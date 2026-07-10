package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/cleanup"
	"github.com/aryanwalia2003/falafal/internal/drive"
	"github.com/aryanwalia2003/falafal/internal/grep"
	"github.com/aryanwalia2003/falafal/internal/index"
	"github.com/aryanwalia2003/falafal/internal/model"
	"github.com/aryanwalia2003/falafal/internal/pattern"
	"github.com/aryanwalia2003/falafal/internal/report"
	"github.com/aryanwalia2003/falafal/internal/scanner"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

const usage = `falafal - make sense of a folder full of tons and tons of data

Usage:
  falafal <path> [flags]              Scan a local folder: tree, sizes, duplicates
  falafal drive [<folder>] [flags]    Scan a Google Drive folder (same output)
  falafal find <path> [flags]         List files matching a type/size/name filter
  falafal index <path> [flags]        Build a searchable name index for a folder
  falafal search <path> <query> [flags]  Find files by name across a folder
  falafal grep <path> <pattern> [flags]  Search file contents (like grep -r)
  falafal patterns <path> [flags]     Detect naming patterns and group matching files
  falafal --version                   Print version and exit
  falafal --help                      Show this help

Common flags (scan/drive):
  --all              include dotfiles and noise dirs (.git, node_modules, ...)
  --format FORMAT     term|text|html|json (default "term")
  --out FILE          write report to a file instead of stdout
  --top N             number of largest files to list (default 10)
  --clean             interactively move duplicate files to a local trash dir (local scan only)

find flags:
  --ext EXTS          comma-separated extensions to match, e.g. "pdf,docx" (default: any)
  --name TEXT          only files whose name contains TEXT (case-insensitive)
  --min-size SIZE      only files at least SIZE, e.g. "100MB", "1.5GiB"
  --max-size SIZE      only files at most SIZE

index/search flags:
  --all               include dotfiles and noise dirs

grep flags:
  --ignore-case, -i   case-insensitive match
  --fixed, -F          treat pattern as a literal string, not a regex
  --ext EXTS           only search files with these extensions, e.g. "go,py"
  --max-size SIZE      skip files larger than this (default 50MiB)
  --all                include dotfiles and noise dirs

patterns flags:
  --min-group N        minimum files sharing a template to report it (default 2)
  --format string      output format: text|json (default "text")
  --out string         write results to file instead of stdout
  --all                include dotfiles and noise dirs

Examples:
  falafal ~/Downloads
  falafal ~/Downloads --format html --out report.html
  falafal ~/Downloads --clean
  falafal drive "Research Data"
  falafal find ~/Downloads --ext pdf,docx
  falafal find ~/Downloads --min-size 500MB --format json --out big-files.json
  falafal search ~/Downloads "graphs report"
  falafal index ~/Downloads --out index.json
  falafal grep ~/project "TODO" --ext go,py
  falafal grep ~/project "func Test.*Error" -i
  falafal patterns ~/Downloads
  falafal patterns ~/Downloads --format json --out patterns.json

Run "falafal <command> --help" for flags specific to a command.
`

func main() {
	args := os.Args[1:]

	if len(args) == 0 || isHelpFlag(args[0]) {
		fmt.Print(usage)
		return
	}

	var err error
	switch args[0] {
	case "drive":
		err = runDrive(args[1:])
	case "find":
		err = runFind(args[1:])
	case "index":
		err = runIndex(args[1:])
	case "search":
		err = runSearch(args[1:])
	case "grep":
		err = runGrep(args[1:])
	case "patterns":
		err = runPatterns(args[1:])
	case "help":
		fmt.Print(usage)
		return
	default:
		err = run(args)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "falafal:", err)
		os.Exit(1)
	}
}

func isHelpFlag(a string) bool {
	return a == "-h" || a == "--help"
}

// boolFlags are flags that take no value, needed to correctly split flags
// from positional args regardless of the order the user types them in
// (Go's flag package otherwise stops parsing at the first positional arg).
var boolFlags = map[string]bool{
	"-all": true, "--all": true,
	"-clean": true, "--clean": true,
	"-i": true, "--ignore-case": true,
	"-F": true, "--fixed": true,
}

// reorderArgs partitions args into flag tokens (with their values) and
// positional tokens, returning flags first so flag.Parse sees them all
// regardless of where the user placed the path argument.
func reorderArgs(args []string) []string {
	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			continue
		}
		flags = append(flags, a)
		if strings.Contains(a, "=") || boolFlags[a] {
			continue
		}
		if i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}
	return append(flags, positional...)
}

