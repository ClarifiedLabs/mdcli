#!/usr/bin/env bash
set -euo pipefail

: "${VERSION:?VERSION is required}"
: "${GOARCH:?GOARCH is required}"

stage_dir="${STAGE_DIR:-dist/bin/linux-${GOARCH}}"
dist_dir="${DIST_DIR:-dist}"
version="${VERSION#v}"

case "$GOARCH" in
	amd64) deb_arch="amd64" ;;
	arm64) deb_arch="arm64" ;;
	*) echo "unsupported GOARCH for deb: ${GOARCH}" >&2; exit 2 ;;
esac

mkdir -p "$dist_dir"

name="md"
binary="md"
summary="Terminal Markdown viewer with ASCII Mermaid diagrams"
description="md renders Markdown in the terminal with styling, syntax highlighting for fenced code blocks, and draws mermaid fences as ASCII diagrams."
pkgroot="${WORK_DIR:-dist/package-deb}/${name}_${version}_${deb_arch}"

rm -rf "$pkgroot"
mkdir -p "${pkgroot}/DEBIAN" "${pkgroot}/usr/bin" "${pkgroot}/usr/share/doc/${name}"

install -m 0755 "${stage_dir}/${binary}" "${pkgroot}/usr/bin/${binary}"
install -m 0644 README.md "${pkgroot}/usr/share/doc/${name}/README.md"
install -m 0644 LICENSE "${pkgroot}/usr/share/doc/${name}/LICENSE"

installed_size="$(du -sk "${pkgroot}/usr" | awk '{print $1}')"
cat >"${pkgroot}/DEBIAN/control" <<CONTROL
Package: ${name}
Version: ${version}
Section: utils
Priority: optional
Architecture: ${deb_arch}
Maintainer: Clarified Labs <opensource@clarifiedlabs.com>
Installed-Size: ${installed_size}
Homepage: https://github.com/ClarifiedLabs/mdcli
Description: ${summary}
 ${description}
CONTROL

dpkg-deb --build --root-owner-group "$pkgroot" "${dist_dir}/${name}_${version}_${deb_arch}.deb"
