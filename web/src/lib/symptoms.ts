// The fixed list of warning symptoms of "Report a symptom" (D-251,
// D-263). The list comes from the EIM form (EV-8), with three items for
// the muscles and the joints. The phone keeps no report, and no report
// syncs.
export type Symptom = { id: string; label: string; text: string };

export const SYMPTOMS: readonly Symptom[] = [
  { id: "chest_pain", label: "Chest pain or pressure", text: "chest pain or pressure" },
  { id: "short_breath", label: "Unusual shortness of breath", text: "unusual shortness of breath" },
  { id: "dizziness", label: "Dizziness or feeling faint", text: "dizziness or feeling faint" },
  { id: "heartbeat", label: "A racing or irregular heartbeat", text: "a racing or irregular heartbeat" },
  { id: "sharp_pain", label: "Sharp pain in a joint or muscle", text: "sharp pain in a joint or muscle" },
  { id: "numbness", label: "Numbness or tingling", text: "numbness or tingling" },
  { id: "nausea", label: "Nausea", text: "nausea" },
];

// The warning of a pain report on a set or a cardio log: the text of
// PainWarning in go/internal/policy/warning.go (D-169).
export const PAIN_WARNING = "You reported pain. Stop this exercise.";

// symptomWarning gives the warning of a symptom (D-153, D-263). It names
// the symptom plainly and tells the owner to stop the exercise. It holds
// no referral, no first aid step, no emergency number, and no diagnosis
// (D-36). The owner can continue after a confirmation (D-40).
export function symptomWarning(s: Symptom): string {
  return `You reported ${s.text}. Stop this exercise.`;
}
