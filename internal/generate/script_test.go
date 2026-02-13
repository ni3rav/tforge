package generate

import (
	"strings"
	"testing"

	"tforge/internal/snapshot"
)

func TestScriptIncludesIdempotentAttachFlow(t *testing.T) {
	s := snapshot.Session{
		Name:         "hive",
		ActiveWindow: 0,
		Windows: []snapshot.Window{{
			Index:      0,
			Name:       "editor",
			Layout:     "abcd",
			ActivePane: 0,
			Panes:      []snapshot.Pane{{Index: 0, Path: "/workspace"}, {Index: 1, Path: "/workspace"}},
		}},
	}
	out, err := Script(s)
	if err != nil {
		t.Fatal(err)
	}
	checks := []string{"tmux has-session -t \"$SESSION\"", "tmux switch-client -t \"$SESSION\"", "tmux attach-session -t \"$SESSION\"", "tmux new-session -d -s \"hive\" -n \"editor\" -c \"/workspace\"", "tmux kill-session -t \"$SESSION\""}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Fatalf("expected generated script to contain %q", c)
		}
	}
}
