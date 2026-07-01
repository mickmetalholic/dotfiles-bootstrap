## 1. Installer Release Download And Dot Installation

- [x] 1.1 Update `install.sh` to download and checksum-verify the public `dot` release artifact for the detected POSIX OS and architecture.
- [x] 1.2 Update `install.sh` to install or replace `dot` at `~/.local/share/dotfiles/bin/dot` without requiring an existing `dot` command.
- [x] 1.3 Update `install.ps1` to download and checksum-verify the public `dot` release artifact for the detected Windows architecture.
- [x] 1.4 Update `install.ps1` to install or replace `dot.exe` at `%USERPROFILE%\.local\share\dotfiles\bin\dot.exe` without requiring an existing `dot` command.
- [x] 1.5 Update release artifact naming or checksum handling so bootstrapper and `dot` artifacts are first-class public release assets.

## 2. Bootstrap Flow Ownership

- [x] 2.1 Keep Git, OpenSSH, and chezmoi install-or-confirm behavior before any private repository initialization.
- [x] 2.2 Verify private repository access with `git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD` before `chezmoi init`.
- [x] 2.3 Search for an existing user-owned public SSH key before generating a new Ed25519 keypair when repository access fails.
- [x] 2.4 Print the public key, `https://github.com/settings/keys`, and the exact `git ls-remote` verification command when SSH setup is required.
- [x] 2.5 Preserve interactive wait-and-retry behavior after key registration.
- [x] 2.6 Preserve non-interactive fail-fast behavior with clear setup instructions and no blocking prompt.
- [x] 2.7 Run `chezmoi init git@github.com:mickmetalholic/dotfiles.git` only after repository-specific SSH access is verified.

## 3. Remove Private Handoff

- [x] 3.1 Remove execution of private source `install.sh` from the POSIX bootstrap path.
- [x] 3.2 Remove execution of private source `install.ps1` from the Windows bootstrap path.
- [x] 3.3 Ensure the bootstrapper succeeds after verified `chezmoi init` and public `dot` installation without requiring private top-level install scripts.
- [x] 3.4 Ensure public bootstrapper artifacts do not include private dotfiles contents, GitHub tokens, SSH private keys, host metadata, or private package catalogs.

## 4. Documentation

- [x] 4.1 Update README install documentation to describe the public first-run surfaces and `dot` binary install or repair behavior.
- [x] 4.2 Update README boundary documentation to state that private repository install scripts are not run and are not required.
- [x] 4.3 Document command responsibilities: `dotfiles-bootstrap` owns first-run bootstrap and `dot` binary install/repair; `dot` owns system package setup, doctor, and validation; chezmoi owns config apply/update/diff/edit; mise owns runtime/tool versions.
- [x] 4.4 Document non-interactive SSH requirements and manual GitHub key registration instructions for POSIX and Windows users.

## 5. Tests And Validation

- [x] 5.1 Add or update Go tests for existing repository access, existing public key selection, generated Ed25519 key, interactive retry, non-interactive failure, and guarded `chezmoi init`.
- [x] 5.2 Add or update Go tests proving private `install.sh` and `install.ps1` are not executed after chezmoi initialization.
- [x] 5.3 Add POSIX installer tests or validation covering `dot` artifact download, checksum verification, and installation to `~/.local/share/dotfiles/bin/dot`.
- [x] 5.4 Add Windows installer tests or validation covering `dot.exe` artifact download, checksum verification, and installation to `%USERPROFILE%\.local\share\dotfiles\bin\dot.exe`.
- [x] 5.5 Run `go test ./...`.
- [x] 5.6 Run POSIX shell syntax validation for `install.sh`.
- [x] 5.7 Run PowerShell syntax validation for `install.ps1` when `pwsh` is available.
- [x] 5.8 Run `openspec validate --all --strict`.
