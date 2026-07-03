# Design: Expanded Interactive WebAssembly Exercises

## Context

`nesmath`'s `web/` directory already contains a working interactive exercise page: a Go program compiled to WebAssembly (`web/wasm/main.go`, standard toolchain, `//go:build js && wasm`, no TinyGo, no bundler) exporting five functions via `syscall/js`, driven by a plain HTML/CSS/JS frontend (`web/static/`) with no framework. It covers ADC, `Q4_4.Split`, accumulator oscillation, `HorizontalMotion.Step`, and `VerticalMotion.Step`.

This spec expands that to the full set of exercises originally proposed for the "learn `nesmath` progressively" curriculum, for a single learner (not a public course): someone who already knows binary/hex and general programming, but wants to internalize the 6502-specific carry-chain architecture itself - the parts this whole project's session repeatedly found subtle and easy to get wrong (the `Dummy`/`MoveForce` ordering, the two-condition clamp, sign extension, etc.).

## Goals

- Cover the remaining curriculum topics that map naturally onto "tinker with inputs, observe visualized output" - the same interaction style the existing five exercises already use.
- Keep the architecture identical: real `nesmath` Go code compiled to WASM, zero reimplementation in JavaScript, zero frontend framework, zero new external dependencies.
- Every new exported WASM function is a thin pass-through to existing, already-tested `nesmath` functions - no new logic lives in `web/wasm/main.go` beyond marshalling.

## Non-goals (explicitly deferred)

