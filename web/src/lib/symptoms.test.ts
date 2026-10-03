import { describe, expect, it } from "vitest";

import { PAIN_WARNING, SYMPTOMS, symptomWarning } from "./symptoms";

describe("the symptoms of D-263", () => {
  it("holds the seven symptoms, each with a unique id", () => {
    expect(SYMPTOMS.map((s) => s.label)).toEqual([
      "Chest pain or pressure",
      "Unusual shortness of breath",
      "Dizziness or feeling faint",
      "A racing or irregular heartbeat",
      "Sharp pain in a joint or muscle",
      "Numbness or tingling",
      "Nausea",
    ]);
    expect(new Set(SYMPTOMS.map((s) => s.id)).size).toBe(SYMPTOMS.length);
  });

  it("names the symptom and tells the owner to stop, with no other advice (D-153)", () => {
    expect(symptomWarning(SYMPTOMS[2])).toBe("You reported dizziness or feeling faint. Stop this exercise.");
    for (const s of SYMPTOMS) {
      const text = symptomWarning(s);
      expect(text).toMatch(/^You reported .+\. Stop this exercise\.$/);
      expect(text).not.toMatch(/doctor|call|911|emergency|seek|hospital/i);
    }
  });

  it("keeps the pain warning of the policy (D-169)", () => {
    expect(PAIN_WARNING).toBe("You reported pain. Stop this exercise.");
  });
});
