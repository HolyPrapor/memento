param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:USERPROFILE\bin"
)

$ErrorActionPreference = "Stop"
$repo = "HolyPrapor/memento"

function Get-Arch {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
        "Arm64"  { return "arm64" }
        default  { return "amd64" }
    }
}

$os = "windows"
$arch = Get-Arch
$ext = ".zip"

if ($Version -eq "latest") {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
    $pattern = "${os}_${arch}${ext}"
    $asset = $release.assets | Where-Object { $_.name -like "*$pattern" } | Select-Object -First 1
    if (-not $asset) {
        Write-Error "No release asset found matching $pattern"
        exit 1
    }
    $url = $asset.browser_download_url
} else {
    $tag = $Version
    if ($tag -notlike "v*") { $tag = "v$tag" }
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/tags/$tag"
    $pattern = "${os}_${arch}${ext}"
    $asset = $release.assets | Where-Object { $_.name -like "*$pattern" } | Select-Object -First 1
    if (-not $asset) {
        Write-Error "No release asset found matching $pattern"
        exit 1
    }
    $url = $asset.browser_download_url
}

Write-Host "Downloading memento from $url ..."

$tmp = "$env:TEMP\memento_install"
Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Path $tmp -Force | Out-Null

Invoke-WebRequest -Uri $url -OutFile "$tmp\memento.zip"
Expand-Archive -Path "$tmp\memento.zip" -DestinationPath "$tmp" -Force

$exe = Get-ChildItem -Path $tmp -Recurse -Filter "memento.exe" | Select-Object -First 1
if (-not $exe) { Write-Error "memento.exe not found in archive"; exit 1 }

New-Item -ItemType Directory -Path $InstallDir -Force -ErrorAction SilentlyContinue | Out-Null
Copy-Item -Force -LiteralPath $exe.FullName -Destination "$InstallDir\memento.exe"
Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$InstallDir", "User")
    Write-Host "Added $InstallDir to PATH. Reopen your terminal."
}

Write-Host "memento installed to $InstallDir\memento.exe"
