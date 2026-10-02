import { create } from "@bufbuild/protobuf";
import { Code, ConnectError } from "@connectrpc/connect";
import { describe, expect, it } from "vitest";

import {
  CatalogExerciseSchema,
  CatalogMachineSchema,
  DumbbellSetSchema,
  InventoryMachineSchema,
} from "../gen/workoutapp/v1/inventory_service_pb";
import { changeErrorText, isNoConnection, loadErrorText } from "./errors";
import {
  addWeight,
  available,
  checkDumbbells,
  checkEstimate,
  checkNote,
  formatPounds,
  groupByKind,
  guessStep,
  parsePounds,
  rangeWeights,
  removeWeight,
  searchCatalog,
  sortByName,
} from "./inventory";

// A synthetic catalog in the form of GetCatalog, with the kinds out of
// the order of D-202 and the names out of A to Z order, so the tests prove
// both orders.
const machines = [
  ["treadmill", "Treadmill", "cardio"],
  ["leg_press", "Leg press", "machine"],
  ["dumbbells", "Dumbbells with an adjustable bench", "dumbbell"],
  ["cable_station", "Cable station", "cable"],
  ["seated_leg_curl", "Seated leg curl", "machine"],
  ["lying_leg_curl", "Lying leg curl", "machine"],
].map(([id, name, kind]) => create(CatalogMachineSchema, { id, name, kind }));

const exercises = [
  ["leg_press", "Leg press", "leg_press"],
  ["lat_pulldown", "Lat pulldown", "cable_station"],
  ["db_goblet_squat", "Dumbbell goblet squat", "dumbbells"],
  ["seated_leg_curl", "Seated leg curl", "seated_leg_curl"],
  ["lying_leg_curl", "Lying leg curl", "lying_leg_curl"],
  ["treadmill", "Treadmill", "treadmill"],
].map(([id, name, machineId]) => create(CatalogExerciseSchema, { id, name, machineId }));

const ids = (list: { id: string }[]) => list.map((m) => m.id);

describe("groupByKind", () => {
  it("orders the groups by kind (D-202), and each group A to Z by name (D-205)", () => {
    const groups = groupByKind(machines);
    expect(groups.map((g) => g.title)).toEqual(["Machines", "Cable station", "Dumbbells", "Cardio"]);
    expect(ids(groups[0].machines)).toEqual(["leg_press", "lying_leg_curl", "seated_leg_curl"]);
    expect(groups.flatMap((g) => g.machines)).toHaveLength(machines.length);
  });

  it("leaves out an empty group", () => {
    expect(groupByKind(machines.filter((m) => m.kind !== "cable")).map((g) => g.kind)).toEqual([
      "machine",
      "dumbbell",
      "cardio",
    ]);
  });
});

describe("sortByName", () => {
  const names = new Map(machines.map((m) => [m.id, m.name]));
  const entry = (machineId: string) => create(InventoryMachineSchema, { machineId });

  it("orders the machines of the inventory A to Z by the catalog name, in any case (D-205)", () => {
    const list = ["treadmill", "seated_leg_curl", "cable_station", "leg_press"].map(entry);
    const lower = new Map([...names, ["cable_station", "cable station"]]);
    expect(sortByName(list, lower).map((m) => m.machineId)).toEqual(["cable_station", "leg_press", "seated_leg_curl", "treadmill"]);
    expect(list.map((m) => m.machineId)[0]).toBe("treadmill");
  });

  it("sorts a machine that the catalog does not name by its id", () => {
    const list = ["treadmill", "abc_unknown"].map(entry);
    expect(sortByName(list, names).map((m) => m.machineId)).toEqual(["abc_unknown", "treadmill"]);
  });
});

