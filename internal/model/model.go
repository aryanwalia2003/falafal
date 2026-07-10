package model

import "time"

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
