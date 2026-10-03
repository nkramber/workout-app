import { expect, test, type Page } from "@playwright/test";

import { callApi, controlApi, holdSync, makePlanOwner, reopenOffline, signIn, stopAndOpen, syncLine, uniqueEmail } from "./support";

// The workout screen and the set log of work area 6.1. The API uses the
// fake provider of Luna (D-241), so each plan has 3 sessions with the
// chest press and the seated row, and 20 minutes on the treadmill. Each
// machine has the weights 10 to 200 lb, in steps of 10 lb. No test calls
// OpenAI (D-24). The phone keeps each log. A test that reads the outbox
// holds the sync, so the entries stay (work area 6.3).

const MACHINES = ["chest_press", "seated_row", "treadmill"];

const button = (page: Page, name: string) => page.getByRole("button", { name, exact: true });
const logger = (page: Page) => page.getByTestId("set-logger");
const exercise = (page: Page, id: string) => page.locator(`[data-testid="workout-exercise"][data-exercise-id="${id}"]`);

// openWithPlan signs in, makes a plan with the fake, and goes back to the
// home screen.
async function openWithPlan(page: Page, email: string) {
  await page.goto("/");
  await signIn(page, email);
  await button(page, "Plan").click();
  await button(page, "Make a plan").click();
  await expect(page.getByTestId("plan-summary")).toHaveText("A plan at the targets of the rules.");
  await button(page, "Back").click();
}

const outbox = (page: Page) =>
  page.evaluate(async () => (await window.workoutAppE2E!.pendingOutbox()).map((e) => ({ entity: e.entity, entityId: e.entityId })));

// The acceptance story: the owner starts the next session, and logs a set
// in three taps or fewer. A symptom report shows the warning, and the
// owner continues after the confirmation. After a stop and an open of the
// app, the workout and each logged set stay on the phone.
test("the owner starts the next session, logs a set in three taps, reports a symptom, and keeps the workout", async ({ page, context, request }, info) => {
  const email = uniqueEmail("workout", info);
  await makePlanOwner(request, email, MACHINES);
  await holdSync(page);
  await openWithPlan(page, email);

  // Tap 1: the workout screen. Tap 2: the next session (D-248).
  await button(page, "Workout").click();
  await expect(page.getByTestId("next-session")).toContainText("Session 1");
  await button(page, "Start Session 1").click();

  // The reps and the weight come from the target (D-249).
  await expect(logger(page).getByTestId("logger-exercise")).toHaveText("Chest press");
  await expect(logger(page).getByTestId("set-label")).toHaveText(/^(Calibration set|Set) 1 of \d+$/);
  const firstLabel = (await logger(page).getByTestId("set-label").textContent()) ?? "";
  const target = (await logger(page).getByTestId("set-target").textContent()) ?? "";
  const [, reps, weight] = /^Target: (\d+) reps? at ([\d.]+ lb)/.exec(target) ?? [];
  expect(reps, target).toBeTruthy();
  await expect(logger(page).getByTestId("reps")).toHaveText(reps);
  await expect(logger(page).getByTestId("weight")).toHaveText(weight);

  // Tap 3: the reps in reserve log the set.
  await button(page, "2 in reserve").click();
  const logged = exercise(page, "chest_press").getByTestId("logged-sets").getByRole("listitem");
  await expect(logged).toHaveCount(1);
  await expect(logged.first()).toContainText(`${reps} reps at ${weight}, 2 in reserve`);
  await expect(logger(page).getByTestId("set-label")).not.toHaveText(firstLabel);

  // The workout and the set each have an outbox entry (D-132).
  await expect.poll(async () => (await outbox(page)).map((e) => e.entity)).toEqual(["workout", "set"]);

  // A symptom report shows the warning of D-153, and the owner continues
  // after the confirmation (D-40, D-251).
  await button(page, "Report a symptom").click();
  await button(page, "Dizziness or feeling faint").click();
  const dialog = page.getByRole("alertdialog");
  await expect(dialog.getByTestId("warning-text")).toHaveText("You reported dizziness or feeling faint. Stop this exercise.");
  await button(page, "Continue the workout").click();
  await expect(page.getByRole("alertdialog")).toHaveCount(0);
  await expect(logger(page)).toBeVisible();

  // After a stop and an open of the app, the workout and the set stay.
  const again = await stopAndOpen(context);
  await expect(again.getByTestId("me-uid")).toBeVisible();
  await button(again, "Continue workout").click();
  await expect(again.getByRole("heading", { name: "Session 1", exact: true })).toBeVisible();
  const kept = exercise(again, "chest_press").getByTestId("logged-sets").getByRole("listitem");
  await expect(kept).toHaveCount(1);
  await expect(kept.first()).toContainText(`${reps} reps at ${weight}, 2 in reserve`);
  const stored = await again.evaluate(async () => ({
    workouts: (await window.workoutAppE2E!.workouts()).map((w) => ({ finished: w.finished, sessionIndex: w.sessionIndex })),
    sets: (await window.workoutAppE2E!.sets()).length,
  }));
  expect(stored).toEqual({ workouts: [{ finished: false, sessionIndex: 0 }], sets: 1 });
});

