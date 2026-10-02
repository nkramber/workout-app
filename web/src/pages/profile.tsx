import { useQuery } from "@connectrpc/connect-query";
import { useId, useState, type ReactNode } from "react";

import { InventoryService, type GetCatalogResponse } from "../gen/workoutapp/v1/inventory_service_pb";
import {
  ProfileService,
  type GetProfileOptionsResponse,
  type Option,
  type Profile,
} from "../gen/workoutapp/v1/profile_service_pb";
import { loadErrorText } from "../lib/errors";
import {
  applyTemplate,
  cardioExercises,
  checkDraft,
  draftOf,
  HEIGHT_FEET,
  HEIGHT_INCHES,
  INJURY_WARNING,
  MAX_TEXT_CHARS,
  toggle,
  TRAINING_DAYS,
  type Draft,
} from "../lib/profile";
import { useProfileApi } from "../lib/profile-api";
import { ErrorText, field, primary, secondary, Title, useAction } from "./inventory/ui";

// ProfilePage is the onboarding screen (work area 5.1). It holds each
// input of the profile on one screen, with one save (D-41, D-42, D-208 to
// D-211). Most inputs are large buttons, so the owner types little
// (D-71). The lists come from GetProfileOptions and the cardio exercises
// from GetCatalog, so the server holds the one copy of each list.
//
// With no saved profile, the app opens this screen before the home screen
// (D-223). It then has no Back button. After the first save, the home
// screen opens it to change the profile, and a save goes back to home.
// The save is a direct call to the API (D-196).
export function ProfilePage({ onBack, onSignOut }: { onBack?: () => void; onSignOut?: () => void }) {
  const options = useQuery(ProfileService.method.getProfileOptions, {}, { staleTime: Infinity });
  const catalog = useQuery(InventoryService.method.getCatalog, {}, { staleTime: Infinity });
  const profile = useQuery(ProfileService.method.getProfile, {});
  const title = onBack ? "Profile" : "Set up your profile";

  const error = options.error ?? catalog.error ?? profile.error;
  if (error) {
    return (
      <div className="space-y-4">
        <Title onBack={onBack}>{title}</Title>
        <ErrorText testId="load-error">{loadErrorText(error)}</ErrorText>
        <button
          type="button"
          className={secondary}
          onClick={() => {
            void options.refetch();
            void catalog.refetch();
            void profile.refetch();
          }}
        >
          Try again
        </button>
      </div>
    );
  }
  if (!options.data || !catalog.data || !profile.data) return <p className="text-slate-400">Loading…</p>;

  return (
    <ProfileForm
      title={title}
      options={options.data}
      catalog={catalog.data}
      stored={profile.data.profile}
      onBack={onBack}
      onSignOut={onSignOut}
    />
  );
}

type FormProps = {
  title: string;
  options: GetProfileOptionsResponse;
  catalog: GetCatalogResponse;
  stored: Profile | undefined;
  onBack?: () => void;
  onSignOut?: () => void;
};

