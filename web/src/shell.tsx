import type { ReactNode } from "react";

import { UpdateBanner } from "./update-banner";

// The shell fills the whole window of the Home Screen app, with no gap at
// the bottom edge (D-120). Its height is the dynamic viewport height
// (`h-dvh`), and it pads itself with the safe areas of the screen. The
// main region is the one part that scrolls (decktome:D-625). The app has
// the phone layout alone (D-20).
export function Shell({ children }: { children: ReactNode }) {
  return (
    <div
      data-testid="shell"
      className="flex h-dvh flex-col bg-[#0b1220] pt-[env(safe-area-inset-top)] pr-[env(safe-area-inset-right)] pb-[env(safe-area-inset-bottom)] pl-[env(safe-area-inset-left)]"
    >
      <header className="flex items-center justify-between border-b border-slate-800 px-4 py-3">
        <h1 className="text-lg font-semibold text-slate-100">Gym Route</h1>
      </header>
      <UpdateBanner />
      <main className="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 py-4">{children}</main>
    </div>
  );
}
