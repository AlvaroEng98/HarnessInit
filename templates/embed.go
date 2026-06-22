package templates

import "embed"

//go:embed CLAUDE.md AGENTS.md init.sh feature_list.json claude-progress.md session-handoff.md clean-state-checklist.md evaluator-rubric.md quality-document.md docs data agents claude-settings
var FS embed.FS
