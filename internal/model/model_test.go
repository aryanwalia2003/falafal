package model

import "testing"

func TestComputeStatsDuplicateGrouping(t *testing.T) {
	a := &Node{Name: "a", Path: "/a", Size: 10, Hash: "h1"}
	b := &Node{Name: "b", Path: "/b", Size: 10, Hash: "h1"}
	c := &Node{Name: "c", Path: "/c", Size: 5, Hash: "h2"}
	root := &Node{Name: "root", IsDir: true, Size: 25}

	stats := ComputeStats(root, []*Node{a, b, c}, 0, 10)

	if stats.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, want 3", stats.TotalFiles)
	}
	if len(stats.DupGroups) != 1 {
		t.Fatalf("DupGroups = %d, want 1", len(stats.DupGroups))
	}
	if stats.WastedSize != 10 {
		t.Errorf("WastedSize = %d, want 10", stats.WastedSize)
	}
	if a.DupGroup == "" || a.DupGroup != b.DupGroup {
		t.Errorf("a and b should share a DupGroup, got %q and %q", a.DupGroup, b.DupGroup)
	}
	if c.DupGroup != "" {
		t.Errorf("c should not be marked a duplicate, got %q", c.DupGroup)
	}
}

func TestComputeStatsTopNCap(t *testing.T) {
	files := []*Node{
		{Name: "a", Path: "/a", Size: 30, Hash: "h1"},
		{Name: "b", Path: "/b", Size: 20, Hash: "h2"},
		{Name: "c", Path: "/c", Size: 10, Hash: "h3"},
	}
	root := &Node{Name: "root", IsDir: true, Size: 60}

	stats := ComputeStats(root, files, 0, 2)

	if len(stats.LargestFiles) != 2 {
		t.Fatalf("LargestFiles len = %d, want 2", len(stats.LargestFiles))
	}
	if stats.LargestFiles[0].Size != 30 || stats.LargestFiles[1].Size != 20 {
		t.Errorf("LargestFiles not sorted by size desc: %+v", stats.LargestFiles)
	}
}
