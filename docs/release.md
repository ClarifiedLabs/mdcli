# Release

Release builds produce a single binary, `md`, which reports its version:

```sh
md --version
```

Release builds inject the repository tag (`v*`), the commit, and the build date
into that output through `internal/buildinfo`. Development builds report
`md dev`.

## Tagging

Create release tags with:

```sh
make release VERSION=patch
make release VERSION=minor
make release VERSION=major
make release VERSION=1.2.3
make release VERSION=patch AUTOPUSH=1
```

`patch`, `minor`, and `major` are computed from the latest `vX.Y.Z` git tag.
`patch` starts at `v0.0.1` when no prior tag exists. The target requires a clean
worktree, runs `go build ./...`, `go vet ./...`, and `go test ./...`, then
creates an annotated `vX.Y.Z` tag. `AUTOPUSH=1` pushes the tag to `origin`,
which is what starts the release workflow.

## Artifacts

Pushing a `v*` tag runs `.github/workflows/release.yml`. The workflow builds:

- macOS arm64 and amd64 on `macos-26`
- Linux amd64
- Linux arm64

It publishes tarballs, `.deb` and `.rpm` packages, a signed and notarized macOS
`.pkg` for Apple silicon, Homebrew bottles for macOS arm64, macOS Intel, Linux
amd64, and Linux arm64, SHA-256 checksums, and GitHub artifact attestations.
The workflow then updates `ClarifiedLabs/homebrew-tap` through a GitHub App
installation token, and adds the Linux packages to the org-wide package
repositories in `ClarifiedLabs/linux-packages`. After publishing the release
assets it also updates the generated release-artifact block in `README.md` on
the default branch and commits the versioned package links. Rerunning an older
release does not replace links for a newer latest release.

Asset names for version `vX.Y.Z`:

| Asset | Name |
|---|---|
| Tarball | `md_vX.Y.Z_{linux,darwin}_{amd64,arm64}.tar.gz` |
| Debian | `md_X.Y.Z_{amd64,arm64}.deb` |
| RPM | `md-X.Y.Z-1.{x86_64,aarch64}.rpm` |
| macOS installer | `md_vX.Y.Z_darwin_arm64.pkg` |
| Checksums | `checksums.txt` |

The tap repository must already exist with an initialized default branch. No
formula file is required ahead of time; the release workflow writes
`Formula/md.rb` and merges the generated bottle metadata.

### Linux package repositories

