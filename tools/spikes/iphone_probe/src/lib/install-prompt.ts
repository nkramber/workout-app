// Chrome on Android and on a computer sends `beforeinstallprompt` once,
// early in the load. This module listens from the first import, so the
// install page can offer the prompt later. iOS sends no such event. On
// an iPhone, the owner adds the probe from the share menu of Chrome.
type InstallPromptEvent = Event & { prompt: () => Promise<void>; userChoice: Promise<{ outcome: string }> };

let deferred: InstallPromptEvent | null = null;
const listeners = new Set<() => void>();

window.addEventListener("beforeinstallprompt", (e) => {
  e.preventDefault();
  deferred = e as InstallPromptEvent;
  listeners.forEach((l) => l());
});

export function installPrompt(): InstallPromptEvent | null {
  return deferred;
}

export function onInstallPrompt(l: () => void): () => void {
  listeners.add(l);
  return () => listeners.delete(l);
}
