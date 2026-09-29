import { useState, type FormEvent } from "react";

import { authErrorCode, signIn } from "../lib/firebase";

const field = "mt-1 block w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-base";

// The sign-in form: email and password on Firebase Authentication (D-75).
// The app has no form that makes an account. The owner makes the one
// account in the Firebase console, and self sign-up is off (D-117). The
// page shows the Firebase error code, and never the email (D-80).
export function SignInPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async (e: FormEvent) => {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      await signIn(email, password);
    } catch (err) {
      setError(authErrorCode(err));
      setBusy(false);
    }
  };

  return (
    <form aria-label="Sign in" className="space-y-4" onSubmit={(e) => void submit(e)}>
      <h2 className="text-base font-semibold text-slate-100">Sign in</h2>
      <label className="block">
        <span className="text-sm text-slate-400">Email</span>
        <input
          type="email"
          autoComplete="username"
          required
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className={field}
        />
      </label>
      <label className="block">
        <span className="text-sm text-slate-400">Password</span>
        <input
          type="password"
          autoComplete="current-password"
          required
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className={field}
        />
      </label>
      {error && (
        <p role="alert" className="text-sm text-red-400" data-testid="sign-in-error">
          {error}
        </p>
      )}
      <button
        type="submit"
        disabled={busy}
        className="min-h-11 w-full rounded-lg bg-sky-600 px-4 font-medium text-white active:bg-sky-700 disabled:opacity-40"
      >
        Sign in
      </button>
    </form>
  );
}
