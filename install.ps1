#Requires -Version 5.1
<#
Installs the latest falafal release for Windows: downloads the binary,
puts it in a per-user folder, and adds that folder to the user PATH so
`falafal` works in any new terminal without manual setup.

Usage (PowerShell):
  irm https://raw.githubusercontent.com/aryanwalia2003/falafal/main/install.ps1 | iex
#>

$ErrorActionPreference = "Stop"

$repo = "aryanwalia2003/falafal"
$assetName = "falafal_windows_amd64.zip"
# github.com/<repo>/releases/latest/download/<asset> redirects straight to the
# right file without ever touching api.github.com, which some campus/lab
# networks block even though github.com and the download CDN work fine.
$downloadUrl = "https://github.com/$repo/releases/latest/download/$assetName"
$installDir = Join-Path $env:LOCALAPPDATA "Programs\falafal"

$zipPath = Join-Path $env:TEMP $assetName
Write-Host "Downloading $assetName..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath

Write-Host "Installing to $installDir..."
if (Test-Path $installDir) {
    Remove-Item -Recurse -Force $installDir
}
New-Item -ItemType Directory -Path $installDir -Force | Out-Null

$extractTemp = Join-Path $env:TEMP "falafal-extract"
if (Test-Path $extractTemp) { Remove-Item -Recurse -Force $extractTemp }
Expand-Archive -Path $zipPath -DestinationPath $extractTemp -Force

$exe = Get-ChildItem -Path $extractTemp -Filter "falafal.exe" -Recurse | Select-Object -First 1
Copy-Item $exe.FullName -Destination (Join-Path $installDir "falafal.exe") -Force

Remove-Item -Recurse -Force $extractTemp
Remove-Item -Force $zipPath

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    Write-Host "Adding $installDir to your PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
}

Write-Host ""
Write-Host "falafal installed to $installDir"
Write-Host "Open a NEW terminal window, then run: falafal --version"
