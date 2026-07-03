// Wires the HTML controls in index.html to the functions nesmath.wasm
// exports (see web/wasm/main.go). No framework, no build step - this is
// plain DOM manipulation and 2D canvas drawing.

const statusEl = document.getElementById("wasm-status");
const buttons = () => document.querySelectorAll("button");

buttons().forEach((b) => (b.disabled = true));

const go = new Go();

function onWasmReady(result) {
  go.run(result.instance);
  buttons().forEach((b) => (b.disabled = false));
}

// No visible banner on success - only surfaced if loading genuinely fails,
// so users aren't left with silently-disabled buttons and no explanation.
function onWasmError(err) {
  statusEl.textContent = "Failed to load nesmath.wasm: " + err;
  statusEl.className = "status error";
  statusEl.hidden = false;
}

// instantiateStreaming requires the server to send Content-Type:
// application/wasm; most hosts (including GitHub Pages) do, but some
// static hosts, proxies, or CDNs don't set it correctly. Fall back to
// fetching the raw bytes and instantiating from an ArrayBuffer, which
// works regardless of the Content-Type header.
WebAssembly.instantiateStreaming(fetch("nesmath.wasm"), go.importObject)
  .then(onWasmReady)
  .catch(() => {
    fetch("nesmath.wasm")
      .then((response) => response.arrayBuffer())
      .then((bytes) => WebAssembly.instantiate(bytes, go.importObject))
      .then(onWasmReady)
      .catch(onWasmError);
  });

function hexByte(str) {
  const n = parseInt(str, 16);
  return Number.isNaN(n) ? 0 : n & 0xff;
}

function toHex(n, digits = 2) {
  return "0x" + (n & (Math.pow(16, digits) - 1)).toString(16).padStart(digits, "0").toUpperCase();
}

function bitsOf(byte) {
  return byte
    .toString(2)
    .padStart(8, "0")
    .split("")
    .map(Number);
}

function renderByteBits(container, label, byte, splitNybbles) {
  const row = document.createElement("div");
  row.className = "byte-row";

  const labelEl = document.createElement("div");
  labelEl.className = "byte-label";
  labelEl.textContent = `${label} = ${toHex(byte)} (${byte})`;
  row.appendChild(labelEl);

  const bitsEl = document.createElement("div");
  bitsEl.className = "bits";
  const bits = bitsOf(byte);

  if (splitNybbles) {
    const high = document.createElement("div");
    high.className = "nybble high";
    bits.slice(0, 4).forEach((bit) => high.appendChild(makeBit(bit)));
    const low = document.createElement("div");
    low.className = "nybble low";
    bits.slice(4, 8).forEach((bit) => low.appendChild(makeBit(bit)));
    bitsEl.appendChild(high);
    bitsEl.appendChild(low);
  } else {
    bits.forEach((bit) => bitsEl.appendChild(makeBit(bit)));
  }

  row.appendChild(bitsEl);
  container.appendChild(row);
}

function makeBit(bit) {
  const el = document.createElement("div");
  el.className = "bit" + (bit ? " one" : "");
  el.textContent = bit;
  return el;
}

// --- Exercise 1: ADC -------------------------------------------------

function runADC() {
  const a = parseInt(document.getElementById("adc-a").value, 10) & 0xff;
  const b = parseInt(document.getElementById("adc-b").value, 10) & 0xff;
  const c = parseInt(document.getElementById("adc-c").value, 10);

  const out = nesmathADC(a, b, c);

  const grid = document.getElementById("adc-grid");
  grid.innerHTML = "";
  renderByteBits(grid, "a", a, false);
  renderByteBits(grid, "b", b, false);
  renderByteBits(grid, "result", out.result, false);

  document.getElementById("adc-result").textContent =
    `${a} + ${b} + carry(${c}) = ${out.result} (${toHex(out.result)}), carry out = ${out.carry}` +
    (out.carry ? "  <- overflowed past 255" : "");
}

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

// --- Exercise 5: Q4_4 Split -------------------------------------------

function runSplit() {
  const raw = hexByte(document.getElementById("split-hex").value);
  const out = nesmathSplit(raw);

  const grid = document.getElementById("split-grid");
  grid.innerHTML = "";
  renderByteBits(grid, "speed", raw, true);

  let line =
    `high nybble = ${out.highNybble}, low nybble = ${out.lowNybble}\n` +
    `Split() -> whole = ${out.whole}, frac = ${toHex(out.frac)}\n` +
    `real decimal value = ${out.decimal}`;

  if (raw === 0xfc) {
    line +=
      "\n\nThis is the famous case: some write-ups claim 0xFC means -4.0 by " +
      "reading int8(0xFC) directly. That is only true if this byte is used " +
      "as a raw VERTICAL speed (never split at all). If it were split as a " +
      "horizontal Q4_4 value, the real result is whole=-1, frac=0xC0, " +
      "i.e. -0.25 - both numbers are 'correct', for two different bytes " +
      "that happen to share the same bit pattern.";
  }

  document.getElementById("split-result").textContent = line;
}

// --- Exercise 6: Accumulator oscillation ------------------------------

