// Package pattern detects naming patterns across a set of files -- e.g.
// "Aryan_BE_2607.pdf", "Aryan_FE_2607.pdf", "Aryan_BE_2609.pdf" all follow
// the template "Aryan_{1}_{2}.pdf" -- and groups the files that share one.
package pattern

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
)

type segKind int

const (
	kindSep segKind = iota // punctuation/whitespace, e.g. "_", "-", " ", "."
	kindAlpha
	kindDigit
)

type segment struct {
	kind  segKind
	value string
}

var tokenRe = regexp.MustCompile(`[A-Za-z]+|[0-9]+|[^A-Za-z0-9]+`)

// splitName breaks a file's base name (including extension) into typed
// segments: runs of letters, runs of digits, and separator runs.
func splitName(name string) []segment {
	parts := tokenRe.FindAllString(name, -1)
	segs := make([]segment, 0, len(parts))
	for _, p := range parts {
		k := kindSep
		switch {
		case isAlpha(p):
			k = kindAlpha
		case isDigit(p):
			k = kindDigit
		}
		segs = append(segs, segment{kind: k, value: p})
	}
	return segs
}

func isAlpha(s string) bool {
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')) {
			return false
		}
	}
	return true
}

func isDigit(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// shapeKey summarizes the segment structure: separators contribute their
// literal text (they must match exactly for two names to share a pattern),
// while alpha/digit runs contribute only their kind, since their value is
// exactly what a pattern is allowed to vary.
func shapeKey(segs []segment) string {
	var sb strings.Builder
	for _, s := range segs {
		switch s.kind {
		case kindSep:
			sb.WriteString("S:")
			sb.WriteString(s.value)
		case kindAlpha:
			sb.WriteString("A")
		case kindDigit:
			sb.WriteString("D")
		}
		sb.WriteByte(',')
	}
	return sb.String()
}

// Group is a set of files that share a detected naming template.
type Group struct {
	Template string
	Files    []*model.Node
	Fields   [][]string // per-file, in template placeholder order
}

// Options controls pattern detection.
type Options struct {
	MinGroupSize int // minimum files sharing a template to report; default 2
}

// bucket is a set of files whose stems have identical segment shape (same
// count/kind of alpha/digit runs and identical separators), and thus are
// candidates for sharing a naming template.
type bucket struct {
	ext   string
	segs  [][]segment
	files []*model.Node
}

// Detect groups files by naming template. Files are first bucketed by
// extension + segment shape, then, within each bucket, recursively split on
// whichever varying segment cleanly divides the bucket into sub-groups that
// are still each at least minGroupSize -- so a mix of "Aryan_BE_2607.pdf" /
// "Aryan_FE_2607.pdf" / "Aryan_BE_2609.pdf" and "Rahul_CS_1001.pdf" /
// "Rahul_CS_1002.pdf" yields two specific templates ("Aryan_{1}_{2}.pdf" and
// "Rahul_CS_{1}.pdf") rather than one over-generalized one. Only files whose
// base name has at least one segment that varies across the final group are
// reported -- files that merely share an identical name aren't a "pattern,"
// they're duplicates (already handled by content-hash dedup).
func Detect(files []*model.Node, opts Options) []Group {
	minSize := opts.MinGroupSize
	if minSize <= 0 {
		minSize = 2
	}

	buckets := make(map[string]*bucket)
	for _, n := range files {
		stem := strings.TrimSuffix(n.Name, n.Ext)
		segs := splitName(stem)
		key := n.Ext + "|" + shapeKey(segs)
		b, ok := buckets[key]
		if !ok {
			b = &bucket{ext: n.Ext}
			buckets[key] = b
		}
		b.segs = append(b.segs, segs)
		b.files = append(b.files, n)
	}

	var groups []Group
	for _, b := range buckets {
		if len(b.files) < minSize {
			continue
		}
		indices := make([]int, len(b.files))
		for i := range indices {
			indices[i] = i
		}
		refine(b, indices, minSize, &groups)
	}

	sort.Slice(groups, func(i, j int) bool {
		if len(groups[i].Files) != len(groups[j].Files) {
			return len(groups[i].Files) > len(groups[j].Files)
		}
		return groups[i].Template < groups[j].Template
	})
	return groups
}

// varyingPositions reports, for each segment position, whether it differs
// across the given subset of a bucket's files.
func varyingPositions(b *bucket, indices []int) []bool {
	n := len(b.segs[indices[0]])
	varying := make([]bool, n)
	for i := 0; i < n; i++ {
		if b.segs[indices[0]][i].kind == kindSep {
			continue
		}
		first := b.segs[indices[0]][i].value
		for _, idx := range indices[1:] {
			if b.segs[idx][i].value != first {
				varying[i] = true
				break
			}
		}
	}
	return varying
}

// refine tries to split indices (a subset of one shape bucket) into more
// specific sub-groups; if no split keeps every resulting sub-group at least
// minSize, it reports indices itself as a leaf group (when it has at least
// one varying segment).
func refine(b *bucket, indices []int, minSize int, groups *[]Group) {
	varying := varyingPositions(b, indices)

	for pos, v := range varying {
		if !v {
			continue
		}
		partitions := make(map[string][]int)
		var order []string
		for _, idx := range indices {
			val := b.segs[idx][pos].value
			if _, ok := partitions[val]; !ok {
				order = append(order, val)
			}
			partitions[val] = append(partitions[val], idx)
		}
		if len(order) < 2 {
			continue
		}
		clean := true
		for _, val := range order {
			if len(partitions[val]) < minSize {
				clean = false
				break
			}
		}
		if !clean {
			continue
		}
		for _, val := range order {
			refine(b, partitions[val], minSize, groups)
		}
		return
	}

	if len(indices) < minSize {
		return
	}
	anyVarying := false
	for _, v := range varying {
		if v {
			anyVarying = true
			break
		}
	}
	if !anyVarying {
		return
	}

	*groups = append(*groups, buildGroup(b, indices, varying))
}

func buildGroup(b *bucket, indices []int, varying []bool) Group {
	n := len(varying)
	rep := b.segs[indices[0]]

	var tmpl strings.Builder
	varIdx := 0
	for i := 0; i < n; i++ {
		if varying[i] {
			varIdx++
			fmt.Fprintf(&tmpl, "{%d}", varIdx)
		} else {
			tmpl.WriteString(rep[i].value)
		}
	}
	tmpl.WriteString(b.ext)

	sorted := append([]int{}, indices...)
	sort.Slice(sorted, func(i, j int) bool { return b.files[sorted[i]].Path < b.files[sorted[j]].Path })

	files := make([]*model.Node, len(sorted))
	fields := make([][]string, len(sorted))
	for i, idx := range sorted {
		files[i] = b.files[idx]
		var vals []string
		for p := 0; p < n; p++ {
			if varying[p] {
				vals = append(vals, b.segs[idx][p].value)
			}
		}
		fields[i] = vals
	}

	return Group{Template: tmpl.String(), Files: files, Fields: fields}
}
