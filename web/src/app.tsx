import { useQueryClient } from "@tanstack/react-query";
import type { User } from "firebase/auth";
import { useEffect, useState } from "react";

import { authErrorCode, signOutOfApp, watchUser } from "./lib/firebase";
import { HomePage } from "./pages/home";
import { SignInPage } from "./pages/sign-in";
import { Shell } from "./shell";

// App shows the sign-in page to a signed-out owner and the home screen to
// a signed-in owner. Phase 2 has no other screen.
export function App() {
  const queryClient = useQueryClient();
  const [user, setUser] = useState<User | null | undefined>(undefined);
  const [error, setError] = useState("");

  useEffect(() => watchUser(setUser, (err) => setError(authErrorCode(err))), []);

  // A sign-out removes the answers of the API for the last user.
  const signOut = async () => {
    await signOutOfApp();
    queryClient.clear();
  };

  let body;
  if (error) {
    body = (
      <p role="alert" className="text-red-400" data-testid="auth-load-error">
        {error}
      </p>
    );
  } else if (user === undefined) {
    body = <p className="text-slate-400">Loading…</p>;
  } else if (user === null) {
    body = <SignInPage />;
  } else {
    body = <HomePage onSignOut={() => void signOut()} />;
  }
  return <Shell>{body}</Shell>;
}
