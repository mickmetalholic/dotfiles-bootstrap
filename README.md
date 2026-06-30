# dotfiles-bootstrap

Public bootstrapper for the private `mickmetalholic/dotfiles` repository.

This repository is intentionally small and stable. It only prepares a new machine to read the private dotfiles repository over GitHub SSH, initializes chezmoi, and hands off to the private repository for all real setup.

## Install

Windows:

```powershell
irm https://raw.githubusercontent.com/mickmetalholic/dotfiles-bootstrap/main/install.ps1 | iex
```

macOS/Linux:

```sh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/mickmetalholic/dotfiles-bootstrap/main/install.sh)"
```

The bootstrapper:

- installs or confirms Git, OpenSSH, and chezmoi where supported
- checks access to `git@github.com:mickmetalholic/dotfiles.git`
- generates `~/.ssh/id_ed25519` when no default public key exists
- prints the public key and points to <https://github.com/settings/keys>
- runs `chezmoi init` after private repository access works
- runs the private repository setup handoff

For a fork or test remote:

```sh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/mickmetalholic/dotfiles-bootstrap/main/install.sh)" -- --repo git@github.com:owner/dotfiles.git
```

Non-interactive setup requires GitHub SSH access to be configured before running:

```sh
dotfiles-bootstrap --non-interactive
```

## Boundary

This repository does not contain private dotfiles, package catalogs, host profiles, SSH private keys, GitHub tokens, or daily `dot` workflows. After `chezmoi init`, those responsibilities belong to the private `dotfiles` repository.

Rule of thumb: if a task can run after `chezmoi init`, it belongs in `dotfiles`, not here.

## Release Artifacts

Version tags publish:

```text
dotfiles-bootstrap_darwin_amd64.tar.gz
dotfiles-bootstrap_darwin_arm64.tar.gz
dotfiles-bootstrap_linux_amd64.tar.gz
dotfiles-bootstrap_linux_arm64.tar.gz
dotfiles-bootstrap_windows_amd64.zip
dotfiles-bootstrap_windows_arm64.zip
checksums.txt
```

Install scripts verify checksums before executing downloaded artifacts.

The same public release may also contain `dot_*` archives and `dot_checksums.txt` mirrored from the private `dotfiles` repository. Those artifacts are consumed by the private repository's `scripts/install-dot-binary.*` helpers after `chezmoi init`.

## Development

```sh
go test ./...
openspec validate --all --strict
```
