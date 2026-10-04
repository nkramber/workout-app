import { create, toJson } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import { GetPlanResponseSchema, PlanSchema } from "../gen/workoutapp/v1/plan_service_pb";
import { nextState, nextTargets, revisedAt, waitsForSync } from "./revision";

const press = (reps: number, reason: string, reasonSource: string) => ({
  exerciseId: "chest_press",
  name: "Chest press",
  workingSets: [
    { reps, loadTenthLb: 250, rirTarget: 3 },
    { reps, loadTenthLb: 250, rirTarget: 3 },
  ],
  reason,
  source: "rules",
  reasonSource,
});

// A plan of two sessions after the revision of workout w1. Each session
// holds the chest press with its new target (D-290).
const plan = create(PlanSchema, {
  createdAt: "2026-10-01T08:00:00Z",
  sessions: [
    {
      title: "Session 1",
      exercises: [
        press(8, "Set 1: 12 reps at 25 lb, 1 in reserve.", "luna"),
        { exerciseId: "seated_row", name: "Seated row", workingSets: [{ reps: 12, loadTenthLb: 400, rirTarget: 2 }], reason: "You skipped this exercise. The target stays the same.", reasonSource: "rules" },
      ],
    },
    { title: "Session 2", exercises: [press(8, "Set 1: 12 reps at 25 lb, 1 in reserve.", "luna")] },
  ],
  lastRevision: { workoutId: "w1", revisedAt: "2026-10-04T09:00:00Z", exerciseIds: ["chest_press", "seated_row", "leg_press"] },
});

describe("nextTargets", () => {
  it("gives each revised exercise once, with its sets and its reason", () => {
    expect(nextTargets(plan)).toEqual([
      {
        exerciseId: "chest_press",
        name: "Chest press",
        sets: ["8 reps at 25 lb, 3 reps in reserve", "8 reps at 25 lb, 3 reps in reserve"],
        reason: "Set 1: 12 reps at 25 lb, 1 in reserve.",
        fromLuna: true,
      },
      {
        exerciseId: "seated_row",
        name: "Seated row",
        sets: ["12 reps at 40 lb, 2 reps in reserve"],
        reason: "You skipped this exercise. The target stays the same.",
        fromLuna: false,
      },
    ]);
  });

  it("names a calibration set", () => {
    const p = create(PlanSchema, {
      sessions: [{ exercises: [{ exerciseId: "leg_press", name: "Leg press", calibrationSets: [{ reps: 8, loadTenthLb: 500 }], workingSets: [{ reps: 8, loadTenthLb: 500, rirTarget: 3 }] }] }],
      lastRevision: { workoutId: "w1", exerciseIds: ["leg_press"] },
    });
    expect(nextTargets(p)[0].sets).toEqual(["Calibration: 8 reps at 50 lb", "8 reps at 50 lb, 3 reps in reserve"]);
  });
});

describe("nextState", () => {
  it("shows the targets when the plan copy holds the revision of the workout", () => {
    expect(nextState(plan, "w1", true)).toMatchObject({ kind: "revised" });
  });

  it("waits for the sync, then tells that the plan has no revision of the workout", () => {
    expect(nextState(plan, "w2", true)).toEqual({ kind: "waiting" });
    expect(nextState(undefined, "w2", true)).toEqual({ kind: "waiting" });
    expect(nextState(plan, "w2", false)).toEqual({ kind: "none" });
  });
});

describe("waitsForSync", () => {
  it("reads the header, the sets, and the cardio logs of the workout alone", () => {
    expect(waitsForSync({ entity: "workout", entityId: "w1", payload: {} }, "w1")).toBe(true);
    expect(waitsForSync({ entity: "set", entityId: "s1", payload: { workoutId: "w1" } }, "w1")).toBe(true);
    expect(waitsForSync({ entity: "cardio", entityId: "c1", payload: { workoutId: "w1" } }, "w1")).toBe(true);
    expect(waitsForSync({ entity: "set", entityId: "s1", payload: { workoutId: "w2" } }, "w1")).toBe(false);
    expect(waitsForSync({ entity: "machine", entityId: "w1", payload: {} }, "w1")).toBe(false);
    expect(waitsForSync({ entity: "set", entityId: "s1", payload: null }, "w1")).toBe(false);
  });
});

describe("revisedAt", () => {
  it("reads the time of the last revision of a plan copy", () => {
    expect(revisedAt(toJson(GetPlanResponseSchema, create(GetPlanResponseSchema, { plan })))).toBe("2026-10-04T09:00:00Z");
    expect(revisedAt(toJson(GetPlanResponseSchema, create(GetPlanResponseSchema, { plan: create(PlanSchema, {}) })))).toBe("");
    expect(revisedAt(undefined)).toBe("");
    expect(revisedAt(null)).toBe("");
  });
});
