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

$arch = Get-BootstrapArch
$archive = "dotfiles-bootstrap_windows_$arch.zip"
$baseUrl = "https://github.com/$Repo/releases/$Version/download"
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("dotfiles-bootstrap-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
  $archivePath = Join-Path $tmp $archive
  $checksumsPath = Join-Path $tmp "checksums.txt"
  Invoke-WebRequest -Uri "$baseUrl/$archive" -OutFile $archivePath
  Invoke-WebRequest -Uri "$baseUrl/checksums.txt" -OutFile $checksumsPath

  $line = Get-ChecksumLine -Checksums $checksumsPath -Archive $archive
  if (-not $line) {
    throw "checksum entry not found for $archive"
  }
  $expected = ($line -split "\s+")[0].ToLowerInvariant()
  $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $archivePath).Hash.ToLowerInvariant()
  if ($expected -ne $actual) {
    throw "checksum verification failed for $archive"
  }

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
