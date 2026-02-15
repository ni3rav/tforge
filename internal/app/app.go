package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"tforge/internal/cli"
	"tforge/internal/config"
	"tforge/internal/fsutil"
	"tforge/internal/generate"
	"tforge/internal/journal"
	"tforge/internal/snapshot"
	"tforge/internal/tmux"
)

func Main() {
	if err := Run(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		cli.Error(os.Stderr, "%v", err)
		os.Exit(1)
	}
}

func Run(ctx context.Context, args []string, in io.Reader, out io.Writer, _ io.Writer) error {
	if len(args) == 0 {
		return usageError(out, "missing command")
	}

	switch args[0] {
	case "--help", "-h", "help":
		printHelp(out)
		return nil
	case "capture":
		return runCapture(ctx, args[1:], in, out)
	case "restore":
		return runRestore(ctx, args[1:], in, out)
	default:
		return usageError(out, fmt.Sprintf("unknown command %q", args[0]))
	}
}

func runCapture(ctx context.Context, args []string, in io.Reader, out io.Writer) error {
	fs := flag.NewFlagSet("capture", flag.ContinueOnError)
	fs.SetOutput(out)

	sessionName := fs.String("session", "", "tmux session name to capture")
	saveName := fs.String("name", "", "name to save generated script as")
	bindKey := fs.String("key", "", "tmux key to bind (prefix + key), empty to skip")
	noBind := fs.Bool("no-bind", false, "do not modify ~/.tmux.conf")

	if err := fs.Parse(args); err != nil {
		return err
	}

	runner := tmux.NewCommandRunner()
	service := tmux.NewService(runner)
	prompt := cli.NewPrompter(in, out)

	if *sessionName == "" {
		s, ok, err := selectTmuxSession(ctx, service, prompt, out)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("capture cancelled")
		}
		*sessionName = s
	}

	if exists, err := service.SessionExists(ctx, *sessionName); err != nil {
		return err
	} else if !exists {
		return fmt.Errorf("tmux session %q does not exist", *sessionName)
	}

	if *saveName == "" {
		v, err := prompt.AskDefault("Save layout as", *sessionName)
		if err != nil {
			return err
		}
		*saveName = v
	}

	capturer := snapshot.NewCapturer(service)
	snap, err := capturer.CaptureSession(ctx, *sessionName)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	scriptPath := filepath.Join(home, ".tforge", "sessions", *saveName+".sh")
	content, err := generate.Script(snap)
	if err != nil {
		return err
	}
	if err := fsutil.WriteExecutable(scriptPath, []byte(content)); err != nil {
		return err
	}
	cli.Info(out, "Wrote script: %s", scriptPath)

	if err := updateJournal(home, snap, scriptPath); err != nil {
		cli.Warn(out, "unable to update journal: %v", err)
	}

	if !*noBind {
		if *bindKey == "" {
			wantsBind, err := prompt.AskYesNo("Add tmux keybinding", false)
			if err != nil {
				return err
			}
			if wantsBind {
				v, err := prompt.Ask("Bind to key (prefix + key)")
				if err != nil {
					return err
				}
				*bindKey = v
			}
		}
		if strings.TrimSpace(*bindKey) != "" {
			if warning := config.CommonKeyWarning(*bindKey); warning != "" {
				cli.Warn(out, "%s", warning)
			}
			tmuxConf := filepath.Join(home, ".tmux.conf")
			updated, changed, err := config.UpdateFile(tmuxConf, *saveName, *bindKey, scriptPath)
			if err != nil {
				return err
			}
			if changed {
				cli.Info(out, "Updated tmux config: %s", tmuxConf)
			} else {
				cli.Info(out, "Tmux config already up-to-date: %s", tmuxConf)
			}
			if updated {
				if err := service.ReloadConfig(ctx, tmuxConf); err != nil {
					cli.Warn(out, "unable to reload tmux config automatically: %v", err)
				} else {
					cli.Info(out, "Reloaded tmux config.")
				}
			}
			cli.Info(out, "Bound key: prefix + %s", *bindKey)
		} else {
			cli.Info(out, "Keybinding skipped.")
		}
	}

	cli.Info(out, "Done.")
	return nil
}

