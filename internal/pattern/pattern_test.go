package pattern

import (
	"reflect"
	"testing"

	"github.com/aryanwalia2003/falafal/internal/model"
)

func TestDetectGroupsSharedTemplate(t *testing.T) {
	files := []*model.Node{
		{Name: "Aryan_BE_2607.pdf", Path: "/d/Aryan_BE_2607.pdf", Ext: ".pdf"},
		{Name: "Aryan_FE_2607.pdf", Path: "/d/Aryan_FE_2607.pdf", Ext: ".pdf"},
		{Name: "Aryan_BE_2609.pdf", Path: "/d/Aryan_BE_2609.pdf", Ext: ".pdf"},
		{Name: "notes.txt", Path: "/d/notes.txt", Ext: ".txt"},
	}

	groups := Detect(files, Options{})
	if len(groups) != 1 {
		t.Fatalf("groups = %d, want 1: %+v", len(groups), groups)
	}
	g := groups[0]
	if g.Template != "Aryan_{1}_{2}.pdf" {
		t.Errorf("template = %q, want %q", g.Template, "Aryan_{1}_{2}.pdf")
	}
	if len(g.Files) != 3 {
		t.Fatalf("files = %d, want 3", len(g.Files))
	}
	wantFields := [][]string{{"BE", "2607"}, {"BE", "2609"}, {"FE", "2607"}}
	if !reflect.DeepEqual(g.Fields, wantFields) {
		t.Errorf("fields = %v, want %v", g.Fields, wantFields)
	}
}

func TestDetectSkipsBelowMinGroupSize(t *testing.T) {
	files := []*model.Node{
		{Name: "Aryan_BE_2607.pdf", Path: "/d/Aryan_BE_2607.pdf", Ext: ".pdf"},
		{Name: "Rahul_XY_1111.docx", Path: "/d/Rahul_XY_1111.docx", Ext: ".docx"},
	}
	if groups := Detect(files, Options{}); len(groups) != 0 {
		t.Errorf("groups = %+v, want none (each shape has only 1 file)", groups)
	}
}

func TestDetectSkipsIdenticalNamesWithNoVariation(t *testing.T) {
	files := []*model.Node{
		{Name: "report.pdf", Path: "/a/report.pdf", Ext: ".pdf"},
		{Name: "report.pdf", Path: "/b/report.pdf", Ext: ".pdf"},
	}
	if groups := Detect(files, Options{}); len(groups) != 0 {
		t.Errorf("groups = %+v, want none (no varying segment)", groups)
	}
}

func TestDetectDifferentShapesDoNotMix(t *testing.T) {
	files := []*model.Node{
		{Name: "Aryan_BE_2607.pdf", Path: "/d/Aryan_BE_2607.pdf", Ext: ".pdf"},
		{Name: "Aryan_FE_2607.pdf", Path: "/d/Aryan_FE_2607.pdf", Ext: ".pdf"},
		{Name: "Aryan-BE-2609.pdf", Path: "/d/Aryan-BE-2609.pdf", Ext: ".pdf"},    // different separator
		{Name: "Aryan_BE_26_09.pdf", Path: "/d/Aryan_BE_26_09.pdf", Ext: ".pdf"},  // different segment count
		{Name: "Aryan_BE_2610.docx", Path: "/d/Aryan_BE_2610.docx", Ext: ".docx"}, // different extension
	}
	groups := Detect(files, Options{})
	if len(groups) != 1 {
		t.Fatalf("groups = %d, want 1: %+v", len(groups), groups)
	}
	if len(groups[0].Files) != 2 {
		t.Errorf("files in group = %d, want 2 (only the matching underscore-shape pair)", len(groups[0].Files))
	}
}

func TestDetectRefinesMixedNamesIntoSpecificTemplates(t *testing.T) {
	files := []*model.Node{
		{Name: "Aryan_BE_2607.pdf", Path: "/d/Aryan_BE_2607.pdf", Ext: ".pdf"},
		{Name: "Aryan_FE_2607.pdf", Path: "/d/Aryan_FE_2607.pdf", Ext: ".pdf"},
		{Name: "Aryan_BE_2609.pdf", Path: "/d/Aryan_BE_2609.pdf", Ext: ".pdf"},
		{Name: "Rahul_CS_1001.pdf", Path: "/d/Rahul_CS_1001.pdf", Ext: ".pdf"},
		{Name: "Rahul_CS_1002.pdf", Path: "/d/Rahul_CS_1002.pdf", Ext: ".pdf"},
	}

	groups := Detect(files, Options{})
	if len(groups) != 2 {
		t.Fatalf("groups = %d, want 2: %+v", len(groups), groups)
	}

	byTemplate := map[string]Group{}
	for _, g := range groups {
		byTemplate[g.Template] = g
	}

	aryan, ok := byTemplate["Aryan_{1}_{2}.pdf"]
	if !ok {
		t.Fatalf("missing Aryan_{{1}}_{{2}}.pdf template, got %+v", groups)
	}
	if len(aryan.Files) != 3 {
		t.Errorf("Aryan group files = %d, want 3", len(aryan.Files))
	}

	rahul, ok := byTemplate["Rahul_CS_{1}.pdf"]
	if !ok {
		t.Fatalf("missing Rahul_CS_{{1}}.pdf template (name split should also fix the constant CS segment), got %+v", groups)
	}
	if len(rahul.Files) != 2 {
		t.Errorf("Rahul group files = %d, want 2", len(rahul.Files))
	}
}

func TestDetectMinGroupSizeOption(t *testing.T) {
	files := []*model.Node{
		{Name: "Aryan_BE_2607.pdf", Path: "/d/Aryan_BE_2607.pdf", Ext: ".pdf"},
		{Name: "Aryan_FE_2607.pdf", Path: "/d/Aryan_FE_2607.pdf", Ext: ".pdf"},
		{Name: "Aryan_BE_2609.pdf", Path: "/d/Aryan_BE_2609.pdf", Ext: ".pdf"},
	}
	if groups := Detect(files, Options{MinGroupSize: 4}); len(groups) != 0 {
		t.Errorf("groups = %+v, want none with MinGroupSize 4", groups)
	}
}
