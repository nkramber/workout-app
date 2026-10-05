import { Code, ConnectError } from "@connectrpc/connect";
import { useQuery } from "@connectrpc/connect-query";
import { useLiveQuery } from "dexie-react-hooks";
import { useEffect, useId, useState, type ReactNode } from "react";

import { UserService } from "../gen/workoutapp/v1/user_service_pb";
import { db } from "../lib/db";
import { canDelete, DELETE_CONFIRMATION } from "../lib/history";
import { activeWorkout } from "../lib/workout";
import { megabytes, requestPersistenceOnce, type StorageState } from "../lib/storage";
import { useUserApi } from "../lib/user-api";
import { danger, field, secondary } from "./inventory/ui";

// The home screen proves the whole path from the sign-in to the API: it
// calls GetMe, and it shows the uid that the API read from the token. The
// uid is an id, so the screen can show it (D-80). It opens the workout
// (work area 6.1), the plan (work area 5.2), the equipment inventory (work
// area 4.1), and the profile (work area 5.1). An open workout on the
// phone gives the button "Continue workout". The diagnostics rows serve the device check
// of PR-11, and they hold the deletion of all data (D-314).
export function HomePage({
  onSignOut,
  onOpenInventory,
  onOpenProfile,
  onOpenPlan,
  onOpenWorkout,
}: {
  onSignOut: () => void;
  onOpenInventory: () => void;
  onOpenProfile: () => void;
  onOpenPlan: () => void;
  onOpenWorkout: () => void;
}) {
  const me = useQuery(UserService.method.getMe, {});
  const pending = useLiveQuery(() => db.outbox.count(), [], null);
  const active = useLiveQuery(() => activeWorkout(db), [], null);
  const [storage, setStorage] = useState<StorageState | null>(null);

  // The first sign-in on this device asks for persistent storage (REC-5,
  // D-134). A later start reads the state alone.
  useEffect(() => {
    let live = true;
    requestPersistenceOnce(db)
      .then((s) => live && setStorage(s))
      .catch(() => live && setStorage({ persisted: null, requestedAt: null, usage: null, quota: null }));
    return () => {
      live = false;
    };
  }, []);

  return (
    <div className="space-y-6">
      <section className="space-y-2">
        <h2 className="text-base font-semibold text-slate-100">Home</h2>
        {me.isPending && <p className="text-slate-400">Loading…</p>}
        {me.data && (
          <Row label="User id" testId="me-uid">
            {me.data.uid}
          </Row>
        )}
        {me.error && (
          <p role="alert" className="text-sm text-red-400">
            {errorText(me.error)} <span data-testid="me-error">{Code[ConnectError.from(me.error).code]}</span>
          </p>
        )}
      </section>

      <button
        type="button"
        onClick={onOpenWorkout}
        className="min-h-14 w-full rounded-lg bg-sky-600 px-4 text-lg font-medium text-white active:bg-sky-700"
      >
        {active ? "Continue workout" : "Workout"}
      </button>
      <button
        type="button"
        onClick={onOpenPlan}
        className="min-h-11 w-full rounded-lg border border-slate-700 px-4 font-medium text-slate-100 active:bg-slate-800"
      >
        Plan
      </button>
      <button
        type="button"
        onClick={onOpenInventory}
        className="min-h-11 w-full rounded-lg border border-slate-700 px-4 font-medium text-slate-100 active:bg-slate-800"
      >
        Equipment
      </button>
      <button
        type="button"
        onClick={onOpenProfile}
        className="min-h-11 w-full rounded-lg border border-slate-700 px-4 font-medium text-slate-100 active:bg-slate-800"
      >
        Profile
      </button>

      <section className="space-y-2 text-sm">
        <h2 className="text-base font-semibold text-slate-100">Diagnostics</h2>
        <Row label="Build" testId="build-id">
          {__BUILD_ID__}
        </Row>
        <Row label="Changes waiting to sync" testId="outbox-count">
          {pending ?? "…"}
        </Row>
        <Row label="Storage kept" testId="storage-persisted">
          {storage === null ? "…" : storage.persisted === null ? "unknown" : storage.persisted ? "yes" : "no"}
        </Row>
        <Row label="Storage used" testId="storage-usage">
          {storage === null ? "…" : `${megabytes(storage.usage)} of ${megabytes(storage.quota)}`}
        </Row>
        <DeleteData />
      </section>

      <button
        type="button"
        onClick={onSignOut}
        className="min-h-11 rounded-lg border border-slate-700 px-4 font-medium text-slate-100 active:bg-slate-800"
      >
        Sign out
      </button>
    </div>
  );
}

