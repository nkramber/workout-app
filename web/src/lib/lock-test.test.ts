import { describe, expect, it, vi } from "vitest";

import { logLine, probeVideo, probeWakeLock, type ProbeVideo } from "./lock-test";

function page() {
  const doc = Object.assign(new EventTarget(), { visibilityState: "visible" });
  const win = new EventTarget();
  const back = () => {
    doc.visibilityState = "hidden";
    doc.dispatchEvent(new Event("visibilitychange"));
    doc.visibilityState = "visible";
    doc.dispatchEvent(new Event("visibilitychange"));
  };
  return { doc, win, back };
}

function sentinel() {
  const t = new EventTarget();
  const s = {
    released: false,
    release: async () => {
      s.released = true;
      t.dispatchEvent(new Event("release"));
    },
    addEventListener: (type: "release", fn: () => void) => t.addEventListener(type, fn),
  };
  return s;
}

const flush = () => new Promise((r) => setTimeout(r, 0));

describe("probeWakeLock", () => {
  it("asks for the lock at the start and again at a return with no tap, and logs a refusal", async () => {
    const { doc, win, back } = page();
    const first = sentinel();
    const request = vi
      .fn()
      .mockResolvedValueOnce(first)
      .mockRejectedValueOnce(Object.assign(new Error("no gesture"), { name: "NotAllowedError" }));
    const log = vi.fn();
    const stop = probeWakeLock({ wakeLock: { request } }, doc, win, log);
    await flush();
    expect(log).toHaveBeenLastCalledWith("start: the lock is on.");

    await first.release();
    expect(log).toHaveBeenLastCalledWith("The phone released the lock.");
    back();
    await flush();
    expect(log).toHaveBeenLastCalledWith("return: the phone refused the lock (NotAllowedError).");
    expect(request).toHaveBeenCalledTimes(2);

    stop();
    win.dispatchEvent(new Event("focus"));
    await flush();
    expect(request).toHaveBeenCalledTimes(2);
  });

  it("says when the phone has no Screen Wake Lock API", () => {
    const { doc, win } = page();
    const log = vi.fn();
    probeWakeLock({}, doc, win, log);
    expect(log).toHaveBeenCalledWith("No Screen Wake Lock API.");
  });
});

describe("probeVideo", () => {
  it("plays the video, and plays it again at a return with no tap when it stopped", async () => {
    const { doc, win, back } = page();
    const t = new EventTarget();
    const video: ProbeVideo = {
      paused: true,
      play: vi.fn(async () => {
        video.paused = false;
      }),
      pause: vi.fn(() => {
        video.paused = true;
        t.dispatchEvent(new Event("pause"));
      }),
      addEventListener: (type, fn) => t.addEventListener(type, fn),
      removeEventListener: (type, fn) => t.removeEventListener(type, fn),
    };
    const log = vi.fn();
    const stop = probeVideo(video, doc, win, log);
    await flush();
    expect(log).toHaveBeenLastCalledWith("start: the video plays again.");

    win.dispatchEvent(new Event("pageshow"));
    await flush();
    expect(log).toHaveBeenLastCalledWith("pageshow: the video plays.");

    video.pause();
    expect(log).toHaveBeenLastCalledWith("The video stopped.");
    vi.mocked(video.play).mockRejectedValueOnce(Object.assign(new Error("no gesture"), { name: "NotAllowedError" }));
    back();
    await flush();
    expect(log).toHaveBeenLastCalledWith("return: the phone refused to play the video (NotAllowedError).");

    stop();
    expect(video.pause).toHaveBeenCalledTimes(2);
    expect(log).not.toHaveBeenLastCalledWith("The video stopped.");
  });
});

describe("logLine", () => {
  it("gives the time since the start", () => {
    expect(logLine(1_000, 66_500, "The video plays.")).toBe("1:05 The video plays.");
    expect(logLine(1_000, 500, "x")).toBe("0:00 x");
  });
});
