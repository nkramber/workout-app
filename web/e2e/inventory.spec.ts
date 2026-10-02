import { expect, test, type Page, type TestInfo } from "@playwright/test";

import { allow, makeAccount, signIn, stopAndOpen, uniqueEmail } from "./support";

// The acceptance story of work area 4.1. Each test runs in WebKit and in
// Chromium with phone emulation, against the API of go/ and the Firestore
// emulator. The owner adds a machine by selection and a machine by text
// entry, and confirms both. A change of the weights makes a confirmed
// machine a draft again. A text with no match stays as a note. Each test
// uses its own account, so it starts with an empty inventory.

async function openInventory(page: Page, info: TestInfo) {
  const email = uniqueEmail("inventory", info);
  await allow(page.request, await makeAccount(page.request, email));
  await page.goto("/");
  await signIn(page, email);
  await expect(page.getByTestId("me-uid")).toBeVisible();
  await page.getByRole("button", { name: "Equipment" }).click();
  await expect(page.getByText("No machine yet.")).toBeVisible();
}

const button = (page: Page, name: string) => page.getByRole("button", { name, exact: true });
const weightButtons = (page: Page) => page.getByTestId("weight-list").getByRole("button");
const state = (page: Page) => page.getByTestId("machine-state");

// fillRange enters the lightest weight, the heaviest weight, and the step
// of a stack, and makes the list (D-195).
async function fillRange(page: Page, lightest: string, heaviest: string, step: string) {
  await page.getByLabel("Lightest (lb)").fill(lightest);
  await page.getByLabel("Heaviest (lb)").fill(heaviest);
  await page.getByLabel("Step (lb)").fill(step);
  await button(page, "Make the list").click();
}

// addStack adds a machine of the catalog list with a range of weights,
// and stops on its review screen.
async function addStack(page: Page, machineId: string, lightest: string, heaviest: string, step: string) {
  await button(page, "Add a machine").click();
  await page.getByTestId(`catalog-${machineId}`).click();
  await fillRange(page, lightest, heaviest, step);
  await button(page, "Save").click();
  await expect(state(page)).toHaveText("Draft");
}

test("the owner adds a machine by selection and a machine by text entry, and confirms both", async ({ page, context }, info) => {
  await openInventory(page, info);

  // Selection: the catalog list shows the kinds in the order of D-202.
  await button(page, "Add a machine").click();
  await expect(page.getByRole("heading", { level: 3 })).toHaveText(["Machines", "Cable station", "Dumbbells", "Cardio"]);
  await page.getByTestId("catalog-chest_press").click();

  // The range makes the list, then the owner removes and adds one weight.
  await fillRange(page, "10", "100", "10");
  await expect(weightButtons(page)).toHaveCount(10);
  await page.getByRole("button", { name: "Remove 50 lb" }).click();
  await page.getByLabel("One weight (lb)").fill("52.5");
  await button(page, "Add").click();
  await expect(weightButtons(page)).toHaveCount(10);
  await page.getByLabel("Chest press (lb)").fill("40");
  await button(page, "Save").click();

  // The review screen shows the stored weights, and confirms them.
  await expect(state(page)).toHaveText("Draft");
  await expect(page.getByTestId("shown-weights")).toHaveText("10, 20, 30, 40, 52.5, 60, 70, 80, 90, 100 lb");
  await expect(page.getByTestId("estimate-chest_press")).toHaveText("40 lb");
  await button(page, "Confirm these weights").click();
  await expect(state(page)).toHaveText("Confirmed");
  await expect(button(page, "Confirm these weights")).toHaveCount(0);
  await button(page, "Back").click();

  // Text entry: the text finds the cable station by the name of an
  // exercise, and the owner selects the match.
  await button(page, "Add a machine").click();
  await page.getByLabel("Search the catalog, or enter a name").fill("lat pull");
  await expect(page.getByRole("region", { name: "Matches" }).getByRole("button", { name: /Cable station/ })).toBeVisible();
  await page.getByTestId("catalog-cable_station").click();
  await fillRange(page, "5", "50", "5");
  await button(page, "Save").click();
  await expect(state(page)).toHaveText("Draft");
  await button(page, "Confirm these weights").click();
  await expect(state(page)).toHaveText("Confirmed");
  await button(page, "Back").click();

  // The server keeps both confirmed machines after a stop of the app.
  const again = await stopAndOpen(context);
  await again.getByRole("button", { name: "Equipment" }).click();
  await expect(again.getByTestId("machine-chest_press").getByTestId("machine-state")).toHaveText("Confirmed");
  await expect(again.getByTestId("machine-cable_station").getByTestId("machine-state")).toHaveText("Confirmed");
  await expect(again.getByTestId("note")).toHaveCount(0);
});