- **Predict-then-reveal mode** and **Bug Hunt mode** (reproducing this session's actual historical bugs as "spot what's wrong" challenges) - a real, considered alternative (Approach B/C during brainstorming), deliberately deferred in favor of breadth first. Worth revisiting later.
- Any exercise that requires the learner to write/submit their own code (e.g. "implement your own `isNegative` via bit masking," "implement `LSR16`") - these remain as suggested paper/editor exercises in prose, not WASM widgets, since there's no in-browser Go compiler here.
- Public distribution concerns (hosting, SEO, accessibility audits, analytics) - this is explicitly a personal tool, not a public course.

## The 10 exercises

Numbering reflects a suggested reading order in `index.html`, not necessarily new HTML element IDs (existing sections keep their current IDs; new ones get new IDs following the same `exercise-<name>` convention).

### 1. ADC *(existing, unchanged)*

### 2. SBC *(new)*

Mirrors exercise 1's layout exactly (bit-grids for `a`, `b`, result) but calls `nesmath.SBC`. The result line explicitly states whether a borrow occurred, using the same "carry=1 means NO borrow" framing already established in `carry.go`'s doc comment, so the exercise reinforces the exact wording the learner will see if they read the source.

**WASM export:** `runSBC(a, b, c uint8/Carry) -> {result, carry}` - structurally identical to `runADC`.

### 3. 16-bit chained addition *(new)*

Four byte inputs: `aLow, aHigh, bLow, bHigh`. Runs two `ADC` calls (low bytes with carry-in 0, then high bytes with the carry the first call produced) and visualizes both as a pair of connected bit-grids with an arrow showing the carry flowing from the low-byte result into the high-byte addition. This is the general multi-byte pattern, deliberately shown before the NES-specific `Position16`/`Accumulator8` wrappers that build on it - the point is seeing the *raw* mechanism once, unwrapped.

**WASM export:** `runChainedAdd16(aLow, aHigh, bLow, bHigh uint8) -> {lowResult, lowCarry, highResult, highCarry, combined16}`.

### 4. Two's complement / `Negate` *(new)*

One byte input. Shows the bit-flip (one's complement) as an intermediate step, then `+1` to reach `Negate`'s actual result, with the bits visually highlighted. Includes a note (not a separate control) that entering `0x80` demonstrates the wrap-to-itself edge case, matching `TestNegate`'s "minimum signed value wraps to itself" case.

**WASM export:** `runNegate(v uint8) -> {onesComplement, result}`.

### 5. `Q4_4` Split *(existing, unchanged)*

### 6. Accumulator oscillation *(existing, unchanged)*

### 7. Standalone `Position16` movement *(new)*

Directly demonstrates `doc.go`'s "Beyond a single game's design" section: no `Q4_4`, no `Accumulator8` - just a raw signed velocity byte added to a `Position16` every frame via `AddSigned`. Same canvas style as the existing horizontal-motion track. The point is showing the primitives work standalone for games/objects that don't need this specific sub-pixel scheme.

**WASM export:** `runSimpleMovement(velocity int8, startPage, startPixel uint8, frameCount int) -> []{frame, page, pixel}`.

### 8. Horizontal motion *(existing, enhanced)*

Add a "compare walk vs run" checkbox. When checked, run two independent `HorizontalMotion` traces (fixed at `$10` and `$19`) side by side on two stacked tracks instead of one, using the existing per-frame animation loop duplicated per track. No change to the underlying WASM export - both traces already come from two separate calls to the existing `runHorizontalTrace`.

### 9. Vertical motion / jump arc *(existing, enhanced)*

Expose `upForce` as a new numeric input (currently hardcoded to `0` in `runVerticalTrace`). Update the WASM export's signature to accept it and pass it through to `VerticalMotion.Step`'s third parameter instead of hardcoding `0`. Existing behavior when the field is left at `0` is unchanged.

**WASM export change:** `runVerticalTrace(speed int8, force uint8, maxSpeed int8, upForce uint8, frameCount int)` (adds the `upForce` parameter before `frameCount`).

### 10. Capstone: combined jump trajectory *(new)*

Drives a `HorizontalMotion` and a `VerticalMotion` together, frame by frame, and plots the result as a single X-Y curve (style C from the visual review: X = cumulative horizontal displacement, Y = cumulative vertical displacement, height inverted so "up" on screen means "up" in-game). This is the only exercise that composes two motion types in one simulation, serving as the capstone that ties the whole set together.

**WASM export:** `runJumpTrajectory(hSpeed uint8, vSpeed int8, vForce uint8, vMaxSpeed int8, frameCount int) -> []{frame, x, y}` where `x`/`y` are cumulative displacement (mirroring how the existing vertical exercise already computes cumulative displacement in `script.js`, just now also tracking the horizontal axis).

## Architecture notes (unchanged from the existing 5)

- `web/wasm/main.go` keeps its `//go:build js && wasm` constraint; new functions are added alongside the existing five in the same file, following the same "thin pass-through, no logic beyond marshalling" pattern already established.
- `web/static/index.html` gains five new `<section class="exercise">` blocks in the order above; existing five are reordered slightly (SBC/16-bit/Negate inserted before the existing Q4_4 Split section, Position16 inserted before Horizontal Motion) to match the numbering above.
- `web/static/script.js` gains one new render function per new exercise, following the existing patterns (bit-grid rendering reused from exercises 1/2, canvas line/track drawing reused from exercises 3/4/5 as applicable).
- `web/static/style.css` - no new classes anticipated; existing `.bit-grid`, `.nybble-grid`, canvas, and `.controls` styles cover all ten exercises.
- No changes to the core `nesmath` package - every new WASM export calls existing, already-tested functions. If `runVerticalTrace`'s signature change breaks anything, it's contained entirely within `web/wasm/main.go` and `script.js` (both outside the module's test suite, per the existing build-tag isolation).

## Testing / verification
yes
- `gofmt`, `go vet`, `go build` (native) unaffected, as with the current five - build-tag isolation is unchanged.
- `GOOS=js GOARCH=wasm go build ./web/wasm/` must succeed.
- Manual verification via the `browser-use` subagent, same as the original five: click through all ten exercises, confirm expected numeric output (hand-derivable from already-tested `nesmath` functions) and confirm zero console errors.
- No new Go unit tests are needed in the core module - all underlying logic (`ADC`, `SBC`, `Negate`, `Position16.AddSigned`, `HorizontalMotion.Step`, `VerticalMotion.Step`) is already covered at 100% in the existing test suite; the WASM layer only marshals inputs/outputs.

## Open questions / risks

- The capstone's `x`/`y` cumulative displacement, drawn as a single curve, will only look like a recognizable "arc" for certain input combinations (e.g. horizontal speed nonzero, vertical speed starting negative with gravity pulling it positive). Inputs that don't produce a visually interesting arc are just less didactic, not broken - no validation/clamping of "sensible" inputs is planned.
- Inserting new sections before existing ones renumbers the visible exercise headings (`index.html`'s "1.", "2." prefixes) - existing exercises 2-5 become 5, 6, 8, 9 in the new order. This is a cosmetic renumbering, not a breaking change to any function or test.
