# Pull request 7 - author response

Date: 2026-09-29. Review round 1 recorded effective head `61d9309`, with the verdict "Changes required".

## P2-1: The startup result does not establish five cold starts

**Result: full merit.**

The trigger reproduces at `61d9309`. D-118 asked for 5 cold starts. The report showed 5 new page loads in 11 seconds, each after an app stop in the app switcher. The probe gives a new launch id to each page load, but no value of the page shows that iOS ended the web view process. Section 5 of the report said that iOS can keep parts of the web view in memory.

**Correction.**

The author asked the owner. The options were five new samples after a restart of the phone, a new definition of a cold start, or new samples after a pause. The owner chose a new definition (Q-139, D-121).

- `docs/decisions.md`: D-121 defines a cold start as a new page load of the Home Screen app after an app stop in the app switcher. The row of D-118 says that D-121 amends it.
- `docs/questions.md`: Q-139 holds the question and the answer.
- `docs/research/iphone-platform-spike.md`: sections 1, 2.3, and 3.5 cite D-121. Section 5 keeps the memory limit. It says that a launch after a restart or a pause can be slower.
- `docs/roadmaps/phase-1-risk-spikes.md`: the result line of PR-6 cites D-121.

The author also corrected an error of its own in the same text. Section 3.1 said that the worker served the page in each of the five launches. The screenshots show that value for the fifth launch only, and the report now says so.

**Regression checks.**

- The five samples meet D-121. Each launch had a new launch id, the `standalone` display mode, and the navigation type `navigate`. The owner stopped the app in the app switcher before each launch.
- The median of 34, 33, 30, 32, and 33 ms is 33 ms.
- `make ste-check`, `make ref-check`, and `make verify` pass.
