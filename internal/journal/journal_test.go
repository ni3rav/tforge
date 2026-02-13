package journal

import (
	"path/filepath"
	"testing"
	"time"
)

func TestUpsert(t *testing.T) {
	d := Data{}
	d = Upsert(d, Entry{Session: "b", ScriptPath: "/tmp/b.sh", CapturedAt: time.Now()})
	d = Upsert(d, Entry{Session: "a", ScriptPath: "/tmp/a.sh", CapturedAt: time.Now()})
	d = Upsert(d, Entry{Session: "a", ScriptPath: "/tmp/new-a.sh", CapturedAt: time.Now()})
	if len(d.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(d.Entries))
	}
	if d.Entries[0].Session != "a" || d.Entries[0].ScriptPath != "/tmp/new-a.sh" {
		t.Fatalf("unexpected first entry: %+v", d.Entries[0])
	}
}

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "journal.json")
	in := Data{Entries: []Entry{{Session: "hive", ScriptPath: "/tmp/hive.sh", Windows: 2, Panes: 3, CapturedAt: time.Now().UTC()}}}
	if err := Save(path, in); err != nil {
		t.Fatal(err)
	}
	out, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Entries) != 1 || out.Entries[0].Session != "hive" {
		t.Fatalf("unexpected output: %+v", out)
	}
}
