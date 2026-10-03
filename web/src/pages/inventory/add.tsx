import { useState } from "react";

import type { CatalogMachine } from "../../gen/workoutapp/v1/inventory_service_pb";
import { checkNote, groupByKind, MAX_NOTE_CHARS, searchCatalog } from "../../lib/inventory";
import { useInventoryApi } from "../../lib/inventory-api";
import type { ScreenProps } from "./types";
import { ErrorText, field, secondary, Title, localErrorText, useAction } from "./ui";

// AddScreen adds a machine in one of two ways (D-51, D-55):
//
//   - selection: the catalog list, by kind (D-202), A to Z (D-205),
//   - text entry: the text searches the names of the catalog, and the
//     owner selects a match. A text with no match stays as a note, which
//     no plan uses (D-191).
//
// A machine that the inventory holds opens its review screen, because the
// inventory holds one entry for each machine (D-200).
export function AddScreen({ catalog, inventory, go }: ScreenProps) {
  const api = useInventoryApi();
  const action = useAction(localErrorText);
  const [text, setText] = useState("");
  const held = new Set((inventory?.machines ?? []).map((m) => m.machineId));

  const select = (m: CatalogMachine) =>
    go(held.has(m.id) ? { name: "review", machineId: m.id } : { name: "edit", machineId: m.id });

  const saveNote = async () => {
    const note = checkNote(text);
    if (!note.ok) {
      action.setError(note.error);
      return;
    }
    if (await action.run(() => api.saveNote(note.value))) go({ name: "list" });
  };

  const searching = text.trim() !== "";
  const matches = searching ? searchCatalog(catalog.machines, catalog.exercises, text) : [];

  const row = (m: CatalogMachine) => (
    <li key={m.id}>
      <button
        type="button"
        data-testid={`catalog-${m.id}`}
        onClick={() => select(m)}
        className="flex min-h-11 w-full items-center justify-between gap-3 rounded-lg border border-slate-800 px-3 py-2 text-left text-slate-100 active:bg-slate-800"
      >
        <span>{m.name}</span>
        {held.has(m.id) && <span className="shrink-0 text-xs text-slate-400">In the inventory</span>}
      </button>
    </li>
  );

  return (
    <div className="space-y-6">
      <Title onBack={() => go({ name: "list" })}>Add a machine</Title>

      <label className="block">
        <span className="text-sm text-slate-400">Search the catalog, or enter a name</span>
        <input
          type="text"
          enterKeyHint="search"
          autoComplete="off"
          maxLength={MAX_NOTE_CHARS}
          value={text}
          onChange={(e) => {
            setText(e.target.value);
            action.setError("");
          }}
          className={field}
        />
      </label>

      {searching ? (
        <section className="space-y-2" aria-label="Matches">
          <h3 className="text-sm font-semibold text-slate-300">Matches</h3>
          {matches.length === 0 ? (
            <p className="text-sm text-slate-400" data-testid="no-match">
              No machine of the catalog matches this text.
            </p>
          ) : (
            <ul className="space-y-2">{matches.map(row)}</ul>
          )}
          <p className="text-xs text-slate-500">
            When no match is your machine, keep the text as a note. No plan uses a note.
          </p>
          <button type="button" disabled={action.busy} className={`${secondary} w-full`} onClick={() => void saveNote()}>
            Keep the text as a note
          </button>
        </section>
      ) : (
        groupByKind(catalog.machines).map((g) => (
          <section key={g.kind} className="space-y-2" aria-label={g.title}>
            <h3 className="text-sm font-semibold text-slate-300">{g.title}</h3>
            <ul className="space-y-2">{g.machines.map(row)}</ul>
          </section>
        ))
      )}

      <ErrorText>{action.error}</ErrorText>
    </div>
  );
}
