## 1. Project Foundation

- [x] 1.1 Add Go module and bootstrapper command layout.
- [x] 1.2 Add README with public install commands and private repository explanation.
- [x] 1.3 Add `.gitignore` for build outputs and temporary files.

## 2. Bootstrapper Core

- [x] 2.1 Implement flags for repository override, dry-run, and non-interactive mode.
- [x] 2.2 Implement command runner abstraction for tests.
- [x] 2.3 Implement minimum tool checks for Git, OpenSSH, and chezmoi.
- [x] 2.4 Implement private repository access verification with `git ls-remote`.
- [x] 2.5 Implement SSH key detection, generation, and public key printing.
- [x] 2.6 Implement interactive wait and non-interactive failure behavior.
- [x] 2.7 Implement `chezmoi init` and existing source reuse.
- [x] 2.8 Implement private repository handoff to `install.sh` or `install.ps1`.

## 3. Install Entrypoints And Release

- [x] 3.1 Add POSIX install script that downloads, verifies, extracts, and runs the release artifact.
- [x] 3.2 Add PowerShell install script that downloads, verifies, extracts, and runs the release artifact.
- [x] 3.3 Add GoReleaser configuration for macOS, Linux, and Windows artifacts.
- [x] 3.4 Add GitHub Actions release workflow for version tags.
- [x] 3.5 Add snapshot workflow for release artifact validation.

## 4. Tests And Validation

- [x] 4.1 Add Go tests for existing access, missing key, existing public key, generated key, non-interactive failure, chezmoi init, and handoff behavior.
- [x] 4.2 Run `go test ./...`.
- [x] 4.3 Run `openspec validate --all --strict`.
- [x] 4.4 Run POSIX syntax validation for install scripts.
- [x] 4.5 Run PowerShell syntax validation when `pwsh` is available.

Verification note: PowerShell 7.6.3 was downloaded into ignored `.tmp/` and used only for local syntax parsing.
