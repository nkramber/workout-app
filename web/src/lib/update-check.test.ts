import { describe, expect, it, vi } from "vitest";

import { bannerShown, checkForUpdate, updateAllowed } from "./update-check";

function registration(installing: object | null = null) {
  return { installing: installing as ServiceWorker | null, update: vi.fn(async () => undefined) };
}

function deps(online: boolean, status: number | Error) {
  return {
    online: () => online,
    fetch: vi.fn(async () => {
      if (status instanceof Error) throw status;
      return new Response("", { status });
    }),
  };
}

describe("checkForUpdate", () => {
  it("reads the worker file with no cache, then asks for the update", async () => {
    const reg = registration();
    const d = deps(true, 200);
    expect(await checkForUpdate(reg, "/sw.js", d)).toBe(true);
    expect(d.fetch).toHaveBeenCalledWith("/sw.js", { cache: "no-store", headers: { "cache-control": "no-cache" } });
    expect(reg.update).toHaveBeenCalledTimes(1);
  });

  it("makes no check when the phone is offline", async () => {
    const reg = registration();
    const d = deps(false, 200);
    expect(await checkForUpdate(reg, "/sw.js", d)).toBe(false);
    expect(d.fetch).not.toHaveBeenCalled();
    expect(reg.update).not.toHaveBeenCalled();
  });

  it("makes no check while a worker installs", async () => {
    const reg = registration({});
    expect(await checkForUpdate(reg, "/sw.js", deps(true, 200))).toBe(false);
    expect(reg.update).not.toHaveBeenCalled();
  });

  it("makes no update when the server fails or the fetch throws", async () => {
    for (const status of [404, 503, new TypeError("network")]) {
      const reg = registration();
      expect(await checkForUpdate(reg, "/sw.js", deps(true, status))).toBe(false);
      expect(reg.update).not.toHaveBeenCalled();
    }
  });
});

describe("updateAllowed", () => {
  it("applies no update during a workout", () => {
    expect(updateAllowed(true)).toBe(false);
    expect(updateAllowed(false)).toBe(true);
  });
});

describe("bannerShown", () => {
  it("shows the banner after the workout alone (D-323)", () => {
    expect(bannerShown(true, true)).toBe(false);
    expect(bannerShown(true, false)).toBe(true);
    expect(bannerShown(false, false)).toBe(false);
  });
});
