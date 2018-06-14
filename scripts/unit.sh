#!/usr/bin/env bash
set -euo pipefail

export ROOT="$( dirname "${BASH_SOURCE[0]}" )/.."
$ROOT/scripts/install_tools.sh

cd $ROOT/src/r/
ginkgo -r -skipPackage=brats,integration
