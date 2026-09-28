---
name: ste-writing
description: Write and review text in ASD-STE100 Simplified Technical English (Issue 8). Load before you write any .md, skill, or agent file in this repo (D-83).
---

# STE writing skill

Use this skill before you write text in this repo. The owner requires ASD-STE100 for every document and skill (D-83). The skill is a port of the Decktome skill.

Source: ASD-STE100 Issue 8 (2021-04-30), Part 1, Writing rules. Issue 9 (2025-01) supersedes it with the same 53 rules. The full standard is free at https://www.asd-ste100.org/. This skill gives the 53 rules in short form. It does not copy the dictionary.

## Procedure

1. Write the text.
2. Check each sentence against the checklist below.
3. Correct each sentence that fails.
4. Read the text again as a reader who does not know the subject.

## Checklist (the rules that fail most often)

- Max 20 words in a procedural sentence. Max 25 words in a descriptive sentence (5.1, 6.3).
- One instruction per sentence (5.2).
- Instructions in the imperative: "Load the file." Not "The file should be loaded." (5.3).
- Active voice in procedures. Active voice as much as possible in descriptions (3.6).
- No "-ing" verb forms. "Sync the data" not "Syncing the data". The rules permit an "-ing" word only in a technical name (3.5).
- No helping verbs for complex tenses: `we did`, not `we have been doing` (3.4).
- Tenses allowed: infinitive, imperative, simple present, simple past, past participle as adjective, future (3.2).
- No semicolons (8.1).
- No contractions (4.2).
- Max three words in a noun cluster. Write longer names in full, then use hyphens or a short name (2.1, 2.2).
- Use "the", "a", "this" before nouns (2.3).
- One term per concept. Do not use synonyms for variety (1.11, 9.4).
- Each paragraph: one topic, max six sentences (6.5, 6.6).
- Use vertical lists for complex content (4.3).
- Start a safety note with the risk word: WARNING, CAUTION (7.1).
- Notes give information, not instructions (5.5).
- American English spelling (1.14).
- No phrasal verbs: "remove" not "take out" (9.3).
- Do not use a technical name as a verb (1.7). Write "make a backup", not "backup the data".

## The 53 rules in short form

### Section 1 - Words
- 1.1 Use only approved dictionary words, technical names, and technical verbs.
- 1.2 Use approved words only as the part of speech given.
- 1.3 Use approved words only with their approved meaning.
- 1.4 Use only approved forms of verbs and adjectives.
- 1.5 You can use words that fit a technical name category.
- 1.6 Use an unapproved word only when it is a technical name or part of one.
- 1.7 Do not use technical names as verbs.
- 1.8 Use technical names that agree with approved nomenclature.
- 1.9 Select technical names that are short and easy to understand.
- 1.10 Do not use slang or jargon as technical names.
- 1.11 Do not use different technical names for the same item.
- 1.12 You can use verbs that fit a technical verb category.
- 1.13 Do not use technical verbs as nouns.
- 1.14 Use American English spelling, unless an official directive says otherwise.

### Section 2 - Noun clusters
- 2.1 Write noun clusters of max three words.
- 2.2 Write a long technical name in full, then give a short name or use hyphens.
- 2.3 Use an article or demonstrative adjective before a noun.

### Section 3 - Verbs
- 3.1 Use only verb forms given in the dictionary.
- 3.2 Make only: infinitive, imperative, simple present, simple past, past participle as adjective, future.
- 3.3 Use the past participle only as an adjective.
- 3.4 Do not use helping verbs to make complex verb structures.
- 3.5 Use the "-ing" form only as a technical name or in a technical name.
- 3.6 Use the active voice in procedures. Use it as much as possible in descriptions.
- 3.7 Use an approved verb to describe an action, not a noun.

### Section 4 - Sentences
- 4.1 Write short and clear sentences.
- 4.2 Do not omit words or use contractions to make sentences shorter.
- 4.3 Use a vertical list for complex text.
- 4.4 Use connecting words to connect sentences with related topics.

### Section 5 - Procedures
- 5.1 Max 20 words in each sentence.
- 5.2 One instruction in each sentence, unless actions occur at the same time.
- 5.3 Write instructions in the imperative.
- 5.4 Divide a descriptive statement from the command with a comma.
- 5.5 Write notes only to give information, not instructions.

### Section 6 - Descriptions
- 6.1 Give information gradually.
- 6.2 Use key words and phrases to organize the text.
- 6.3 Max 25 words in each sentence.
- 6.4 Use paragraphs to show related information.
- 6.5 Each paragraph has only one topic.
- 6.6 No paragraph has more than six sentences.

### Section 7 - Safety instructions
- 7.1 Use a word such as "WARNING" or "CAUTION" to identify the risk level.
- 7.2 Start a safety instruction with a clear command or condition.
- 7.3 Give an explanation that shows the risk or the possible result.

### Section 8 - Punctuation and word count
- 8.1 Use all standard punctuation except the semicolon.
- 8.2 Use hyphens to connect closely related words.
- 8.3 Use parentheses for references, item identifiers, step identifiers, abbreviations, and singular/plural forms.
- 8.4 In a vertical list, a colon counts as the end of a sentence.
- 8.5 Text in parentheses counts as one word.
- 8.6 Count each number, unit, abbreviation, identifier, quoted text, and title as one word.
- 8.7 A hyphenated word counts as one word.

