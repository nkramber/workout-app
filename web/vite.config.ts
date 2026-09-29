import { execSync } from "node:child_process";

import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";

import { webManifest } from "./src/lib/manifest";

// The build id is the short commit of the checkout. The device check of
// PR-11 reads it on the phone, so the check names the build it ran.
function buildId(): string {
  try {
    return execSync("git rev-parse --short HEAD", { encoding: "utf8" }).trim();
  } catch {
    return "unknown";
  }
}

export default defineConfig({
  define: {
    __BUILD_ID__: JSON.stringify(buildId()),
  },
  plugins: [
    react(),
    tailwindcss(),
    // The install item needs a manifest and a service worker. The worker
    // waits for the owner before it takes over (REC-3, D-133), so an open
    // form never loses its data to a reload. src/lib/pwa.ts registers it.
    VitePWA({
      registerType: "prompt",
      injectRegister: null,
      includeAssets: ["favicon.svg", "apple-touch-icon.png"],
      manifest: { ...webManifest, icons: [...webManifest.icons] },
      workbox: {
        globPatterns: ["**/*.{js,css,html,svg,png}"],
      },
      devOptions: { enabled: false },
    }),
  ],
  server: { port: 5273, strictPort: true },
  // WebKit refuses some ports, such as 4190 (tools/spikes/iphone_probe/README.md).
  preview: { port: 4273, strictPort: true },
});
