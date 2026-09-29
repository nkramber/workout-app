// The IndexedDB store of the probe. It holds the records of the storage
// page and the startup records. The device checklist writes a record,
// stops the app, opens it again, and reads the record back (PR-6, item 1).
// The raw API keeps the probe free of a storage library, so the result
// is the result of the platform.

const DB_NAME = "gym-route-probe";
const DB_VERSION = 1;

export type Note = { id?: number; text: string; at: string; launchId: string };

export type Startup = {
  id?: number;
  at: string;
  launchId: string;
  build: string;
  displayMode: string;
  navigationType: string;
  controlled: boolean;
  htmlMs: number | null;
  mainMs: number | null;
  renderMs: number | null;
  fcpMs: number | null;
  domContentLoadedMs: number | null;
};

type StoreName = "notes" | "startups";

let pending: Promise<IDBDatabase> | null = null;

function open(): Promise<IDBDatabase> {
  pending ??= new Promise((resolve, reject) => {
    const req = indexedDB.open(DB_NAME, DB_VERSION);
    req.onupgradeneeded = () => {
      const db = req.result;
      if (!db.objectStoreNames.contains("notes")) db.createObjectStore("notes", { keyPath: "id", autoIncrement: true });
      if (!db.objectStoreNames.contains("startups")) db.createObjectStore("startups", { keyPath: "id", autoIncrement: true });
    };
    req.onsuccess = () => resolve(req.result);
    req.onerror = () => {
      pending = null;
      reject(req.error);
    };
  });
  return pending;
}

// run opens one transaction and resolves when the transaction commits,
// so a caller that reads after a write reads the written value.
async function run<T>(store: StoreName, mode: IDBTransactionMode, body: (s: IDBObjectStore) => IDBRequest<T>): Promise<T> {
  const db = await open();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(store, mode);
    const req = body(tx.objectStore(store));
    tx.oncomplete = () => resolve(req.result);
    tx.onerror = () => reject(tx.error);
    tx.onabort = () => reject(tx.error);
  });
}

export const addNote = (note: Note) => run("notes", "readwrite", (s) => s.add(note));
export const listNotes = () => run<Note[]>("notes", "readonly", (s) => s.getAll());
export const clearNotes = () => run("notes", "readwrite", (s) => s.clear());
export const addStartup = (rec: Startup) => run("startups", "readwrite", (s) => s.add(rec));
export const listStartups = () => run<Startup[]>("startups", "readonly", (s) => s.getAll());
export const clearStartups = () => run("startups", "readwrite", (s) => s.clear());
