# Incident evals

**English** · [Português (Brasil)](04-incident-evals.pt-BR.md)

Every incident that review should have caught becomes a regression case. The reviewer runs against the whole set daily; any prompt, model or tool change is scored against all cases and **kept only if the score goes up**.

## The pattern already exists here
The deterministic engine is built this way. [`bench/corpus/<case-id>/`](../bench/README.md) holds:

| File | Content |
|---|---|
| `change.diff` | The unified diff that was reviewed |
| `expected.json` | `description`, `source`, and `findings: [{path, line, class}]` where `line` is the new-file line |

An empty `findings` array means the reviewer must stay silent (clean-PR case). `make bench` (`go run ./cmd/bench`, code in `internal/bench`) scores the engine against the corpus and writes `bench/results.json`. Labels reflect reviewer judgement, never the engine's current output.

## Turn an incident into a case
1. Find the PR that introduced it and the fix-forward or revert.
2. **An engineer confirms** the bug location (`path:line` in the original diff). No confirmation, no case.
3. Save the original diff as `change.diff`; write `expected.json` with the confirmed location and a `class`.
4. Add the clean sibling too (same code with the check present, `findings: []`), so a fix that over-flags is caught.
5. Run the bench. A new failing case is the point: fix the rule, prompt or skill until it passes without breaking others.
6. Record in the incident review which layer should have caught it.

## For LLM reviewers
Use the same corpus shape for `reviewers/*` prompts: run each role over every `change.diff`, compare findings against `expected.json` by `path:line`. Gate a change:

| Change | Accept only if |
|---|---|
| Prompt, model, tool or skill edit | Recall on the full set does not drop **and** the clean-PR test does not get worse |
| New case added | It fails first, then passes with the fix |

Scoring never edits `expected.json` to match output.

## Zero-cost fallback
Skip the LLM: the deterministic engine over `bench/corpus` is free and runs in seconds. For LLM roles, run the set weekly by hand in a harness you already have.

## Failure modes
| Failure | Guard |
|---|---|
| Unverified labels | Engineer confirms location; `source` says who |
| Overfitting to the set | Keep `bench/blind/` cases unseen when tuning |
| Set rots | Daily run; a case that stops failing for the right reason is reviewed |

Next: [05 human roles](05-human-roles.md).
