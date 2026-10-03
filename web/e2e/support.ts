import { expect, type APIRequestContext, type BrowserContext, type Page } from "@playwright/test";

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
