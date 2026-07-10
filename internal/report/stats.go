package report

import (
	"fmt"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
)

// RenderStats writes a plain-text summary of scan stats to sb.
func RenderStats(sb *strings.Builder, stats model.Stats) {
	sb.WriteString("\nSummary\n")
	sb.WriteString("-------\n")
	fmt.Fprintf(sb, "Files: %d   Dirs: %d   Total size: %s\n", stats.TotalFiles, stats.TotalDirs, HumanSize(stats.TotalSize))

	if len(stats.ByExt) > 0 {
		sb.WriteString("\nBy type:\n")
		for _, e := range stats.ByExt {
			fmt.Fprintf(sb, "  %-16s %6d files   %s\n", e.Ext, e.Count, HumanSize(e.Size))
		}
	}

	if len(stats.LargestFiles) > 0 {
		sb.WriteString("\nLargest files:\n")
		for _, f := range stats.LargestFiles {
			fmt.Fprintf(sb, "  %10s  %s\n", HumanSize(f.Size), f.Path)
		}
	}

	if len(stats.DupGroups) > 0 {
		fmt.Fprintf(sb, "\nDuplicates: %d groups, %s wasted\n", len(stats.DupGroups), HumanSize(stats.WastedSize))
		for _, g := range stats.DupGroups {
			fmt.Fprintf(sb, "  [%s] %s (%d copies)\n", g.Nodes[0].DupGroup, HumanSize(g.Nodes[0].Size), len(g.Nodes))
			for _, n := range g.Nodes {
				fmt.Fprintf(sb, "      %s\n", n.Path)
			}
		}
	} else {
		sb.WriteString("\nNo duplicates found.\n")
	}
}
