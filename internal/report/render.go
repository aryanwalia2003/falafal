package report

import (
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
)

// RenderText renders the full report (tree + stats) as plain text.
func RenderText(tree *model.Tree, color bool) string {
	var sb strings.Builder
	RenderTree(&sb, tree.Root, color)
	RenderStats(&sb, tree.Stats)
	return sb.String()
}
