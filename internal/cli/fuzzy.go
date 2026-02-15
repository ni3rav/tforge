package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Option struct {
	ID      string
	Label   string
	Details string
}

// SelectFuzzy tries an interactive arrow-key selector when a TTY is available.
// If TTY interaction is unavailable, it falls back to a numbered prompt selector.
func SelectFuzzy(p *Prompter, out io.Writer, title string, options []Option) (string, bool, error) {
	if len(options) == 0 {
		return "", false, nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return selectFallback(p, out, title, options)
	}
	defer tty.Close()

	state, err := sttyState(tty)
	if err != nil {
		return selectFallback(p, out, title, options)
	}
	defer restoreTTY(tty, state)

	if err := setRawTTY(tty); err != nil {
		return selectFallback(p, out, title, options)
	}

	all := append([]Option{{ID: "", Label: "Exit", Details: "Cancel"}}, options...)
	query := ""
	selected := 0

	for {
		filtered := filterOptions(all, query)
		if len(filtered) == 0 {
			filtered = []Option{{ID: "", Label: "Exit", Details: "No matches (press Enter to cancel)"}}
		}
		if selected >= len(filtered) {
			selected = len(filtered) - 1
		}
		if selected < 0 {
			selected = 0
		}

		renderSelect(tty, title, query, filtered, selected)

		buf := make([]byte, 3)
		n, rErr := tty.Read(buf)
		if rErr != nil {
			return "", false, rErr
		}
		b := buf[:n]

		switch {
		case len(b) == 1 && (b[0] == '\r' || b[0] == '\n'):
			pick := filtered[selected]
			fmt.Fprint(tty, "\n")
			if pick.ID == "" {
				return "", false, nil
			}
			return pick.ID, true, nil
		case len(b) == 1 && (b[0] == 3 || b[0] == 'q'):
			fmt.Fprint(tty, "\n")
			return "", false, nil
		case len(b) == 1 && (b[0] == 127 || b[0] == 8):
			if len(query) > 0 {
				query = query[:len(query)-1]
				selected = 0
			}
		case len(b) == 3 && b[0] == 27 && b[1] == 91 && b[2] == 65: // up
			selected--
		case len(b) == 3 && b[0] == 27 && b[1] == 91 && b[2] == 66: // down
			selected++
		case len(b) == 1 && b[0] >= 32 && b[0] <= 126:
			query += string(b[0])
			selected = 0
		}
	}
}

func selectFallback(p *Prompter, out io.Writer, title string, options []Option) (string, bool, error) {
	fmt.Fprintf(out, "\n%s\n", title)
	fmt.Fprintln(out, "(fallback mode: type a number, or q to cancel)")
	for i, opt := range options {
		if strings.TrimSpace(opt.Details) == "" {
			fmt.Fprintf(out, "  %d) %s\n", i+1, opt.Label)
		} else {
			fmt.Fprintf(out, "  %d) %s (%s)\n", i+1, opt.Label, opt.Details)
		}
	}
	for {
		in, err := p.AskAllowEmpty("Select number (q to cancel)")
		if err != nil {
			return "", false, err
		}
		in = strings.TrimSpace(strings.ToLower(in))
		if in == "q" || in == "exit" {
			return "", false, nil
		}
		idx, err := strconv.Atoi(in)
		if err != nil || idx < 1 || idx > len(options) {
			fmt.Fprintln(out, "Invalid selection.")
			continue
		}
		return options[idx-1].ID, true, nil
	}
}

func renderSelect(tty *os.File, title, query string, options []Option, selected int) {
	fmt.Fprint(tty, "\033[H\033[2J")
	fmt.Fprintf(tty, "%s\n", title)
	fmt.Fprint(tty, "Type to filter • ↑/↓ to move • Enter to select • q to cancel\n")
	fmt.Fprintf(tty, "Filter: %s\n\n", query)

	max := len(options)
	if max > 12 {
		max = 12
	}
	for i := 0; i < max; i++ {
		prefix := "  "
		if i == selected {
			prefix = "▸ "
		}
		opt := options[i]
		if strings.TrimSpace(opt.Details) == "" {
			fmt.Fprintf(tty, "%s%s\n", prefix, opt.Label)
		} else {
			fmt.Fprintf(tty, "%s%s (%s)\n", prefix, opt.Label, opt.Details)
		}
	}
}

func filterOptions(options []Option, query string) []Option {
	if strings.TrimSpace(query) == "" {
		return options
	}
	q := strings.ToLower(query)
	out := make([]Option, 0, len(options))
	for _, opt := range options {
		text := strings.ToLower(opt.Label + " " + opt.Details)
		if strings.Contains(text, q) {
			out = append(out, opt)
		}
	}
	return out
}

func setRawTTY(tty *os.File) error {
	cmd := exec.Command("stty", "-icanon", "-echo", "min", "1", "time", "0")
	cmd.Stdin = tty
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func sttyState(tty *os.File) (string, error) {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = tty
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func restoreTTY(tty *os.File, state string) {
	if state == "" {
		return
	}
	cmd := exec.Command("stty", state)
	cmd.Stdin = tty
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	_ = cmd.Run()
}
