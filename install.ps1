[CmdletBinding()]
param(
  [string]$Repo = $(if ($env:DOTFILES_BOOTSTRAP_REPO) { $env:DOTFILES_BOOTSTRAP_REPO } else { "mickmetalholic/dotfiles-bootstrap" }),
  [string]$Version = $(if ($env:DOTFILES_BOOTSTRAP_VERSION) { $env:DOTFILES_BOOTSTRAP_VERSION } else { "latest" })
)

$ErrorActionPreference = "Stop"

function Get-BootstrapArch {
  switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64"; return }
    "ARM64" { "arm64"; return }
    default { throw "unsupported architecture for dotfiles-bootstrap: $env:PROCESSOR_ARCHITECTURE" }
  }
}

function Get-ChecksumLine {
  param([string]$Checksums, [string]$Archive)
  Get-Content -LiteralPath $Checksums | Where-Object { $_ -match "\s$([regex]::Escape($Archive))$" } | Select-Object -First 1
}

function Test-ArchiveChecksum {
  param([string]$ArchivePath, [string]$ChecksumsPath)

  $archiveName = Split-Path -Leaf $ArchivePath
  $line = Get-ChecksumLine -Checksums $ChecksumsPath -Archive $archiveName
  if (-not $line) {
    return $false
  }
  $expected = ($line -split "\s+")[0].ToLowerInvariant()
  $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $ArchivePath).Hash.ToLowerInvariant()
  return $expected -eq $actual
}

function Install-DotBinary {
  param([string]$ArchivePath, [string]$TempRoot)

  $extractPath = Join-Path $TempRoot "dot"
  $installDir = if ($env:DOTFILES_BIN_DIR) {
    $env:DOTFILES_BIN_DIR
  } else {
    Join-Path $HOME ".local/share/dotfiles/bin"
  }
  New-Item -ItemType Directory -Path $extractPath -Force | Out-Null
  New-Item -ItemType Directory -Path $installDir -Force | Out-Null
  Expand-Archive -LiteralPath $ArchivePath -DestinationPath $extractPath -Force
  $dotBinary = Join-Path $extractPath "dot.exe"
  if (-not (Test-Path -LiteralPath $dotBinary)) {
    $dotBinary = Get-ChildItem -LiteralPath $extractPath -Filter "dot.exe" -File -Recurse | Select-Object -First 1 -ExpandProperty FullName
  }
  if (-not $dotBinary) {
    throw "archive did not contain dot.exe"
  }
  $target = Join-Path $installDir "dot.exe"
  Copy-Item -LiteralPath $dotBinary -Destination $target -Force
  Write-Host "installed dot to $target"
}

$arch = Get-BootstrapArch
$archive = "dotfiles-bootstrap_windows_$arch.zip"
$dotArchive = "dot_windows_$arch.zip"
$baseUrl = "https://github.com/$Repo/releases/$Version/download"
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("dotfiles-bootstrap-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
  $archivePath = Join-Path $tmp $archive
  $dotArchivePath = Join-Path $tmp $dotArchive
  $checksumsPath = Join-Path $tmp "checksums.txt"
  $dotChecksumsPath = Join-Path $tmp "dot_checksums.txt"
  Invoke-WebRequest -Uri "$baseUrl/$archive" -OutFile $archivePath
  Invoke-WebRequest -Uri "$baseUrl/$dotArchive" -OutFile $dotArchivePath
  Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile $checksumsPath

  if (-not (Test-ArchiveChecksum -ArchivePath $archivePath -ChecksumsPath $checksumsPath)) {
    throw "checksum verification failed for $archive"
  }
  if (-not (Test-ArchiveChecksum -ArchivePath $dotArchivePath -ChecksumsPath $checksumsPath)) {
    Invoke-WebRequest -Uri "$baseUrl/dot_checksums.txt" -OutFile $dotChecksumsPath
    if (-not (Test-ArchiveChecksum -ArchivePath $dotArchivePath -ChecksumsPath $dotChecksumsPath)) {
      throw "checksum verification failed for $dotArchive"
    }
  }
  Install-DotBinary -ArchivePath $dotArchivePath -TempRoot $tmp

  Expand-Archive -LiteralPath $archivePath -DestinationPath $tmp -Force
  $binary = Join-Path $tmp "dotfiles-bootstrap.exe"
  if (-not (Test-Path -LiteralPath $binary)) {
    throw "archive did not contain dotfiles-bootstrap.exe"
  }
  & $binary @args
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} finally {
  Remove-Item -LiteralPath $tmp -Recurse -Force -ErrorAction SilentlyContinue
}
