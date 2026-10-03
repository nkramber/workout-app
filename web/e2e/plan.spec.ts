import { expect, test, type Page, type Route } from "@playwright/test";

import { makePlanOwner, signIn, uniqueEmail } from "./support";

// The plan screen of work area 5.2. The API uses the fake provider of
// Luna with a wait of 1 s for each call, so each test sees the progress
// (D-241). The fake gives each part of a plan: the warm-up, the work
// sets, the rest, the cardio, the cool-down, and one mobility and one
// recovery item (D-44). No test calls OpenAI (D-24).

const MACHINES = ["chest_press", "seated_row", "treadmill"];

const button = (page: Page, name: string) => page.getByRole("button", { name, exact: true });
const sessions = (page: Page) => page.getByTestId("plan-session");
const progress = (page: Page) => page.getByTestId("plan-progress");
const planError = (page: Page) => page.getByTestId("plan-error");
const exercise = (page: Page, id: string) => page.locator(`[data-testid="plan-exercise"][data-exercise-id="${id}"]`);

async function openPlan(page: Page, email: string) {
  await page.goto("/");
  await signIn(page, email);
  await button(page, "Plan").click();
}

// The acceptance story: the owner requests a plan, sees the progress and
// each part of the plan from the API, excludes an exercise with a reason,
// and sees the new plan.
test("the owner requests a plan, sees each part, excludes an exercise, and sees the new plan", async ({ page, request }, info) => {
  const email = uniqueEmail("plan", info);
  await makePlanOwner(request, email, MACHINES);
  await openPlan(page, email);

  await expect(page.getByTestId("no-plan")).toBeVisible();
  await button(page, "Make a plan").click();
  await expect(progress(page)).toHaveText("Asking Luna for a plan. Try 1 of 4.");
  await expect(button(page, "Back")).toHaveCount(0);

  await expect(page.getByTestId("plan-summary")).toHaveText("A plan at the targets of the rules.");
  await expect(progress(page)).toHaveCount(0);
  await expect(sessions(page)).toHaveCount(3);
  for (const s of await sessions(page).all()) {
    await expect(s.getByTestId("warm-up")).toContainText("Do 5 minutes of easy cardio.");
    await expect(s.getByTestId("cool-down")).toContainText("Walk at an easy pace for 5 minutes.");
    await expect(s.getByTestId("cardio")).toHaveText("Cardio (optional): Treadmill, 10 min");
    await expect(s.getByTestId("plan-exercise")).toHaveCount(2);
  }
  const press = exercise(page, "chest_press").first();
  await expect(press.getByRole("heading")).toHaveText("Chest press");
  await expect(press.getByTestId("sets").getByRole("listitem").first()).toHaveText(/^(Calibration set|Set) 1: \d+ reps? at [\d.]+ lb/);
  await expect(press.getByTestId("rest")).toHaveText(/^Rest .+ between sets\.$/);
  const guidance = page.getByTestId("guidance");
  await expect(guidance).toContainText("Mobility: Do 10 slow leg swings");
  await expect(guidance).toContainText("Recovery: Leave one rest day");

  // The exclusion of the chest press, with a reason.
  await press.getByRole("button", { name: "Exclude" }).click();
  const form = page.getByTestId("exclude-form");
  await form.getByLabel("Reason (optional)").fill("  The seat does not fit.  ");
  await expect(form).toContainText("22 of 200 characters");
  await form.getByRole("button", { name: "Exclude and make a new plan" }).click();
  await expect(progress(page)).toHaveText("Asking Luna for a plan. Try 1 of 4.");
  await expect(progress(page)).toBeInViewport();

  const excluded = page.getByTestId("exclusion");
  await expect(excluded).toHaveCount(1);
  await expect(excluded).toContainText("Chest press");
  await expect(excluded).toContainText("The seat does not fit.");
  await expect(exercise(page, "chest_press")).toHaveCount(0);
  await expect(sessions(page)).toHaveCount(3);
  await expect(exercise(page, "seated_row")).toHaveCount(3);

  // The server holds the new plan and the exclusion.
  await page.reload();
  await button(page, "Plan").click();
  await expect(exercise(page, "seated_row")).toHaveCount(3);
  await expect(exercise(page, "chest_press")).toHaveCount(0);
  await expect(page.getByTestId("exclusion")).toHaveCount(1);
});

