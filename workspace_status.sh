#!/bin/sh

set -eu

echo STABLE_VERSION "$(git describe --tags --always | sed -r 's/^v//')"

exit 0
