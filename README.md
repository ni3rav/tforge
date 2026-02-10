# tforge

`tforge` is a Go CLI that captures a live tmux session layout and generates a reusable shell script that can recreate it on demand. It can also add a tmux key binding so your layout can be restored with `prefix + <key>`.

## Features

- Capture a tmux session by name.
- Record windows, pane paths, pane counts, layout strings, and active window/pane.
- Generate idempotent scripts in `~/.tforge/sessions/<name>.sh`.
- Safely update `~/.tmux.conf` with an `unbind-key` + `bind-key` block.
- Reload tmux config automatically when possible.
- Supports interactive prompts and CLI flags.
- Uses a single `tforge` binary; `tf` can be a symlink to that same binary.

## Install

```bash
go build -o tforge ./cmd/tforge
```

Place the binaries in your `PATH`, for example:

```bash
install -m 755 tforge /usr/local/bin/tforge
ln -sf /usr/local/bin/tforge /usr/local/bin/tf
```

## Usage

`tforge` and `tf` are interoperable via one binary: `tf` should point to `tforge` (symlink).


```bash
tf capture
```

Flag-based usage:

```bash
tforge capture --session hive --name hive --key g
```

Skip tmux config update:

```bash
tforge capture --session hive --no-bind
```

## How keybinding works

When keybinding is enabled, `tforge` updates `~/.tmux.conf` by writing a managed block:

```tmux
# tforge begin: hive
unbind-key g
bind-key g run-shell "/usr/bin/env bash /home/you/.tforge/sessions/hive.sh"
# tforge end: hive
```

This ensures the selected key is unbound first and avoids duplicated managed blocks for the same saved session.

## Generated script behavior

Generated scripts are idempotent:

- If the session already exists:
  - inside tmux: `switch-client`
  - outside tmux: `attach-session`
- If the session does not exist:
  - create session + windows + panes
  - restore each window layout
  - restore active pane/window
  - attach/switch appropriately

Scripts are plain shell and can be edited manually after generation.

## Development checks

Run full local verification:

```bash
make check
```

Or run manually:

```bash
go test -count=1 ./...
go build ./...
go vet ./...
```

## Help

```bash
tforge --help
```

Shows command and flag documentation.
