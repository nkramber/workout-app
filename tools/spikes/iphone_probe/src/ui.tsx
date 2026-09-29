import type { ReactNode } from "react";

export function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="rounded-xl border border-slate-700 bg-slate-900 p-4">
      <h2 className="mb-3 text-base font-semibold text-slate-100">{title}</h2>
      <div className="space-y-3 text-sm">{children}</div>
    </section>
  );
}

// Row shows one measured value. The test id is the handle of the
// browser tests and the name that the report of PR-6 uses.
export function Row({ label, value, testId }: { label: string; value: ReactNode; testId: string }) {
  return (
    <div className="flex items-baseline justify-between gap-4 border-b border-slate-800 pb-1 last:border-0">
      <span className="text-slate-400">{label}</span>
      <span className="text-right font-mono text-slate-100" data-testid={testId}>
        {value}
      </span>
    </div>
  );
}

export function Button({ children, onClick, disabled }: { children: ReactNode; onClick: () => void; disabled?: boolean }) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="min-h-11 rounded-lg bg-sky-600 px-4 font-medium text-white active:bg-sky-700 disabled:opacity-40"
    >
      {children}
    </button>
  );
}

export function yesNo(v: boolean | null | undefined): string {
  if (v === null || v === undefined) return "unknown";
  return v ? "yes" : "no";
}

export function ms(v: number | null | undefined): string {
  return v === null || v === undefined ? "n/a" : `${v} ms`;
}
