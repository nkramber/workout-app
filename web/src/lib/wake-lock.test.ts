import { describe, expect, it, vi } from "vitest";

import { holdWakeLock, type WakeDocument, type WakeNavigator, type WakeState } from "./wake-lock";

type FakeSentinel = {
  released: boolean;
  release: () => Promise<void>;
  addEventListener: (type: "release", fn: () => void) => void;
  drop: () => void;
};

// A fake lock: each request gives a sentinel. hide() releases it as the
// phone does when the app goes to the back, and drop() releases it while
// the app shows, as a power-save mode does.
function fakes(refuse = false) {
  const listeners = new Set<() => void>();
  const sentinels: FakeSentinel[] = [];
  const doc: WakeDocument & { show: () => void; hide: () => void } = {
    visibilityState: "visible",
    addEventListener: (_t, fn) => listeners.add(fn),
    removeEventListener: (_t, fn) => listeners.delete(fn),
    show() {
      this.visibilityState = "visible";
      listeners.forEach((fn) => fn());
    },
    hide() {
      this.visibilityState = "hidden";
      sentinels.forEach((s) => s.drop());
      listeners.forEach((fn) => fn());
    },
  };
  const request = vi.fn(async () => {
    if (refuse) throw new DOMException("refused", "NotAllowedError");
    const onRelease: (() => void)[] = [];
    const s: FakeSentinel = {
      released: false,
      release: vi.fn(async () => s.drop()),
      addEventListener: (_t, fn) => onRelease.push(fn),
      drop: () => {
        if (s.released) return;
        s.released = true;
        onRelease.forEach((fn) => fn());
      },
    };
    sentinels.push(s);
    return s;
  });
  const nav: WakeNavigator = { wakeLock: { request } };
  return { doc, nav, request, sentinels, listeners };
}

const settle = () => new Promise((r) => setTimeout(r, 0));

describe("holdWakeLock (D-265)", () => {
  it("holds the lock, requests it again after a return to the front, and releases it at the stop", async () => {
    const { doc, nav, request, sentinels, listeners } = fakes();
    const states: WakeState[] = [];
    const stop = holdWakeLock(nav, doc, (s) => states.push(s));
    await settle();
    expect(request).toHaveBeenCalledTimes(1);
    expect(states).toEqual(["on"]);

    doc.hide();
    await settle();
    expect(request).toHaveBeenCalledTimes(1);
    doc.show();
    await settle();
    expect(request).toHaveBeenCalledTimes(2);

    stop();
    expect(sentinels[1].release).toHaveBeenCalled();
    expect(listeners.size).toBe(0);
  });

  it("requests the lock again after a release while the app shows, and gives off after a second release", async () => {
    const { doc, nav, request, sentinels } = fakes();
    const states: WakeState[] = [];
    holdWakeLock(nav, doc, (s) => states.push(s));
    await settle();

    sentinels[0].drop();
    await settle();
    expect(request).toHaveBeenCalledTimes(2);
    expect(states).toEqual(["on", "on"]);

    sentinels[1].drop();
    await settle();
    expect(request).toHaveBeenCalledTimes(2);
    expect(states).toEqual(["on", "on", "off"]);

    // A return to the front requests the lock again.
    doc.hide();
    doc.show();
    await settle();
    expect(request).toHaveBeenCalledTimes(3);
    expect(states.at(-1)).toBe("on");
  });

  it("gives off when the phone refuses the lock", async () => {
    const { doc, nav } = fakes(true);
    const states: WakeState[] = [];
    holdWakeLock(nav, doc, (s) => states.push(s));
    await settle();
    expect(states).toEqual(["off"]);
  });

  it("gives off when the phone has no Screen Wake Lock API", () => {
    const { doc } = fakes();
    const states: WakeState[] = [];
    holdWakeLock({}, doc, (s) => states.push(s));
    expect(states).toEqual(["off"]);
  });

  it("releases a lock that arrives after the stop", async () => {
    const { doc, nav, sentinels } = fakes();
    const onState = vi.fn();
    const stop = holdWakeLock(nav, doc, onState);
    stop();
    await settle();
    expect(sentinels[0].release).toHaveBeenCalled();
    expect(onState).not.toHaveBeenCalled();
  });
});
