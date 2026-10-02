import { create } from "@bufbuild/protobuf";
import { describe, expect, it } from "vitest";

import { CatalogExerciseSchema, GetCatalogResponseSchema } from "../gen/workoutapp/v1/inventory_service_pb";
import { GoalTemplateSchema, ProfileSchema } from "../gen/workoutapp/v1/profile_service_pb";
import { applyTemplate, cardioExercises, checkDraft, draftOf, INJURY_WARNING, toggle, type Draft } from "./profile";

const GROUPS = ["chest", "back", "shoulders", "biceps", "triceps", "quadriceps", "hamstrings", "glutes", "calves", "core"];

// A complete form with synthetic values.
const full = (change: Partial<Draft> = {}): Draft => ({
  experience: "intermediate",
  goalTemplate: "strength",
  muscleGroups: ["chest", "back"],
  freeText: "  more back work  ",
  injuredAreas: ["knee"],
  injuryText: "",
  age: "40",
  heightFeet: "5",
  heightInches: "10",
  weight: "180",
  cardio: ["treadmill_walk"],
  trainingDays: 3,
  ...change,
});

const error = (d: Draft) => {
  const r = checkDraft(d);
  return r.ok ? "" : r.error;
};

describe("draftOf", () => {
  it("gives an empty form with no profile", () => {
    const d = draftOf(undefined);
    expect(d.experience).toBe("");
    expect(d.muscleGroups).toEqual([]);
    expect(d.age).toBe("");
    expect(d.heightFeet).toBe("");
    expect(d.trainingDays).toBe(0);
  });

  it("gives the form of a stored profile, with the height in feet and inches", () => {
    const p = create(ProfileSchema, {
      experience: "advanced",
      goalTemplate: "general_fitness",
      muscleGroups: ["chest"],
      ageYears: 52,
      heightIn: 71,
      weightLb: 200,
      trainingDays: 4,
    });
    const d = draftOf(p);
    expect(d).toMatchObject({ experience: "advanced", age: "52", heightFeet: "5", heightInches: "11", weight: "200", trainingDays: 4 });
    d.muscleGroups.push("back");
    expect(p.muscleGroups).toEqual(["chest"]);
  });

  it("gives the round trip of each height of the bound (D-215)", () => {
    for (const heightIn of [48, 60, 72, 95, 96]) {
      const r = checkDraft({ ...full(), ...pick(draftOf(create(ProfileSchema, { heightIn })), "heightFeet", "heightInches") });
      expect(r.ok && r.value.heightIn).toBe(heightIn);
    }
  });
});

function pick<T, K extends keyof T>(o: T, ...keys: K[]): Pick<T, K> {
  return Object.fromEntries(keys.map((k) => [k, o[k]])) as Pick<T, K>;
}

describe("toggle", () => {
  it("adds and removes an id, in the order of the fixed list", () => {
    expect(toggle(["core"], "chest", GROUPS)).toEqual(["chest", "core"]);
    expect(toggle(["chest", "core"], "chest", GROUPS)).toEqual(["core"]);
  });
});

describe("applyTemplate", () => {
  it("selects the template and its groups, and keeps the other fields (D-210, D-220)", () => {
    const strength = create(GoalTemplateSchema, {
      id: "strength",
      name: "Strength",
      muscleGroups: ["chest", "back", "shoulders", "quadriceps", "hamstrings", "glutes"],
    });
    const d = applyTemplate(full({ goalTemplate: "general_fitness", muscleGroups: ["core"] }), strength);
    expect(d.goalTemplate).toBe("strength");
    expect(d.muscleGroups).toEqual(["chest", "back", "shoulders", "quadriceps", "hamstrings", "glutes"]);
    expect(d.injuredAreas).toEqual(["knee"]);
    d.muscleGroups.pop();
    expect(strength.muscleGroups).toHaveLength(6);
  });
});

describe("cardioExercises", () => {
  it("gives the cardio exercises alone, in catalog order (D-217)", () => {
    const catalog = create(GetCatalogResponseSchema, {
      exercises: [
        create(CatalogExerciseSchema, { id: "rower", region: "cardio" }),
        create(CatalogExerciseSchema, { id: "leg_press", region: "lower_push" }),
        create(CatalogExerciseSchema, { id: "bike", region: "cardio" }),
      ],
    });
    expect(cardioExercises(catalog).map((e) => e.id)).toEqual(["rower", "bike"]);
  });
});

describe("checkDraft", () => {
  it("gives the profile of a full form, with each text trimmed", () => {
    const r = checkDraft(full());
    expect(r).toEqual({
      ok: true,
      value: {
        experience: "intermediate",
        goalTemplate: "strength",
        muscleGroups: ["chest", "back"],
        freeText: "more back work",
        injuredAreas: ["knee"],
        injuryText: "",
        ageYears: 40,
        heightIn: 70,
        weightLb: 180,
        cardioExerciseIds: ["treadmill_walk"],
        trainingDays: 3,
      },
    });
  });

  it("accepts no injured area and no cardio (D-217)", () => {
    expect(checkDraft(full({ injuredAreas: [], cardio: [] })).ok).toBe(true);
  });

  it("asks for each selection", () => {
    expect(error(full({ experience: "" }))).toBe("Select your experience.");
    expect(error(full({ trainingDays: 0 }))).toBe("Select the training days in each week.");
    expect(error(full({ trainingDays: 5 }))).toBe("Select the training days in each week.");
    expect(error(full({ goalTemplate: "" }))).toBe("Select a goal.");
    expect(error(full({ muscleGroups: [] }))).toBe("Select one muscle group or more.");
    expect(error(full({ heightFeet: "" }))).toBe("Select your height in feet and inches.");
  });

  it("refuses a value outside each bound of the server (D-211, D-215)", () => {
    for (const age of ["17", "91", "", "40.5", "forty", "-20"]) expect(error(full({ age }))).toMatch(/^Enter an age from 18 to 90/);
    for (const age of ["18", "90", " 33 "]) expect(error(full({ age }))).toBe("");
    for (const weight of ["79", "501", "180.5"]) expect(error(full({ weight }))).toMatch(/^Enter a weight from 80 to 500 lb/);
    for (const weight of ["80", "500"]) expect(error(full({ weight }))).toBe("");
    expect(error(full({ heightFeet: "8", heightInches: "1" }))).toMatch(/^Select a height from/);
    expect(error(full({ heightFeet: "8", heightInches: "0" }))).toBe("");
    expect(error(full({ heightFeet: "4", heightInches: "0" }))).toBe("");
  });

  it("counts the characters of each text after the trim, as the server does (D-215)", () => {
    expect(error(full({ freeText: "a".repeat(500) }))).toBe("");
    expect(error(full({ freeText: ` ${"a".repeat(500)} ` }))).toBe("");
    expect(error(full({ freeText: "a".repeat(501) }))).toMatch(/text for your plan has 500 characters/);
    // One emoji is one character and two UTF-16 units.
    expect(error(full({ injuryText: "\u{1F4AA}".repeat(500) }))).toBe("");
    expect(error(full({ injuryText: "\u{1F4AA}".repeat(501) }))).toMatch(/injury text has 500 characters/);
  });
});

describe("the injury warning", () => {
  it("states the avoidance and the fitness boundary, and no diagnosis (D-36, D-222)", () => {
    expect(INJURY_WARNING).toBe(
      "Your plan avoids exercises that load the areas you selected. This app gives fitness guidance only, and it does not diagnose or treat an injury.",
    );
  });
});
