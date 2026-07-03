# Interactive Exercises Expansion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Expand `web/`'s interactive WebAssembly exercise page from 5 to 10 exercises, covering the remaining curriculum topics (SBC, multi-byte carry chaining, two's complement, standalone `Position16` movement, and a combined horizontal+vertical capstone) plus two enhancements to existing exercises (walk-vs-run comparison, optional upward force).

**Architecture:** Unchanged from the existing 5 exercises - every new WASM export in `web/wasm/main.go` is a thin pass-through to already-tested `nesmath` functions, called from `web/static/script.js` (vanilla JS, no framework) via `syscall/js`.

**Tech Stack:** Go (standard toolchain, `GOOS=js GOARCH=wasm`), vanilla HTML/CSS/JS, no new dependencies.

**Scope check:** All ten exercises share the same three files (`web/wasm/main.go`, `web/static/index.html`, `web/static/script.js`) and the identical thin-wrapper pattern - this is one cohesive subsystem, not independent projects, so it stays as a single plan.

**Note on commit steps:** The standard `writing-plans` task template ends each task with a `git commit` step. This repo's workflow rule is to commit only when the user explicitly asks. Each task below ends with a **Checkpoint** (build + verify) instead of a commit - stage the changes if you like, but do not run `git commit` unless asked.

---

## Task 1: Add SBC, 16-bit chained addition, Negate, and standalone Position16 exercises

**Files:**
- Modify: `web/wasm/main.go` (add 4 functions + 4 registrations)
- Modify: `web/static/index.html` (insert 4 new `<section>` blocks, renumber 2 existing headings)
- Modify: `web/static/script.js` (add 4 wiring functions)

- [ ] **Step 1: Add the four new WASM exports to `web/wasm/main.go`**

Add these four functions anywhere after `runADC` (e.g. directly below it), and add their four `js.Global().Set(...)` registrations to `main()`:

```go
func main() {
	js.Global().Set("nesmathADC", js.FuncOf(runADC))
	js.Global().Set("nesmathSBC", js.FuncOf(runSBC))
	js.Global().Set("nesmathChainedAdd16", js.FuncOf(runChainedAdd16))
	js.Global().Set("nesmathNegate", js.FuncOf(runNegate))
	js.Global().Set("nesmathSplit", js.FuncOf(runSplit))
	js.Global().Set("nesmathAccumulatorTrace", js.FuncOf(runAccumulatorTrace))
	js.Global().Set("nesmathSimpleMovement", js.FuncOf(runSimpleMovement))
	js.Global().Set("nesmathHorizontalTrace", js.FuncOf(runHorizontalTrace))
	js.Global().Set("nesmathVerticalTrace", js.FuncOf(runVerticalTrace))

	// Keep the program alive: without this, main returns and the exported
	// functions become unreachable the moment the page calls one.
	select {}
}
```

```go
// runSBC exposes nesmath.SBC directly: three bytes in (a, b, carry-in),
// the new value of a and the carry-out come back out. SBC's carry
// convention is inverted from ADC's: carry=1 means NO borrow occurred,
// carry=0 means a borrow occurred - this exercise exists specifically to
// make that counterintuitive convention visible.
func runSBC(_ js.Value, args []js.Value) any {
	a := uint8(args[0].Int())
	b := uint8(args[1].Int())
	c := nesmath.Carry(args[2].Int())

	carry := nesmath.SBC(&a, b, c)

	return map[string]any{
		"result": int(a),
		"carry":  int(carry),
	}
}

// runChainedAdd16 adds two 16-bit numbers (each given as low/high byte
// pairs) using two chained ADC calls: the low bytes first with carry-in
// 0, then the high bytes using whatever carry the low-byte addition
// produced. This is the general multi-byte pattern underneath
// Position16.AddSigned and every other multi-byte operation in the
// library, shown here unwrapped.
func runChainedAdd16(_ js.Value, args []js.Value) any {
	aLow := uint8(args[0].Int())
	aHigh := uint8(args[1].Int())
	bLow := uint8(args[2].Int())
	bHigh := uint8(args[3].Int())

	lowResult := aLow
	lowCarry := nesmath.ADC(&lowResult, bLow, 0)

	highResult := aHigh
	highCarry := nesmath.ADC(&highResult, bHigh, lowCarry)

	combined16 := uint16(highResult)<<8 | uint16(lowResult)

	return map[string]any{
		"lowResult":  int(lowResult),
		"lowCarry":   int(lowCarry),
		"highResult": int(highResult),
		"highCarry":  int(highCarry),
		"combined16": int(combined16),
	}
}

// runNegate exposes nesmath.Negate along with the intermediate one's
// complement (bit-flip) step, so the frontend can show both stages of
// "invert all bits, then add 1." Entering 0x80 demonstrates the two's
// complement edge case where negation wraps back to itself.
func runNegate(_ js.Value, args []js.Value) any {
	v := uint8(args[0].Int())

	onesComplement := ^v
	result := nesmath.Negate(v)

	return map[string]any{
		"onesComplement": int(onesComplement),
		"result":         int(result),
	}
}

// runSimpleMovement drives Position16.AddSigned directly with a raw
// signed velocity and no accumulator at all - the "Beyond Super Mario
// Bros." pattern from doc.go, for games that don't need SMB's specific
// sub-pixel scheme.
func runSimpleMovement(_ js.Value, args []js.Value) any {
	velocity := int8(args[0].Int())
	startPage := uint8(args[1].Int())
	startPixel := uint8(args[2].Int())
	frameCount := args[3].Int()

	pos := nesmath.Position16{Page: startPage, Pixel: startPixel}

	frames := make([]any, 0, frameCount)
	for frame := 1; frame <= frameCount; frame++ {
		pos.AddSigned(velocity, 0)
		frames = append(frames, map[string]any{
			"frame": frame,
			"page":  int(pos.Page),
			"pixel": int(pos.Pixel),
		})
	}
	return frames
}
```

