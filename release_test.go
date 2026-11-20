package mdcli_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const releaseWorkflow = ".github/workflows/release.yml"

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// The tap token comes from a GitHub App: the workflow needs the app's client ID
// as a secret, not as a variable and not passed as app-id.
func TestReleaseWorkflowUsesSecretForHomebrewTapAppClientID(t *testing.T) {
	text := readFile(t, releaseWorkflow)

	const want = "client-id: ${{ secrets.HOMEBREW_TAP_APP_CLIENT_ID }}"
	if !strings.Contains(text, want) {
		t.Fatalf("release workflow should use %q", want)
	}

	for _, forbidden := range []string{
		"vars.HOMEBREW_TAP_APP_CLIENT_ID",
		"app-id: ${{ secrets.HOMEBREW_TAP_APP_CLIENT_ID }}",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("release workflow should not contain %q", forbidden)
		}
	}
}

// The workflow invokes the release scripts directly, so they have to stay
// executable in git.
func TestReleaseScriptsAreExecutable(t *testing.T) {
	scripts, err := filepath.Glob(filepath.Join("scripts", "release", "*.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if len(scripts) == 0 {
		t.Fatal("no scripts/release/*.sh files found")
	}
	for _, script := range scripts {
		info, err := os.Stat(script)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm()&0o111 == 0 {
			t.Errorf("%s is not executable (mode %v)", script, info.Mode().Perm())
		}
	}
}

// update-readme-release-links.sh owns everything between the markers, and its
// awk pass fails unless there is exactly one complete pair.
func TestReadmeHasOneReleaseArtifactMarkerPair(t *testing.T) {
	text := readFile(t, "README.md")
	for marker, want := range map[string]int{
		"<!-- release-artifacts:start -->": 1,
		"<!-- release-artifacts:end -->":   1,
	} {
		if got := strings.Count(text, marker); got != want {
			t.Errorf("README has %d %q markers, want %d", got, marker, want)
		}
	}
}

func TestUpdateReadmeReleaseLinks(t *testing.T) {
	readme := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(readme, []byte(readFile(t, "README.md")), 0o644); err != nil {
		t.Fatal(err)
	}

	run := func(version string) string {
		t.Helper()
		cmd := exec.Command("scripts/release/update-readme-release-links.sh", readme)
		cmd.Env = append(os.Environ(), "VERSION="+version)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("update-readme-release-links.sh %s: %v\n%s", version, err, out)
		}
		return readFile(t, readme)
	}

	first := run("v1.2.3")
	if !strings.Contains(first, "md_v1.2.3_darwin_arm64.pkg") {
		t.Error("rendered block is missing the macOS pkg link")
	}
	if !strings.Contains(first, "md_1.2.3_amd64.deb") {
		t.Error("rendered block is missing the amd64 deb link, which drops the leading v")
	}

	// Re-rendering replaces the previous block rather than stacking a second
	// one, so releases stay idempotent.
	if again := run("v1.2.3"); again != first {
		t.Error("re-rendering the same version changed the README")
	}
	second := run("v4.5.6")
	if strings.Contains(second, "1.2.3") {
		t.Error("README still references v1.2.3 after rendering v4.5.6")
	}
	if !strings.Contains(second, "md_v4.5.6_linux_arm64.tar.gz") {
		t.Error("rendered block is missing the linux arm64 tarball link")
	}

	if err := exec.Command("scripts/release/update-readme-release-links.sh", readme).Run(); err == nil {
		t.Error("script should fail without VERSION")
	}
	cmd := exec.Command("scripts/release/update-readme-release-links.sh", readme)
	cmd.Env = append(os.Environ(), "VERSION=1.2.3")
	if err := cmd.Run(); err == nil {
		t.Error("script should reject a VERSION without the v prefix")
	}
}

