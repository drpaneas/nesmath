#!/usr/bin/env bash
# Builds the interactive exercises' WebAssembly module and copies the Go
# runtime's JS glue code alongside it. Run this from anywhere; it always
# resolves paths relative to this script.
#
# Usage:
#   ./build.sh          # build web/static/nesmath.wasm
#   ./build.sh serve    # build, then serve web/static on :8080
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

echo "Building nesmath.wasm..."
GOOS=js GOARCH=wasm go build -o static/nesmath.wasm ./wasm

echo "Copying wasm_exec.js from the Go toolchain..."
cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" static/wasm_exec.js

echo "Done: web/static/nesmath.wasm ($(du -h static/nesmath.wasm | cut -f1))"

if [ "${1:-}" = "serve" ]; then
    echo "Serving web/static on http://localhost:8080 ..."
    cd static
    python3 -m http.server 8080
fi
