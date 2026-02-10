package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"tforge/internal/cli"
	"tforge/internal/config"
	"tforge/internal/fsutil"
	"tforge/internal/generate"
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
	default:
		return usageError(out, fmt.Sprintf("unknown command %q", args[0]))
	}
}

func runCapture(ctx context.Context, args []string, in io.Reader, out io.Writer) error {
	fs := flag.NewFlagSet("capture", flag.ContinueOnError)
	fs.SetOutput(out)

	sessionName := fs.String("session", "", "tmux session name to capture")
	saveName := fs.String("name", "", "name to save generated script as")
	bindKey := fs.String("key", "", "tmux key to bind (prefix + key)")
	noBind := fs.Bool("no-bind", false, "do not modify ~/.tmux.conf")

	if err := fs.Parse(args); err != nil {
		return err
	}

	runner := tmux.NewCommandRunner()
	service := tmux.NewService(runner)
	prompt := cli.NewPrompter(in, out)

	if *sessionName == "" {
		detected, err := service.DetectCurrentSession(ctx)
		if err == nil && detected != "" {
			cli.Info(out, "Current tmux session detected: %s", detected)
			*sessionName = detected
		} else {
			available, listErr := service.ListSessions(ctx)
			if listErr == nil && len(available) > 0 {
				cli.Info(out, "Available tmux sessions: %s", strings.Join(available, ", "))
			}
			v, pErr := prompt.Ask("Session name to capture")
			if pErr != nil {
				return pErr
			}
			*sessionName = v
		}
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

	if !*noBind {
		if *bindKey == "" {
			v, err := prompt.Ask("Bind to key (prefix + key)")
			if err != nil {
				return err
			}
			*bindKey = v
		}
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
	}

	cli.Info(out, "Done.")
	return nil
}

func printHelp(out io.Writer) {
	fmt.Fprint(out, `tforge - capture tmux layouts into reusable scripts

Usage:
  tforge capture [flags]

Command:
  capture     Capture a tmux session and generate a reusable script

Flags:
  --session <name>   tmux session name to capture
  --name <name>      output script name (default: same as session)
  --key <key>        bind key (prefix + key)
  --no-bind          skip updating ~/.tmux.conf

Examples:
  tf capture
  tforge capture --session hive --name hive --key g
  tforge capture --session dev --no-bind
`)
}

func usageError(out io.Writer, msg string) error {
	printHelp(out)
	return errors.New(msg)
}
