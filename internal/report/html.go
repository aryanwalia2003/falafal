package report

import (
	"fmt"
	"html"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
)

const htmlHead = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>falafal report</title>
<style>
  body { font-family: -apple-system, Segoe UI, Arial, sans-serif; background: #1e1e1e; color: #ddd; padding: 1.5rem; }
  h1, h2 { color: #fff; }
  .stats { background: #2a2a2a; border-radius: 8px; padding: 1rem 1.5rem; margin-bottom: 1.5rem; }
  .stats table { border-collapse: collapse; }
  .stats td { padding: 2px 12px 2px 0; }
  details { margin-left: 1.2rem; }
  summary { cursor: pointer; color: #7cb7ff; }
  summary::marker { color: #7cb7ff; }
  .file { margin-left: 1.2rem; padding: 1px 0; }
  .meta { color: #999; font-size: 0.85em; }
  .dup { color: #f0c419; font-weight: bold; }
  a.dup-link { color: #f0c419; text-decoration: none; }
  a.dup-link:hover { text-decoration: underline; }
</style>
</head>
<body>
<h1>falafal report</h1>
`

const htmlFoot = `</body>
</html>
`

// RenderHTML builds a single self-contained HTML report with a collapsible tree
// and a stats panel.
func RenderHTML(tree *model.Tree) string {
	var sb strings.Builder
	sb.WriteString(htmlHead)

	writeStatsHTML(&sb, tree.Stats)

	sb.WriteString("<h2>Tree</h2>\n")
	sb.WriteString("<details open><summary>")
	sb.WriteString(html.EscapeString(tree.Root.Name))
	sb.WriteString("/</summary>\n")
	for _, child := range tree.Root.Children {
		writeNodeHTML(&sb, child)
	}
	sb.WriteString("</details>\n")

	sb.WriteString(htmlFoot)
	return sb.String()
}

func writeNodeHTML(sb *strings.Builder, n *model.Node) {
	if n.IsDir {
		fmt.Fprintf(sb, "<details><summary>%s/ <span class=\"meta\">(%s)</span></summary>\n",
			html.EscapeString(n.Name), HumanSize(n.Size))
		for _, child := range n.Children {
			writeNodeHTML(sb, child)
		}
		sb.WriteString("</details>\n")
		return
	}

	dupBadge := ""
	if n.DupGroup != "" {
		dupBadge = fmt.Sprintf(` <a class="dup-link" href="#%s">[DUP:%s]</a>`, n.DupGroup, n.DupGroup)
	}
	fmt.Fprintf(sb, `<div class="file" id="node-%s">%s <span class="meta">(%s, %s)</span>%s</div>`+"\n",
		html.EscapeString(n.Path), html.EscapeString(n.Name), html.EscapeString(n.TypeLabel), HumanSize(n.Size), dupBadge)
}

func writeStatsHTML(sb *strings.Builder, stats model.Stats) {
	sb.WriteString(`<div class="stats">` + "\n<h2>Summary</h2>\n<table>\n")
	fmt.Fprintf(sb, "<tr><td>Files</td><td>%d</td></tr>\n", stats.TotalFiles)
	fmt.Fprintf(sb, "<tr><td>Dirs</td><td>%d</td></tr>\n", stats.TotalDirs)
	fmt.Fprintf(sb, "<tr><td>Total size</td><td>%s</td></tr>\n", HumanSize(stats.TotalSize))
	fmt.Fprintf(sb, "<tr><td>Duplicate groups</td><td>%d</td></tr>\n", len(stats.DupGroups))
	fmt.Fprintf(sb, "<tr><td>Wasted space</td><td>%s</td></tr>\n", HumanSize(stats.WastedSize))
	sb.WriteString("</table>\n")

	if len(stats.ByExt) > 0 {
		sb.WriteString("<h2>By type</h2>\n<table>\n")
		for _, e := range stats.ByExt {
			fmt.Fprintf(sb, "<tr><td>%s</td><td>%d files</td><td>%s</td></tr>\n",
				html.EscapeString(e.Ext), e.Count, HumanSize(e.Size))
		}
		sb.WriteString("</table>\n")
	}

	if len(stats.LargestFiles) > 0 {
		sb.WriteString("<h2>Largest files</h2>\n<table>\n")
		for _, f := range stats.LargestFiles {
			fmt.Fprintf(sb, "<tr><td>%s</td><td>%s</td></tr>\n", HumanSize(f.Size), html.EscapeString(f.Path))
		}
		sb.WriteString("</table>\n")
	}

	if len(stats.DupGroups) > 0 {
		sb.WriteString("<h2>Duplicates</h2>\n")
		for _, g := range stats.DupGroups {
			groupID := g.Nodes[0].DupGroup
			fmt.Fprintf(sb, `<div id="%s" class="dup">[%s] %s (%d copies)</div>`+"\n<ul>\n",
				groupID, groupID, HumanSize(g.Nodes[0].Size), len(g.Nodes))
			for _, n := range g.Nodes {
				fmt.Fprintf(sb, `<li><a href="#node-%s">%s</a></li>`+"\n", html.EscapeString(n.Path), html.EscapeString(n.Path))
			}
			sb.WriteString("</ul>\n")
		}
	}
	sb.WriteString("</div>\n")
}