### Section 9 - Writing practices
- 9.1 Use a different construction when a word-for-word replacement is not enough.
- 9.2 Use each approved word correctly.
- 9.3 Do not make phrasal verbs.
- 9.4 Use a consistent style for terminology and wording.

## Technical names in this project

The rules permit these as written. They are technical names (rule 1.5):
- Fitness terms: training, loading, rounding, rowing, running, cycling, stretching, programming, screening, rating, conditioning, and the names of exercises.
- Software names: Go, Protobuf, Connect-RPC, React, Cloud Run, Cloud Build, Firebase Hosting, Firebase Auth, Firestore, Codex, Claude Code.
- Model names, such as OpenAI Luna.
- Code identifiers in backticks.

The checker holds each fitness noun of this list in its list of allowed "-ing" words. Add a new term to that list in the same pull request that uses it.

## Glossary

One term per concept (rule 1.11). Use the term of the left column. Do not use a synonym for variety.

| Term | Use for | Do not use |
|---|---|---|
| owner | the person who owns the repo and answers each question | maintainer, the name of the owner in prose |
| user | the person who uses the app | athlete, customer, client |
| session | one harness invocation, bound to one pull request (D-12) | run, conversation, chat |
| clean session | a new top-level session that holds no work of another pull request (D-12) | fresh context, new chat |
| pull request | a GitHub pull request | PR in prose, MR, change request |
| milestone | the cohesive result of one pull request, with one acceptance story (D-10) | concern, when you mean the whole result |
| concern | one part of a milestone, such as one tool or one document (D-12) | topic, task |
| acceptance story | the one story that proves the whole milestone (D-10) | exit test, gate test |
| work area | a unit of pull request size in the high-level roadmap (D-9) | PR number, ticket |
| hand-over point | the end of the work of a session on its pull request | hand-off, which names `docs/session-handoff.md` |
| hand-off | the file `docs/session-handoff.md` | handover, notes |
| start read | the files a session reads at start: `AGENTS.md`, the hand-off, and the skills of the task | read order, onboarding |
| context compaction | the harness step that replaces the conversation with a summary | compaction alone |
| reference file | a file of the `references` folder of a skill, which the skill loads for one case | appendix, sub-skill |
| effective head | the newest commit outside the metadata set of the review record | head, which names the branch tip |
| paid target | a `make` target that costs money or spends a plan of the owner | expensive target, live target |

## The checker

`make ste-check` runs `docs/tools/ste-check.py` on every hand-written `.md` file. The checker flags passive voice (3.6) and modal and helper verbs (3.2, 3.4). It flags sentence-initial and preposition-led "-ing" forms (3.5), and the 20-word limit in a numbered step (5.1). It also flags semicolons, contractions, and the 25-word limit.

The passive rule and the participle rule are heuristics. A past participle is an irregular form of the list in the checker, or a word that ends in "ed". So "is required" is a finding. Rewrite the sentence with the actor as the subject: "the build needs the key". The words "can", "must", and "will" pass, because the standard approves them.

The pre-commit hook of `make hooks` runs the checker on each staged `.md` file.

## The reference check

`make ref-check` runs `docs/tools/ref_check.py` on the same file list (D-6). It reads two rules:

- REF 1: a cited `D-` or `Q-` id that no register defines.
- REF 2: a path of this repo in backticks that no file and no folder holds.

`docs/decisions.md` defines each `D-` id with a table row. `docs/questions.md` defines each `Q-` id with a table row. The first cell of the row starts with the id.

- A token with the prefix `decktome:` names an id or a path of the Decktome repo, for example `decktome:D-811`. The check skips it.
- A label of the roadmap, such as "Work area 2.1" or "Phase 3", is no register id. The check skips it.
- REF 2 reads a path with a slash and a first part that names a top-level entry. A bare file name is ambiguous, so the rule skips it.
- A placeholder path, such as `docs/reviews/pr-<n>.md`, takes no rule.
- A path under `.local` takes no rule, because a local run creates it.
- A dated record is history, and a rewrite of it falsifies the record. So the check reads no file that ends with a date.

## The size rules of the context budget

`make context-budget` gives each file of the start read a byte limit (D-6). A session reads each one in full.

| File | Limit |
|---|---|
| `AGENTS.md` | 11,000 bytes |
| `CLAUDE.md` | 1,000 bytes |
| `docs/session-handoff.md` | 24,000 bytes |
| The `## Resume here` section of the hand-off | 6,000 bytes |
| The session records of the hand-off | 3 records, one `Author provider:` line each |
| Each `SKILL.md` file of `.claude/skills` | 36,864 bytes |

The skill limit is 36 KB, and one kilobyte is 1024 bytes. Move the detail to a file of the `references` folder when a skill comes near its limit. A reference file loads for one case, and the skill file loads each time.

The check also compares the paid targets. Each `make` target whose help text says CAUTION must be in the paid-target sentence of `AGENTS.md`.

## Markdown notes

- Tables, code blocks, and front matter are exempt from sentence-length counts. Keep cell text short.
- Headings are titles. They count as one word (8.6).
- Text in backticks, in double quotes, or in parentheses counts as one word (8.5, 8.6).
- A numbered list item is a procedural step, under any heading. Rule 5.1 applies, max 20 words.
- A bullet list item is one unit. Rule 6.3 applies, max 25 words.
