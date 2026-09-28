# Pull request 3 - author response

Date: 2026-09-28. Review round 1 recorded head `365c204`, with the verdict "Changes required".

## P2-1: Retries can spend beyond the reserved cap

**Result: full merit.**

The trigger reproduces at `365c204`. An opener that always times out made the OpenAI provider send 3 attempts under a budget of 1.5 worst-case calls. The budget counted 0 USD. The worst-case charge of the 3 attempts was 2 times the cap. So the cap of D-98 did not bound the spend of a run with retries.

The paid run of the report made no retry: 60 calls, 0 retries, 0 errors. So each plan made one attempt, and the old guard also bounded that run. The numbers of the report do not change.

**Correction.**

- `tools/spikes/luna_plan/harness.py`: the new `Gate` reserves the worst-case cost before each attempt, retries included, and settles each attempt after it. An attempt with an unknown charge settles at the worst-case cost. So the spend of the budget stays an upper bound under the cap.
- `tools/spikes/luna_plan/providers.py`: each provider asks the gate before each attempt. The OpenAI provider stops the retries when the cap can not cover one more attempt. A timeout, a network error, a bad body, and a server error report an unknown charge. A client error, 429 included, reports no charge.
- The record of each plan holds the known cost and the count of attempts with an unknown charge. The summary holds the total count and the spend bound.
- `tools/spikes/luna_plan/README.md` and section 2.2 of `docs/research/luna-plan-spike.md` state the new guard. The report also states that the run used the first guard, with no retry.

**Regression checks.**

- `RetryCapTest.test_timeouts_stop_at_the_cap`: the cap covers 1.5 worst-case calls, and each attempt times out. The provider sends 1 attempt, and the spend stays under the cap.
- `RetryCapTest.test_each_unknown_charge_counts_at_the_worst_case`: 3 timeouts under a large cap count 3 worst-case charges.
- `RetryCapTest.test_server_error_is_unknown_and_429_is_free`: a 503 counts at the worst case, a 429 counts nothing, and the success counts its real cost.
- `RetryCapTest.test_no_attempt_when_the_cap_is_spent`: no attempt starts when the cap can not cover it.
- The first three tests fail on the code of `365c204`: 1 failure and 2 errors. All 44 tests of the spike pass on the correction.
