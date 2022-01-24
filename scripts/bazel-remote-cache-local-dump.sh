#!/bin/sh

set -ue

bazel_cachedir="${BAZEL_LOCAL_CACHEDIR:-$HOME/.cache/bazel-cache}"

exec find "$bazel_cachedir/ac.v2" -type f -printf "%T@ %f\n" |
    sort -n |
    awk '{ print $2 }' |
    sed -r 's/-.*//' |
    xargs -r bazel-remote-cache-client "$@" ac get
