import "./index.css";

import { TransportProvider } from "@connectrpc/connect-query";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { App } from "./app";
import { transport } from "./lib/api";
import { startServiceWorker } from "./lib/pwa";

// One query client for the app. A failed call does not retry by itself,
// so an auth error shows at once, as in Decktome.
const queryClient = new QueryClient({
  defaultOptions: { queries: { retry: false, refetchOnWindowFocus: false, staleTime: 30_000 } },
});

const root = document.getElementById("root");
if (!root) throw new Error("root element not found");

createRoot(root).render(
  <StrictMode>
    <TransportProvider transport={transport}>
      <QueryClientProvider client={queryClient}>
        <App />
      </QueryClientProvider>
    </TransportProvider>
  </StrictMode>,
);
startServiceWorker();

// The browser tests build the app in the e2e mode alone. Vite removes
// this branch from each other build.
if (import.meta.env.MODE === "e2e") void import("./lib/e2e-hooks").then((m) => m.install());
