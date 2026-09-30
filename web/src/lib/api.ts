import type { Interceptor } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";

import { currentIdToken } from "./firebase";

// The bearer interceptor adds the Firebase ID token to each call, as in
// Decktome. A call with nobody signed in goes out with no header, and the
// API answers `unauthenticated`. The web app never reads Firestore itself.
export const bearerInterceptor: Interceptor = (next) => async (req) => {
  const token = await currentIdToken();
  if (token) req.header.set("Authorization", `Bearer ${token}`);
  return next(req);
};

// The API is on the default Cloud Run URL, another origin than the web
// app (D-82). VITE_API_BASE_URL names it at build time. cloudbuild/web.yaml
// sets it for the deploy. Empty means the origin of the page.
export const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_BASE_URL ?? "",
  interceptors: [bearerInterceptor],
});
