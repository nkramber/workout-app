import "fake-indexeddb/auto";

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { GymRouteDB } from "./db";
import { PERSIST_KEY, megabytes, requestPersistenceOnce } from "./storage";

let store: GymRouteDB;
let n = 0;

beforeEach(async () => {
  store = new GymRouteDB(`storage-${n++}`);
  await store.open();
});

afterEach(async () => {
  await store.delete();
});

function fakeApi(granted: boolean) {
  return {
    persist: vi.fn(async () => granted),
    persisted: vi.fn(async () => granted),
    estimate: vi.fn(async () => ({ usage: 2_500_000, quota: 1_000_000_000 })),
  };
}

describe("requestPersistenceOnce", () => {
  it("asks on the first call and keeps the answer", async () => {
    const api = fakeApi(true);
    const now = new Date("2026-09-29T12:00:00.000Z");
    const got = await requestPersistenceOnce(store, api, now);
    expect(got).toEqual({ persisted: true, requestedAt: now.toISOString(), usage: 2_500_000, quota: 1_000_000_000 });
    expect(api.persist).toHaveBeenCalledTimes(1);
    expect((await store.meta.get(PERSIST_KEY))?.value).toEqual({ at: now.toISOString(), granted: true });
  });

  it("never asks again after the first call, and reads the state", async () => {
    const api = fakeApi(false);
    const first = await requestPersistenceOnce(store, api);
    const again = await requestPersistenceOnce(store, api);
    expect(api.persist).toHaveBeenCalledTimes(1);
    expect(api.persisted).toHaveBeenCalledTimes(1);
    expect(again.persisted).toBe(false);
    expect(again.requestedAt).toBe(first.requestedAt);
  });

  it("gives null values when the browser has no storage API", async () => {
    expect(await requestPersistenceOnce(store, undefined)).toEqual({ persisted: null, requestedAt: null, usage: null, quota: null });
  });

  it("keeps the answer when the estimate fails", async () => {
    const api = { ...fakeApi(true), estimate: vi.fn(async () => Promise.reject(new Error("no estimate"))) };
    const got = await requestPersistenceOnce(store, api);
    expect(got.persisted).toBe(true);
    expect(got.usage).toBeNull();
  });
});

describe("megabytes", () => {
  it("gives MB with one decimal", () => {
    expect(megabytes(2_500_000)).toBe("2.5 MB");
    expect(megabytes(null)).toBe("n/a");
  });
});
