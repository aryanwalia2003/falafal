# falafal

A command-line tool that scans a folder and shows a tree of every file —
full name, type, size — and flags duplicate files by hashing their content.
Exports to colored terminal output, plain text, a self-contained HTML report,
or JSON, and can interactively move duplicate copies to a reversible trash
folder.

## Install

### Linux

Download the latest `linux_amd64` (or `linux_arm64`) tarball from the
[releases page](https://github.com/aryanwalia2003/falafal/releases), then:

```bash
tar -xzf falafal_*_linux_amd64.tar.gz -C /tmp
sudo mv /tmp/falafal_*_linux_amd64/falafal /usr/local/bin/falafal
falafal --version
```

Or, with Go installed:

```bash
go install github.com/aryanwalia2003/falafal@latest
```

### Windows

1. Download `falafal_*_windows_amd64.zip` from the
   [releases page](https://github.com/aryanwalia2003/falafal/releases).
2. Extract it, then move `falafal.exe` somewhere on your `PATH` (e.g.
   `C:\Windows\System32` or a custom folder added to `PATH`).
3. Open PowerShell or cmd and run:
   ```powershell
   falafal --version
   ```

Or, with Go installed:

```powershell
go install github.com/aryanwalia2003/falafal@latest
```

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

## How duplicate detection works

Every file's content is hashed with SHA-256. Files sharing a hash are grouped
and tagged `[DUP:Dn]` in the tree, with total wasted space (all but one copy
per group) shown in the summary.

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

- Google Drive folder scanning (reuses the same tree model and reports)
