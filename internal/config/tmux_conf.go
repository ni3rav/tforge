package config

import (
	"fmt"
	"os"
	"strings"
)

func CommonKeyWarning(key string) string {
	common := map[string]bool{"c": true, "n": true, "p": true, "l": true, "z": true, "%": true, "\"": true}
	if common[key] {
		return fmt.Sprintf("key %q is commonly used by tmux defaults; rebinding may override expected behavior", key)
	}
	return ""
}

func UpdateFile(path, sessionName, key, scriptPath string) (updated bool, changed bool, err error) {
	content, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return false, false, err
	}
	newContent := UpdateContent(string(content), sessionName, key, scriptPath)
	if newContent == string(content) {
		return false, false, nil
	}
	if err := os.WriteFile(path, []byte(newContent), 0o644); err != nil {
		return false, false, err
	}
	return true, true, nil
}

func UpdateContent(content, sessionName, key, scriptPath string) string {
	begin := fmt.Sprintf("# tforge begin: %s", sessionName)
	end := fmt.Sprintf("# tforge end: %s", sessionName)

	lines := strings.Split(content, "\n")
	var kept []string
	inBlock := false
	bindLine := fmt.Sprintf("bind-key %s run-shell \"/usr/bin/env bash %s\"", key, scriptPath)

	for _, line := range lines {
		trim := strings.TrimSpace(line)
		switch {
		case trim == begin:
			inBlock = true
			continue
		case trim == end:
			inBlock = false
			continue
		case inBlock:
			continue
		case strings.Contains(trim, fmt.Sprintf("run-shell \"/usr/bin/env bash %s\"", scriptPath)):
			continue
		default:
			kept = append(kept, line)
		}
	}

	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}

	block := []string{
		begin,
		fmt.Sprintf("unbind-key %s", key),
		bindLine,
		end,
	}
	if len(kept) > 0 {
		kept = append(kept, "")
	}
	kept = append(kept, block...)
	return strings.Join(kept, "\n") + "\n"
}
