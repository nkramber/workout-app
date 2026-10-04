// The local date of the owner. A module of its own, so the sync and the
// plan screen read it with no import cycle.

// planRequest gives the request of GetPlan with the local date, so each
// target is the target of the rules on the date of the next session
// (D-151, D-179, D-294, D-295).
export function planRequest(now = new Date()): { today: string } {
  return { today: localDate(now) };
}

// localDate gives the local date of the owner as YYYY-MM-DD, the "today"
// of a plan request.
export function localDate(now = new Date()): string {
  const pad = (n: number) => String(n).padStart(2, "0");
  return `${now.getFullYear()}-${pad(now.getMonth() + 1)}-${pad(now.getDate())}`;
}
