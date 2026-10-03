import { clone, create, fromJson, toJson, type JsonValue, type MessageInitShape } from "@bufbuild/protobuf";
import { useLiveQuery } from "dexie-react-hooks";
import { useMemo } from "react";

import {
  ConfirmMachineRequestSchema,
  GetCatalogResponseSchema,
  GetInventoryResponseSchema,
  InventoryMachineSchema,
  InventoryNoteSchema,
  InventorySchema,
  MachineState,
  SaveMachineRequestSchema,
  type DumbbellSet,
  type GetCatalogResponse,
  type Inventory,
  type InventoryMachine,
} from "../gen/workoutapp/v1/inventory_service_pb";
import { MachineConfirmSchema, MachineSaveSchema, NoteSaveSchema } from "../gen/workoutapp/v1/workout_service_pb";
import { db, INVENTORY_ENTITIES, OUTBOX_SCHEMA_VERSION, withReopen, type OutboxEntry, type WorkoutAppDB } from "./db";
import { nextId } from "./uuidv7";

// The inventory of the phone (D-250, D-272). The phone keeps a copy of
// the inventory that the server gave, and each change of the owner goes
// into the outbox, with no call. The screens show the copy with the
// inventory entries of the outbox on it, by the rules of the server. When
// the sync reads the inventory again, the server wins, and the entries
// that wait go on the new copy (D-258). A confirmation with no connection
// shows the machine as confirmed until the server answers. When the
// server refuses it, the copy shows the draft again (D-273).

// addEntry writes one inventory entry to the outbox. The payload is the
// JSON form of the payload field of OutboxEntry, such as
// {"saveMachine": {...}}.
function addEntry(store: WorkoutAppDB, entity: "machine" | "note", entityId: string, payload: Record<string, JsonValue>, now: Date): Promise<OutboxEntry> {
  return withReopen(store, async () => {
    const entry: OutboxEntry = {
      opId: nextId(now.getTime()),
      entity,
      entityId,
      baseVersion: 0,
      payload,
      at: now.toISOString(),
      attempts: 0,
      schemaVersion: OUTBOX_SCHEMA_VERSION,
    };
    await store.outbox.add(entry);
    return entry;
  });
}

export function saveMachine(store: WorkoutAppDB, req: MessageInitShape<typeof SaveMachineRequestSchema>, now = new Date()): Promise<OutboxEntry> {
  const m = create(SaveMachineRequestSchema, req);
  const save = create(MachineSaveSchema, { weightsTenthLb: m.weightsTenthLb, dumbbells: m.dumbbells, estimates: m.estimates });
  return addEntry(store, "machine", m.machineId, { saveMachine: toJson(MachineSaveSchema, save) }, now);
}

export function confirmMachine(store: WorkoutAppDB, req: MessageInitShape<typeof ConfirmMachineRequestSchema>, now = new Date()): Promise<OutboxEntry> {
  const m = create(ConfirmMachineRequestSchema, req);
  const confirm = create(MachineConfirmSchema, { weightsTenthLb: m.weightsTenthLb, dumbbells: m.dumbbells });
  return addEntry(store, "machine", m.machineId, { confirmMachine: toJson(MachineConfirmSchema, confirm) }, now);
}

export function removeMachine(store: WorkoutAppDB, machineId: string, now = new Date()): Promise<OutboxEntry> {
  return addEntry(store, "machine", machineId, { removeMachine: {} }, now);
}

// saveNote adds a note, with an id that the phone makes (D-272), or
// changes the text of the note id.
export function saveNote(store: WorkoutAppDB, text: string, id = "", now = new Date()): Promise<OutboxEntry> {
  const noteId = id || nextId(now.getTime());
  return addEntry(store, "note", noteId, { saveNote: toJson(NoteSaveSchema, create(NoteSaveSchema, { text })) }, now);
}

export function removeNote(store: WorkoutAppDB, id: string, now = new Date()): Promise<OutboxEntry> {
  return addEntry(store, "note", id, { removeNote: {} }, now);
}

type Weights = { weightsTenthLb: number[]; dumbbells?: DumbbellSet };

// sameWeights tells whether two machines give the same weights: the same
// list, or the same dumbbell set, or none, as the server reads them.
function sameWeights(a: Weights, b: Weights): boolean {
  if (!a.dumbbells !== !b.dumbbells) return false;
  if (a.dumbbells && b.dumbbells) {
    const [x, y] = [a.dumbbells, b.dumbbells];
    if (x.lightestTenthLb !== y.lightestTenthLb || x.heaviestTenthLb !== y.heaviestTenthLb || x.stepTenthLb !== y.stepTenthLb) return false;
  }
  return a.weightsTenthLb.length === b.weightsTenthLb.length && a.weightsTenthLb.every((w, i) => w === b.weightsTenthLb[i]);
}

// Pending names the machines and the notes with an entry that waits in
// the outbox.
export type Pending = { machines: ReadonlySet<string>; notes: ReadonlySet<string> };

