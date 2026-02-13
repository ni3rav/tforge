package generate

import (
	"fmt"
	"path/filepath"
	"strings"

	"tforge/internal/snapshot"
)

func Script(s snapshot.Session) (string, error) {
	if len(s.Windows) == 0 {
		return "", fmt.Errorf("session has no windows")
	}

	var b strings.Builder
	b.WriteString("#!/usr/bin/env bash\n")
	b.WriteString("set -euo pipefail\n\n")
	b.WriteString(fmt.Sprintf("SESSION=%q\n", s.Name))
	b.WriteString("\n")
	b.WriteString("if tmux has-session -t \"$SESSION\" 2>/dev/null; then\n")
	b.WriteString("  WINDOWS=$(tmux list-windows -t \"$SESSION\" 2>/dev/null | wc -l | tr -d ' ')\n")
	b.WriteString("  PANES=$(tmux list-panes -t \"$SESSION\" 2>/dev/null | wc -l | tr -d ' ')\n")
	b.WriteString("  if [ \"${WINDOWS:-0}\" = \"1\" ] && [ \"${PANES:-0}\" = \"1\" ]; then\n")
	b.WriteString("    tmux kill-session -t \"$SESSION\"\n")
	b.WriteString("  else\n")
	b.WriteString("    if [ -n \"${TMUX:-}\" ]; then\n")
	b.WriteString("      tmux switch-client -t \"$SESSION\"\n")
	b.WriteString("    else\n")
	b.WriteString("      tmux attach-session -t \"$SESSION\"\n")
	b.WriteString("    fi\n")
	b.WriteString("    exit 0\n")
	b.WriteString("  fi\n")
	b.WriteString("fi\n\n")

	for i, w := range s.Windows {
		if len(w.Panes) == 0 {
			return "", fmt.Errorf("window %q has no panes", w.Name)
		}
		firstPath := filepath.Clean(w.Panes[0].Path)
		if i == 0 {
			b.WriteString(fmt.Sprintf("tmux new-session -d -s %q -n %q -c %q\n", s.Name, w.Name, firstPath))
		} else {
			b.WriteString(fmt.Sprintf("tmux new-window -t %q -n %q -c %q\n", s.Name, w.Name, firstPath))
		}
		for paneIdx := 1; paneIdx < len(w.Panes); paneIdx++ {
			pane := w.Panes[paneIdx]
			b.WriteString(fmt.Sprintf("tmux split-window -t %q:%d -c %q\n", s.Name, w.Index, filepath.Clean(pane.Path)))
		}
		b.WriteString(fmt.Sprintf("tmux select-layout -t %q:%d %q\n", s.Name, w.Index, w.Layout))
		b.WriteString(fmt.Sprintf("tmux select-pane -t %q:%d.%d\n", s.Name, w.Index, w.ActivePane))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("tmux select-window -t %q:%d\n", s.Name, s.ActiveWindow))
	b.WriteString("if [ -n \"${TMUX:-}\" ]; then\n")
	b.WriteString("  tmux switch-client -t \"$SESSION\"\n")
	b.WriteString("else\n")
	b.WriteString("  tmux attach-session -t \"$SESSION\"\n")
	b.WriteString("fi\n")
	return b.String(), nil
}
