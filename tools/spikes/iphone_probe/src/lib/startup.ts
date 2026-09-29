import { addStartup, type Startup } from "./db";
import { displayMode, launchId } from "./launch";

// The startup timer (PR-6, item 6). Each time is in milliseconds from
// the start of the navigation:
//   htmlMs    the first inline script of index.html ran,
//   mainMs    the app bundle started,
//   renderMs  React committed the first render,
//   fcpMs     the browser painted the first content.
// A Home Screen app that comes back from the background does not load
// again, so each record is a cold start or a reload.

function mark(name: string): number | null {
  const entry = performance.getEntriesByName(name, "mark")[0];
  return entry ? Math.round(entry.startTime) : null;
}

function firstContentfulPaint(timeoutMs: number): Promise<number | null> {
  const now = performance.getEntriesByName("first-contentful-paint")[0];
  if (now) return Promise.resolve(Math.round(now.startTime));
  return new Promise((resolve) => {
    let done = false;
    const finish = (v: number | null) => {
      if (done) return;
      done = true;
      observer?.disconnect();
      resolve(v);
    };
    let observer: PerformanceObserver | undefined;
    try {
      observer = new PerformanceObserver((list) => {
        const fcp = list.getEntriesByName("first-contentful-paint")[0];
        if (fcp) finish(Math.round(fcp.startTime));
      });
      observer.observe({ type: "paint", buffered: true });
    } catch {
      finish(null);
    }
    setTimeout(() => finish(null), timeoutMs);
  });
}

let recorded: Promise<Startup> | null = null;

// recordStartup runs once for each launch, after the first render.
export function recordStartup(): Promise<Startup> {
  recorded ??= (async () => {
    performance.mark("probe:render");
    const nav = performance.getEntriesByType("navigation")[0] as PerformanceNavigationTiming | undefined;
    const rec: Startup = {
      at: new Date().toISOString(),
      launchId,
      build: __BUILD_ID__,
      displayMode: displayMode(),
      navigationType: nav?.type ?? "unknown",
      controlled: Boolean(navigator.serviceWorker?.controller),
      htmlMs: mark("probe:html"),
      mainMs: mark("probe:main"),
      renderMs: mark("probe:render"),
      fcpMs: await firstContentfulPaint(3000),
      domContentLoadedMs: nav ? Math.round(nav.domContentLoadedEventEnd) : null,
    };
    await addStartup(rec);
    return rec;
  })();
  return recorded;
}

export function median(values: number[]): number | null {
  if (values.length === 0) return null;
  const s = [...values].sort((a, b) => a - b);
  const mid = Math.floor(s.length / 2);
  return s.length % 2 ? s[mid] : Math.round((s[mid - 1] + s[mid]) / 2);
}
