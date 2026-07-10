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
