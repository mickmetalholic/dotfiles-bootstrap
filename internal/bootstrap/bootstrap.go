package bootstrap

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const defaultRepo = "git@github.com:mickmetalholic/dotfiles.git"

type Runner interface {
	Run(name string, args []string, env []string) error
	Output(name string, args []string, env []string) ([]byte, error)
	LookPath(name string) (string, error)
}

type Options struct {
	OS     string
	Env    []string
	Home   string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
	Runner Runner
}

type realRunner struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func (r realRunner) Run(name string, args []string, env []string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdin = r.stdin
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr
	return cmd.Run()
}

func (r realRunner) Output(name string, args []string, env []string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stderr = io.Discard
	return cmd.Output()
}

func (r realRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

type config struct {
	repo           string
	dryRun         bool
	nonInteractive bool
}

func Execute(args []string, opts Options) int {
	stdout := opts.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}
	stdin := opts.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	env := opts.Env
	if env == nil {
		env = os.Environ()
	} else {
		env = append([]string{}, env...)
	}
	goos := opts.OS
	if goos == "" {
		goos = runtime.GOOS
	}
	home := opts.Home
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(stderr, "could not locate home directory")
			return 1
		}
	}
	runner := opts.Runner
	if runner == nil {
		runner = realRunner{stdin: stdin, stdout: stdout, stderr: stderr}
	}

	cfg, err := parseArgs(args, env)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if envBool(env, "DOTFILES_NONINTERACTIVE") || envBool(env, "CI") {
		cfg.nonInteractive = true
	}

	return run(cfg, runner, goos, home, stdin, stdout, stderr, env)
}

func parseArgs(args []string, env []string) (config, error) {
	cfg := config{repo: firstNonEmpty(envValue(env, "DOTFILES_REPO"), defaultRepo)}
	fs := flag.NewFlagSet("dotfiles-bootstrap", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&cfg.repo, "repo", cfg.repo, "private dotfiles SSH remote")
	fs.BoolVar(&cfg.dryRun, "dry-run", false, "print actions without changing the machine")
	fs.BoolVar(&cfg.nonInteractive, "non-interactive", false, "fail instead of waiting for manual GitHub key registration")
	fs.BoolVar(&cfg.nonInteractive, "yes", false, "alias for --non-interactive")
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if fs.NArg() > 0 {
		return cfg, fmt.Errorf("unknown argument: %s", fs.Arg(0))
	}
	return cfg, nil
}

func run(cfg config, r Runner, goos, home string, stdin io.Reader, stdout, stderr io.Writer, env []string) int {
	fmt.Fprintln(stdout, "dotfiles bootstrap")
	if !ensureMinimumTools(cfg, r, goos, home, stdout, stderr, &env) {
		return 1
	}
	if !repoReadable(r, cfg.repo, env) {
		if !prepareSSHKey(cfg, r, home, stdin, stdout, stderr, env) {
			return 1
		}
		if cfg.dryRun {
			fmt.Fprintf(stdout, "dry-run: would retry private repository access for %s\n", cfg.repo)
		} else if !repoReadable(r, cfg.repo, env) {
			fmt.Fprintf(stderr, "private repository is still not readable: %s\n", cfg.repo)
			fmt.Fprintf(stderr, "verify with: git ls-remote %s HEAD\n", cfg.repo)
			return 1
		}
	}
	source, ok := ensureChezmoiSource(cfg, r, cfg.repo, stdout, stderr, env)
	if !ok {
		return 1
	}
	if !runPrivateHandoff(cfg, r, goos, source, stdout, stderr, env) {
		return 1
	}
	fmt.Fprintln(stdout, "bootstrap complete")
	return 0
}

func ensureMinimumTools(cfg config, r Runner, goos, home string, stdout, stderr io.Writer, env *[]string) bool {
	ok := true
	if !ensureTool(cfg, r, goos, "git", "Git", stdout, stderr, env) {
		ok = false
	}
	if !ensureTool(cfg, r, goos, "ssh", "OpenSSH", stdout, stderr, env) {
		ok = false
	}
	if !ensureChezmoi(cfg, r, goos, home, stdout, stderr, env) {
		ok = false
	}
	return ok
}