describe("searchCatalog", () => {
  it("finds a machine by a part of each word of its name, in any case", () => {
    expect(ids(searchCatalog(machines, exercises, "LEG cur"))).toEqual(["lying_leg_curl", "seated_leg_curl"]);
    expect(ids(searchCatalog(machines, exercises, "  leg   PRESS "))).toEqual(["leg_press"]);
  });

  it("finds a machine by the name of one of its exercises", () => {
    expect(ids(searchCatalog(machines, exercises, "lat pull"))).toEqual(["cable_station"]);
    expect(ids(searchCatalog(machines, exercises, "goblet"))).toEqual(["dumbbells"]);
  });

  it("finds a plural of a name", () => {
    expect(ids(searchCatalog(machines, exercises, "leg curls"))).toEqual(["lying_leg_curl", "seated_leg_curl"]);
  });

  it("keeps the order of the catalog list (D-202, D-205)", () => {
    expect(ids(searchCatalog(machines, exercises, "l"))).toEqual([
      "leg_press",
      "lying_leg_curl",
      "seated_leg_curl",
      "cable_station",
    ]);
  });

  it("finds nothing for a text with no match, or with no words", () => {
    expect(searchCatalog(machines, exercises, "smith machine")).toEqual([]);
    expect(searchCatalog(machines, exercises, " - ")).toEqual([]);
    expect(searchCatalog(machines, exercises, "")).toEqual([]);
  });
});

describe("parsePounds and formatPounds", () => {
  it("reads whole pounds and one decimal place as tenths", () => {
    expect(parsePounds("10")).toBe(100);
    expect(parsePounds(" 12.5 ")).toBe(125);
    expect(parsePounds("0.5")).toBe(5);
  });

  it("refuses an empty text, 0, a sign, two decimal places, and a word", () => {
    for (const t of ["", "0", "0.0", "-5", "+5", "12.25", "1e3", "ten", "5 lb", ".5"]) {
      expect(parsePounds(t), t).toBeNull();
    }
  });

  it("writes tenths as pounds", () => {
    expect(formatPounds(100)).toBe("10");
    expect(formatPounds(125)).toBe("12.5");
    expect(formatPounds(5)).toBe("0.5");
  });
});

describe("rangeWeights", () => {
  it("makes the list from the lightest weight, the heaviest weight, and the step (D-195)", () => {
    expect(rangeWeights(100, 500, 100)).toEqual({ ok: true, value: [100, 200, 300, 400, 500] });
    expect(rangeWeights(100, 100, 50)).toEqual({ ok: true, value: [100] });
  });

  it("ends with the heaviest weight when the step does not reach it", () => {
    expect(rangeWeights(100, 250, 100)).toEqual({ ok: true, value: [100, 200, 250] });
  });

  it("refuses a missing value, a reversed range, and the bounds of D-199", () => {
    expect(rangeWeights(null, 500, 100).ok).toBe(false);
    expect(rangeWeights(500, 100, 100).ok).toBe(false);
    expect(rangeWeights(100, 10_010, 100).ok).toBe(false);
    expect(rangeWeights(10, 2_000, 10).ok).toBe(true);
    expect(rangeWeights(10, 2_010, 10).ok).toBe(false);
  });
});

describe("addWeight and removeWeight", () => {
  it("adds a weight in order, and removes it", () => {
    const added = addWeight([100, 300], 200);
    expect(added).toEqual({ ok: true, value: [100, 200, 300] });
    expect(removeWeight([100, 200, 300], 200)).toEqual([100, 300]);
  });

  it("refuses a weight two times, a bad weight, and a full list", () => {
    expect(addWeight([100], 100).ok).toBe(false);
    expect(addWeight([100], null).ok).toBe(false);
    expect(addWeight([100], 10_001).ok).toBe(false);
    expect(addWeight(Array.from({ length: 200 }, (_, i) => i + 1), 1_000).ok).toBe(false);
  });
});

describe("checkDumbbells", () => {
  it("accepts a set whose step goes from the lightest to the heaviest dumbbell", () => {
    expect(checkDumbbells(50, 500, 50)).toEqual({ ok: true, value: { lightest: 50, heaviest: 500, step: 50 } });
  });

  it("refuses a missing value, a reversed set, a step that does not divide, and the bound of D-166", () => {
    expect(checkDumbbells(50, null, 50).ok).toBe(false);
    expect(checkDumbbells(500, 50, 50).ok).toBe(false);
    expect(checkDumbbells(50, 520, 50).ok).toBe(false);
    expect(checkDumbbells(50, 1_050, 50).ok).toBe(false);
  });
});

