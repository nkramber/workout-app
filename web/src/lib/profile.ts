import type { MessageInitShape } from "@bufbuild/protobuf";

import type { GetCatalogResponse } from "../gen/workoutapp/v1/inventory_service_pb";
import type { GoalTemplate, Profile, ProfileSchema } from "../gen/workoutapp/v1/profile_service_pb";
import type { Result } from "./inventory";

// The logic of the onboarding screen (work area 5.1). The server reads
// each value again, so these checks only give the owner an early message.
// The bounds copy go/internal/profile (D-211, D-215).

export const MIN_AGE = 18;
export const MAX_AGE = 90;
export const MIN_HEIGHT_IN = 48;
export const MAX_HEIGHT_IN = 96;
export const MIN_WEIGHT_LB = 80;
export const MAX_WEIGHT_LB = 500;
export const TRAINING_DAYS: readonly number[] = [2, 3, 4];
// MAX_TEXT_CHARS is the length limit of the free text and of the injury
// text, in characters.
export const MAX_TEXT_CHARS = 500;

// HEIGHT_FEET gives the feet of the height selection. 4 ft 0 in is 48 in,
// and 8 ft 0 in is 96 in.
export const HEIGHT_FEET: readonly number[] = [4, 5, 6, 7, 8];
export const HEIGHT_INCHES: readonly number[] = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11];

// INJURY_WARNING shows after the owner selects an injured area (D-35,
// D-222). It stays inside the fitness boundary of D-36.
export const INJURY_WARNING =
  "Your plan avoids exercises that load the areas you selected. This app gives fitness guidance only, and it does not diagnose or treat an injury.";

// Draft is the state of the onboarding form. A number field holds the
// text of the owner until the save, so the form can show a partial entry.
export type Draft = {
  experience: string;
  goalTemplate: string;
  muscleGroups: string[];
  freeText: string;
  injuredAreas: string[];
  injuryText: string;
  age: string;
  heightFeet: string;
  heightInches: string;
  weight: string;
  cardio: string[];
  // trainingDays is 0 until the owner selects a count.
  trainingDays: number;
};

// draftOf gives the form state of a stored profile, or an empty form when
// the owner saved no profile.
export function draftOf(p: Profile | undefined): Draft {
  const number = (n: number | undefined) => (n ? String(n) : "");
  const h = p?.heightIn ?? 0;
  return {
    experience: p?.experience ?? "",
    goalTemplate: p?.goalTemplate ?? "",
    muscleGroups: [...(p?.muscleGroups ?? [])],
    freeText: p?.freeText ?? "",
    injuredAreas: [...(p?.injuredAreas ?? [])],
    injuryText: p?.injuryText ?? "",
    age: number(p?.ageYears),
    heightFeet: h ? String(Math.floor(h / 12)) : "",
    heightInches: h ? String(h % 12) : "",
    weight: number(p?.weightLb),
    cardio: [...(p?.cardioExerciseIds ?? [])],
    trainingDays: p?.trainingDays ?? 0,
  };
}

// toggle adds the id to the selection or removes it. The result keeps the
// order of the fixed list, as the server stores it.
export function toggle(selected: readonly string[], id: string, order: readonly string[]): string[] {
  const next = new Set(selected);
  if (next.has(id)) next.delete(id);
  else next.add(id);
  return order.filter((v) => next.has(v));
}

// applyTemplate selects the template and the groups that it selects
// (D-210, D-220). The owner can change the groups after.
export function applyTemplate(d: Draft, t: GoalTemplate): Draft {
  return { ...d, goalTemplate: t.id, muscleGroups: [...t.muscleGroups] };
}

// cardioExercises gives the cardio exercises of the catalog, in catalog
// order. The cardio preference names them (D-217).
export function cardioExercises(catalog: GetCatalogResponse) {
  return catalog.exercises.filter((e) => e.region === "cardio");
}

const chars = (s: string) => [...s].length;

function wholeNumber(text: string): number | null {
  const t = text.trim();
  return /^\d{1,4}$/.test(t) ? Number(t) : null;
}

// checkDraft reads each field of the form, in the order of the screen,
// and gives the profile to save, or the message of the first field that
// fails.
export function checkDraft(d: Draft): Result<MessageInitShape<typeof ProfileSchema>> {
  const fail = (error: string) => ({ ok: false as const, error });
  if (!d.experience) return fail("Select your experience.");
  if (!TRAINING_DAYS.includes(d.trainingDays)) return fail("Select the training days in each week.");
  if (!d.goalTemplate) return fail("Select a goal.");
  if (d.muscleGroups.length === 0) return fail("Select one muscle group or more.");
  const freeText = d.freeText.trim();
  if (chars(freeText) > MAX_TEXT_CHARS) return fail(`The text for your plan has ${MAX_TEXT_CHARS} characters or fewer.`);
  const injuryText = d.injuryText.trim();
  if (chars(injuryText) > MAX_TEXT_CHARS) return fail(`The injury text has ${MAX_TEXT_CHARS} characters or fewer.`);
  const age = wholeNumber(d.age);
  if (age === null || age < MIN_AGE || age > MAX_AGE) return fail(`Enter an age from ${MIN_AGE} to ${MAX_AGE} years.`);
  const feet = wholeNumber(d.heightFeet);
  const inches = wholeNumber(d.heightInches);
  if (feet === null || inches === null) return fail("Select your height in feet and inches.");
  const height = feet * 12 + inches;
  if (height < MIN_HEIGHT_IN || height > MAX_HEIGHT_IN) return fail("Select a height from 4 ft 0 in to 8 ft 0 in.");
  const weight = wholeNumber(d.weight);
  if (weight === null || weight < MIN_WEIGHT_LB || weight > MAX_WEIGHT_LB) {
    return fail(`Enter a weight from ${MIN_WEIGHT_LB} to ${MAX_WEIGHT_LB} lb, as a whole number.`);
  }
  return {
    ok: true,
    value: {
      experience: d.experience,
      goalTemplate: d.goalTemplate,
      muscleGroups: d.muscleGroups,
      freeText,
      injuredAreas: d.injuredAreas,
      injuryText,
      ageYears: age,
      heightIn: height,
      weightLb: weight,
      cardioExerciseIds: d.cardio,
      trainingDays: d.trainingDays,
    },
  };
}
