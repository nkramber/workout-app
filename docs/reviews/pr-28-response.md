# Pull request 28 - author response

Date: 2026-10-03. Review round 1 recorded effective head `26f1283`, with the verdict "Changes required".

## P2-1: The fake delay accepts values above its limit after overflow

**Result: full merit.**

The trigger reproduces at `26f1283`. `strconv.Atoi` accepts `9223372036855`. The product with `time.Millisecond` overflows `time.Duration` to a negative value, and a negative value is not greater than `maxFakeDelay`. So the API started with a fake with no delay, and the range of D-241 did not hold.

**Correction.**

- `go/cmd/api/main.go`: the bound reads the parsed number against `maxFakeDelay.Milliseconds()` before the conversion. A value above 60000 can not overflow into a short delay.

**Regression checks.**

- `TestProviderFromEnvDelay` of `go/cmd/api/main_test.go` adds `9223372036855` to the refused values. The test fails on the code of `26f1283`, and it passes after the correction.
- `make go-test` passes.

## P2-2: The no-exercise error omits the unchanged-plan text

**Result: full merit, after an owner decision.**

The text of D-240 that the owner approved had no "Your plan did not change." for the error of no allowed exercise. The summaries of Q-253 and of `docs/design.md` section 3.3 said that each error has that sentence. The author asked the owner, and quoted both. The owner added the sentence (Q-255).

**Correction.**

- `web/src/lib/plan.ts`: `NO_EXERCISE` is now "No confirmed machine gives an exercise that you can do. Your plan did not change. Confirm a machine, or change the injuries in your profile."
- `docs/decisions.md`: D-240 gives the new text, and names Q-255.
- `docs/questions.md`: Q-255 records the answer of the owner.

**Regression checks.**

- `web/src/lib/plan.test.ts` asserts the full text of `NO_EXERCISE`. The test fails on the text of `26f1283`.
- The browser test "an owner with no confirmed machine gets the error of no allowed exercise" of `web/e2e/plan.spec.ts` asserts the full text in WebKit and in Chromium.
- `make web` and `make verify` pass.
