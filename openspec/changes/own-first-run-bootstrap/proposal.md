## Why

The public bootstrapper should be the only first-run entrypoint for private dotfiles setup, because the private repository is removing its top-level install scripts and a missing `dot` binary cannot repair itself. This change clarifies ownership so public bootstrap artifacts can initialize access and install/repair `dot` without embedding private dotfiles content or requiring GitHub API authentication.

## What Changes

- Make `dotfiles-bootstrap/install.sh` and `install.ps1` the supported first-run installation surfaces for POSIX and Windows.
- Install or confirm the minimum pre-init tools: Git, OpenSSH, and chezmoi.
- Download the released `dot` binary from public dotfiles-bootstrap GitHub Releases and install it into `~/.local/share/dotfiles/bin/dot` or the Windows equivalent.
- Verify private repository SSH access with `git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD` before running `chezmoi init`.
- When repository access fails, find an existing public SSH key or generate a user-owned Ed25519 key, print the public key and GitHub SSH key settings URL, then retry after user confirmation in interactive mode.
- In non-interactive mode, exit with clear key-registration instructions instead of blocking.
- Run `chezmoi init git@github.com:mickmetalholic/dotfiles.git` after repository-specific SSH access is verified.
- **BREAKING** Stop running private-repo `install.sh` or `install.ps1` handoff scripts; the private dotfiles repository will no longer provide those first-run entrypoints.
- Keep bootstrapper-owned `dot` binary repair and update logic.
- Document command responsibilities for `dotfiles-bootstrap`, `dot`, `chezmoi`, and `mise`.
- Update README and OpenSpec coverage for POSIX and Windows behavior.

## Capabilities

### New Capabilities

- `first-run-bootstrap`: Public first-run bootstrap and `dot` binary install/repair flow for a private dotfiles setup.

### Modified Capabilities

- None.

## Impact

- Updates POSIX and Windows install scripts, bootstrapper command behavior, release download/install logic, and tests.
- Updates README responsibilities and setup documentation.
- Removes bootstrapper dependency on private-repository install scripts while preserving private repository SSH initialization through chezmoi.
- Maintains a public artifact boundary: no GitHub CLI auth, GitHub API token, private dotfiles contents, private keys, or private-repo script execution.
