import { useEffect, useState } from "react";

import "./lib/install-prompt";
import { displayMode, launchId } from "./lib/launch";
import { recordStartup } from "./lib/startup";
import { InstallPage } from "./pages/install";
import { SignInPage } from "./pages/sign-in";
import { StartupPage } from "./pages/startup";
import { StoragePage } from "./pages/storage";
import { WakeLockPage } from "./pages/wake-lock";

// One page for each device item of the checklist of PR-6. The probe has
// no camera page (D-112). The route lives in the hash, so each URL of
// the probe serves the same index.html from Firebase Hosting.
const pages = [
  { id: "storage", label: "Storage", Page: StoragePage },
  { id: "wake-lock", label: "Wake", Page: WakeLockPage },
  { id: "install", label: "Install", Page: InstallPage },
  { id: "sign-in", label: "Sign-in", Page: SignInPage },
  { id: "startup", label: "Startup", Page: StartupPage },
] as const;

type PageId = (typeof pages)[number]["id"];

function routeFromHash(): PageId {
  const id = window.location.hash.replace(/^#\/?/, "");
  return pages.some((p) => p.id === id) ? (id as PageId) : "storage";
}

export function App() {
  const [route, setRoute] = useState<PageId>(routeFromHash);

  useEffect(() => {
    const onHash = () => setRoute(routeFromHash());
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, []);

  useEffect(() => {
    void recordStartup();
  }, []);

  const { Page, label } = pages.find((p) => p.id === route)!;

  return (
    <div className="mx-auto flex h-full max-w-md flex-col">
      <header className="px-4 pt-[max(env(safe-area-inset-top),1rem)] pb-3">
        <h1 className="text-lg font-bold">Gym Route probe</h1>
        <p className="font-mono text-xs text-slate-400">
          build <span data-testid="build-id">{__BUILD_ID__}</span> · launch <span data-testid="launch-id">{launchId}</span> ·{" "}
          <span data-testid="display-mode">{displayMode()}</span>
        </p>
      </header>
      <main className="flex-1 space-y-4 overflow-y-auto px-4 pb-4" aria-label={label}>
        <Page />
      </main>
      <nav className="grid grid-cols-5 border-t border-slate-800 bg-slate-950 pb-[env(safe-area-inset-bottom)]" aria-label="Probe pages">
        {pages.map((p) => (
          <a
            key={p.id}
            href={`#/${p.id}`}
            aria-current={p.id === route ? "page" : undefined}
            className={`py-3 text-center text-xs ${p.id === route ? "font-semibold text-sky-400" : "text-slate-400"}`}
          >
            {p.label}
          </a>
        ))}
      </nav>
    </div>
  );
}
