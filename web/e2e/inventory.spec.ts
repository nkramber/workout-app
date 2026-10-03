import { expect, test, type Page, type TestInfo } from "@playwright/test";

import { callApi, makeOwner, reopenOffline, signIn, stopAndOpen, syncLine, uniqueEmail } from "./support";

// The acceptance story of work area 4.1. Each test runs in WebKit and in
// Chromium with phone emulation, against the API of go/ and the Firestore
// emulator. The owner adds a machine by selection and a machine by text
// entry, and confirms both. The save makes the weight list (D-245), and
// the save of a cardio machine confirms it (D-246). A change of the
// weights makes a confirmed machine a draft again. A text with no match
// stays as a note. Each change goes into the outbox, and the sync sends
// it (D-250). Each test uses its own account, so it starts with an empty
// inventory.

// openInventory signs in a new owner, opens the inventory, and gives the
// email of the account.
async function openInventory(page: Page, info: TestInfo): Promise<string> {
  const email = uniqueEmail("inventory", info);
  await makeOwner(page.request, email);
  await page.goto("/");
  await signIn(page, email);
  await expect(page.getByTestId("me-uid")).toBeVisible();
  await page.getByRole("button", { name: "Equipment" }).click();
  await expect(page.getByText("No machine yet.")).toBeVisible();
  return email;
}

// aToZ gives a copy of the names in the order of D-205.
const aToZ = (names: string[]) => [...names].sort((a, b) => a.localeCompare(b, "en", { sensitivity: "base", numeric: true }));

const button = (page: Page, name: string) => page.getByRole("button", { name, exact: true });
const weightButtons = (page: Page) => page.getByTestId("weight-list").getByRole("button");
const state = (page: Page) => page.getByTestId("machine-state");

