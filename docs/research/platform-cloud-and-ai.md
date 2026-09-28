# Gym Route - platform, cloud, and AI research

Status: research input for the roadmap. This document holds evidence and synthesis. It holds no owner decision. The owner decisions live in `docs/decisions.md`, and the design lives in `docs/design.md`.

Date of this document: 2026-09-28. Access dates of the sources: 2026-09-27 and 2026-09-28.

## 1. Purpose, method, and limits

### 1.1 Purpose

This document gives the platform, cloud, and AI choices of Gym Route a cited base. Gym Route is an installable, phone-first web app on a default Firebase Hosting URL. It never goes to an app store, and no native app exists (D-17). A desktop browser shows the one phone layout (D-20). The owner accepted the limits of a web app on iOS (D-21). The app serves the owner alone (D-67).

Chrome on iPhone leads, and it uses the WebKit engine (D-29). The web client uses React 19.3, Vite 8, vite-plugin-pwa, TanStack Query with connect-query, and Tailwind 4 (D-84). Luna, the OpenAI model `gpt-6-luna` at medium effort, plans, revises, and reads equipment photos through a role layer with a fake provider (D-22, D-24). A deterministic policy checks every output (D-23). Google Cloud holds one development project in `us-central1` (D-18, D-76).

### 1.2 Method

- The researchers read official documentation, pricing pages, standards, source code, issue trackers, and papers.
- The fetch tool returns a model summary of each page, not the raw text. So the quotes here are near-verbatim. Check load-bearing wording on the live page.
- Browser support data comes mainly from the MDN browser-compat-data npm package 8.1.3, with a build date of 2026-09-24. The researchers queried it locally.
- Package versions come from the npm registry and the Go module proxy on the access dates.
- The first research targeted native iOS and Android apps for a public audience. D-17 and D-67 came later. Sections 4.5 and 9 keep those findings for a future reopening.
- Search result pages served for discovery only. They are not sources in the register.

| Class | Meaning |
|---|---|
| evidence | The fetched primary source states it. |
| recommendation | A vendor or standards body advises it, or the researchers propose it as a design choice. |
| assumption | An inference or estimate that no fetched source states. |
| unresolved | The researchers could not fetch or confirm it, or the sources conflict. |

### 1.3 Limits

- The current iOS is iOS 27 with Safari 27.0, released on 2026-09-14. The Safari 27 release notes page did not render, so its web app fixes stay unresolved (PC-18).
- The capability findings hold as of iOS 26.x. The research found no iOS 27 regression.
- Nobody tested on a physical iPhone. Playwright WebKit differs from branded Safari, and it can not run a Home Screen web app (PC-28).
- Prices are USD list prices on the access dates. Vendors change them without notice.
- Pages on openai.com returned HTTP 403. The launch post of Luna stays unresolved (PC-70). A dated copy of the usage policies gives their text (PC-101).

## 2. Source register

