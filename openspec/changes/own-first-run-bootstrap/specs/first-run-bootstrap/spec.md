## ADDED Requirements

### Requirement: First-run installation surfaces
`dotfiles-bootstrap` SHALL provide POSIX and Windows install entrypoints as the only supported first-run setup surfaces for the private dotfiles setup.

#### Scenario: POSIX first-run entrypoint
- **WHEN** a user runs `dotfiles-bootstrap/install.sh` on macOS or Linux
- **THEN** the script downloads verified public release artifacts and runs the bootstrapper flow without requiring files from the private dotfiles repository

#### Scenario: Windows first-run entrypoint
- **WHEN** a user runs `dotfiles-bootstrap/install.ps1` on Windows
- **THEN** the script downloads verified public release artifacts and runs the bootstrapper flow without requiring files from the private dotfiles repository

#### Scenario: Private direct installers are absent
- **WHEN** the private dotfiles repository has no top-level `install.sh` or `install.ps1`
- **THEN** first-run setup remains available through the public bootstrapper install entrypoints

### Requirement: Minimum pre-init tools
The bootstrapper SHALL install or confirm Git, OpenSSH, and chezmoi before attempting private repository initialization.

#### Scenario: Required tools already exist
- **WHEN** Git, OpenSSH, and chezmoi are already available on PATH
- **THEN** the bootstrapper reports them as available and does not reinstall them

#### Scenario: Required tool can be installed
- **WHEN** a required pre-init tool is missing and a supported platform installer is available
- **THEN** the bootstrapper installs the missing tool before continuing

#### Scenario: Required tool cannot be installed
- **WHEN** a required pre-init tool is missing and no supported installer path is available
- **THEN** the bootstrapper exits with clear manual installation guidance

### Requirement: Dot binary install and repair
The installer SHALL download and install the released `dot` binary from public dotfiles-bootstrap GitHub Releases into the stable per-user dotfiles binary path.

#### Scenario: POSIX dot binary installation
- **WHEN** the POSIX installer runs successfully
- **THEN** it installs an executable `dot` binary at `~/.local/share/dotfiles/bin/dot`

#### Scenario: Windows dot binary installation
- **WHEN** the Windows installer runs successfully
- **THEN** it installs `dot.exe` at `%USERPROFILE%\.local\share\dotfiles\bin\dot.exe`

#### Scenario: Dot binary is missing
- **WHEN** the installer runs and the target `dot` binary is absent
- **THEN** the installer installs the released `dot` binary without relying on an existing `dot` command

#### Scenario: Dot binary repair or update
- **WHEN** the installer runs and the target `dot` binary is outdated, broken, or replaced by a different release version
- **THEN** the installer replaces it with the verified release artifact selected for the bootstrap run

#### Scenario: Public artifact verification
- **WHEN** the installer downloads a `dot` binary artifact
- **THEN** it verifies the artifact against public release checksums before installing it

### Requirement: GitHub SSH access preparation
The bootstrapper SHALL prepare manual GitHub SSH access for `git@github.com:mickmetalholic/dotfiles.git` without requiring GitHub CLI authentication or GitHub API tokens.

#### Scenario: Repository access already works
- **WHEN** `git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD` succeeds
- **THEN** the bootstrapper skips SSH key generation and proceeds toward chezmoi initialization

#### Scenario: Existing public SSH key is available
- **WHEN** repository access fails and an existing user-owned public SSH key is found
- **THEN** the bootstrapper prints that public key, prints `https://github.com/settings/keys`, and prints the repository verification command

#### Scenario: Public SSH key is missing
- **WHEN** repository access fails and no suitable public SSH key is found
- **THEN** the bootstrapper generates a user-owned Ed25519 keypair and prints the generated public key, `https://github.com/settings/keys`, and the repository verification command

#### Scenario: Interactive key registration
- **WHEN** the bootstrapper runs interactively after printing SSH key instructions
- **THEN** it waits for the user to add the key to GitHub before retrying `git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD`

#### Scenario: Non-interactive key registration
- **WHEN** the bootstrapper runs non-interactively and repository access is unavailable
- **THEN** it exits with clear key-registration instructions instead of waiting for input

### Requirement: Chezmoi private repository initialization
The bootstrapper SHALL run `chezmoi init git@github.com:mickmetalholic/dotfiles.git` only after repository-specific SSH access is verified.

#### Scenario: Access is verified
- **WHEN** `git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD` succeeds
- **THEN** the bootstrapper runs `chezmoi init git@github.com:mickmetalholic/dotfiles.git`

#### Scenario: Access remains unavailable
- **WHEN** repository access still fails after SSH key guidance
- **THEN** the bootstrapper does not run `chezmoi init`

#### Scenario: Repository override is provided
- **WHEN** a supported repository override is provided for testing or forks
- **THEN** repository access verification and `chezmoi init` use the same overridden SSH remote

### Requirement: Private repository boundary
The bootstrapper SHALL NOT embed private dotfiles contents or execute private-repository install scripts.

#### Scenario: Public artifacts are built
- **WHEN** dotfiles-bootstrap release artifacts are produced
- **THEN** they contain no private dotfiles contents, SSH private keys, GitHub tokens, host metadata, or private package catalogs

#### Scenario: Chezmoi source is initialized
- **WHEN** `chezmoi init` has completed successfully
- **THEN** the bootstrapper does not execute `install.sh`, `install.ps1`, or other setup scripts from the private dotfiles source

#### Scenario: GitHub credentials are unavailable
- **WHEN** GitHub CLI auth or GitHub API tokens are not configured
- **THEN** the bootstrapper can still complete SSH-guided first-run setup after the user registers a key manually

### Requirement: Command responsibility documentation
The README SHALL document the responsibilities of `dotfiles-bootstrap`, `dot`, chezmoi, and mise.

#### Scenario: Bootstrapper responsibility is documented
- **WHEN** a user reads the setup documentation
- **THEN** `dotfiles-bootstrap` is described as owning first-run bootstrap and `dot` binary install or repair

#### Scenario: Dot responsibility is documented
- **WHEN** a user reads the setup documentation
- **THEN** `dot` is described as owning system package setup, doctor checks, and validation

#### Scenario: Chezmoi responsibility is documented
- **WHEN** a user reads the setup documentation
- **THEN** chezmoi is described as owning config apply, update, diff, and edit workflows

#### Scenario: Mise responsibility is documented
- **WHEN** a user reads the setup documentation
- **THEN** mise is described as owning runtime and tool versions
