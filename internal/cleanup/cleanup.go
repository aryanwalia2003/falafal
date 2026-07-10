package cleanup

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
	"github.com/aryanwalia2003/falafal/internal/report"
)

const TrashDirName = ".falafal-trash"

// Result summarizes what an interactive cleanup run did.
type Result struct {
	GroupsHandled int
	FilesMoved    int
	BytesFreed    int64
}

// Run walks each duplicate group, asks the user which copy to keep, and moves the
// rest into a trash directory under root (never deletes directly, so it's reversible).
func Run(root string, groups []model.DupGroupInfo, in io.Reader, out io.Writer) (Result, error) {
	var res Result
	scanner := bufio.NewScanner(in)
	trashDir := filepath.Join(root, TrashDirName)

	for _, g := range groups {
		fmt.Fprintf(out, "\nDuplicate group (%s, %d copies):\n", report.HumanSize(g.Nodes[0].Size), len(g.Nodes))
		for i, n := range g.Nodes {
			fmt.Fprintf(out, "  [%d] %s  (modified %s)\n", i+1, n.Path, n.ModTime.Format("2006-01-02 15:04:05"))
		}
		fmt.Fprintf(out, "Keep which one? [1-%d, default 1, s=skip]: ", len(g.Nodes))

		keep := 1
		if scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "s" || line == "S" {
				continue
			}
			if line != "" {
				if v, err := strconv.Atoi(line); err == nil && v >= 1 && v <= len(g.Nodes) {
					keep = v
				}
			}
		}

		res.GroupsHandled++
		for i, n := range g.Nodes {
			if i+1 == keep {
				continue
			}
			if err := moveToTrash(n, trashDir); err != nil {
				fmt.Fprintf(out, "  ! failed to move %s: %v\n", n.Path, err)
				continue
			}
			res.FilesMoved++
			res.BytesFreed += n.Size
			fmt.Fprintf(out, "  moved %s -> trash\n", n.Path)
		}
	}

	return res, nil
}

// moveToTrash relocates a file into trashDir, flattening its path into the
// filename so duplicate basenames from different folders don't collide.
func moveToTrash(n *model.Node, trashDir string) error {
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return err
	}
	flatName := strings.ReplaceAll(strings.TrimPrefix(n.Path, string(filepath.Separator)), string(filepath.Separator), "__")
	dest := filepath.Join(trashDir, flatName)

	if err := os.Rename(n.Path, dest); err != nil {
		return copyThenRemove(n.Path, dest)
	}
	return nil
}

func copyThenRemove(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Remove(src)
}
