package tmux

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Runner interface {
	Run(ctx context.Context, args ...string) (string, error)
}

type commandRunner struct{}

func NewCommandRunner() Runner {
	return commandRunner{}
}

func (commandRunner) Run(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "tmux", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return strings.TrimRight(string(out), "\n"), nil
}

type Service struct {
	runner Runner
}

func NewService(r Runner) *Service {
	return &Service{runner: r}
}

func (s *Service) DetectCurrentSession(ctx context.Context) (string, error) {
	if os.Getenv("TMUX") == "" {
		return "", errors.New("not in tmux")
	}
	out, err := s.runner.Run(ctx, "display-message", "-p", "#S")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (s *Service) ListSessions(ctx context.Context) ([]string, error) {
	out, err := s.runner.Run(ctx, "list-sessions", "-F", "#{session_name}")
	if err != nil {
		return nil, err
	}
	lines := splitLines(out)
	return lines, nil
}

func (s *Service) SessionExists(ctx context.Context, session string) (bool, error) {
	sessions, err := s.ListSessions(ctx)
	if err != nil {
		return false, err
	}
	for _, x := range sessions {
		if x == session {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) ListWindows(ctx context.Context, session string) ([]string, error) {
	return splitCommand(s.runner.Run(ctx, "list-windows", "-t", session, "-F", "#{window_index}|#{window_name}|#{window_layout}|#{window_active}"))
}

func (s *Service) ListPanes(ctx context.Context, target string) ([]string, error) {
	return splitCommand(s.runner.Run(ctx, "list-panes", "-t", target, "-F", "#{pane_index}|#{pane_id}|#{pane_current_path}|#{pane_active}"))
}

func (s *Service) ReloadConfig(ctx context.Context, path string) error {
	_, err := s.runner.Run(ctx, "source-file", path)
	return err
}

func splitCommand(out string, err error) ([]string, error) {
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

func splitLines(out string) []string {
	if strings.TrimSpace(out) == "" {
		return nil
	}
	parts := strings.Split(strings.TrimSpace(out), "\n")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
