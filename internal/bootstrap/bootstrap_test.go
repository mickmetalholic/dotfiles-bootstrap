package bootstrap

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type recordedCommand struct {
	name string
	args []string
}

type fakeRunner struct {
	paths    map[string]bool
	outputs  map[string][]byte
	errs     map[string]error
	commands []recordedCommand
	onRun    func(name string, args []string) error
	onOutput func(name string, args []string) ([]byte, error)
}

func (r *fakeRunner) Run(name string, args []string, env []string) error {
	r.commands = append(r.commands, recordedCommand{name: name, args: append([]string{}, args...)})
	if r.onRun != nil {
		if err := r.onRun(name, args); err != nil {
			return err
		}
	}
	if err := r.errs[name+" "+strings.Join(args, " ")]; err != nil {
		return err
	}
	return nil
}

func (r *fakeRunner) Output(name string, args []string, env []string) ([]byte, error) {
	key := name + " " + strings.Join(args, " ")
	r.commands = append(r.commands, recordedCommand{name: name, args: append([]string{}, args...)})
	if r.onOutput != nil {
		return r.onOutput(name, args)
	}
	if err := r.errs[key]; err != nil {
		return nil, err
	}
	return r.outputs[key], nil
}

func (r *fakeRunner) LookPath(name string) (string, error) {
	if r.paths[name] {
		return "/bin/" + name, nil
	}
	return "", errors.New("not found")
}

func TestExistingAccessReusesSourceWithoutPrivatePosixHandoff(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	script := filepath.Join(source, "install.sh")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{
		paths: map[string]bool{"git": true, "ssh": true, "chezmoi": true, "sh": true},
		outputs: map[string][]byte{
			"git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD": []byte("ref\n"),
			"chezmoi source-path": []byte(source + "\n"),
		},
		errs: map[string]error{},
	}
	var out, errOut strings.Builder
	code := Execute(nil, Options{
		OS:     "linux",
		Home:   home,
		Env:    []string{"PATH=/bin"},
		Runner: runner,
		Stdout: &out,
		Stderr: &errOut,
	})
	if code != 0 {
		t.Fatalf("code = %d, want 0\nstdout:\n%s\nstderr:\n%s", code, out.String(), errOut.String())
	}
	want := recordedCommand{name: "sh", args: []string{script}}
	if containsCommand(runner.commands, want) {
		t.Fatalf("private install.sh should not run: %#v", runner.commands)
	}
	if containsCommandNameArgs(runner.commands, "chezmoi", []string{"init", defaultRepo}) {
		t.Fatalf("should not run chezmoi init when source exists: %#v", runner.commands)
	}
}

func TestExistingAccessInitializesChezmoiWhenSourceMissing(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	sourceReady := false
	runner := &fakeRunner{
		paths: map[string]bool{"git": true, "ssh": true, "chezmoi": true},
		outputs: map[string][]byte{
			"git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD": []byte("ref\n"),
		},
		errs: map[string]error{},
		onRun: func(name string, args []string) error {
			if name == "chezmoi" && reflect.DeepEqual(args, []string{"init", defaultRepo}) {
				sourceReady = true
			}
			return nil
		},
		onOutput: func(name string, args []string) ([]byte, error) {
			if name == "git" && reflect.DeepEqual(args, []string{"ls-remote", defaultRepo, "HEAD"}) {
				return []byte("ref\n"), nil
			}
			if name == "chezmoi" && reflect.DeepEqual(args, []string{"source-path"}) && sourceReady {
				return []byte(source + "\n"), nil
			}
			return nil, nil
		},
	}
	var out, errOut strings.Builder
	code := Execute(nil, Options{
		OS:     "linux",
		Home:   home,
		Env:    []string{"PATH=/bin"},
		Runner: runner,
		Stdout: &out,
		Stderr: &errOut,
	})
	if code != 0 {
		t.Fatalf("code = %d, want 0\nstdout:\n%s\nstderr:\n%s", code, out.String(), errOut.String())
	}
	if !containsCommandNameArgs(runner.commands, "git", []string{"ls-remote", defaultRepo, "HEAD"}) {
		t.Fatalf("repo access was not verified: %#v", runner.commands)
	}
	if !containsCommandNameArgs(runner.commands, "chezmoi", []string{"init", defaultRepo}) {
		t.Fatalf("chezmoi init was not run: %#v", runner.commands)
	}
	if indexOfCommand(runner.commands, "git", []string{"ls-remote", defaultRepo, "HEAD"}) > indexOfCommand(runner.commands, "chezmoi", []string{"init", defaultRepo}) {
		t.Fatalf("chezmoi init ran before repository verification: %#v", runner.commands)
	}
}

