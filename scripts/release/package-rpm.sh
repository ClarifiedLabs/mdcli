#!/usr/bin/env bash
set -euo pipefail

: "${VERSION:?VERSION is required}"
: "${GOARCH:?GOARCH is required}"

repo_root="$(pwd -P)"

abs_path() {
	case "$1" in
		/*) printf '%s\n' "$1" ;;
		*) printf '%s/%s\n' "$repo_root" "$1" ;;
	esac
}

stage_dir="$(abs_path "${STAGE_DIR:-dist/bin/linux-${GOARCH}}")"
dist_dir="$(abs_path "${DIST_DIR:-dist}")"
version="${VERSION#v}"

case "$GOARCH" in
	amd64) rpm_arch="x86_64" ;;
	arm64) rpm_arch="aarch64" ;;
	*) echo "unsupported GOARCH for rpm: ${GOARCH}" >&2; exit 2 ;;
esac

work_dir="$(abs_path "${WORK_DIR:-dist/package-rpm}")"
readme_path="$(abs_path README.md)"
license_path="$(abs_path LICENSE)"
rm -rf "$work_dir"
mkdir -p "$dist_dir"

name="md"
binary="md"
summary="Terminal Markdown viewer with ASCII Mermaid diagrams"
description="md renders Markdown in the terminal with styling, syntax highlighting for fenced code blocks, and draws mermaid fences as ASCII diagrams."
topdir="${work_dir}/${name}/rpmbuild"
spec="${topdir}/SPECS/${name}.spec"

mkdir -p "${topdir}/BUILD" "${topdir}/BUILDROOT" "${topdir}/RPMS" "${topdir}/SOURCES" "${topdir}/SPECS"
cat >"$spec" <<SPEC
Name: ${name}
Version: ${version}
Release: 1%{?dist}
Summary: ${summary}
License: MIT
URL: https://github.com/ClarifiedLabs/mdcli
BuildArch: ${rpm_arch}
AutoReqProv: no

%description
${description}

%prep

%build

%install
mkdir -p %{buildroot}/usr/bin
mkdir -p %{buildroot}/usr/share/doc/${name}
mkdir -p %{buildroot}/usr/share/licenses/${name}
install -m 0755 "${stage_dir}/${binary}" %{buildroot}/usr/bin/${binary}
install -m 0644 "${readme_path}" %{buildroot}/usr/share/doc/${name}/README.md
install -m 0644 "${license_path}" %{buildroot}/usr/share/licenses/${name}/LICENSE

%files
/usr/bin/${binary}
%doc /usr/share/doc/${name}/README.md
%license /usr/share/licenses/${name}/LICENSE
SPEC

rpmbuild --define "_topdir ${topdir}" --target "${rpm_arch}" -bb "$spec"
rpm_path="$(find "${topdir}/RPMS" -type f -name '*.rpm' -print | sort | awk 'END {print}')"
if [[ -z "$rpm_path" ]]; then
	echo "rpmbuild did not produce an rpm for ${name}" >&2
	exit 1
fi
cp "$rpm_path" "${dist_dir}/${name}-${version}-1.${rpm_arch}.rpm"