func run(args []string) error {
	fs := flag.NewFlagSet("falafal", flag.ContinueOnError)
	all := fs.Bool("all", false, "include dotfiles and noise dirs (.git, node_modules, ...)")
	format := fs.String("format", "term", "output format: term|text|html|json")
	out := fs.String("out", "", "write report to file instead of stdout")
	clean := fs.Bool("clean", false, "interactively review and move duplicate files to trash")
	topN := fs.Int("top", 10, "number of largest files to show in stats")

	showVersion := fs.Bool("version", false, "print version and exit")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: falafal <path> [flags]

Scans a local folder and reports a full tree (names, types, sizes) plus
duplicate files detected by content hash.

`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return err
	}
	if *showVersion {
		fmt.Println("falafal", version)
		return nil
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("usage: falafal <path> [flags]")
	}
	root := fs.Arg(0)

	tree, err := scanner.Scan(root, scanner.Options{IncludeAll: *all, TopN: *topN})
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	if err := writeReport(tree, *format, *out); err != nil {
		return err
	}

	if *clean {
		if len(tree.Stats.DupGroups) == 0 {
			fmt.Println("No duplicates to clean up.")
			return nil
		}
		res, err := cleanup.Run(tree.Root.Path, tree.Stats.DupGroups, os.Stdin, os.Stdout)
		if err != nil {
			return fmt.Errorf("cleanup: %w", err)
		}
		fmt.Printf("\nCleanup done: %d group(s) handled, %d file(s) moved to %s, %s freed.\n",
			res.GroupsHandled, res.FilesMoved, cleanup.TrashDirName, report.HumanSize(res.BytesFreed))
	}

	return nil
}

func runDrive(args []string) error {
	fs := flag.NewFlagSet("falafal drive", flag.ContinueOnError)
	id := fs.String("id", "", "Drive folder ID (skips name search)")
	format := fs.String("format", "term", "output format: term|text|html|json")
	out := fs.String("out", "", "write report to file instead of stdout")
	topN := fs.Int("top", 10, "number of largest files to show in stats")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: falafal drive [<folder-name>] [flags]
       falafal drive --id <folder-id> [flags]

Scans a Google Drive folder (or "My Drive" root if no folder is given) and
reports a full tree plus duplicate files. First run opens a browser to sign
in with Google.

`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return err
	}
	target := fs.Arg(0) // empty target means "My Drive" root

	tree, err := drive.Scan(context.Background(), target, drive.Options{ExplicitID: *id, TopN: *topN})
	if err != nil {
		return fmt.Errorf("scanning drive: %w", err)
	}

	return writeReport(tree, *format, *out)
}

func writeReport(tree *model.Tree, format, out string) error {
	var output string
	switch format {
	case "term":
		output = report.RenderText(tree, true)
	case "text":
		output = report.RenderText(tree, false)
	case "html":
		output = report.RenderHTML(tree)
	case "json":
		b, err := report.RenderJSON(tree)
		if err != nil {
			return fmt.Errorf("rendering json: %w", err)
		}
		output = string(b)
	default:
		return fmt.Errorf("unknown format %q (want term|text|html|json)", format)
	}

	return writeOutput(output, out)
}

func writeOutput(output, out string) error {
	if out != "" {
		if err := os.WriteFile(out, []byte(output), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", out, err)
		}
		fmt.Printf("Report written to %s\n", out)
	} else {
		fmt.Println(output)
	}
	return nil
}