- [ ] **Step 2: Build and verify the WASM target compiles**

Run: `cd web && GOOS=js GOARCH=wasm go build -o /tmp/step_check.wasm ./wasm`
Expected: exits with no output (success). If it errors, the most likely cause is a typo in one of the four new functions or a missing registration - compare against the code above exactly.

- [ ] **Step 3: Insert the four new sections into `web/static/index.html`**

Neither the existing `exercise-adc` nor `exercise-split` sections are deleted or rewritten in this step - insert the following three new sections as new content between them, directly after the ADC section's closing `</section>` tag and directly before the existing `<section class="exercise" id="exercise-split">` opening tag:

```html
  <section class="exercise" id="exercise-sbc">
    <h2>2. SBC: subtract-with-borrow</h2>
    <p>The 6502's <code>SBC</code> instruction has one famously counterintuitive rule: the carry flag means the <strong>opposite</strong> of what you'd guess. <code>carry=1</code> means no borrow happened; <code>carry=0</code> means a borrow occurred. Try 5-3 with carry=1 first (clean subtraction), then 3-5 with carry=1 (watch it borrow).</p>
    <div class="controls">
      <label>a <input type="number" id="sbc-a" min="0" max="255" value="5"></label>
      <label>b <input type="number" id="sbc-b" min="0" max="255" value="3"></label>
      <label>carry in <select id="sbc-c"><option value="1">1</option><option value="0">0</option></select></label>
      <button onclick="runSBC()">Run SBC</button>
    </div>
    <div class="output">
      <div class="bit-grid" id="sbc-grid"></div>
      <p id="sbc-result" class="result-line">&nbsp;</p>
    </div>
    <p class="hint">Try a=3, b=5, carry=1: a is less than b, so this borrows - carry out becomes 0, not 1.</p>
  </section>

  <section class="exercise" id="exercise-chain16">
    <h2>3. Chaining ADC across two bytes</h2>
    <p>The 6502 has no 16-bit registers, so a 16-bit add is two 8-bit <code>ADC</code> calls: add the low bytes first, then add the high bytes <strong>using the carry the low-byte addition produced</strong> - no <code>CLC</code> in between. This is the raw mechanism behind every multi-byte operation in this library, including <code>Position16.AddSigned</code>.</p>
    <div class="controls">
      <label>a low <input type="number" id="chain-alow" min="0" max="255" value="255"></label>
      <label>a high <input type="number" id="chain-ahigh" min="0" max="255" value="0"></label>
      <label>b low <input type="number" id="chain-blow" min="0" max="255" value="1"></label>
      <label>b high <input type="number" id="chain-bhigh" min="0" max="255" value="0"></label>
      <button onclick="runChainedAdd16()">Run</button>
    </div>
    <div class="output">
      <div class="bit-grid" id="chain-grid"></div>
      <p id="chain-result" class="result-line">&nbsp;</p>
    </div>
    <p class="hint">Try a=255 (low=255,high=0) + b=1 (low=1,high=0): the low byte overflows to 0 with carry=1, and that carry becomes the +1 that increments the high byte - producing 256, correctly, with two 8-bit additions.</p>
  </section>

  <section class="exercise" id="exercise-negate">
    <h2>4. Two's complement: Negate</h2>
    <p>To negate a byte on the 6502: invert every bit, then add 1. Enter a byte and watch both steps. Try 0x80 - the most negative signed byte, -128 - and see it negate back to itself, because there is no +128 to represent in a signed byte.</p>
    <div class="controls">
      <label>v <input type="number" id="negate-v" min="0" max="255" value="40"></label>
      <button onclick="runNegate()">Negate</button>
    </div>
    <div class="output">
      <div class="bit-grid" id="negate-grid"></div>
      <p id="negate-result" class="result-line">&nbsp;</p>
    </div>
    <p class="hint">40 (0x28) is Mario's max run speed, +2.5 px/frame. Negate it and get 0xD8, -2.5 px/frame (running left). Then try 128 (0x80) for the wrap-to-itself edge case.</p>
  </section>
```

