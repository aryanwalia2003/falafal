package report

import (
	"strings"
	"testing"

	"github.com/aryanwalia2003/falafal/internal/model"
)

func sampleTree() *model.Tree {
	fileA := &model.Node{Name: "a.txt", Path: "/root/a.txt", Size: 5, TypeLabel: "Text document", DupGroup: "D1"}
	fileB := &model.Node{Name: "b.txt", Path: "/root/sub/b.txt", Size: 5, TypeLabel: "Text document", DupGroup: "D1"}
	sub := &model.Node{Name: "sub", IsDir: true, Size: 5, Children: []*model.Node{fileB}}
	root := &model.Node{Name: "root", IsDir: true, Size: 10, Children: []*model.Node{fileA, sub}}
	return &model.Tree{
		Root: root,
		Stats: model.Stats{
			TotalFiles: 2,
			TotalDirs:  1,
			TotalSize:  10,
			DupGroups:  []model.DupGroupInfo{{Hash: "x", Nodes: []*model.Node{fileA, fileB}}},
			WastedSize: 5,
		},
	}
}

func TestRenderText(t *testing.T) {
	out := RenderText(sampleTree(), false)
	for _, want := range []string{"root/", "a.txt", "[DUP:D1]", "Duplicates: 1 groups"} {
		if !strings.Contains(out, want) {
			t.Errorf("RenderText output missing %q\n---\n%s", want, out)
		}
	}
}

func TestRenderJSON(t *testing.T) {
	b, err := RenderJSON(sampleTree())
	if err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	if !strings.Contains(string(b), `"Name": "a.txt"`) {
		t.Errorf("RenderJSON missing expected field:\n%s", b)
	}
}

func TestRenderHTML(t *testing.T) {
	out := RenderHTML(sampleTree())
	for _, want := range []string{"<html", "a.txt", "DUP:D1", "Summary"} {
		if !strings.Contains(out, want) {
			t.Errorf("RenderHTML output missing %q", want)
		}
	}
}

func TestHumanSize(t *testing.T) {
	cases := map[int64]string{
		500:             "500 B",
		2048:            "2.0 KiB",
		5 * 1024 * 1024: "5.0 MiB",
	}
	for size, want := range cases {
		if got := HumanSize(size); got != want {
			t.Errorf("HumanSize(%d) = %q, want %q", size, got, want)
		}
	}
}
