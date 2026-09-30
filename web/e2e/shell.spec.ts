import { expect, test, type APIRequestContext, type BrowserContext, type Page } from "@playwright/test";

// The acceptance story of work area 2.2. Each test runs in WebKit and in
// Chromium with phone emulation. The tests sign in against the Auth
// emulator, reach the home screen with the uid from the API, prove that
// the page blocks the pinch zoom, and write a change and its outbox entry.
// A test that stops the app closes each page of the context and opens a
// new page. The context keeps its storage, as the phone does.

const authHost = process.env.FIREBASE_AUTH_EMULATOR_HOST;
const firestoreHost = process.env.FIRESTORE_EMULATOR_HOST;
const project = "demo-workout-app";
const password = "emulator-only-1";

// The Auth emulator makes the account, so the app needs no form for it.
async function makeAccount(request: APIRequestContext, email: string): Promise<string> {
  const res = await request.post(`http://${authHost}/identitytoolkit.googleapis.com/v1/accounts:signUp?key=demo-key`, {
    data: { email, password, returnSecureToken: true },
  });
  expect(res.ok(), await res.text()).toBe(true);
  return (await res.json()).localId;
}

// allow writes the allowlist document of a uid (D-131). The emulator
// accepts the `owner` token as an admin, so no rule blocks the write.
async function allow(request: APIRequestContext, uid: string) {
  const url = `http://${firestoreHost}/v1/projects/${project}/databases/(default)/documents/allowlist?documentId=${uid}`;
  const res = await request.post(url, { headers: { Authorization: "Bearer owner" }, data: { fields: {} } });
  expect(res.ok(), await res.text()).toBe(true);
}

async function signIn(page: Page, email: string, pass = password) {
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(pass);
  await page.getByRole("button", { name: "Sign in" }).click();
}

async function stopAndOpen(context: BrowserContext): Promise<Page> {
  for (const p of context.pages()) await p.close();
  const page = await context.newPage();
  await page.goto("/");
  return page;
}

function uniqueEmail(name: string, info: { project: { name: string } }) {
  return `${name}-${info.project.name}-${Date.now()}-${Math.floor(Math.random() * 1e6)}@example.com`;
}

test("the app has one sign-in form and no form that makes an account", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByRole("form", { name: "Sign in" })).toBeVisible();
  await expect(page.locator("form")).toHaveCount(1);
  await expect(page.getByRole("button")).toHaveText(["Sign in"]);
  await expect(page.getByText(/create|sign up|register/i)).toHaveCount(0);
});

test("sign-in reaches the home screen with the uid from the API, and stays after a stop", async ({ page, context }, info) => {
  const email = uniqueEmail("owner", info);
  const uid = await makeAccount(page.request, email);
  await allow(page.request, uid);

  await page.goto("/");
  await signIn(page, email, "wrong-pass-1");
  await expect(page.getByTestId("sign-in-error")).toHaveText(/^auth\/(invalid-credential|wrong-password)$/);

  await signIn(page, email);
  await expect(page.getByTestId("me-uid")).toHaveText(uid);
  await expect(page.getByText(email)).toHaveCount(0);

  const again = await stopAndOpen(context);
  await expect(again.getByTestId("me-uid")).toHaveText(uid);

  await again.getByRole("button", { name: "Sign out" }).click();
  await expect(again.getByRole("form", { name: "Sign in" })).toBeVisible();
});

test("a uid outside the allowlist reaches the home screen and gets permission_denied from the API", async ({ page }, info) => {
  const email = uniqueEmail("stranger", info);
  await makeAccount(page.request, email);

  await page.goto("/");
  await signIn(page, email);
  await expect(page.getByTestId("me-error")).toHaveText("PermissionDenied");
  await expect(page.getByTestId("me-uid")).toHaveCount(0);
});

test("the viewport meta blocks the pinch zoom", async ({ page }) => {
  await page.goto("/");
  const content = await page.locator('meta[name="viewport"]').getAttribute("content");
  const parts = new Set((content ?? "").split(",").map((p) => p.trim()));
  expect(parts).toContain("maximum-scale=1");
  expect(parts).toContain("user-scalable=no");
  expect(parts).toContain("viewport-fit=cover");
});

