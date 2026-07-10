// Package grep searches file contents for a pattern, like a recursive
// grep, but reuses falafal's scan tree so results share the same file
// filtering (extensions, dotfiles/noise dirs) as the rest of the tool.
package grep

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"

	"github.com/aryanwalia2003/falafal/internal/model"
)

// maxScanBytes caps how much of a file is read to decide whether it's binary
// and, for text files, how much is scanned for matches -- keeps grep from
// stalling on multi-gigabyte files.
const maxScanBytes = 50 * 1024 * 1024 // 50 MiB

// Options controls a content search.
type Options struct {
	Pattern    string
	IgnoreCase bool
	Fixed      bool  // treat Pattern as a literal string, not a regex
	MaxBytes   int64 // skip files larger than this; 0 = use maxScanBytes
}

// Match is a single matching line.
type Match struct {
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

// Compile builds the matcher regexp from opts, escaping the pattern first if
// Fixed is set.
func Compile(opts Options) (*regexp.Regexp, error) {
	pattern := opts.Pattern
	if opts.Fixed {
		pattern = regexp.QuoteMeta(pattern)
	}
	if opts.IgnoreCase {
		pattern = "(?i)" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern %q: %w", opts.Pattern, err)
	}
	return re, nil
}

// Search scans every file's content for lines matching re, skipping files
// that look binary or exceed the size cap. It returns matches in file order,
// plus the paths of any files skipped for being too large.
func Search(files []*model.Node, re *regexp.Regexp, opts Options) (matches []Match, skippedTooLarge []string) {
	limit := opts.MaxBytes
	if limit <= 0 {
		limit = maxScanBytes
	}

	for _, n := range files {
		if n.Size > limit {
			skippedTooLarge = append(skippedTooLarge, n.Path)
			continue
		}

		f, err := os.Open(n.Path)
		if err != nil {
			continue
		}
		fileMatches, binary := searchFile(f, n.Path, re)
		f.Close()
		if binary {
			continue
		}
		matches = append(matches, fileMatches...)
	}
	return matches, skippedTooLarge
}

func searchFile(f *os.File, path string, re *regexp.Regexp) (matches []Match, binary bool) {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		if lineNo == 1 && looksBinary(line) {
			return nil, true
		}
		if re.Match(line) {
			matches = append(matches, Match{Path: path, Line: lineNo, Text: string(line)})
		}
	}
	return matches, false
}

func looksBinary(sample []byte) bool {
	return bytes.IndexByte(sample, 0) != -1
}