test("the plus and minus buttons change the reps and step the weight on the list, and pain gives the warning", async ({ page, request }, info) => {
  const email = uniqueEmail("workout-steps", info);
  await makePlanOwner(request, email, MACHINES);
  await openWithPlan(page, email);
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();

  const weightText = (await logger(page).getByTestId("weight").textContent()) ?? "";
  const start = Number.parseFloat(weightText);
  const reps = Number((await logger(page).getByTestId("reps").textContent()) ?? "");

  // The weight steps to the next weight of the list of the machine (D-264).
  // A new exercise starts at the lightest weight, 10 lb, so a tap of
  // minus there keeps the weight (D-150).
  await button(page, "Heavier").click();
  await expect(logger(page).getByTestId("weight")).toHaveText(`${start + 10} lb`);
  await button(page, "Lighter").click();
  await expect(logger(page).getByTestId("weight")).toHaveText(`${start} lb`);
  const final = start > 10 ? start - 10 : 10;
  await button(page, "Lighter").click();
  await expect(logger(page).getByTestId("weight")).toHaveText(`${final} lb`);
  await button(page, "More reps").click();
  await expect(logger(page).getByTestId("reps")).toHaveText(String(reps + 1));
  await button(page, "Fewer reps").click();
  await button(page, "Fewer reps").click();
  await expect(logger(page).getByTestId("reps")).toHaveText(String(reps - 1));

  // Pain and a note are behind one tap (D-57, D-162). A pain report gives
  // the warning of D-169. The first set of a new exercise is the
  // calibration set, so it offers 0 to 6+ (D-268).
  await expect(logger(page).getByTestId("set-label")).toHaveText("Calibration set 1 of 1");
  await button(page, "Add pain or a note").click();
  await button(page, "Pain 3").click();
  await logger(page).getByLabel("Note (optional)").fill("Synthetic note.");
  await expect(button(page, "4+ in reserve")).toHaveCount(0);
  await button(page, "4 in reserve").click();
  await expect(page.getByRole("alertdialog").getByTestId("warning-text")).toHaveText("You reported pain. Stop this exercise.");
  await button(page, "Continue the workout").click();
  await expect(exercise(page, "chest_press").getByTestId("logged-sets")).toContainText(
    `${reps - 1} reps at ${final} lb, 4 in reserve, pain 3`,
  );
  // A working set offers 0 to 4+ (D-249).
  await expect(logger(page).getByTestId("set-label")).toHaveText("Set 1 of 3");
  await expect(button(page, "4+ in reserve")).toBeVisible();
  await expect(button(page, "6+ in reserve")).toHaveCount(0);
});