function ProfileForm({ title, options, catalog, stored, onBack, onSignOut }: FormProps) {
  const api = useProfileApi();
  const action = useAction();
  const [draft, setDraft] = useState<Draft>(() => draftOf(stored));
  const set = (change: Partial<Draft>) => setDraft((d) => ({ ...d, ...change }));
  const ids = (list: readonly Option[]) => list.map((o) => o.id);
  const cardio = cardioExercises(catalog);

  const save = async () => {
    const r = checkDraft(draft);
    if (!r.ok) {
      action.setError(r.error);
      return;
    }
    const saved = await action.run(() => api.saveProfile(r.value));
    if (saved) onBack?.();
  };

  return (
    <div className="space-y-6">
      <Title onBack={onBack}>{title}</Title>
      {!onBack && (
        <p className="text-sm text-slate-400">Answer these questions one time. You can change them later from the home screen.</p>
      )}

      <Choices title="Experience" testId="experience">
        {options.experiences.map((o) => (
          <Toggle key={o.id} pressed={draft.experience === o.id} onClick={() => set({ experience: o.id })}>
            {o.name}
          </Toggle>
        ))}
      </Choices>

      <Choices title="Training days in each week" testId="training-days">
        {TRAINING_DAYS.map((n) => (
          <Toggle key={n} pressed={draft.trainingDays === n} onClick={() => set({ trainingDays: n })}>
            {`${n} days`}
          </Toggle>
        ))}
      </Choices>

      <Choices title="Goal" testId="goal-templates">
        {options.goalTemplates.map((t) => (
          <Toggle key={t.id} pressed={draft.goalTemplate === t.id} onClick={() => setDraft((d) => applyTemplate(d, t))}>
            {t.name}
          </Toggle>
        ))}
      </Choices>

      <Choices title="Muscle groups" testId="muscle-groups" hint="The goal selects these groups. Change them as you like.">
        {options.muscleGroups.map((o) => (
          <Toggle
            key={o.id}
            pressed={draft.muscleGroups.includes(o.id)}
            onClick={() => set({ muscleGroups: toggle(draft.muscleGroups, o.id, ids(options.muscleGroups)) })}
          >
            {o.name}
          </Toggle>
        ))}
      </Choices>

      <TextField
        label="Other goals for your plan (optional)"
        value={draft.freeText}
        onChange={(freeText) => set({ freeText })}
      />

      <Choices title="Injured areas" testId="injured-areas" hint="Select each area with an injury. Select none when you have no injury.">
        {options.injuryAreas.map((o) => (
          <Toggle
            key={o.id}
            pressed={draft.injuredAreas.includes(o.id)}
            onClick={() => set({ injuredAreas: toggle(draft.injuredAreas, o.id, ids(options.injuryAreas)) })}
          >
            {o.name}
          </Toggle>
        ))}
      </Choices>
      {draft.injuredAreas.length > 0 && (
        <p role="status" className="rounded-lg border border-amber-800 bg-amber-950 p-3 text-sm text-amber-200" data-testid="injury-warning">
          {INJURY_WARNING}
        </p>
      )}

      <TextField
        label="Other injuries or limits (optional)"
        hint="No AI model reads this text."
        value={draft.injuryText}
        onChange={(injuryText) => set({ injuryText })}
      />

      <section className="space-y-3" aria-label="Body">
        <h3 className="text-sm font-semibold text-slate-100">Body</h3>
        <NumberField label="Age (years)" value={draft.age} onChange={(age) => set({ age })} />
        <div className="grid grid-cols-2 gap-3">
          <SelectField
            label="Height (ft)"
            value={draft.heightFeet}
            values={HEIGHT_FEET}
            onChange={(heightFeet) => set({ heightFeet })}
          />
          <SelectField
            label="Height (in)"
            value={draft.heightInches}
            values={HEIGHT_INCHES}
            onChange={(heightInches) => set({ heightInches })}
          />
        </div>
        <NumberField label="Weight (lb)" value={draft.weight} onChange={(weight) => set({ weight })} />
      </section>

      <Choices title="Cardio that you like" testId="cardio" hint="Select none for no optional cardio.">
        {cardio.map((e) => (
          <Toggle
            key={e.id}
            pressed={draft.cardio.includes(e.id)}
            onClick={() => set({ cardio: toggle(draft.cardio, e.id, cardio.map((c) => c.id)) })}
          >
            {e.name}
          </Toggle>
        ))}
      </Choices>

      <div className="space-y-3">
        <ErrorText>{action.error}</ErrorText>
        <button type="button" className={`${primary} w-full`} disabled={action.busy} onClick={() => void save()}>
          Save
        </button>
        {onSignOut && (
          <button type="button" className={secondary} onClick={onSignOut}>
            Sign out
          </button>
        )}
      </div>
    </div>
  );
}

// Choices is one group of large buttons with its title.
function Choices({ title, hint, testId, children }: { title: string; hint?: string; testId: string; children: ReactNode }) {
  const id = useId();
  return (
    <section className="space-y-2" role="group" aria-labelledby={id} data-testid={testId}>
      <h3 id={id} className="text-sm font-semibold text-slate-100">
        {title}
      </h3>
      {hint && <p className="text-sm text-slate-400">{hint}</p>}
      <div className="flex flex-wrap gap-2">{children}</div>
    </section>
  );
}

// Toggle is a large button that shows its state with aria-pressed.
function Toggle({ pressed, onClick, children }: { pressed: boolean; onClick: () => void; children: ReactNode }) {
  return (
    <button
      type="button"
      aria-pressed={pressed}
      onClick={onClick}
      className={`min-h-11 rounded-lg border px-4 font-medium ${
        pressed ? "border-sky-500 bg-sky-950 text-sky-100" : "border-slate-700 text-slate-100 active:bg-slate-800"
      }`}
    >
      {children}
    </button>
  );
}

function TextField({ label, hint, value, onChange }: { label: string; hint?: string; value: string; onChange: (v: string) => void }) {
  const count = [...value.trim()].length;
  return (
    <label className="block">
      <span className="text-sm font-semibold text-slate-100">{label}</span>
      {hint && <span className="block text-sm text-slate-400">{hint}</span>}
      <textarea rows={3} value={value} onChange={(e) => onChange(e.target.value)} className={field} />
      <span className={`text-xs ${count > MAX_TEXT_CHARS ? "text-red-400" : "text-slate-500"}`}>
        {count} of {MAX_TEXT_CHARS} characters
      </span>
    </label>
  );
}

function NumberField({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  return (
    <label className="block">
      <span className="text-sm text-slate-400">{label}</span>
      <input
        type="text"
        inputMode="numeric"
        autoComplete="off"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={field}
      />
    </label>
  );
}

function SelectField({
  label,
  value,
  values,
  onChange,
}: {
  label: string;
  value: string;
  values: readonly number[];
  onChange: (v: string) => void;
}) {
  return (
    <label className="block">
      <span className="text-sm text-slate-400">{label}</span>
      <select value={value} onChange={(e) => onChange(e.target.value)} className={`${field} min-h-11`}>
        <option value="">-</option>
        {values.map((v) => (
          <option key={v} value={String(v)}>
            {v}
          </option>
        ))}
      </select>
    </label>
  );
}
