// Package nesmath models the arithmetic architecture of the Ricoh 2A03
// (the NES's 6502-derived CPU), as used by classic NES platformers, for
// the purpose of porting NES physics to modern languages with bit-exact
// accuracy.
//
// # This is not a fixed-point number library
//
// Generic fixed-point libraries (Q16.16, Q8.8, and similar) model a
// numeric type: a value with a decimal point pinned at a known bit
// position, supporting +, -, *, /. That is the wrong abstraction for NES
// game physics.
//
// The 6502 has no multiply, no divide, and only one arithmetic primitive
// that matters here: add-with-carry (ADC) and subtract-with-borrow (SBC).
// Game physics on the NES is built from independent bytes connected by
// carry propagation:
//
//	Force -> carry -> Speed -> carry -> SubPixel -> carry -> Pixel -> carry -> Page
//
// Most of those bytes are not numbers. A sub-pixel accumulator such as a
// horizontal move-force byte is not "the fractional part of position" -
// it is a delay counter that accumulates fractional crumbs each frame
// until it overflows, at which point the carry promotes one whole pixel
// of motion into the position byte. nesmath models that role directly
// with [Accumulator8], distinct from [Q4_4], the one type in this
// package that does have a meaningful decimal interpretation.
//
// # Design principle
//
// Model hardware roles, not decimal values:
//
//   - [Q4_4] is a speed byte. You split it and read it.
//   - [Accumulator8] is a carry-producing bucket. You add to it and check
//     for overflow. It has no meaningful "value".
//   - [Carry] is a uint8 (0 or 1) that threads through [ADC] and [SBC]
//     calls like a pipeline. It is never branched on with an if statement,
//     because the 6502 never branches on carry mid-chain - it feeds the
//     carry into the next ADC.
//   - [Position16] is a page and a pixel, not a 16-bit integer.
//
// # Ground truth
//
// Every type and method in this package documents the assembly routine
// it mirrors, using names chosen to describe hardware roles rather than
// any one disassembly's original internal labels. The full conceptual
// background - why fixed-point instead of float, the history of the 4.4
// and 8.8 formats, and a worked frame-by-frame trace of the carry chain -
// is in [NES_MATH_FOR_PORTERS.md] in this repo.
//
// # What this package does not do
//
// It does not provide friction tables, jump-force tables, or speed tiers -
// those are game data that belongs in a port, not in this library. It does
// not convert from float64, because the NES works in bytes, not decimals.
// It does not multiply or divide, because NES platformers generally do
// not either. It has no String method that prints "1.5625" - the decimal
// interpretation is a human convenience, not something the game computes.
//
// # Beyond a single game's design
//
// The arithmetic modeled here - a nybble-split [Q4_4] speed feeding a
// separate [Accumulator8] sub-pixel byte - is one NES game design
// choice, not the NES's only design choice. Plenty of NES games use
// simpler schemes:
//
//   - A single signed velocity byte added directly into the fractional
//     half of a 16-bit position, with no separate speed/accumulator
//     split at all (a plain Q8.8 pair, sometimes taught as the "default"
//     NES sub-pixel technique). That degrades to a single call to
//     [Position16.AddSigned] with the raw velocity as the delta - no
//     [Q4_4] or [Accumulator8] involved.
//   - No fractional component whatsoever: some objects (a bullet, a
//     simple enemy drifting one direction) just move a fixed whole
//     number of pixels every frame via plain ADC/SBC. That is exactly
//     what [Position16.AddSigned] does when called with carryIn always 0
//     and no accumulator feeding it.
//
// nesmath's layered primitives are built to support both without forcing
// a three-tier structure onto a simpler game: [ADC], [SBC], and
// [Position16] work standalone, and [Q4_4]/[Accumulator8] are opt-in for
// games (or objects within a game) that actually need this style of
// sub-pixel precision.
//
// # Validation
//
// Every function is tested against hand-verified fixtures traceable to
// real 6502 assembly patterns (see the test files in this package). A
// further validation step - replaying real ROM traces captured via an
// emulator and comparing them frame-for-frame against this package - is a
// natural extension for callers building a full port, but is
// intentionally left out of this package to keep it dependency-free.
//
// [NES_MATH_FOR_PORTERS.md]: https://github.com/drpaneas/nesmath/blob/main/NES_MATH_FOR_PORTERS.md
package nesmath
