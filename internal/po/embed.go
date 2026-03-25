package po

import "embed"

// WorkspaceFS embeds the default PO workspace (CLAUDE.md + playbooks/) into the binary.
// No external files needed at runtime — the binary is fully self-contained.
//
//go:embed all:po_workspace
var WorkspaceFS embed.FS
