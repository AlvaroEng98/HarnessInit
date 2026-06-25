# Project Name

> AI-assisted Spec-Driven Development scaffold.

This project uses **Spec-Driven Development (SDD)** with AI agents (Claude Code)
to guide features from specification to implementation to review.

## Getting started

1. Run `./init.sh` to verify the environment.
2. Read `AGENT.md` to understand the project structure and workflow.
3. Edit `feature_list.json` to define your project name and features.
4. Start with the first `pending` feature — the AI agent workflow handles the rest.

## Prerequisites

- Python 3 (for `init.sh` validation)
- Claude Code or compatible AI agent (`.claude/agents/` contains agent definitions)

## Project structure

```
├── .claude/agents/       AI agent definitions (orchestrator, developer, reviewer)
├── docs/                 Project documentation and process guides
├── specs/<feature>/      Per-feature specs (requirements, design, tasks)
├── progress/             Session tracking (current + history)
├── src/                  Application source code
├── tests/                Automated tests
├── AGENT.md              Navigation map for AI agents
├── feature_list.json     Feature tracking manifest
├── init.sh               Environment verification
└── CHECKPOINTS.md        Quality criteria for feature completion
```

## Workflow

```
pending → [spec author] → spec_ready → ⏸ human approval → in_progress → [developer → reviewer] → done
```

Each feature flows through the SDD pipeline. Features with `"sdd": true` require
a full spec (requirements, design, tasks) before any code is written.

## Commands

- `./init.sh` — Verify environment and validate state
