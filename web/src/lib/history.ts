import { REST_KEY, withReopen, type OutboxEntry, type WorkoutAppDB } from "./db";

// DELETE_CONFIRMATION is the text that the owner types before the
// deletion of all data (D-314). The server checks the same text.
export const DELETE_CONFIRMATION = "Delete all data";

// The entities of the outbox that hold the history (D-315). The entries
// of the inventory stay.
const HISTORY_ENTITIES: ReadonlySet<string> = new Set(["workout", "set", "cardio"]);

// canDelete tells whether the button of the deletion works: the switch
// is at "Yes", and the text is the confirmation, with no change (D-314).
export function canDelete(yes: boolean, text: string): boolean {
  return yes && text === DELETE_CONFIRMATION;
}

// deleteLocalHistory deletes the history on the phone in one
// transaction (D-315): each workout, set, and cardio log, their outbox
// and refused entries, the plan copy, and the rest timer. The profile,
// the inventory, and the catalog stay.
export function deleteLocalHistory(store: WorkoutAppDB): Promise<void> {
  const history = (e: OutboxEntry) => HISTORY_ENTITIES.has(e.entity);
  return withReopen(store, () =>
    store.transaction("rw", [store.workouts, store.sets, store.cardio, store.outbox, store.refused, store.copies, store.meta], async () => {
      await store.workouts.clear();
      await store.sets.clear();
      await store.cardio.clear();
      await store.outbox.filter(history).delete();
      await store.refused.filter(history).delete();
      await store.copies.delete("plan");
      await store.meta.delete(REST_KEY);
    }),
  );
}

// deleteAllData deletes the history on the phone, then on the server
// (D-314, D-315). The phone deletes its copy first, so no later sync
// sends a deleted entry again. The first sync waits for a sync that runs,
// because that sync can hold an entry that it read before the delete.
// The second sync reads the plan of a user with no plan. A failed call
// of the server throws, and a second call deletes the rest. It gives the
// count of the workouts that the server deleted.
export async function deleteAllData(
  store: WorkoutAppDB,
  deps: { sync: () => Promise<void>; deleteOnServer: () => Promise<number> },
): Promise<number> {
  await deleteLocalHistory(store);
  await deps.sync();
  const deleted = await deps.deleteOnServer();
  await deps.sync();
  return deleted;
}