After the GitHub release is published, `packages-publish` adds the `md` `.deb`
and `.rpm` packages to the shared Clarified Labs repositories hosted in
[`ClarifiedLabs/linux-packages`](https://github.com/ClarifiedLabs/linux-packages)
and served by GitHub Pages at
`https://clarifiedlabs.github.io/linux-packages`:

- The APT repository (reprepro, `stable`/`main`, amd64 and arm64) files `md`
under `deb/pool/main/m/md/` automatically; no distributions change is needed.
- The RPM repository (createrepo_c) serves `rpm/x86_64/md-*.rpm` and
`rpm/aarch64/md-*.rpm`.

The repositories are org-wide: the same signing key (`Clarified Labs, Inc.
Packages <hello@clarified.io>`, published as
`clarifiedlabs-archive-keyring.asc`), the same GitHub App, and the same update
scripts serve `harness` and `md` packages, so both products install side by
side from one repository entry (`clarifiedlabs.list` / `clarifiedlabs.repo`).
The update scripts live in `ClarifiedLabs/linux-packages`
(`scripts/apt-repo-update.sh`, `scripts/rpm-repo-update.sh`); this workflow
checks that repository out and runs them from there. Versions are independent
per product, so `md` and `harness` releases never collide.

RPM payload signing happens here, in `build-linux`, before attestation and
upload: the GitHub release assets, `checksums.txt`, attestations, and the
package repositories all serve identical signed files. RPMs are signed on
every run, including dry runs.

## CI Dry Runs

Push a branch named `release-ci` or under `release-ci/`, or run the `release`
workflow manually, to exercise the release workflow without publishing. Dry-run
builds use version `v0.0.0` and the pushed commit archive as the Homebrew
source. They build and upload the normal workflow artifacts, generate
checksums, build Homebrew bottles from a local tap, and dry-run the Homebrew
formula merge.

Dry runs do not publish a GitHub release, push to the Homebrew tap, push to
the package repositories, or create artifact attestations. The macOS `.pkg`
is built unsigned in dry runs so Apple Developer ID and notarization secrets
are only required for real `v*` tag releases. They render the README
release-artifact block with `v0.0.0` and show its diff without committing it.

The `packages-publish-dry-run` job exercises the package repository pipeline
end to end with the production key: `build-linux` signs the dry-run RPMs with
`PACKAGES_GPG_PRIVATE_KEY`, the job verifies them with `rpm -K` against the
published `clarifiedlabs-archive-keyring.asc`, then builds scratch APT and RPM
repositories with the same update scripts from the `linux-packages` checkout
and verifies the `InRelease` and `repomd.xml` signatures and the expected `md`
entries for both architectures. Nothing is pushed.

The dry run also mints a GitHub App token with the same inputs as the real
`packages-publish` job and immediately discards it. Together with the GPG
import/sign/verify steps this validates all three package secrets —
`PACKAGES_GPG_PRIVATE_KEY`, `PACKAGES_APP_PRIVATE_KEY`, and
`PACKAGES_APP_CLIENT_ID` — on every `release-ci` run, so a typo'd or missing
secret fails fast instead of at the first tag. The token is never used to push
anything.

## Required Secrets

- `MACOS_DEVELOPER_ID_APPLICATION_P12_BASE64`: base64 of a `.p12` exported from
  Certificates, Identifiers & Profiles -> Certificates -> **Developer ID
  Application**. Export it with the private key from Keychain Access.
- `MACOS_DEVELOPER_ID_APPLICATION_PASSWORD`: password used when exporting that
  Application `.p12`.
- `MACOS_DEVELOPER_ID_INSTALLER_P12_BASE64`: base64 of a `.p12` exported from
  Certificates, Identifiers & Profiles -> Certificates -> **Developer ID
  Installer**.
- `MACOS_DEVELOPER_ID_INSTALLER_PASSWORD`: password used when exporting that
  Installer `.p12`.
- `APPLE_TEAM_ID`: the Apple Developer Team ID, visible in the developer account
  membership page and in Developer ID certificate subjects.
- `APPLE_NOTARY_KEY_ID`, `APPLE_NOTARY_ISSUER_ID`,
  `APPLE_NOTARY_KEY_P8_BASE64`: an **App Store Connect API key** for
  notarization. This is created in App Store Connect under Users and Access ->
  Integrations -> App Store Connect API, not in the
  Certificates/Identifiers/Profiles certificate list. Download the `.p8` key
  once and base64 it for the secret.

  Signing identities are discovered from the imported certificates by team ID.
  `package-pkg.sh` accepts `MACOS_DEVELOPER_ID_APPLICATION_IDENTITY` and
  `MACOS_DEVELOPER_ID_INSTALLER_IDENTITY` as local overrides for the exact
  certificate common names, but the workflow does not pass them.
- `HOMEBREW_TAP_APP_PRIVATE_KEY`: private key for the GitHub App installed on
  `ClarifiedLabs/homebrew-tap`.
- `HOMEBREW_TAP_APP_CLIENT_ID`: the GitHub App Client ID.
- `PACKAGES_GPG_PRIVATE_KEY`: ASCII-armored private half of the org-wide
  `Clarified Labs, Inc. Packages <hello@clarified.io>` GPG key — the same key
  `ClarifiedLabs/harness` uses. Signs the RPM packages and (on publish) the
  APT `InRelease`/`Release` metadata and `repomd.xml` in
  `ClarifiedLabs/linux-packages`.
- `PACKAGES_APP_PRIVATE_KEY`: private key for the GitHub App installed on
  `ClarifiedLabs/linux-packages` — the same App `ClarifiedLabs/harness` uses.
- `PACKAGES_APP_CLIENT_ID`: that GitHub App's Client ID.

The Homebrew App only needs to be installed on `ClarifiedLabs/homebrew-tap`
with repository Contents read/write permission; the packages App only needs to
be installed on `ClarifiedLabs/linux-packages` with repository Contents
read/write permission.
