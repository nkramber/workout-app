import { expect, test, type Page, type TestInfo } from "@playwright/test";

import { allow, makeAccount, signIn, stopAndOpen, uniqueEmail } from "./support";

// The acceptance story of the onboarding screen (work area 5.1). Each test
// runs in WebKit and in Chromium with phone emulation, against the API of
// go/ and the Firestore emulator. The owner fills each input, sees the
// injury warning, selects a template, changes its groups, and saves the
// profile. A second visit shows the saved profile. Each test uses its own
// account, so it starts with no profile.

const WARNING =
  "Your plan avoids exercises that load the areas you selected. This app gives fitness guidance only, and it does not diagnose or treat an injury.";

// The groups of the template "Strength" (D-220).
const STRENGTH = ["Chest", "Back", "Shoulders", "Quadriceps", "Hamstrings", "Glutes"];

async function openOnboarding(page: Page, info: TestInfo) {
  const email = uniqueEmail("profile", info);
  await allow(page.request, await makeAccount(page.request, email));
  await page.goto("/");
  await signIn(page, email);
  await expect(page.getByRole("heading", { name: "Set up your profile" })).toBeVisible();
}

const choice = (page: Page, group: string, name: string) =>
  page.getByTestId(group).getByRole("button", { name, exact: true });
const pressed = (page: Page, group: string) => page.getByTestId(group).locator('button[aria-pressed="true"]');
// back is the Back button of the title. The muscle group "Back" is a
// toggle, so it has aria-pressed.
const back = (page: Page) => page.locator("button:not([aria-pressed])", { hasText: /^Back$/ });
const save = (page: Page) => page.getByRole("button", { name: "Save", exact: true });
const error = (page: Page) => page.getByTestId("change-error");

// fillRequired fills each input that the server needs, with no injury.
async function fillRequired(page: Page) {
  await choice(page, "experience", "Intermediate").click();
  await choice(page, "training-days", "3 days").click();
  await choice(page, "goal-templates", "General fitness").click();
  await page.getByLabel("Age (years)").fill("40");
  await page.getByLabel("Height (ft)").selectOption("5");
  await page.getByLabel("Height (in)").selectOption("10");
  await page.getByLabel("Weight (lb)").fill("180");
}

