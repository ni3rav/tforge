# tforge

`tforge` is a lightweight Go CLI that captures a live tmux session and turns it into a reusable restore script.

## Features

- Single binary build (`tforge`) that can be invoked as `tforge` or `tf`.
- Interactive arrow-key fuzzy selector for capture/restore (`↑/↓`, type to filter, Enter to select, `q` to cancel).
- Automatic fallback to a numbered selector when interactive TTY controls are unavailable.
- Save scripts to `~/.tforge/sessions/<name>.sh`.
- Optional keybinding during wizard (skip with `n`) or via `--no-bind`.
- Restore sessions via `tforge restore` with session details (`windows`, `panes`, capture timestamp).
- Journal metadata in `~/.tforge/journal.json`.
- Fresh-session override: if same-name session is only 1 window + 1 pane, restore script replaces it with saved layout.

## Install

Build once:

```bash
go build -o tforge ./cmd/tforge
```

Install the same binary for both command names:

```bash
install -m 755 tforge /usr/local/bin/tforge
ln -f /usr/local/bin/tforge /usr/local/bin/tf
```

> `tf` and `tforge` now point to the same built binary (single build artifact).

## Usage

Capture interactively:

```bash
tf capture
```

Capture with flags:

```bash
tforge capture --session hive --name hive --key g
```

Skip keybinding:

```bash
tforge capture --session hive --no-bind
```

Interactive wizard skip:

- answer `n` to `Add tmux keybinding [y/N]`

Restore interactively:

```bash
tforge restore
```

Restore by name:

```bash
tforge restore --session hive
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
