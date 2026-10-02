import { expect, type APIRequestContext, type BrowserContext, type Page } from "@playwright/test";

// The shared steps of the browser tests. Each test makes its own account
// on the Auth emulator, so the tests can run in parallel.

const authHost = process.env.FIREBASE_AUTH_EMULATOR_HOST;
const firestoreHost = process.env.FIRESTORE_EMULATOR_HOST;
const project = "demo-workout-app";
const password = "emulator-only-1";

// The Auth emulator makes the account, so the app needs no form for it.
export async function makeAccount(request: APIRequestContext, email: string): Promise<string> {
  const res = await request.post(`http://${authHost}/identitytoolkit.googleapis.com/v1/accounts:signUp?key=demo-key`, {
    data: { email, password, returnSecureToken: true },
  });
  expect(res.ok(), await res.text()).toBe(true);
  return (await res.json()).localId;
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
