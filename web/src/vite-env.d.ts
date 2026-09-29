/// <reference types="vite/client" />

declare const __BUILD_ID__: string;

interface ImportMetaEnv {
  // The local Auth emulator, such as 127.0.0.1:9299. The browser tests
  // set it. Empty means the project of src/lib/firebase-config.ts.
  readonly VITE_AUTH_EMULATOR_HOST?: string;
  // The origin of the API, such as the Cloud Run URL of work area 2.3.
  readonly VITE_API_BASE_URL?: string;
}
