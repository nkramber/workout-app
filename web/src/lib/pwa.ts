import { registerSW } from "virtual:pwa-register";

import { checkForUpdate, swUpdateIntervalMs, updateAllowed } from "./update-check";

// The registration of the service worker, and the state of a waiting
// update (REC-3, D-133). src/lib/update-check.ts holds the rules.

// The state of the waiting update, for the banner of src/update-banner.tsx.
let waiting = false;
let applyUpdate: (() => Promise<void>) | null = null;
const listeners = new Set<() => void>();

export function subscribeUpdate(listener: () => void): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

export function updateWaiting(): boolean {
  return waiting;
}

// applyWaitingUpdate activates the waiting worker, and the page reloads.
export async function applyWaitingUpdate(workoutActive: boolean): Promise<void> {
  if (!waiting || !applyUpdate || !updateAllowed(workoutActive)) return;
  await applyUpdate();
}

// startServiceWorker registers the worker of the build.
export function startServiceWorker(): void {
  if (!("serviceWorker" in navigator)) return;
  const update = registerSW({
    immediate: true,
    onNeedRefresh() {
      waiting = true;
      for (const l of listeners) l();
    },
    onRegisteredSW(swUrl, registration) {
      if (!registration) return;
      const check = () => void checkForUpdate(registration, swUrl);
      setInterval(check, swUpdateIntervalMs);
      document.addEventListener("visibilitychange", () => {
        if (document.visibilityState === "visible") check();
      });
    },
  });
  applyUpdate = () => update(true);
}