func ensureTool(cfg config, r Runner, goos, command, label string, stdout, stderr io.Writer, env *[]string) bool {
	if hasCommand(r, command) {
		fmt.Fprintf(stdout, "ok: %s available\n", label)
		return true
	}
	if cfg.dryRun {
		fmt.Fprintf(stdout, "dry-run: would install or require %s\n", label)
		return true
	}
	switch {
	case goos == "linux" && command == "git" && hasCommand(r, "apt-get"):
		return runInstall(r, "sudo", []string{"apt-get", "install", "-y", "git"}, *env, label, stderr)
	case goos == "linux" && command == "ssh" && hasCommand(r, "apt-get"):
		return runInstall(r, "sudo", []string{"apt-get", "install", "-y", "openssh-client"}, *env, label, stderr)
	case goos == "windows" && command == "git" && hasCommand(r, "winget"):
		return runInstall(r, "winget", []string{"install", "--id", "Git.Git", "--exact", "--accept-package-agreements", "--accept-source-agreements"}, *env, label, stderr)
	default:
		fmt.Fprintf(stderr, "%s is required. Install %s and re-run dotfiles-bootstrap.\n", label, label)
		return false
	}
}

func ensureChezmoi(cfg config, r Runner, goos, home string, stdout, stderr io.Writer, env *[]string) bool {
	if hasCommand(r, "chezmoi") {
		fmt.Fprintln(stdout, "ok: chezmoi available")
		return true
	}
	if cfg.dryRun {
		fmt.Fprintln(stdout, "dry-run: would install or require chezmoi")
		return true
	}
	if goos == "windows" && hasCommand(r, "winget") {
		return runInstall(r, "winget", []string{"install", "--id", "twpayne.chezmoi", "--exact", "--accept-package-agreements", "--accept-source-agreements"}, *env, "chezmoi", stderr)
	}
	if goos != "windows" && hasCommand(r, "sh") && (hasCommand(r, "curl") || hasCommand(r, "wget")) {
		bindir := filepath.Join(home, ".local", "bin")
		args := []string{"-c", fmt.Sprintf("curl -fsLS https://get.chezmoi.io | sh -s -- -b %s", shellQuote(bindir))}
		if !hasCommand(r, "curl") {
			args = []string{"-c", fmt.Sprintf("wget -qO- https://get.chezmoi.io | sh -s -- -b %s", shellQuote(bindir))}
		}
		if !runInstall(r, "sh", args, *env, "chezmoi", stderr) {
			return false
		}
		*env = prependPath(*env, bindir)
		_ = os.Setenv("PATH", envValue(*env, "PATH"))
		return true
	}
	fmt.Fprintln(stderr, "chezmoi is required. Install it from https://www.chezmoi.io/install/ and re-run dotfiles-bootstrap.")
	return false
}

func runInstall(r Runner, name string, args []string, env []string, label string, stderr io.Writer) bool {
	if err := r.Run(name, args, env); err != nil {
		fmt.Fprintf(stderr, "failed to install %s: %v\n", label, err)
		return false
	}
	return true
}

func repoReadable(r Runner, repo string, env []string) bool {
	if !hasCommand(r, "git") {
		return false
	}
	_, err := r.Output("git", []string{"ls-remote", repo, "HEAD"}, env)
	return err == nil
}

