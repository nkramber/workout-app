import "fake-indexeddb/auto";

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { GymRouteDB, OUTBOX_SCHEMA_VERSION, pendingOutbox, saveSetting } from "./db";
import * as ids from "./uuidv7";

let store: GymRouteDB;
let n = 0;

beforeEach(async () => {
  store = new GymRouteDB(`test-${n++}`);
  await store.open();
});

afterEach(async () => {
  vi.restoreAllMocks();
  await store.delete();
});

describe("saveSetting", () => {
  it("writes the change and its outbox entry", async () => {
    const now = new Date("2026-09-29T12:00:00.000Z");
    const entry = await saveSetting(store, "rest-seconds", 90, now);

    expect(await store.settings.get("rest-seconds")).toEqual({
      id: "rest-seconds",
      value: 90,
      version: 0,
      updatedAt: now.toISOString(),
    });
    expect(await pendingOutbox(store)).toEqual([entry]);
    expect(entry).toMatchObject({
      entity: "setting",
      entityId: "rest-seconds",
      baseVersion: 0,
      payload: 90,
      at: now.toISOString(),
      attempts: 0,
      schemaVersion: OUTBOX_SCHEMA_VERSION,
    });
    expect(entry.opId).toMatch(/^[0-9a-f-]{36}$/);
  });

  it("starts each change from the last server version that the phone knows", async () => {
    await store.settings.put({ id: "rest-seconds", value: 60, version: 4, updatedAt: "2026-09-28T00:00:00.000Z" });
    const entry = await saveSetting(store, "rest-seconds", 120);
    expect(entry.baseVersion).toBe(4);
    expect((await store.settings.get("rest-seconds"))?.version).toBe(4);
  });

  it("gives each change its own op id", async () => {
    const a = await saveSetting(store, "a", 1);
    const b = await saveSetting(store, "a", 2);
    expect(a.opId).not.toBe(b.opId);
    expect(await store.outbox.count()).toBe(2);
  });

  it("keeps neither write when the outbox write fails", async () => {
    vi.spyOn(ids, "uuidv7").mockReturnValue("0190a000-0000-7000-8000-000000000001");
    await saveSetting(store, "rest-seconds", 90);
    // The same op id again makes the outbox write fail, so the change of
    // the setting must roll back too.
    await expect(saveSetting(store, "rest-seconds", 120)).rejects.toThrow();
    expect((await store.settings.get("rest-seconds"))?.value).toBe(90);
    expect(await store.outbox.count()).toBe(1);
  });

  it("opens the store again after the browser closed it", async () => {
    store.close();
    const entry = await saveSetting(store, "rest-seconds", 90);
    expect(await pendingOutbox(store)).toEqual([entry]);
  });
});
