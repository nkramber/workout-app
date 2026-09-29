import { useEffect, useState } from "react";

import { installPrompt, onInstallPrompt } from "../lib/install-prompt";
import { displayMode } from "../lib/launch";
import { Button, Row, Section, yesNo } from "../ui";

type ManifestInfo = { name: string; display: string; icons: number } | null;

// PR-6, item 3: add the probe to the Home Screen from Chrome. The page
// shows the facts that the install rests on: the manifest, the service
// worker, and the display mode after the install.
export function InstallPage() {
  const [manifest, setManifest] = useState<ManifestInfo | "missing" | "error" | null>(null);
  const [swState, setSwState] = useState("unknown");
  const [controlled, setControlled] = useState<boolean | null>(null);
  const [canPrompt, setCanPrompt] = useState(installPrompt() !== null);
  const [outcome, setOutcome] = useState("");

  useEffect(() => onInstallPrompt(() => setCanPrompt(true)), []);

  useEffect(() => {
    const link = document.querySelector<HTMLLinkElement>('link[rel="manifest"]');
    if (!link) {
      setManifest("missing");
    } else {
      fetch(link.href)
        .then((r) => r.json())
        .then((m) => setManifest({ name: m.name, display: m.display, icons: m.icons?.length ?? 0 }))
        .catch(() => setManifest("error"));
    }

    if (!("serviceWorker" in navigator)) {
      setSwState("no API");
      return;
    }
    setControlled(Boolean(navigator.serviceWorker.controller));
    const onChange = () => setControlled(Boolean(navigator.serviceWorker.controller));
    navigator.serviceWorker.addEventListener("controllerchange", onChange);
    // The worker goes from "activating" to "activated", so the page
    // follows each change of state.
    void navigator.serviceWorker.ready.then((reg) => {
      const worker = reg.active;
      if (!worker) return setSwState("none");
      setSwState(worker.state);
      worker.addEventListener("statechange", () => setSwState(worker.state));
    });
    void navigator.serviceWorker.getRegistration().then((reg) => {
      if (!reg) setSwState("not registered");
    });
    return () => navigator.serviceWorker.removeEventListener("controllerchange", onChange);
  }, []);

  const prompt = async () => {
    const e = installPrompt();
    if (!e) return;
    await e.prompt();
    setOutcome((await e.userChoice).outcome);
  };

  const manifestText =
    manifest === null ? "…" : typeof manifest === "string" ? manifest : `${manifest.name}, ${manifest.display}, ${manifest.icons} icons`;

  return (
    <>
      <Section title="Install">
        <Row label="Display mode" value={displayMode()} testId="install-display-mode" />
        <Row label="Manifest" value={manifestText} testId="install-manifest" />
        <Row label="Service worker" value={swState} testId="install-sw-state" />
        <Row label="Page controlled by the worker" value={yesNo(controlled)} testId="install-sw-controlled" />
        <Row label="Install prompt available" value={yesNo(canPrompt)} testId="install-prompt" />
        {canPrompt && <Button onClick={() => void prompt()}>Install</Button>}
        {outcome && <Row label="Prompt outcome" value={outcome} testId="install-outcome" />}
      </Section>
      <Section title="On the iPhone, in Chrome">
        <ol className="list-decimal space-y-1 pl-5 text-slate-300">
          <li>Tap the share button in the address bar.</li>
          <li>Tap "Add to Home Screen".</li>
          <li>Open the probe from the Home Screen.</li>
          <li>Make sure that the display mode is "standalone".</li>
        </ol>
      </Section>
    </>
  );
}
