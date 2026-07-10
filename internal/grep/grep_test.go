package grep

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aryanwalia2003/falafal/internal/model"
)

func writeFile(t *testing.T, dir, name, content string) *model.Node {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return &model.Node{Name: name, Path: path, Size: info.Size()}
}

func TestSearch(t *testing.T) {
	dir := t.TempDir()
	files := []*model.Node{
		writeFile(t, dir, "a.go", "package main\n// TODO: fix this\nfunc main() {}\n"),
		writeFile(t, dir, "b.go", "package main\nfunc helper() {}\n"),
		writeFile(t, dir, "bin.dat", "abc\x00def\n"),
	}

	re, err := Compile(Options{Pattern: "TODO"})
	if err != nil {
		t.Fatal(err)
	}
	matches, skipped := Search(files, re, Options{})
	if len(skipped) != 0 {
		t.Errorf("skipped = %v, want none", skipped)
	}
	if len(matches) != 1 {
		t.Fatalf("matches = %v, want 1", matches)
	}
	if matches[0].Line != 2 || matches[0].Path != files[0].Path {
		t.Errorf("match = %+v, want line 2 in %s", matches[0], files[0].Path)
	}
}

func TestSearchIgnoreCaseAndFixed(t *testing.T) {
	dir := t.TempDir()
	files := []*model.Node{
		writeFile(t, dir, "a.txt", "Hello (World)\nhello world\n"),
	}

	re, err := Compile(Options{Pattern: "(world)", IgnoreCase: true, Fixed: true})
	if err != nil {
		t.Fatal(err)
	}
	matches, _ := Search(files, re, Options{})
	if len(matches) != 1 || matches[0].Line != 1 {
		t.Fatalf("matches = %+v, want 1 match on line 1", matches)
	}
}

func TestSearchSkipsLargeFiles(t *testing.T) {
	dir := t.TempDir()
	f := writeFile(t, dir, "big.txt", "needle\n")

	re, err := Compile(Options{Pattern: "needle"})
	if err != nil {
		t.Fatal(err)
	}
	matches, skipped := Search([]*model.Node{f}, re, Options{MaxBytes: 1})
	if len(matches) != 0 {
		t.Errorf("matches = %v, want none", matches)
	}
	if len(skipped) != 1 || skipped[0] != f.Path {
		t.Errorf("skipped = %v, want [%s]", skipped, f.Path)
	}
}
