package journal

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Entry struct {
	Session    string    `json:"session"`
	ScriptPath string    `json:"script_path"`
	Windows    int       `json:"windows"`
	Panes      int       `json:"panes"`
	CapturedAt time.Time `json:"captured_at"`
}

type Data struct {
	Entries []Entry `json:"entries"`
}

func Path(home string) string {
	return filepath.Join(home, ".tforge", "journal.json")
}

func Load(path string) (Data, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Data{}, nil
		}
		return Data{}, err
	}
	var d Data
	if err := json.Unmarshal(b, &d); err != nil {
		return Data{}, err
	}
	return d, nil
}

func Save(path string, d Data) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func Upsert(d Data, e Entry) Data {
	found := false
	for i := range d.Entries {
		if d.Entries[i].Session == e.Session {
			d.Entries[i] = e
			found = true
			break
		}
	}
	if !found {
		d.Entries = append(d.Entries, e)
	}
	sort.Slice(d.Entries, func(i, j int) bool { return d.Entries[i].Session < d.Entries[j].Session })
	return d
}
