import { expect, test, type BrowserContext, type Page } from "@playwright/test";

// The browser tests of the probe (D-113). Each test runs in WebKit and in
// Chromium. A test that stops the app closes each page of the context and
// opens a new page. The context keeps its storage, as the phone does.

async function stopAndOpen(context: BrowserContext, hash: string): Promise<Page> {
  for (const p of context.pages()) await p.close();
  const page = await context.newPage();
  await page.goto(`/#/${hash}`);
  return page;
}

test("the probe has one page for each device item and no camera page", async ({ page }) => {
  await page.goto("/");
  const nav = page.getByRole("navigation", { name: "Probe pages" });
  await expect(nav.getByRole("link")).toHaveText(["Storage", "Wake", "Install", "Sign-in", "Startup"]);
  await expect(page.getByText(/camera/i)).toHaveCount(0);
  await expect(page.getByTestId("display-mode")).toHaveText("browser");
  await expect(page.getByTestId("build-id")).not.toHaveText("");
});

test("an IndexedDB record stays after the app stops", async ({ page, context }) => {
  await page.goto("/#/storage");
  await expect(page.getByTestId("note-count")).toHaveText("0");
  await page.getByRole("button", { name: "Write a record" }).click();
  await expect(page.getByTestId("note-count")).toHaveText("1");
  await expect(page.getByTestId("note-earlier")).toHaveText("0");

  const again = await stopAndOpen(context, "storage");
  await expect(again.getByTestId("note-count")).toHaveText("1");
  await expect(again.getByTestId("note-earlier")).toHaveText("1");
  await expect(again.getByTestId("note-list")).toContainText("record 1");
  await expect(again.getByTestId("storage-error")).toHaveCount(0);
});

test("the wake lock page requests the lock and shows the result", async ({ page }) => {
  await page.goto("/#/wake-lock");
  const supported = await page.getByTestId("wake-supported").textContent();
  test.skip(supported !== "yes", "this engine gives no Screen Wake Lock API");
  await expect(page.getByTestId("wake-state")).toHaveText("released");
  await page.getByRole("button", { name: "Keep the screen on" }).click();
  // A headless browser can refuse the lock. The page must show the
  // result either way, and never stay silent.
  await expect(page.getByTestId("wake-state")).toHaveText(/^(active|error)$/);
  if ((await page.getByTestId("wake-state").textContent()) === "active") {
    await page.getByRole("button", { name: "Release" }).click();
    await expect(page.getByTestId("wake-state")).toHaveText("released");
  } else {
    await expect(page.getByTestId("wake-error")).not.toBeEmpty();
  }
});

test("the probe is installable: a manifest and a service worker that controls the page", async ({ page }) => {
  await page.goto("/#/install");
  await expect(page.getByTestId("install-manifest")).toHaveText("Gym Route probe, standalone, 3 icons");
  await expect(page.getByTestId("install-sw-state")).toHaveText("activated");
  await page.reload();
  await expect(page.getByTestId("install-sw-controlled")).toHaveText("yes");

  const icon = await page.request.get("/apple-touch-icon.png");
  expect(icon.ok()).toBe(true);
  expect(icon.headers()["content-type"]).toBe("image/png");
});

// The Auth emulator makes the account, so the probe needs no form for it.
async function makeAccount(page: Page, email: string, password: string) {
  const res = await page.request.post(
    "http://127.0.0.1:9099/identitytoolkit.googleapis.com/v1/accounts:signUp?key=demo-key",
    { data: { email, password, returnSecureToken: true } },
  );
  expect(res.ok(), await res.text()).toBe(true);
}

test("a sign-in with email and password stays after the app stops", async ({ page, context }, info) => {
  const email = `probe-${info.project.name}-${Date.now()}@example.com`;
  const password = "probe-pass-1";
  await makeAccount(page, email, password);

  await page.goto("/#/sign-in");
  await expect(page.getByTestId("auth-state")).toHaveText("signed out");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill("wrong-pass-1");
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page.getByTestId("auth-error")).toHaveText(/^auth\/(invalid-credential|wrong-password)$/);

  await page.getByLabel("Password").fill(password);
  await page.getByRole("button", { name: "Sign in" }).click();
  await expect(page.getByTestId("auth-state")).toHaveText("signed in");
  await expect(page.getByText(email)).toHaveCount(0);

  const again = await stopAndOpen(context, "sign-in");
  await expect(again.getByTestId("auth-state")).toHaveText("signed in");
  await expect(again.getByTestId("auth-ready")).toHaveText(/^\d+ ms$/);

  await again.getByRole("button", { name: "Sign out" }).click();
  await expect(again.getByTestId("auth-state")).toHaveText("signed out");
});

test("the startup timer records each launch", async ({ page, context }) => {
  await page.goto("/#/startup");
  await expect(page.getByTestId("startup-render")).toHaveText(/^\d+ ms$/);
  await expect(page.getByTestId("startup-fcp")).toHaveText(/^\d+ ms$/);
  await expect(page.getByTestId("startup-count")).toHaveText("1");

  const again = await stopAndOpen(context, "startup");
  await expect(again.getByTestId("startup-count")).toHaveText("2");
  await expect(again.getByTestId("startup-median-browser")).toHaveText(/^\d+ ms$/);
});
