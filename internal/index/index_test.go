package index

import (
	"reflect"
	"testing"

	"github.com/aryanwalia2003/falafal/internal/model"
)

func TestBuildAndSearch(t *testing.T) {
	root := &model.Node{
		Name:  "root",
		IsDir: true,
		Children: []*model.Node{
			{Name: "Graphs Ag on Bi.pdf", Path: "/root/Graphs Ag on Bi.pdf"},
			{Name: "Old graphs inference.docx", Path: "/root/Old graphs inference.docx"},
			{Name: "notes.txt", Path: "/root/notes.txt"},
			{
				Name: "sub", IsDir: true,
				Children: []*model.Node{
					{Name: "graphs_backup.pdf", Path: "/root/sub/graphs_backup.pdf"},
				},
			},
		},
	}

	idx := Build(root)
	if idx.Files != 4 {
		t.Fatalf("Files = %d, want 4", idx.Files)
	}

	got := idx.Search("graphs")
	want := []string{"/root/Graphs Ag on Bi.pdf", "/root/Old graphs inference.docx", "/root/sub/graphs_backup.pdf"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Search(graphs) = %v, want %v", got, want)
	}

	got = idx.Search("graphs pdf")
	want = []string{"/root/Graphs Ag on Bi.pdf", "/root/sub/graphs_backup.pdf"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Search(graphs pdf) = %v, want %v", got, want)
	}

	if got := idx.Search("nonexistent"); got != nil {
		t.Errorf("Search(nonexistent) = %v, want nil", got)
	}
}

func TestTokenize(t *testing.T) {
	got := Tokenize("Graphs Ag_on-Bi (v2).pdf")
	want := []string{"graphs", "ag", "on", "bi", "v2", "pdf"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Tokenize = %v, want %v", got, want)
	}
}
