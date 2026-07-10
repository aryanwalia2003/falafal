package report

import (
	"encoding/json"

	"github.com/aryanwalia2003/falafal/internal/model"
)

// RenderJSON marshals the full tree (nodes + stats) as indented JSON.
func RenderJSON(tree *model.Tree) ([]byte, error) {
	return json.MarshalIndent(tree, "", "  ")
}
