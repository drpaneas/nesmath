# nesmath interactive exercises

The real `nesmath` Go library, compiled to WebAssembly with the standard Go toolchain (no TinyGo, no bundler, no JS framework), driving ten interactive exercises in the browser.

## Run it

```sh
./build.sh serve
```

Then open <http://localhost:8080>. This builds `static/nesmath.wasm`, copies `wasm_exec.js` from your local Go installation, and serves `static/` with Python's built-in HTTP server.

To just build without serving:

```sh
./build.sh
```

## How it's wired

- `wasm/main.go` - a `//go:build js && wasm` command that exports ten plain functions on the JS global object via `syscall/js`: `nesmathADC`, `nesmathSBC`, `nesmathChainedAdd16`, `nesmathNegate`, `nesmathSplit`, `nesmathAccumulatorTrace`, `nesmathSimpleMovement`, `nesmathHorizontalTrace`, `nesmathVerticalTrace`, `nesmathJumpTrajectory`. Each one calls straight into the real `nesmath` package - nothing is reimplemented in JavaScript.
- `static/script.js` - loads the WASM module and wires HTML inputs/buttons to those ten functions, rendering results as bit-grids, bar charts, and line/trajectory graphs on `<canvas>`.
- `static/index.html` / `static/style.css` - the page itself, one section per exercise.

Because `wasm/main.go` carries the `js && wasm` build constraint, it never compiles as part of `go build ./...` or `go test ./...` for the main module - the library and its test suite are completely unaffected by this directory.

## Deploying to GitHub Pages

`.github/workflows/deploy-pages.yml` builds `nesmath.wasm` and `wasm_exec.js` (both gitignored - they're regenerated, not committed) and deploys `static/` on every push to `main` that touches `web/**` or the Go module. It requires GitHub Pages to be enabled on the repository with the source set to "GitHub Actions" (Settings -> Pages -> Build and deployment -> Source).

`static/.nojekyll` is committed directly so GitHub Pages never routes the site through its default Jekyll processing pipeline - unnecessary for a plain static site and a common source of subtle asset-handling surprises.

`script.js` also falls back to `WebAssembly.instantiate` on a raw `ArrayBuffer` if `instantiateStreaming` fails, so the page still loads even on a host that doesn't serve `.wasm` with the `application/wasm` Content-Type `instantiateStreaming` requires (GitHub Pages itself sets this correctly).