test("the cardio log holds the duration and the effort, and the optional fields", async ({ page, request }, info) => {
  const email = uniqueEmail("workout-cardio", info);
  await makePlanOwner(request, email, MACHINES);
  await holdSync(page);
  await openWithPlan(page, email);
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();

  const card = page.getByTestId("cardio-card");
  await expect(card).toContainText("Treadmill");
  await expect(card.getByTestId("cardio-minutes")).toHaveText("20");
  await expect(button(page, "Log cardio")).toBeDisabled();
  await button(page, "More minutes").click();
  await button(page, "Effort 6").click();
  await button(page, "Add distance, level, pain, or a note").click();
  await card.getByLabel("Distance in miles (optional)").fill("2.55");
  await expect(button(page, "Log cardio")).toBeDisabled();
  await card.getByLabel("Distance in miles (optional)").fill("2.5");
  await card.getByLabel("Resistance level (optional)").fill("4");
  await button(page, "Log cardio").click();
  await expect(card.getByTestId("cardio-logged")).toHaveText("Logged: 21 min, effort 6 of 10");
  await expect.poll(async () => (await outbox(page)).map((e) => e.entity)).toEqual(["workout", "cardio"]);
});

test("finish now skips the exercises with no set, and the next session comes next", async ({ page, request }, info) => {
  const email = uniqueEmail("workout-finish", info);
  await makePlanOwner(request, email, MACHINES);
  await openWithPlan(page, email);
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();
  await button(page, "3 in reserve").click();
  await expect(exercise(page, "chest_press").getByTestId("logged-sets").getByRole("listitem")).toHaveCount(1);

  await button(page, "Finish now").click();
  await expect(page.getByRole("alertdialog").getByTestId("finish-text")).toHaveText(
    "2 exercises have a set with no log, so the workout ends early. An exercise with no logged set counts as skipped.",
  );
  await page.getByRole("alertdialog").getByRole("button", { name: "Finish now", exact: true }).click();
  await expect(button(page, "Workout")).toBeVisible();

  // Session 1 is done, so Session 2 is next, and the owner can pick
  // another session (D-248).
  await button(page, "Workout").click();
  await expect(page.getByTestId("next-session")).toContainText("Session 2");
  await button(page, "Start Session 3").click();
  await expect(page.getByRole("heading", { name: "Session 3", exact: true })).toBeVisible();

  const workouts = await page.evaluate(async () =>
    (await window.workoutAppE2E!.workouts()).map((w) => ({
      sessionIndex: w.sessionIndex,
      finished: w.finished,
      endedEarly: w.endedEarly,
      skipped: w.skippedExerciseIds,
    })),
  );
  expect(workouts).toContainEqual({ sessionIndex: 0, finished: true, endedEarly: true, skipped: ["seated_row"] });
  expect(workouts).toContainEqual({ sessionIndex: 2, finished: false, endedEarly: false, skipped: [] });
});

test("the plan screen refuses a new plan and an exclusion during a workout", async ({ page, request }, info) => {
  const email = uniqueEmail("workout-lock", info);
  await makePlanOwner(request, email, MACHINES);
  await openWithPlan(page, email);
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();
  await button(page, "Back").click();

  // D-252: the owner finishes the workout first.
  await button(page, "Plan").click();
  await expect(page.getByTestId("workout-lock")).toHaveText(
    "A workout is in progress. Finish it before you make a new plan or exclude an exercise.",
  );
  await expect(button(page, "Make a new plan")).toBeDisabled();
  for (const b of await page.getByRole("button", { name: "Exclude", exact: true }).all()) await expect(b).toBeDisabled();

  await button(page, "Back").click();
  await button(page, "Continue workout").click();
  await button(page, "Finish now").click();
  await page.getByRole("alertdialog").getByRole("button", { name: "Finish now", exact: true }).click();
  await button(page, "Plan").click();
  await expect(page.getByTestId("workout-lock")).toHaveCount(0);
  await expect(button(page, "Make a new plan")).toBeEnabled();
});

