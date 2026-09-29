import type { GymRouteDB } from "./db";

// The persistent storage request (REC-5, D-134). Without it, the browser
// can remove the offline store when the phone needs space. The app asks
// one time on each device, after the first sign-in, and keeps the answer
// in the meta table. The home screen shows the state and the use of the
// store, so the device check can read them.
export const PERSIST_KEY = "storage-persist";

export type StorageState = {
  persisted: boolean | null; // null when the browser gives no API
  requestedAt: string | null;
  usage: number | null;
  quota: number | null;
};

type StorageApi = Pick<StorageManager, "persist" | "persisted" | "estimate">;

// requestPersistenceOnce asks for persistent storage on the first call on
// this device. A later call reads the state and asks again never.
export async function requestPersistenceOnce(
  store: GymRouteDB,
  api: StorageApi | undefined = navigator.storage,
  now: Date = new Date(),
): Promise<StorageState> {
  if (!api?.persist) return { persisted: null, requestedAt: null, usage: null, quota: null };
  const earlier = (await store.meta.get(PERSIST_KEY))?.value as { at: string } | undefined;
  let requestedAt = earlier?.at ?? null;
  let persisted: boolean;
  if (requestedAt === null) {
    persisted = await api.persist();
    requestedAt = now.toISOString();
    await store.meta.put({ key: PERSIST_KEY, value: { at: requestedAt, granted: persisted } });
  } else {
    persisted = await api.persisted();
  }
  const { usage, quota } = await api.estimate().catch(() => ({ usage: undefined, quota: undefined }));
  return { persisted, requestedAt, usage: usage ?? null, quota: quota ?? null };
}

// megabytes gives a byte count in MB with one decimal, or "n/a".
export function megabytes(bytes: number | null): string {
  return bytes === null ? "n/a" : `${(bytes / 1_000_000).toFixed(1)} MB`;
}
