import { useSyncExternalStore } from "react";

import { applyWaitingUpdate, subscribeUpdate, updateWaiting } from "./lib/pwa";
import { updateAllowed } from "./lib/update-check";

// The workout screen of Phase 4 sets this state. Until then no workout
// runs, so the owner can apply each update.
const workoutActive = false;

// UpdateBanner shows "Update ready" when a new service worker waits
// (REC-3, D-133). The owner applies it with the button. During a workout
// the banner stays, and the button waits until the workout ends.
export function UpdateBanner() {
  const waiting = useSyncExternalStore(subscribeUpdate, updateWaiting);
  if (!waiting) return null;
  const allowed = updateAllowed(workoutActive);
  return (
    <div role="status" className="flex items-center justify-between gap-3 bg-sky-950 px-4 py-2 text-sm text-sky-100">
      <span>{allowed ? "Update ready" : "Update ready after the workout"}</span>
      {allowed && (
        <button
          type="button"
          onClick={() => void applyWaitingUpdate(workoutActive)}
          className="min-h-11 rounded-lg bg-sky-600 px-4 font-medium text-white active:bg-sky-700"
        >
          Update
        </button>
      )}
    </div>
  );
}
