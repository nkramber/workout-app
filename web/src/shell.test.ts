import { readdirSync, readFileSync } from "node:fs";
import path from "node:path";

import { describe, expect, it } from "vitest";

// A headless browser has no status bar and no Home Screen, so it can not
// show the gap at the bottom edge of D-120. This test holds the rule in
// the source: the shell reads the dynamic viewport height, and no file
// sizes a screen by 100vh, which is taller than the visible part on a
// phone (decktome:D-625). The device check of PR-11 reads the phone.
const src = import.meta.dirname;

function sourceFiles(dir: string): string[] {
  return readdirSync(dir, { withFileTypes: true }).flatMap((e) => {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) return e.name === "gen" ? [] : sourceFiles(p);
    return /\.(tsx?|css)$/.test(e.name) && !e.name.endsWith(".test.ts") ? [p] : [];
  });
}

describe("the shell", () => {
  it("reads the dynamic viewport height and the safe areas", () => {
    const shell = readFileSync(path.join(src, "shell.tsx"), "utf8");
    expect(shell).toContain("h-dvh");
    for (const side of ["top", "right", "bottom", "left"]) expect(shell).toContain(`env(safe-area-inset-${side})`);
  });

  it("is the one size of the screen: no file uses 100vh or h-screen", () => {
    const offenders = sourceFiles(src).filter((f) => /\b100vh\b|\bh-screen\b|\bmin-h-screen\b/.test(readFileSync(f, "utf8")));
    expect(offenders).toEqual([]);
  });
});
