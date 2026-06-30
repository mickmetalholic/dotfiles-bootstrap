## Context

`dotfiles-bootstrap` is a public sibling repository for a private `dotfiles` repository. Its responsibility ends once the private repository is initialized and its own setup flow is running.

The bootstrapper must be safe to publish. It can contain repository coordinates and installer logic, but it must not contain private dotfiles content, SSH private keys, GitHub tokens, host metadata from the private repository, or package catalogs.

## Goals / Non-Goals

**Goals:**

- Provide one public POSIX and one public Windows command for first-run setup.
- Keep the Go command small and deterministic.
- Confirm or install Git, OpenSSH, and chezmoi before private repository initialization.
- Generate a default Ed25519 SSH key only when needed.
- Give the user clear manual instructions to add the public key to GitHub.
- Verify the configured private repository remote with `git ls-remote`.
- Run `chezmoi init` and then execute the private repository's install handoff.
- Support non-interactive mode only when SSH access is already configured.

**Non-Goals:**

- Do not manage packages, host profiles, or chezmoi templates.
- Do not duplicate the private repository's `dot` CLI.
- Do not upload keys to GitHub automatically.
- Do not depend on `gh auth login`.
- Do not embed a private repository snapshot.

## Decisions

1. Use Go for the bootstrapper.

   Go gives a single cross-platform binary with simple process orchestration and testable command planning. Shell and PowerShell remain thin download/run wrappers.

2. Keep only one primary command.

   The default command performs bootstrap. Flags configure `--repo`, `--yes`, `--non-interactive`, and `--dry-run`. A larger subcommand surface would encourage this repo to become a second dotfiles tool.

3. Verify access with `git ls-remote`.

   `ssh -T git@github.com` is useful guidance, but the authoritative gate is whether the configured private repository can be read. `git ls-remote <repo> HEAD` is the final check before `chezmoi init`.

4. Generate `~/.ssh/id_ed25519` by default.

   This matches the private dotfiles SSH config expectation. If a key already exists, the bootstrapper prints the existing public key rather than replacing it.

5. Use private repository handoff after `chezmoi init`.

   After `chezmoi init`, the bootstrapper runs `install.sh` or `install.ps1` from the private source if present. This keeps binary installation, `chezmoi apply`, `dot setup`, and `dot doctor` in the private repo.

## Risks / Trade-offs

- [Risk] The bootstrapper may be run on machines without supported package managers. -> Mitigation: report exact missing tools and manual install guidance.
- [Risk] Users can have custom SSH key names. -> Mitigation: first try repository access; only generate or print the default key when access fails.
- [Risk] Interactive key registration is manual. -> Mitigation: print the full public key and GitHub settings URL.
- [Risk] Running private `install.sh` after `chezmoi init` may re-run `chezmoi init`. -> Mitigation: private direct installer is idempotent enough for authenticated setup, and the bootstrapper can prefer direct script execution only after source path is available.

## Open Questions

- Should the bootstrapper open the GitHub SSH settings URL automatically, or only print it?
- Should a future release support a dotfiles-specific key name through a flag?