describe("available", () => {
  it("gives the weights of a stack, and each load of a dumbbell set", () => {
    expect(available(create(InventoryMachineSchema, { weightsTenthLb: [100, 200] }))).toEqual([100, 200]);
    const dumbbells = create(DumbbellSetSchema, { lightestTenthLb: 50, heaviestTenthLb: 150, stepTenthLb: 50 });
    expect(available(create(InventoryMachineSchema, { dumbbells }))).toEqual([50, 100, 150]);
  });
});

describe("checkEstimate", () => {
  it("reads an empty text as no estimate (D-192)", () => {
    expect(checkEstimate(" ", [100, 200])).toEqual({ ok: true, value: null });
  });

  it("accepts an estimate from the lightest to the heaviest weight (D-198)", () => {
    expect(checkEstimate("10", [100, 200])).toEqual({ ok: true, value: 100 });
    expect(checkEstimate("15", [100, 200])).toEqual({ ok: true, value: 150 });
    expect(checkEstimate("20", [100, 200])).toEqual({ ok: true, value: 200 });
  });

  it("refuses an estimate outside the weights, a bad text, and a machine with no weights", () => {
    expect(checkEstimate("9.5", [100, 200]).ok).toBe(false);
    expect(checkEstimate("20.5", [100, 200]).ok).toBe(false);
    expect(checkEstimate("x", [100, 200]).ok).toBe(false);
    expect(checkEstimate("10", []).ok).toBe(false);
  });
});

describe("checkNote", () => {
  it("trims the text, and keeps 1 to 200 characters (D-199)", () => {
    expect(checkNote("  Smith machine ")).toEqual({ ok: true, value: "Smith machine" });
    expect(checkNote("   ").ok).toBe(false);
    expect(checkNote("a".repeat(200)).ok).toBe(true);
    expect(checkNote("a".repeat(201)).ok).toBe(false);
  });
});

describe("guessStep", () => {
  it("gives the smallest difference of a list, or null", () => {
    expect(guessStep([100, 200, 250, 350])).toBe(50);
    expect(guessStep([100])).toBeNull();
  });
});

describe("the error text", () => {
  it("reads a failed fetch and an unavailable server as no connection (D-196)", () => {
    const failedFetch = new ConnectError("Failed to fetch", Code.Unknown, undefined, undefined, new TypeError("Failed to fetch"));
    expect(isNoConnection(failedFetch, true)).toBe(true);
    expect(isNoConnection(new ConnectError("x", Code.Unavailable), true)).toBe(true);
    expect(isNoConnection(new ConnectError("x", Code.Internal), false)).toBe(true);
    expect(isNoConnection(new ConnectError("x", Code.Unknown), true)).toBe(false);
    expect(changeErrorText(failedFetch, true)).toMatch(/^No connection\. The change is not saved\./);
  });

  it("names a server fault, and does not say that the API did not answer (D-206)", () => {
    const fault = new ConnectError("x", Code.Internal);
    expect(changeErrorText(fault, true)).toBe("The server failed. The change is not saved.");
    expect(loadErrorText(fault, true)).toBe("The server failed.");
    expect(changeErrorText(fault, false)).toMatch(/^No connection\./);
  });

  it("names a changed weight list for a refused confirmation (D-201)", () => {
    expect(changeErrorText(new ConnectError("x", Code.FailedPrecondition), true)).toMatch(/weights changed/);
  });

  it("names a refused account in a failed read, so the profile gate can show it (D-223)", () => {
    expect(loadErrorText(new ConnectError("x", Code.PermissionDenied), true)).toBe("This account is not on the allowlist.");
    expect(loadErrorText(new ConnectError("x", Code.Unauthenticated), true)).toBe("The API did not accept the sign-in.");
    expect(loadErrorText(new ConnectError("x", Code.DeadlineExceeded), true)).toBe("The API did not answer.");
  });
});
