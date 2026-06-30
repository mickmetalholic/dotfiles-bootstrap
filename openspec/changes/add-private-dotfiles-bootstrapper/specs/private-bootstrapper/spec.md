## ADDED Requirements

### Requirement: Public pre-init bootstrapper
The repository SHALL provide a public bootstrapper that can initialize a private dotfiles repository without requiring the dotfiles repository itself to be public.

#### Scenario: User runs bootstrapper before chezmoi init
- **WHEN** the bootstrapper runs on a machine without an initialized chezmoi source
- **THEN** it completes all pre-init work without reading private dotfiles repository files

#### Scenario: User overrides repository remote
- **WHEN** the user passes a supported repository override
- **THEN** SSH verification and `chezmoi init` use the overridden remote consistently

### Requirement: Minimum tooling readiness
The bootstrapper SHALL install or confirm Git, OpenSSH, and chezmoi before attempting private repository initialization.

#### Scenario: Required tool exists
- **WHEN** Git, OpenSSH, or chezmoi is already available
- **THEN** the bootstrapper does not reinstall that tool

#### Scenario: Required tool cannot be installed
- **WHEN** a required tool is missing and no supported installer path is available
- **THEN** the bootstrapper exits with clear manual installation guidance

### Requirement: GitHub SSH key guidance
The bootstrapper SHALL guide the user through manual GitHub SSH key registration when private repository access is unavailable.

#### Scenario: Private repository access already works
- **WHEN** `git ls-remote` succeeds for the configured private repository
- **THEN** the bootstrapper skips SSH key generation and proceeds to `chezmoi init`

#### Scenario: Default public key exists
- **WHEN** private repository access is unavailable and `~/.ssh/id_ed25519.pub` exists
- **THEN** the bootstrapper prints that public key and GitHub SSH settings guidance

#### Scenario: Default public key is missing
- **WHEN** private repository access is unavailable and no default public key exists
- **THEN** the bootstrapper generates a user-owned Ed25519 keypair and prints the public key

#### Scenario: Interactive setup waits for key registration
- **WHEN** the bootstrapper runs interactively after printing a public key
- **THEN** it waits for user confirmation before retrying private repository access

#### Scenario: Non-interactive setup lacks access
- **WHEN** the bootstrapper runs non-interactively and private repository access is unavailable
- **THEN** it exits with setup instructions instead of blocking

### Requirement: Private repository initialization
The bootstrapper SHALL initialize chezmoi from the configured private SSH remote only after repository access is verified.

#### Scenario: Repository access is verified
- **WHEN** the configured private repository is readable over SSH
- **THEN** the bootstrapper runs `chezmoi init` with that remote

#### Scenario: Chezmoi source already exists
- **WHEN** chezmoi is already initialized
- **THEN** the bootstrapper reuses the existing source path and does not clobber it

### Requirement: Private setup handoff
The bootstrapper SHALL hand off post-init setup to scripts in the private dotfiles repository.

#### Scenario: POSIX source is initialized
- **WHEN** the bootstrapper runs on POSIX after private source initialization
- **THEN** it runs the private source `install.sh` when present

#### Scenario: Windows source is initialized
- **WHEN** the bootstrapper runs on Windows after private source initialization
- **THEN** it runs the private source `install.ps1` when present

### Requirement: Public release artifacts
The repository SHALL publish versioned bootstrapper artifacts and checksums without private dotfiles contents.

#### Scenario: Release tag is pushed
- **WHEN** a version tag is pushed
- **THEN** the release workflow publishes platform-specific bootstrapper archives and `checksums.txt`

#### Scenario: User runs install entrypoint
- **WHEN** a POSIX or Windows install entrypoint downloads a bootstrapper archive
- **THEN** it verifies the archive checksum before executing the bootstrapper
