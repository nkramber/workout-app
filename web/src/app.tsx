import { toJson } from "@bufbuild/protobuf";
import { useQuery } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useLiveQuery } from "dexie-react-hooks";
import type { User } from "firebase/auth";
import { useEffect, useState } from "react";

import { GetProfileResponseSchema, ProfileService } from "./gen/workoutapp/v1/profile_service_pb";
import { db, withReopen } from "./lib/db";
import { isNoConnection, loadErrorText } from "./lib/errors";
import { keepCopy } from "./lib/sync";
import { useSync } from "./lib/sync-engine";
import { authErrorCode, signOutOfApp, watchUser } from "./lib/firebase";
import { useWakeLock } from "./lib/wake-lock";
import { activeWorkout } from "./lib/workout";
import { HomePage } from "./pages/home";
import { InventoryPage } from "./pages/inventory";
import { LockTestPage } from "./pages/lock-test";
import { PlanPage } from "./pages/plan";
import { ErrorText, secondary } from "./pages/inventory/ui";
import { ProfilePage } from "./pages/profile";
import { SignInPage } from "./pages/sign-in";
import { WorkoutPage } from "./pages/workout";
import { Shell } from "./shell";
import { SyncLine } from "./sync-line";

// App shows the sign-in page to a signed-out owner. A signed-in owner gets
// the home screen, and from it the equipment inventory (work area 4.1),
// the profile (work area 5.1), the plan (work area 5.2), and the workout
// (work area 6.1).
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
  return <Shell status={user ? <SyncLine /> : null}>{body}</Shell>;
}

// SignedIn reads the profile first. With no saved profile, it opens the
// onboarding screen before the home screen (D-223). The save puts the
// profile in the query cache, so the home screen then shows. The screen
// wake lock holds from the start of a workout to its end, on each screen
// (D-265). The sync of the outbox runs while the owner is signed in
// (D-277). Each read of the profile goes into its offline copy. With no
// connection, a copy that holds a saved profile opens the home screen,
// so a workout starts with no connection (D-278).
function SignedIn({ onSignOut }: { onSignOut: () => void }) {
  useSync();
  const profile = useQuery(ProfileService.method.getProfile, {});
  const profileCopy = useLiveQuery(() => withReopen(db, () => db.copies.get("profile")), [], undefined);
  const workout = useLiveQuery(() => activeWorkout(db), [], null);
  const wake = useWakeLock(!!workout);
  const [page, setPage] = useState<"home" | "inventory" | "profile" | "plan" | "workout" | "lock-test">("home");
  useEffect(() => {
    if (profile.data?.profile) void keepCopy(db, "profile", toJson(GetProfileResponseSchema, profile.data));
  }, [profile.data]);
  const offline = !!profile.error && isNoConnection(profile.error) && !!(profileCopy?.json as { profile?: unknown } | undefined)?.profile;

  if (profile.error && !offline) {
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
  if (!profile.data && !offline) return <p className="text-slate-400">Loading…</p>;
  if (profile.data && !profile.data.profile) return <ProfilePage onSignOut={onSignOut} />;

  const home = () => setPage("home");
  switch (page) {
    case "inventory":
      return <InventoryPage onBack={home} />;
    case "profile":
      return <ProfilePage onBack={home} />;
    case "plan":
      return <PlanPage onBack={home} />;
    case "workout":
      return <WorkoutPage onBack={home} wake={wake} />;
    case "lock-test":
      return <LockTestPage onBack={home} />;
    case "home":
      return (
        <HomePage
          onSignOut={onSignOut}
          onOpenInventory={() => setPage("inventory")}
          onOpenProfile={() => setPage("profile")}
          onOpenPlan={() => setPage("plan")}
          onOpenWorkout={() => setPage("workout")}
          onOpenLockTest={() => setPage("lock-test")}
        />
      );
  }
}
