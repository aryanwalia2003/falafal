package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/grep"
)

// RenderGrepText renders content-search matches as "path:line:text" lines,
// one per match, like grep -n.
func RenderGrepText(matches []grep.Match) string {
	if len(matches) == 0 {
		return "No matches.\n"
	}
	var sb strings.Builder
	for _, m := range matches {
		fmt.Fprintf(&sb, "%s:%d:%s\n", m.Path, m.Line, m.Text)
	}
	fmt.Fprintf(&sb, "\n%d match(es)\n", len(matches))
	return sb.String()
}

// RenderGrepJSON marshals content-search matches as a JSON array.
func RenderGrepJSON(matches []grep.Match) ([]byte, error) {
	if matches == nil {
		matches = []grep.Match{}
	}
	return json.MarshalIndent(matches, "", "  ")
}
