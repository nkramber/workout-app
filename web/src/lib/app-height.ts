// The height of the shell of src/shell.tsx (D-120).
//
// In the Home Screen app on iOS 27.0 with Chrome 154, the page draws
// below the translucent status bar, but the dynamic viewport height
// leaves out the band of that bar. So a shell of `100dvh` ended 62 pt
// above the bottom edge of an iPhone 16 Pro, the height of its top safe
// area (PR-28). A Home Screen app always fills the whole screen, so in
// the standalone display mode the shell takes the height of the screen.
// A browser tab keeps `100dvh`. The app has the phone layout alone
// (D-20), so a standalone window on a computer is out of scope.

export interface Viewport {
  standalone: boolean;
  innerWidth: number;
  innerHeight: number;
  screenWidth: number;
  screenHeight: number;
}

// appHeight gives the height of the shell in CSS pixels, or null for the
// default of `100dvh`. iOS gives the screen size in portrait in each
// orientation, so the side of the screen comes from the orientation of
// the window.
export function appHeight(v: Viewport): number | null {
  if (!v.standalone) return null;
  const landscape = v.innerWidth > v.innerHeight;
  const side = landscape ? Math.min(v.screenWidth, v.screenHeight) : Math.max(v.screenWidth, v.screenHeight);
  return Math.max(v.innerHeight, side);
}

// applyAppHeight sets the CSS variable `--app-height` of the document
// from the window, or removes it for a browser tab.
export function applyAppHeight(win: Window = window): void {
  const standalone =
    win.matchMedia("(display-mode: standalone)").matches ||
    (win.navigator as Navigator & { standalone?: boolean }).standalone === true;
  const h = appHeight({
    standalone,
    innerWidth: win.innerWidth,
    innerHeight: win.innerHeight,
    screenWidth: win.screen.width,
    screenHeight: win.screen.height,
  });
  const style = win.document.documentElement.style;
  if (h === null) style.removeProperty("--app-height");
  else style.setProperty("--app-height", `${h}px`);
}

// watchAppHeight applies the height now, and again after each resize and
// each turn of the screen.
export function watchAppHeight(win: Window = window): void {
  const apply = () => applyAppHeight(win);
  apply();
  win.addEventListener("resize", apply);
  win.addEventListener("orientationchange", apply);
}
