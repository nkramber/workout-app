# Gym Route - recognition spike report

Status: spike report of work area 1.2 of `docs/roadmaps/high-level-roadmap.md`. It holds measurements and recommendations. The owner decisions live in `docs/decisions.md`.

Date of the run: 2026-09-29 00:50 UTC, which is 2026-09-28 in the time zone of the owner. The harness and the results are in `tools/spikes/recognition/`. The test set is in `tools/spikes/recognition_set/`.

## 1. Result

**No-go.** The recognition risk fails one part of the go bar of D-107. Luna gave 15 wrong answers with a stated confidence of 0.8 or more. That is 8.6% of the photos, and the bar is 2% or less.

| Measure | Result | Go bar (D-107) | Pass |
|---|---|---|---|
| Wrong answers with a stated confidence of 0.8 or more | 8.6% (15 of 174) | 2% or less | No |
| Correct answers | 78.2% (136 of 174) | 70% or more | Yes |
| Mean cost per photo | 0.00037 USD | 0.01 USD or less | Yes |

The owner read this result and decided the change of the roadmap (D-110, D-111, D-112). Phase 4 gives manual selection and text entry only. Photo recognition leaves the roadmap until a later owner decision adds it again. Section 8 gives the details.

## 2. Method

### 2.1 Inputs

- The recognition test set of PR-3: 130 licensed images of 72 gyms, and 44 degraded photos (D-102 to D-106). The run sent all 174 photos (D-108).
- 98 originals show a catalog type. 32 originals show a machine outside the catalog, with the label `none_of_these`.
- The catalog shortlist of 26 types in `tools/spikes/recognition_set/catalog.json`. Each type has a distinguishing detail.
- The set holds no user photo (D-53). Each image has a license of D-95.

### 2.2 Run

| Item | Value |
|---|---|
| Model and effort | `gpt-6-luna`, medium, from the role configuration `tools/spikes/recognition/roles.json` (D-24) |
| Endpoint | Responses API, strict structured outputs, `store` false |
| Photo | EXIF orientation applied, long side 1536 px or less, JPEG quality 80, no metadata, detail `high` |
| Prompt | `recognition-prompt-v1`, hash `24851131a227b755` |
| Answer | Observed text, evidence, quality flags, then a machine type or an abstention, then a stated confidence from 0 to 1 |
| Calls | One call for each photo, 4 calls at a time |
| Owner approval | The owner approved the smoke call and the run in the session of 2026-09-28, before the first paid call (D-25) |
| Cap | 2 USD in total (D-94): 0.05 USD for the smoke call, and 1.95 USD for the run |
| Real cost | 0.0013 USD for the smoke call of 2 photos, and 0.0638 USD for the run. Total: 0.0651 USD |

The answer has two abstentions: `none_of_these` for a machine outside the catalog, and `abstain` for a photo that does not show enough. The request holds the photo and the catalog, and never the photo id or the true type.

The smoke call answered a question that section 7 of `docs/research/platform-cloud-and-ai.md` marked as unresolved. Luna accepts an image together with a strict JSON schema, and each of the 176 answers passed the schema.

### 2.3 Measures

The harness scores each answer:

| True type | Answer | Outcome |
|---|---|---|
| A catalog type | The same type | correct |
| A catalog type | An abstention | abstained |
| A catalog type | Another catalog type | wrong |
| `none_of_these` | An abstention | correct |
| `none_of_these` | A catalog type | wrong |

A stated confidence of 0.8 or more is high (D-107). Each rate of the go bar counts against all the photos of the run. The cost per photo is the total cost over the photos that Luna answered. The prices come from `docs/research/platform-cloud-and-ai.md` section 7, read 2026-09-28.

## 3. Numbers

| Photos | Count | Correct | Wrong | Abstained | Invalid |
|---|---|---|---|---|---|
| All | 174 | 136 | 30 | 8 | 0 |
| Originals | 130 | 105 | 23 | 2 | 0 |
| Degraded | 44 | 31 | 7 | 6 | 0 |
| True type in the catalog | 139 | 109 | 22 | 8 | 0 |
| True type outside the catalog | 35 | 27 | 8 | 0 | 0 |
| Calibration split | 39 | 33 | 6 | 0 | 0 |
| Test split | 135 | 103 | 24 | 8 | 0 |
| Hard negatives | 88 | 63 | 19 | 6 | 0 |

Luna seldom abstains on a catalog photo. It abstained on 8 of 139 catalog photos, and 6 of the 8 come from the degraded group. On the 35 photos outside the catalog, it selected `none_of_these` 27 times, and a catalog type 8 times.

| Degradation | Correct | Wrong | Abstained |
|---|---|---|---|
| Blur | 7 | 0 | 4 |
| Low light | 10 | 1 | 0 |
| Glare | 9 | 2 | 0 |
| Occlusion | 5 | 4 | 2 |

### 3.1 The stated confidence

This table counts the answers that name a catalog type:

| Stated confidence | Answers | Correct | Precision |
|---|---|---|---|
| Under 0.5 | 0 | 0 | none |
| 0.5 to under 0.8 | 27 | 12 | 44.4% |
| 0.8 to under 0.9 | 28 | 20 | 71.4% |
| 0.9 to 1.0 | 84 | 77 | 91.7% |

The confidence has an order: a higher value gives a higher precision. But the value does not agree with the real precision. An answer of 0.9 or more was wrong 7 times in 84. Section 6.3 of `docs/research/platform-cloud-and-ai.md` predicted this result.

A threshold from the calibration split does not remove the risk. The lowest threshold with no wrong answer on the calibration split is 0.89. On the 135 photos of the test split, that threshold still gives 7 wrong answers, 5.2% of the photos, with a precision of 90.0%.