Then update the two existing headings that now shift position - change `<h2>2. Q4_4: splitting a speed byte</h2>` to `<h2>5. Q4_4: splitting a speed byte</h2>`, and change `<h2>3. The sub-pixel accumulator: why movement alternates 1, 2, 1, 2</h2>` to `<h2>6. The sub-pixel accumulator: why movement alternates 1, 2, 1, 2</h2>`.

Then insert the new Position16 section immediately after the (now-numbered-6) accumulator section's `</section>`, before the horizontal motion section:

```html
  <section class="exercise" id="exercise-simple-movement">
    <h2>7. Position16 alone: movement without Q4_4</h2>
    <p>Not every NES game needs Super Mario Bros.' specific sub-pixel scheme. Some objects - a bullet, a simple enemy drifting one direction - just move a fixed whole number of pixels every frame, with carryIn always 0 and no accumulator feeding it. This is <code>Position16.AddSigned</code> used completely on its own.</p>
    <div class="controls">
      <label>velocity (px/frame) <input type="number" id="simple-velocity" min="-128" max="127" value="4"></label>
      <label>start pixel <input type="number" id="simple-pixel" min="0" max="255" value="250"></label>
      <label>frames <input type="number" id="simple-frames" min="1" max="64" value="6"></label>
      <button onclick="runSimpleMovement()">Run</button>
    </div>
    <div class="output">
      <canvas id="simple-canvas" width="640" height="100"></canvas>
      <p id="simple-result" class="result-line">&nbsp;</p>
    </div>
    <p class="hint">Compare this to exercise 8 (Horizontal motion): same page-crossing carry mechanism, but no Q4_4 split and no sub-pixel accumulator - just a raw integer step.</p>
  </section>
```

Finally, change `<h2>4. Horizontal motion: speed becomes movement</h2>` to `<h2>8. Horizontal motion: speed becomes movement</h2>`, and change `<h2>5. Vertical motion: the gravity chain and the jump arc</h2>` to `<h2>9. Vertical motion: the gravity chain and the jump arc</h2>`. (These two heading-number edits are the only change in this step to those two sections - their bodies are untouched until Tasks 2 and 3.)

- [ ] **Step 4: Add the four new wiring functions to `web/static/script.js`**

Insert immediately after the existing `// --- Exercise 1: ADC` block (before `// --- Exercise 2: Q4_4 Split`):

```js
// --- Exercise 2: SBC -------------------------------------------------

function runSBC() {
  const a = parseInt(document.getElementById("sbc-a").value, 10) & 0xff;
  const b = parseInt(document.getElementById("sbc-b").value, 10) & 0xff;
  const c = parseInt(document.getElementById("sbc-c").value, 10);

  const out = nesmathSBC(a, b, c);

  const grid = document.getElementById("sbc-grid");
  grid.innerHTML = "";
  renderByteBits(grid, "a", a, false);
  renderByteBits(grid, "b", b, false);
  renderByteBits(grid, "result", out.result, false);

  document.getElementById("sbc-result").textContent =
    `${a} - ${b} - (1 - carry(${c})) = ${out.result} (${toHex(out.result)}), carry out = ${out.carry}` +
    (out.carry ? "  <- carry=1 means NO borrow occurred" : "  <- carry=0 means a borrow occurred");
}

// --- Exercise 3: 16-bit chained addition -------------------------------

function runChainedAdd16() {
  const aLow = parseInt(document.getElementById("chain-alow").value, 10) & 0xff;
  const aHigh = parseInt(document.getElementById("chain-ahigh").value, 10) & 0xff;
  const bLow = parseInt(document.getElementById("chain-blow").value, 10) & 0xff;
  const bHigh = parseInt(document.getElementById("chain-bhigh").value, 10) & 0xff;

  const out = nesmathChainedAdd16(aLow, aHigh, bLow, bHigh);

  const grid = document.getElementById("chain-grid");
  grid.innerHTML = "";
  renderByteBits(grid, "a low", aLow, false);
  renderByteBits(grid, "b low", bLow, false);
  renderByteBits(grid, "low result", out.lowResult, false);
  renderByteBits(grid, "a high", aHigh, false);
  renderByteBits(grid, "b high", bHigh, false);
  renderByteBits(grid, "high result", out.highResult, false);

  document.getElementById("chain-result").textContent =
    `low byte: ${aLow} + ${bLow} = ${out.lowResult}, carry out = ${out.lowCarry}\n` +
    `high byte: ${aHigh} + ${bHigh} + carry(${out.lowCarry}) = ${out.highResult}, carry out = ${out.highCarry}\n` +
    `combined 16-bit value = ${out.combined16} (${toHex(out.combined16, 4)})`;
}

// --- Exercise 4: Two's complement / Negate ------------------------------

function runNegate() {
  const v = parseInt(document.getElementById("negate-v").value, 10) & 0xff;
  const out = nesmathNegate(v);

  const grid = document.getElementById("negate-grid");
  grid.innerHTML = "";
  renderByteBits(grid, "v", v, false);
  renderByteBits(grid, "one's complement (^v)", out.onesComplement, false);
  renderByteBits(grid, "Negate(v) = ^v + 1", out.result, false);

  let line = `Negate(${v}) = ${out.result} (${toHex(out.result)})`;
  if (v === 0x80) {
    line +=
      "\n\nEdge case: 0x80 is -128, the most negative signed byte. There is " +
      "no positive +128 to represent, so negating it wraps back to itself " +
      "instead of overflowing - this is a real hardware behavior, not a bug.";
  }
  document.getElementById("negate-result").textContent = line;
}
```

