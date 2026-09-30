import type { Auth, User } from "firebase/auth";

import { projectConfig } from "./firebase-config";

// Firebase Auth loads after the first paint, as in Decktome. No module
// imports the SDK at module scope, so the first chunk carries no Auth
// code. Each caller awaits loadAuth().
type AuthModule = typeof import("firebase/auth");

// The browser tests build the app with VITE_AUTH_EMULATOR_HOST, and the
// app then talks to the local Auth emulator of `firebase-tools` (D-115).
// `npm run dev` uses the emulator of firebase.json too. The `demo-`
// prefix tells the emulator that no real project exists, and the Go API
// of the tests reads the same project id.
const emulatorHost = import.meta.env.VITE_AUTH_EMULATOR_HOST ?? (import.meta.env.DEV ? "127.0.0.1:9299" : "");
const emulatorConfig = { apiKey: "demo-key", projectId: "demo-workout-app", authDomain: "localhost" };

let pending: Promise<{ auth: Auth; mod: AuthModule }> | null = null;

// loadAuth downloads the SDK once, and returns the same instance after
// that. It resolves after the SDK reads the stored session.
export function loadAuth(): Promise<{ auth: Auth; mod: AuthModule }> {
  pending ??= start();
  return pending;
}

async function start() {
  const [{ initializeApp }, mod] = await Promise.all([import("firebase/app"), import("firebase/auth")]);
  const app = initializeApp(emulatorHost ? emulatorConfig : projectConfig);
  // initializeAuth with no popup and no redirect resolver keeps the
  // bundle small. Email and password use neither (platform research,
  // section 5.5). The session stays in IndexedDB after a stop of the
  // Home Screen app, as the probe proved.
  const auth = mod.initializeAuth(app, {
    persistence: [mod.indexedDBLocalPersistence, mod.browserLocalPersistence],
  });
  if (emulatorHost) mod.connectAuthEmulator(auth, `http://${emulatorHost}`, { disableWarnings: true });
  await auth.authStateReady();
  return { auth, mod };
}

// watchUser calls onUser with the current user, then with each change.
// It returns the function that stops the watch.
export function watchUser(onUser: (user: User | null) => void, onError: (err: unknown) => void): () => void {
  let off = () => {};
  let stopped = false;
  loadAuth()
    .then(({ auth, mod }) => {
      if (stopped) return;
      off = mod.onAuthStateChanged(auth, onUser);
    })
    .catch(onError);
  return () => {
    stopped = true;
    off();
  };
}

// A token refresh that fails with one of these codes can never succeed
// again, so the app signs out, as Decktome does.
const deadSessionCodes = new Set(["auth/user-token-expired", "auth/user-disabled"]);

// currentIdToken gives the ID token of the signed-in user, or an empty
// string when nobody is signed in.
export async function currentIdToken(): Promise<string> {
  const { auth, mod } = await loadAuth();
  const user = auth.currentUser;
  if (!user) return "";
  try {
    return await user.getIdToken();
  } catch (err) {
    if (deadSessionCodes.has(authErrorCode(err))) {
      await mod.signOut(auth);
      return "";
    }
    throw err;
  }
}

// signIn serves the sign-in form. The app has no form that makes an
// account: the owner makes the one account in the Firebase console, and
// self sign-up is off (D-117).
export async function signIn(email: string, password: string): Promise<void> {
  const { auth, mod } = await loadAuth();
  await mod.signInWithEmailAndPassword(auth, email, password);
}

export async function signOutOfApp(): Promise<void> {
  const { auth, mod } = await loadAuth();
  await mod.signOut(auth);
}

// authErrorCode gives the Firebase code of an error, such as
// `auth/invalid-credential`. The app shows the code and never the email.
export function authErrorCode(err: unknown): string {
  if (err && typeof err === "object" && "code" in err && typeof err.code === "string") return err.code;
  return err instanceof Error ? err.message : "unknown error";
}
