import { Code, ConnectError } from "@connectrpc/connect";
import { describe, expect, it } from "vitest";

import {
  CAP_REACHED,
  localDate,
  MAX_REASON_CHARS,
  NO_EXERCISE,
  NO_VALID_PLAN,
  planErrorText,
  progressText,
  reasonLength,
  restText,
  setText,
} from "./plan";
import { InventoryNotSyncedError } from "./sync";

const call = (attempt: number, previousStatus = "") => ({ step: "call", attempt, maxAttempts: 4, previousStatus });

describe("progressText", () => {
  it("gives the texts of D-239", () => {
    expect(progressText(call(1))).toBe("Asking Luna for a plan. Try 1 of 4.");
    expect(progressText({ ...call(0), step: "check" })).toBe("Checking each set and load with the safety rules.");
    expect(progressText({ ...call(0), step: "save" })).toBe("Saving the plan.");
  });

  it("names the cause of each retry", () => {
    const want: Record<string, string> = {
      malformed: "it was not in the right form",
      refusal: "Luna refused",
      incomplete: "it was cut off",
      timeout: "it took too long",
      error: "the call failed",
      new_status: "the call failed",
    };
    for (const [status, cause] of Object.entries(want)) {
      expect(progressText(call(2, status))).toBe(`The last answer of Luna was not valid (${cause}). Asking again. Try 2 of 4.`);
    }
  });
});

describe("planErrorText", () => {
  const err = (code: Code) => new ConnectError("plan: an error", code);

  it("gives the three errors of D-240", () => {
    expect(planErrorText(err(Code.Unavailable), { sawProgress: true, isOnline: true })).toBe(NO_VALID_PLAN);
    expect(planErrorText(err(Code.ResourceExhausted), { isOnline: true })).toBe(CAP_REACHED);
    expect(planErrorText(err(Code.FailedPrecondition), { isOnline: true })).toBe(NO_EXERCISE);
    expect(NO_VALID_PLAN).toBe("Luna gave no valid plan in 4 tries. Your plan did not change. Try again later.");
    expect(NO_EXERCISE).toBe("No confirmed machine gives an exercise that you can do. Your plan did not change. Confirm a machine, or change the injuries in your profile.");
  });

  it("starts the text of a failed exclusion with the exclusion", () => {
    expect(planErrorText(err(Code.ResourceExhausted), { exclude: true, isOnline: true })).toBe(
      `The exercise is not excluded. ${CAP_REACHED}`,
    );
  });

  it("gives no connection for UNAVAILABLE with no progress, and with no network", () => {
    const none = "No connection. Try again when the phone is online.";
    expect(planErrorText(err(Code.Unavailable), { isOnline: true })).toBe(none);
    expect(planErrorText(err(Code.Unavailable), { sawProgress: true, isOnline: false })).toBe(none);
    const fetchFailed = new ConnectError("fetch failed", Code.Unknown, undefined, undefined, new TypeError("Load failed"));
    expect(planErrorText(fetchFailed, { sawProgress: true, isOnline: true })).toBe(none);
  });

  it("gives a fixed text for each other code", () => {
    expect(planErrorText(err(Code.Aborted), { isOnline: true })).toMatch(/^The exclusions changed during the request\./);
    expect(planErrorText(err(Code.DeadlineExceeded), { isOnline: true })).toMatch(/^The request took too long\./);
    expect(planErrorText(err(Code.Internal), { isOnline: true })).toBe("The server failed. Your plan did not change.");
    expect(planErrorText(err(Code.PermissionDenied), { isOnline: true })).toBe("This account is not on the allowlist.");
  });
});

describe("formats", () => {
  it("gives a calibration set with no RIR, and a working set with its RIR", () => {
    expect(setText({ reps: 8, loadTenthLb: 125, rirTarget: 0 }, true)).toBe("8 reps at 12.5 lb");
    expect(setText({ reps: 10, loadTenthLb: 500, rirTarget: 2 }, false)).toBe("10 reps at 50 lb, 2 reps in reserve");
    expect(setText({ reps: 1, loadTenthLb: 500, rirTarget: 1 }, false)).toBe("1 rep at 50 lb, 1 rep in reserve");
  });

  it("gives the rest in seconds and minutes", () => {
    expect(restText(45)).toBe("45 s");
    expect(restText(60)).toBe("1 min");
    expect(restText(150)).toBe("2 min 30 s");
  });

  it("gives the local date", () => {
    expect(localDate(new Date(2026, 0, 5, 23, 30))).toBe("2026-01-05");
  });

  it("counts the characters of the trimmed reason, as the server does", () => {
    expect(reasonLength("  knee  ")).toBe(4);
    expect(reasonLength("💪".repeat(MAX_REASON_CHARS))).toBe(MAX_REASON_CHARS);
  });
});

describe("planErrorText of a change that waits", () => {
  it("says that the equipment changes did not reach the server (D-272)", () => {
    const text = "Your equipment changes did not reach the server. Your plan did not change. Try again when the line above says Synced.";
    expect(planErrorText(new InventoryNotSyncedError())).toBe(text);
    expect(planErrorText(new InventoryNotSyncedError(), { exclude: true })).toBe(`The exercise is not excluded. ${text}`);
  });
});
