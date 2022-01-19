#!/bin/sh

set -xeu

binname="bazel-remote-cache-client"

tmpdir="$(mktemp -d --tmpdir brcc.XXXXXX)"
version="$(./workspace_status.sh | grep STABLE_VERSION | awk '{ print $2 }')"

build()
{
    local os="${1}" arch="${2}"

    bazel build \
        --stamp \
        --@io_bazel_rules_go//go/config:pure \
        --platforms "@io_bazel_rules_go//go/toolchain:${os}_${arch}" \
        "//cmd/${binname}"

    cp -av \
        "bazel-bin/cmd/${binname}/${binname}_/${binname}" \
        "${tmpdir}/${binname}-${version}-${os}-${arch}"
}

build linux amd64
build darwin amd64

tree -pugsh "$tmpdir"

exit 0
