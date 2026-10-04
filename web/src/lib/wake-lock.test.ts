import { describe, expect, it, vi } from "vitest";

import { holdWakeLock, type WakeDocument, type WakeNavigator, type WakeState, type WakeWindow } from "./wake-lock";

type FakeSentinel = {
  released: boolean;
  release: () => Promise<void>;
  addEventListener: (type: "release", fn: () => void) => void;
  drop: () => void;
};

// A fake lock: each request gives a sentinel. hide() releases it as the
// phone does when the app goes to the back, and drop() releases it while
// the app shows, as a power-save mode does. tap(), focus(), and
// pageshow() send the other events of D-283. With refuse, each request
// fails until refuse is false.
function fakes(refuse = false) {
  const listeners = new Map<string, Set<() => void>>();
  const on = (t: string, fn: () => void) => {
    if (!listeners.has(t)) listeners.set(t, new Set());
    listeners.get(t)!.add(fn);
  };
  const off = (t: string, fn: () => void) => listeners.get(t)?.delete(fn);
  const send = (t: string) => listeners.get(t)?.forEach((fn) => fn());
  const sentinels: FakeSentinel[] = [];
  const state = { refuse };
  const doc: WakeDocument & { show: () => void; hide: () => void; tap: () => void } = {
    visibilityState: "visible",
    addEventListener: on,
    removeEventListener: off,
    show() {
      this.visibilityState = "visible";
      send("visibilitychange");
    },
    hide() {
      this.visibilityState = "hidden";
      sentinels.forEach((s) => s.drop());
      send("visibilitychange");
    },
    tap: () => send("pointerdown"),
  };
  const win: WakeWindow & { focus: () => void; pageshow: () => void } = {
    addEventListener: on,
    removeEventListener: off,
    focus: () => send("focus"),
    pageshow: () => send("pageshow"),
  };
  const request = vi.fn(async () => {
    if (state.refuse) throw new DOMException("refused", "NotAllowedError");
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
  const count = () => [...listeners.values()].reduce((n, l) => n + l.size, 0);
  return { doc, win, nav, request, sentinels, count, state };
}

const settle = () => new Promise((r) => setTimeout(r, 0));

// track gives the state list and the onState function of a test.
function track() {
  const states: WakeState[] = [];
  const errors: string[] = [];
  return { states, errors, onState: (st: WakeState, e: string) => (states.push(st), errors.push(e)) };
}

describe("holdWakeLock (D-265, D-283)", () => {
  it("holds the lock, gets it again at a return to the front with no tap, and releases it at the stop", async () => {
    const { doc, win, nav, request, sentinels, count } = fakes();
    const t = track();
    const stop = holdWakeLock(nav, doc, win, t.onState);
    await settle();
    expect(request).toHaveBeenCalledTimes(1);
    expect(t.states).toEqual(["on"]);

    doc.hide();
    await settle();
    expect(request).toHaveBeenCalledTimes(1);
    doc.show();
    await settle();
    expect(request).toHaveBeenCalledTimes(2);
    expect(t.states).toEqual(["on", "pending", "on"]);

    stop();
    expect(sentinels[1].release).toHaveBeenCalled();
    expect(count()).toBe(0);
  });

  it("makes no request at a release while the app shows, and the next event gets the lock with no notice", async () => {
    const { doc, win, nav, request, sentinels } = fakes();
    const t = track();
    holdWakeLock(nav, doc, win, t.onState);
    await settle();

    // A release gives no notice, and makes no request of its own (D-283).
    sentinels[0].drop();
    await settle();
    expect(request).toHaveBeenCalledTimes(1);
    expect(t.states).toEqual(["on", "pending"]);

    // A focus requests the lock again, with no tap.
    win.focus();
    await settle();
    expect(request).toHaveBeenCalledTimes(2);
    expect(t.states.at(-1)).toBe("on");

    // A release of an earlier lock does not change the state.
    sentinels[0].drop();
    await settle();
    expect(t.states.at(-1)).toBe("on");

    // A tap requests the lock after a release too, and a tap while the
    // lock holds makes no request.
    sentinels[1].drop();
    doc.tap();
    await settle();
    expect(request).toHaveBeenCalledTimes(3);
    expect(t.states.at(-1)).toBe("on");
    doc.tap();
    await settle();
    expect(request).toHaveBeenCalledTimes(3);
  });

  it("keeps the state of the newest request when an earlier request answers late", async () => {
    // The first request of this phone answers only when the test says so.
    let answer: { ok: () => void; fail: () => void } | null = null;
    const late = { released: false, release: vi.fn(async () => {}), addEventListener: () => {} };
    const fresh = { released: false, release: vi.fn(async () => {}), addEventListener: () => {} };
    const request = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((ok, fail) => {
            answer = { ok: () => ok(late), fail: () => fail(new DOMException("refused", "NotAllowedError")) };
          }),
      )
      .mockImplementation(async () => fresh);
    const { doc, win } = fakes();
    const t = track();
    holdWakeLock({ wakeLock: { request } }, doc, win, t.onState);

    // A return to the front sends a second request, which answers first.
    doc.show();
    await settle();
    expect(request).toHaveBeenCalledTimes(2);
    expect(t.states).toEqual(["on"]);

    // The late refusal of the first request gives no notice.
    answer!.fail();
    await settle();
    expect(t.states).toEqual(["on"]);
    expect(fresh.release).not.toHaveBeenCalled();
  });

  it("gives a lock of an earlier request that answers late back to the phone", async () => {
    let ok: (() => void) | null = null;
    const late = { released: false, release: vi.fn(async () => {}), addEventListener: () => {} };
    const fresh = { released: false, release: vi.fn(async () => {}), addEventListener: () => {} };
    const request = vi
      .fn()
      .mockImplementationOnce(() => new Promise((resolve) => (ok = () => resolve(late))))
      .mockImplementation(async () => fresh);
    const { doc, win } = fakes();
    const t = track();
    holdWakeLock({ wakeLock: { request } }, doc, win, t.onState);
    win.pageshow();
    await settle();
    ok!();
    await settle();
    expect(late.release).toHaveBeenCalled();
    expect(fresh.release).not.toHaveBeenCalled();
    expect(t.states).toEqual(["on"]);
  });

  it("gets the lock at a focus and a pageshow event after a refusal at the return", async () => {
    const { doc, win, nav, request, state } = fakes();
    const t = track();
    holdWakeLock(nav, doc, win, t.onState);
    await settle();

    // The phone refuses the request of the return to the front.
    state.refuse = true;
    doc.hide();
    doc.show();
    await settle();
    expect(t.states.at(-1)).toBe("off");
    expect(t.errors.at(-1)).toBe("NotAllowedError");

    state.refuse = false;
    win.focus();
    await settle();
    expect(t.states.at(-1)).toBe("on");
    expect(request).toHaveBeenCalledTimes(3);

    doc.hide();
    win.pageshow(); // A hidden app makes no request.
    await settle();
    expect(request).toHaveBeenCalledTimes(3);
    doc.visibilityState = "visible";
    win.pageshow();
    await settle();
    expect(request).toHaveBeenCalledTimes(4);
    expect(t.states.at(-1)).toBe("on");
  });

  it("gives off with the error name when the phone refuses the lock", async () => {
    const { doc, win, nav } = fakes(true);
    const t = track();
    holdWakeLock(nav, doc, win, t.onState);
    await settle();
    expect(t.states).toEqual(["off"]);
    expect(t.errors).toEqual(["NotAllowedError"]);
  });

  it("gives off when the phone has no Screen Wake Lock API", () => {
    const { doc, win } = fakes();
    const t = track();
    holdWakeLock({}, doc, win, t.onState);
    expect(t.states).toEqual(["off"]);
    expect(t.errors).toEqual(["unsupported"]);
  });

  it("releases a lock that arrives after the stop", async () => {
    const { doc, win, nav, sentinels } = fakes();
    const onState = vi.fn();
    const stop = holdWakeLock(nav, doc, win, onState);
    stop();
    await settle();
    expect(sentinels[0].release).toHaveBeenCalled();
    expect(onState).not.toHaveBeenCalled();
  });
});