test("a change of the weights makes a confirmed machine a draft again, and a change of an estimate alone does not", async ({ page }, info) => {
  await openInventory(page, info);
  await addStack(page, "leg_press", "50", "300", "50");
  await button(page, "Confirm these weights").click();
  await expect(state(page)).toHaveText("Confirmed");

  // A change of the estimates alone keeps the confirmation (D-200).
  await button(page, "Change").click();
  await expect(page.getByTestId("draft-warning")).toBeVisible();
  await page.getByLabel("Leg press (lb)").fill("150");
  await button(page, "Save").click();
  await expect(page.getByTestId("estimate-leg_press")).toHaveText("150 lb");
  await expect(state(page)).toHaveText("Confirmed");

  // A change of the weights makes a draft again (D-193).
  await button(page, "Change").click();
  await page.getByRole("button", { name: "Remove 300 lb" }).click();
  await button(page, "Save").click();
  await expect(state(page)).toHaveText("Draft");
  await expect(page.getByTestId("shown-weights")).toHaveText("50, 100, 150, 200, 250 lb");
  await button(page, "Back").click();
  await expect(page.getByTestId("machine-leg_press").getByTestId("machine-state")).toHaveText("Draft");
});

test("a text with no match stays as a note, and the owner removes the note", async ({ page }, info) => {
  await openInventory(page, info);
  await button(page, "Add a machine").click();
  await page.getByLabel("Search the catalog, or enter a name").fill("  Smith machine ");
  await expect(page.getByTestId("no-match")).toBeVisible();
  await button(page, "Keep the text as a note").click();

  await expect(page.getByTestId("note")).toHaveCount(1);
  await expect(page.getByTestId("note")).toContainText("Smith machine");
  await expect(page.getByText("No machine yet.")).toBeVisible();

  await page.getByRole("button", { name: "Remove the note Smith machine" }).click();
  await expect(page.getByText("No note.")).toBeVisible();
});

test("the dumbbells use the dumbbell set, and a cardio machine has no weights", async ({ page }, info) => {
  await openInventory(page, info);

  await button(page, "Add a machine").click();
  await page.getByTestId("catalog-dumbbells").click();
  await page.getByLabel("Lightest (lb)").fill("5");
  await page.getByLabel("Heaviest (lb)").fill("50");
  await page.getByLabel("Step (lb)").fill("5");
  await expect(page.getByTestId("dumbbell-count")).toHaveText("10 pairs of dumbbells.");
  await expect(button(page, "Make the list")).toHaveCount(0);
  await page.getByLabel("Dumbbell goblet squat (lb)").fill("25");
  await button(page, "Save").click();
  await expect(page.getByTestId("shown-weights")).toHaveText("5 to 50 lb, step 5 lb (10 pairs)");
  await button(page, "Confirm these weights").click();
  await expect(state(page)).toHaveText("Confirmed");
  await button(page, "Back").click();

  await button(page, "Add a machine").click();
  await page.getByTestId("catalog-treadmill").click();
  await expect(page.getByTestId("no-weights")).toBeVisible();
  await expect(page.getByRole("region", { name: "Estimates" })).toHaveCount(0);
  await button(page, "Save").click();
  await expect(page.getByTestId("shown-weights")).toHaveText("No weights");
  await button(page, "Confirm these weights").click();
  await expect(state(page)).toHaveText("Confirmed");
});

test("the review screen removes a machine", async ({ page }, info) => {
  await openInventory(page, info);
  await addStack(page, "seated_row", "10", "150", "10");
  await button(page, "Remove the machine").click();
  await button(page, "Cancel").click();
  await button(page, "Remove the machine").click();
  await button(page, "Yes, remove").click();
  await expect(page.getByText("No machine yet.")).toBeVisible();

  // The catalog list no longer marks the machine as held.
  await button(page, "Add a machine").click();
  await expect(page.getByTestId("catalog-seated_row")).not.toContainText("In the inventory");
});

test("the server refuses a confirmation of weights that changed after the review screen showed them", async ({ page, context }, info) => {
  await openInventory(page, info);
  await addStack(page, "biceps_curl", "10", "50", "10");
  await expect(page.getByTestId("shown-weights")).toHaveText("10, 20, 30, 40, 50 lb");

  // A second page of the app changes the weights.
  const other = await context.newPage();
  await other.goto("/");
  await other.getByRole("button", { name: "Equipment" }).click();
  await other.getByTestId("machine-biceps_curl").click();
  await button(other, "Change").click();
  await other.getByRole("button", { name: "Remove 50 lb" }).click();
  await button(other, "Save").click();
  await expect(other.getByTestId("shown-weights")).toHaveText("10, 20, 30, 40 lb");
  await other.close();

  // The first page still shows the old weights, so the server refuses its
  // confirmation (D-201). The screen then reads the new weights.
  await button(page, "Confirm these weights").click();
  await expect(page.getByTestId("change-error")).toContainText("The weights changed");
  await expect(page.getByTestId("shown-weights")).toHaveText("10, 20, 30, 40 lb");
  await expect(state(page)).toHaveText("Draft");
  await button(page, "Confirm these weights").click();
  await expect(state(page)).toHaveText("Confirmed");
});

test("a change with no connection shows an error and saves nothing (D-196)", async ({ page, context }, info) => {
  await openInventory(page, info);
  await button(page, "Add a machine").click();
  await page.getByTestId("catalog-shoulder_press").click();
  await fillRange(page, "10", "100", "10");

  await context.setOffline(true);
  await button(page, "Save").click();
  await expect(page.getByTestId("change-error")).toHaveText(/^No connection\. The change is not saved\./);
  await expect(button(page, "Save")).toBeEnabled();

  await context.setOffline(false);
  await button(page, "Save").click();
  await expect(state(page)).toHaveText("Draft");
});