func runRestore(ctx context.Context, args []string, in io.Reader, out io.Writer) error {
	fs := flag.NewFlagSet("restore", flag.ContinueOnError)
	fs.SetOutput(out)
	sessionName := fs.String("session", "", "session name from journal")
	if err := fs.Parse(args); err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	jPath := journal.Path(home)
	data, err := journal.Load(jPath)
	if err != nil {
		return err
	}
	if len(data.Entries) == 0 {
		return errors.New("no saved sessions found; run 'tforge capture' first")
	}

	prompt := cli.NewPrompter(in, out)
	if *sessionName == "" {
		opts := make([]cli.Option, 0, len(data.Entries))
		for _, e := range data.Entries {
			opts = append(opts, cli.Option{
				ID:      e.Session,
				Label:   e.Session,
				Details: fmt.Sprintf("windows=%d panes=%d captured=%s", e.Windows, e.Panes, e.CapturedAt.Format(time.RFC3339)),
			})
		}
		sel, ok, err := cli.SelectFuzzy(prompt, out, "Select a saved session to restore", opts)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("restore cancelled")
		}
		*sessionName = sel
	}

	var entry *journal.Entry
	for i := range data.Entries {
		if data.Entries[i].Session == *sessionName {
			entry = &data.Entries[i]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("session %q is not in journal %s", *sessionName, jPath)
	}
	cli.Info(out, "Restoring %s (windows=%d, panes=%d)", entry.Session, entry.Windows, entry.Panes)

	cmd := exec.CommandContext(ctx, "/usr/bin/env", "bash", entry.ScriptPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore script failed: %w", err)
	}
	return nil
}

func updateJournal(home string, snap snapshot.Session, scriptPath string) error {
	path := journal.Path(home)
	data, err := journal.Load(path)
	if err != nil {
		return err
	}
	panes := 0
	for _, w := range snap.Windows {
		panes += len(w.Panes)
	}
	data = journal.Upsert(data, journal.Entry{
		Session:    snap.Name,
		ScriptPath: scriptPath,
		Windows:    len(snap.Windows),
		Panes:      panes,
		CapturedAt: time.Now().UTC(),
	})
	return journal.Save(path, data)
}

func selectTmuxSession(ctx context.Context, service *tmux.Service, prompt *cli.Prompter, out io.Writer) (string, bool, error) {
	detected, err := service.DetectCurrentSession(ctx)
	if err == nil && detected != "" {
		cli.Info(out, "Current tmux session detected: %s", detected)
		return detected, true, nil
	}
	sessions, err := service.ListSessions(ctx)
	if err != nil {
		return "", false, err
	}
	options := make([]cli.Option, 0, len(sessions))
	for _, s := range sessions {
		options = append(options, cli.Option{ID: s, Label: s})
	}
	return cli.SelectFuzzy(prompt, out, "Select tmux session to capture", options)
}

func printHelp(out io.Writer) {
	cmd := filepath.Base(os.Args[0])
	if cmd == "" {
		cmd = "tforge"
	}
	fmt.Fprintf(out, `%s - capture tmux layouts into reusable scripts

Usage:
  %s capture [flags]
  %s restore [flags]

Commands:
  capture     Capture a tmux session and generate a reusable script
  restore     Restore a captured session from ~/.tforge/journal.json

Flags (capture):
  --session <name>   tmux session name to capture
  --name <name>      output script name (default: same as session)
  --key <key>        bind key (prefix + key), empty to skip
  --no-bind          skip updating ~/.tmux.conf

Flags (restore):
  --session <name>   restore a specific saved session (else fuzzy select)

Examples:
  tf capture
  tforge capture --session hive --name hive --key g
  tforge restore
`, cmd, cmd, cmd)
}

func usageError(out io.Writer, msg string) error {
	printHelp(out)
	return errors.New(msg)
}