// localInventory puts the inventory entries of the outbox on the copy of
// the server, in the order of the op ids, by the rules of InventoryService:
//
//   - a save makes a draft, and keeps a confirmed machine confirmed when
//     its weights do not change. The save of a cardio machine confirms it
//     (D-193, D-246).
//   - a confirmation confirms the machine when the weights agree. Else the
//     machine stays as it is, and the server refuses the entry (D-201).
//   - a removal of an unknown machine or note changes nothing.
//
// The machines stay in catalog order. An entry that the phone can not
// read changes nothing, and the sync moves it to the refused entries.
export function localInventory(
  server: Inventory | undefined,
  entries: readonly OutboxEntry[],
  catalog: GetCatalogResponse,
): { inventory: Inventory; pending: Pending } {
  const inv = server ? clone(InventorySchema, server) : create(InventorySchema);
  const machines = new Set<string>();
  const notes = new Set<string>();
  const kind = new Map(catalog.machines.map((m) => [m.id, m.kind]));
  for (const e of entries) {
    if (!INVENTORY_ENTITIES.has(e.entity)) continue;
    const p = e.payload as Record<string, JsonValue>;
    try {
      if (e.entity === "machine") {
        machines.add(e.entityId);
        applyMachine(inv, e.entityId, p, kind.get(e.entityId) === "cardio");
      } else {
        notes.add(e.entityId);
        applyNote(inv, e.entityId, p);
      }
    } catch {
      // A payload that the contract does not accept changes nothing.
    }
  }
  const order = new Map(catalog.machines.map((m, i) => [m.id, i]));
  const pos = (id: string) => order.get(id) ?? order.size;
  inv.machines.sort((a, b) => pos(a.machineId) - pos(b.machineId));
  return { inventory: inv, pending: { machines, notes } };
}

function applyMachine(inv: Inventory, id: string, p: Record<string, JsonValue>, cardio: boolean) {
  const i = inv.machines.findIndex((m) => m.machineId === id);
  if ("saveMachine" in p) {
    const save = fromJson(MachineSaveSchema, p.saveMachine);
    const m: InventoryMachine = create(InventoryMachineSchema, {
      machineId: id,
      weightsTenthLb: save.weightsTenthLb,
      dumbbells: save.dumbbells,
      estimates: save.estimates,
      state: cardio ? MachineState.CONFIRMED : MachineState.DRAFT,
    });
    if (i < 0) {
      inv.machines.push(m);
      return;
    }
    if (inv.machines[i].state === MachineState.CONFIRMED && sameWeights(inv.machines[i], m)) m.state = MachineState.CONFIRMED;
    inv.machines[i] = m;
  } else if ("confirmMachine" in p) {
    const shown = fromJson(MachineConfirmSchema, p.confirmMachine);
    if (i >= 0 && sameWeights(inv.machines[i], shown)) inv.machines[i].state = MachineState.CONFIRMED;
  } else if ("removeMachine" in p && i >= 0) {
    inv.machines.splice(i, 1);
  }
}

function applyNote(inv: Inventory, id: string, p: Record<string, JsonValue>) {
  const i = inv.notes.findIndex((n) => n.id === id);
  if ("saveNote" in p) {
    const text = fromJson(NoteSaveSchema, p.saveNote).text.trim();
    if (i >= 0) inv.notes[i].text = text;
    else inv.notes.push(create(InventoryNoteSchema, { id, text }));
  } else if ("removeNote" in p && i >= 0) {
    inv.notes.splice(i, 1);
  }
}

// OfflineInventory is what the screens of the inventory read: the catalog
// and the inventory of the phone, with the items that wait to sync.
export type OfflineInventory = { catalog: GetCatalogResponse; inventory: Inventory; pending: Pending };

// readOfflineInventory reads the copies and the outbox of the store. It
// gives null when the phone has no copy of the catalog or the inventory.
export async function readOfflineInventory(store: WorkoutAppDB): Promise<OfflineInventory | null> {
  const [catalogCopy, inventoryCopy, entries] = await Promise.all([
    store.copies.get("catalog"),
    store.copies.get("inventory"),
    store.outbox.toArray(),
  ]);
  if (!catalogCopy || !inventoryCopy) return null;
  const catalog = fromJson(GetCatalogResponseSchema, catalogCopy.json as JsonValue, { ignoreUnknownFields: true });
  const server = fromJson(GetInventoryResponseSchema, inventoryCopy.json as JsonValue, { ignoreUnknownFields: true }).inventory;
  return { catalog, ...localInventory(server, entries, catalog) };
}

// useOfflineInventory gives the catalog and the inventory of the phone.
// It gives undefined while the store loads, and null when the phone has
// no copy yet.
export function useOfflineInventory(): OfflineInventory | null | undefined {
  return useLiveQuery(() => withReopen(db, () => readOfflineInventory(db)), [], undefined);
}

// useInventoryApi gives the changes of the inventory. Each change goes
// into the outbox of the phone, and the sync sends it (D-250).
export function useInventoryApi() {
  return useMemo(
    () => ({
      saveMachine: async (req: MessageInitShape<typeof SaveMachineRequestSchema>) => void (await saveMachine(db, req)),
      confirmMachine: async (req: MessageInitShape<typeof ConfirmMachineRequestSchema>) => void (await confirmMachine(db, req)),
      removeMachine: async (machineId: string) => void (await removeMachine(db, machineId)),
      saveNote: async (text: string) => void (await saveNote(db, text)),
      removeNote: async (id: string) => void (await removeNote(db, id)),
    }),
    [],
  );
}
