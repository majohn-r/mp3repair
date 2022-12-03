#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
    tracing=true
else
    tracing=false
fi

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
cd "${DIR}/build"
if [[ "${tracing}" == "true" ]]; then
    go run . -v "$@"
else
    go run . "$@"
fi