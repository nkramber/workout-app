import { useEffect, useRef, useState } from "react";

import { Button, Row, Section } from "../ui";

// PR-6, item 2: keep the screen on with Screen Wake Lock. The browser
// releases the lock when the page goes to the background, so the page
// requests it again when the page comes back, while the owner wants it.
// The timer shows how long the current lock is held.
export function WakeLockPage() {
  const supported = typeof navigator !== "undefined" && "wakeLock" in navigator;
  const sentinel = useRef<WakeLockSentinel | null>(null);
  const [wanted, setWanted] = useState(false);
  const [state, setState] = useState<"released" | "active" | "error">("released");
  const [error, setError] = useState("");
  const [since, setSince] = useState<number | null>(null);
  const [now, setNow] = useState(Date.now());
  const [log, setLog] = useState<string[]>([]);

  const note = (line: string) => setLog((l) => [`${new Date().toISOString().slice(11, 19)} ${line}`, ...l].slice(0, 20));

  const acquire = async () => {
    if (!supported) return;
    try {
      const s = await navigator.wakeLock.request("screen");
      sentinel.current = s;
      setState("active");
      setError("");
      setSince(Date.now());
      note("lock acquired");
      s.addEventListener("release", () => {
        setState("released");
        setSince(null);
        note("lock released");
      });
    } catch (err) {
      setState("error");
      setError(err instanceof Error ? `${err.name}: ${err.message}` : String(err));
      note("request failed");
    }
  };

  const release = async () => {
    setWanted(false);
    await sentinel.current?.release();
    sentinel.current = null;
  };

  useEffect(() => {
    if (!wanted) return;
    const onVisible = () => {
      if (document.visibilityState === "visible" && sentinel.current?.released !== false) void acquire();
    };
    document.addEventListener("visibilitychange", onVisible);
    return () => document.removeEventListener("visibilitychange", onVisible);
    // acquire reads refs and setters only, so it is not a dependency.
  }, [wanted]);

  useEffect(() => {
    const t = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(t);
  }, []);

  useEffect(() => () => void sentinel.current?.release(), []);

  return (
    <Section title="Screen Wake Lock">
      <Row label="API present" value={supported ? "yes" : "no"} testId="wake-supported" />
      <Row label="Lock state" value={state} testId="wake-state" />
      <Row label="Held for" value={since ? `${Math.floor((now - since) / 1000)} s` : "n/a"} testId="wake-held" />
      {error && (
        <p className="text-red-400" data-testid="wake-error">
          {error}
        </p>
      )}
      <div className="flex flex-wrap gap-2">
        <Button
          disabled={!supported}
          onClick={() => {
            setWanted(true);
            void acquire();
          }}
        >
          Keep the screen on
        </Button>
        <Button disabled={!supported} onClick={() => void release()}>
          Release
        </Button>
      </div>
      <ul className="space-y-1 font-mono text-xs text-slate-400" data-testid="wake-log">
        {log.map((l, i) => (
          <li key={i}>{l}</li>
        ))}
      </ul>
    </Section>
  );
}
