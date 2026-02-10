package fsutil

import (
	"os"
	"path/filepath"
)

func WriteExecutable(path string, content []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o755)
}
