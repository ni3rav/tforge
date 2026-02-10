package config

import (
	"strings"
	"testing"
)

func TestUpdateContentAddsSingleBlock(t *testing.T) {
	initial := "set -g mouse on\n"
	out := UpdateContent(initial, "hive", "g", "/home/me/.tmux/sessions/hive.sh")
	if !strings.Contains(out, "# tforge begin: hive") {
		t.Fatal("missing block begin")
	}
	if strings.Count(out, "\nbind-key g ") != 1 {
		t.Fatalf("expected exactly one bind line, got %d", strings.Count(out, "\nbind-key g "))
	}

	again := UpdateContent(out, "hive", "g", "/home/me/.tmux/sessions/hive.sh")
	if strings.Count(again, "# tforge begin: hive") != 1 {
		t.Fatal("expected idempotent block update")
	}
}
