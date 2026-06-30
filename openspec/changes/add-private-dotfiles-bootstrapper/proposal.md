## Why

New machines need a public, stable entrypoint that can initialize a private dotfiles repository without making the dotfiles repository public. The bootstrapper should solve only the pre-init problem: establish GitHub SSH access, run `chezmoi init`, and hand off to the private repository.

## What Changes

- Add a small Go-based `dotfiles-bootstrap` command for macOS, Linux, and Windows.
- Add POSIX and PowerShell install entrypoints that download the released bootstrapper, verify checksums, and run it.
- Install or confirm minimum pre-init tools: Git, OpenSSH, and chezmoi.
- Detect usable private repository access before generating keys.
- Generate a user-owned Ed25519 key when SSH access and a public key are missing.
- Print the public key and GitHub SSH key settings URL, then wait for interactive confirmation before retrying.
- Verify access to the configured private dotfiles SSH remote before running `chezmoi init`.
- After initialization, run the private repository's install handoff so long-term setup remains owned by `dotfiles`.
- Add release automation and checksums for bootstrapper artifacts.

## Capabilities

### New Capabilities

- `private-bootstrapper`: Public pre-init bootstrap flow for private dotfiles setup.

### Modified Capabilities

- None.

## Impact

- Adds Go CLI source, install scripts, tests, release workflow, and README documentation in this repository.
- Publishes public GitHub Release artifacts that intentionally contain no private dotfiles contents or credentials.
- Depends on platform tools and official installers for Git, OpenSSH, and chezmoi where supported.
