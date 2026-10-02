import { Code, ConnectError } from "@connectrpc/connect";
import { useQuery } from "@connectrpc/connect-query";
import { useLiveQuery } from "dexie-react-hooks";
import { useEffect, useState, type ReactNode } from "react";

import { UserService } from "../gen/workoutapp/v1/user_service_pb";
import { db } from "../lib/db";
import { megabytes, requestPersistenceOnce, type StorageState } from "../lib/storage";

// The home screen proves the whole path from the sign-in to the API: it
// calls GetMe, and it shows the uid that the API read from the token. The
// uid is an id, so the screen can show it (D-80). It opens the equipment
// inventory (work area 4.1). The diagnostics rows serve the device check
// of PR-11.
export function HomePage({ onSignOut, onOpenInventory }: { onSignOut: () => void; onOpenInventory: () => void }) {
  const me = useQuery(UserService.method.getMe, {});
  const pending = useLiveQuery(() => db.outbox.count(), [], null);
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
        onClick={onOpenInventory}
        className="min-h-11 w-full rounded-lg bg-sky-600 px-4 font-medium text-white active:bg-sky-700"
      >
        Equipment
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