// runFind lists full paths of files matching a type/size/name filter -- for
// someone who just wants "show me every PDF in this folder" without wading
// through the full tree.
func runFind(args []string) error {
	fs := flag.NewFlagSet("falafal find", flag.ContinueOnError)
	all := fs.Bool("all", false, "include dotfiles and noise dirs")
	ext := fs.String("ext", "", "comma-separated extensions to match, e.g. \"pdf,docx\"")
	name := fs.String("name", "", "only files whose name contains this text")
	minSize := fs.String("min-size", "", "only files at least this size, e.g. \"100MB\"")
	maxSize := fs.String("max-size", "", "only files at most this size, e.g. \"1GiB\"")
	format := fs.String("format", "text", "output format: text|json")
	out := fs.String("out", "", "write results to file instead of stdout")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: falafal find <path> [flags]

Lists full paths of files matching a filter, grouped by type -- for finding
"every PDF" or "everything over 500MB" in a folder without reading the whole
tree.

`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("usage: falafal find <path> [flags]")
	}
	root := fs.Arg(0)

	filter := model.Filter{NameContains: *name}
	if *ext != "" {
		for _, e := range strings.Split(*ext, ",") {
			e = strings.ToLower(strings.TrimSpace(e))
			if e == "" {
				continue
			}
			if !strings.HasPrefix(e, ".") {
				e = "." + e
			}
			filter.Exts = append(filter.Exts, e)
		}
	}
	if *minSize != "" {
		v, err := report.ParseSize(*minSize)
		if err != nil {
			return fmt.Errorf("--min-size: %w", err)
		}
		filter.MinSize = v
	}
	if *maxSize != "" {
		v, err := report.ParseSize(*maxSize)
		if err != nil {
			return fmt.Errorf("--max-size: %w", err)
		}
		filter.MaxSize = v
	}

	tree, err := scanner.Scan(root, scanner.Options{IncludeAll: *all})
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	matches := model.ApplyFilter(model.AllFiles(tree.Root), filter)
	groups := model.GroupByExt(matches)

	var output string
	switch *format {
	case "text":
		output = report.RenderFindText(groups)
	case "json":
		b, err := report.RenderFindJSON(groups)
		if err != nil {
			return fmt.Errorf("rendering json: %w", err)
		}
		output = string(b)
	default:
		return fmt.Errorf("unknown format %q (want text|json)", *format)
	}

	return writeOutput(output, *out)
}

// runIndex builds an inverted index (name token -> file paths) for a folder
// and writes it as JSON, for later reuse or inspection.
func runIndex(args []string) error {
	fs := flag.NewFlagSet("falafal index", flag.ContinueOnError)
	all := fs.Bool("all", false, "include dotfiles and noise dirs")
	out := fs.String("out", "", "write index JSON to file instead of stdout")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: falafal index <path> [flags]

Builds an inverted index (name word -> file paths) for a folder and prints
it as JSON. Useful for scripting or feeding into other tools; for quick
interactive lookups use "falafal search" instead.

`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("usage: falafal index <path> [flags]")
	}
	root := fs.Arg(0)

	tree, err := scanner.Scan(root, scanner.Options{IncludeAll: *all})
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	idx := index.Build(tree.Root)
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("rendering json: %w", err)
	}

	if err := writeOutput(string(b), *out); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Indexed %d files, %d unique name tokens.\n", idx.Files, len(idx.Tokens))
	return nil
}

