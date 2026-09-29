// One launch is one load of the page. On the phone, a new launch after a
// stop of the app is a cold start. Each record holds the launch id, so a
// page shows which records came from an earlier launch.
export const launchId = Math.random().toString(36).slice(2, 10);

// displayMode says how the page runs: in a browser tab, or as a Home
// Screen app. iOS gives `navigator.standalone`. Other browsers give the
// display-mode media query.
export function displayMode(): "standalone" | "browser" {
  const nav = navigator as Navigator & { standalone?: boolean };
  if (nav.standalone === true) return "standalone";
  if (window.matchMedia?.("(display-mode: standalone)").matches) return "standalone";
  return "browser";
}