// Playwright sends a touch of two fingers in Chromium alone, through the
// DevTools protocol. WebKit reads the same meta, and the device check of
// PR-11 pinches the real phone. `Input.synthesizePinchGesture` zooms no
// page in Chromium on Linux, so the test moves two touch points apart.
test.describe("in Chromium", () => {
  // Playwright can not change a page that a service worker serves, and
  // the control below changes the page.
  test.use({ serviceWorkers: "block" });

  test("a pinch does not zoom the page", async ({ context, browserName }) => {
    test.skip(browserName !== "chromium", "Playwright sends a touch of two fingers in Chromium alone");

    // pinchIn loads a page in a new tab, waits until the page shows the
    // shell and draws two frames, and then moves two fingers apart. It
    // gives the scale of the visual viewport after the pinch.
    const pinchIn = async (path: string) => {
      const page = await context.newPage();
      await page.goto(path);
      await expect(page.getByTestId("shell")).toBeVisible();
      await page.evaluate(() => new Promise((r) => requestAnimationFrame(() => requestAnimationFrame(r))));
      const cdp = await context.newCDPSession(page);
      const fingers = (d: number) => [
        { x: 200 - d, y: 400, id: 0 },
        { x: 200 + d, y: 400, id: 1 },
      ];
      await cdp.send("Input.dispatchTouchEvent", { type: "touchStart", touchPoints: fingers(20) });
      for (let d = 30; d <= 150; d += 10) await cdp.send("Input.dispatchTouchEvent", { type: "touchMove", touchPoints: fingers(d) });
      await cdp.send("Input.dispatchTouchEvent", { type: "touchEnd", touchPoints: [] });
      return page.evaluate(() => new Promise<number>((r) => requestAnimationFrame(() => r(window.visualViewport?.scale ?? 0))));
    };

    // The control comes first: the same page with a viewport meta that
    // allows the zoom. The pinch zooms it, so a pinch in this browser
    // works, and the check below proves the block of the meta. Chromium
    // reads the meta at load, so the route changes the page before the
    // load.
    await context.route((url) => url.pathname === "/" && url.search === "?control", async (route) => {
      const res = await route.fetch();
      const html = (await res.text()).replace("maximum-scale=1, user-scalable=no, ", "");
      expect(html).not.toContain("user-scalable=no");
      await route.fulfill({ response: res, body: html });
    });
    expect(await pinchIn("/?control")).toBeGreaterThan(1.5);
    expect(await pinchIn("/")).toBe(1);
  });
});

test("the shell fills the whole screen, and the main region is the one part that scrolls", async ({ page }) => {
  await page.goto("/");
  await expect(page.getByTestId("shell")).toBeVisible();
  const got = await page.evaluate(() => {
    const shell = document.querySelector('[data-testid="shell"]')!.getBoundingClientRect();
    const main = document.querySelector("main")!;
    return {
      top: shell.top,
      bottom: shell.bottom,
      width: shell.width,
      innerHeight: window.innerHeight,
      innerWidth: window.innerWidth,
      docHeight: document.documentElement.scrollHeight,
      docWidth: document.documentElement.scrollWidth,
      bodyOverflow: getComputedStyle(document.body).overflowY,
      mainOverflow: getComputedStyle(main).overflowY,
    };
  });
  expect(got.top).toBe(0);
  expect(got.bottom, "a gap stays below the shell").toBeCloseTo(got.innerHeight, 0);
  expect(got.width).toBeCloseTo(got.innerWidth, 0);
  expect(got.docHeight, "the document is taller than the screen").toBeLessThanOrEqual(got.innerHeight + 1);
  expect(got.docWidth, "the page scrolls sideways").toBeLessThanOrEqual(got.innerWidth);
  expect(got.bodyOverflow).toBe("hidden");
  expect(got.mainOverflow).toBe("auto");
});

test("a change and its outbox entry go into the offline store together, and stay after a stop", async ({ page, context }, info) => {
  const email = uniqueEmail("outbox", info);
  await allow(page.request, await makeAccount(page.request, email));
  await page.goto("/");
  await signIn(page, email);
  await expect(page.getByTestId("outbox-count")).toHaveText("0");

  const entry = await page.evaluate(() => window.workoutAppE2E!.saveSetting("rest-seconds", 90));
  expect(entry).toMatchObject({ entity: "setting", entityId: "rest-seconds", baseVersion: 0, payload: 90, attempts: 0, schemaVersion: 1 });
  expect(entry.opId).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/);
  await expect(page.getByTestId("outbox-count")).toHaveText("1");

  const again = await stopAndOpen(context);
  await expect(again.getByTestId("outbox-count")).toHaveText("1");
  const stored = await again.evaluate(async () => ({
    outbox: await window.workoutAppE2E!.pendingOutbox(),
    settings: await window.workoutAppE2E!.settings(),
  }));
  expect(stored.outbox).toEqual([entry]);
  expect(stored.settings).toEqual([{ id: "rest-seconds", value: 90, version: 0, updatedAt: entry.at }]);
});

test("the first sign-in asks for persistent storage and shows the state", async ({ page }, info) => {
  const email = uniqueEmail("storage", info);
  await allow(page.request, await makeAccount(page.request, email));
  await page.goto("/");
  await signIn(page, email);
  await expect(page.getByTestId("storage-persisted")).toHaveText(/^(yes|no)$/);
  await expect(page.getByTestId("storage-usage")).toHaveText(/^\d+\.\d MB of \d+\.\d MB$/);
});

test("the app is installable: a manifest and a service worker that controls the page", async ({ page }) => {
  await page.goto("/");
  const href = await page.locator('link[rel="manifest"]').getAttribute("href");
  const manifest = await (await page.request.get(href!)).json();
  expect(manifest).toMatchObject({ name: "Workout App", display: "standalone", start_url: "/" });
  expect(manifest.icons).toHaveLength(3);

  await page.waitForFunction(async () => (await navigator.serviceWorker.getRegistration())?.active?.state === "activated");
  // A later start of the app runs under the worker. The worker takes
  // control a moment after it reports the activated state.
  await expect
    .poll(async () => {
      await page.goto("/");
      return page.evaluate(() => navigator.serviceWorker.controller?.scriptURL ?? "");
    })
    .toMatch(/\/sw\.js$/);

  const icon = await page.request.get("/apple-touch-icon.png");
  expect(icon.ok()).toBe(true);
  expect(icon.headers()["content-type"]).toBe("image/png");
});
