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
# Go 1.24 moved wasm_exec.js from misc/wasm to lib/wasm (go.dev/doc/go1.24#wasm).
# Check both locations so this works on Go 1.22 (our go.mod minimum) through
# whatever is currently installed, without pinning to one Go version's layout.
GOROOT="$(go env GOROOT)"
if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
    WASM_EXEC="$GOROOT/lib/wasm/wasm_exec.js"
elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
    WASM_EXEC="$GOROOT/misc/wasm/wasm_exec.js"
else
    echo "error: could not find wasm_exec.js under \$GOROOT/lib/wasm or \$GOROOT/misc/wasm ($GOROOT)" >&2
    exit 1
fi
cp "$WASM_EXEC" static/wasm_exec.js

echo "Done: web/static/nesmath.wasm ($(du -h static/nesmath.wasm | cut -f1))"

if [ "${1:-}" = "serve" ]; then
    echo "Serving web/static on http://localhost:8080 ..."
    cd static
    python3 -m http.server 8080
fi
