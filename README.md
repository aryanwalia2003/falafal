# falafal

A command-line tool for making sense of a folder full of tons and tons of
data. It scans a folder and shows a tree of every file — full name, type,
size — flags duplicate files by hashing their content, lets you pull out
"every PDF" or "everything over 500MB" without reading the whole tree, and
lets you search for a file by name across huge trees. Exports to colored
terminal output, plain text, a self-contained HTML report, or JSON, and can
interactively move duplicate copies to a reversible trash folder.

## Install

### Windows

Open **PowerShell** and run:

```powershell
irm https://raw.githubusercontent.com/aryanwalia2003/falafal/main/install.ps1 | iex
```

This downloads the latest `falafal.exe`, puts it in
`%LOCALAPPDATA%\Programs\falafal`, and adds that folder to your PATH — no
manual zip extraction or Environment Variables dialog needed. **Close and
reopen your terminal**, then run `falafal --version` to confirm.

<details>
<summary>Manual install (if the script doesn't work, e.g. restricted lab computers)</summary>

1. Download `falafal_windows_amd64.zip` from the
   [releases page](https://github.com/aryanwalia2003/falafal/releases).
2. Extract it, then move `falafal.exe` somewhere on your `PATH`, or add its
   folder to PATH via Settings → System → About → Advanced system settings →
   Environment Variables.
3. Open a new PowerShell/cmd window and run `falafal --version`.

</details>

### Linux / macOS

```bash
curl -fsSL https://raw.githubusercontent.com/aryanwalia2003/falafal/main/install.sh | bash
```

This downloads the right binary for your OS/architecture, installs it to
`~/.local/bin`, and adds that to your PATH if it isn't already there. Restart
your terminal (or `source` your shell rc file as instructed), then run
`falafal --version`.

<details>
<summary>Manual install</summary>

Download the matching tarball from the
[releases page](https://github.com/aryanwalia2003/falafal/releases):

```bash
tar -xzf falafal_linux_amd64.tar.gz -C /tmp
sudo mv /tmp/falafal_linux_amd64/falafal /usr/local/bin/falafal
falafal --version
```

</details>

### With Go installed (any OS)

```bash
go install github.com/aryanwalia2003/falafal@latest
```

### Troubleshooting: "No such host is known" / network errors

Some campus, lab, or corporate networks block specific GitHub subdomains
(most commonly `api.github.com`) while allowing `github.com` itself. The
installers above only ever talk to `github.com`'s release-download redirect,
not the API, so this shouldn't happen — but if it still does:
- Try a different network (e.g. a phone hotspot) to confirm it's a network
  policy issue, not something wrong with your machine.
- Fall back to the manual install steps above, which only need a browser.

## Usage

```
falafal <path> [flags]

  --all               include dotfiles and noise dirs (.git, node_modules, ...)
  --format string     output format: term|text|html|json (default "term")
  --out string        write report to file instead of stdout
  --clean             interactively review and move duplicate files to trash
  --top int           number of largest files to show in stats (default 10)
  --version           print version and exit
```

### Examples

```bash
# Colored tree in the terminal
falafal ~/Downloads

# Plain text, redirect-friendly
falafal ~/Downloads --format text > tree.txt

# Self-contained HTML report (collapsible tree, duplicate links, stats panel)
falafal ~/Downloads --format html --out report.html

# JSON, for scripting
falafal ~/Downloads --format json --out report.json

# Interactively move duplicate copies to .falafal-trash (reversible)
falafal ~/Downloads --clean
```

By default, dotfiles and common noise directories (`.git`, `node_modules`,
`vendor`, `__pycache__`, `.idea`, `.vscode`) are skipped. Pass `--all` to
include them.

## Finding files in tons of data

Five commands help when you're lost in a huge folder and don't want the
full tree.

### `falafal find` — pull out files by type/size/name

```
falafal find <path> [flags]

  --ext string          comma-separated extensions to match, e.g. "pdf,docx"
  --name string         only files whose name contains this text
  --min-size string     only files at least this size, e.g. "100MB", "1.5GiB"
  --max-size string     only files at most this size
  --format string       output format: text|json (default "text")
  --out string          write results to file instead of stdout
  --all                 include dotfiles and noise dirs
```

```bash
# Every PDF and Word doc, full paths, grouped by type
falafal find ~/Downloads --ext pdf,docx

# Everything over 500MB, exported to JSON
falafal find ~/Downloads --min-size 500MB --format json --out big-files.json

# Anything with "invoice" in the name
falafal find ~/Downloads --name invoice
```

### `falafal search` — find a file by name

```
falafal search <path> <query> [flags]

  --format string    output format: text|json (default "text")
  --out string       write results to file instead of stdout
  --all              include dotfiles and noise dirs
```

All words in the query must appear somewhere in a file's name (in any order,
case-insensitive) — you don't need the exact title:

```bash
falafal search ~/Downloads "graphs report"
falafal search ~/Downloads final draft --format json --out matches.json
```

### `falafal index` — build a reusable name index

Builds an inverted index (name word → file paths) for a folder and prints it
as JSON — useful for scripting, or feeding into another tool. For one-off
interactive lookups, use `falafal search` instead.

```
falafal index <path> [flags]

  --out string    write index JSON to file instead of stdout
  --all           include dotfiles and noise dirs
```

```bash
falafal index ~/Downloads --out index.json
```

### `falafal grep` — search file contents (like `grep -r`)

```
falafal grep <path> <pattern> [flags]

  --ignore-case, -i    case-insensitive match
  --fixed, -F           treat pattern as a literal string, not a regex
  --ext string          only search files with these extensions, e.g. "go,py"
  --max-size string     skip files larger than this (default 50MiB)
  --format string       output format: text|json (default "text")
  --out string          write results to file instead of stdout
  --all                 include dotfiles and noise dirs
```

`<pattern>` is a regular expression by default (Go's `regexp` syntax, similar
to `egrep`); pass `--fixed`/`-F` to match it literally instead. Files that
look binary (a null byte in their first line) are skipped automatically, and
matches print as `path:line:text`, like `grep -rn`.

```bash
# All TODO/FIXME comments in Go and Python files
falafal grep ~/project "TODO|FIXME" --ext go,py

# Case-insensitive test function names
falafal grep ~/project "func Test.*Error" -i

# Literal string match, exported as JSON
falafal grep ~/project "api.github.com" -F --format json --out matches.json
```

### `falafal patterns` — detect naming patterns and group matching files

If a folder is full of files like `Aryan_BE_2607.pdf`, `Aryan_FE_2607.pdf`,
`Aryan_BE_2609.pdf`, `Rahul_CS_1001.pdf`, `Rahul_CS_1002.pdf`, this finds the
naming templates automatically and groups the files that follow each one —
here, `Aryan_{1}_{2}.pdf` (3 files) and `Rahul_CS_{1}.pdf` (2 files) — along
with the actual values each file has for the varying parts.

```
falafal patterns <path> [flags]

  --min-group int      minimum files sharing a template to report it (default 2)
  --format string      output format: text|json (default "text")
  --out string         write results to file instead of stdout
  --all                include dotfiles and noise dirs
```

```bash
falafal patterns ~/Downloads
falafal patterns ~/Downloads --min-group 3
falafal patterns ~/Downloads --format json --out patterns.json
```

Detection works per file extension: names are split into runs of letters,
digits, and separators (`_`, `-`, space, `.`); files whose separators and
segment structure match exactly are candidates for the same template, and
whichever segments actually differ (the "obvious" varying parts) become
`{1}`, `{2}`, ... placeholders. It automatically finds the most specific
template that still covers at least `--min-group` files — favoring
`Rahul_CS_{1}.pdf` over a vaguer `{1}_{2}_{3}.pdf` when the data supports it.
Files with an identical name (no varying segment) aren't reported here —
that's exact duplication, already covered by duplicate detection.

## Google Drive

```
falafal drive <folder-name> [flags]
falafal drive --id <folder-id> [flags]

  --id string          Drive folder ID (skips name search)
  --format string      output format: term|text|html|json (default "term")
  --out string         write report to file instead of stdout
  --top int            number of largest files to show in stats (default 10)
```

```bash
# Scan a folder by name (searches your whole Drive)
falafal drive "Research Data"

# Scan your entire My Drive
falafal drive

# Scan by folder ID (use this if the name matches more than one folder)
falafal drive --id 1AbCdEfGhIjKlMnOpQrStUvWxYz

# HTML report of a Drive folder
falafal drive "Research Data" --format html --out drive-report.html
```

The first time you run `falafal drive`, it opens your browser to sign in to
Google and asks for **read-only** access to your Drive. After that it caches
a token locally so you won't need to sign in again on that machine.

**Note:** this app is currently in Google's "Testing" publishing mode, which
means only Google accounts the app owner has explicitly allowlisted can sign
in (a hard cap of 100 accounts). If you get an "access blocked" error when
signing in, ask whoever set this up to add your Google email to the test
users list in their Google Cloud Console project.

Duplicate detection on Drive works differently for two kinds of files:
- **Regular files** (PDFs, images, zips, etc.) are compared by Drive's own
  MD5 checksum of their content — same guarantee as local scanning.
- **Native Google formats** (Docs, Sheets, Slides, Forms, Drawings) have no
  binary content to hash, so they're compared by name + reported size
  instead. This is a looser signal than a real hash, so treat `[DUP]` tags
  on these as "worth checking," not certain.

`--clean` (interactive duplicate cleanup) is local-only for now; Drive
scanning is read-only.

## How duplicate detection works

Every local file's content is hashed with SHA-256. Files sharing a hash are
grouped and tagged `[DUP:Dn]` in the tree, with total wasted space (all but
one copy per group) shown in the summary.

## Building from source

```bash
git clone https://github.com/aryanwalia2003/falafal.git
cd falafal
go build -o falafal .
```

To cross-compile release archives for Linux, macOS, and Windows:

```bash
make release VERSION=v0.1.0
```

Output lands in `dist/`.

## Roadmap

- `--clean` support for Google Drive (move duplicates to Drive Trash)
- Move the Google OAuth app out of Testing mode so anyone can sign in
  without being allowlisted
