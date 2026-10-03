import { useEffect, useState } from "react";

// The screen wake lock of a workout (D-265, D-271). The app requests the
// lock while a workout is open, on each screen. The phone releases the
// lock when the app goes to the back, and it can refuse a request or
// release the lock while the app shows, for example in a power-save mode.
// So the app requests the lock again at each return to the front, at each
// focus, at each `pageshow` event, and at each tap. It never stops while
// the workout is open. After a release while the app shows, the app
// requests the lock one more time, and a second release gives "off"
// until the next of these events, so a phone that always releases the
// lock does not get a loop of requests. A phone with no Screen Wake Lock
// API, or one that refuses the lock, gives "off", and the screen shows a
// notice with the error name.

export type WakeState = "pending" | "on" | "off";

// WakeStatus is the state of the lock. For "off", error names the cause:
// the name of the error of the last refused request, "released" after a
// second release while the app shows, or "unsupported" for a phone with
// no Screen Wake Lock API.
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
// D-271. It gives each state to onState, and gives a stop function that
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
  // True after the request again of a release while the app shows. Each
  // event of D-271 sets it to false.
  let retried = false;

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
      s.addEventListener("release", () => onRelease(s));
      onState("on", "");
    } catch (err) {
      if (!stopped) onState("off", err instanceof Error && err.name ? err.name : "refused");
    }
  };

  // A release while the app goes to the back needs no step: the return to
  // the front requests the lock again.
  const onRelease = (s: Sentinel) => {
    if (stopped || s !== sentinel || doc.visibilityState !== "visible") return;
    if (retried) {
      onState("off", "released");
      return;
    }
    retried = true;
    void request();
  };

  // Each event of D-271. A tap gives the user activation that a phone can
  // need for the request.
  const onEvent = () => {
    if (doc.visibilityState === "visible") retried = false;
    void request();
  };
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
