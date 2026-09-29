# Recognition spike

This folder holds the harness of the recognition spike, work area 1.2 of `docs/roadmaps/high-level-roadmap.md`. The report is `docs/research/recognition-spike.md`. The spike code stays outside the product code (D-100).

The harness sends each photo of the recognition test set of `tools/spikes/recognition_set/` to Luna, one photo in each call. Each call also holds the catalog shortlist. Luna selects one machine type with a stated confidence, or it abstains.

## Files

| File | Content |
|---|---|
| `roles.json` | The role configuration. It holds the model id, the effort, the photo settings, the prices, and the cap of the run (D-24, D-94). |
| `photos.py` | The photo list of the test set, with the true machine type of each photo. It prepares each photo for the call. |
| `vision.py` | The prompt with its version and hash, the answer schema, the fake provider, and the OpenAI provider. |
| `recognize.py` | The harness. It sends each photo, scores each answer, and applies the go bar of D-107. |
| `test_recognition.py` | The unit tests. `make test` runs them with the fake provider only. |
| `results/` | The results of the paid run. |

The harness uses three modules of `tools/spikes/luna_plan/`: the call loop of the OpenAI provider, the cap gate, and the schema check. So both spikes use the same retries and the same cap rules.

## The answer

The answer schema is strict. It holds these fields, in this order:

1. `observed_text`: the text that Luna can read on the machine.
2. `evidence`: one or two sentences about the parts of the machine.
3. `quality_flags`: blur, low light, glare, occlusion, a partial view, or no nameplate.
4. `machine_type`: a `type_id` of `catalog.json`, or an abstention.
5. `confidence`: the probability from 0 to 1 that the answer is correct.

The two abstentions are `none_of_these`, for a machine outside the catalog, and `abstain`, for a photo that does not show enough. The schema enum comes from `catalog.json`, so a new type needs no change here.

## The scores

| True type | Answer | Outcome |
|---|---|---|
| A catalog type | The same type | correct |
| A catalog type | An abstention | abstained |
| A catalog type | Another catalog type | wrong |
| `none_of_these` | An abstention | correct |
| `none_of_these` | A catalog type | wrong |
| Any | No valid answer | invalid |

A stated confidence of 0.8 or more is high (D-107). The go bar of D-107 counts each rate against all the photos of the run.

## The photos

The harness sends each photo in the form that the app will send (section 5.7 of `docs/research/platform-cloud-and-ai.md`). It applies the EXIF orientation, sets the long side to 1536 px or less, and encodes a JPEG with no metadata. This step needs Pillow, a local tool that CI does not install. So the image test skips in CI.

Fill the cache before a run:

```bash
python3 tools/spikes/recognition_set/download.py
python3 tools/spikes/recognition_set/degrade.py
```

## Free run

The fake provider makes no network call. It answers with the true type, and it breaks some answers in a known way. So a free run shows each part of the summary.

```bash
python3 tools/spikes/recognition/recognize.py
```

## Paid run

CAUTION: A paid run spends the OpenAI API account of the owner. The owner approves each paid run at run time (D-25). The cap in `roles.json` is 2 USD (D-94), and the harness refuses a higher cap.

1. Get the approval of the owner for the run.
2. Make sure that the `.env` file at the root of the main checkout holds the line `OPENAI_API_KEY=...` (D-109). Git ignores this file.
3. Load the file into the command of the harness only, and run a smoke call. The commands start in the worktree, next to the main checkout:

```bash
(set -a; . ../workout-app/.env; set +a; python3 tools/spikes/recognition/recognize.py --provider openai --owner-approved --photos R001,R002 --cap-usd 0.05)
```

4. Run all the photos with the rest of the cap:

```bash
(set -a; . ../workout-app/.env; set +a; python3 tools/spikes/recognition/recognize.py --provider openai --owner-approved --cap-usd 1.95)
```

Before each attempt, retries included, the harness reserves the worst-case cost of the attempt. It skips each photo that the cap can not cover, and it records the photo as `cap_skip`. An attempt with an unknown charge, such as a timeout or a server error, counts at its worst-case cost. Each request sets `store` to false. The request holds the photo and the catalog, and never the photo id or the true type.

## Results

The harness writes each run to `.local/spikes/recognition/<run id>/`, which Git ignores:

- `answers.jsonl`: one record for each photo, with the raw text, the answer, the outcome, the tokens, and the cost.
- `summary.json`: the counts, the rates, each group, the confidence bins, each high-confidence wrong answer, the cost, and the go result of D-107.
- `summary.md`: the same summary as Markdown tables.
