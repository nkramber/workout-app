import type { FirebaseOptions } from "firebase/app";

// The web configuration of the development project `gym-route-dev` (D-76,
// D-99, D-116). Firebase publishes these values to each browser that
// loads the app, so they are not a secret. The access rules of the
// project protect the data, not these values. The probe uses Firebase
// Authentication alone, so the configuration holds no storage bucket
// and no messaging sender.
export const projectConfig: FirebaseOptions | null = {
  apiKey: "AIzaSyAlu62ipBFG_lUiOApvQ6gAcOKxqZpH_Ac",
  authDomain: "gym-route-dev.firebaseapp.com",
  projectId: "gym-route-dev",
  appId: "1:338650836774:web:c0f99e8e111c007df29be7",
};