// The rest timer, the preview, and the automatic advance of work area 6.2.
// The fake plan starts the chest press at the lightest weight, 10 lb,
// with one calibration set and 3 working sets, and a rest of 60 seconds
// (D-150, D-279).

// lock sends the app to the back as a screen lock does, moves the clock
// of the page by ms with no timer, and brings the app back. The phone
// runs no timer during a lock, so the timer must read its stored end
// time (D-270).
async function lockFor(page: Page, ms: number) {
  const visibility = (state: "hidden" | "visible") =>
    page.evaluate((s) => {
      Object.defineProperty(document, "visibilityState", { configurable: true, get: () => s });
      document.dispatchEvent(new Event("visibilitychange"));
    }, state);
  await visibility("hidden");
  const now = await page.evaluate(() => Date.now());
  await page.clock.setSystemTime(now + ms);
  await visibility("visible");
}

const restLeft = (page: Page) => page.getByTestId("rest-timer").getByTestId("rest-left");

// The acceptance story of PR-31: the timer shows the correct time after a
// screen lock and a return, and the advance comes after the last set.
test("the rest timer is correct after a screen lock, and the next machine comes after the last set", async ({ page, request }, info) => {
  const email = uniqueEmail("workout-rest", info);
  await makePlanOwner(request, email, MACHINES);
  await page.clock.install();
  await openWithPlan(page, email);
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();
  await expect(page.getByTestId("rest-timer")).toHaveCount(0);

  // The log of a set starts the timer with the rest of the target (D-59).
  await button(page, "3 in reserve").click();
  await expect(restLeft(page)).toHaveText(/^(1:00|0:59)$/);
  await expect(page.getByTestId("rest-state")).toHaveText("Rest");

  // A lock of 15 s leaves 45 s.
  await lockFor(page, 15_000);
  await expect(restLeft(page)).toHaveText(/^0:4[3-5]$/);

  // The controls of D-270.
  await button(page, "15 seconds more rest").click();
  await expect(restLeft(page)).toHaveText(/^(0:5[89]|1:00)$/);
  await button(page, "15 seconds less rest").click();
  await expect(restLeft(page)).toHaveText(/^0:4[3-5]$/);

  // The end is a visual cue alone (D-58).
  await lockFor(page, 60_000);
  await expect(restLeft(page)).toHaveText("0:00");
  await expect(page.getByTestId("rest-state")).toHaveText("Rest done");
  await expect(page.getByTestId("rest-timer")).toHaveAttribute("data-done", "true");
  await button(page, "Dismiss").click();
  await expect(page.getByTestId("rest-timer")).toHaveCount(0);

  // The 3 working sets. After the last one, the preview shows the next
  // machine, and the rest timer runs (D-60, D-269).
  for (let i = 1; i <= 3; i++) {
    await expect(logger(page).getByTestId("set-label")).toHaveText(`Set ${i} of 3`);
    await button(page, "3 in reserve").click();
  }
  const preview = page.getByTestId("next-preview");
  await expect(preview.getByTestId("preview-done")).toHaveText("Chest press is done.");
  await expect(preview.getByTestId("preview-next")).toHaveText("Next: Seated row");
  await expect(preview.getByTestId("preview-seconds")).toHaveText(/^The next machine shows in (10|9) s\.$/);
  await expect(logger(page)).toHaveCount(0);
  await expect(page.getByTestId("rest-state")).toHaveText("Rest");

  // The automatic advance after 10 s.
  await page.clock.fastForward(10_000);
  await expect(preview).toHaveCount(0);
  await expect(logger(page).getByTestId("logger-exercise")).toHaveText("Seated row");
  await expect(restLeft(page)).toHaveText(/^0:[45]\d$/);
});

