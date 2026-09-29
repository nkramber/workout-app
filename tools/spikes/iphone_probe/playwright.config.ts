import { defineConfig, devices } from "@playwright/test";

// The browser tests of the probe (D-113). They run in WebKit, the engine
// of every browser on the iPhone, and in Chromium. `npm run test:e2e`
// starts the Auth emulator of firebase-tools around this run (D-115).
// The web server serves a build of the probe that talks to the emulator.
const phone = { width: 390, height: 844 };

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
    baseURL: "http://127.0.0.1:4173",
    trace: "retain-on-failure",
  },
  projects: [
    { name: "webkit", use: { ...devices["iPhone 15"] } },
    { name: "chromium", use: { ...devices["Desktop Chrome"], viewport: phone, hasTouch: true } },
  ],
  webServer: {
    command: "npm run build:e2e && npx vite preview --outDir dist-e2e --host 127.0.0.1",
    url: "http://127.0.0.1:4173",
    reuseExistingServer: false,
    timeout: 120_000,
  },
});
