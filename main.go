package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/cleanup"
	"github.com/aryanwalia2003/falafal/internal/report"
	"github.com/aryanwalia2003/falafal/internal/scanner"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "falafal:", err)
		os.Exit(1)
	}
}

// boolFlags are flags that take no value, needed to correctly split flags
// from positional args regardless of the order the user types them in
// (Go's flag package otherwise stops parsing at the first positional arg).
var boolFlags = map[string]bool{"-all": true, "--all": true, "-clean": true, "--clean": true}

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

	if err := fs.Parse(reorderArgs(args)); err != nil {
		return err
	}
	if *showVersion {
		fmt.Println("falafal", version)
		return nil
	}
	if fs.NArg() < 1 {
		return fmt.Errorf("usage: falafal <path> [flags]")
	}
	root := fs.Arg(0)

	tree, err := scanner.Scan(root, scanner.Options{IncludeAll: *all, TopN: *topN})
	if err != nil {
		return fmt.Errorf("scanning: %w", err)
	}

	var output string
	switch *format {
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
		return fmt.Errorf("unknown format %q (want term|text|html|json)", *format)
	}

	if *out != "" {
		if err := os.WriteFile(*out, []byte(output), 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", *out, err)
		}
		fmt.Printf("Report written to %s\n", *out)
	} else {
		fmt.Println(output)
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
