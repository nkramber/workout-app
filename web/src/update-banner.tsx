import { useLiveQuery } from "dexie-react-hooks";
import { useSyncExternalStore } from "react";

import { db } from "./lib/db";
import { applyWaitingUpdate, subscribeUpdate, updateWaiting } from "./lib/pwa";
import { bannerShown } from "./lib/update-check";
import { activeWorkout } from "./lib/workout";

// UpdateBanner shows "Update ready" when a new service worker waits
// (REC-3, D-133). The owner applies it with the button. During a workout
// the banner does not show, so the workout screen does not move. It
// shows when the workout ends (D-323).
export function UpdateBanner() {
  const waiting = useSyncExternalStore(subscribeUpdate, updateWaiting);
  // An open workout on the phone holds the update. While the store
  // loads, the banner holds it too.
  const workout = useLiveQuery(() => activeWorkout(db), [], null);
  const workoutActive = workout !== undefined;
  if (!bannerShown(waiting, workoutActive)) return null;
  return (
    <div role="status" className="flex items-center justify-between gap-3 bg-sky-950 px-4 py-2 text-sm text-sky-100">
      <span>Update ready</span>
      <button
        type="button"
        onClick={() => void applyWaitingUpdate(workoutActive)}
        className="min-h-11 rounded-lg bg-sky-600 px-4 font-medium text-white active:bg-sky-700"
      >
        Update
      </button>
    </div>
  );
}
