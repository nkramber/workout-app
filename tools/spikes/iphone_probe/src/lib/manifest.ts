// The web manifest of the probe. A browser installs a web app when the
// manifest has a name, a start URL, a display mode other than `browser`,
// and an icon of 192 pixels or more. The browser tests read these fields.
export const pageBackground = "#0b1220";

export const webManifest = {
  name: "Gym Route probe",
  short_name: "Probe",
  description: "The iPhone web platform probe of Gym Route, Phase 1.",
  start_url: "/",
  scope: "/",
  display: "standalone",
  orientation: "portrait",
  theme_color: pageBackground,
  background_color: pageBackground,
  icons: [
    { src: "/icon-192.png", sizes: "192x192", type: "image/png" },
    { src: "/icon-512.png", sizes: "512x512", type: "image/png" },
    { src: "/icon-maskable-512.png", sizes: "512x512", type: "image/png", purpose: "maskable" },
  ],
} as const;