test("a reason over 200 characters can not go to the server", async ({ page, request }, info) => {
  const email = uniqueEmail("plan-reason", info);
  await makePlanOwner(request, email, MACHINES);
  await openPlan(page, email);
  await button(page, "Make a plan").click();
  await expect(sessions(page)).toHaveCount(3);

  await exercise(page, "chest_press").first().getByRole("button", { name: "Exclude" }).click();
  const form = page.getByTestId("exclude-form");
  await form.getByLabel("Reason (optional)").fill("x".repeat(201));
  await expect(form).toContainText("201 of 200 characters");
  await expect(form.getByRole("button", { name: "Exclude and make a new plan" })).toBeDisabled();
  await form.getByLabel("Reason (optional)").fill("x".repeat(200));
  await expect(form.getByRole("button", { name: "Exclude and make a new plan" })).toBeEnabled();
  await form.getByRole("button", { name: "Cancel" }).click();
  await expect(form).toHaveCount(0);
});

test("an owner with no confirmed machine gets the error of no allowed exercise", async ({ page, request }, info) => {
  const email = uniqueEmail("plan-none", info);
  await makePlanOwner(request, email, []);
  await openPlan(page, email);

  await button(page, "Make a plan").click();
  await expect(planError(page)).toHaveText(
    "No confirmed machine gives an exercise that you can do. Confirm a machine, or change the injuries in your profile.",
  );
  await expect(page.getByTestId("no-plan")).toBeVisible();
  await expect(button(page, "Back")).toBeVisible();
});

// The Connect stream of a stubbed call: each message in an envelope, then
// the end of the stream with an error (Connect protocol, server stream).
function envelope(flags: number, body: unknown): Buffer {
  const json = Buffer.from(JSON.stringify(body));
  const head = Buffer.alloc(5);
  head.writeUInt8(flags, 0);
  head.writeUInt32BE(json.length, 1);
  return Buffer.concat([head, json]);
}

function stubStream(messages: unknown[], code: string, message: string) {
  return async (route: Route) => {
    const origin = route.request().headers()["origin"] ?? "*";
    const cors = {
      "access-control-allow-origin": origin,
      "access-control-allow-headers": "*",
      "access-control-allow-methods": "POST",
      "access-control-expose-headers": "*",
    };
    if (route.request().method() === "OPTIONS") return route.fulfill({ status: 204, headers: cors });
    const body = Buffer.concat([...messages.map((m) => envelope(0, m)), envelope(2, { error: { code, message } })]);
    return route.fulfill({ status: 200, headers: { ...cors, "content-type": "application/connect+json" }, body });
  };
}

const call = (attempt: number, previousStatus = "") => ({ progress: { step: "call", attempt, maxAttempts: 4, previousStatus } });

// A stub gives the two errors that the fake can not give: 4 failed calls
// and the cap. The plan and the exclusions do not change (D-230, D-234).
test("4 failed calls and the cap give their errors, and the plan does not change", async ({ page, request }, info) => {
  const email = uniqueEmail("plan-errors", info);
  await makePlanOwner(request, email, MACHINES);
  await openPlan(page, email);
  await button(page, "Make a plan").click();
  const summary = page.getByTestId("plan-summary");
  await expect(summary).toHaveText("A plan at the targets of the rules.");

  await page.route(
    "**/workoutapp.v1.PlanService/RequestPlan",
    stubStream([call(1), call(2, "malformed"), call(3, "timeout"), call(4, "refusal")], "unavailable", "plan: Luna gave no valid plan"),
  );
  await button(page, "Make a new plan").click();
  await expect(planError(page)).toHaveText("Luna gave no valid plan in 4 tries. Your plan did not change. Try again later.");
  await expect(summary).toBeVisible();
  await expect(sessions(page)).toHaveCount(3);

  await page.unroute("**/workoutapp.v1.PlanService/RequestPlan");
  await page.route(
    "**/workoutapp.v1.PlanService/ExcludeExercise",
    stubStream([], "resource_exhausted", "plan: the monthly AI cap can not cover the call"),
  );
  await exercise(page, "chest_press").first().getByRole("button", { name: "Exclude" }).click();
  await page.getByRole("button", { name: "Exclude and make a new plan" }).click();
  await expect(planError(page)).toHaveText(
    "The exercise is not excluded. The monthly AI limit is reached. Your plan did not change. The limit resets on the first day of the month (UTC).",
  );
  await expect(planError(page)).toBeInViewport();
  await expect(exercise(page, "chest_press")).toHaveCount(3);
  await expect(page.getByTestId("exclusion")).toHaveCount(0);
});
