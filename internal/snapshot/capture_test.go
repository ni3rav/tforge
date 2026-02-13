package snapshot

import (
	"context"
	"testing"
)

type fakeTmux struct{}

func (fakeTmux) ListWindows(ctx context.Context, session string) ([]string, error) {
	return []string{"0|editor|a,b,c|1", "1|logs|d,e,f|0"}, nil
}

func (fakeTmux) ListPanes(ctx context.Context, target string) ([]string, error) {
	if target == "hive:0" {
		return []string{"0|%1|/repo|1", "1|%2|/repo|0"}, nil
	}
	return []string{"0|%3|/tmp|1"}, nil
}

func TestCaptureSession(t *testing.T) {
	c := NewCapturer(fakeTmux{})
	s, err := c.CaptureSession(context.Background(), "hive")
	if err != nil {
		t.Fatal(err)
	}
	if s.ActiveWindow != 0 {
		t.Fatalf("expected active window 0, got %d", s.ActiveWindow)
	}
	if len(s.Windows) != 2 {
		t.Fatalf("expected 2 windows, got %d", len(s.Windows))
	}
	if s.Windows[0].ActivePane != 0 {
		t.Fatalf("expected active pane 0, got %d", s.Windows[0].ActivePane)
	}
}
