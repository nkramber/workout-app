import type {
  CatalogExercise,
  CatalogMachine,
  DumbbellSet,
  InventoryMachine,
} from "../gen/workoutapp/v1/inventory_service_pb";

// The logic of the inventory screens (work area 4.1). Each load is a whole
// number of tenths of a pound, as in the contract, so 12.5 lb is 125
// (D-122). The server reads each value again, so these checks only give
// the owner an early message. The bounds copy go/internal/inventory and
// go/internal/domain (D-166, D-199).

// MAX_WEIGHT is the heaviest weight of a machine or of the cable station.
export const MAX_WEIGHT = 10_000;
// MAX_WEIGHTS is the length limit of the weight list of one machine.
export const MAX_WEIGHTS = 200;
// DUMBBELL_MAX is the heaviest dumbbell of a set.
export const DUMBBELL_MAX = 1_000;
// MAX_NOTE_CHARS is the length limit of the text of a note.
export const MAX_NOTE_CHARS = 200;

export type Kind = "machine" | "cable" | "dumbbell" | "cardio";

// KIND_ORDER is the order of the groups of the catalog list (D-202).
export const KIND_ORDER: readonly { kind: Kind; title: string }[] = [
  { kind: "machine", title: "Machines" },
  { kind: "cable", title: "Cable station" },
  { kind: "dumbbell", title: "Dumbbells" },
  { kind: "cardio", title: "Cardio" },
];

export type CatalogGroup = { kind: Kind; title: string; machines: CatalogMachine[] };

// compareNames orders two names A to Z, with no difference between upper
// and lower case (D-205).
export function compareNames(a: string, b: string): number {
  return a.localeCompare(b, "en", { sensitivity: "base", numeric: true });
}

function byName(a: { id: string; name: string }, b: { id: string; name: string }): number {
  return compareNames(a.name, b.name) || compareNames(a.id, b.id);
}

// groupByKind gives the groups of the catalog list, by kind (D-202), each
// A to Z by name (D-205). A group with no machine does not show.
export function groupByKind(machines: readonly CatalogMachine[]): CatalogGroup[] {
  return KIND_ORDER.map((g) => ({ ...g, machines: machines.filter((m) => m.kind === g.kind).sort(byName) })).filter(
    (g) => g.machines.length > 0,
  );
}

// sortByName gives the machines of the inventory A to Z by the catalog
// name (D-205). A machine that the catalog does not name sorts by its id.
export function sortByName(machines: readonly InventoryMachine[], names: ReadonlyMap<string, string>): InventoryMachine[] {
  const name = (m: InventoryMachine) => names.get(m.machineId) ?? m.machineId;
  return [...machines].sort((a, b) => byName({ id: a.machineId, name: name(a) }, { id: b.machineId, name: name(b) }));
}

function words(text: string): string[] {
  return text
    .toLowerCase()
    .split(/[^a-z0-9]+/)
    .filter((w) => w.length > 0);
}

// A word of the text matches a word of a name when it starts the word. A
// plural of the text, such as "curls", also matches "curl".
function wordMatches(query: string, word: string): boolean {
  if (word.startsWith(query)) return true;
  return query.length > 3 && query.endsWith("s") && word.startsWith(query.slice(0, -1));
}

function nameMatches(queryWords: string[], name: string): boolean {
  const nameWords = words(name);
  return queryWords.every((q) => nameWords.some((w) => wordMatches(q, w)));
}

// searchCatalog searches the names of the catalog for a text (D-191). A
// machine matches when each word of the text matches a word of its name,
// or of the name of one of its exercises. So "lat pull" finds the cable
// station. The matches keep the order of the catalog list (D-202, D-205). A text
// with no words matches nothing.
export function searchCatalog(
  machines: readonly CatalogMachine[],
  exercises: readonly CatalogExercise[],
  text: string,
): CatalogMachine[] {
  const queryWords = words(text);
  if (queryWords.length === 0) return [];
  const found = machines.filter(
    (m) =>
      nameMatches(queryWords, m.name) ||
      exercises.some((e) => e.machineId === m.id && nameMatches(queryWords, e.name)),
  );
  return groupByKind(found).flatMap((g) => g.machines);
}

// parsePounds reads a load in pounds, with one decimal place or none, as
// tenths of a pound. It gives null for an empty or bad text, and for 0.
export function parsePounds(text: string): number | null {
  const t = text.trim();
  const m = /^(\d{1,5})(?:\.(\d))?$/.exec(t);
  if (!m) return null;
  const tenths = Number(m[1]) * 10 + Number(m[2] ?? "0");
  return tenths > 0 ? tenths : null;
}

// formatPounds gives a load in tenths of a pound as text, such as "12.5".
export function formatPounds(tenths: number): string {
  return tenths % 10 === 0 ? String(tenths / 10) : (tenths / 10).toFixed(1);
}

export type Result<T> = { ok: true; value: T } | { ok: false; error: string };