// The calibration step of D-267 with no network: 6+ reps in reserve on
// the calibration set at 10 lb give two 5 lb steps, and the machine has
// 20 lb. "Go now" advances at once, and a logged set can be edited.
test("the calibration set gives the working load with no network, go now advances, and a set can be edited", async ({ page, context, request }, info) => {
  const email = uniqueEmail("workout-calibration", info);
  await makePlanOwner(request, email, MACHINES);
  await holdSync(page);
  await openWithPlan(page, email);
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();
  await expect(logger(page).getByTestId("set-target")).toContainText("at 10 lb");
  await context.setOffline(true);

  await button(page, "6+ in reserve").click();
  await expect(logger(page).getByTestId("set-label")).toHaveText("Set 1 of 3");
  await expect(logger(page).getByTestId("set-target")).toContainText("at 20 lb");
  await expect(logger(page).getByTestId("weight")).toHaveText("20 lb");
  await expect(logger(page).getByTestId("calibration-note")).toHaveText("The calibration set gave this load.");
  for (let i = 1; i <= 3; i++) await button(page, "3 in reserve").click();
  await expect(page.getByTestId("next-preview")).toBeVisible();
  await button(page, "Go now").click();
  await expect(logger(page).getByTestId("logger-exercise")).toHaveText("Seated row");

  // The edit of the calibration set keeps its kind and its place (D-63).
  const rows = exercise(page, "chest_press").getByTestId("logged-text");
  await expect(rows.first()).toContainText("Calibration: ");
  await expect(rows.first()).toContainText("at 10 lb, 6+ in reserve");
  await button(page, "Edit set 1 of Chest press").click();
  const dialog = page.getByRole("alertdialog");
  await expect(dialog.getByRole("button", { name: "6+ in reserve", exact: true })).toHaveAttribute("aria-pressed", "true");
  await dialog.getByRole("button", { name: "More reps", exact: true }).click();
  await dialog.getByRole("button", { name: "5 in reserve", exact: true }).click();
  await dialog.getByRole("button", { name: "Save the change", exact: true }).click();
  await expect(page.getByRole("alertdialog")).toHaveCount(0);
  await expect(rows.first()).toContainText("at 10 lb, 5 in reserve");
  await expect(rows).toHaveCount(4);

  const sets = await page.evaluate(async () => (await window.workoutAppE2E!.pendingOutbox()).filter((e) => e.entity === "set"));
  const firstId = sets[0].entityId;
  expect(sets.filter((e) => e.entityId === firstId)).toHaveLength(2);
  expect(sets.at(-1)!.payload).toMatchObject({ kind: "calibration", rir: 5 });
  await context.setOffline(false);
});

