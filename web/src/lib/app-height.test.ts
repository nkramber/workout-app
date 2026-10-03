import { describe, expect, it } from "vitest";

import { appHeight } from "./app-height";

// The iPhone 16 Pro of PR-28: a screen of 402 by 874 pt, and a window of
// 812 pt in the Home Screen app, 62 pt short of the screen.
const iphone = { innerWidth: 402, innerHeight: 812, screenWidth: 402, screenHeight: 874 };

describe("appHeight", () => {
  it("gives the whole height of the screen to a Home Screen app (D-120)", () => {
    expect(appHeight({ ...iphone, standalone: true })).toBe(874);
  });

  it("keeps 100dvh in a browser tab", () => {
    expect(appHeight({ ...iphone, standalone: false })).toBeNull();
  });

  it("uses the short side of the screen in landscape, because iOS gives the screen size in portrait", () => {
    expect(appHeight({ standalone: true, innerWidth: 874, innerHeight: 380, screenWidth: 402, screenHeight: 874 })).toBe(402);
  });

  it("never gives less than the window height", () => {
    expect(appHeight({ ...iphone, standalone: true, innerHeight: 900 })).toBe(900);
  });
});
