package scanner

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aryanwalia2003/falafal/internal/model"
)

var noiseDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".DS_Store":    true,
	"__pycache__":  true,
	".idea":        true,
	".vscode":      true,
}

// Options controls scan behavior.
type Options struct {
	IncludeAll bool // include dotfiles and noise dirs
	TopN       int  // largest-files list length
}

// Scan walks root and builds a full tree with hashes, duplicate groups, and stats.
func Scan(root string, opts Options) (*model.Tree, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}
	info, err := os.Stat(absRoot)
	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", absRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", absRoot)
	}

	var allFiles []*model.Node
	var totalDirs int

	rootNode := &model.Node{
		Name:    filepath.Base(absRoot),
		Path:    absRoot,
		IsDir:   true,
		ModTime: info.ModTime(),
	}

	var walk func(dirPath string, dirNode *model.Node) error
	walk = func(dirPath string, dirNode *model.Node) error {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return fmt.Errorf("reading dir %s: %w", dirPath, err)
		}
		sort.Slice(entries, func(i, j int) bool {
			return strings.ToLower(entries[i].Name()) < strings.ToLower(entries[j].Name())
		})

		var dirSize int64
		for _, entry := range entries {
			name := entry.Name()
			if !opts.IncludeAll {
				if strings.HasPrefix(name, ".") {
					continue
				}
				if entry.IsDir() && noiseDirs[name] {
					continue
				}
			}

			childPath := filepath.Join(dirPath, name)

			if entry.IsDir() {
				totalDirs++
				childNode := &model.Node{Name: name, Path: childPath, IsDir: true}
				if err := walk(childPath, childNode); err != nil {
					return err
				}
				dirNode.Children = append(dirNode.Children, childNode)
				dirSize += childNode.Size
				continue
			}

			fi, err := entry.Info()
			if err != nil {
				return fmt.Errorf("stat %s: %w", childPath, err)
			}
			if fi.Mode()&os.ModeSymlink != 0 {
				continue
			}

			hash, err := hashFile(childPath)
			if err != nil {
				return fmt.Errorf("hashing %s: %w", childPath, err)
			}
			ext, label := TypeLabel(name)

			fileNode := &model.Node{
				Name:      name,
				Path:      childPath,
				IsDir:     false,
				Size:      fi.Size(),
				ModTime:   fi.ModTime(),
				Ext:       ext,
				TypeLabel: label,
				Hash:      hash,
			}
			dirNode.Children = append(dirNode.Children, fileNode)
			dirSize += fi.Size()

			allFiles = append(allFiles, fileNode)
		}

		dirNode.Size = dirSize
		return nil
	}

	if err := walk(absRoot, rootNode); err != nil {
		return nil, err
	}

	return &model.Tree{
		Root:  rootNode,
		Stats: model.ComputeStats(rootNode, allFiles, totalDirs, opts.TopN),
	}, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