// The acceptance story of PR-31: a skip and "finish now" give the correct
// session log (D-63, D-170).
test("a skip and finish now give the correct session log", async ({ page, request }, info) => {
  const email = uniqueEmail("workout-skip", info);
  await makePlanOwner(request, email, MACHINES);
  await holdSync(page);
  await openWithPlan(page, email);
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();

  await expect(logger(page).getByTestId("logger-exercise")).toHaveText("Chest press");
  await button(page, "Skip this exercise").click();
  await expect(page.getByRole("alertdialog").getByRole("heading")).toHaveText("Skip Chest press?");
  await page.getByRole("alertdialog").getByRole("button", { name: "Skip", exact: true }).click();
  await expect(logger(page).getByTestId("logger-exercise")).toHaveText("Seated row");
  await expect(exercise(page, "chest_press").getByTestId("exercise-skipped")).toHaveText("Skipped");

  // The table applies to the weight that the owner logged (D-249, D-267):
  // 20 lb in place of 10 lb, at 6+ reps in reserve, gives 30 lb.
  await button(page, "Heavier").click();
  await expect(logger(page).getByTestId("weight")).toHaveText("20 lb");
  await button(page, "6+ in reserve").click();
  await expect(logger(page).getByTestId("set-target")).toContainText("at 30 lb");
  await button(page, "Finish now").click();
  await expect(page.getByRole("alertdialog").getByTestId("finish-text")).toHaveText(
    "1 exercise has a set with no log, so the workout ends early.",
  );
  await page.getByRole("alertdialog").getByRole("button", { name: "Finish now", exact: true }).click();
  await expect(button(page, "Workout")).toBeVisible();

  const log = await page.evaluate(async () => ({
    workouts: (await window.workoutAppE2E!.workouts()).map((w) => ({ finished: w.finished, endedEarly: w.endedEarly, skipped: w.skippedExerciseIds })),
    sets: (await window.workoutAppE2E!.sets()).map((s) => ({ exerciseId: s.exerciseId, kind: s.kind })),
    headers: (await window.workoutAppE2E!.pendingOutbox())
      .filter((e) => e.entity === "workout")
      .map((e) => e.payload as { skippedExerciseIds?: string[]; endedEarly?: boolean; finished?: boolean }),
  }));
  expect(log.workouts).toEqual([{ finished: true, endedEarly: true, skipped: ["chest_press"] }]);
  expect(log.sets).toEqual([{ exerciseId: "seated_row", kind: "calibration" }]);
  expect(log.headers).toHaveLength(3);
  expect(log.headers[1]).toMatchObject({ skippedExerciseIds: ["chest_press"] });
  expect(log.headers[2]).toMatchObject({ skippedExerciseIds: ["chest_press"], endedEarly: true, finished: true });
});

// D-271: the notice names the error, and a tap gets the lock again. The
// fake lock of this test refuses each request until the test allows it.
test("a refused wake lock shows the error name, and a tap gets the lock again", async ({ page, request }, info) => {
  const email = uniqueEmail("workout-wake", info);
  await makePlanOwner(request, email, MACHINES);
  await page.addInitScript(() => {
    const w = window as unknown as { wakeRefuse: boolean; wakeRequests: number };
    w.wakeRefuse = true;
    w.wakeRequests = 0;
    const fake = {
      request: async () => {
        w.wakeRequests++;
        if (w.wakeRefuse) throw new DOMException("refused", "NotAllowedError");
        return { released: false, release: async () => {}, addEventListener: () => {} };
      },
    };
    Object.defineProperty(navigator, "wakeLock", { configurable: true, get: () => fake });
  });
  await openWithPlan(page, email);
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();
  await expect(page.getByTestId("wake-off")).toContainText("The screen can turn off. Tap the screen to try again.");
  await expect(page.getByTestId("wake-error")).toHaveText("(NotAllowedError)");

  await page.evaluate(() => ((window as unknown as { wakeRefuse: boolean }).wakeRefuse = false));
  await page.getByRole("heading", { name: "Session 1", exact: true }).click();
  await expect(page.getByTestId("wake-off")).toHaveCount(0);
});

// The state of the workout screen: the preview of the next machine, the
// next set of the set log, or the end of the strength work. The page reads
// it in one step, because the set log leaves the screen when the preview
// comes, and a second read of a locator would wait for it.
function screenState(page: Page): Promise<string> {
  return page.evaluate(() => {
    if (document.querySelector('[data-testid="next-preview"]')) return "preview";
    const log = document.querySelector('[data-testid="set-logger"]');
    if (!log) return "done";
    const name = log.querySelector('[data-testid="logger-exercise"]')?.textContent;
    const label = log.querySelector('[data-testid="set-label"]')?.textContent;
    return `${name}: ${label}`;
  });
}

