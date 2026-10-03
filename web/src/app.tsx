import { useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import type { User } from "firebase/auth";
import { useEffect, useState } from "react";

import { ProfileService } from "./gen/workoutapp/v1/profile_service_pb";
import { loadErrorText } from "./lib/errors";
import { authErrorCode, signOutOfApp, watchUser } from "./lib/firebase";
import { HomePage } from "./pages/home";
import { InventoryPage } from "./pages/inventory";
import { PlanPage } from "./pages/plan";
import { ErrorText, secondary } from "./pages/inventory/ui";
import { ProfilePage } from "./pages/profile";
import { SignInPage } from "./pages/sign-in";
import { Shell } from "./shell";

// App shows the sign-in page to a signed-out owner. A signed-in owner gets
// the home screen, and from it the equipment inventory (work area 4.1),
// the profile (work area 5.1), and the plan (work area 5.2).
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
    body = <SignedIn key={user.uid} onSignOut={() => void signOut()} />;
  }
  return <Shell>{body}</Shell>;
}

// SignedIn reads the profile first. With no saved profile, it opens the
// onboarding screen before the home screen (D-223). The save puts the
// profile in the query cache, so the home screen then shows.
function SignedIn({ onSignOut }: { onSignOut: () => void }) {
  const profile = useQuery(ProfileService.method.getProfile, {});
  const [page, setPage] = useState<"home" | "inventory" | "profile" | "plan">("home");

  if (profile.error) {
    return (
      <div className="space-y-4">
        <ErrorText testId="profile-load-error">{loadErrorText(profile.error)}</ErrorText>
        <div className="flex gap-3">
          <button type="button" className={secondary} onClick={() => void profile.refetch()}>
            Try again
          </button>
          <button type="button" className={secondary} onClick={onSignOut}>
            Sign out
          </button>
        </div>
      </div>
    );
  }
  if (!profile.data) return <p className="text-slate-400">Loading…</p>;
  if (!profile.data.profile) return <ProfilePage onSignOut={onSignOut} />;

  const home = () => setPage("home");
  switch (page) {
    case "inventory":
      return <InventoryPage onBack={home} />;
    case "profile":
      return <ProfilePage onBack={home} />;
    case "plan":
      return <PlanPage onBack={home} />;
    case "home":
      return (
        <HomePage
          onSignOut={onSignOut}
          onOpenInventory={() => setPage("inventory")}
          onOpenProfile={() => setPage("profile")}
          onOpenPlan={() => setPage("plan")}
        />
      );
  }
}
