import { expect, test, type Page } from "@playwright/test";

import { makeOwner, signIn, uniqueEmail } from "./support";

// The "Screen lock test" screen of D-282. The real result comes from the
// iPhone after the deploy. Here each method starts, logs the result of
// its start, and stops. An engine with no MediaRecorder shows the error
// of the start instead.

const button = (page: Page, name: string) => page.getByRole("button", { name, exact: true });

test("the screen lock test runs each method, logs its start, and stops it", async ({ page }, info) => {
  const email = uniqueEmail("lock-test", info);
  await makeOwner(page.request, email);
  await page.goto("/");
  await signIn(page, email);
  await button(page, "Screen lock test").click();
  await expect(page.getByTestId("lock-test-steps")).toContainText("Set Auto-Lock of the iPhone to 30 seconds.");

  const log = page.getByTestId("lock-test-log").getByRole("listitem");
  for (const title of ["Wake Lock, no tap", "Silent video file", "Silent live video"]) {
    await button(page, `Start ${title}`).click();
    await expect(log.first()).toHaveText(new RegExp(`^0:00 Start: ${title}\\.$`));
    const result = log.nth(1).or(page.getByTestId("lock-test-error"));
    await expect(result).toHaveText(/^(0:0\d start: .+|The method did not start \(.+\)\.)$/, { timeout: 5_000 });
    if (await button(page, `Stop ${title}`).count()) await button(page, `Stop ${title}`).click();
    await expect(button(page, `Start ${title}`)).toBeVisible();
  }

  await button(page, "Back").click();
  await expect(button(page, "Screen lock test")).toBeVisible();
});
