import { useEffect, useState } from "react";

// The screen wake lock of a workout (D-265). The app requests the lock
// while a workout is open, on each screen. The phone releases the lock when the app goes
// to the back, so the app requests it again when the app comes back to
// the front. A phone with no Screen Wake Lock API, or one that refuses
// the lock, gives the state "off", and the screen shows a notice.

export type WakeState = "pending" | "on" | "off";

type Sentinel = { released: boolean; release: () => Promise<void>; addEventListener: (type: "release", fn: () => void) => void };
export type WakeNavigator = { wakeLock?: { request: (type: "screen") => Promise<Sentinel> } };
export type WakeDocument = {
  visibilityState: string;
  addEventListener: (type: "visibilitychange", fn: () => void) => void;
  removeEventListener: (type: "visibilitychange", fn: () => void) => void;
};

// holdWakeLock requests the lock, and requests it again at each return to
// the front. It gives each state to onState, and gives a stop function
// that releases the lock.
export function holdWakeLock(nav: WakeNavigator, doc: WakeDocument, onState: (s: WakeState) => void): () => void {
  const api = nav.wakeLock;
  if (!api) {
    onState("off");
    return () => {};
  }
  let stopped = false;
  let sentinel: Sentinel | null = null;

  const request = async () => {
    if (stopped || doc.visibilityState !== "visible") return;
    if (sentinel && !sentinel.released) return;
    try {
      const s = await api.request("screen");
      if (stopped) {
        await s.release();
        return;
      }
      sentinel = s;
      onState("on");
    } catch {
      if (!stopped) onState("off");
    }
  };

  const onVisible = () => void request();
  doc.addEventListener("visibilitychange", onVisible);
  void request();

  return () => {
    stopped = true;
    doc.removeEventListener("visibilitychange", onVisible);
    if (sentinel && !sentinel.released) void sentinel.release().catch(() => {});
  };
}

// useWakeLock holds the lock while `enabled` is true.
export function useWakeLock(enabled: boolean): WakeState {
  const [state, setState] = useState<WakeState>("pending");
  useEffect(() => {
    if (!enabled) return;
    setState("pending");
    return holdWakeLock(navigator as unknown as WakeNavigator, document, setState);
  }, [enabled]);
  return state;
}
