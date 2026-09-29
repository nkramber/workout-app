import { execSync } from "node:child_process";

import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";

import { webManifest } from "./src/lib/manifest";

// The build id is the short commit of the checkout. The owner reads it
// on the phone, so the device checklist of PR-6 names the build it ran.
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
    // The install item needs a manifest and a service worker. The plugin
    // writes both from the build, so the precache list holds the hashed
    // asset names of that build.
    VitePWA({
      registerType: "autoUpdate",
      injectRegister: "script",
      includeAssets: ["favicon.svg", "apple-touch-icon.png"],
      manifest: { ...webManifest, icons: [...webManifest.icons] },
      workbox: {
        globPatterns: ["**/*.{js,css,html,svg,png}"],
      },
      devOptions: { enabled: false },
    }),
  ],
  server: { port: 5173, strictPort: true },
  preview: { port: 4173, strictPort: true },
});