// DeleteData is the deletion of all data (D-314, D-315). It is hard to
// reach: a closed section holds its button, and the button opens a
// dialog. The button of the dialog stays off until the switch is at
// "Yes" and the owner types "Delete all data".
function DeleteData() {
  const api = useUserApi();
  const [open, setOpen] = useState(false);
  const [yes, setYes] = useState(false);
  const [text, setText] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState<string | null>(null);
  const titleId = useId();
  const switchId = useId();
  const inputId = useId();
  const reset = () => {
    setYes(false);
    setText("");
    setError(null);
  };
  const close = () => {
    if (busy) return;
    setOpen(false);
    reset();
  };
  const submit = async () => {
    if (!canDelete(yes, text) || busy) return;
    setBusy(true);
    setError(null);
    try {
      const n = await api.deleteAllData();
      setDone(`The phone and the server deleted your history: ${n} ${n === 1 ? "workout" : "workouts"}.`);
      setOpen(false);
      reset();
    } catch (err) {
      setError(deleteErrorText(err));
    } finally {
      setBusy(false);
    }
  };

  return (
    <details className="pt-2" data-testid="delete-data">
      <summary className="cursor-pointer text-slate-500">Delete all data</summary>
      <div className="space-y-2 pt-2">
        <p className="text-slate-400">This deletes your workouts and your plan. Your profile and your machines stay.</p>
        <button
          type="button"
          onClick={() => {
            setDone(null);
            setOpen(true);
          }}
          className="rounded-lg border border-red-900 px-3 py-1 text-sm text-red-300 active:bg-red-950"
        >
          Delete all data…
        </button>
        {done && (
          <p className="text-slate-300" data-testid="delete-done">
            {done}
          </p>
        )}
      </div>
      {open && (
        <div className="fixed inset-0 z-20 flex items-end justify-center bg-black/70 p-4 sm:items-center">
          <div
            role="alertdialog"
            aria-modal="true"
            aria-labelledby={titleId}
            className="max-h-full w-full max-w-md space-y-4 overflow-y-auto rounded-lg border border-red-900 bg-slate-900 p-4"
            onKeyDown={(e) => {
              if (e.key === "Escape") close();
            }}
          >
            <h3 id={titleId} className="text-lg font-semibold text-slate-100">
              Delete all data
            </h3>
            <p className="text-sm text-slate-300" data-testid="delete-text">
              This deletes each workout, each logged set, and the plan with its history, on this phone and on the server. Your profile, your
              machines, and your exclusions stay. A backup keeps a copy for up to 10 days.
            </p>
            <div className="flex items-center justify-between gap-4">
              <span id={switchId} className="text-sm text-slate-200">
                I want to delete all my history
              </span>
              <button
                type="button"
                role="switch"
                aria-checked={yes}
                aria-labelledby={switchId}
                disabled={busy}
                onClick={() => setYes(!yes)}
                className={`min-h-11 min-w-16 rounded-lg border px-3 font-medium ${yes ? "border-red-700 bg-red-900 text-red-100" : "border-slate-700 text-slate-300"}`}
              >
                {yes ? "Yes" : "No"}
              </button>
            </div>
            <label htmlFor={inputId} className="block text-sm text-slate-200">
              {`Type ${DELETE_CONFIRMATION} to confirm`}
              <input
                id={inputId}
                className={field}
                value={text}
                disabled={busy}
                onChange={(e) => setText(e.target.value)}
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
                data-testid="delete-confirmation"
              />
            </label>
            {error && (
              <p role="alert" className="text-sm text-red-300" data-testid="delete-error">
                {error}
              </p>
            )}
            <div className="flex gap-2">
              <button type="button" className={secondary} disabled={busy} onClick={close}>
                Cancel
              </button>
              <button type="button" className={danger} disabled={!canDelete(yes, text) || busy} onClick={() => void submit()}>
                {busy ? "Deleting…" : "Delete"}
              </button>
            </div>
          </div>
        </div>
      )}
    </details>
  );
}

// deleteErrorText gives the error of a deletion. The phone deletes its
// copy first, so a failed call leaves the history on the server alone.
function deleteErrorText(err: unknown): string {
  switch (ConnectError.from(err).code) {
    case Code.InvalidArgument:
      return "The server did not accept the confirmation. Nothing changed on the server.";
    case Code.Unavailable:
    case Code.Unknown:
      return "The phone deleted its copy, but the server did not delete your history. Connect and try again.";
    default:
      return "The phone deleted its copy, but the API did not answer. Try again.";
  }
}

function errorText(err: unknown): string {
  switch (ConnectError.from(err).code) {
    case Code.PermissionDenied:
      return "This account is not on the allowlist.";
    case Code.Unauthenticated:
      return "The API did not accept the sign-in.";
    default:
      return "The API did not answer.";
  }
}

function Row({ label, testId, children }: { label: string; testId: string; children: ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-4 border-b border-slate-800 pb-1">
      <span className="text-slate-400">{label}</span>
      <span className="min-w-0 text-right font-mono break-all text-slate-100" data-testid={testId}>
        {children}
      </span>
    </div>
  );
}
