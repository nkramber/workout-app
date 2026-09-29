import { defineConfig, devices } from "@playwright/test";

// The browser tests of the web client (work area 2.2). They run in WebKit,
// the engine of each browser on the iPhone, and in Chromium, each with
// phone emulation (D-20). `make web` starts the Auth and Firestore
// emulators around this run and builds the API into .bin/api (D-115).
// The config starts the API and a build of the web client that talks to
// the emulators and to that API.
const apiPort = 8480;
const webPort = 4273;
const webOrigin = `http://127.0.0.1:${webPort}`;

for (const name of ["FIREBASE_AUTH_EMULATOR_HOST", "FIRESTORE_EMULATOR_HOST"]) {
  if (!process.env[name]) throw new Error(`${name} is not set. Run the browser tests with \`make web\`.`);
}

export default defineConfig({
  testDir: "./e2e",
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: 0,
  reporter: [["list"]],
  timeout: 30_000,
  expect: { timeout: 10_000 },
  outputDir: "test-results",
  use: {
    baseURL: webOrigin,
    trace: "retain-on-failure",
  },
  projects: [
    { name: "webkit", use: { ...devices["iPhone 15"] } },
    { name: "chromium", use: { ...devices["Pixel 7"] } },
  ],
  webServer: [
    {
      // The emulator variables of `firebase emulators:exec` reach the API
      // through the environment of this process.
      command: "../.bin/api",
      url: `http://127.0.0.1:${apiPort}/version`,
      env: { PORT: String(apiPort), GOOGLE_CLOUD_PROJECT: "demo-gym-route", ALLOWED_ORIGIN: webOrigin },
      reuseExistingServer: false,
      timeout: 60_000,
    },
    {
      command: `npm run build:e2e && npx vite preview --outDir dist-e2e --host 127.0.0.1 --port ${webPort}`,
      url: webOrigin,
      reuseExistingServer: false,
      timeout: 120_000,
    },
  ],
});