// rangeWeights makes the weight list of a stack from the lightest weight,
// the heaviest weight, and the step (D-195). When the step does not reach
// the heaviest weight, the list ends with the heaviest weight, so the
// list holds each weight that the owner entered.
export function rangeWeights(lightest: number | null, heaviest: number | null, step: number | null): Result<number[]> {
  if (lightest === null || heaviest === null || step === null) {
    return { ok: false, error: "Enter the lightest weight, the heaviest weight, and the step, each more than 0." };
  }
  if (heaviest < lightest) return { ok: false, error: "The heaviest weight is less than the lightest weight." };
  if (heaviest > MAX_WEIGHT) return { ok: false, error: `The heaviest weight is more than ${formatPounds(MAX_WEIGHT)} lb.` };
  const list: number[] = [];
  for (let w = lightest; w <= heaviest; w += step) {
    list.push(w);
    if (list.length > MAX_WEIGHTS) break;
  }
  if (list[list.length - 1] !== heaviest) list.push(heaviest);
  if (list.length > MAX_WEIGHTS) return { ok: false, error: `The list has more than ${MAX_WEIGHTS} weights. Use a larger step.` };
  return { ok: true, value: list };
}

// addWeight adds one weight to a list, and keeps the list sorted with no
// weight two times.
export function addWeight(list: readonly number[], weight: number | null): Result<number[]> {
  if (weight === null) return { ok: false, error: "Enter a weight of more than 0." };
  if (weight > MAX_WEIGHT) return { ok: false, error: `A weight is ${formatPounds(MAX_WEIGHT)} lb or less.` };
  if (list.includes(weight)) return { ok: false, error: `The list has ${formatPounds(weight)} lb.` };
  if (list.length >= MAX_WEIGHTS) return { ok: false, error: `The list has ${MAX_WEIGHTS} weights.` };
  return { ok: true, value: [...list, weight].sort((a, b) => a - b) };
}

// removeWeight removes one weight from a list.
export function removeWeight(list: readonly number[], weight: number): number[] {
  return list.filter((w) => w !== weight);
}

// checkDumbbells reads a dumbbell set as the server does (D-166).
export function checkDumbbells(lightest: number | null, heaviest: number | null, step: number | null): Result<{
  lightest: number;
  heaviest: number;
  step: number;
}> {
  if (lightest === null || heaviest === null || step === null) {
    return { ok: false, error: "Enter the lightest dumbbell, the heaviest dumbbell, and the step, each more than 0." };
  }
  if (heaviest < lightest) return { ok: false, error: "The heaviest dumbbell is less than the lightest dumbbell." };
  if (heaviest > DUMBBELL_MAX) return { ok: false, error: `The heaviest dumbbell is more than ${formatPounds(DUMBBELL_MAX)} lb.` };
  if ((heaviest - lightest) % step !== 0) {
    return { ok: false, error: "The step does not go from the lightest dumbbell to the heaviest dumbbell." };
  }
  return { ok: true, value: { lightest, heaviest, step } };
}

// dumbbellWeights gives each load of a dumbbell set, lightest first.
export function dumbbellWeights(d: Pick<DumbbellSet, "lightestTenthLb" | "heaviestTenthLb" | "stepTenthLb">): number[] {
  const list: number[] = [];
  if (d.stepTenthLb <= 0) return list;
  for (let w = d.lightestTenthLb; w <= d.heaviestTenthLb; w += d.stepTenthLb) list.push(w);
  return list;
}

// available gives the loads of a machine of the inventory, lightest first.
export function available(m: Pick<InventoryMachine, "weightsTenthLb" | "dumbbells">): number[] {
  return m.dumbbells ? dumbbellWeights(m.dumbbells) : [...m.weightsTenthLb];
}

// checkEstimate reads the optional estimate of one exercise (D-192). An
// empty text is no estimate. A load is from the lightest to the heaviest
// weight of the machine (D-198).
export function checkEstimate(text: string, loads: readonly number[]): Result<number | null> {
  if (text.trim() === "") return { ok: true, value: null };
  const load = parsePounds(text);
  if (load === null) return { ok: false, error: "Enter the estimate in pounds, more than 0." };
  if (loads.length === 0) return { ok: false, error: "Enter the weights before an estimate." };
  const lo = loads[0];
  const hi = loads[loads.length - 1];
  if (load < lo || load > hi) {
    return { ok: false, error: `The estimate is not from ${formatPounds(lo)} lb to ${formatPounds(hi)} lb.` };
  }
  return { ok: true, value: load };
}

// checkNote reads the text of a note: 1 to MAX_NOTE_CHARS characters after
// the trim of the spaces at each end (D-199).
export function checkNote(text: string): Result<string> {
  const t = text.trim();
  if (t.length === 0) return { ok: false, error: "Enter a text." };
  if ([...t].length > MAX_NOTE_CHARS) return { ok: false, error: `A note has ${MAX_NOTE_CHARS} characters or fewer.` };
  return { ok: true, value: t };
}

// guessStep gives the step of a stored list: the smallest difference of
// two weights. It fills the step field when the owner changes a machine.
export function guessStep(list: readonly number[]): number | null {
  let step: number | null = null;
  for (let i = 1; i < list.length; i++) {
    const d = list[i] - list[i - 1];
    if (step === null || d < step) step = d;
  }
  return step;
}
