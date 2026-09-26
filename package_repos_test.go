package mdcli_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func requirePackageRepoTools(t *testing.T, tools ...string) {
	t.Helper()
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("required tool %q not in PATH; skipping", tool)
		}
	}
}

// jobBoundaryRegexp matches the start of the next top-level job in a GitHub
// Actions workflow file.
func jobBoundaryRegexp() *regexp.Regexp {
	return regexp.MustCompile(`\n  [a-z0-9-]+:\n`)
}

// newPackageTestKey creates an isolated GNUPGHOME holding a single
// passphrase-less RSA signing key and returns an environment for commands
// that must use it.
func newPackageTestKey(t *testing.T) []string {
	t.Helper()
	gnupgHome := filepath.Join(t.TempDir(), "gnupg")
	if err := os.Mkdir(gnupgHome, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = exec.Command("gpgconf", "--homedir", gnupgHome, "--kill", "gpg-agent").Run()
	})
	env := append(os.Environ(), "GNUPGHOME="+gnupgHome)
	params := `Key-Type: RSA
Key-Length: 2048
Name-Real: mdcli Test Packages
Name-Email: mdcli-packages-test@example.invalid
Expire-Date: 0
%no-protection
%commit
`
	cmd := exec.Command("gpg", "--batch", "--gen-key")
	cmd.Env = env
	cmd.Stdin = strings.NewReader(params)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate test signing key: %v\n%s", err, output)
	}
	return env
}

// RPM payload signing happens in this repository so the GitHub release
// assets, checksums, attestations, and the package repositories serve
// byte-identical signed files.
func TestPackageRpmSigns(t *testing.T) {
	requirePackageRepoTools(t, "rpmbuild", "rpm", "gpg")
	env := newPackageTestKey(t)

	workDir := t.TempDir()
	stageDir := filepath.Join(workDir, "stage")
	if err := os.Mkdir(stageDir, 0o755); err != nil {
		t.Fatal(err)
	}
	stub := "#!/bin/sh\necho \"md v0.0.0\"\n"
	if err := os.WriteFile(filepath.Join(stageDir, "md"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}

	distDir := filepath.Join(workDir, "dist")
	build := exec.Command("bash", "scripts/release/package-rpm.sh")
	build.Dir = "."
	build.Env = append(env,
		"VERSION=v0.0.0",
		"GOARCH=amd64",
		"RPM_GPG_SIGN=1",
		"STAGE_DIR="+stageDir,
		"DIST_DIR="+distDir,
		"WORK_DIR="+filepath.Join(workDir, "rpmbuild-work"),
	)
	// package-rpm.sh resolves README.md and LICENSE relative to its working
	// directory, so run it from the repository root like the workflow does.
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("package-rpm.sh: %v\n%s", err, output)
	}

	rpmPath := filepath.Join(distDir, "md-0.0.0-1.x86_64.rpm")
	if _, err := os.Stat(rpmPath); err != nil {
		t.Fatalf("expected rpm %s: %v", rpmPath, err)
	}

	query := exec.Command("rpm", "-qpi", rpmPath)
	query.Env = env
	output, err := query.CombinedOutput()
	if err != nil {
		t.Fatalf("rpm -qpi %s: %v\n%s", rpmPath, err, output)
	}
	signature := ""
	for line := range strings.Lines(string(output)) {
		if strings.HasPrefix(line, "Signature") {
			signature = line
		}
	}
	if signature == "" || strings.Contains(signature, "(none)") {
		t.Fatalf("md rpm is not signed (Signature line %q):\n%s", signature, output)
	}

	// rpm -K against an isolated rpmdb holding only the test public key.
	exportKey := exec.Command("gpg", "--batch", "--armor", "--export")
	exportKey.Env = env
	pubKey, err := exportKey.Output()
	if err != nil {
		t.Fatalf("export test public key: %v", err)
	}
	pubKeyPath := filepath.Join(workDir, "pubkey.asc")
	if err := os.WriteFile(pubKeyPath, pubKey, 0o644); err != nil {
		t.Fatal(err)
	}
	rpmDBPath := filepath.Join(workDir, "rpmdb")
	if err := os.Mkdir(rpmDBPath, 0o755); err != nil {
		t.Fatal(err)
	}
	importKey := exec.Command("rpmkeys", "--dbpath", rpmDBPath, "--import", pubKeyPath)
	importKey.Env = env
	if output, err := importKey.CombinedOutput(); err != nil {
		t.Fatalf("rpm --import: %v\n%s", err, output)
	}
	check := exec.Command("rpmkeys", "--dbpath", rpmDBPath, "-K", rpmPath)
	check.Env = env
	output, err = check.CombinedOutput()
	if err != nil {
		t.Fatalf("rpm -K %s: %v\n%s", rpmPath, err, output)
	}
	if !strings.Contains(string(output), "signatures OK") {
		t.Fatalf("rpm -K %s did not report signatures OK:\n%s", rpmPath, output)
	}
}

