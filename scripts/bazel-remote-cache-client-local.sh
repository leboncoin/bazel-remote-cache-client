#!/bin/sh

set -eu

bazel_cachedir="${BAZEL_LOCAL_CACHEDIR:-$HOME/.cache/bazel-cache}"

if [ $# -eq 0 ]; then
    echo "Error: Missing bazel cache object type to get" >&2
    exit 1
fi

object_type="$1"
shift

case "$object_type" in
    ac) ;;
    cas) ;;
    *)
        echo "Error: Object type \"$object_type\" not suppoorted" >&2
        exit 1
        ;;
esac

if [ $# -eq 0 ]; then
    echo "Error: Missing operation" >&2
    exit 1
fi

operation="$1"
shift

if [ "$operation" != "get" ]; then
    echo "Error: Operation \"$operation\" not supported" >&2
    exit 1
fi

if [ $# -eq 0 ]; then
    echo "Error: Missing digest to get" >&2
    exit 1
fi

digest="$1"
shift

case "$object_type" in
    ac)
        bazel-remote-cache-client ac get "$@" "$digest"
        ;;
    cas)
        cf="$(find "${bazel_cachedir}/cas.v2" -name "${digest}-*-*" -type f -printf "%f\n" | head -1)"

        size="${cf%-*}"
        size="${size##*-}"

        echo "$size"

        bazel-remote-cache-client cas get "$@" "$digest/$size"
        ;;
esac

exit 0
