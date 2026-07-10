package drive

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/aryanwalia2003/falafal/internal/model"
	"github.com/aryanwalia2003/falafal/internal/scanner"
)

const folderMimeType = "application/vnd.google-apps.folder"

var nativeTypeLabels = map[string]string{
	"application/vnd.google-apps.document":     "Google Doc",
	"application/vnd.google-apps.spreadsheet":  "Google Sheet",
	"application/vnd.google-apps.presentation": "Google Slide",
	"application/vnd.google-apps.form":         "Google Form",
	"application/vnd.google-apps.drawing":      "Google Drawing",
	"application/vnd.google-apps.script":       "Google Apps Script",
	"application/vnd.google-apps.shortcut":     "Google Drive shortcut",
}

// Options controls a Drive scan.
type Options struct {
	ExplicitID string // Drive folder ID; skips name search if set
	TopN       int
}

// Scan walks a Drive folder (by name or ID) and builds a tree the same shape
// as a local scan, so every report renderer works unmodified.
func Scan(ctx context.Context, target string, opts Options) (*model.Tree, error) {
	client, err := HTTPClient(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("creating Drive client: %w", err)
	}

	rootID, rootName, err := resolveFolder(svc, target, opts.ExplicitID)
	if err != nil {
		return nil, err
	}

	rootNode := &model.Node{Name: rootName, Path: rootName, IsDir: true}
	var allFiles []*model.Node
	var totalDirs int

	var walk func(folderID string, node *model.Node) error
	walk = func(folderID string, node *model.Node) error {
		children, err := listChildren(svc, folderID)
		if err != nil {
			return err
		}
		sort.Slice(children, func(i, j int) bool {
			return strings.ToLower(children[i].Name) < strings.ToLower(children[j].Name)
		})

		var dirSize int64
		for _, f := range children {
			childPath := node.Path + "/" + f.Name

			if f.MimeType == folderMimeType {
				totalDirs++
				childNode := &model.Node{Name: f.Name, Path: childPath, IsDir: true}
				if err := walk(f.Id, childNode); err != nil {
					return err
				}
				node.Children = append(node.Children, childNode)
				dirSize += childNode.Size
				continue
			}

			ext, label := fileTypeLabel(f)
			fileNode := &model.Node{
				Name:      f.Name,
				Path:      childPath,
				Size:      f.Size,
				ModTime:   parseTime(f.ModifiedTime),
				Ext:       ext,
				TypeLabel: label,
				Hash:      dedupKey(f),
			}
			node.Children = append(node.Children, fileNode)
			dirSize += f.Size
			allFiles = append(allFiles, fileNode)
		}

		node.Size = dirSize
		return nil
	}

	if err := walk(rootID, rootNode); err != nil {
		return nil, err
	}

	return &model.Tree{
		Root:  rootNode,
		Stats: model.ComputeStats(rootNode, allFiles, totalDirs, opts.TopN),
	}, nil
}

// resolveFolder turns a folder name or explicit ID into a (folder ID, display
// name) pair. An empty target (or "root"/"my drive") means the Drive root.
func resolveFolder(svc *drive.Service, target, explicitID string) (id string, name string, err error) {
	if explicitID != "" {
		f, err := svc.Files.Get(explicitID).Fields("id,name,mimeType").Do()
		if err != nil {
			return "", "", fmt.Errorf("looking up folder id %q: %w", explicitID, err)
		}
		if f.MimeType != folderMimeType {
			return "", "", fmt.Errorf("%q is not a folder", explicitID)
		}
		return f.Id, f.Name, nil
	}

	if target == "" || strings.EqualFold(target, "root") || strings.EqualFold(target, "my drive") {
		return "root", "My Drive", nil
	}

	escaped := strings.ReplaceAll(target, "'", "\\'")
	q := fmt.Sprintf("name = '%s' and mimeType = '%s' and trashed = false", escaped, folderMimeType)
	res, err := svc.Files.List().Q(q).Fields("files(id, name)").Do()
	if err != nil {
		return "", "", fmt.Errorf("searching for folder %q: %w", target, err)
	}

	switch len(res.Files) {
	case 0:
		return "", "", fmt.Errorf("no folder named %q found in your Drive", target)
	case 1:
		return res.Files[0].Id, res.Files[0].Name, nil
	default:
		var sb strings.Builder
		fmt.Fprintf(&sb, "multiple folders named %q found; use --id to pick one:\n", target)
		for _, f := range res.Files {
			fmt.Fprintf(&sb, "  %s  (id: %s)\n", f.Name, f.Id)
		}
		return "", "", errors.New(sb.String())
	}
}

func listChildren(svc *drive.Service, folderID string) ([]*drive.File, error) {
	var all []*drive.File
	pageToken := ""
	for {
		call := svc.Files.List().
			Q(fmt.Sprintf("'%s' in parents and trashed = false", folderID)).
			Fields("nextPageToken, files(id, name, mimeType, size, md5Checksum, modifiedTime)").
			PageSize(1000)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		res, err := call.Do()
		if err != nil {
			return nil, fmt.Errorf("listing Drive folder: %w", err)
		}
		all = append(all, res.Files...)
		if res.NextPageToken == "" {
			break
		}
		pageToken = res.NextPageToken
	}
	return all, nil
}

// fileTypeLabel labels native Google formats (Docs/Sheets/Slides/...)
// directly since they have no file extension, and falls back to the same
// extension-based labeling used for local files otherwise.
func fileTypeLabel(f *drive.File) (ext, label string) {
	if lbl, ok := nativeTypeLabels[f.MimeType]; ok {
		return "", lbl
	}
	return scanner.TypeLabel(f.Name)
}

// dedupKey returns the content hash for regular files. Native Google formats
// have no binary content or checksum, so they fall back to a name+size
// heuristic instead (a looser signal, but the best available for those).
func dedupKey(f *drive.File) string {
	if f.Md5Checksum != "" {
		return "md5:" + f.Md5Checksum
	}
	return fmt.Sprintf("namesize:%s:%d", f.Name, f.Size)
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}