// logSets logs the next sets at 3 reps in reserve, and goes past each
// preview, until the strength work ends or `count` sets are logged. It
// gives the count of logged sets.
async function logSets(page: Page, count = Infinity): Promise<number> {
  await expect(logger(page).or(page.getByTestId("next-preview"))).toBeVisible();
  let logged = 0;
  while (logged < count) {
    const state = await screenState(page);
    if (state === "done") break;
    if (state === "preview") await button(page, "Go now").click();
    else {
      await button(page, "3 in reserve").click();
      logged++;
    }
    try {
      await expect.poll(() => screenState(page)).not.toBe(state);
    } catch {
      // The error holds the screen, so a failure in CI shows its cause.
      const shown = await page.getByTestId("shell").ariaSnapshot();
      throw new Error(`the screen stayed at "${state}" after ${logged} sets:\n${shown}`);
    }
  }
  return logged;
}

type ServerWorkout = { workoutId: string; finished?: boolean; exercises: { exerciseId: string; sets?: { setId: string }[] }[]; cardio?: unknown[] };

// The acceptance story of PR-32 (work area 6.3). The owner opens the app
// with no connection, and completes a full workout from the offline
// copies (D-278), with a stop of the app after two sets. WebKit of
// Playwright can not open a page with no connection, so there the app
// stays open (see reopenOffline). At the
// reconnect, the server applies the first sync call, but its answer does
// not reach the phone. The phone sends the batch again, and the server
// holds each set one time (D-257).
test("a full workout with no connection, a stop of the app, and a dropped answer gives each set on the server one time", async ({ page, context, request, browserName }, info) => {
  const api = await controlApi(context);
  const email = uniqueEmail("workout-offline", info);
  await makePlanOwner(request, email, MACHINES);
  await openWithPlan(page, email);
  await expect(syncLine(page)).toHaveText("Synced");

  await context.setOffline(true);
  ({ page } = await reopenOffline(context, browserName));
  await button(page, "Workout").click();
  await button(page, "Start Session 1").click();
  let logged = await logSets(page, 2);
  expect(logged).toBe(2);

  let stopped: boolean;
  ({ page, stopped } = await reopenOffline(context, browserName));
  if (stopped) await button(page, "Continue workout").click();
  logged += await logSets(page);
  const card = page.getByTestId("cardio-card");
  await button(page, "Effort 6").click();
  await button(page, "Log cardio").click();
  await expect(card.getByTestId("cardio-logged")).toBeVisible();
  await button(page, "Finish workout").click();
  await expect(button(page, "Workout")).toBeVisible();
  const waiting = logged + 3; // the sets, the cardio log, and the start and the end of the workout
  await expect(syncLine(page)).toHaveText(`Offline · ${waiting} waiting`);

  // The first sync call reaches the server, and its answer drops.
  let dropped = 0;
  api.sync = async (route) => {
    if (dropped > 0) return route.fallback();
    dropped++;
    await route.fetch();
    await route.abort("connectionreset");
  };
  await context.setOffline(false);
  await expect.poll(() => dropped).toBe(1);
  await expect(syncLine(page)).toHaveText("Synced", { timeout: 20_000 });

  const phone = await page.evaluate(async () => ({
    sets: (await window.workoutAppE2E!.sets()).map((s) => ({ id: s.id, version: s.version })),
    outbox: (await window.workoutAppE2E!.pendingOutbox()).length,
  }));
  expect(phone.outbox).toBe(0);
  expect(phone.sets).toHaveLength(logged);
  expect(phone.sets.every((s) => s.version >= 1)).toBe(true);

  const { workouts } = await callApi<{ workouts: ServerWorkout[] }>(request, email, "WorkoutService/ListWorkouts", {});
  expect(workouts).toHaveLength(1);
  expect(workouts[0].finished).toBe(true);
  expect(workouts[0].cardio).toHaveLength(1);
  const serverSets = workouts[0].exercises.flatMap((e) => (e.sets ?? []).map((s) => s.setId));
  expect(serverSets).toHaveLength(logged);
  expect(new Set(serverSets).size).toBe(logged);
  expect([...serverSets].sort()).toEqual(phone.sets.map((s) => s.id).sort());
});
