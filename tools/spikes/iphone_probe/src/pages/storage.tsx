import { useCallback, useEffect, useState } from "react";

import { addNote, clearNotes, listNotes, type Note } from "../lib/db";
import { launchId } from "../lib/launch";
import { Button, Row, Section, yesNo } from "../ui";

// PR-6, item 1: write a record, stop the app, open it again. A record of
// an earlier launch that the page still reads is the pass. The page also
// shows whether the browser marks the storage as persistent, because
// Safari can clear the data of a site that the user does not open.
export function StoragePage() {
  const [notes, setNotes] = useState<Note[] | null>(null);
  const [error, setError] = useState("");
  const [persisted, setPersisted] = useState<boolean | null>(null);
  const [usage, setUsage] = useState("unknown");

  const refresh = useCallback(async () => {
    try {
      setNotes(await listNotes());
      setError("");
    } catch (err) {
      setError(String(err));
    }
    if (navigator.storage?.persisted) setPersisted(await navigator.storage.persisted());
    if (navigator.storage?.estimate) {
      const est = await navigator.storage.estimate();
      setUsage(`${Math.round((est.usage ?? 0) / 1024)} KB`);
    }
  }, []);

  useEffect(() => {
    void refresh();
  }, [refresh]);

  const write = async () => {
    const n = (notes?.length ?? 0) + 1;
    await addNote({ text: `record ${n}`, at: new Date().toISOString(), launchId });
    await refresh();
  };

  const askPersist = async () => {
    if (navigator.storage?.persist) setPersisted(await navigator.storage.persist());
  };

  const earlier = notes?.filter((n) => n.launchId !== launchId).length ?? 0;

  return (
    <>
      <Section title="IndexedDB after an app stop">
        <Row label="Records" value={notes === null ? "…" : notes.length} testId="note-count" />
        <Row label="Records from an earlier launch" value={notes === null ? "…" : earlier} testId="note-earlier" />
        <Row label="Persistent storage" value={yesNo(persisted)} testId="storage-persisted" />
        <Row label="Storage used" value={usage} testId="storage-usage" />
        {error && (
          <p className="text-red-400" data-testid="storage-error">
            {error}
          </p>
        )}
        <div className="flex flex-wrap gap-2">
          <Button onClick={() => void write()}>Write a record</Button>
          <Button onClick={() => void askPersist()}>Ask for persistent storage</Button>
          <Button onClick={() => void clearNotes().then(refresh)}>Clear</Button>
        </div>
      </Section>
      <Section title="Stored records">
        <ul className="space-y-1 font-mono text-xs" data-testid="note-list">
          {[...(notes ?? [])].reverse().map((n) => (
            <li key={n.id}>
              {n.text} · {n.at.slice(11, 19)} · {n.launchId === launchId ? "this launch" : `launch ${n.launchId}`}
            </li>
          ))}
        </ul>
      </Section>
    </>
  );
}
