import { useEffect, useState } from "react";

import { clearStartups, listStartups, type Startup } from "../lib/db";
import { launchId } from "../lib/launch";
import { median, recordStartup } from "../lib/startup";
import { Button, Row, Section, ms } from "../ui";

// PR-6, item 6: the startup time of the React build (D-84). The page
// shows this launch, and the median of the first contentful paint for
// each display mode over the stored launches.
export function StartupPage() {
  const [current, setCurrent] = useState<Startup | null>(null);
  const [all, setAll] = useState<Startup[]>([]);

  const refresh = async () => {
    setCurrent(await recordStartup());
    setAll(await listStartups());
  };

  useEffect(() => {
    void refresh();
  }, []);

  const fcpOf = (mode: string) =>
    all.filter((s) => s.displayMode === mode && s.fcpMs !== null).map((s) => s.fcpMs as number);
  const standalone = fcpOf("standalone");
  const browser = fcpOf("browser");

  return (
    <>
      <Section title="This launch">
        <Row label="Inline script of the page" value={ms(current?.htmlMs)} testId="startup-html" />
        <Row label="App bundle start" value={ms(current?.mainMs)} testId="startup-main" />
        <Row label="First React render" value={ms(current?.renderMs)} testId="startup-render" />
        <Row label="First contentful paint" value={ms(current?.fcpMs)} testId="startup-fcp" />
        <Row label="DOMContentLoaded end" value={ms(current?.domContentLoadedMs)} testId="startup-dcl" />
        <Row label="Navigation type" value={current?.navigationType ?? "…"} testId="startup-nav" />
        <Row label="Served by the worker" value={current ? (current.controlled ? "yes" : "no") : "…"} testId="startup-controlled" />
      </Section>
      <Section title="All stored launches">
        <Row label="Launches" value={all.length} testId="startup-count" />
        <Row label={`Median paint, standalone (${standalone.length})`} value={ms(median(standalone))} testId="startup-median-standalone" />
        <Row label={`Median paint, browser (${browser.length})`} value={ms(median(browser))} testId="startup-median-browser" />
        <table className="w-full font-mono text-xs">
          <thead className="text-slate-400">
            <tr>
              <th className="text-left">time</th>
              <th className="text-left">mode</th>
              <th className="text-right">render</th>
              <th className="text-right">paint</th>
            </tr>
          </thead>
          <tbody data-testid="startup-list">
            {[...all].reverse().slice(0, 30).map((s) => (
              <tr key={s.id} className={s.launchId === launchId ? "text-sky-400" : ""}>
                <td>{s.at.slice(5, 19).replace("T", " ")}</td>
                <td>{s.displayMode}</td>
                <td className="text-right">{s.renderMs ?? "n/a"}</td>
                <td className="text-right">{s.fcpMs ?? "n/a"}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <Button onClick={() => void clearStartups().then(refresh)}>Clear</Button>
      </Section>
    </>
  );
}
