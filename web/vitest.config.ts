import { defineConfig } from "vitest/config";

// The unit tests of src/lib. They run in Node. fake-indexeddb gives
// Dexie an IndexedDB, so the tests of the store need no browser. The
// browser tests of e2e/ run under Playwright.
export default defineConfig({
  define: {
    __BUILD_ID__: JSON.stringify("test"),
  },
  test: {
    include: ["src/**/*.test.ts"],
    environment: "node",
  },
});