// The package repository jobs check out ClarifiedLabs/linux-packages, verify
// the two signed RPMs against the org-wide keyring, and run the update
// scripts from that checkout. The old harness-specific keyring name must not
// come back.
func TestReleaseWorkflowPublishesPackageRepos(t *testing.T) {
	workflow, err := os.ReadFile(releaseWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)

	for _, want := range []string{
		"client-id: ${{ secrets.PACKAGES_APP_CLIENT_ID }}",
		"private-key: ${{ secrets.PACKAGES_APP_PRIVATE_KEY }}",
		"PACKAGES_GPG_PRIVATE_KEY: ${{ secrets.PACKAGES_GPG_PRIVATE_KEY }}",
		"repository: ClarifiedLabs/linux-packages",
		"RPM_GPG_SIGN: '1'",
		"rpm --dbpath \"$rpmdb\" --import",
		"packages-publish-dry-run:",
		"linux-packages/clarifiedlabs-archive-keyring.asc",
		"linux-packages/scripts/apt-repo-update.sh",
		"linux-packages/scripts/rpm-repo-update.sh",
		"dist/md-${pkg_version}-1.x86_64.rpm",
		"dist/md-${pkg_version}-1.aarch64.rpm",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("release workflow should contain %q", want)
		}
	}

	publishStart := strings.Index(text, "\n  packages-publish:\n")
	if publishStart < 0 {
		t.Fatal("release workflow should define a packages-publish job")
	}
	publishEnd := len(text)
	if loc := jobBoundaryRegexp().FindStringIndex(text[publishStart+1:]); loc != nil {
		publishEnd = publishStart + 1 + loc[0]
	}
	publishJob := text[publishStart:publishEnd]

	if !strings.Contains(publishJob, "if: ${{ startsWith(github.ref, 'refs/tags/v') }}") {
		t.Fatal("packages-publish job should be gated on v* tags")
	}
	if !strings.Contains(publishJob, "needs: publish") {
		t.Fatal("packages-publish job should wait for the release publish job")
	}
	if !strings.Contains(publishJob, "chore: add md ${TAG} to package repositories") {
		t.Fatal("packages-publish job should commit md package repository updates")
	}

	// The dry-run job mints (and discards) an App token so a typo'd or
	// missing PACKAGES_APP_* secret fails during release-ci, not at the first
	// tag.
	dryRunStart := strings.Index(text, "\n  packages-publish-dry-run:\n")
	if dryRunStart < 0 {
		t.Fatal("release workflow should define a packages-publish-dry-run job")
	}
	dryRunEnd := len(text)
	dryRun := text[dryRunStart:dryRunEnd]
	if !strings.Contains(dryRun, "actions/create-github-app-token") {
		t.Fatal("packages-publish-dry-run should validate the packages App secrets")
	}
	if !strings.Contains(dryRun, "PACKAGES_GPG_PRIVATE_KEY: ${{ secrets.PACKAGES_GPG_PRIVATE_KEY }}") {
		t.Fatal("packages-publish-dry-run should import the production signing key")
	}
	if !strings.Contains(dryRun, "clarifiedlabs.github.io/linux-packages/clarifiedlabs-archive-keyring.asc") {
		t.Fatal("packages-publish-dry-run should verify against the published org-wide keyring")
	}
	if !strings.Contains(dryRun, "Label: clarifiedlabs") {
		t.Fatal("packages-publish-dry-run scratch distributions should use the org-wide Label")
	}
	if !strings.Contains(dryRun, "|${arch}: md ${pkg_version}") {
		t.Fatal("packages-publish-dry-run should assert the md apt repo entries")
	}

	for _, forbidden := range []string{
		"vars.PACKAGES_APP_CLIENT_ID",
		"app-id: ${{ secrets.PACKAGES_APP_CLIENT_ID }}",
		// Dry runs sign with the production key, like tag builds.
		"RPM_GPG_SIGN: ${{ startsWith(github.ref, 'refs/tags/v') && '1' || '0' }}",
		"gpg --batch --gen-key",
		// Non-root rpm -K cannot see keys imported via sudo (per-user rpmdb).
		"sudo rpm --import",
		// Old harness-specific public names are gone org-wide.
		"harness-archive-keyring.asc",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("release workflow should not contain %q", forbidden)
		}
	}
}

// build-linux imports the signing key and signs RPMs on every run so
// release-ci exercises the production signing path, dry runs included.
func TestReleaseWorkflowSignsRpmsOnEveryRun(t *testing.T) {
	workflow, err := os.ReadFile(releaseWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	text := string(workflow)

	buildStart := strings.Index(text, "\n  build-linux:\n")
	if buildStart < 0 {
		t.Fatal("release workflow should define a build-linux job")
	}
	buildEnd := len(text)
	boundary := jobBoundaryRegexp()
	if loc := boundary.FindStringIndex(text[buildStart+1:]); loc != nil {
		buildEnd = buildStart + 1 + loc[0]
	}
	buildLinux := text[buildStart:buildEnd]

	importIdx := strings.Index(buildLinux, "- name: Import package signing key")
	packageIdx := strings.Index(buildLinux, "- name: Package archives")
	if importIdx < 0 || packageIdx < 0 || importIdx > packageIdx {
		t.Fatal("build-linux should import the package signing key before packaging")
	}
	importStep := buildLinux[importIdx:packageIdx]
	if strings.Contains(importStep, "refs/tags/v") {
		t.Fatal("build-linux signing key import should not be gated on tags; dry runs sign too")
	}
	if !strings.Contains(buildLinux, "RPM_GPG_SIGN: '1'") {
		t.Fatal("build-linux should set RPM_GPG_SIGN unconditionally")
	}
}
