---
name: commit
description: >-
  Runs project checks, stages all changes, reviews the diff, matches recent commit
  message style, commits with a conventional subject and substantive body when useful,
  and pushes. Use when the user asks to commit, ship changes, run the full pre-commit
  workflow, or mirror a "make check → stage → review → commit → push" flow.
---

# Commit workflow

This skill is scoped to **this repository**: the root `Makefile` is the canonical static verification entrypoint (`make lint`).

## Steps

1. **Checks** — From the repository root run `make lint`. If it fails, attempt **trivial** fixes (formatting, obvious typos, small config drift). If the failure is **non-trivial** (unclear root cause, risky behavior change, or anything you cannot fix confidently), **stop and ask** the user before committing.
2. **Status** — Run `git status` and understand what will be included.
3. **Stage** — `git add -A` unless the user asked to commit only specific paths.
4. **Review** — Run `git diff --cached` and confirm the staged change matches intent.
5. **Message** — Inspect `git log -10` and match style. This repo uses **conventional commits** for the subject, e.g. `chore(scope): summary`, `feat: …` (omit scope when none fits).
6. **Commit** — Use a **subject line** plus a **body** when it adds value (see below). Prefer multiple `-m` arguments so Git records subject and body: e.g. `git commit -m "feat(api): add user pagination" -m "Wire cursor-based page params and align OpenAPI."` For longer bodies, additional `-m` blocks or a here-doc are fine.
7. **Push** — `git push` using the current branch’s upstream; set upstream if missing. Assume the developer **intends** to publish this branch—they are aware which branch they are on.

### When to include a commit body

Use a body (not subject-only) when any apply:

- Non-obvious **why** or tradeoffs
- **Breaking changes** or migration notes (consider `BREAKING CHANGE:` footer if the project uses that convention)
- Multiple **concerns** in one commit that need separating in prose
- **Issue / ticket** references or links the team expects in history
- Anything a future reader would miss from the subject alone

A trivial one-line subject is fine when the diff is small and self-explanatory.

## Notes

- **Staged files** — `.gitignore` should keep noise out, but treat the staged set as **untrusted until reviewed**. Flag accidental, irrelevant, or suspicious paths (secrets, local tooling, editor junk, large binaries, unexpected lockfile-only churn, etc.). Do not commit them; ask or unstage as appropriate.
- If there is nothing to stage after checks, say so and stop—do not create an empty commit.
