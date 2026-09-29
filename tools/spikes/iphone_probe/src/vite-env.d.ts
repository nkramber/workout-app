/// <reference types="vite/client" />

declare const __BUILD_ID__: string;

interface ImportMetaEnv {
  readonly VITE_AUTH_EMULATOR_HOST?: string;
}
