package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type Option struct {
	ID      string
	Label   string
	Details string
}

func SelectFuzzy(_ *Prompter, _ io.Writer, title string, options []Option) (string, bool, error) {
	if len(options) == 0 {
		return "", false, nil
	}

	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", false, err
	}
	defer tty.Close()

	state, err := sttyState()
	if err != nil {
		return "", false, err
	}
	defer restoreTTY(state)
	if err := exec.Command("stty", "-icanon", "-echo", "min", "1", "time", "0").Run(); err != nil {
		return "", false, err
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

func sttyState() (string, error) {
	cmd := exec.Command("stty", "-g")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func restoreTTY(state string) {
	if state == "" {
		return
	}
	_ = exec.Command("stty", state).Run()
}
