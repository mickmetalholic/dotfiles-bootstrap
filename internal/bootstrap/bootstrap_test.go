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

func TestExistingAccessReusesSourceAndRunsPosixHandoff(t *testing.T) {
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
	if !containsCommand(runner.commands, want) {
		t.Fatalf("commands = %#v, missing %#v", runner.commands, want)
	}
	if containsCommandNameArgs(runner.commands, "chezmoi", []string{"init", defaultRepo}) {
		t.Fatalf("should not run chezmoi init when source exists: %#v", runner.commands)
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
