package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type Prompter struct {
	in  *bufio.Reader
	out io.Writer
}

func NewPrompter(in io.Reader, out io.Writer) *Prompter {
	return &Prompter{in: bufio.NewReader(in), out: out}
}

func (p *Prompter) Ask(label string) (string, error) {
	for {
		fmt.Fprintf(p.out, "%s: ", label)
		line, err := p.in.ReadString('\n')
		if err != nil && len(line) == 0 {
			return "", err
		}
		line = strings.TrimSpace(line)
		if line != "" {
			return line, nil
		}
	}
}

func (p *Prompter) AskDefault(label, def string) (string, error) {
	fmt.Fprintf(p.out, "%s [%s]: ", label, def)
	line, err := p.in.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

func (p *Prompter) AskAllowEmpty(label string) (string, error) {
	fmt.Fprintf(p.out, "%s: ", label)
	line, err := p.in.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (p *Prompter) AskYesNo(label string, def bool) (bool, error) {
	defLabel := "y/N"
	if def {
		defLabel = "Y/n"
	}
	for {
		fmt.Fprintf(p.out, "%s [%s]: ", label, defLabel)
		line, err := p.in.ReadString('\n')
		if err != nil && len(line) == 0 {
			return false, err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			return def, nil
		}
		if line == "y" || line == "yes" {
			return true, nil
		}
		if line == "n" || line == "no" {
			return false, nil
		}
	}
}
