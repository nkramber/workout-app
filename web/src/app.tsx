import { useQueryClient } from "@tanstack/react-query";
import type { User } from "firebase/auth";
import { useEffect, useState } from "react";

import { authErrorCode, signOutOfApp, watchUser } from "./lib/firebase";
import { HomePage } from "./pages/home";
import { InventoryPage } from "./pages/inventory";
import { SignInPage } from "./pages/sign-in";
import { Shell } from "./shell";

// App shows the sign-in page to a signed-out owner. A signed-in owner gets
// the home screen, and from it the equipment inventory (work area 4.1).
export function App() {
  const queryClient = useQueryClient();
  const [user, setUser] = useState<User | null | undefined>(undefined);
  const [error, setError] = useState("");
  const [page, setPage] = useState<"home" | "inventory">("home");

  useEffect(() => watchUser(setUser, (err) => setError(authErrorCode(err))), []);

  // A sign-out removes the answers of the API for the last user.
  const signOut = async () => {
    await signOutOfApp();
    queryClient.clear();
    setPage("home");
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
  } else if (page === "inventory") {
    body = <InventoryPage onBack={() => setPage("home")} />;
  } else {
    body = <HomePage onSignOut={() => void signOut()} onOpenInventory={() => setPage("inventory")} />;
  }
  return <Shell>{body}</Shell>;
}
