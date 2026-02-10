package snapshot

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type TmuxReader interface {
	ListWindows(ctx context.Context, session string) ([]string, error)
	ListPanes(ctx context.Context, target string) ([]string, error)
}

type Session struct {
	Name          string
	Windows       []Window
	ActiveWindow  int
	ActivePaneIDs map[int]int
}

type Window struct {
	Index      int
	Name       string
	Layout     string
	Panes      []Pane
	ActivePane int
}

type Pane struct {
	Index int
	ID    string
	Path  string
}

type Capturer struct {
	tmux TmuxReader
}

func NewCapturer(tmux TmuxReader) *Capturer {
	return &Capturer{tmux: tmux}
}

func (c *Capturer) CaptureSession(ctx context.Context, session string) (Session, error) {
	windowLines, err := c.tmux.ListWindows(ctx, session)
	if err != nil {
		return Session{}, err
	}
	if len(windowLines) == 0 {
		return Session{}, fmt.Errorf("session %q has no windows", session)
	}

	snap := Session{Name: session, ActivePaneIDs: map[int]int{}}
	for _, line := range windowLines {
		win, activeWindow, err := parseWindow(line)
		if err != nil {
			return Session{}, err
		}
		if activeWindow {
			snap.ActiveWindow = win.Index
		}
		paneLines, err := c.tmux.ListPanes(ctx, fmt.Sprintf("%s:%d", session, win.Index))
		if err != nil {
			return Session{}, err
		}
		for _, pLine := range paneLines {
			pane, activePane, err := parsePane(pLine)
			if err != nil {
				return Session{}, err
			}
			win.Panes = append(win.Panes, pane)
			if activePane {
				win.ActivePane = pane.Index
				snap.ActivePaneIDs[win.Index] = pane.Index
			}
		}
		snap.Windows = append(snap.Windows, win)
	}
	return snap, nil
}

func parseWindow(line string) (Window, bool, error) {
	parts := strings.Split(line, "|")
	if len(parts) != 4 {
		return Window{}, false, fmt.Errorf("invalid tmux window row: %q", line)
	}
	index, err := strconv.Atoi(parts[0])
	if err != nil {
		return Window{}, false, err
	}
	active := parts[3] == "1"
	return Window{Index: index, Name: parts[1], Layout: parts[2]}, active, nil
}

func parsePane(line string) (Pane, bool, error) {
	parts := strings.Split(line, "|")
	if len(parts) != 4 {
		return Pane{}, false, fmt.Errorf("invalid tmux pane row: %q", line)
	}
	index, err := strconv.Atoi(parts[0])
	if err != nil {
		return Pane{}, false, err
	}
	active := parts[3] == "1"
	return Pane{Index: index, ID: parts[1], Path: parts[2]}, active, nil
}