test("the owner fills each input, sees the injury warning, changes the template groups, and a second visit shows the profile", async ({
  page,
  context,
}, info) => {
  await openOnboarding(page, info);
  // Onboarding comes before the home screen, with no way past it (D-223).
  await expect(page.getByTestId("me-uid")).toHaveCount(0);
  await expect(back(page)).toHaveCount(0);

  await save(page).click();
  await expect(error(page)).toHaveText("Select your experience.");

  await choice(page, "experience", "Advanced").click();
  await choice(page, "training-days", "3 days").click();
  await choice(page, "goal-templates", "Strength").click();
  await expect(pressed(page, "goal-templates")).toHaveText(["Strength"]);
  await expect(pressed(page, "muscle-groups")).toHaveText(STRENGTH);
  await choice(page, "muscle-groups", "Shoulders").click();
  await choice(page, "muscle-groups", "Core").click();
  await expect(pressed(page, "muscle-groups")).toHaveText(["Chest", "Back", "Quadriceps", "Hamstrings", "Glutes", "Core"]);
  await page.getByLabel("Other goals for your plan (optional)").fill("More pulling than pushing.");

  await expect(page.getByTestId("injury-warning")).toHaveCount(0);
  await choice(page, "injured-areas", "Knee").click();
  await expect(page.getByTestId("injury-warning")).toHaveText(WARNING);
  await choice(page, "injured-areas", "Lower back").click();
  await choice(page, "injured-areas", "Knee").click();
  await expect(page.getByTestId("injury-warning")).toHaveText(WARNING);
  await page.getByLabel("Other injuries or limits (optional)").fill("Synthetic test text.");

  await page.getByLabel("Age (years)").fill("44");
  await page.getByLabel("Height (ft)").selectOption("6");
  await page.getByLabel("Height (in)").selectOption("1");
  await page.getByLabel("Weight (lb)").fill("205");
  await choice(page, "cardio", "Rowing machine").click();
  await choice(page, "cardio", "Treadmill").click();

  await save(page).click();
  await expect(page.getByTestId("me-uid")).toBeVisible();

  // The second visit goes to the home screen, and the profile screen
  // shows each saved value.
  const again = await stopAndOpen(context);
  await expect(again.getByTestId("me-uid")).toBeVisible();
  await again.getByRole("button", { name: "Profile", exact: true }).click();
  await expect(again.getByRole("heading", { name: "Profile", exact: true })).toBeVisible();
  await expect(pressed(again, "experience")).toHaveText(["Advanced"]);
  await expect(pressed(again, "training-days")).toHaveText(["3 days"]);
  await expect(pressed(again, "goal-templates")).toHaveText(["Strength"]);
  await expect(pressed(again, "muscle-groups")).toHaveText(["Chest", "Back", "Quadriceps", "Hamstrings", "Glutes", "Core"]);
  await expect(again.getByLabel("Other goals for your plan (optional)")).toHaveValue("More pulling than pushing.");
  await expect(pressed(again, "injured-areas")).toHaveText(["Lower back"]);
  await expect(again.getByTestId("injury-warning")).toHaveText(WARNING);
  await expect(again.getByLabel("Other injuries or limits (optional)")).toHaveValue("Synthetic test text.");
  await expect(again.getByLabel("Age (years)")).toHaveValue("44");
  await expect(again.getByLabel("Height (ft)")).toHaveValue("6");
  await expect(again.getByLabel("Height (in)")).toHaveValue("1");
  await expect(again.getByLabel("Weight (lb)")).toHaveValue("205");
  // The cardio list keeps the catalog order.
  await expect(pressed(again, "cardio")).toHaveText(["Treadmill", "Rowing machine"]);

  // A change saves and goes back to the home screen.
  await choice(again, "training-days", "4 days").click();
  await save(again).click();
  await expect(again.getByTestId("me-uid")).toBeVisible();
  await again.getByRole("button", { name: "Profile", exact: true }).click();
  await expect(pressed(again, "training-days")).toHaveText(["4 days"]);

  // Back leaves with no save.
  await choice(again, "training-days", "2 days").click();
  await back(again).click();
  await again.getByRole("button", { name: "Profile", exact: true }).click();
  await expect(pressed(again, "training-days")).toHaveText(["4 days"]);
});

test("a value outside its bound gives a message and saves nothing (D-215)", async ({ page, context }, info) => {
  await openOnboarding(page, info);
  await fillRequired(page);
  await page.getByLabel("Age (years)").fill("17");
  await save(page).click();
  await expect(error(page)).toHaveText("Enter an age from 18 to 90 years.");

  const again = await stopAndOpen(context);
  await expect(again.getByRole("heading", { name: "Set up your profile" })).toBeVisible();
});

test("a save with no connection shows an error and saves nothing (D-196)", async ({ page, context }, info) => {
  await openOnboarding(page, info);
  await fillRequired(page);

  await context.setOffline(true);
  await save(page).click();
  await expect(error(page)).toHaveText(/^No connection\. The change is not saved\./);
  await expect(save(page)).toBeEnabled();

  await context.setOffline(false);
  await save(page).click();
  await expect(page.getByTestId("me-uid")).toBeVisible();
});

test("the server refusal of a value shows that the server did not accept it", async ({ page }, info) => {
  await openOnboarding(page, info);
  await fillRequired(page);

  await page.route("**/workoutapp.v1.ProfileService/SaveProfile", (route) =>
    route.fulfill({
      status: 400,
      contentType: "application/json",
      headers: { "Access-Control-Allow-Origin": "*" },
      body: JSON.stringify({ code: "invalid_argument", message: "invalid: profile age: want 18 to 90" }),
    }),
  );
  await save(page).click();
  await expect(error(page)).toHaveText("The server did not accept the values. Read them again.");
  await expect(page.getByRole("heading", { name: "Set up your profile" })).toBeVisible();
});

test("the onboarding screen can sign out", async ({ page }, info) => {
  await openOnboarding(page, info);
  await page.getByRole("button", { name: "Sign out" }).click();
  await expect(page.getByRole("form", { name: "Sign in" })).toBeVisible();
});
