package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestScanBasic(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.txt"), "hello")
	writeFile(t, filepath.Join(dir, "sub", "b.txt"), "hello") // duplicate of a.txt
	writeFile(t, filepath.Join(dir, "sub", "c.go"), "package main")
	writeFile(t, filepath.Join(dir, ".git", "config"), "ignored")
	writeFile(t, filepath.Join(dir, "node_modules", "pkg", "index.js"), "ignored")

	tree, err := Scan(dir, Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if tree.Stats.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, want 3 (noise dirs should be skipped)", tree.Stats.TotalFiles)
	}
	if len(tree.Stats.DupGroups) != 1 {
		t.Fatalf("DupGroups = %d, want 1", len(tree.Stats.DupGroups))
	}
	if len(tree.Stats.DupGroups[0].Nodes) != 2 {
		t.Errorf("dup group size = %d, want 2", len(tree.Stats.DupGroups[0].Nodes))
	}
	wantWasted := int64(len("hello"))
	if tree.Stats.WastedSize != wantWasted {
		t.Errorf("WastedSize = %d, want %d", tree.Stats.WastedSize, wantWasted)
	}
}

func TestScanIncludeAll(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".git", "config"), "ignored")

	tree, err := Scan(dir, Options{IncludeAll: true})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if tree.Stats.TotalFiles != 1 {
		t.Errorf("TotalFiles = %d, want 1 when --all is set", tree.Stats.TotalFiles)
	}
}

func TestTypeLabel(t *testing.T) {
	cases := []struct{ name, wantExt, wantLabel string }{
		{"main.go", ".go", "Go source"},
		{"README", "", "File (no extension)"},
		{".bashrc", "", "File (no extension)"},
		{"archive.tar.gz", ".gz", "Gzip archive"},
	}
	for _, c := range cases {
		ext, label := TypeLabel(c.name)
		if ext != c.wantExt || label != c.wantLabel {
			t.Errorf("TypeLabel(%q) = (%q, %q), want (%q, %q)", c.name, ext, label, c.wantExt, c.wantLabel)
		}
	}
}
