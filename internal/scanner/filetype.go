package scanner

import "strings"

var extLabels = map[string]string{
	".go":   "Go source",
	".py":   "Python source",
	".js":   "JavaScript source",
	".ts":   "TypeScript source",
	".jsx":  "React source",
	".tsx":  "React source",
	".java": "Java source",
	".c":    "C source",
	".h":    "C header",
	".cpp":  "C++ source",
	".rs":   "Rust source",
	".rb":   "Ruby source",
	".php":  "PHP source",
	".sh":   "Shell script",
	".html": "HTML document",
	".css":  "Stylesheet",
	".json": "JSON data",
	".yaml": "YAML data",
	".yml":  "YAML data",
	".xml":  "XML data",
	".toml": "TOML data",
	".md":   "Markdown document",
	".txt":  "Text document",
	".pdf":  "PDF document",
	".doc":  "Word document",
	".docx": "Word document",
	".xls":  "Excel spreadsheet",
	".xlsx": "Excel spreadsheet",
	".ppt":  "PowerPoint presentation",
	".pptx": "PowerPoint presentation",
	".csv":  "CSV data",
	".jpg":  "JPEG image",
	".jpeg": "JPEG image",
	".png":  "PNG image",
	".gif":  "GIF image",
	".svg":  "SVG image",
	".bmp":  "Bitmap image",
	".webp": "WebP image",
	".ico":  "Icon",
	".mp3":  "MP3 audio",
	".wav":  "WAV audio",
	".flac": "FLAC audio",
	".mp4":  "MP4 video",
	".mov":  "QuickTime video",
	".mkv":  "Matroska video",
	".avi":  "AVI video",
	".zip":  "ZIP archive",
	".tar":  "TAR archive",
	".gz":   "Gzip archive",
	".rar":  "RAR archive",
	".7z":   "7-Zip archive",
	".exe":  "Windows executable",
	".dll":  "Windows library",
	".so":   "Shared library",
	".dmg":  "macOS disk image",
	".sql":  "SQL script",
	".db":   "Database file",
	".log":  "Log file",
	".env":  "Environment config",
	".lock": "Lock file",
	".mod":  "Go module file",
	".sum":  "Go checksum file",
}

// TypeLabel returns a human-readable label for a file name based on its extension.
func TypeLabel(name string) (ext, label string) {
	ext = strings.ToLower(filepathExt(name))
	if lbl, ok := extLabels[ext]; ok {
		return ext, lbl
	}
	if ext == "" {
		return ext, "File (no extension)"
	}
	return ext, strings.TrimPrefix(ext, ".") + " file"
}

func filepathExt(name string) string {
	for i := len(name) - 1; i >= 0 && name[i] != '/'; i-- {
		if name[i] == '.' {
			if i == 0 {
				return "" // dotfile like ".bashrc" has no extension
			}
			return name[i:]
		}
	}
	return ""
}
