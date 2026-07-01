# dotfiles-bootstrap

Public bootstrapper for the private `mickmetalholic/dotfiles` repository.

This repository is intentionally small and stable. It owns first-run setup for a new machine: install or confirm the pre-init tools, install or repair the public `dot` binary, prepare GitHub SSH access, and initialize the private dotfiles repository with chezmoi.

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
- installs or repairs `dot` from public dotfiles-bootstrap GitHub Releases
- writes `dot` to `~/.local/share/dotfiles/bin/dot` on macOS/Linux
- writes `dot.exe` to `%USERPROFILE%\.local\share\dotfiles\bin\dot.exe` on Windows
- checks access to `git@github.com:mickmetalholic/dotfiles.git`
- prints an existing public SSH key, or generates `~/.ssh/id_ed25519` when no public key exists
- prints the public key and points to <https://github.com/settings/keys>
- runs `chezmoi init` after private repository access works

For a fork or test remote:

```sh
sh -c "$(curl -fsSL https://raw.githubusercontent.com/mickmetalholic/dotfiles-bootstrap/main/install.sh)" -- --repo git@github.com:owner/dotfiles.git
```

Non-interactive setup requires GitHub SSH access to be configured before running:

```sh
dotfiles-bootstrap --non-interactive
```

If repository access is not already configured, the bootstrapper prints the public key to add, the GitHub SSH key settings URL, and the exact verification command:

```sh
git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD
```

Interactive runs wait for you to add the key and then retry. Non-interactive runs exit with those instructions instead of blocking.

## Boundary

This repository does not contain private dotfiles, package catalogs, host profiles, SSH private keys, GitHub tokens, or private-repository install scripts. It does not run `install.sh`, `install.ps1`, or other setup scripts from the private dotfiles source.

Command responsibilities:

- `dotfiles-bootstrap`: first-run bootstrap, Git/OpenSSH/chezmoi readiness, GitHub SSH key guidance, `chezmoi init`, and `dot` binary install or repair
- `dot`: system package setup, doctor checks, and validation
- `chezmoi`: config apply, update, diff, and edit workflows
- `mise`: runtime and tool versions

Rule of thumb: if a task needs to repair a missing `dot` binary or initialize private source access, it belongs here. If it manages the configured system after `chezmoi init`, it belongs to `dot`, chezmoi, or mise.

## Release Artifacts

Version tags publish:

```text
dotfiles-bootstrap_darwin_amd64.tar.gz
dotfiles-bootstrap_darwin_arm64.tar.gz
dotfiles-bootstrap_linux_amd64.tar.gz
dotfiles-bootstrap_linux_arm64.tar.gz
dotfiles-bootstrap_windows_amd64.zip
dotfiles-bootstrap_windows_arm64.zip
dot_darwin_amd64.tar.gz
dot_darwin_arm64.tar.gz
dot_linux_amd64.tar.gz
dot_linux_arm64.tar.gz
dot_windows_amd64.zip
dot_windows_arm64.zip
checksums.txt
dot_checksums.txt
```

Install scripts verify checksums before executing or installing downloaded artifacts. `dot` artifacts may be listed in `checksums.txt`; when they are published with a separate checksum file, installers verify them with `dot_checksums.txt`.

## Development

```sh
go test ./...
openspec validate --all --strict
```
