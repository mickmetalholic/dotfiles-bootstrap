## Context

`dotfiles-bootstrap` is the public entrypoint for a private `mickmetalholic/dotfiles` setup. The current bootstrap flow already installs or confirms Git, OpenSSH, and chezmoi, verifies repository access with `git ls-remote`, and initializes chezmoi. It also runs `install.sh` or `install.ps1` from the private source after initialization.

That private handoff is no longer valid because the private repository will remove its top-level install scripts. The bootstrapper must own first-run setup through the point where the private repository is initialized and the public `dot` binary is available. After that, day-to-day setup and validation are owned by `dot`, chezmoi, and mise.

The public release boundary remains strict: bootstrapper artifacts can include installer code, repository coordinates, and public binaries, but not private dotfiles contents, SSH private keys, GitHub tokens, host profiles, or private package catalogs.

## Goals / Non-Goals

**Goals:**

- Keep POSIX and Windows install scripts as the only first-run surfaces.
- Ensure Git, OpenSSH, and chezmoi exist before repository initialization.
- Install or repair the released `dot` binary into the bootstrapper-owned user bin path:
  - POSIX: `~/.local/share/dotfiles/bin/dot`
  - Windows: the equivalent per-user dotfiles binary directory under the user profile.
- Verify `git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD` before `chezmoi init`.
- Use manual GitHub SSH key registration without GitHub CLI auth or GitHub API tokens.
- Support interactive retry after key registration and non-interactive fail-fast instructions.
- Run `chezmoi init git@github.com:mickmetalholic/dotfiles.git` only after repository-specific SSH access is verified.
- Document command ownership across `dotfiles-bootstrap`, `dot`, `chezmoi`, and `mise`.
- Cover POSIX and Windows behavior in tasks and validation.

**Non-Goals:**

- Do not run private repository install scripts or any private post-init handoff.
- Do not embed private dotfiles contents in public release artifacts.
- Do not make `dot` responsible for repairing a missing `dot` binary.
- Do not use GitHub CLI authentication, GitHub API tokens, or automatic SSH key upload.
- Do not move chezmoi apply/update/diff/edit responsibilities into the bootstrapper.
- Do not move runtime/tool version management out of mise.

## Decisions

1. Keep installer scripts thin but make them install the public `dot` binary.

   The POSIX and PowerShell scripts should continue to download verified public GitHub Release artifacts. They should install `dotfiles-bootstrap` into a temporary execution path and install `dot` into the stable per-user dotfiles bin path. This keeps first-run repair possible even when `dot` is missing or broken.

   Alternative considered: have `dotfiles-bootstrap` download `dot` after startup. That would work, but the scripts already own release artifact download and checksum verification, and placing binary installation there keeps bootstrapping consistent across POSIX and Windows.

2. Treat private repository readability as the SSH authority.

   The bootstrapper should use `git ls-remote <repo> HEAD` as the gate before `chezmoi init`. General GitHub SSH authentication can succeed while repository-level access still fails, so the check must target the private dotfiles repository.

   Alternative considered: use `ssh -T git@github.com` as the primary gate. That is less precise and can report successful authentication for an account that lacks repository access.

3. Generate or print user-owned SSH keys only after access fails.

   The bootstrapper should first check repository access. If access fails, it should locate an existing public SSH key, preferring common user-owned keys such as Ed25519, or generate a new Ed25519 keypair when no suitable public key exists. It should print the public key and `https://github.com/settings/keys`.

   Alternative considered: always generate a dedicated key. That is predictable but creates unnecessary keys on machines where an existing key already works or only needs to be added to GitHub.

4. Split interactive and non-interactive behavior explicitly.

   Interactive runs should wait for the user to add the printed key, then retry `git ls-remote`. Non-interactive runs should print exact instructions and exit without waiting. This prevents CI or scripted setup from hanging while still giving humans a simple first-run loop.

   Alternative considered: loop with polling until access appears. That obscures what action is required and can hang indefinitely.

5. Stop executing private install scripts after `chezmoi init`.

   The bootstrapper should finish after verified `chezmoi init` and public `dot` installation. Private setup continues through documented commands: `dot` for system package setup, doctor, and validation; chezmoi for config apply/update/diff/edit; mise for runtime and tool versions.

   Alternative considered: invoke a `dot` setup command automatically. That would blur the ownership boundary and make the first-run bootstrapper responsible for post-init system setup. The first-run surface should establish prerequisites and make the tools available, then leave explicit next steps.

## Risks / Trade-offs

- [Risk] Public releases must contain usable `dot` artifacts without private contents. -> Mitigation: release packaging and tests verify artifact names, checksums, and install destinations while keeping private source files out of bootstrapper artifacts.
- [Risk] Machines can have custom SSH key names. -> Mitigation: check repository access first, then search common public key locations before generating a default Ed25519 key.
- [Risk] Non-interactive first-run cannot complete without preconfigured SSH access. -> Mitigation: exit with the public key, GitHub settings URL, and `git ls-remote` verification command.
- [Risk] Removing private handoff changes the end of the first-run experience. -> Mitigation: README documents the command responsibility split and follow-up commands clearly.
- [Risk] Windows and POSIX path handling can diverge. -> Mitigation: add tests and syntax validation for both installers, plus bootstrap tests for OS-specific dot binary destination behavior.

## Migration Plan

1. Update installer scripts to download, checksum-verify, and install public `dot` release artifacts into the stable per-user bin directory.
2. Update bootstrapper behavior to remove private handoff execution and keep SSH/chezmoi initialization.
3. Update release metadata and tests so `dot` artifacts are first-class public bootstrapper release assets.
4. Update README and specs to describe the new responsibility boundary and remove private install handoff instructions.
5. Validate POSIX shell syntax, PowerShell syntax when available, Go tests, and OpenSpec strict validation.

Rollback is to restore the previous release and bootstrapper behavior that expected private `install.sh` or `install.ps1`, but that only works while those private scripts still exist.

## Open Questions

- What exact Windows install path should be standardized as the equivalent of `~/.local/share/dotfiles/bin/dot`: `%USERPROFILE%\\.local\\share\\dotfiles\\bin\\dot.exe` or `%LOCALAPPDATA%\\dotfiles\\bin\\dot.exe`?
- Should first-run print suggested next commands only, or run a safe read-only validation such as `dot doctor` when the binary is installed?