Insert immediately after the (renumbered) `// --- Exercise 6: Accumulator oscillation` block's closing `}` (before `// --- Exercise 8: Horizontal motion` - note the comment banners for exercises 5/6 should also have their numbers updated from "Exercise 2"/"Exercise 3" to "Exercise 5"/"Exercise 6" to stay consistent with the new numbering, though this is cosmetic and does not affect behavior):

```js
// --- Exercise 7: Position16 alone ---------------------------------------

function runSimpleMovement() {
  const velocity = parseInt(document.getElementById("simple-velocity").value, 10);
  const startPixel = parseInt(document.getElementById("simple-pixel").value, 10) & 0xff;
  const frames = parseInt(document.getElementById("simple-frames").value, 10);

  const trace = nesmathSimpleMovement(velocity, 0, startPixel, frames);

  const canvas = document.getElementById("simple-canvas");
  const ctx = canvas.getContext("2d");
  let frameIndex = 0;

  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const trackY = canvas.height / 2;

    ctx.strokeStyle = "#34344a";
    ctx.beginPath();
    ctx.moveTo(10, trackY);
    ctx.lineTo(canvas.width - 10, trackY);
    ctx.stroke();

    const f = trace[frameIndex];
    const x = 10 + (f.pixel / 255) * (canvas.width - 20);

    ctx.fillStyle = "#5fd68a";
    ctx.beginPath();
    ctx.arc(x, trackY, 8, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = "#e6e6ef";
    ctx.font = "12px monospace";
    ctx.fillText(`frame ${f.frame}  page=${f.page}  pixel=${f.pixel}`, 10, 20);

    document.getElementById("simple-result").textContent =
      `frame ${f.frame}/${trace.length}: page=${f.page} pixel=${f.pixel} - no accumulator, no Q4_4, just AddSigned(${velocity}, 0) every frame.`;

    frameIndex++;
    if (frameIndex < trace.length) {
      setTimeout(draw, 200);
    }
  }
  draw();
}
```

- [ ] **Checkpoint: rebuild and sanity-check**

Run: `cd web && ./build.sh`
Expected: `Done: web/static/nesmath.wasm (...)` with no errors. Do not run the browser check yet - that happens once, comprehensively, in Task 5.

---

## Task 2: Enhance Horizontal Motion with a walk-vs-run comparison mode