function runAccumulatorTrace() {
  const value = parseInt(document.getElementById("acc-value").value, 10) & 0xff;
  const frames = parseInt(document.getElementById("acc-frames").value, 10);

  const trace = nesmathAccumulatorTrace(value, frames);

  const canvas = document.getElementById("acc-canvas");
  const ctx = canvas.getContext("2d");
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  const barWidth = canvas.width / trace.length;
  let carries = 0;

  trace.forEach((f, i) => {
    const x = i * barWidth;
    const barHeight = (f.value / 255) * (canvas.height - 30);

    ctx.fillStyle = "#34344a";
    ctx.fillRect(x + 2, canvas.height - 20 - barHeight, barWidth - 4, barHeight);

    if (f.carry) {
      carries++;
      ctx.fillStyle = "#ef4a5c";
      ctx.fillRect(x + 2, canvas.height - 20 - barHeight - 6, barWidth - 4, 4);
    }

    ctx.fillStyle = "#9494ab";
    ctx.font = "9px monospace";
    ctx.fillText(String(i + 1), x + barWidth / 2 - 3, canvas.height - 6);
  });

  document.getElementById("acc-result").textContent =
    `${carries} carr${carries === 1 ? "y" : "ies"} out of ${trace.length} frames ` +
    `(red marks = carry fired = one extra pixel promoted that frame). ` +
    `Naive expectation from "every other frame" would be ${Math.floor(trace.length / 2)}.`;
}

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

// --- Exercise 8: Horizontal motion -------------------------------------

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

document.getElementById("h-compare").addEventListener("change", (e) => {
  document.getElementById("h-speed").disabled = e.target.checked;
});

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

// --- Exercise 9: Vertical motion / jump arc -----------------------------

function runVerticalTrace() {
  const speed = parseInt(document.getElementById("v-speed").value, 10);
  const force = parseInt(document.getElementById("v-force").value, 10) & 0xff;
  const maxSpeed = parseInt(document.getElementById("v-max").value, 10);
  const upForce = parseInt(document.getElementById("v-upforce").value, 10) & 0xff;
  const frames = parseInt(document.getElementById("v-frames").value, 10);

  const trace = nesmathVerticalTrace(speed, force, maxSpeed, upForce, frames);

  const canvas = document.getElementById("v-canvas");
  const ctx = canvas.getContext("2d");
  ctx.clearRect(0, 0, canvas.width, canvas.height);

  const padding = 30;
  const plotW = canvas.width - padding * 2;
  const plotH = canvas.height - padding * 2;

  // Cumulative displacement (world position relative to start) so the
  // arc is visible even though Position16 wraps at page boundaries.
  let cumulative = 0;
  const positions = trace.map((f) => {
    cumulative += f.delta;
    return cumulative;
  });
  const posMin = Math.min(0, ...positions);
  const posMax = Math.max(0, ...positions);
  const posRange = Math.max(1, posMax - posMin);

  const speedMin = Math.min(...trace.map((f) => f.speed), maxSpeed * -1);
  const speedMax = Math.max(...trace.map((f) => f.speed), maxSpeed);
  const speedRange = Math.max(1, speedMax - speedMin);

  function xFor(i) {
    return padding + (i / (trace.length - 1)) * plotW;
  }

  // Position line (cyan): displacement from start. Position increases
  // downward in-game (NES convention: negative = up, positive = falling),
  // and canvas Y also increases downward, so mapping posMin (the jump's
  // peak) to a small screen-Y (top of canvas) and posMax to a large
  // screen-Y (bottom) is a direct, un-flipped scale - the same fix
  // applied to the capstone exercise's screenY, for the same reason.
  ctx.strokeStyle = "#4ac0ef";
  ctx.beginPath();
  positions.forEach((p, i) => {
    const y = padding + ((p - posMin) / posRange) * plotH;
    if (i === 0) ctx.moveTo(xFor(i), y);
    else ctx.lineTo(xFor(i), y);
  });
  ctx.stroke();

  // Speed line (red)
  ctx.strokeStyle = "#ef4a5c";
  ctx.beginPath();
  trace.forEach((f, i) => {
    const y = canvas.height - padding - ((f.speed - speedMin) / speedRange) * plotH;
    if (i === 0) ctx.moveTo(xFor(i), y);
    else ctx.lineTo(xFor(i), y);
  });
  ctx.stroke();

  // Zero-speed reference line
  ctx.strokeStyle = "#34344a";
  ctx.setLineDash([4, 4]);
  const zeroY = canvas.height - padding - ((0 - speedMin) / speedRange) * plotH;
  ctx.beginPath();
  ctx.moveTo(padding, zeroY);
  ctx.lineTo(canvas.width - padding, zeroY);
  ctx.stroke();
  ctx.setLineDash([]);

  // Find the peak: first frame where speed crosses from negative to >= 0.
  let peakFrame = null;
  for (let i = 0; i < trace.length; i++) {
    if (trace[i].speed >= 0) {
      peakFrame = trace[i].frame;
      break;
    }
  }

  ctx.fillStyle = "#9494ab";
  ctx.font = "11px monospace";
  ctx.fillText("cyan = position (displacement from start)", padding, 14);
  ctx.fillText("red = Speed", padding, 26);

  document.getElementById("v-result").textContent =
    `final speed=${trace[trace.length - 1].speed}, total displacement=${cumulative}px` +
    (peakFrame ? `. Peak (Speed crosses from negative to non-negative) at frame ${peakFrame}.` : ". Speed never reached zero in this many frames - try more frames or less gravity force.");
}

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
    // y increases downward in-game (NES convention: negative = up,
    // positive = down/falling). Canvas Y also increases downward, so no
    // outer inversion is needed - mapping yMin to the top (small screen-Y)
    // and yMax to the bottom (large screen-Y) is a direct, un-flipped scale.
    return padding + ((y - yMin) / yRange) * (canvas.height - padding * 2);
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
