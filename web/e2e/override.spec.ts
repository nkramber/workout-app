import { expect, test, type Page } from "@playwright/test";

import { holdSync, makePlanOwner, signIn, stopAndOpen, uniqueEmail } from "./support";

// An override of the owner (D-69, D-293). The API uses the fake provider
// of Luna (D-241), so the chest press of Session 1 starts with 3 working
// sets at 10 lb, and the first set is the calibration. Each machine has the
// weights 10 to 200 lb, in steps of 10 lb. The policy of the server
// checks each override (D-23).

const MACHINES = ["chest_press", "seated_row", "treadmill"];

const button = (page: Page, name: string) => page.getByRole("button", { name, exact: true });
const logger = (page: Page) => page.getByTestId("set-logger");
const card = (page: Page) => page.locator('[data-testid="plan-exercise"][data-exercise-id="chest_press"]').first();

// The acceptance story: an override shows in the next workout, and the
// recommendation and the reason stay as separate records.
test("an override shows in the next workout, with the recommendation and the reason apart", async ({ page, context, request }, info) => {
  const email = uniqueEmail("override", info);
  await makePlanOwner(request, email, MACHINES);
  await holdSync(page);
  await page.goto("/");
  await signIn(page, email);
  await button(page, "Plan").click();
  await button(page, "Make a plan").click();
  await expect(page.getByTestId("plan-summary")).toHaveText("A plan at the targets of the rules.");

  const recommended = (await card(page).getByTestId("sets").getByRole("listitem").allTextContents()).filter((t) => t.startsWith("Set "));
  expect(recommended.length).toBeGreaterThan(0);
  expect(recommended[0]).toContain("at 10 lb");

  // A load that the machine does not have is refused, and nothing saves.
  await card(page).getByRole("button", { name: "Change target", exact: true }).click();
  const form = card(page).getByTestId("override-form");
  const save = form.getByRole("button", { name: "Save the change", exact: true });
  await expect(save).toBeDisabled();
  await form.getByRole("textbox").last().fill("My shoulder feels fine this week.");
  await form.getByLabel("Set 1 load").fill("15");
  await save.click();
  await expect(card(page).getByTestId("override-error")).toContainText("The policy did not accept the change.");
  await expect(card(page).getByTestId("override")).toHaveCount(0);

  // A valid change saves. The recommendation stays, and the reason shows.
  for (let i = 1; i <= recommended.length; i++) await form.getByLabel(`Set ${i} load`).fill("20");
  await save.click();
  await expect(card(page).getByTestId("override-form")).toHaveCount(0);
  const shown = card(page).getByTestId("override");
  await expect(shown).toContainText("Your change for the next session");
  await expect(shown.getByRole("listitem").filter({ hasText: /^Set 1[,:] / })).toContainText("at 20 lb");
  await expect(shown.getByTestId("override-reason")).toHaveText("Your reason: My shoulder feels fine this week.");
  await expect(card(page).getByTestId("sets").getByRole("listitem").filter({ hasText: /^Set 1[,:] / })).toHaveText(recommended[0]);

  // "Use the recommendation" removes the override, and the owner saves it
  // again. One test holds both, so the file makes one plan, because the
  // plan requests of parallel tests share the cap documents.
  await card(page).getByRole("button", { name: "Use the recommendation", exact: true }).click();
  await expect(card(page).getByTestId("override")).toHaveCount(0);
  await expect(card(page).getByRole("button", { name: "Use the recommendation", exact: true })).toHaveCount(0);
  await card(page).getByRole("button", { name: "Change target", exact: true }).click();
  for (let i = 1; i <= recommended.length; i++) await form.getByLabel(`Set ${i} load`).fill("20");
  await form.getByRole("textbox").last().fill("My shoulder feels fine this week.");
  await save.click();
  await expect(card(page).getByTestId("override")).toBeVisible();

  // The offline copy keeps the override after a stop of the app.
  const again = await stopAndOpen(context);
  await expect(again.getByTestId("me-uid")).toBeVisible();
  await button(again, "Workout").click();
  await button(again, "Start Session 1").click();

  // The next workout uses the override: each working set starts at 20 lb,
  // and the first set stays the calibration (D-301).
  await expect(logger(again).getByTestId("logger-exercise")).toHaveText("Chest press");
  await expect(logger(again).getByTestId("set-label")).toHaveText(/^Set 1 of \d+, the calibration$/);
  await expect(logger(again).getByTestId("set-target")).toContainText("at 20 lb");
  await button(again, "3 in reserve").click();
  await expect(logger(again).getByTestId("set-label")).toHaveText(/^Set 2 of \d+$/);
  await expect(logger(again).getByTestId("set-target")).toContainText("at 20 lb");
  await expect(logger(again).getByTestId("override-note")).toContainText("Your change. Recommended:");
  await expect(logger(again).getByTestId("override-note")).toContainText("at 10 lb");

  // The workout keeps the override, the recommendation, and the reason as
  // separate records, and its header sends them (D-69, D-291). The unit
  // tests of `src/lib/workout.test.ts` read the header. In WebKit, the hold
  // of the sync does not apply after the stop, so this test reads the
  // workout of the phone.
  const exercises = await again.evaluate(async () => (await window.workoutAppE2E!.workouts())[0].exercises);
  const chest = exercises.find((e) => e.exerciseId === "chest_press")!;
  expect(chest.workingSets.every((s) => s.loadTenthLb === 200)).toBe(true);
  expect(chest.override?.recommendedWorkingSets.every((s) => s.loadTenthLb === 100)).toBe(true);
  expect(chest.override?.reason).toBe("My shoulder feels fine this week.");
  expect(exercises.find((e) => e.exerciseId === "seated_row")!.override).toBeUndefined();
});

