// The update strategy of the service worker (REC-3, D-133). The plugin
// runs in prompt mode: a new worker waits, and the app shows "Update
// ready". The owner applies the update, and never during a workout. A
// reload during a workout could lose a set that the owner is typing.
// This file holds the parts that need no worker, so the unit tests read
// them. src/lib/pwa.ts registers the worker.

// The browser checks for a new worker on a navigation alone. An installed
// app also asks each hour and on each return to view, as in Decktome.
export const swUpdateIntervalMs = 60 * 60 * 1000;

// updateAllowed says if the app can apply a waiting update now. Phase 4
// adds the workout screen and its state. Until then no workout runs.
export function updateAllowed(workoutActive: boolean): boolean {
  return !workoutActive;
}

type Deps = {
  online: () => boolean;
  fetch: typeof fetch;
};

// checkForUpdate asks the registration for a new worker. It first reads
// the worker file with no cache, so an offline phone or a failed server
// makes no update check that fails half way.
export async function checkForUpdate(
  registration: { installing: ServiceWorker | null; update: () => Promise<unknown> },
  swUrl: string,
  deps: Deps = { online: () => navigator.onLine, fetch: (...a) => fetch(...a) },
): Promise<boolean> {
  if (registration.installing || !deps.online()) return false;
  try {
    const res = await deps.fetch(swUrl, { cache: "no-store", headers: { "cache-control": "no-cache" } });
    if (res.status !== 200) return false;
    await registration.update();
    return true;
  } catch {
    return false;
  }
}
