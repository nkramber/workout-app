import type { Auth } from "firebase/auth";

import { projectConfig } from "./firebase-config";

// Firebase Auth loads after the first paint, as in Decktome. No module
// imports the SDK at module scope, so the startup timer measures the
// shell alone. Each caller awaits loadAuth().
type AuthModule = typeof import("firebase/auth");

// The browser tests build the probe with VITE_AUTH_EMULATOR_HOST, and the
// probe then talks to the local Auth emulator of `firebase-tools` (D-115).
// The `demo-` prefix tells the emulator that no real project exists.
const emulatorHost = import.meta.env.VITE_AUTH_EMULATOR_HOST ?? "";
const emulatorConfig = { apiKey: "demo-key", projectId: "demo-gym-route", authDomain: "localhost" };

export function authConfigured(): boolean {
  return Boolean(emulatorHost) || projectConfig !== null;
}

let pending: Promise<{ auth: Auth; mod: AuthModule; readyMs: number }> | null = null;

// loadAuth downloads the SDK once. readyMs is the time from the call to
// the moment the SDK read the stored session. The device checklist reads
// it after a stop of the app (PR-6, item 5).
export function loadAuth(): Promise<{ auth: Auth; mod: AuthModule; readyMs: number }> {
  pending ??= start();
  return pending;
}

async function start() {
  const t0 = performance.now();
  const config = emulatorHost ? emulatorConfig : projectConfig;
  if (!config) throw new Error("No Firebase project is configured.");
  const [{ initializeApp }, mod] = await Promise.all([import("firebase/app"), import("firebase/auth")]);
  const app = initializeApp(config);
  // IndexedDB first, the default of the web SDK. The probe states it,
  // because the session must stay after a stop of the Home Screen app.
  const auth = mod.initializeAuth(app, {
    persistence: [mod.indexedDBLocalPersistence, mod.browserLocalPersistence],
  });
  if (emulatorHost) mod.connectAuthEmulator(auth, `http://${emulatorHost}`, { disableWarnings: true });
  await auth.authStateReady();
  return { auth, mod, readyMs: Math.round(performance.now() - t0) };
}

export async function signIn(email: string, password: string): Promise<void> {
  const { auth, mod } = await loadAuth();
  await mod.signInWithEmailAndPassword(auth, email, password);
}

export async function signOutOfProbe(): Promise<void> {
  const { auth, mod } = await loadAuth();
  await mod.signOut(auth);
}

// authErrorCode gives the Firebase code of an error, such as
// `auth/invalid-credential`. The page shows the code and never the email.
export function authErrorCode(err: unknown): string {
  if (err && typeof err === "object" && "code" in err && typeof err.code === "string") return err.code;
  return err instanceof Error ? err.message : "unknown error";
}
