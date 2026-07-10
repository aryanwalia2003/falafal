package model

import (
	"fmt"
	"sort"
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