// fillRange enters the lightest weight, the heaviest weight, and the step
// of a stack. The save makes the list (D-195, D-245).
async function fillRange(page: Page, lightest: string, heaviest: string, step: string) {
  await page.getByLabel("Lightest (lb)").fill(lightest);
  await page.getByLabel("Heaviest (lb)").fill(heaviest);
  await page.getByLabel("Step (lb)").fill(step);
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

  // Selection: the catalog list shows the kinds in the order of D-202,
  // and each kind A to Z by name (D-205).
  await button(page, "Add a machine").click();
  await expect(page.getByRole("heading", { level: 3 })).toHaveText(["Machines", "Cable station", "Dumbbells", "Cardio"]);
  for (const kind of ["Machines", "Cardio"]) {
    const shown = await page.getByRole("region", { name: kind }).getByRole("listitem").allTextContents();
    expect(shown.length).toBeGreaterThan(1);
    expect(shown).toEqual(aToZ(shown));
  }
  await page.getByTestId("catalog-chest_press").click();

  // A save with no range makes no list, and tells the owner why.
  await expect(page.getByTestId("list-hint")).toHaveText("Save makes the list of weights from these three values.");
  await button(page, "Save").click();
  await expect(page.getByTestId("change-error")).toHaveText(/^Enter the lightest weight, the heaviest weight, and the step/);

  // The save makes the list from the range, with no other step (D-245).
  await fillRange(page, "10", "100", "10");
  await expect(weightButtons(page)).toHaveCount(0);
  await page.getByLabel("Chest press (lb)").fill("40");
  await button(page, "Save").click();
  await expect(state(page)).toHaveText("Draft");
  await expect(page.getByTestId("shown-weights")).toHaveText("10, 20, 30, 40, 50, 60, 70, 80, 90, 100 lb");

  // On a later visit, the owner removes and adds one weight (D-195).
  await button(page, "Change").click();
  await expect(weightButtons(page)).toHaveCount(10);
  await page.getByRole("button", { name: "Remove 50 lb" }).click();
  await page.getByLabel("One weight (lb)").fill("52.5");
  await button(page, "Add").click();
  await expect(weightButtons(page)).toHaveCount(10);
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

  // The list shows the machines A to Z, not in the order of the adds (D-205).
  await expect(again.getByRole("region", { name: "Machines" }).getByRole("button")).toHaveText([/^Cable station/, /^Chest press/]);
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

  // A change of the range makes a new list at the save (D-245).
  await button(page, "Change").click();
  await page.getByLabel("Heaviest (lb)").fill("350");
  await button(page, "Save").click();
  await expect(page.getByTestId("shown-weights")).toHaveText("50, 100, 150, 200, 250, 300, 350 lb");
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

test("the dumbbells use the dumbbell set, and the save of a cardio machine confirms it", async ({ page }, info) => {
  await openInventory(page, info);

  await button(page, "Add a machine").click();
  await page.getByTestId("catalog-dumbbells").click();
  await page.getByLabel("Lightest (lb)").fill("5");
  await page.getByLabel("Heaviest (lb)").fill("50");
  await page.getByLabel("Step (lb)").fill("5");
  await expect(page.getByTestId("dumbbell-count")).toHaveText("10 pairs of dumbbells.");
  await expect(page.getByTestId("list-hint")).toHaveCount(0);
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

  // A cardio machine has no weights to read, so the save confirms it (D-246).
  await expect(state(page)).toHaveText("Confirmed");
  await expect(page.getByTestId("shown-weights")).toHaveText("A cardio machine has no weights");
  await expect(button(page, "Confirm these weights")).toHaveCount(0);
  await expect(button(page, "Confirm this machine")).toHaveCount(0);
  await expect(button(page, "Change")).toHaveCount(0);
  await button(page, "Back").click();
  await expect(page.getByTestId("machine-treadmill").getByTestId("machine-state")).toHaveText("Confirmed");
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

// D-201 and D-273: a confirmation with no connection waits in the
// outbox. Another device changes the weights on the server, so the server
// refuses the confirmation at the reconnect. The machine shows as a draft
// with the new weights, and the line of the sync shows the refusal until
// the owner dismisses it.
test("a confirmation with no connection waits, and the server refuses it when the weights changed on another device", async ({ page, context, request }, info) => {
  const email = await openInventory(page, info);
  await addStack(page, "biceps_curl", "10", "50", "10");
  await expect(page.getByTestId("shown-weights")).toHaveText("10, 20, 30, 40, 50 lb");
  await expect(syncLine(page)).toHaveText("Synced");

  await context.setOffline(true);
  await button(page, "Confirm these weights").click();
  await expect(state(page)).toHaveText("Confirmed");
  await expect(page.getByTestId("waiting-badge")).toBeVisible();
  await expect(syncLine(page)).toHaveText("Offline · 1 waiting");

  await callApi(request, email, "InventoryService/SaveMachine", { machineId: "biceps_curl", weightsTenthLb: [100, 200, 300, 400] });
  await context.setOffline(false);
  await expect(syncLine(page)).toHaveText("Synced · 1 refused");
  await expect(state(page)).toHaveText("Draft");
  await expect(page.getByTestId("shown-weights")).toHaveText("10, 20, 30, 40 lb");
  await expect(page.getByTestId("waiting-badge")).toHaveCount(0);

  await syncLine(page).click();
  const refused = page.getByTestId("refused-entry");
  await expect(refused).toContainText("Confirmation of Biceps curl");
  await expect(refused).toContainText("The server holds other weights");
  await refused.getByRole("button", { name: "Dismiss" }).click();
  await expect(syncLine(page)).toHaveText("Synced");

  await button(page, "Confirm these weights").click();
  await expect(syncLine(page)).toHaveText("Synced");
  await expect(state(page)).toHaveText("Confirmed");
});

// The inventory part of the acceptance story of PR-32 (D-250): the owner
// adds and confirms a machine and keeps a note with no connection. Each
// change shows at once, waits in the outbox, and stays after a stop of
// the app in Chromium (see reopenOffline). After the reconnect, the
// server holds each change.
test("a change with no connection waits in the outbox, stays after a stop, and reaches the server after the reconnect", async ({ page, context, request, browserName }, info) => {
  const email = await openInventory(page, info);
  await expect(syncLine(page)).toHaveText("Synced");
  await context.setOffline(true);
  await addStack(page, "shoulder_press", "10", "100", "10");
  await expect(state(page)).toHaveText("Draft");
  await button(page, "Confirm these weights").click();
  await expect(state(page)).toHaveText("Confirmed");
  await button(page, "Back").click();
  await button(page, "Add a machine").click();
  await page.getByLabel("Search the catalog, or enter a name").fill("Rope handle");
  await button(page, "Keep the text as a note").click();
  await expect(page.getByTestId("note")).toContainText("Rope handle");
  await expect(syncLine(page)).toHaveText("Offline · 3 waiting");

  let stopped: boolean;
  ({ page, stopped } = await reopenOffline(context, browserName));
  await expect(syncLine(page)).toHaveText("Offline · 3 waiting");
  if (stopped) await button(page, "Equipment").click();
  await expect(page.getByTestId("machine-shoulder_press").getByTestId("machine-state")).toHaveText("Confirmed");
  await expect(page.getByTestId("note")).toContainText("Rope handle");
  await expect(page.getByTestId("waiting-badge")).toHaveCount(2);

  await context.setOffline(false);
  await expect(syncLine(page)).toHaveText("Synced");
  await expect(page.getByTestId("waiting-badge")).toHaveCount(0);
  const { inventory } = await callApi<{ inventory: { machines: { machineId: string; weightsTenthLb: number[]; state: string }[]; notes: { text: string }[] } }>(
    request,
    email,
    "InventoryService/GetInventory",
    {},
  );
  expect(inventory.machines).toEqual([expect.objectContaining({ machineId: "shoulder_press", state: "MACHINE_STATE_CONFIRMED" })]);
  expect(inventory.machines[0].weightsTenthLb).toHaveLength(10);
  expect(inventory.notes.map((n) => n.text)).toEqual(["Rope handle"]);
});

// A server fault of the sync shows "Sync failed", and the change waits.
// "Sync now" sends it after the fault ends (D-276, D-277).
test("a failed sync shows that the sync failed, and sync now sends the change", async ({ page }, info) => {
  await openInventory(page, info);
  await expect(syncLine(page)).toHaveText("Synced");
  await page.route("**/workoutapp.v1.WorkoutService/SyncOutbox", (route) =>
    route.fulfill({
      status: 500,
      contentType: "application/json",
      headers: { "Access-Control-Allow-Origin": "*" },
      body: JSON.stringify({ code: "internal", message: "the workout store failed" }),
    }),
  );
  await addStack(page, "leg_press", "10", "100", "10");
  await expect(syncLine(page)).toHaveText("Sync failed · 1 waiting");
  await expect(page.getByTestId("waiting-badge")).toBeVisible();

  await page.unroute("**/workoutapp.v1.WorkoutService/SyncOutbox");
  await syncLine(page).click();
  await button(page, "Sync now").click();
  await expect(syncLine(page)).toHaveText("Synced");
  await expect(page.getByTestId("waiting-badge")).toHaveCount(0);
  await expect(state(page)).toHaveText("Draft");
});
