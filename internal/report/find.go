package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
)

// RenderFindText renders file groups (as produced by model.GroupByExt) as a
// plain-text report: one section per extension, listing full paths.
func RenderFindText(groups []model.ExtGroup) string {
	var sb strings.Builder

	total, totalSize := 0, int64(0)
	for _, g := range groups {
		total += g.Count
		totalSize += g.Size
	}

	if total == 0 {
		return "No files matched.\n"
	}

	for _, g := range groups {
		fmt.Fprintf(&sb, "%s (%d files, %s)\n", g.Ext, g.Count, HumanSize(g.Size))
		for _, n := range g.Files {
			fmt.Fprintf(&sb, "  %s\n", n.Path)
		}
		sb.WriteString("\n")
	}
	fmt.Fprintf(&sb, "Total: %d files, %s\n", total, HumanSize(totalSize))
	return sb.String()
}

type findJSONFile struct {
	Path      string `json:"path"`
	Size      int64  `json:"size"`
	TypeLabel string `json:"type_label"`
}

type findJSONGroup struct {
	Ext   string         `json:"ext"`
	Count int            `json:"count"`
	Size  int64          `json:"size"`
	Files []findJSONFile `json:"files"`
}

// RenderFindJSON marshals file groups as indented JSON.
func RenderFindJSON(groups []model.ExtGroup) ([]byte, error) {
	out := make([]findJSONGroup, 0, len(groups))
	for _, g := range groups {
		jg := findJSONGroup{Ext: g.Ext, Count: g.Count, Size: g.Size, Files: make([]findJSONFile, 0, len(g.Files))}
		for _, n := range g.Files {
			jg.Files = append(jg.Files, findJSONFile{Path: n.Path, Size: n.Size, TypeLabel: n.TypeLabel})
		}
		out = append(out, jg)
	}
	return json.MarshalIndent(out, "", "  ")
}

// RenderPathsText renders a flat list of file paths, one per line.
func RenderPathsText(paths []string) string {
	if len(paths) == 0 {
		return "No matches.\n"
	}
	var sb strings.Builder
	for _, p := range paths {
		sb.WriteString(p)
		sb.WriteString("\n")
	}
	fmt.Fprintf(&sb, "\n%d match(es)\n", len(paths))
	return sb.String()
}

// RenderPathsJSON marshals a flat list of file paths as a JSON array.
func RenderPathsJSON(paths []string) ([]byte, error) {
	if paths == nil {
		paths = []string{}
	}
	return json.MarshalIndent(paths, "", "  ")
}
