package model

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Node is a single file or directory in the scanned tree.
type Node struct {
	Name      string
	Path      string
	IsDir     bool
	Size      int64 // for dirs, sum of children
	ModTime   time.Time
	Ext       string
	TypeLabel string
	Hash      string // content hash, files only
	DupGroup  string // shared key for files with identical content; empty if unique
	Children  []*Node
}

// DupGroup describes a set of files that share identical content.
type DupGroupInfo struct {
	Hash  string
	Nodes []*Node
}

// ExtStat holds aggregate numbers for one file extension.
type ExtStat struct {
	Ext   string
	Count int
	Size  int64
}

// Stats summarizes a scanned tree.
type Stats struct {
	TotalFiles   int
	TotalDirs    int
	TotalSize    int64
	ByExt        []ExtStat // sorted by Size desc
	LargestFiles []*Node   // sorted by Size desc, capped at TopN
	DupGroups    []DupGroupInfo
	WastedSize   int64 // sum of all-but-one copy in each dup group
}

// Tree is the full result of a scan.
type Tree struct {
	Root  *Node
	Stats Stats
}

// ComputeStats groups files by their Hash field into duplicate groups (assigning
// each group's Nodes a shared DupGroup id), aggregates per-extension totals, and
// picks the topN largest files. Shared by every scan backend (local, Drive, ...)
// so dedup/stats logic lives in exactly one place.
func ComputeStats(root *Node, allFiles []*Node, totalDirs, topN int) Stats {
	hashToNodes := make(map[string][]*Node)
	extToStat := make(map[string]*ExtStat)

	for _, n := range allFiles {
		hashToNodes[n.Hash] = append(hashToNodes[n.Hash], n)

		key := n.Ext
		if key == "" {
			key = "(no extension)"
		}
		st, ok := extToStat[key]
		if !ok {
			st = &ExtStat{Ext: key}
			extToStat[key] = st
		}
		st.Count++
		st.Size += n.Size
	}

	var dupGroups []DupGroupInfo
	var wasted int64
	groupIdx := 0
	for hash, nodes := range hashToNodes {
		if len(nodes) < 2 {
			continue
		}
		groupIdx++
		groupID := fmt.Sprintf("D%d", groupIdx)
		for _, n := range nodes {
			n.DupGroup = groupID
		}
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].Path < nodes[j].Path })
		dupGroups = append(dupGroups, DupGroupInfo{Hash: hash, Nodes: nodes})
		wasted += nodes[0].Size * int64(len(nodes)-1)
	}
	sort.Slice(dupGroups, func(i, j int) bool {
		return dupGroups[i].Nodes[0].Path < dupGroups[j].Nodes[0].Path
	})

	var extStats []ExtStat
	for _, st := range extToStat {
		extStats = append(extStats, *st)
	}
	sort.Slice(extStats, func(i, j int) bool { return extStats[i].Size > extStats[j].Size })

	sorted := append([]*Node{}, allFiles...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Size > sorted[j].Size })
	if topN <= 0 {
		topN = 10
	}
	if topN > len(sorted) {
		topN = len(sorted)
	}

	return Stats{
		TotalFiles:   len(allFiles),
		TotalDirs:    totalDirs,
		TotalSize:    root.Size,
		ByExt:        extStats,
		LargestFiles: append([]*Node{}, sorted[:topN]...),
		DupGroups:    dupGroups,
		WastedSize:   wasted,
	}
}

// AllFiles returns every non-directory node in the tree rooted at root, in
// tree order. Used by commands that need the full file list rather than the
// capped/aggregated view in Stats (e.g. `find`, `index`, `search`).
func AllFiles(root *Node) []*Node {
	var out []*Node
	var walk func(n *Node)
	walk = func(n *Node) {
		for _, c := range n.Children {
			if c.IsDir {
				walk(c)
				continue
			}
			out = append(out, c)
		}
	}
	walk(root)
	return out
}

// Filter describes criteria for narrowing a file list down to files of
// interest. A zero-value Filter matches every file.
type Filter struct {
	Exts         []string // normalized (lowercase, leading dot); empty = any extension
	NameContains string   // case-insensitive substring match on file name; empty = any
	MinSize      int64    // bytes; 0 = no minimum
	MaxSize      int64    // bytes; 0 = no maximum
}

// ApplyFilter returns the subset of files matching f.
func ApplyFilter(files []*Node, f Filter) []*Node {
	var out []*Node
	for _, n := range files {
		if len(f.Exts) > 0 && !containsExt(f.Exts, n.Ext) {
			continue
		}
		if f.NameContains != "" && !strings.Contains(strings.ToLower(n.Name), strings.ToLower(f.NameContains)) {
			continue
		}
		if f.MinSize > 0 && n.Size < f.MinSize {
			continue
		}
		if f.MaxSize > 0 && n.Size > f.MaxSize {
			continue
		}
		out = append(out, n)
	}
	return out
}

func containsExt(exts []string, ext string) bool {
	for _, e := range exts {
		if e == ext {
			return true
		}
	}
	return false
}

// ExtGroup is a set of files sharing the same extension.
type ExtGroup struct {
	Ext   string
	Files []*Node
	Count int
	Size  int64
}

// GroupByExt buckets files by extension, sorted by group size descending;
// within each group, files are sorted by path.
func GroupByExt(files []*Node) []ExtGroup {
	byExt := make(map[string]*ExtGroup)
	for _, n := range files {
		key := n.Ext
		if key == "" {
			key = "(no extension)"
		}
		g, ok := byExt[key]
		if !ok {
			g = &ExtGroup{Ext: key}
			byExt[key] = g
		}
		g.Files = append(g.Files, n)
		g.Count++
		g.Size += n.Size
	}

	var groups []ExtGroup
	for _, g := range byExt {
		sort.Slice(g.Files, func(i, j int) bool { return g.Files[i].Path < g.Files[j].Path })
		groups = append(groups, *g)
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i].Size > groups[j].Size })
	return groups
}
