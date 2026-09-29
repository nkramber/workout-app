import type { FirebaseOptions } from "firebase/app";

// The web configuration of the development project of D-76 and D-99.
// Firebase publishes these values to each browser that loads the app, so
// they are not a secret. The access rules of the project protect the
// data, not these values. The project does not exist yet, so the value
// is null, and the sign-in page says so.
export const projectConfig: FirebaseOptions | null = null;