func prepareSSHKey(cfg config, r Runner, home string, stdin io.Reader, stdout, stderr io.Writer, env []string) bool {
	sshDir := filepath.Join(home, ".ssh")
	privateKey := filepath.Join(sshDir, "id_ed25519")
	publicKey := privateKey + ".pub"
	if _, err := os.Stat(publicKey); errors.Is(err, os.ErrNotExist) {
		if cfg.dryRun {
			fmt.Fprintf(stdout, "dry-run: would generate SSH key at %s\n", privateKey)
		} else {
			if err := os.MkdirAll(sshDir, 0700); err != nil {
				fmt.Fprintf(stderr, "could not create %s: %v\n", sshDir, err)
				return false
			}
			if err := r.Run("ssh-keygen", []string{"-t", "ed25519", "-f", privateKey, "-N", "", "-C", "dotfiles-bootstrap"}, env); err != nil {
				fmt.Fprintf(stderr, "failed to generate SSH key: %v\n", err)
				return false
			}
		}
	}
	keyText := ""
	if !cfg.dryRun {
		data, err := os.ReadFile(publicKey)
		if err != nil {
			fmt.Fprintf(stderr, "could not read public key %s: %v\n", publicKey, err)
			return false
		}
		keyText = strings.TrimSpace(string(data))
	}
	fmt.Fprintln(stdout, "\nGitHub SSH key setup required")
	fmt.Fprintln(stdout, "Add this public key at https://github.com/settings/keys:")
	if keyText == "" {
		fmt.Fprintf(stdout, "  %s\n", publicKey)
	} else {
		fmt.Fprintf(stdout, "\n%s\n\n", keyText)
	}
	fmt.Fprintf(stdout, "Then verify with: git ls-remote %s HEAD\n", cfg.repo)
	if cfg.nonInteractive {
		fmt.Fprintln(stderr, "non-interactive bootstrap requires GitHub SSH access to be configured before running.")
		return false
	}
	if cfg.dryRun {
		return true
	}
	fmt.Fprint(stdout, "Press Enter after adding the key to GitHub...")
	_, _ = bufio.NewReader(stdin).ReadString('\n')
	return true
}

func ensureChezmoiSource(cfg config, r Runner, repo string, stdout, stderr io.Writer, env []string) (string, bool) {
	if source := chezmoiSource(r, env); source != "" {
		fmt.Fprintf(stdout, "ok: chezmoi source exists at %s\n", source)
		return source, true
	}
	if cfg.dryRun {
		fmt.Fprintf(stdout, "dry-run: would run chezmoi init %s\n", repo)
		return "", true
	}
	if err := r.Run("chezmoi", []string{"init", repo}, env); err != nil {
		fmt.Fprintf(stderr, "chezmoi init failed: %v\n", err)
		return "", false
	}
	source := chezmoiSource(r, env)
	if source == "" {
		fmt.Fprintln(stderr, "chezmoi source-path returned empty after init")
		return "", false
	}
	return source, true
}

func chezmoiSource(r Runner, env []string) string {
	if !hasCommand(r, "chezmoi") {
		return ""
	}
	out, err := r.Output("chezmoi", []string{"source-path"}, env)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func runPrivateHandoff(cfg config, r Runner, goos, source string, stdout, stderr io.Writer, env []string) bool {
	if source == "" {
		if cfg.dryRun {
			fmt.Fprintln(stdout, "dry-run: would run private repository handoff after chezmoi init")
			return true
		}
		fmt.Fprintln(stderr, "cannot run private handoff without chezmoi source path")
		return false
	}
	if goos == "windows" {
		script := filepath.Join(source, "install.ps1")
		if _, err := os.Stat(script); err != nil {
			fmt.Fprintf(stderr, "private handoff script missing: %s\n", script)
			return false
		}
		if cfg.dryRun {
			fmt.Fprintf(stdout, "dry-run: would run %s\n", script)
			return true
		}
		return r.Run("pwsh", []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script}, env) == nil
	}
	script := filepath.Join(source, "install.sh")
	if _, err := os.Stat(script); err != nil {
		fmt.Fprintf(stderr, "private handoff script missing: %s\n", script)
		return false
	}
	if cfg.dryRun {
		fmt.Fprintf(stdout, "dry-run: would run %s\n", script)
		return true
	}
	if err := r.Run("sh", []string{script}, env); err != nil {
		fmt.Fprintf(stderr, "private handoff failed: %v\n", err)
		return false
	}
	return true
}

func hasCommand(r Runner, name string) bool {
	_, err := r.LookPath(name)
	return err == nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}

func envBool(env []string, key string) bool {
	value := strings.ToLower(envValue(env, key))
	return value == "1" || value == "true" || value == "yes"
}

func prependPath(env []string, dir string) []string {
	next := append([]string{}, env...)
	for i, item := range next {
		if strings.HasPrefix(item, "PATH=") {
			next[i] = "PATH=" + dir + string(os.PathListSeparator) + strings.TrimPrefix(item, "PATH=")
			return next
		}
	}
	return append(next, "PATH="+dir)
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
