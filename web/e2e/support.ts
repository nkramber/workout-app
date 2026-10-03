import { expect, type APIRequestContext, type BrowserContext, type Page, type Route } from "@playwright/test";

// The shared steps of the browser tests. Each test makes its own account
// on the Auth emulator, so the tests can run in parallel.

const authHost = process.env.FIREBASE_AUTH_EMULATOR_HOST;
const firestoreHost = process.env.FIRESTORE_EMULATOR_HOST;
const project = "demo-workout-app";
const password = "emulator-only-1";

// apiOrigin is the API of playwright.config.ts.
const apiOrigin = "http://127.0.0.1:8480";

// The Auth emulator makes the account, so the app needs no form for it.
async function signUp(request: APIRequestContext, email: string): Promise<{ uid: string; idToken: string }> {
  const res = await request.post(`http://${authHost}/identitytoolkit.googleapis.com/v1/accounts:signUp?key=demo-key`, {
    data: { email, password, returnSecureToken: true },
  });
  expect(res.ok(), await res.text()).toBe(true);
  const body = await res.json();
  return { uid: body.localId, idToken: body.idToken };
}

export async function makeAccount(request: APIRequestContext, email: string): Promise<string> {
  return (await signUp(request, email)).uid;
}

// makeOwner makes an account on the allowlist with a saved profile, so
// the sign-in goes past onboarding to the home screen (D-223). The save
// goes through the API, with the token of the account.
export async function makeOwner(request: APIRequestContext, email: string): Promise<string> {
  const { uid, idToken } = await signUp(request, email);
  await allow(request, uid);
  const res = await request.post(`${apiOrigin}/workoutapp.v1.ProfileService/SaveProfile`, {
    headers: { Authorization: `Bearer ${idToken}` },
    data: {
      profile: {
        experience: "intermediate",
        goalTemplate: "general_fitness",
        muscleGroups: ["chest", "back"],
        ageYears: 40,
        heightIn: 70,
        weightLb: 180,
        trainingDays: 3,
      },
    },
  });
  expect(res.ok(), await res.text()).toBe(true);
  return uid;
}

// call sends one Connect call to the API with the token of an account.
async function call(request: APIRequestContext, idToken: string, method: string, data: unknown) {
  const res = await request.post(`${apiOrigin}/workoutapp.v1.${method}`, {
    headers: { Authorization: `Bearer ${idToken}` },
    data,
  });
  expect(res.ok(), await res.text()).toBe(true);
}

// makePlanOwner makes an owner whose profile asks for chest and back work
// and the treadmill as cardio, on 3 training days. Each machine of the
// list is saved and confirmed through the API, so a plan can use it
// (D-193). With no machine, the plan request finds no allowed exercise.
export async function makePlanOwner(request: APIRequestContext, email: string, machines: string[]): Promise<string> {
  const { uid, idToken } = await signUp(request, email);
  await allow(request, uid);
  await call(request, idToken, "ProfileService/SaveProfile", {
    profile: {
      experience: "intermediate",
      goalTemplate: "general_fitness",
      muscleGroups: ["chest", "back"],
      ageYears: 40,
      heightIn: 70,
      weightLb: 180,
      cardioExerciseIds: ["treadmill"],
      trainingDays: 3,
    },
  });
  for (const machineId of machines) {
    const weightsTenthLb = machineId === "treadmill" ? [] : Array.from({ length: 20 }, (_, i) => (i + 1) * 100);
    await call(request, idToken, "InventoryService/SaveMachine", { machineId, weightsTenthLb });
    await call(request, idToken, "InventoryService/ConfirmMachine", { machineId, weightsTenthLb });
  }
  return uid;
}

// allow writes the allowlist document of a uid (D-131). The emulator
// accepts the `owner` token as an admin, so no rule blocks the write.
export async function allow(request: APIRequestContext, uid: string) {
  const url = `http://${firestoreHost}/v1/projects/${project}/databases/(default)/documents/allowlist?documentId=${uid}`;
  const res = await request.post(url, { headers: { Authorization: "Bearer owner" }, data: { fields: {} } });
  expect(res.ok(), await res.text()).toBe(true);
}

export async function signIn(page: Page, email: string, pass = password) {
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Password").fill(pass);
  await page.getByRole("button", { name: "Sign in" }).click();
}

export async function stopAndOpen(context: BrowserContext): Promise<Page> {
  for (const p of context.pages()) await p.close();
  const page = await context.newPage();
  await page.goto("/");
  return page;
}

export function uniqueEmail(name: string, info: { project: { name: string } }) {
  return `${name}-${info.project.name}-${Date.now()}-${Math.floor(Math.random() * 1e6)}@example.com`;
}

// tokenOf signs in to the Auth emulator, and gives the ID token of the
// account, so a test can read the server as another device does.
async function tokenOf(request: APIRequestContext, email: string): Promise<string> {
  const res = await request.post(`http://${authHost}/identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=demo-key`, {
    data: { email, password, returnSecureToken: true },
  });
  expect(res.ok(), await res.text()).toBe(true);
  return (await res.json()).idToken;
}

// callApi sends one Connect call to the API for the account of the email,
// and gives the JSON answer.
export async function callApi<T = Record<string, unknown>>(request: APIRequestContext, email: string, method: string, data: unknown): Promise<T> {
  const res = await request.post(`${apiOrigin}/workoutapp.v1.${method}`, {
    headers: { Authorization: `Bearer ${await tokenOf(request, email)}` },
    data,
  });
  expect(res.ok(), await res.text()).toBe(true);
  return res.json();
}

// holdSync makes each sync call of the browser context fail as with no
// connection, so the outbox keeps its entries. The copies still load
// (work area 6.3). A new page of the context gets the hold too.
export async function holdSync(page: Page) {
  await page.context().route("**/workoutapp.v1.WorkoutService/SyncOutbox", (route) => route.abort("internetdisconnected"));
}

// syncLine is the line of the state of the sync (D-276).
export const syncLine = (page: Page) => page.getByTestId("sync-line");

// ApiControl changes the calls of the API in a test. blocked refuses
// each call, as with no connection. sync, when set, handles each
// SyncOutbox call.
export type ApiControl = { blocked: boolean; sync?: (route: Route) => Promise<void> };

// controlApi puts one route of the API on the context. Call it at the
// start of a test: WebKit of Playwright does not apply a route that comes
// after the context was offline.
export async function controlApi(context: BrowserContext): Promise<ApiControl> {
  const control: ApiControl = { blocked: false };
  await context.route(
    (url) => url.origin === apiOrigin,
    async (route) => {
      if (control.blocked) return route.abort("internetdisconnected");
      if (control.sync && new URL(route.request().url()).pathname.endsWith("/SyncOutbox")) return control.sync(route);
      return route.fallback();
    },
  );
  return control;
}

// reopenOffline stops the app, and opens it again with no connection,
// in Chromium. It gives the page and true for a stop. WebKit of
// Playwright refuses each navigation of an offline context, the
// navigations that the service worker can serve too. A route of the
// context also did not apply to the calls of a page that opened again
// in WebKit, so the test can not refuse the calls of the API there
// either (PR-32). So in WebKit, the app stays open, and the function
// gives the same page and false.
export async function reopenOffline(context: BrowserContext, browserName: string): Promise<{ page: Page; stopped: boolean }> {
  if (browserName === "webkit") return { page: context.pages()[0], stopped: false };
  return { page: await stopAndOpen(context), stopped: true };
}
