---
name: warm-up
description: Load project context into the conversation before substantial work — planning, coding, or reviewing.
---

# /warm-up

Front-load project understanding so subsequent tasks are faster and better informed. Run this before substantial work —
skip it for quick questions.

## Steps

Execute in order. Maximize parallelism where noted.

### Step 1 — Load foundation (sequential, main context)

These go into YOUR context directly — you need them for all downstream work:

1. **Read `Makefile`** — defined workflows. Use these; do not invent commands.
2. **Load these following skills** — always, regardless of the task:
   - `go-best-practices`, `diagura-go-best-practices`
   - `frontend-design`, `vercel-react-best-practices`

### Step 2 — Explore the codebase (parallel subagents)

Spawn parallel **Explore subagents** to map the codebase fast. Each returns a **concise summary** — not raw file
contents. This keeps your main context lean.

### Step 3 — Synthesize and report

After all subagents return, produce a brief status report. This confirms warm-up is done and gives the user a chance to
correct misunderstandings before real work begins.