At 0.95 or more, all 63 answers of the full set were correct. The author selected this threshold after a look at all the results, so it is not a valid measurement. A set of this size can not show an error rate under about 5% with confidence (rule of three).

### 3.2 Cost

| Item | Value |
|---|---|
| Input tokens per photo | 1,332 to 3,751, mean 2,815. The prompt caching read 190,850 of the 489,830 input tokens. |
| Output tokens per photo | Mean 368, of which 299 are reasoning tokens |
| Cost per photo | 0.00037 USD at the prices of section 7 |
| Time per photo | Mean 5.5 seconds, max 22.7 seconds |

This measurement answers Q-96. One photo costs about 0.0004 USD, near the estimate of section 6.4 of `docs/research/platform-cloud-and-ai.md`. The input token count includes the text of the prompt, so the harness can not give the image tokens alone.

28 attempts failed and passed on a retry, with 1 or 2 retries for 22 photos. The harness counts each failed attempt as an unknown charge at its worst case. So the spend of the run is 0.2651 USD or less, and the cost per photo is 0.0016 USD or less. Both bounds pass the go bar. The run did not keep the reason of each failed attempt. This pull request corrects the harness, so a later run keeps it.

## 4. Wrong answers with a high stated confidence

| Photo | Group | True type | Answer | Confidence | Note |
|---|---|---|---|---|---|
| R001 | original | `ab_crunch` | `hip_abduction` | 0.93 | |
| R002 | original | `ab_crunch` | `hip_abduction` | 0.88 | |
| R007 | original | `biceps_curl` | `seated_row` | 0.86 | Hard negative |
| R022 | original | `chest_press_stack` | `pec_fly` | 0.92 | Hard negative |
| R033 | original | `hack_squat` | `leg_press_plate` | 0.86 | Hard negative |
| R042 | original | `lat_pulldown` | `assisted_pullup` | 0.88 | Hard negative |
| R044 | original | `leg_curl_lying` | `hip_abduction` | 0.90 | Hard negative |
| R046 | original | `leg_curl_seated` | `hip_abduction` | 0.88 | Hard negative |
| R074 | original | `shoulder_press` | `chest_press_plate` | 0.92 | |
| R104 | original | `none_of_these` | `calf_raise` | 0.90 | A hyperextension bench |
| R113 | original | `none_of_these` | `leg_curl_lying` | 0.82 | A decline bench with leg rollers |
| R126 | original | `none_of_these` | `leg_extension` | 0.86 | Doubtful label, see below |
| D016 | occlusion 2 | `rower` | `stair_climber` | 0.94 | Doubtful item, see below |
| D036 | occlusion 1 | `chest_press_stack` | `pec_fly` | 0.90 | Hard negative |
| D044 | occlusion 1 | `leg_curl_lying` | `hip_abduction` | 0.84 | Hard negative |

The author examined the photos of R001, R044, R113, R126, and D016. Two items are doubtful:

- R126 shows a cable tower with a leg attachment in front. The attachment has rollers, so the answer `leg_extension` has some support.
- D016 is a strong occlusion of a rower. The occlusion hides the rower, and a stair climber stays visible. Luna read "Life Fitness PowerMill" on its side panel.

Without these two items, 13 of 172 photos (7.6%) have a wrong answer with a high confidence. The result stays a no-go.

## 5. Confusion pairs

Each pair gives the true type, then the answer, for all 30 wrong answers:

| True type | Answer | Count |
|---|---|---|
| `biceps_curl` | `seated_row` | 3 |
| `ab_crunch` | `hip_abduction` | 2 |
| `chest_press_stack` | `pec_fly` | 2 |
| `leg_curl_lying` | `hip_abduction` | 2 |
| `leg_curl_lying` | `leg_extension` | 2 |
| `triceps_extension` | `chest_press_stack` | 2 |
| `none_of_these` | `leg_curl_lying` | 2 |
| `none_of_these` | six other types, once each | 6 |
| Other pairs, once each | | 9 |

Most wrong answers are between seated, selectorized machines with pads beside the body. The distinguishing detail of the catalog did not separate them. `hip_abduction` is the answer of 6 wrong answers.

## 6. Limits

- The set holds 174 photos, not the 200 of the roadmap (D-103). Seven types have images from fewer than two gyms, and two types have no image.
- Many images come from manufacturer or event photos, not from the phone of a gym member. The photos of the owner can be easier or harder.
- The label of each image comes from its source title and from the author. Section 4 names two doubtful items. The run found no other label error, but the author examined only five photos.
- The run used one sample for each photo. It did not measure agreement among samples, the top-three candidates, or an OCR match.
- Luna has no dated snapshot (PC-61). A later model version can give other numbers.

## 7. Recommendations (not owner decisions)

- Do not show one machine card from the stated confidence of Luna alone.
- A later photo phase can measure other signals on the same set. Use the calibration split for the thresholds. Examples: agreement among 3 to 5 samples, the top-three candidates, and an OCR match with the catalog aliases.
- A candidate list needs a sharper detail for the seated machines of section 5. For example: "the pads push the knees outward" for `hip_abduction`.
- Keep the fake provider and the harness of this spike for a rerun on a new model or a new prompt.

## 8. The owner decisions after the result

The owner read the result on 2026-09-28 and made three decisions:

- D-110: Phase 4 gives manual selection and text entry only. Photo recognition leaves Phase 4.
- D-111: No phase holds photo recognition. The high-level roadmap records it as deferred, and a later owner decision adds a phase.
- D-112: The iPhone probe of PR-5 has no camera page.

`docs/roadmaps/high-level-roadmap.md`, `docs/roadmaps/phase-1-risk-spikes.md`, and `docs/design.md` record the change.
