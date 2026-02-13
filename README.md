# tforge

`tforge` is a Go CLI that captures a live tmux session and turns it into a reusable restore script.

## Features

- Capture tmux sessions with fuzzy selection (with cancel/exit option).
- Save scripts to `~/.tforge/sessions/<name>.sh`.
- Optional keybinding (you can skip binding during capture).
- Restore sessions via `tforge restore` using fuzzy selection + details.
- Keeps metadata in `~/.tforge/journal.json` (session name, script path, windows, panes, capture time).
- If restore script is run against a fresh 1-window/1-pane session with the same name, it overrides that session and rebuilds layout.

## Install

Build both commands directly (no symlink needed):

```bash
go build -o tforge ./cmd/tforge
go build -o tf ./cmd/tf
```

Install:

```bash
install -m 755 tforge /usr/local/bin/tforge
install -m 755 tf /usr/local/bin/tf
```

## Usage

Capture interactively:

```bash
tf capture
```

Capture with flags:

```bash
tforge capture --session hive --name hive --key g
```

Capture but skip keybinding:

```bash
tforge capture --session hive --no-bind
```

In the interactive wizard, you can also skip keybinding by answering `n` when asked `Add tmux keybinding [y/N]`.

Restore interactively from journal:

```bash
tforge restore
```

Restore by name:

```bash
tforge restore --session hive
```

## Keybinding behavior

When binding is enabled, `tforge` writes a managed block in `~/.tmux.conf`:

```tmux
# tforge begin: hive
unbind-key g
bind-key g run-shell "/usr/bin/env bash /home/you/.tforge/sessions/hive.sh"
# tforge end: hive
```

## Development checks

```bash
make check
```

Manual checks:

```bash
go test -count=1 ./...
go build ./...
go vet ./...
```

## Help

```bash
tforge --help
```
