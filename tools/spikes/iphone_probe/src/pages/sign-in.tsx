import type { User } from "firebase/auth";
import { useEffect, useState, type FormEvent } from "react";

import { authConfigured, authErrorCode, loadAuth, signIn, signOutOfProbe } from "../lib/firebase";
import { Button, Row, Section, ms } from "../ui";

// PR-6, items 4 and 5: sign in with email and password in the Home Screen
// app, stop the app, open it again, and read the sign-in state. The page
// shows the user id only. It never shows or logs the email (D-80).
// The probe has no form to make an account. The owner makes the account
// in the Firebase console.
export function SignInPage() {
  const configured = authConfigured();
  const [user, setUser] = useState<User | null | undefined>(undefined);
  const [readyMs, setReadyMs] = useState<number | null>(null);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");

  useEffect(() => {
    if (!configured) return;
    let off = () => {};
    loadAuth()
      .then(({ auth, mod, readyMs }) => {
        setReadyMs(readyMs);
        setUser(auth.currentUser);
        off = mod.onAuthStateChanged(auth, setUser);
      })
      .catch((err) => setError(authErrorCode(err)));
    return () => off();
  }, [configured]);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await signIn(email, password);
      setPassword("");
    } catch (err) {
      setError(authErrorCode(err));
    } finally {
      setBusy(false);
    }
  };

  if (!configured) {
    return (
      <Section title="Sign-in">
        <p data-testid="auth-state">not configured</p>
      </Section>
    );
  }

  const state = user === undefined ? "loading" : user ? "signed in" : "signed out";

  return (
    <Section title="Sign-in">
      <Row label="State" value={state} testId="auth-state" />
      <Row label="User id" value={user ? `${user.uid.slice(0, 6)}…` : "n/a"} testId="auth-uid" />
      <Row label="Session read in" value={ms(readyMs)} testId="auth-ready" />
      {error && (
        <p className="text-red-400" data-testid="auth-error">
          {error}
        </p>
      )}
      {user ? (
        <Button onClick={() => void signOutOfProbe()}>Sign out</Button>
      ) : (
        <form className="space-y-2" onSubmit={(e) => void submit(e)}>
          <label className="block">
            <span className="text-slate-400">Email</span>
            <input
              type="email"
              autoComplete="username"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="mt-1 block w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-base"
            />
          </label>
          <label className="block">
            <span className="text-slate-400">Password</span>
            <input
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="mt-1 block w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-base"
            />
          </label>
          <button
            type="submit"
            disabled={busy || user === undefined}
            className="min-h-11 rounded-lg bg-sky-600 px-4 font-medium text-white disabled:opacity-40"
          >
            Sign in
          </button>
        </form>
      )}
    </Section>
  );
}