**Files:**
- Modify: `web/static/index.html:` the `exercise-horizontal` section
- Modify: `web/static/script.js:` the `// --- Exercise 8: Horizontal motion` block (was "Exercise 4" before Task 1's renumbering)

- [ ] **Step 1: Update the horizontal motion section in `web/static/index.html`**

Replace the entire `<section class="exercise" id="exercise-horizontal">` block with:

```html
  <section class="exercise" id="exercise-horizontal">
    <h2>8. Horizontal motion: speed becomes movement</h2>
    <p>Combine a Q4_4 speed with the sub-pixel accumulator and a page/pixel position, and you get SMB's actual per-frame movement routine. The dot below is the object; the vertical line marks a page boundary (pixel 255 to 0).</p>
    <div class="controls">
      <label>speed byte (hex) <input type="text" id="h-speed" value="19" maxlength="2"></label>
      <label>start pixel <input type="number" id="h-pixel" min="0" max="255" value="240"></label>
      <label>frames <input type="number" id="h-frames" min="1" max="64" value="12"></label>
      <label>compare walk ($10) vs run ($19) <input type="checkbox" id="h-compare"></label>
      <button onclick="runHorizontalTrace()">Run</button>
    </div>
    <div class="output">
      <canvas id="h-canvas" width="640" height="160"></canvas>
      <p id="h-result" class="result-line">&nbsp;</p>
    </div>
    <p class="hint">Start pixel near 255 to watch it cross a page boundary within a few frames. Check the comparison box to see walk speed never oscillate while run speed alternates 1,2,1,2.</p>
  </section>
```

(This changes: heading number to `8.`, adds the comparison checkbox, and grows the canvas from `height="100"` to `height="160"` to fit two tracks when comparison mode is on.)

- [ ] **Step 2: Replace the horizontal motion wiring in `web/static/script.js`**

Replace the entire `function runHorizontalTrace() { ... }` block (the one under the `// --- Exercise 8: Horizontal motion` comment) with:

```js
function runHorizontalTrace() {
  const compare = document.getElementById("h-compare").checked;
  const startPixel = parseInt(document.getElementById("h-pixel").value, 10) & 0xff;
  const frames = parseInt(document.getElementById("h-frames").value, 10);

  const canvas = document.getElementById("h-canvas");
  const ctx = canvas.getContext("2d");

  if (!compare) {
    const speed = hexByte(document.getElementById("h-speed").value);
    const trace = nesmathHorizontalTrace(speed, 0, startPixel, frames);
    animateSingleTrack(canvas, ctx, trace, "h-result");
    return;
  }

  const walkTrace = nesmathHorizontalTrace(0x10, 0, startPixel, frames);
  const runTrace = nesmathHorizontalTrace(0x19, 0, startPixel, frames);
  animateComparisonTracks(canvas, ctx, walkTrace, runTrace);
}

function animateSingleTrack(canvas, ctx, trace, resultElId) {
  let frameIndex = 0;
  let pageCrossings = 0;
  let lastPage = 0;

  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const trackY = canvas.height / 2;

    ctx.strokeStyle = "#34344a";
    ctx.beginPath();
    ctx.moveTo(10, trackY);
    ctx.lineTo(canvas.width - 10, trackY);
    ctx.stroke();

    const f = trace[frameIndex];
    const x = 10 + (f.pixel / 255) * (canvas.width - 20);

    ctx.fillStyle = "#4ac0ef";
    ctx.beginPath();
    ctx.arc(x, trackY, 8, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = "#e6e6ef";
    ctx.font = "12px monospace";
    ctx.fillText(`frame ${f.frame}  page=${f.page}  pixel=${f.pixel}  moved=${f.delta}`, 10, 20);

    if (f.page !== lastPage) {
      pageCrossings++;
      lastPage = f.page;
    }

    document.getElementById(resultElId).textContent =
      `frame ${f.frame}/${trace.length}: page=${f.page} pixel=${f.pixel}, moveForce=${toHex(f.moveForce)}, moved ${f.delta} px this frame` +
      (pageCrossings ? `  (crossed a page boundary ${pageCrossings}x so far)` : "");

    frameIndex++;
    if (frameIndex < trace.length) {
      setTimeout(draw, 250);
    }
  }
  draw();
}

function animateComparisonTracks(canvas, ctx, walkTrace, runTrace) {
  let frameIndex = 0;

  function draw() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const walkY = canvas.height * 0.3;
    const runY = canvas.height * 0.75;

    [walkY, runY].forEach((y) => {
      ctx.strokeStyle = "#34344a";
      ctx.beginPath();
      ctx.moveTo(10, y);
      ctx.lineTo(canvas.width - 10, y);
      ctx.stroke();
    });

    const w = walkTrace[frameIndex];
    const r = runTrace[frameIndex];
    const wx = 10 + (w.pixel / 255) * (canvas.width - 20);
    const rx = 10 + (r.pixel / 255) * (canvas.width - 20);

    ctx.fillStyle = "#f2c14e";
    ctx.beginPath();
    ctx.arc(wx, walkY, 7, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = "#4ac0ef";
    ctx.beginPath();
    ctx.arc(rx, runY, 7, 0, Math.PI * 2);
    ctx.fill();

    ctx.fillStyle = "#9494ab";
    ctx.font = "11px monospace";
    ctx.fillText("yellow = walk ($10, no oscillation)", 10, walkY - 14);
    ctx.fillText("cyan = run ($19, oscillates 1,2,1,2)", 10, runY - 14);

    document.getElementById("h-result").textContent =
      `frame ${w.frame}/${walkTrace.length}: walk moved ${w.delta}px this frame, run moved ${r.delta}px this frame`;

    frameIndex++;
    if (frameIndex < walkTrace.length) {
      setTimeout(draw, 250);
    }
  }
  draw();
}
```

- [ ] **Checkpoint: rebuild**

Run: `cd web && ./build.sh`
Expected: succeeds with no errors (this task touches no Go code, so this mainly re-copies `wasm_exec.js`; the real verification of this task is visual, deferred to Task 5).

---

## Task 3: Enhance Vertical Motion with the optional upward-force parameter

**Files:**
- Modify: `web/wasm/main.go:` `runVerticalTrace`
- Modify: `web/static/index.html:` the `exercise-vertical` section
- Modify: `web/static/script.js:` the `// --- Exercise 9: Vertical motion` block (was "Exercise 5" before Task 1's renumbering)

- [ ] **Step 1: Update `runVerticalTrace` in `web/wasm/main.go`**

Replace the entire `runVerticalTrace` function with:

```go
// runVerticalTrace drives VerticalMotion.Step for frameCount frames with a
// fixed gravity force, returning per-frame speed/position/accumulator
// state so the frontend can plot a jump arc and mark its peak (the frame
// where Speed crosses from negative to non-negative). upForce is the
// optional mirrored upward-deceleration force (SMBDIS.ASM:7736-7758);
// pass 0 to skip that section entirely, matching the player.
func runVerticalTrace(_ js.Value, args []js.Value) any {
	initialSpeed := int8(args[0].Int())
	force := uint8(args[1].Int())
	maxSpeed := int8(args[2].Int())
	upForce := uint8(args[3].Int())
	frameCount := args[4].Int()

	v := nesmath.VerticalMotion{Speed: initialSpeed}

	frames := make([]any, 0, frameCount)
	for frame := 1; frame <= frameCount; frame++ {
		delta := v.Step(force, maxSpeed, upForce)
		frames = append(frames, map[string]any{
			"frame":     frame,
			"moveForce": int(v.MoveForce.Value()),
			"dummy":     int(v.Dummy.Value()),
			"speed":     int(v.Speed),
			"page":      int(v.Position.Page),
			"pixel":     int(v.Position.Pixel),
			"delta":     int(delta),
		})
	}
	return frames
}
```

(The only change from the existing version is the new `upForce` parameter at position 4, threaded into `v.Step(force, maxSpeed, upForce)` instead of the hardcoded `0`.)

- [ ] **Step 2: Build and verify the WASM target compiles**

Run: `cd web && GOOS=js GOARCH=wasm go build -o /tmp/step_check.wasm ./wasm`
Expected: exits with no output (success).

- [ ] **Step 3: Update the vertical motion section in `web/static/index.html`**

Replace the entire `<section class="exercise" id="exercise-vertical">` block with:

```html
  <section class="exercise" id="exercise-vertical">
    <h2>9. Vertical motion: the gravity chain and the jump arc</h2>
    <p>Vertical speed is <strong>never</strong> nybble-split - it is a raw signed byte, added directly to position. Sub-pixel precision comes from two accumulators sharing a byte across two different points in their lifecycle each frame. Simulate a jump and watch Speed climb from negative (upward) toward positive (falling) while Position traces the arc.</p>
    <div class="controls">
      <label>initial speed <input type="number" id="v-speed" min="-128" max="127" value="-4"></label>
      <label>gravity force <input type="number" id="v-force" min="0" max="255" value="32"></label>
      <label>max speed <input type="number" id="v-max" min="-128" max="127" value="4"></label>
      <label>upward force (0 = off) <input type="number" id="v-upforce" min="0" max="255" value="0"></label>
      <label>frames <input type="number" id="v-frames" min="1" max="128" value="40"></label>
      <button onclick="runVerticalTrace()">Run</button>
    </div>
    <div class="output">
      <canvas id="v-canvas" width="640" height="220"></canvas>
      <p id="v-result" class="result-line">&nbsp;</p>
    </div>
    <p class="hint">Watch for the frame where Speed crosses from negative to non-negative - that is the peak of the jump. Set upward force > 0 to see the optional mirrored deceleration path (used by e.g. Red Koopa Paratroopa, never by the player).</p>
  </section>
```

- [ ] **Step 4: Update the vertical motion wiring in `web/static/script.js`**

In the `function runVerticalTrace() { ... }` block, replace only the first five lines (the parameter reads and the WASM call) with:

```js
function runVerticalTrace() {
  const speed = parseInt(document.getElementById("v-speed").value, 10);
  const force = parseInt(document.getElementById("v-force").value, 10) & 0xff;
  const maxSpeed = parseInt(document.getElementById("v-max").value, 10);
  const upForce = parseInt(document.getElementById("v-upforce").value, 10) & 0xff;
  const frames = parseInt(document.getElementById("v-frames").value, 10);

  const trace = nesmathVerticalTrace(speed, force, maxSpeed, upForce, frames);
```

Leave the rest of the function (canvas drawing, peak detection) unchanged.

- [ ] **Checkpoint: rebuild**

Run: `cd web && ./build.sh`
Expected: succeeds with no errors.

---

## Task 4: Add the capstone combined-trajectory exercise

**Files:**
- Modify: `web/wasm/main.go` (add `runJumpTrajectory` + registration)
- Modify: `web/static/index.html` (add the capstone section at the end of `<main>`)
- Modify: `web/static/script.js` (add the capstone wiring function)

- [ ] **Step 1: Add `runJumpTrajectory` to `web/wasm/main.go`**

Add this function anywhere after `runVerticalTrace`, and add its registration to `main()` (append `js.Global().Set("nesmathJumpTrajectory", js.FuncOf(runJumpTrajectory))` to the list already built up in Task 1):

```go
// runJumpTrajectory drives a HorizontalMotion and a VerticalMotion
// together, frame by frame, returning cumulative X/Y displacement so the
// frontend can plot the combined trajectory as a single curve. This is
// the only exercise that composes both motion types in one simulation.
func runJumpTrajectory(_ js.Value, args []js.Value) any {
	hSpeed := nesmath.Q4_4(uint8(args[0].Int()))
	vSpeed := int8(args[1].Int())
	vForce := uint8(args[2].Int())
	vMaxSpeed := int8(args[3].Int())
	frameCount := args[4].Int()

	h := nesmath.HorizontalMotion{Speed: hSpeed}
	v := nesmath.VerticalMotion{Speed: vSpeed}

	var x, y int
	frames := make([]any, 0, frameCount)
	for frame := 1; frame <= frameCount; frame++ {
		x += int(h.Step())
		y += int(v.Step(vForce, vMaxSpeed, 0))
		frames = append(frames, map[string]any{
			"frame": frame,
			"x":     x,
			"y":     y,
		})
	}
	return frames
}
```

- [ ] **Step 2: Build and verify the WASM target compiles**

Run: `cd web && GOOS=js GOARCH=wasm go build -o /tmp/step_check.wasm ./wasm`
Expected: exits with no output (success).

- [ ] **Step 3: Add the capstone section to `web/static/index.html`**

Insert this new section immediately before the closing `</main>` tag (after the `exercise-vertical` section):

```html
  <section class="exercise" id="exercise-capstone">
    <h2>10. Capstone: the full jump trajectory</h2>
    <p>The finale: HorizontalMotion and VerticalMotion driven together, frame by frame, plotted as a single X-Y curve - horizontal displacement against vertical displacement, with "up" drawn upward to match the in-game convention. This is everything above, composed.</p>
    <div class="controls">
      <label>horizontal speed (hex) <input type="text" id="capstone-hspeed" value="19" maxlength="2"></label>
      <label>initial vertical speed <input type="number" id="capstone-vspeed" min="-128" max="127" value="-8"></label>
      <label>gravity force <input type="number" id="capstone-vforce" min="0" max="255" value="20"></label>
      <label>max fall speed <input type="number" id="capstone-vmax" min="-128" max="127" value="4"></label>
      <label>frames <input type="number" id="capstone-frames" min="1" max="128" value="50"></label>
      <button onclick="runJumpTrajectory()">Run</button>
    </div>
    <div class="output">
      <canvas id="capstone-canvas" width="640" height="260"></canvas>
      <p id="capstone-result" class="result-line">&nbsp;</p>
    </div>
    <p class="hint">This is the shape of a Mario jump: rising, arcing over, falling - produced entirely by integer carry chains, no floating point anywhere.</p>
  </section>
```

- [ ] **Step 4: Add the capstone wiring to `web/static/script.js`**

Append this to the end of the file:

```js
// --- Exercise 10: Capstone - combined jump trajectory --------------------

function runJumpTrajectory() {
  const hSpeed = hexByte(document.getElementById("capstone-hspeed").value);
  const vSpeed = parseInt(document.getElementById("capstone-vspeed").value, 10);
  const vForce = parseInt(document.getElementById("capstone-vforce").value, 10) & 0xff;
  const vMax = parseInt(document.getElementById("capstone-vmax").value, 10);
  const frames = parseInt(document.getElementById("capstone-frames").value, 10);

  const trace = nesmathJumpTrajectory(hSpeed, vSpeed, vForce, vMax, frames);

  const canvas = document.getElementById("capstone-canvas");
  const ctx = canvas.getContext("2d");
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  const padding = 20;
  const xs = trace.map((f) => f.x);
  const ys = trace.map((f) => f.y);
  const xMin = Math.min(0, ...xs);
  const xMax = Math.max(1, ...xs);
  const yMin = Math.min(0, ...ys);
  const yMax = Math.max(1, ...ys);
  const xRange = Math.max(1, xMax - xMin);
  const yRange = Math.max(1, yMax - yMin);

  function screenX(x) {
    return padding + ((x - xMin) / xRange) * (canvas.width - padding * 2);
  }
  function screenY(y) {
    // Invert: more negative y (upward, NES convention) draws higher on screen.
    return canvas.height - padding - ((y - yMin) / yRange) * (canvas.height - padding * 2);
  }

  ctx.strokeStyle = "#4ac0ef";
  ctx.lineWidth = 2;
  ctx.beginPath();
  trace.forEach((f, i) => {
    const sx = screenX(f.x);
    const sy = screenY(f.y);
    if (i === 0) ctx.moveTo(sx, sy);
    else ctx.lineTo(sx, sy);
  });
  ctx.stroke();

  ctx.fillStyle = "#5fd68a";
  ctx.beginPath();
  ctx.arc(screenX(trace[0].x), screenY(trace[0].y), 5, 0, Math.PI * 2);
  ctx.fill();

  const last = trace[trace.length - 1];
  ctx.fillStyle = "#ef4a5c";
  ctx.beginPath();
  ctx.arc(screenX(last.x), screenY(last.y), 5, 0, Math.PI * 2);
  ctx.fill();

  document.getElementById("capstone-result").textContent =
    `after ${trace.length} frames: horizontal displacement=${last.x}px, vertical displacement=${last.y}px ` +
    `(green = start, red = end)`;
}
```

- [ ] **Checkpoint: rebuild**

Run: `cd web && ./build.sh`
Expected: succeeds with no errors.

---

## Task 5: Full verification

**Files:** none (verification only)

- [ ] **Step 1: Verify the main module is unaffected**

Run: `cd /Users/pgeorgia/gocode/src/github.com/drpaneas/nesmath && gofmt -l . && go vet ./... && go build ./... && go test ./... -count=1`
Expected: `gofmt -l .` prints nothing, `go vet`/`go build` produce no output, `go test` prints `ok` for both `github.com/drpaneas/nesmath` and `github.com/drpaneas/nesmath/examples`.

- [ ] **Step 2: Verify the WASM target builds cleanly one more time, fully**

Run: `cd web && ./build.sh`
Expected: `Done: web/static/nesmath.wasm (...)` with no errors.

- [ ] **Step 3: Serve and browser-verify all ten exercises**

Run: `cd web/static && python3 -m http.server 8080` (background this)

Using a browser-testing subagent (or manual browser check), verify for each of the 10 exercises:
1. **ADC** - a=144, b=144, carry=0 -> result=32, carry=1 (unchanged from before)
2. **SBC** - a=3, b=5, carry=1 -> result=254 (0xFE), carry=0 (borrow occurred)
3. **16-bit chained addition** - a=(255,0), b=(1,0) -> lowResult=0, lowCarry=1, highResult=1, highCarry=0, combined16=256
4. **Negate** - v=0x80 -> onesComplement=0x7F, result=0x80 (wraps to itself); v=40 -> result=216 (0xD8)
5. **Q4_4 Split** - unchanged behavior, still shows the `$FC` callout
6. **Accumulator oscillation** - unchanged behavior (9 carries out of 16 frames at the default 144/16)
7. **Position16 alone** - velocity=4, start pixel=250, 6 frames -> pixel sequence 254, 2 (page crosses to 1), 6, 10, 14, 18
8. **Horizontal motion** - unchecked comparison box behaves as before; checking it shows two tracks (walk never oscillates, run alternates 1,2,1,2)
9. **Vertical motion** - upForce=0 behaves identically to before; setting upForce > 0 visibly changes the trace (Speed decreases faster due to the mirrored SBC path)
10. **Capstone** - produces a visible curve; green start dot, red end dot, no console errors

Confirm zero browser console errors throughout.

- [ ] **Step 4: Update `web/README.md`'s exercise count reference if present**

Check `web/README.md` for any text mentioning "five interactive exercises" (the current file's intro line says exactly this: `driving five interactive exercises in the browser`). Update it to `driving ten interactive exercises in the browser`.

---

## Self-review notes (completed during plan authoring)

- **Spec coverage:** All 10 exercises from the spec (`docs/superpowers/specs/2026-07-03-interactive-exercises-expansion-design.md`) map to a task above: 1 (Task-independent, unchanged), 2/3/4/7 (Task 1), 5/6 (unchanged, renumbered in Task 1), 8 (Task 2), 9 (Task 3), 10 (Task 4). The spec's "Architecture notes" section (no new CSS classes needed) is honored - no task touches `style.css`.
- **Type consistency:** `runVerticalTrace`'s new signature (`speed, force, maxSpeed, upForce, frameCount`) is used identically in Task 3 Steps 1 and 4 - the Go export and the JS call site agree on argument order.
- **Placeholder scan:** no TBD/TODO markers; every step has complete, runnable code.
