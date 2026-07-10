# falafal

A command-line tool that scans a folder and shows a tree of every file —
full name, type, size — and flags duplicate files by hashing their content.
Exports to colored terminal output, plain text, a self-contained HTML report,
or JSON, and can interactively move duplicate copies to a reversible trash
folder.

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

1. Download `falafal_*_windows_amd64.zip` from the
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
tar -xzf falafal_*_linux_amd64.tar.gz -C /tmp
sudo mv /tmp/falafal_*_linux_amd64/falafal /usr/local/bin/falafal
falafal --version
```

</details>

### With Go installed (any OS)

```bash
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
