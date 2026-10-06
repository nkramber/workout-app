import { createContext, useContext, useState, type ReactNode } from "react";
import { createPortal } from "react-dom";

import { UpdateBanner } from "./update-banner";

// The shell fills the whole screen of the Home Screen app, with no gap at
// the bottom edge (D-120). Its height is `--app-height` of
// src/lib/app-height.ts in the Home Screen app, and the dynamic viewport
// height (`100dvh`) in a browser tab. It pads itself with the top and
// side safe areas. The main region is the one part that scrolls
// (decktome:D-625), and it reaches the bottom edge. Its bottom padding
// holds the bottom safe area, so the content scrolls below the home
// indicator and ends above it. The app has the phone layout alone (D-20).
//
// Since iOS 26, the Home Screen app blurs a band of the page below the
// status bar, and the page can not turn the blur off. So the header adds
// 16 px above its text when a status bar covers the page, the least space
// that keeps the title out of the band. With no top safe area, as in a
// browser tab, it adds nothing.
//
// The status slot below the header holds the line of the sync for a
// signed-in owner (D-276). The band below the main region holds a notice
// of a screen that must not move its content, such as the workout
// (D-321, D-322). The band is the last part of the column, so it reaches
// the bottom edge, and the main region gets shorter but does not move.
export function Shell({ children, status }: { children: ReactNode; status?: ReactNode }) {
  const [band, setBand] = useState<HTMLElement | null>(null);
  return (
    <div
      data-testid="shell"
      className="flex h-[var(--app-height,100dvh)] flex-col bg-[#0b1220] pt-[env(safe-area-inset-top)] pr-[env(safe-area-inset-right)] pl-[env(safe-area-inset-left)]"
    >
      <header className="flex items-center justify-between border-b border-slate-800 px-4 pt-[calc(0.75rem+min(env(safe-area-inset-top),16px))] pb-3">
        <h1 className="text-lg font-semibold text-slate-100">Workout App</h1>
      </header>
      {status}
      <UpdateBanner />
      <BandSlot.Provider value={band}>
        <main className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 pt-4 pb-[calc(1rem+env(safe-area-inset-bottom))]">{children}</main>
      </BandSlot.Provider>
      <div
        ref={setBand}
        data-testid="bottom-band"
        className="space-y-1 border-t border-slate-800 px-4 pt-2 pb-[calc(0.5rem+env(safe-area-inset-bottom))] empty:hidden"
      />
    </div>
  );
}

const BandSlot = createContext<HTMLElement | null>(null);

// BottomBand shows its children in the band at the bottom of the shell.
export function BottomBand({ children }: { children: ReactNode }) {
  const slot = useContext(BandSlot);
  return slot ? createPortal(children, slot) : null;
}
