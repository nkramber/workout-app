import { useEffect, useState } from "react";

// The screen wake lock of a workout (D-265, D-271, D-283). The app
// requests the lock while a workout is open, on each screen. The phone
// releases the lock when the app goes to the back, and it can release
// the lock while the app shows, for example in a power-save mode. So the
// app requests the lock again at each return to the front, at each
// focus, and at each `pageshow` event, with no tap. This is the method
// "Wake Lock, no tap" that the owner tested on the iPhone (D-283). A tap
// requests the lock too, and it never stops while the workout is open.
// A release makes no request of its own: the next of these events
// requests the lock. Only the newest request sets the state, so a slow
// request of an earlier event can not change it. A phone with no Screen
// Wake Lock API, or one that refuses the lock, gives "off", and the
// screen shows a notice with the error name.

export type WakeState = "pending" | "on" | "off";

// WakeStatus is the state of the lock. "pending" is the state before the
// first answer, and after a release until the next request. For "off",
// error names the cause: the name of the error of the last refused
// request, or "unsupported" for a phone with no Screen Wake Lock API.
export type WakeStatus = { state: WakeState; error: string };

type Sentinel = { released: boolean; release: () => Promise<void>; addEventListener: (type: "release", fn: () => void) => void };
export type WakeNavigator = { wakeLock?: { request: (type: "screen") => Promise<Sentinel> } };
type Events = {
  addEventListener: (type: string, fn: () => void, capture?: boolean) => void;
  removeEventListener: (type: string, fn: () => void, capture?: boolean) => void;
};
// The document gives "visibilitychange" and "pointerdown", and the window
// gives "focus" and "pageshow".
export type WakeDocument = Events & { visibilityState: string };
export type WakeWindow = Events;

// holdWakeLock requests the lock, and requests it again at each event of
// D-283. It gives each state to onState, and gives a stop function that
// releases the lock.
export function holdWakeLock(
  nav: WakeNavigator,
  doc: WakeDocument,
  win: WakeWindow,
  onState: (state: WakeState, error: string) => void,
): () => void {
  const api = nav.wakeLock;
  if (!api) {
    onState("off", "unsupported");
    return () => {};
  }
  let stopped = false;
  let sentinel: Sentinel | null = null;
  // The number of the newest request. An answer of an earlier request
  // does not change the state, and its lock goes back to the phone.
  let latest = 0;

  const request = async () => {
    if (stopped || doc.visibilityState !== "visible") return;
    if (sentinel && !sentinel.released) return;
    const n = ++latest;
    try {
      const s = await api.request("screen");
      if (stopped || n !== latest) {
        await s.release();
        return;
      }
      sentinel = s;
      s.addEventListener("release", () => {
        if (!stopped && s === sentinel) onState("pending", "");
      });
      onState("on", "");
    } catch (err) {
      if (!stopped && n === latest) onState("off", err instanceof Error && err.name ? err.name : "refused");
    }
  };

  // Each event of D-283. A tap gives the user activation that a phone can
  // need for the request.
  const onEvent = () => void request();
  doc.addEventListener("visibilitychange", onEvent);
  doc.addEventListener("pointerdown", onEvent, true);
  win.addEventListener("focus", onEvent);
  win.addEventListener("pageshow", onEvent);
  void request();

  return () => {
    stopped = true;
    doc.removeEventListener("visibilitychange", onEvent);
    doc.removeEventListener("pointerdown", onEvent, true);
    win.removeEventListener("focus", onEvent);
    win.removeEventListener("pageshow", onEvent);
    if (sentinel && !sentinel.released) void sentinel.release().catch(() => {});
  };
}

// useWakeLock holds the lock while `enabled` is true.
export function useWakeLock(enabled: boolean): WakeStatus {
  const [status, setStatus] = useState<WakeStatus>({ state: "pending", error: "" });
  useEffect(() => {
    if (!enabled) return;
    setStatus({ state: "pending", error: "" });
    return holdWakeLock(navigator as unknown as WakeNavigator, document, window, (state, error) => setStatus({ state, error }));
  }, [enabled]);
  return status;
}