func TestExistingAccessDoesNotRunPrivateWindowsHandoff(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	script := filepath.Join(source, "install.ps1")
	if err := os.WriteFile(script, []byte("Write-Host bootstrap\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{
		paths: map[string]bool{"git": true, "ssh": true, "chezmoi": true, "pwsh": true},
		outputs: map[string][]byte{
			"git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD": []byte("ref\n"),
			"chezmoi source-path": []byte(source + "\n"),
		},
		errs: map[string]error{},
	}
	code := Execute(nil, Options{
		OS:     "windows",
		Home:   home,
		Env:    []string{"PATH=/bin"},
		Runner: runner,
		Stdout: &strings.Builder{},
		Stderr: &strings.Builder{},
	})
	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if containsCommand(runner.commands, recordedCommand{name: "pwsh", args: []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script}}) {
		t.Fatalf("private install.ps1 should not run: %#v", runner.commands)
	}
}

func TestNonInteractiveMissingAccessPrintsKeyGuidance(t *testing.T) {
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	publicKey := filepath.Join(home, ".ssh", "id_ed25519.pub")
	if err := os.WriteFile(publicKey, []byte("ssh-ed25519 AAAATEST user@example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{
		paths:   map[string]bool{"git": true, "ssh": true, "chezmoi": true},
		outputs: map[string][]byte{},
		errs: map[string]error{
			"git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD": errors.New("denied"),
		},
	}
	var out, errOut strings.Builder
	code := Execute([]string{"--non-interactive"}, Options{
		OS:     "linux",
		Home:   home,
		Env:    []string{"PATH=/bin"},
		Runner: runner,
		Stdout: &out,
		Stderr: &errOut,
	})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	got := out.String() + errOut.String()
	for _, want := range []string{"ssh-ed25519 AAAATEST", "https://github.com/settings/keys", "non-interactive bootstrap requires GitHub SSH access"} {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q:\n%s", want, got)
		}
	}
	if containsCommandNameArgs(runner.commands, "chezmoi", []string{"init", defaultRepo}) {
		t.Fatalf("chezmoi init should not run without repository access: %#v", runner.commands)
	}
}

func TestExistingNonDefaultPublicKeyIsPrinted(t *testing.T) {
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_rsa.pub"), []byte("ssh-rsa AAAARSA user@example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runner := &fakeRunner{
		paths:   map[string]bool{"git": true, "ssh": true, "ssh-keygen": true, "chezmoi": true},
		outputs: map[string][]byte{},
		errs: map[string]error{
			"git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD": errors.New("denied"),
		},
	}
	var out, errOut strings.Builder
	code := Execute([]string{"--non-interactive"}, Options{
		OS:     "linux",
		Home:   home,
		Env:    []string{"PATH=/bin"},
		Runner: runner,
		Stdout: &out,
		Stderr: &errOut,
	})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	if !strings.Contains(out.String(), "ssh-rsa AAAARSA") {
		t.Fatalf("stdout missing existing public key:\n%s", out.String())
	}
	if containsCommandNameArgs(runner.commands, "ssh-keygen", []string{"-t", "ed25519", "-f", filepath.Join(home, ".ssh", "id_ed25519"), "-N", "", "-C", "dotfiles-bootstrap"}) {
		t.Fatalf("ssh-keygen should not run when an existing public key is available: %#v", runner.commands)
	}
}

func TestMissingKeyGeneratesDefaultEd25519Key(t *testing.T) {
	home := t.TempDir()
	runner := &fakeRunner{
		paths:   map[string]bool{"git": true, "ssh": true, "ssh-keygen": true, "chezmoi": true},
		outputs: map[string][]byte{},
		errs: map[string]error{
			"git ls-remote git@github.com:mickmetalholic/dotfiles.git HEAD": errors.New("denied"),
		},
	}
	runner.onRun = func(name string, args []string) error {
		if name != "ssh-keygen" {
			return nil
		}
		keyPath := filepath.Join(home, ".ssh", "id_ed25519.pub")
		return os.WriteFile(keyPath, []byte("ssh-ed25519 GENERATED dotfiles-bootstrap\n"), 0644)
	}
	var out, errOut strings.Builder
	code := Execute([]string{"--non-interactive"}, Options{
		OS:     "linux",
		Home:   home,
		Env:    []string{"PATH=/bin"},
		Runner: runner,
		Stdout: &out,
		Stderr: &errOut,
	})
	if code != 1 {
		t.Fatalf("code = %d, want 1", code)
	}
	want := recordedCommand{name: "ssh-keygen", args: []string{"-t", "ed25519", "-f", filepath.Join(home, ".ssh", "id_ed25519"), "-N", "", "-C", "dotfiles-bootstrap"}}
	if !containsCommand(runner.commands, want) {
		t.Fatalf("commands = %#v, missing %#v", runner.commands, want)
	}
	if !strings.Contains(out.String(), "ssh-ed25519 GENERATED") {
		t.Fatalf("stdout missing generated public key:\n%s", out.String())
	}
}

