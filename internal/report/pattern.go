package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/pattern"
)

// RenderPatternsText renders detected naming-pattern groups as plain text:
// each group's template, followed by every matching file with its extracted
// variable fields.
func RenderPatternsText(groups []pattern.Group) string {
	if len(groups) == 0 {
		return "No naming patterns found.\n"
	}
	var sb strings.Builder
	for _, g := range groups {
		fmt.Fprintf(&sb, "%s  (%d files)\n", g.Template, len(g.Files))
		for i, f := range g.Files {
			fmt.Fprintf(&sb, "  [%s]  %s\n", strings.Join(g.Fields[i], ", "), f.Path)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

type patternJSONFile struct {
	Path   string   `json:"path"`
	Fields []string `json:"fields"`
}

type patternJSONGroup struct {
	Template string            `json:"template"`
	Count    int               `json:"count"`
	Files    []patternJSONFile `json:"files"`
}

// RenderPatternsJSON marshals detected naming-pattern groups as JSON.
func RenderPatternsJSON(groups []pattern.Group) ([]byte, error) {
	out := make([]patternJSONGroup, 0, len(groups))
	for _, g := range groups {
		jg := patternJSONGroup{Template: g.Template, Count: len(g.Files), Files: make([]patternJSONFile, 0, len(g.Files))}
		for i, f := range g.Files {
			jg.Files = append(jg.Files, patternJSONFile{Path: f.Path, Fields: g.Fields[i]})
		}
		out = append(out, jg)
	}
	return json.MarshalIndent(out, "", "  ")
}
