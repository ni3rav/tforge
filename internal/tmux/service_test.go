package tmux

import (
	"context"
	"errors"
	"os"
	"testing"
)

type fakeRunner struct {
	fn func(args ...string) (string, error)
}

func (f fakeRunner) Run(_ context.Context, args ...string) (string, error) {
	return f.fn(args...)
}

func TestSessionExists(t *testing.T) {
	svc := NewService(fakeRunner{fn: func(args ...string) (string, error) {
		return "hive\nops", nil
	}})
	ok, err := svc.SessionExists(context.Background(), "hive")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected session to exist")
	}
}

func TestDetectCurrentSessionRequiresTmuxEnv(t *testing.T) {
	t.Setenv("TMUX", "")
	svc := NewService(fakeRunner{fn: func(args ...string) (string, error) {
		return "hive", nil
	}})
	if _, err := svc.DetectCurrentSession(context.Background()); err == nil {
		t.Fatal("expected error when TMUX is empty")
	}
}

func TestDetectCurrentSession(t *testing.T) {
	_ = os.Setenv("TMUX", "1")
	t.Cleanup(func() { _ = os.Unsetenv("TMUX") })
	called := false
	svc := NewService(fakeRunner{fn: func(args ...string) (string, error) {
		called = true
		if len(args) < 3 || args[0] != "display-message" {
			return "", errors.New("unexpected call")
		}
		return "hive", nil
	}})
	s, err := svc.DetectCurrentSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !called || s != "hive" {
		t.Fatalf("unexpected result: called=%v session=%q", called, s)
	}
}