// checksums.sh writes its temp file inside dist, and the shell creates that
// file before find walks the directory: it must not check itself in.
func TestChecksumsSkipsItsOwnFiles(t *testing.T) {
	dist := t.TempDir()
	for _, name := range []string{"md_v1.2.3_linux_amd64.tar.gz", "md_1.2.3_amd64.deb"} {
		if err := os.WriteFile(filepath.Join(dist, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	cmd := exec.Command("scripts/release/checksums.sh")
	cmd.Env = append(os.Environ(), "DIST_DIR="+dist)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("checksums.sh: %v\n%s", err, out)
	}

	var names []string
	for line := range strings.SplitSeq(strings.TrimSpace(readFile(t, filepath.Join(dist, "checksums.txt"))), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("unexpected checksum line %q", line)
		}
		names = append(names, fields[1])
	}
	want := []string{"md_1.2.3_amd64.deb", "md_v1.2.3_linux_amd64.tar.gz"}
	if strings.Join(names, " ") != strings.Join(want, " ") {
		t.Errorf("checksums.txt lists %v, want %v", names, want)
	}

	if entries, err := os.ReadDir(dist); err != nil {
		t.Fatal(err)
	} else if len(entries) != 3 {
		t.Errorf("checksums.sh left %d files in dist, want 3", len(entries))
	}
}

// Every asset the README links has to be an asset the workflow actually
// verifies it built, so packaging renames cannot silently produce dead links.
func TestReadmeLinksMatchWorkflowArtifacts(t *testing.T) {
	readme := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(readme, []byte(readFile(t, "README.md")), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("scripts/release/update-readme-release-links.sh", readme)
	cmd.Env = append(os.Environ(), "VERSION=v0.0.0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("update-readme-release-links.sh: %v\n%s", err, out)
	}

	workflow := readFile(t, releaseWorkflow)
	assets := regexp.MustCompile(`releases/download/v0\.0\.0/([^)\s]+)`)
	matches := assets.FindAllStringSubmatch(readFile(t, readme), -1)
	if len(matches) == 0 {
		t.Fatal("no release asset links found in the rendered README")
	}
	for _, match := range matches {
		asset := match[1]
		// The workflow spells asset names with the shell variables the release
		// version is carried in: ${VERSION} keeps the v, ${pkg_version} drops it.
		pattern := strings.ReplaceAll(asset, "v0.0.0", "${VERSION}")
		pattern = strings.ReplaceAll(pattern, "0.0.0", "${pkg_version}")
		if !strings.Contains(workflow, "dist/"+pattern) {
			t.Errorf("README links %s but the workflow never verifies dist/%s", asset, pattern)
		}
	}
}

// The merge writes an entire bottle block, so splitting the JSON files
// across invocations loses earlier platform entries. Exercise the shared
// publisher with a stub that rejects anything but one complete invocation.
func TestMergeHomebrewBottles(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq is required by the Homebrew publisher")
	}
	for _, mode := range []string{"complete", "missing-input", "duplicate-input", "lost-platform", "wrong-digest"} {
		t.Run(mode, func(t *testing.T) {
			tmp := t.TempDir()
			bottles, tap := filepath.Join(tmp, "metadata files"), filepath.Join(tmp, "tap")
			for _, dir := range []string{bottles, filepath.Join(tap, "Formula")} {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			write := func(path, data string, perm os.FileMode) {
				t.Helper()
				if err := os.WriteFile(path, []byte(data), perm); err != nil {
					t.Fatal(err)
				}
			}

			tags := []string{"arm64_linux", "arm64_tahoe", "tahoe", "x86_64_linux"}
			rows := []string{"  bottle do"}
			for i, tag := range tags {
				sum := strings.Repeat(string(rune('a'+i)), 64)
				inputTag := tag
				if mode == "duplicate-input" && i == 1 {
					inputTag = tags[0]
				}
				if mode != "missing-input" || i != 0 {
					write(filepath.Join(bottles, tag+".json"),
						fmt.Sprintf(`{"clarifiedlabs/tap/md":{"bottle":{"tags":{"%s":{"sha256":"%s"}}}}}`, inputTag, sum), 0o644)
				}
				if mode == "lost-platform" && i == 0 {
					continue
				}
				if mode == "wrong-digest" && i == 0 {
					sum = strings.Repeat("0", 64)
				}
				rows = append(rows, fmt.Sprintf("    sha256 cellar: :any_skip_relocation, %s: %q", tag, sum))
			}
			rows = append(rows, "  end", "")
			write(filepath.Join(tmp, "expected.rb"), strings.Join(rows, "\n"), 0o644)
			write(filepath.Join(tmp, "brew"), `#!/usr/bin/env bash
set -euo pipefail
[[ "$PWD" == "$TAP_DIR" && "$#" -eq 8 && "$1 $2 $3 $4" == "bottle --merge --write --no-commit" ]]
touch "$TEST_DIR/called-brew"
cp "$TEST_DIR/expected.rb" Formula/md.rb
`, 0o755)
			cmd := exec.Command("scripts/release/merge-homebrew-bottles.sh", bottles)
			cmd.Env = append(os.Environ(), "TAP_DIR="+tap, "TEST_DIR="+tmp, "PATH="+tmp+":"+os.Getenv("PATH"))
			out, err := cmd.CombinedOutput()
			switch mode {
			case "complete":
				if err != nil {
					t.Fatalf("merge failed: %v\n%s", err, out)
				}
			case "missing-input", "duplicate-input":
				if err == nil || !strings.Contains(string(out), "each expected platform exactly once") {
					t.Fatalf("want input coverage failure, got %v\n%s", err, out)
				}
				if _, err := os.Stat(filepath.Join(tmp, "called-brew")); !os.IsNotExist(err) {
					t.Fatalf("brew should not run on incomplete inputs, stat error: %v", err)
				}
			default:
				if err == nil || !strings.Contains(string(out), "does not match the expected platform checksums") {
					t.Fatalf("want output verification failure, got %v\n%s", err, out)
				}
			}
		})
	}
}

func TestReleaseWorkflowSharesHomebrewBottleMerge(t *testing.T) {
	workflow := readFile(t, releaseWorkflow)
	const call = `scripts/release/merge-homebrew-bottles.sh "${GITHUB_WORKSPACE}/bottles"`
	if got := strings.Count(workflow, call); got != 2 {
		t.Fatalf("release and dry run must both invoke the shared bottle merge; got %d calls", got)
	}
}
