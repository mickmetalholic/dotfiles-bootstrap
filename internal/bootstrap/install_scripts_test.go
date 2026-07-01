package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPosixInstallerDownloadsVerifiesAndInstallsDot(t *testing.T) {
	script := readRepoFile(t, "install.sh")
	for _, want := range []string{
		`dot_archive="dot_${os}_${arch}.tar.gz"`,
		`download_file "$base_url/$dot_archive" "$dot_archive_path"`,
		`verify_with_public_checksums "$dot_archive_path" "$checksums_path" "$dot_checksums_path"`,
		`install_dir="${DOTFILES_BIN_DIR:-$HOME/.local/share/dotfiles/bin}"`,
		`cp "$dot_binary" "$install_dir/dot"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("install.sh missing %q", want)
		}
	}
}

func TestWindowsInstallerDownloadsVerifiesAndInstallsDot(t *testing.T) {
	script := readRepoFile(t, "install.ps1")
	for _, want := range []string{
		`$dotArchive = "dot_windows_$arch.zip"`,
		`Invoke-WebRequest -Uri "$baseUrl/$dotArchive" -OutFile $dotArchivePath`,
		`Test-ArchiveChecksum -ArchivePath $dotArchivePath`,
		`Join-Path $HOME ".local/share/dotfiles/bin"`,
		`Join-Path $installDir "dot.exe"`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("install.ps1 missing %q", want)
		}
	}
}

func readRepoFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", name))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