func TestRepoOverrideIsUsedForVerification(t *testing.T) {
	home := t.TempDir()
	source := t.TempDir()
	script := filepath.Join(source, "install.sh")
	if err := os.WriteFile(script, []byte("#!/usr/bin/env sh\n"), 0755); err != nil {
		t.Fatal(err)
	}
	repo := "git@github.com:example/private.git"
	runner := &fakeRunner{
		paths: map[string]bool{"git": true, "ssh": true, "chezmoi": true, "sh": true},
		outputs: map[string][]byte{
			"git ls-remote " + repo + " HEAD": []byte("ref\n"),
			"chezmoi source-path":             []byte(source + "\n"),
		},
		errs: map[string]error{},
	}
	code := Execute([]string{"--repo", repo}, Options{
		OS:     "linux",
		Home:   home,
		Env:    []string{"PATH=/bin"},
		Runner: runner,
		Stdout: &strings.Builder{},
		Stderr: &strings.Builder{},
	})
	if code != 0 {
		t.Fatalf("code = %d, want 0", code)
	}
	if !containsCommandNameArgs(runner.commands, "git", []string{"ls-remote", repo, "HEAD"}) {
		t.Fatalf("repo override was not verified: %#v", runner.commands)
	}
}

func TestInteractiveKeyRegistrationRetriesRepositoryAccess(t *testing.T) {
	home := t.TempDir()
	sshDir := filepath.Join(home, ".ssh")
	if err := os.MkdirAll(sshDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("ssh-ed25519 AAAATEST user@example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	source := t.TempDir()
	accessChecks := 0
	sourceReady := false
	runner := &fakeRunner{
		paths: map[string]bool{"git": true, "ssh": true, "chezmoi": true},
		errs:  map[string]error{},
		onRun: func(name string, args []string) error {
			if name == "chezmoi" && reflect.DeepEqual(args, []string{"init", defaultRepo}) {
				sourceReady = true
			}
			return nil
		},
		onOutput: func(name string, args []string) ([]byte, error) {
			if name == "git" && reflect.DeepEqual(args, []string{"ls-remote", defaultRepo, "HEAD"}) {
				accessChecks++
				if accessChecks == 1 {
					return nil, errors.New("denied")
				}
				return []byte("ref\n"), nil
			}
			if name == "chezmoi" && reflect.DeepEqual(args, []string{"source-path"}) && sourceReady {
				return []byte(source + "\n"), nil
			}
			return nil, nil
		},
	}
	var out, errOut strings.Builder
	code := Execute(nil, Options{
		OS:     "linux",
		Home:   home,
		Env:    []string{"PATH=/bin"},
		Stdin:  strings.NewReader("\n"),
		Runner: runner,
		Stdout: &out,
		Stderr: &errOut,
	})
	if code != 0 {
		t.Fatalf("code = %d, want 0\nstdout:\n%s\nstderr:\n%s", code, out.String(), errOut.String())
	}
	if accessChecks != 2 {
		t.Fatalf("accessChecks = %d, want 2", accessChecks)
	}
	if !containsCommandNameArgs(runner.commands, "chezmoi", []string{"init", defaultRepo}) {
		t.Fatalf("chezmoi init was not run after access retry: %#v", runner.commands)
	}
}

func containsCommand(commands []recordedCommand, want recordedCommand) bool {
	for _, got := range commands {
		if got.name == want.name && reflect.DeepEqual(got.args, want.args) {
			return true
		}
	}
	return false
}

func containsCommandNameArgs(commands []recordedCommand, name string, args []string) bool {
	return containsCommand(commands, recordedCommand{name: name, args: args})
}

func indexOfCommand(commands []recordedCommand, name string, args []string) int {
	for i, got := range commands {
		if got.name == name && reflect.DeepEqual(got.args, args) {
			return i
		}
	}
	return -1
}