// runSearch finds files by name across a folder using the same tokenized
// index as `index`, but skips straight to matching paths.
func runSearch(args []string) error {
	fs := flag.NewFlagSet("falafal search", flag.ContinueOnError)
	all := fs.Bool("all", false, "include dotfiles and noise dirs")
	format := fs.String("format", "text", "output format: text|json")
	out := fs.String("out", "", "write results to file instead of stdout")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: falafal search <path> <query> [flags]

Finds files by name across a folder. All words in the query must appear in
a file's name (in any order); e.g. "falafal search . report final" matches
"Final Report.docx".

`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		fs.Usage()
		return fmt.Errorf("usage: falafal search <path> <query> [flags]")
	}
	root := fs.Arg(0)
	query := strings.Join(fs.Args()[1:], " ")

	tree, err := scanner.Scan(root, scanner.Options{IncludeAll: *all})
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	idx := index.Build(tree.Root)
	matches := idx.Search(query)

	var output string
	switch *format {
	case "text":
		output = report.RenderPathsText(matches)
	case "json":
		b, err := report.RenderPathsJSON(matches)
		if err != nil {
			return fmt.Errorf("rendering json: %w", err)
		}
		output = string(b)
	default:
		return fmt.Errorf("unknown format %q (want text|json)", *format)
	}

	return writeOutput(output, *out)
}

// runGrep searches file contents for a pattern, like a recursive grep, but
// reuses the same scan/extension-filtering machinery as find/search.
func runGrep(args []string) error {
	fs := flag.NewFlagSet("falafal grep", flag.ContinueOnError)
	all := fs.Bool("all", false, "include dotfiles and noise dirs")
	ignoreCase := fs.Bool("ignore-case", false, "case-insensitive match")
	fs.BoolVar(ignoreCase, "i", false, "shorthand for --ignore-case")
	fixed := fs.Bool("fixed", false, "treat pattern as a literal string, not a regex")
	fs.BoolVar(fixed, "F", false, "shorthand for --fixed")
	ext := fs.String("ext", "", "only search files with these extensions, e.g. \"go,py\"")
	maxSize := fs.String("max-size", "", "skip files larger than this, e.g. \"20MB\" (default 50MiB)")
	format := fs.String("format", "text", "output format: text|json")
	out := fs.String("out", "", "write results to file instead of stdout")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: falafal grep <path> <pattern> [flags]

Searches file contents for a pattern (regex by default), like a recursive
grep -n, but skips binary files and respects the same dotfile/noise-dir
rules as the rest of falafal.

`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 2 {
		fs.Usage()
		return fmt.Errorf("usage: falafal grep <path> <pattern> [flags]")
	}
	root := fs.Arg(0)
	pattern := fs.Arg(1)

	opts := grep.Options{Pattern: pattern, IgnoreCase: *ignoreCase, Fixed: *fixed}
	if *maxSize != "" {
		v, err := report.ParseSize(*maxSize)
		if err != nil {
			return fmt.Errorf("--max-size: %w", err)
		}
		opts.MaxBytes = v
	}
	re, err := grep.Compile(opts)
	if err != nil {
		return err
	}

	tree, err := scanner.Scan(root, scanner.Options{IncludeAll: *all})
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	files := model.AllFiles(tree.Root)
	if *ext != "" {
		var exts []string
		for _, e := range strings.Split(*ext, ",") {
			e = strings.ToLower(strings.TrimSpace(e))
			if e == "" {
				continue
			}
			if !strings.HasPrefix(e, ".") {
				e = "." + e
			}
			exts = append(exts, e)
		}
		files = model.ApplyFilter(files, model.Filter{Exts: exts})
	}

	matches, skipped := grep.Search(files, re, opts)

	var output string
	switch *format {
	case "text":
		output = report.RenderGrepText(matches)
	case "json":
		b, err := report.RenderGrepJSON(matches)
		if err != nil {
			return fmt.Errorf("rendering json: %w", err)
		}
		output = string(b)
	default:
		return fmt.Errorf("unknown format %q (want text|json)", *format)
	}

	if err := writeOutput(output, *out); err != nil {
		return err
	}
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "Skipped %d file(s) larger than the size cap.\n", len(skipped))
	}
	return nil
}

// runPatterns detects naming patterns across a folder's files (e.g.
// "Aryan_BE_2607.pdf", "Aryan_FE_2607.pdf", "Aryan_BE_2609.pdf" all follow
// "Aryan_{1}_{2}.pdf") and groups the files that share one.
func runPatterns(args []string) error {
	fs := flag.NewFlagSet("falafal patterns", flag.ContinueOnError)
	all := fs.Bool("all", false, "include dotfiles and noise dirs")
	minGroup := fs.Int("min-group", 2, "minimum files sharing a template to report it")
	format := fs.String("format", "text", "output format: text|json")
	out := fs.String("out", "", "write results to file instead of stdout")

	fs.Usage = func() {
		fmt.Fprint(os.Stderr, `Usage: falafal patterns <path> [flags]

Detects naming patterns across files in a folder -- e.g. "Aryan_BE_2607.pdf",
"Aryan_FE_2607.pdf", and "Aryan_BE_2609.pdf" all follow the template
"Aryan_{1}_{2}.pdf" -- and groups the files that share one, along with the
values each file has for the varying parts.

`)
		fs.PrintDefaults()
	}

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		fs.Usage()
		return fmt.Errorf("usage: falafal patterns <path> [flags]")
	}
	root := fs.Arg(0)

	tree, err := scanner.Scan(root, scanner.Options{IncludeAll: *all})
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	groups := pattern.Detect(model.AllFiles(tree.Root), pattern.Options{MinGroupSize: *minGroup})

	var output string
	switch *format {
	case "text":
		output = report.RenderPatternsText(groups)
	case "json":
		b, err := report.RenderPatternsJSON(groups)
		if err != nil {
			return fmt.Errorf("rendering json: %w", err)
		}
		output = string(b)
	default:
		return fmt.Errorf("unknown format %q (want text|json)", *format)
	}

	return writeOutput(output, *out)
}
