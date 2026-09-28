# PR review: the project contracts

Part of the `pr-review` skill. Load this file when the pull request changes code, tools, or CI, or the text of a contract below.

## The contracts

Apply each row that the change touches. Record why a row does not apply when its absence can mislead the owner. Today the repo holds documents and process tools alone. A row about product code applies when that code arrives.

| Area | Examine |
|---|---|
| Safety | The model proposes, and a deterministic, versioned policy checks each set, load, and change before the user sees it (D-23). A change of that behavior is a change of safety behavior (D-15). |
| The API contract | The Protobuf files are the one contract of the Connect-RPC API. Generated code comes from the generator alone, and CI fails on stale output. |
| The model layer | The model plans and revises workouts (D-22). No model id at a call site. A new call to a provider names its cost. |
| User data | Log ids, never workout text, photos, prompts, or health details (D-80). |
| Money | A paid target has a spend guard. The reviewer never runs a paid target. |
| Go code | `ctx` is the first parameter. Use `errors.Is` and `errors.As`, and table tests. Accept interfaces, and return structs. |
| The web app | The app has one phone layout (D-20). Check each changed screen at phone width. |
| CI | Each action pins a full commit SHA. A value of the event enters a script through `env`, and never through an expression. A `pull_request_target` job never runs a file of the head. |
| The deploy | Only `main` deploys (D-14). Google Cloud hosts the API, the data, and the web app (D-18). |
| Git | Each commit subject is a Conventional Commits subject, and no commit holds AI attribution (D-14). |
| Documents | Each document follows ASD-STE100 (D-83), and `make ste-check` and `make ref-check` pass. |
