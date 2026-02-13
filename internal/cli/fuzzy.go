package cli

import (
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

type Option struct {
	ID      string
	Label   string
	Details string
}

func SelectFuzzy(p *Prompter, out io.Writer, title string, options []Option) (string, bool, error) {
	if len(options) == 0 {
		return "", false, nil
	}

	query := ""
	for {
		fmt.Fprintf(out, "\n%s\n", title)
		fmt.Fprintln(out, "Type a filter (or press Enter to keep current), a number to select, or 'q' to cancel.")
		if query != "" {
			fmt.Fprintf(out, "Current filter: %q\n", query)
		}

		filtered := filterOptions(options, query)
		if len(filtered) == 0 {
			fmt.Fprintln(out, "No matches. Try a different filter.")
		} else {
			for i, opt := range filtered {
				fmt.Fprintf(out, "  %d) %s\n", i+1, opt.Label)
				if strings.TrimSpace(opt.Details) != "" {
					fmt.Fprintf(out, "     %s\n", opt.Details)
				}
			}
		}

		input, err := p.AskAllowEmpty("Filter/number (q to exit)")
		if err != nil {
			return "", false, err
		}
		input = strings.TrimSpace(input)
		if strings.EqualFold(input, "q") || strings.EqualFold(input, "exit") {
			return "", false, nil
		}
		if input == "" {
			continue
		}
		if idx, err := strconv.Atoi(input); err == nil {
			if idx >= 1 && idx <= len(filtered) {
				return filtered[idx-1].ID, true, nil
			}
			fmt.Fprintln(out, "Invalid selection number.")
			continue
		}
		query = input
	}
}

func filterOptions(options []Option, query string) []Option {
	if query == "" {
		return options
	}
	q := strings.ToLower(query)
	type scored struct {
		opt   Option
		score int
	}
	matches := make([]scored, 0)
	for _, opt := range options {
		label := strings.ToLower(opt.Label + " " + opt.Details)
		score := fuzzyScore(label, q)
		if score >= 0 {
			matches = append(matches, scored{opt: opt, score: score})
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score == matches[j].score {
			return matches[i].opt.Label < matches[j].opt.Label
		}
		return matches[i].score > matches[j].score
	})
	out := make([]Option, 0, len(matches))
	for _, m := range matches {
		out = append(out, m.opt)
	}
	return out
}

func fuzzyScore(target, query string) int {
	if strings.Contains(target, query) {
		return 1000 - len(target)
	}
	ti := 0
	score := 0
	for _, qc := range query {
		found := false
		for ti < len(target) {
			if rune(target[ti]) == qc {
				score += 2
				found = true
				ti++
				break
			}
			ti++
		}
		if !found {
			return -1
		}
	}
	return score
}
