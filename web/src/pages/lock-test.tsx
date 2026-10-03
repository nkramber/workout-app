import { useEffect, useRef, useState } from "react";

import { logLine, METHODS, probeVideo, probeWakeLock, silentSource, type Method } from "../lib/lock-test";
import { ErrorText, primary, secondary, Title } from "./inventory/ui";

// LockTestPage is the "Screen lock test" screen (D-280, D-282). It runs one
// method of src/lib/lock-test.ts at a time, and shows its log. A method
// stops when the owner taps "Stop" or leaves the screen.
export function LockTestPage({ onBack }: { onBack: () => void }) {
  const [running, setRunning] = useState<Method | null>(null);
  const [lines, setLines] = useState<string[]>([]);
  const [error, setError] = useState("");
  const stopRef = useRef<(() => void) | null>(null);
  const video = useRef<HTMLVideoElement>(null);

  const stop = () => {
    stopRef.current?.();
    stopRef.current = null;
    setRunning(null);
  };
  useEffect(() => () => stopRef.current?.(), []);

  const start = async (method: Method) => {
    stop();
    setError("");
    const t0 = Date.now();
    const log = (text: string) => setLines((old) => [...old, logLine(t0, Date.now(), text)]);
    setLines([logLine(t0, t0, `Start: ${METHODS.find((m) => m.id === method)?.title}.`)]);
    setRunning(method);
    try {
      if (method === "wake-lock") {
        stopRef.current = probeWakeLock(navigator as never, document, window, log);
        return;
      }
      const v = video.current;
      if (!v) return;
      const offSource = await silentSource(v, method);
      const offProbe = probeVideo(v, document, window, log);
      stopRef.current = () => {
        offProbe();
        offSource();
      };
    } catch (err) {
      setError(`The method did not start (${err instanceof Error ? err.name : String(err)}).`);
      setRunning(null);
    }
  };

  return (
    <div className="space-y-6">
      <Title onBack={onBack}>Screen lock test</Title>
      <ol className="list-decimal space-y-1 pl-5 text-sm text-slate-300" data-testid="lock-test-steps">
        <li>Set Auto-Lock of the iPhone to 30 seconds.</li>
        <li>Note the battery level, and start one method.</li>
        <li>Leave the app, come back, and do not tap the screen.</li>
        <li>Watch the screen for one minute or more, then note the battery level.</li>
      </ol>
      <ErrorText testId="lock-test-error">{error}</ErrorText>
      {METHODS.map((m) => (
        <section key={m.id} className="space-y-2 rounded-lg border border-slate-800 p-3" aria-label={m.title}>
          <h3 className="text-base font-semibold text-slate-100">{m.title}</h3>
          <p className="text-sm text-slate-400">{m.text}</p>
          {running === m.id ? (
            <button type="button" className={`${secondary} w-full`} onClick={stop}>
              {`Stop ${m.title}`}
            </button>
          ) : (
            <button type="button" className={`${primary} w-full`} onClick={() => void start(m.id)}>
              {`Start ${m.title}`}
            </button>
          )}
        </section>
      ))}
      <video ref={video} muted playsInline loop aria-hidden="true" className="h-px w-px opacity-[0.01]" data-testid="lock-test-video" />
      <section className="space-y-2" aria-label="Log">
        <h3 className="text-sm font-semibold text-slate-300">Log</h3>
        {lines.length === 0 && <p className="text-sm text-slate-400">No method ran yet.</p>}
        <ol className="space-y-1 font-mono text-xs text-slate-300" data-testid="lock-test-log">
          {lines.map((l, i) => (
            <li key={i}>{l}</li>
          ))}
        </ol>
      </section>
    </div>
  );
}
