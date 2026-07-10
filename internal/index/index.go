// Package index builds a simple inverted index (word -> file paths) over a
// scanned tree, so users can search for files by name across huge trees
// without remembering exact paths.
package index

import (
	"sort"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
)

// Index maps a lowercase name token to the sorted, deduplicated list of full
// file paths whose name contains that token.
type Index struct {
	Tokens map[string][]string `json:"tokens"`
	Files  int                 `json:"files"`
}

// Build walks root and indexes every file's name into tokens.
func Build(root *model.Node) *Index {
	idx := &Index{Tokens: make(map[string][]string)}

	for _, n := range model.AllFiles(root) {
		idx.Files++
		seen := make(map[string]bool)
		for _, tok := range Tokenize(n.Name) {
			if seen[tok] {
				continue
			}
			seen[tok] = true
			idx.Tokens[tok] = append(idx.Tokens[tok], n.Path)
		}
	}

	for tok := range idx.Tokens {
		sort.Strings(idx.Tokens[tok])
	}
	return idx
}

// Search returns full paths of files whose name contains every token in
// query (a simple AND across the inverted index), sorted alphabetically.
func (idx *Index) Search(query string) []string {
	terms := Tokenize(query)
	if len(terms) == 0 {
		return nil
	}

	result := idx.Tokens[terms[0]]
	for _, t := range terms[1:] {
		result = intersect(result, idx.Tokens[t])
		if len(result) == 0 {
			return nil
		}
	}
	if len(result) == 0 {
		return nil
	}

	out := append([]string{}, result...)
	sort.Strings(out)
	return out
}

func intersect(a, b []string) []string {
	set := make(map[string]bool, len(b))
	for _, s := range b {
		set[s] = true
	}
	var out []string
	for _, s := range a {
		if set[s] {
			out = append(out, s)
		}
	}
	return out
}

// Tokenize lowercases s and splits it into alphanumeric words, discarding
// punctuation/separators. Both the raw extension-less stem and, for
// convenience, the extension itself (without the dot) become tokens, so
// searching "pdf" or "report" both work as expected.
func Tokenize(s string) []string {
	s = strings.ToLower(s)
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			tokens = append(tokens, cur.String())
			cur.Reset()
		}
	}
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			cur.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return tokens
}
