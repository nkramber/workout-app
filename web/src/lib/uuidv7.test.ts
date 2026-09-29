import { describe, expect, it } from "vitest";

import { uuidv7 } from "./uuidv7";

const form = /^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/;

describe("uuidv7", () => {
  it("has the form, the version, and the variant of RFC 9562", () => {
    for (let i = 0; i < 100; i++) expect(uuidv7()).toMatch(form);
  });

  it("puts the Unix time in milliseconds in the first 48 bits", () => {
    const now = Date.UTC(2026, 8, 29, 12, 0, 0, 123);
    const id = uuidv7(now, (b) => b.fill(0xff));
    expect(parseInt(id.replace(/-/g, "").slice(0, 12), 16)).toBe(now);
    expect(id).toMatch(form);
  });

  it("keeps the version and the variant when the random bits are all zero", () => {
    expect(uuidv7(0, (b) => b.fill(0))).toBe("00000000-0000-7000-8000-000000000000");
  });

  it("sorts the ids of two times in the order of the times", () => {
    const a = uuidv7(1_000);
    const b = uuidv7(2_000);
    expect(a < b).toBe(true);
  });
});