| Id | Source | Supports | Limits | Class |
|---|---|---|---|---|
| PC-1 | [WebKit Features in Safari 26.0](https://webkit.org/blog/17333/webkit-features-in-safari-26-0/), WebKit (Apple) | iOS 26 opens every Home Screen site as a web app by default. Safari has "zero requirements for installability". Web Inspector can inspect service workers. | Summary of a long post | evidence |
| PC-2 | [WebKit Features in Safari 18.4](https://webkit.org/blog/16574/webkit-features-in-safari-18-4/), WebKit | Screen Wake Lock works in Home Screen web apps from iOS 18.4. Declarative Web Push. ImageCapture. | Summary | evidence |
| PC-3 | [Safari 17.0](https://webkit.org/blog/14445/webkit-features-in-safari-17-0/), [Safari 26.5](https://webkit.org/blog/17938/webkit-features-for-safari-26-5/), [WWDC26: Safari 27 beta](https://webkit.org/blog/17967/news-from-wwdc26-webkit-in-safari-27-beta/), WebKit | Safari 17 decodes HEIC. Safari 26.5 fixed IndexedDB connections that broke until a reload. Safari 27 beta adds no PWA or device API. | Beta post, not final 27.0 notes | evidence |
| PC-4 | [Updates to Storage Policy](https://webkit.org/blog/14403/updates-to-storage-policy/), WebKit, 2023 | Quota up to 60% of disk per origin. Home Screen apps get equal quotas. LRU eviction skips persistent origins. persist() heuristics favor Home Screen apps. | 2023 post | evidence |
| PC-5 | [Full Third-Party Cookie Blocking and More](https://webkit.org/blog/10218/full-third-party-cookie-blocking-and-more/), WebKit, 2020 | Safari deletes script-written storage after 7 days of browser use with no interaction. Home Screen apps have their own day counter. | 2020 post, still cited by MDN | evidence |
| PC-6 | [Web Push for Web Apps on iOS and iPadOS](https://webkit.org/blog/13878/web-push-for-web-apps-on-ios-and-ipados/), WebKit, 2023 | Web Push for Home Screen apps only, from iOS 16.4. The permission request needs a user tap. | 2023 post | evidence |
| PC-7 | [WebKit bug 181849](https://bugs.webkit.org/show_bug.cgi?id=181849), WebKit Bugzilla | "Home Screen apps are created as isolated entities without shared state with the browser." | Last update 2022 | evidence |
| PC-8 | [WebKit bug 254545](https://bugs.webkit.org/show_bug.cgi?id=254545), WebKit Bugzilla | Wake Lock failed in Home Screen apps until the iOS 18.4 fix. | None | evidence |
| PC-9 | [WebKit bug 215884](https://bugs.webkit.org/show_bug.cgi?id=215884), WebKit Bugzilla, and [Apple Developer Forums 743049](https://developer.apple.com/forums/thread/743049), Apple | getUserMedia permission does not persist across standalone sessions (reports into 2026). An accept list with image/heic makes Safari 17+ convert photos to HEIC. | User reports, no Apple reply | evidence (weak) |
| PC-10 | [WebKit bug 207088](https://bugs.webkit.org/show_bug.cgi?id=207088), WebKit Bugzilla | Photo library EXIF handling varies by iOS version. iOS 16.4+ strips GPS by default. iOS 17 adds a user option to include location. | Behavior varies by version | evidence |
| PC-11 | [MDN browser-compat-data 8.1.3](https://www.npmjs.com/package/@mdn/browser-compat-data), MDN and Open Web Docs, build 2026-09-24 | Support versions for iOS Safari and Chrome Android: capture, Wake Lock, Vibration, Background Sync, Push, beforeinstallprompt, StorageManager, WebP encode, Shape Detection. | Compatibility data, not behavior | evidence |
| PC-12 | [Storage quotas and eviction criteria](https://developer.mozilla.org/en-US/docs/Web/API/Storage_API/Storage_quotas_and_eviction_criteria), MDN | Safari quotas, proactive 7-day eviction, silent persist() decisions. | None | evidence |
| PC-13 | [Screen Wake Lock API](https://developer.mozilla.org/en-US/docs/Web/API/Screen_Wake_Lock_API), MDN | The lock releases when the page hides, on low battery, or in power-save mode. Re-acquire on visibilitychange. | None | evidence |
| PC-14 | [capture](https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Attributes/capture), [createImageBitmap](https://developer.mozilla.org/en-US/docs/Web/API/Window/createImageBitmap), [toBlob](https://developer.mozilla.org/en-US/docs/Web/API/HTMLCanvasElement/toBlob), MDN | capture="environment" asks for the rear camera. createImageBitmap applies EXIF orientation and resizes. toBlob writes JPEG with a quality value. | Silent on EXIF removal | evidence |
| PC-15 | [Making PWAs installable](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps/Guides/Making_PWAs_installable), MDN | Manifest fields. iOS 16.4+ installs from the Share menu without a manifest. apple-touch-icon advice. | None | evidence |
| PC-16 | [Customize install](https://web.dev/articles/customize-install), [Install criteria](https://web.dev/articles/install-criteria), [WebAPKs](https://web.dev/articles/webapks), web.dev, and [Revisiting installability criteria](https://developer.chrome.com/blog/update-install-criteria), Chrome for Developers | beforeinstallprompt flow. Chrome install criteria. No fetch handler needed since Chrome 108 on mobile. A manifest change re-mints the WebAPK. | Chrome-centric | evidence |
| PC-17 | [Notification Triggers API](https://developer.chrome.com/docs/web-platform/notification-triggers), Chrome for Developers | Development ended. No browser schedules a local notification. | None | evidence |
| PC-18 | [Safari 27 Release Notes](https://developer.apple.com/documentation/safari-release-notes/safari-27-release-notes), Apple | Nothing. The page body did not render. | Not read | unresolved |
| PC-19 | [HTML Standard, canvas](https://html.spec.whatwg.org/multipage/canvas.html), WHATWG | Nothing on metadata in a canvas export. | Inconclusive excerpt | unresolved |
| PC-20 | [Best practices for signInWithRedirect](https://firebase.google.com/docs/auth/web/redirect-best-practices), Firebase, updated 2026-09-24 | Redirect sign-in breaks on Safari 16.1+ because of third-party storage partitioning. Five workarounds. | Silent on email and password | evidence |
| PC-21 | @firebase/auth 1.13.6 source, getAuth function, [npm](https://www.npmjs.com/package/@firebase/auth), Google | getAuth defaults to IndexedDB persistence first, and it bundles the popup and redirect resolver. | Source code, not docs | evidence |
| PC-22 | [App Check with reCAPTCHA Enterprise in web apps](https://firebase.google.com/docs/app-check/web/recaptcha-enterprise-provider), Firebase, and [Compare Fraud Defense tiers](https://docs.cloud.google.com/recaptcha/docs/compare-tiers), Google Cloud | Score-based key, 1 h token TTL. First 10,000 assessments a month free, then $8 flat to 100,000. | reCAPTCHA renamed Fraud Defense | evidence |
| PC-23 | [Configure Hosting behavior](https://firebase.google.com/docs/hosting/full-config), Firebase | Headers per glob in firebase.json. SPA rewrites to index.html. | Default cache statement vague | evidence |
| PC-24 | [Automatic reload](https://vite-pwa-org.netlify.app/guide/auto-update.html), [Prompt for new content](https://vite-pwa-org.netlify.app/guide/prompt-for-update.html), [Periodic updates](https://vite-pwa-org.netlify.app/guide/periodic-sw-updates.html), [React](https://vite-pwa-org.netlify.app/frameworks/react.html), [Frameworks](https://vite-pwa-org.netlify.app/frameworks/), Vite PWA | prompt is the default registerType. autoUpdate can lose form data. useRegisterSW for React. Framework modules for React, Vue, Svelte, Solid, Preact. | None | evidence, recommendation |
| PC-25 | [vite-plugin-pwa releases](https://github.com/vite-pwa/vite-plugin-pwa/releases), GitHub | 1.3.0 supports Vite 7 and Vite 8. | None | evidence |
| PC-26 | [Dexie.js releases](https://github.com/dexie/Dexie.js/releases), GitHub, and [StorageManager](https://dexie.org/docs/StorageManager), Dexie | Dexie 4.4.x maintained, no 5.x. persist() carries no guarantee. | Release years in the summary conflict with npm | evidence |
| PC-27 | [Persistent Storage Options](https://sqlite.org/wasm/doc/trunk/persistence.md), SQLite | The opfs VFS needs COOP and COEP headers. opfs-sahpool needs neither and supports Safari 16.4+. | None | evidence |
| PC-28 | [Emulation](https://playwright.dev/docs/emulation), [Browsers](https://playwright.dev/docs/browsers), [Accessibility testing](https://playwright.dev/docs/accessibility-testing), Playwright | Playwright WebKit "doesn't work with branded Safari". Emulation covers viewport, touch, offline. axe finds some problems, and manual tests find the rest. | Silent on iOS | evidence |
| PC-29 | npm registry and Go module proxy lookups, [registry.npmjs.org](https://registry.npmjs.org/) | Versions and dates: react 19.3.0, vite 8.3.1, vite-plugin-pwa 1.3.0, workbox 7.4.1, dexie 4.4.6, firebase 12.19.0, connect-web 2.2.0, connect-query 2.3.1, protobuf 2.15.0, vitest 5.0.2, playwright 1.63.0, tasks-vision 1.0.1, openai 7.23.0 | Registry metadata only | evidence |
| PC-30 | [connect-query-es](https://github.com/connectrpc/connect-query-es) and [issue 324](https://github.com/connectrpc/connect-query-es/issues/324), Connect RPC | "Connect-Query does require React". A core package exists for other frameworks, with no official wrapper. | FAQ can lag the core package | evidence |
| PC-31 | [connect-es](https://github.com/connectrpc/connect-es) and [Getting started for web](https://connectrpc.com/docs/web/getting-started/), Connect RPC | Framework-agnostic and stable. Fetch-based clients support unary and server streaming only. | None | evidence |
| PC-32 | [React 19.3](https://react.dev/blog/2026/09/09/react-19-3), [React Compiler v1.0](https://react.dev/blog/2025/10/07/react-compiler-1), [The React Foundation](https://react.dev/blog/2026/02/24/the-react-foundation), React team | 19.3 has no noted breaking change. Compiler 1.0 stable. Linux Foundation governance since 2026-02-24. | Performance claims are the vendor's own | evidence |
| PC-33 | [Announcing Vite 8](https://vite.dev/blog/announcing-vite8), Vite team | Rolldown bundler. Most plugins work unchanged. plugin-react v6 needs Vite 8. | Vendor speed claims | evidence |
| PC-34 | [js-framework-benchmark, Chrome 152](https://krausest.github.io/js-framework-benchmark/2026/chrome152.html), Stefan Krause | Compressed size and first paint for each framework (section 4.3). | Desktop Chrome on macOS, no phone | evidence |
| PC-35 | [The Performance Inequality Gap, 2026](https://infrequently.org/2025/11/performance-inequality-gap-2026/), Alex Russell | Galaxy A24 4G as the 75th-percentile device. 0.62 MiB JavaScript budget for a 3 s load. | Models a web page load, not a PWA relaunch | evidence, recommendation |
| PC-36 | [Web-Bench, arXiv 2505.07473](https://arxiv.org/html/2505.07473v1), ByteDance | LLMs scored higher on React than on Vue, Angular, and Svelte. | 2025 models | evidence (dated) |
| PC-37 | [SvelteKit 3 release candidate](https://svelte.dev/blog/sveltekit-3-release-candidate) and [Svelte AI docs](https://svelte.dev/docs/ai/overview), Svelte team | SvelteKit 3 is at RC with breaking changes. LLMs often write outdated Svelte 4 syntax. | Vendor framing | evidence |
| PC-38 | [Solid v2.0.0 RC discussion](https://github.com/solidjs/solid/discussions/2995), SolidJS | Solid 2.0 RC removes several APIs. No stable date. | Kobalte claim conflicts with npm | evidence |
| PC-39 | [Vue v3.6.0-alpha.1 release](https://github.com/vuejs/core/releases/tag/v3.6.0-alpha.1), Vue team | Vapor mode is opt-in with feature gaps. | Alpha-era notes | evidence |
| PC-40 | [Announcing Vitest 5.0](https://vitest.dev/blog/vitest-5), Vitest team | Vitest 5 needs Vite 6.4+ and Node 22.12+. | None | evidence |
| PC-41 | [Switching to Preact](https://preactjs.com/guide/v10/switching-to-preact/), Preact team | No stated React 19 compatibility target. | Leaves compat scope open | unresolved |
| PC-42 | [Expo SDK reference](https://docs.expo.dev/versions/latest/), [SDK 57 changelog](https://expo.dev/changelog/sdk-57), [New Architecture](https://docs.expo.dev/guides/new-architecture/), Expo | SDK 57 ships React Native 0.86 and Hermes V1. New Architecture mandatory since SDK 55. | Summaries | evidence |
| PC-43 | [examples-es react-native](https://github.com/connectrpc/examples-es/tree/main/react-native), [connect-es issue 199](https://github.com/connectrpc/connect-es/issues/199), [issue 1372](https://github.com/connectrpc/connect-es/issues/1372), Connect RPC | Connect-Web runs on Expo with no polyfills. No pure gRPC transport on React Native. | Example pinned to Expo 53 | evidence |
| PC-44 | [expo issue 50213](https://github.com/expo/expo/issues/50213), [issue 50212](https://github.com/expo/expo/issues/50212), [issue 47762](https://github.com/expo/expo/issues/47762), Expo | SDK 57 expo/fetch hangs on iOS after a drop, truncates on Android, loses the abort type, and reorders chunks. | Community reports | evidence |
| PC-45 | [What's new in Flutter 3.47](https://flutter.dev/blog/whats-new-in-flutter-3-47), [Impeller](https://docs.flutter.dev/perf/impeller), Flutter team | Flutter 3.47.5 stable. Minimum iOS 15. Impeller only on iOS. Android opt-out deprecated. | Summaries | evidence |
| PC-46 | [connect-dart](https://github.com/connectrpc/connect-dart) and [Using clients](https://connectrpc.com/docs/dart/using-clients), Connect RPC | Connect-Dart 2.0.0 stable (2026-09-01). The http2 client supports every RPC type. | Small adoption | evidence |
| PC-47 | [connect-swift](https://github.com/connectrpc/connect-swift), Connect RPC | 1.2.3 stable. SwiftPM only. Supports watchOS. | None | evidence |
| PC-48 | [connect-kotlin](https://github.com/connectrpc/connect-kotlin) and [issue 140](https://github.com/connectrpc/connect-kotlin/issues/140), Connect RPC | 0.9.0 "in beta". JVM only, no Kotlin Multiplatform support. | None | evidence |
| PC-49 | [Xiong et al., arXiv 2306.13063](https://arxiv.org/abs/2306.13063), ICLR 2024 | LLMs are overconfident when they state confidence. Consistency across samples helps. | Text LLMs, 2023 models | evidence |
| PC-50 | [Groot and Valdenegro-Toro, arXiv 2405.02917](https://arxiv.org/abs/2405.02917) | LLMs and VLMs "have a high calibration error and are overconfident most of the time". | GPT-4V era | evidence |
| PC-51 | [Ferdous, arXiv 2607.22034](https://arxiv.org/abs/2607.22034), 2026 | Small VLMs state an almost constant confidence. Under underexposure, accuracy fell from 0.99 to 0.22 while stated confidence held. | Two small open models | evidence (limited) |
| PC-52 | [Tian et al., arXiv 2305.14975](https://arxiv.org/abs/2305.14975), and [Kadavath et al., arXiv 2207.05221](https://arxiv.org/abs/2207.05221) | Counterpoint: RLHF text models can state better-calibrated confidence than token probabilities. | Text only, abstracts only | evidence |
| PC-53 | [Wang et al., arXiv 2203.11171](https://arxiv.org/abs/2203.11171), ICLR 2023 | Self-consistency: sample several answers and pick the most consistent. | Reasoning tasks, not vision | evidence |
| PC-54 | [Angelopoulos and Bates, arXiv 2107.07511](https://arxiv.org/abs/2107.07511) | Conformal prediction sets with a coverage guarantee from a calibration set. | Needs exchangeable data | evidence |
| PC-55 | [Geifman and El-Yaniv, arXiv 1705.08500](https://arxiv.org/abs/1705.08500) | Selective classification: a reject option trades coverage for risk. | 2017 CNNs | evidence |
| PC-56 | [Cloud Vision API pricing](https://cloud.google.com/vision/pricing), Google Cloud | First 1,000 units a month free. TEXT_DETECTION $1.50 per 1,000 units up to 5 million. | None | evidence |
| PC-57 | [Agent Platform Pricing](https://cloud.google.com/vertex-ai/generative-ai/pricing), Google Cloud | Gemini 3.5 Flash-Lite $0.30 input and $2.50 output per 1M tokens (global). | Parsed from raw HTML | evidence |
| PC-58 | [Model versions and lifecycle](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/learn/model-versions), Google Cloud | Gemini 2.5 Flash and Flash-Lite retire on 2026-10-20. | Dates can move | evidence |
| PC-59 | [Image understanding](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/multimodal/image-understanding), [Inference reference](https://docs.cloud.google.com/vertex-ai/generative-ai/docs/model-reference/inference), Google Cloud, and [forum thread](https://discuss.ai.google.dev/t/missing-logprobs-support-in-the-newest-gemini-models-3-1-pro-3-6-flash-on-vertex-ai-and-ai-studio/176557), Google AI Developers Forum | Gemini 3 bills 1,120 tokens per image by default. Gemini 3.x returns no logprobs. | Forum reply is not official docs | evidence |
| PC-60 | [Face detection guide for Web](https://developers.google.com/edge/mediapipe/solutions/vision/face_detector/web_js), Google AI Edge | Browser face detector with boxes and keypoints. | No browser support matrix | evidence |
| PC-61 | [GPT-6 Luna Model](https://developers.openai.com/api/docs/models/gpt-6-luna) and [Pricing](https://developers.openai.com/api/docs/pricing), OpenAI | Price, context, effort values, features, rate limits. No logprobs listed. No dated snapshot listed. | Alias only | evidence |
| PC-62 | [Images and vision](https://developers.openai.com/api/docs/guides/images-vision), OpenAI | 32 px patch tokens. Model multipliers. Formats PNG, JPEG, WEBP, GIF. No HEIC. | Luna multiplier absent | evidence |
| PC-63 | [Structured Outputs](https://developers.openai.com/api/docs/guides/structured-outputs), OpenAI | Strict json_schema, all fields required, a refusal field. | Silent on images plus schema | evidence |
| PC-64 | [Reasoning models](https://developers.openai.com/api/docs/guides/reasoning), OpenAI | reasoning.effort. Reasoning tokens bill as output. Encrypted reasoning with store false. | Silent on logprobs | evidence |
| PC-65 | [Data controls in the OpenAI platform](https://developers.openai.com/api/docs/guides/your-data), OpenAI | No training by default. Abuse logs up to 30 days. ZDR needs approval. store true keeps state 30 days. US residency with a 10% uplift. | Summary | evidence |
| PC-66 | [Create a model response](https://developers.openai.com/api/reference/resources/responses/methods/create) and [Using GPT-6](https://developers.openai.com/api/docs/guides/latest-model), OpenAI | The include option can ask for logprobs. With effort other than none, remove top_logprobs and the logprobs include. | None | evidence |
| PC-67 | [Flex processing](https://developers.openai.com/api/docs/guides/flex-processing) and [Batch API](https://developers.openai.com/api/docs/guides/batch), OpenAI | Flex bills at Batch rates and can return 429. Batch is 50% off within 24 h. | None | evidence |
| PC-68 | [Safety best practices](https://developers.openai.com/api/docs/guides/safety-best-practices), OpenAI | Free Moderation API, hashed safety_identifier, adversarial tests. | None | recommendation |
| PC-69 | [Usage policies](https://openai.com/policies/usage-policies/), OpenAI | HTTP 403 again on 2026-09-28. PC-101 gives a dated copy of the page. | Not fetched | unresolved, see PC-101 |
| PC-70 | [Introducing GPT-6 Sol and Luna](https://openai.com/index/introducing-gpt-6-sol-and-luna/), OpenAI | Nothing verified. HTTP 403. | Not fetched | unresolved |
| PC-71 | [Overview of environments](https://firebase.google.com/docs/projects/dev-workflows/overview-environments), Firebase | Firebase advises a separate project for each environment, and no real user data in development. | General guidance | recommendation |
| PC-72 | [Cloud Run pricing](https://cloud.google.com/run/pricing), Google Cloud | Request-based and instance-based rates. Free tier per billing account. us-central1 is Tier 1. | Prices change | evidence |
| PC-73 | [Min instances](https://cloud.google.com/run/docs/configuring/min-instances), [General tips](https://cloud.google.com/run/docs/tips/general), [Task timeout](https://cloud.google.com/run/docs/configuring/task-timeout), [Jobs on a schedule](https://cloud.google.com/run/docs/execute/jobs-on-schedule), Google Cloud | Min instances cost money. Startup CPU boost. Job timeout up to 168 h. Scheduler runs a job. | None | evidence |
| PC-74 | [Firestore locations](https://firebase.google.com/docs/firestore/locations), Firebase, and [Firestore pricing](https://cloud.google.com/firestore/pricing), Google Cloud | Location is permanent. Regional SLA 99.99%. us-central1 operations cost half of nam5. Free quota. | Prices parsed from page data | evidence |
| PC-75 | [Firestore editions overview](https://cloud.google.com/firestore/docs/editions-overview) and [Choose between Firestore APIs](https://cloud.google.com/datastore/docs/firestore-or-datastore), Google Cloud | Standard and Enterprise editions. Datastore mode disables client SDK features. | Enterprise unit prices not read | evidence |
| PC-76 | [PITR](https://firebase.google.com/docs/firestore/pitr), [Backups](https://firebase.google.com/docs/firestore/backups), [TTL](https://firebase.google.com/docs/firestore/ttl), Firebase | PITR keeps 7 days. One daily and one weekly backup schedule, up to 14 weeks. TTL deletes within about 24 h. | In-place restore not read | evidence |
| PC-77 | [Get started with Security Rules](https://firebase.google.com/docs/firestore/security/get-started) and [Avoid insecure rules](https://firebase.google.com/docs/rules/insecure-rules), Firebase | Server libraries bypass rules. Deny-all rules for a server-only backend. | None | evidence, recommendation |
| PC-78 | [Signed URLs](https://cloud.google.com/storage/docs/access-control/signed-urls), [Go storage package](https://pkg.go.dev/cloud.google.com/go/storage), [Uniform bucket-level access](https://cloud.google.com/storage/docs/uniform-bucket-level-access), [Public access prevention](https://cloud.google.com/storage/docs/public-access-prevention), Google Cloud | V4 signed URLs up to 7 days. Keyless signing through signBlob. UBLA advised. PAP does not block signed URLs. | None | evidence |
| PC-79 | [Lifecycle](https://cloud.google.com/storage/docs/lifecycle), [Soft delete](https://cloud.google.com/storage/docs/soft-delete), [Bucket Lock](https://cloud.google.com/storage/docs/bucket-lock), Google Cloud | Lifecycle changes take up to 24 h. Soft delete defaults to 7 days, 0 turns it off, and soft-deleted bytes bill. A locked retention policy is permanent. | None | evidence |
| PC-80 | [Storage pricing](https://cloud.google.com/storage/pricing), Google Cloud, and [Cloud Storage for Firebase billing changes](https://firebase.google.com/docs/storage/faqs-storage-changes-announced-sept-2024), Firebase | Always Free only in us-west1, us-central1, us-east1. Cloud Storage for Firebase needs Blaze from 2026-02-03. | Per-GB price not read | evidence |
| PC-81 | [Secret Manager](https://cloud.google.com/secret-manager/pricing), [Artifact Registry](https://cloud.google.com/artifact-registry/pricing), [Cloud Build](https://cloud.google.com/build/pricing), [Private pool schema](https://cloud.google.com/build/docs/private-pools/private-pool-config-file-schema), [Cloud Scheduler](https://cloud.google.com/scheduler/pricing), Google Cloud | Free tiers and rates in section 8.2. Cloud Build machine types are e2, n2d, and c3 only. | Cloud Build free tier "promotional" | evidence |
| PC-82 | [Budgets and alerts](https://cloud.google.com/billing/docs/how-to/budgets) and [Disable billing with notifications](https://cloud.google.com/billing/docs/how-to/disable-billing-with-notifications), Google Cloud | Budgets do not cap spend. A billing shutdown can delete resources for good. | None | evidence |
| PC-83 | [Manage spend cap budgets](https://cloud.google.com/billing/docs/how-to/budgets-spend-caps), Google Cloud | Spend caps (Preview) pause Vertex AI, Gemini API, Cloud Run, and Cloud Run functions for one project. | Pre-GA terms | evidence |
| PC-84 | [Google Cloud Observability pricing](https://cloud.google.com/stackdriver/pricing), Google Cloud | Logging: first 50 GiB a project a month free. Error Reporting has no charge. | None | evidence |
| PC-85 | [Extensions Deprecation FAQ](https://firebase.google.com/docs/extensions/faq-and-troubleshooting) and [Delete User Data](https://extensions.dev/extensions/firebase/delete-user-data), Firebase | Firebase Extensions shut down on 2027-03-31. | None | evidence |
| PC-86 | [Verify App Check tokens from a custom backend](https://firebase.google.com/docs/app-check/custom-resource-backend), Firebase, and [appcheck package](https://pkg.go.dev/firebase.google.com/go/v4/appcheck), Go | Go Admin SDK VerifyToken, v4.22.0. Replay protection only in Node.js. | None | evidence |
| PC-87 | [Emulator Suite](https://firebase.google.com/docs/emulator-suite), [Auth emulator](https://firebase.google.com/docs/emulator-suite/connect_auth), [Firestore emulator](https://firebase.google.com/docs/emulator-suite/connect_firestore), [Storage emulator](https://firebase.google.com/docs/emulator-suite/connect_storage), Firebase | Environment variables. With the Auth emulator variable set, Admin SDKs accept unsigned ID tokens. No App Check emulator. Java 21 soon. | Signed URL support not stated | evidence |
| PC-88 | [Hosting quotas and pricing](https://firebase.google.com/docs/hosting/usage-quotas-pricing) and [Firebase Pricing](https://firebase.google.com/pricing), Firebase | Free web.app subdomain with SSL. 10 GB storage and 360 MB a day at no cost. Auth no-cost to 50,000 MAU. | None | evidence |
| PC-89 | [App Review Guidelines](https://developer.apple.com/app-store/review/guidelines/) and [Offering account deletion](https://developer.apple.com/support/offering-account-deletion-in-your-app/), Apple | 5.1.1(v) in-app account deletion. 5.1.2(i) disclosure of third-party AI. 1.4.1 medical scrutiny. | No update date shown | evidence |
| PC-90 | [Update on regulated medical device apps](https://developer.apple.com/news/?id=nyqbfz1y), 2026-03-26, and [Declare regulated medical device status](https://developer.apple.com/help/app-store-connect/manage-app-information/declare-regulated-medical-device-status/), Apple | New Health and Fitness apps in the EEA, UK, and US declare a medical device status. | None | evidence |
| PC-91 | [App Privacy Details](https://developer.apple.com/app-store/app-privacy-details/), Apple | Privacy label data types: Health, Fitness, Photos, User ID. | None | evidence |
| PC-92 | [TestFlight overview](https://developer.apple.com/help/app-store-connect/test-a-beta-version/testflight-overview) and [TestFlight](https://developer.apple.com/testflight/), Apple | A build runs for up to 90 days. 100 internal and 10,000 external testers. | None | evidence |
| PC-93 | [Health Content and Services](https://support.google.com/googleplay/android-developer/answer/16679511) and [Health apps declaration](https://support.google.com/googleplay/android-developer/answer/14738291), Google Play | Every app completes the health declaration, testing tracks too. Disclaimer for other health apps. | Scope of disclaimer ambiguous | evidence |
| PC-94 | [Account deletion requirements](https://support.google.com/googleplay/android-developer/answer/13327111) and [Data safety](https://support.google.com/googleplay/android-developer/answer/10787469), Google Play | In-app deletion plus a web link resource. Data safety form for closed, open, and production tracks. | None | evidence |
| PC-95 | [Testing requirements for new personal developer accounts](https://support.google.com/googleplay/android-developer/answer/14151465), Google Play | 12 testers opted in for 14 days before production access. | Personal accounts after 2023-11-13 | evidence |
| PC-96 | [General Wellness: Policy for Low Risk Devices](https://www.fda.gov/regulatory-information/search-fda-guidance-documents/general-wellness-policy-low-risk-devices), FDA, 2026-01-06 | Strength and muscle size claims are wellness claims. Disease and treatment claims fall outside. | Crude PDF text extraction | recommendation |
| PC-97 | [Understanding SC 2.5.8 Target Size (Minimum)](https://www.w3.org/WAI/WCAG22/Understanding/target-size-minimum.html), W3C | Level AA target of at least 24 by 24 CSS pixels, with exceptions. | None | evidence |
| PC-98 | [WCAG2Mobile](https://www.w3.org/TR/wcag2mobile-22/) and [Mobile Accessibility at W3C](https://www.w3.org/WAI/standards-guidelines/mobile/), W3C | Group Draft Note of 2025-05-06. Informative only. Covers mobile web apps. | Draft | recommendation |
| PC-99 | [HIG Accessibility](https://developer.apple.com/design/human-interface-guidelines/accessibility), Apple | Controls 44 by 44 pt by default, 28 by 28 pt minimum. Contrast 4.5:1. | Native guidance | recommendation |
| PC-100 | [Make apps more accessible](https://developer.android.com/guide/topics/ui/accessibility/apps), Android Developers | Targets of at least 48 by 48 dp. | Native guidance | recommendation |
| PC-101 | [Usage policies, print copy](https://docs.databricks.com/aws/ja/assets/files/usage-policies-openai-923e062e72487537e2a8df04bcec7f6d.pdf), a Databricks copy of the OpenAI page | The full text of the policy page, printed 2025-11-07, effective 2025-10-29. Read on 2026-09-28. | A copy, not the live page. Changes after 2025-11-07 stay unverified. | evidence, dated copy |

## 3. Installable web app on iPhone and Android

Chrome on iOS uses the WebKit engine (D-29). So each iOS limit below applies to Chrome on iPhone as it applies to Safari. Android Chrome matters less, because the owner uses an iPhone (D-29, D-67). The table keeps the Android data for comparison.

| Capability | iOS (Safari and Chrome, WebKit) | Android Chrome | Effect on Gym Route | Sources |
|---|---|---|---|---|
| Install flow | No beforeinstallprompt. The user taps Share, then Add to Home Screen. Safari 26 needs no manifest. | beforeinstallprompt, then an install button. Needs name, 192 and 512 px icons, start_url, a standalone display, and HTTPS. | A short instruction screen on iOS. A stable manifest with apple-touch-icon. | PC-11, PC-1, PC-15, PC-16 |
| iOS 26 Home Screen default | Every site that the user adds to the Home Screen opens as a web app. The user can turn off "Open as Web App". | Not applicable | Tell the owner to keep the default. | PC-1 |
| Install from Chrome on iPhone | The sources describe Safari only. The Chrome path to the Home Screen on iOS 26 and 27 stays untested. | Not applicable | Test on the owner's iPhone in Phase 1. | unresolved, PC-1 |
| Storage isolation | Cookies, localStorage, and IndexedDB of a Home Screen app are separate from the browser. This is by design. | Not researched | The owner signs in again inside the installed app. Data logged in a browser tab stays in the tab. | PC-7 |
| Quota, eviction, persist() | Up to 60% of disk per origin. Home Screen apps get equal quotas. LRU eviction skips persistent origins. WebKit grants persist() by heuristics that favor Home Screen apps, with no prompt. | Chrome grants persist() silently by engagement. | Call navigator.storage.persist() in the installed app and log the result. The server copy stays the durable record (D-77). | PC-4, PC-12, PC-26 |
| 7-day ITP cap | Safari deletes script-written storage after 7 days of browser use with no interaction. Home Screen apps count their own days of use, and WebKit does not expect their first-party data to go. | Not applicable | Log workouts in the installed app only. | PC-5, PC-12 |
| IndexedDB connection bug | Safari 26.5 fixed IndexedDB connections that broke until a reload. | Not applicable | Catch the error and reopen the database on older iOS. | PC-3 |
| Camera input | input capture since iOS Safari 10. getUserMedia works in standalone mode, but its permission does not persist across sessions. An accept list with image/heic makes Safari convert photos to HEIC. | input capture since Chrome Android 25. | Use a file input with accept="image/*" and capture="environment". Leave image/heic out. | PC-11, PC-14, PC-9 |
| Screen Wake Lock | Safari tabs from 16.4. Home Screen apps from 18.4 only. The lock releases when the page hides. | Since Chrome 84. | Request the lock at workout start, and request it again on visibilitychange. On older iOS, show "keep the screen on" advice. | PC-2, PC-8, PC-11, PC-13 |
| Vibration | No navigator.vibrate in any iOS version, checked through 27.2. | Since Chrome 32. | No effect. D-58 chose visual cues only. | PC-11 |
| Background Sync | No SyncManager and no PeriodicSyncManager through 27.2. | Background Sync since 49, Periodic Sync since 80. | The app syncs only while it is open, visible, and online. | PC-11 |
| Web Push | Home Screen apps only, from 16.4. The permission needs a user tap. Declarative Web Push from 18.4. | Push since Chrome 42. | Not used. D-61 chose no notifications. | PC-6, PC-2, PC-11 |
| Scheduled local notification | No browser supports it. Chrome ended Notification Triggers. | Same | The rest timer shows on the screen only (D-61). | PC-17 |
| Shape Detection | FaceDetector absent. BarcodeDetector behind a flag on iOS 17. | Not captured | No built-in detection API for recognition on the iPhone. See section 6. | PC-11 |
| Image formats | Safari 17 decodes HEIC. Canvas can not encode WebP on iOS. JPEG encode works from iOS 11. | Not captured | Decode, then encode JPEG. OpenAI accepts JPEG and rejects HEIC. | PC-3, PC-11, PC-62 |
| Health data, watch, Live Activities | No web access | No web access | Not available to a web app. D-17 accepts this. | PC-11 |

## 4. Web client stack

### 4.1 Comparison summary

The researchers scored seven stacks from 1 (poor) to 5 (best). The scores are their judgement from the evidence in this section, so the class is assumption. D-84 records the owner choice.

| Criterion (weight) | React 19 + Vite 8 | Svelte 5 + Kit SPA | Vue 3.5 | Solid 1.9/2.0 | Preact + compat | Angular 22 | Lit 3 |
|---|---|---|---|---|---|---|---|
| AI-agent suitability: corpus, API stability, lint guardrails (25) | 5 | 3 | 4 | 2 | 3 | 2 | 2 |
| Reuse of Decktome conventions and skills (20) | 5 | 1 | 1 | 2 | 3 | 1 | 1 |
| Connect-RPC, protobuf-es, and TanStack Query integration (10) | 5 | 3 | 3 | 3 | 3 | 2 | 2 |
| Accessible primitives, maturity, and mobile (10) | 5 | 4 | 4 | 3 | 2 | 4 | 2 |
| Offline and PWA tooling maturity (10) | 5 | 3 | 5 | 4 | 5 | 3 | 4 |
| Startup cost on mid-range phones (10) | 2 | 5 | 4 | 5 | 5 | 3 | 5 |
| Testing tooling (5) | 5 | 4 | 4 | 3 | 3 | 4 | 2 |
| Governance and maintenance risk (10) | 5 | 3 | 4 | 2 | 3 | 5 | 4 |
| **Weighted total (out of 5)** | **4.70** | **2.95** | **3.40** | **2.75** | **3.30** | **2.60** | **2.50** |

Sensitivity: with the Decktome reuse weight at 0, React still leads with 4.63, against Vue 4.00 and Svelte 3.44. Svelte leads only when reuse is 0, agent suitability drops to about 15, and startup cost rises to about 40.

### 4.2 Why React won

| Reason | Evidence | Sources |
|---|---|---|
| connect-query requires React | The only official TanStack Query binding for Connect declares React peers, and its README says "Connect-Query does require React". Other frameworks need a hand-written wrapper around the core package. | PC-30, PC-29 |
| Decktome reuse | Decktome runs the same stack. Its conventions, skills, test harness, and Connect patterns carry over with no divergence cost (assumption). D-74 makes the Decktome stack the default. | D-74, D-84 |
| Stability | React 19.1 to 19.3 added no breaking client API. React Compiler 1.0 is stable. Vite 8 runs most plugins unchanged, and vite-plugin-pwa 1.3.0 supports it. | PC-32, PC-33, PC-25 |
| Agent reliability | React led Web-Bench. Svelte ships an autofixer because models write outdated syntax. Solid 2.0 and SvelteKit 3 are at release candidate, and Vue Vapor is not stable. | PC-36, PC-37, PC-38, PC-39 |
| Governance | The React Foundation under the Linux Foundation started on 2026-02-24. Technical governance sits with the maintainers. | PC-32 |
| Testing | Vitest 5 runs on Vite 8, and its browser mode has an official React package. Playwright drives the browser tests. | PC-40, PC-28 |
| Shared layers | connect-web, protobuf-es, the Firebase JS SDK, Dexie, and vite-plugin-pwa work with any framework. They do not decide the choice. | PC-31, PC-29, PC-24 |

### 4.3 Startup cost risk and the Phase 1 profile

React has the largest runtime and the slowest first paint in the benchmark. The benchmark ran on desktop Chrome, and nobody measured a phone.

| Implementation | Compressed size (KB) | First paint (ms) | Source |
|---|---|---|---|
| React hooks 19.2.0 | 51.4 | 221.4 | PC-34 |
| React Compiler hooks 19.0.0 | 50.0 | 207.3 | PC-34 |
| Vue 3.5.39 | 23.3 | 93.9 | PC-34 |
| Svelte 5.42.1 | 9.7 | 58.4 | PC-34 |
| Solid 1.9.3 | 4.5 | 47.3 | PC-34 |
| Budget for a JavaScript-heavy page in 3 s on a Galaxy A24 4G | 0.62 MiB of JavaScript | - | PC-35 |

The gap of about 45 KB sits well inside the 0.62 MiB budget. A service worker precache removes the network cost on each relaunch, so parse and execute time remain (assumption). D-84 makes Phase 1 profile the startup of a production build on a phone. The researchers propose these profile targets, and the owner did not accept them yet:

1. First interaction under about 2.5 s on a cold start.
2. No long task over about 200 ms during launch.
3. Code splits for the camera and history routes.
4. Firebase Auth and Dexie start after the first paint.

### 4.4 Runner-up conditions

Vue 3.5 with Vite 8, vite-plugin-pwa, Reka UI, and TanStack Vue Query came second. It wins only when all three conditions below hold.

1. The Phase 1 profile misses its targets, and code splits can not close the gap.
2. Decktome is no longer the shared template, so the reuse criterion drops out.
3. A small connect-query-core wrapper for Vue proves stable in a spike.

Svelte leads only when startup weight reaches about 40 and agent reliability matters less. Solid is the fastest option, but Solid 2.0 is at release candidate and its primitives are 0.x.

### 4.5 Not applicable after D-17: native and cross-platform clients

D-17 removes native apps. The launch prompt asked for this comparison, so the key facts stay here for a future reopening.

| Option | Current state (2026-09-27) | Connect-RPC finding | Main risk | Sources |
|---|---|---|---|---|
| React Native with Expo | SDK 57: React Native 0.86, Hermes V1, New Architecture mandatory since SDK 55. SDK 58 beta. | Connect-Web works through expo/fetch with no polyfills. Unary works. No pure gRPC transport. | Open SDK 57 expo/fetch defects: iOS hang after a drop, Android truncation, lost abort type, chunk reorder. Use idempotent unary calls with a watchdog. | PC-42, PC-43, PC-44 |
| Flutter | 3.47.5 stable with Dart 3.13.4. Minimum iOS 15. Impeller only on iOS. | Connect-Dart 2.0.0 is stable. The http2 client supports every RPC type on phones. | Small Connect-Dart adoption. Impeller problems on low-end Android GPUs. Key plugins pre-1.0. | PC-45, PC-46 |
| Swift and Kotlin | Two codebases. | connect-swift 1.2.3 is stable. connect-kotlin 0.9.0 is in beta. | About 1.6 to 2 times the work (assumption). Sync logic must match on both platforms. | PC-47, PC-48 |
| Kotlin Multiplatform | Compose Multiplatform for iOS is stable. Room is multiplatform. | connect-kotlin runs on the JVM only. No shared Connect client for iOS. | Custom network plumbing for iOS. | PC-48 |

## 5. Offline logging and sync design

D-62 and D-77 make the phone store logs first and sync later. Firestore, through the API, is the source of record after a sync. The designs below are recommendations, and the owner did not accept them yet.

### 5.1 Local store

| Topic | Recommendation | Sources |
|---|---|---|
| Store | Dexie 4.4.6 on IndexedDB. It needs no Worker, no COOP or COEP headers, and no WASM download. useLiveQuery serves React. | PC-26, PC-29 |
| Fallback | SQLite WASM with the opfs-sahpool VFS in a Worker, only if relational queries become necessary. Avoid the opfs VFS, because its COOP and COEP headers affect every cross-origin resource. | PC-27 |
| Persistence | Call navigator.storage.persist() after the first sign-in inside the installed app. Log the result and show storage.estimate() in diagnostics. | PC-4, PC-12 |
| Install first | The owner installs the app and signs in there before the first log. Browser tab storage is separate, and Safari tabs have the 7-day cap. | PC-7, PC-5 |
| iOS bug | Catch DatabaseClosedError and InvalidStateError, then reopen the database. | PC-3 |

### 5.2 Outbox and idempotent unary sync

The client writes each change and its outbox entry in one Dexie transaction. Connect clients that use fetch support unary and server streaming calls only (PC-31). So the sync API uses idempotent unary calls.

1. Each change gets a client op id, a UUIDv7 that the phone makes.
2. An outbox entry holds the op id, entity, entity id, base version, payload, time, attempts, and schema version.
3. The client sends a batch of entries in one unary SyncOutbox call.
4. The server applies each op in a Firestore transaction keyed by the op id.
5. A repeated op id changes nothing, so a retry is safe.
6. The server returns an acknowledgement and the current version for each op.
7. The client deletes each acknowledged entry from the outbox.
8. A logged set is an append-only event, so sets do not conflict.
9. For a plan edit, the server wins, and the client rebases.
10. Unauthenticated triggers a token refresh, not an op failure.
11. InvalidArgument or FailedPrecondition moves the op to a dead table that the owner can see.

Retries use exponential backoff with jitter. Luna calls need a connection. When the phone is offline, the app shows the cached plan, and it asks Luna after the next sync (assumption).

### 5.3 Sync triggers

iOS has no Background Sync (PC-11). So the app syncs at these points:

- app start,
- visibilitychange to visible,
- the online event,
- a short delay after each logged set,
- a "Sync now" action, with a count of pending changes on the screen.

A fetch with keepalive on visibilitychange to hidden can try a last send. Nothing guarantees that send (assumption).

### 5.4 Update strategy

| Item | Recommendation | Sources |
|---|---|---|
| registerType | prompt, not autoUpdate. The docs warn that autoUpdate can lose data in an open form. A switch later "can be a pain". | PC-24 |
| When to apply | When an update waits and no workout runs, show "Update ready". During a workout, wait until the workout ends. | PC-24 |
| Update checks | Hourly and on visibilitychange to visible. First check navigator.onLine and a no-store fetch of the service worker file. | PC-24 |
| Hosting headers | Cache-Control no-cache on sw.js, index.html, and manifest.webmanifest. A one-year immutable cache on hashed files under assets. | PC-23 |
| Client version | Each Connect call carries a client version header. The server supports the previous client for at least one release. Protobuf changes stay additive. | assumption |
| Too old | The server returns FailedPrecondition with a client-too-old reason. The app shows "Update required" and applies the update. | assumption |
| Manifest | Keep it stable. On Android, each change re-mints the WebAPK. | PC-16 |

### 5.5 Sign-in on the phone

D-75 chose Firebase email and password with an allowlist. Redirect sign-in breaks on Safari 16.1 and later because of storage partitioning (PC-20). Email and password calls do not use the redirect frame, so the break does not apply (assumption). getAuth bundles the popup and redirect resolver (PC-21). So initializeAuth with indexedDBLocalPersistence and no resolver keeps the bundle smaller (recommendation). A local log never waits for a token refresh.

### 5.6 Rest timer

D-59 starts the rest timer when the owner logs a set. D-61 shows it on the screen only.

1. On start, store the rest end time in state and in Dexie.
2. Compute the remaining time from the clock on each tick.
3. Compute it again on visibilitychange, because timers freeze in the background.
4. Request a screen wake lock when the workout starts, and release it when the workout ends.
5. Request the lock again on each visibilitychange to visible.
6. On iOS before 18.4, show advice to keep the screen on.

The wake lock works in Home Screen apps from iOS 18.4 only (PC-2, PC-8). The browser releases it when the page hides or the battery is low (PC-13). Cues stay visual (D-58).

### 5.7 Camera flow

| Step | Recommendation | Sources |
|---|---|---|
| Capture | "Take photo" uses a file input with accept="image/*" and capture="environment". "Choose from library" drops capture. No getUserMedia. | PC-14, PC-9 |
| Accept list | Leave image/heic out, so Safari does not convert the file to HEIC. | PC-9 |
| Decode | createImageBitmap with imageOrientation "from-image". Safari 17+ decodes HEIC. On failure, show "unsupported format". | PC-14, PC-3 |
| Downscale | Long edge at most 1024 to 1536 px, through resize options or a canvas. | PC-14 |
| Encode | JPEG at quality 0.8 through toBlob, or convertToBlob in a Worker. No WebP, because iOS can not encode it. | PC-14, PC-11 |
| Metadata | A canvas re-encode writes a new file without the source EXIF (assumption). A unit test asserts no APP1 or EXIF segment in the output. The Go API decodes and encodes again. | PC-19, PC-10 |
| Bystanders | Optional face blur with the MediaPipe face detector, self-hosted on Firebase Hosting. | PC-60 |
| Format check | OpenAI accepts PNG, JPEG, WEBP, and GIF, and it rejects HEIC. JPEG output meets this. | PC-62 |

## 6. Equipment recognition

D-24 assigns equipment photos to Luna. D-49 makes the owner confirm or correct every machine. D-50 asks for one general photo first, and D-51 and D-55 keep a full manual path. D-54 stores identity and available weights only.

### 6.1 Approaches compared

| Approach | Facts | Fit for Gym Route | Sources |
|---|---|---|---|
| On-device OCR and detection in the browser | Safari ships no FaceDetector, and BarcodeDetector sits behind a flag. The research found no shipped text detection API. WASM OCR libraries were not researched. The native OCR of the first research (Apple Vision, ML Kit) is not available to a web app. | Not available in the same way as native. Optional MediaPipe face blur is the only on-device model. | PC-11, PC-60 |
| Server-side Luna with a catalog list | Luna takes image input and strict structured outputs. The output schema holds an enum of catalog machine ids plus none_of_these. | Recommended. It matches D-24. | PC-61, PC-63 |
| Deterministic OCR fallback | Cloud Vision TEXT_DETECTION: first 1,000 units a month free, then $1.50 per 1,000. | Optional, on a nameplate photo only, when Luna abstains. | PC-56 |
| User-confirmed classification | Manual selection and text entry on every screen. | Always present (D-49, D-51, D-55). | D-49, D-55 |
| Gemini on Vertex AI | The first research recommended Gemini 3.5 Flash-Lite. Gemini 2.5 Flash retires on 2026-10-20. | Context only. D-24 chose OpenAI. | PC-57, PC-58 |

### 6.2 Recommended flow

1. The owner takes one general photo (D-50).
2. The phone downscales it, encodes JPEG, and removes metadata (section 5.7).
3. The Go API decodes and encodes the image again, then checks size and type.
4. The role layer calls Luna with the photo, the catalog enum, and a strict schema.
5. The schema writes observed text and evidence before the machine id.
6. The policy maps the answer to a display tier (section 6.3).
7. The owner confirms, picks another machine, or uses manual entry (D-49, D-55).
8. After the confirmation, the server deletes the photo (D-52).

A catalog of fixed-path resistance and cardio machine types (D-45) can fit in one enum (assumption). Google advises fewer and shorter enum values for large schemas (PC-59). If the catalog grows, a shortlist from a nameplate OCR result keeps the enum small (recommendation).

### 6.3 Confidence, thresholds, and abstention

| Finding | Source |
|---|---|
| LLMs that state their confidence tend to be overconfident. Consistency among samples helps. | PC-49 |
| LLMs and VLMs show high calibration error and overconfidence. | PC-50 |
| Small VLMs state an almost constant confidence. In poor light, accuracy fell from 0.99 to 0.22 while stated confidence held. | PC-51 |
| Counterpoint: RLHF text models can state better-calibrated confidence than their token probabilities. | PC-52 |
| Self-consistency picks the most consistent of several sampled answers. | PC-53 |
| Conformal prediction sizes a candidate set to reach a chosen coverage on calibration data. | PC-54 |
| A reject option trades coverage for a lower error rate. | PC-55 |
| The Luna model page lists no logprobs. With effort other than none, OpenAI tells callers to remove logprobs options. D-24 fixes the effort at medium. | PC-61, PC-66 |

The evidence on text tasks points both ways. For VLMs and degraded photos, it is consistently negative, and gym photos have dim light, glare, and odd angles. So the stated confidence of Luna is at most one weak signal. Logprobs are not available at medium effort. The researchers propose these signals instead:

- agreement among 3 to 5 Luna samples for the top machine id,
- agreement between the category that Luna names and the category of the chosen machine,
- a match between any OCR text and the catalog aliases,
- image quality flags from the schema, such as blur, glare, or no nameplate.

| Tier | Rule (thresholds from the evaluation set, not from intuition) | Screen |
|---|---|---|
| High | Score at or above a threshold with at least 98% precision on the evaluation set | One machine card with "Not this?" |
| Medium | Score between the two thresholds | Top 3 candidate cards with their differences |
| Low or none_of_these | Score below the low threshold | "Machine not identified", a retake prompt, and manual search |

Every tier ends with the owner confirmation (D-49). So a tier only chooses the screen. It never assigns a machine silently. The first research also proposed a confusable-machine list and one question to separate close candidates.

### 6.4 Cost

| Item | Arithmetic | Result | Class | Source |
|---|---|---|---|---|
| Luna token prices | Standard tier, per 1M tokens | $0.10 input, $0.50 output, $0.01 cached input | evidence | PC-61 |
| Luna image tokens | The vision guide gives 32 px patches and a multiplier for other models. It gives no Luna multiplier. | Unknown | unresolved | PC-62 |
| One 1024 by 1024 photo, if the 1.2 multiplier applies | 1,229 tokens x $0.10 / 1M | About $0.00012 | assumption | PC-62, PC-61 |
| One photo call with 1,500 text tokens and 300 output tokens | (1,229 + 1,500) x $0.10 / 1M + 300 x $0.50 / 1M | About $0.00042 | assumption | PC-61 |
| Five self-consistency samples | Five separate calls | About $0.0021 | assumption | PC-61 |
| US data residency | 10% uplift for models released on or after 2026-03-05 | x 1.1 | evidence | PC-65 |
| Context: Gemini 3.5 Flash-Lite, 3 photos | 4,860 x $0.30 / 1M + 300 x $2.50 / 1M | About $0.0022 per session | evidence (prices) | PC-57 |
| Context: Gemini 3 image tokens | Default media resolution | 1,120 tokens per image | evidence | PC-59 |
| Cloud Vision OCR fallback | $1.50 per 1,000 units after 1,000 free a month | $0.0015 per photo, free for one user | evidence | PC-56 |

One owner with one gym takes tens of photos in total (assumption). So the recognition cost stays far below one dollar. A paid probe of the true Luna image rate needs owner approval (D-25).

### 6.5 Privacy

| Risk | Control | Sources |
|---|---|---|
| EXIF and GPS | The iOS picker keeps EXIF and strips GPS by default from iOS 16.4, and iOS 17 lets the user include location. So the app removes metadata itself on the phone and again on the server. | PC-10, PC-14 |
| Bystanders | Photos can show other people. Optional face blur on the phone. Luna "cannot identify people" per the vision guide. | PC-60, PC-62 |
| Retention at OpenAI | Send store false. With store true, the Responses API keeps application state for 30 days. | PC-65 |
| Abuse monitoring | OpenAI keeps abuse logs up to 30 days. Image inputs get a CSAM scan, and flagged images stay for review even under ZDR. | PC-65 |
| Zero data retention | Needs prior approval from OpenAI. Not requested. | PC-65 |
| Region | US data residency on us.api.openai.com needs no approval and adds 10%. | PC-65 |
| Training | API data does not train OpenAI models unless the customer opts in. | PC-65 |
| Retention at Gym Route | Delete the photo after the confirmation (D-52). No user photo in an evaluation set (D-53). | D-52, D-53 |
| Telemetry | Ids only, with no photos or prompts (D-80). | D-80 |

### 6.6 Evaluation on licensed images

D-56 limits the recognition test set to public or properly licensed images. D-53 keeps every user photo out of it. The researchers propose this test set design:

- each catalog machine type, with photos from several gyms, lights, and angles,
- hard negatives: same brand with a different series, and the same movement from different brands,
- plate-loaded and selectorized versions of the same movement,
- out-of-catalog machines, to measure correct abstention,
- degraded photos with blur, low light, glare, and partial occlusion.

Split the set by gym, not by photo, and hold out a calibration split for the thresholds. Measure top-1 and top-3 accuracy, precision at each tier, abstention precision, and the risk-coverage curve. Run the set again after each change of model, prompt, schema, or catalog. Luna shows no dated snapshot (PC-61), so a silent model change is possible, and a periodic rerun catches it.

The recognition test set of work area 1.2 applies this design. `tools/spikes/recognition_set/README.md` describes it. A finding of the curation on 2026-09-28: licensed photos of cardio machines are plentiful, but licensed photos of selectorized strength machines are few. Most of them show the face of a person, and few name a gym. Commons and Flickr through Openverse gave no usable photo of a seated row machine or a back extension machine. The manifest records each such gap.

### 6.7 Risk of confident wrong identification

A wrong machine identity gives the policy a wrong exercise and a wrong load scale. The model can state high confidence while it is wrong, as the papers in section 6.3 show. The owner confirmation of D-49 is the main control. Each candidate card needs a distinguishing detail, such as seated or standing, so a quick confirmation stays meaningful (recommendation). The load estimate of D-41 comes from the owner after the confirmation, not from the photo.

## 7. OpenAI gpt-6-luna facts

| Item | Fact | Class | Source |
|---|---|---|---|
| Model id | gpt-6-luna. The page lists the alias only, with no dated snapshot. | evidence, snapshot unresolved | PC-61 |
| Price per 1M tokens | Input $0.10, cached input $0.01, cache writes $0.125, output $0.50. Batch and Flex at 50%. Fast mode at 2x. | evidence | PC-61 |
| Other models | gpt-6-sol $2.00 and $10.00. gpt-6-astra $10.00 and $50.00. | evidence | PC-61 |
| Context | 1,050,000 tokens in. Up to 128,000 tokens out. Knowledge cutoff 2026-05-18. | evidence | PC-61 |
| Reasoning effort | none, low, medium (default), high, xhigh, max. Reasoning tokens bill as output. | evidence | PC-61, PC-64 |
| Endpoints | Chat Completions, Responses, Batch. | evidence | PC-61 |
| Features | Streaming, structured outputs, function calling, file search, image input, web search, prompt caching. No logprobs listed. | evidence | PC-61 |
| Structured outputs | text.format with type json_schema and strict true. Every field required, additionalProperties false. A refusal field replaces the schema output. | evidence | PC-63 |
| Images plus structured outputs | The guide does not state the combination. A smoke test confirms it. | unresolved | PC-63 |
| Image input | PNG, JPEG, WEBP, non-animated GIF. Detail low, high, original, auto. Up to 1,500 images and 512 MB a request. | evidence | PC-62 |
| Logprobs | With effort other than none, remove top_logprobs and the logprobs include. | evidence | PC-66 |
| Data controls | No training by default. Abuse logs up to 30 days. ZDR needs approval. store false avoids the 30-day application state. | evidence | PC-65 |
| Data residency | US residency through us.api.openai.com without approval, with a 10% uplift. EU only on Standard processing. | evidence | PC-65, PC-61 |
| Rate limits | Tier 1: 500 RPM and 500,000 TPM. Tier 5: 30,000 RPM and 180M TPM. | evidence | PC-61 |
| Flex | Batch rates, slower, can return 429 Resource Unavailable with no charge. | evidence | PC-67 |
| Safety | Free Moderation API. Hashed safety_identifier for each user. Adversarial and prompt-injection tests. | recommendation | PC-68 |
| Usage policies | The dated copy forbids tailored advice that needs a license, such as medical advice, without a licensed professional. It forbids automated high-stakes medical decisions without human review. General workout plans are not on the list. The owner accepted this copy (D-93). | evidence, dated copy | PC-101 |
| Launch post | HTTP 403. The release date near 2026-09-22 comes from a changelog title only. | unresolved | PC-70 |

Tier 1 limits exceed the need of one user by far (assumption). The prompts and the policy keep Luna text inside the fitness boundary of D-36 (D-93).

## 8. Google Cloud for this product

### 8.1 Recommended baseline under D-76

D-76 allows one development project in `us-central1`. Firebase advises one project for each environment and no real user data in development (PC-71). D-76 departs from that advice on purpose, so this one project holds the real data of the owner. The rows below are recommendations unless a D- id marks them.

| Item | Baseline | Reason | Sources |
|---|---|---|---|
| Plan | Blaze | Cloud Storage for Firebase needs Blaze from 2026-02-03. Cloud Run needs billing. | PC-80 |
| Region | us-central1 for every service (D-76) | Cloud Run Tier 1. Firestore operations at half the nam5 price. Always Free Cloud Storage. | PC-72, PC-74, PC-80 |
| Firestore | Standard edition, Native mode, regional us-central1, default database | Simple, predictable pricing. The location is permanent. Regional SLA 99.99%. | PC-75, PC-74 |
| Rules | Deny-all Firestore and Storage rules, deployed from the repo | Server libraries bypass rules. Only the API reads and writes data (D-77). | PC-77 |
| Service account | roles/datastore.user, object admin on the photo bucket only, secret access on named secrets, token creator on itself | Least privilege. Keyless URL signing. | PC-77, PC-78 |
| Cloud Run API | Request-based billing, min instances 0, startup CPU boost, default run.app URL with CORS to the app origin (D-82) | No charge while idle. Go binaries start fast. | PC-72, PC-73 |
| Photo bucket | Uniform bucket-level access, public access prevention, signed URLs of 15 minutes or less | Public access prevention does not block signed URLs. | PC-78 |
| Soft delete | 0 on the photo bucket | The default keeps deleted objects 7 days and bills them. D-52 wants the photo gone. | PC-79 |
| Bucket Lock | None | A locked retention policy is permanent and blocks the deletion of D-52. | PC-79 |
| Lifecycle | Delete objects older than 1 day as a backstop | Lifecycle changes take up to 24 h. | PC-79 |
| Photo purge | A Go Cloud Run job, started by Cloud Scheduler | Firebase Extensions shut down on 2027-03-31. Do the deletion in Go. | PC-85, PC-73 |
| Secrets | Secret Manager, versions pinned, read at startup | The OpenAI API key never enters the repo. | PC-81 |
| Images | Artifact Registry in us-central1, cleanup keeps the last 10 | 0.5 GiB free. | PC-81 |
| Build | Cloud Build trigger on main (D-14, D-18) | 2,500 free minutes a month on the default pool. Linux machine types only. | PC-81 |
| Budget | Alerts-only project budget at 50%, 90%, and 100% | Budgets do not cap spend. | PC-82 |
| Spend cap | Cloud Run spend cap (Preview) | Caps cover Vertex AI, Gemini API, Cloud Run, and functions. They do not cover OpenAI. | PC-83 |
| AI caps | Monthly caps in the role layer for the user and the project (D-25) | GCP caps can not see OpenAI spend. OpenAI-side limits were not researched. | PC-83 |
| Billing shutdown | Do not automate it | It can delete resources for good, and this project holds the only copy. | PC-82 |
| Logging | Structured JSON to stdout, ids only (D-80) | 50 GiB a month free. Error Reporting has no charge. | PC-84 |
| Auth | Firebase email and password with an allowlist (D-75) | No-cost up to 50,000 MAU. | PC-88 |
| App Check | Optional. Skip at first under D-67. | A reCAPTCHA Enterprise key for the web.app domain costs nothing below 10,000 assessments a month. The Go Admin SDK verifies tokens. | PC-22, PC-86 |
| Hosting | Default web.app URL (D-17, D-81) | Free subdomain with SSL. | PC-88 |

### 8.2 Costs

| Service | Free tier | Rate after the free tier | Source |
|---|---|---|---|
| Firestore Standard, us-central1 | 1 GiB, 50,000 reads, 20,000 writes, 20,000 deletes a day, one free database a project | Reads $0.03, writes $0.09, deletes $0.01 per 100,000. Storage $0.15 a GiB-month. | PC-74 |
| Firestore PITR and backups | Never free | PITR $0.15 a GiB-month. Backup $0.03 a GiB-month. Restore $0.20 a GiB. | PC-74 |
| Cloud Run, request-based | 180,000 vCPU-s, 360,000 GiB-s, 2 million requests a month, per billing account | $0.000024 a vCPU-s active. $0.40 a million requests. | PC-72 |
| Cloud Run min instance, if ever used | None known | About $9.86 a month for 1 vCPU and 512 MiB idle | PC-72 |
| Cloud Storage | 5 GB-months Standard in us-central1 | Not read | PC-80 |
| Secret Manager | 6 active versions, 10,000 access operations a month | $0.03 per 10,000 accesses | PC-81 |
| Artifact Registry | 0.5 GiB | About $0.10 a GiB-month | PC-81 |
| Cloud Build | 2,500 minutes a month on e2-standard-2 | $0.006 a minute | PC-81 |
| Cloud Scheduler | 3 jobs per billing account | $0.10 a job per 31 days | PC-81 |
| Cloud Logging | 50 GiB a project a month | $0.50 a GiB | PC-84 |
| Firebase Hosting | 10 GB storage, 360 MB a day transfer | $0.026 a GB stored, $0.15 a GB sent | PC-88 |
| reCAPTCHA Enterprise (App Check) | 10,000 assessments a month | $8 flat to 100,000 | PC-22 |

One owner stays inside every free tier above (assumption). The paid parts are PITR, backups, and Luna calls.

### 8.3 Emulators and environment variables

| Variable or fact | Value or rule | Source |
|---|---|---|
| FIRESTORE_EMULATOR_HOST | 127.0.0.1:8080, with no scheme | PC-87 |
| FIREBASE_AUTH_EMULATOR_HOST | 127.0.0.1:9099. With it set, Admin SDKs accept unsigned ID tokens. Firebase says never to set it in production. | PC-87 |
| FIREBASE_STORAGE_EMULATOR_HOST | 127.0.0.1:9199. The Go Admin SDK copies it into STORAGE_EMULATOR_HOST. | PC-87 |
| App Check | No emulator. The Go appcheck package always fetches the live JWKS. | PC-87, PC-86 |
| Java | The Firestore emulator soon requires Java 21. | PC-87 |
| Signed URLs | Emulator support for V4 signed uploads is not stated. | unresolved |
| No emulator | OpenAI, Cloud Run jobs, and Cloud Scheduler. The fake provider of D-24 covers OpenAI. | PC-87 |

FIREBASE_AUTH_EMULATOR_HOST must never exist in a deployed service. With it, the API accepts forged tokens. The researchers propose a startup guard: the Go API refuses to start when an emulator variable exists and K_SERVICE shows Cloud Run.

### 8.4 The backup question

D-76 leaves one project, and that project holds the only server copy of the workout history. The phone copy is not durable, because iOS can evict web storage (PC-4). Firestore offers PITR for 7 days and scheduled backups for up to 14 weeks (PC-76). A backup restores into a new database, and it survives the deletion of its source database. Neither feature is free, but the cost for one user is a few cents a month (assumption). The research did not check whether a backup survives the deletion of the project.

## 9. Store policy and native distribution: not applicable after D-17 and D-67

The app never goes to a store (D-17), and it serves the owner alone (D-67). The rules below do not apply now. They stay here so that a reopening can start from them.

| Area | Finding (2026-09-27) | Source |
|---|---|---|
| Apple account deletion | 5.1.1(v): an app with account creation offers account deletion inside the app. Deactivation alone is not enough. | PC-89 |
| Apple third-party AI | 5.1.2(i), since 2025-11-13: disclose sharing with third-party AI and get explicit permission first. | PC-89 |
| Apple medical scrutiny | 1.4.1: medical apps get greater scrutiny. Remind users to check with a doctor. | PC-89 |
| Apple medical device status | Since 2026-03-26, a new Health and Fitness app in the EEA, UK, or US declares its regulated medical device status. | PC-90 |
| Apple privacy labels | Likely data types: Health, Fitness, Photos or Videos, Email Address, User ID. | PC-91 |
| TestFlight | A build runs for up to 90 days. 100 internal testers. Up to 10,000 external testers after a review of the first build. | PC-92 |
| Play account deletion | An in-app path and a web link resource that works without the app. | PC-94 |
| Play health declaration | Every app completes it, testing tracks too. Gym Route fits "Activity and Fitness". | PC-93 |
| Play disclaimer | Other health apps state "not a medical device and does not diagnose, treat, cure, or prevent any medical condition". | PC-93 |
| Play Data safety | Required for closed, open, and production tracks. Not for internal testing alone. | PC-94 |
| Play new personal accounts | A closed test with at least 12 testers opted in for 14 days before production. | PC-95 |
| FDA general wellness | Claims on strength and muscle size are wellness claims. Disease, treatment, and rehabilitation claims fall outside. This one still frames D-36. | PC-96 |

A reopening also needs public privacy and support pages, which D-81 now excludes.

## 10. Accessibility: deferred by D-72

D-72 defers accessibility work. D-71 still makes large targets and few taps a formal requirement. The numbers below give a scale for D-71 and a start for the reopening gate that `docs/design.md` names.

| Standard | Target size and other facts | Status | Source |
|---|---|---|---|
| WCAG 2.2 SC 2.5.8 (AA) | At least 24 by 24 CSS pixels, with spacing, equivalent, inline, user agent, and essential exceptions | W3C Recommendation | PC-97 |
| Apple HIG | 44 by 44 pt default, 28 by 28 pt minimum. Contrast 4.5:1 for small text. Text scaling to 200%. | Native guidance | PC-99 |
| Android | At least 48 by 48 dp. Contrast 4.5:1 for small text. | Native guidance | PC-100 |
| WCAG2Mobile | Group Draft Note of 2025-05-06. Informative, sets no requirement. The unit of conformance is one screen. | Draft | PC-98 |
| Automated checks | axe with Playwright finds some problems. Manual tests find many others. | Tool guidance | PC-28 |

## 11. Unresolved items and open recommendations

### 11.1 Unresolved items

| Item | Why it matters | Next step | Source |
|---|---|---|---|
| Safari 27.0 web app fixes | A fix or regression can change the capability table. | Read the notes in a browser. | PC-18 |
| Chrome on iPhone Home Screen install | The owner uses Chrome (D-29). | Test on the owner's iPhone in Phase 1. | PC-1 |
| Canvas re-encode removes EXIF | Photo privacy. | Unit test on the output bytes. | PC-19 |
| Luna image token rate | Recognition cost. | PR-4 of `docs/roadmaps/phase-1-risk-spikes.md` measures it, with a cap of 2 USD (D-94). | PC-62 |
| Luna dated snapshot | A silent model change can shift behavior. | Check the model page again. Rerun the test set. | PC-61 |
| Image input with strict structured outputs | The recognition schema needs both. | Smoke test through the fake-provider seam. | PC-63 |
| React startup on a phone | The only weak score of React. | The Phase 1 profile (D-84). | PC-34, PC-35 |
| Preact compat with React 19 | Only relevant to a later switch. | None now. | PC-41 |
| Web OCR in WASM | A possible on-device shortlist. | Research only if the catalog outgrows one enum. | none |
| Cloud Run free tier and idle min instances | A cost only if min instances rise above 0. | Check the billing report. | PC-72 |
| Storage emulator and V4 signed URLs | Local tests of the photo upload. | Spike in Phase 2. | PC-87 |
| Firestore backups after project deletion | The one project holds the only copy (D-76). | Read the backup docs again. | PC-76 |
| OpenAI-side spend limits | D-25 caps. | Research the OpenAI project settings. | none |
| MediaPipe browser support | Optional face blur. | Device test. | PC-60 |

### 11.2 Recommendations not yet owner decisions

Each row below is a recommendation of this research. None is an owner decision until `docs/decisions.md` records it.

| Id | Recommendation | Status | Section |
|---|---|---|---|
| REC-1 | Dexie on IndexedDB, an outbox with UUIDv7 op ids, and idempotent unary sync keyed by op id. | Recommendation, not an owner decision | 5.1, 5.2 |
| REC-2 | Sync on start, visible, online, after each set, and on demand. No reliance on Background Sync. | Recommendation, not an owner decision | 5.3 |
| REC-3 | registerType prompt, no update during a workout, and no-cache headers on sw.js, index.html, and the manifest. | Recommendation, not an owner decision | 5.4 |
| REC-4 | A client version header, additive protobuf changes, and an "Update required" screen. | Recommendation, not an owner decision | 5.4 |
| REC-5 | Install and sign in inside the Home Screen app before the first log. Call persist(). | Recommendation, not an owner decision | 5.1 |
| REC-6 | Rest timer from a stored end time. Screen wake lock with a request again on each return to view. | Recommendation, not an owner decision | 5.6 |
| REC-7 | Camera: file input with capture, no image/heic, downscale, JPEG re-encode, an EXIF test, and a server re-encode. | Recommendation, not an owner decision | 5.7 |
| REC-8 | Luna with a strict schema, a catalog enum, and none_of_these. Confidence from sample agreement, not from stated confidence. | Recommendation, not an owner decision | 6.2, 6.3 |
| REC-9 | Tier thresholds from the licensed test set, rerun after each model, prompt, schema, or catalog change. | Recommendation, not an owner decision | 6.3, 6.6 |
| REC-10 | Cloud Vision OCR on a nameplate photo only when Luna abstains. | Recommendation, not an owner decision | 6.1 |
| REC-11 | OpenAI requests with store false, a hashed safety_identifier, and the US data residency endpoint. | Recommendation, not an owner decision | 6.5, 7 |
| REC-12 | Firestore Standard, Native mode, regional us-central1, with deny-all rules. | Recommendation, not an owner decision | 8.1 |
| REC-13 | PITR and a daily backup on the one project. | Recommendation, not an owner decision | 8.4 |
| REC-14 | Photo bucket with UBLA, PAP, soft delete 0, no Bucket Lock, a 1-day lifecycle backstop, and short signed URLs. | Recommendation, not an owner decision | 8.1 |
| REC-15 | Cloud Run request billing with min instances 0, a Cloud Run spend cap, an alerts budget, and role-layer AI caps. | Recommendation, not an owner decision | 8.1 |
| REC-16 | Photo purge as a Go Cloud Run job from Cloud Scheduler. No Firebase Extension. | Recommendation, not an owner decision | 8.1 |
| REC-17 | A startup guard that refuses emulator variables on Cloud Run. | Recommendation, not an owner decision | 8.3 |
| REC-18 | No App Check at first. Add it with a reCAPTCHA Enterprise key if the audience changes. | Recommendation, not an owner decision | 8.1 |
| REC-19 | Phase 1 profile targets: first interaction under 2.5 s, no launch task over 200 ms. | Recommendation, not an owner decision | 4.3 |
| REC-20 | Touch targets of at least 48 CSS pixels as the scale of D-71. | Recommendation, not an owner decision | 10 |
