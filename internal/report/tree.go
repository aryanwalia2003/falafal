package report

import (
	"fmt"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
)

const (
	colorReset = "\033[0m"
	colorBlue  = "\033[34;1m"
	colorGray  = "\033[90m"
	colorYel   = "\033[33;1m"
)

// RenderTree writes a box-drawing tree for the node to sb. If color is true, ANSI
// codes highlight directories (blue) and duplicate files (yellow).
func RenderTree(sb *strings.Builder, node *model.Node, color bool) {
	sb.WriteString(node.Name)
	sb.WriteString("/\n")
	renderChildren(sb, node, "", color)
}

func renderChildren(sb *strings.Builder, node *model.Node, prefix string, color bool) {
	for i, child := range node.Children {
		last := i == len(node.Children)-1
		connector := "├── "
		nextPrefix := prefix + "│   "
		if last {
			connector = "└── "
			nextPrefix = prefix + "    "
		}

		sb.WriteString(prefix)
		sb.WriteString(connector)
		sb.WriteString(renderLine(child, color))
		sb.WriteString("\n")

		if child.IsDir {
			renderChildren(sb, child, nextPrefix, color)
		}
	}
}

func renderLine(n *model.Node, color bool) string {
	if n.IsDir {
		name := n.Name + "/"
		if color {
			name = colorBlue + name + colorReset
		}
		return fmt.Sprintf("%s (%s)", name, HumanSize(n.Size))
	}

	dup := ""
	if n.DupGroup != "" {
		tag := fmt.Sprintf(" [DUP:%s]", n.DupGroup)
		if color {
			tag = colorYel + tag + colorReset
		}
		dup = tag
	}
	meta := fmt.Sprintf(" (%s, %s)", n.TypeLabel, HumanSize(n.Size))
	if color {
		meta = colorGray + meta + colorReset
	}
	return n.Name + meta + dup
}
