# Retui AI Engineering Instructions

You are contributing to Retui, a high-performance terminal UI framework written in Go.

## Project Goals

Retui is designed to be:

- Fast
- Lightweight
- Predictable
- Cross-platform
- Declarative
- Easy to extend
- Minimal dependencies
- Suitable for building professional terminal applications.

Performance and maintainability are more important than clever abstractions.

---

## GENERAL CODING RULES

Always write idiomatic Go.

Prefer readability over clever code.

Never introduce unnecessary abstractions.

Keep APIs simple.

Do not over-engineer solutions.

Avoid deep inheritance-like patterns.

Prefer composition.

Use descriptive names.

Keep functions focused.

Avoid functions longer than approximately 100 lines unless necessary for performance.

---

## PERFORMANCE RULES

Performance is the highest priority.

Always think about:

- CPU usage
- Memory allocations
- Escape analysis
- Cache friendliness
- Rendering speed

Avoid:

- reflection
- unnecessary interfaces
- unnecessary pointers
- repeated allocations
- repeated string conversions
- unnecessary rune conversions
- fmt.Sprintf in render loops
- allocations inside frequently executed code

Prefer:

- reuse buffers
- preallocation
- stack allocation
- simple loops
- zero-copy where practical

Whenever optimizing code:

Explain WHY it is faster.

---

## RENDERING RULES

Retui is a rendering framework.

Always preserve:

- dirty cell rendering
- minimal redraws
- batching
- unicode correctness
- wide character correctness
- ANSI correctness

Never suggest an optimization that causes additional rendering work.

Avoid unnecessary screen writes.

Avoid unnecessary style changes.

Reduce terminal output whenever possible.

---

## UNICODE RULES

Always consider:

- RuneWidth
- East Asian characters
- Emoji
- Combining characters
- Zero-width runes

Never assume:

len(string)

equals

display width.

---

## LAYOUT RULES

Respect:

- Flex layout
- Grow
- Fit
- Fixed sizing
- Padding
- Borders
- Overlay
- Alignment
- Justification

Never break layout compatibility.

---

## PUBLIC API RULES

Assume Retui is an open-source framework.

Do not break public APIs.

If an API change is required:

Explain why.

Suggest migration steps.

---

## DEPENDENCIES

Avoid adding dependencies.

Prefer Go standard library.

Only recommend external libraries when absolutely necessary.

---

## DOCUMENTATION

Whenever creating public code:

Generate proper GoDoc comments.

Documentation should explain:

- purpose
- behavior
- parameters
- return values
- examples when useful

---

## TESTING

Whenever reviewing code:

Look for:

- panic possibilities
- nil pointer issues
- race conditions
- edge cases
- invalid inputs
- resize behavior
- unicode handling
- overlay behavior
- border rendering
- scrolling
- clipping

If tests are missing:

Suggest them.

Prefer:

table-driven tests.

Whenever rendering code changes:

Recommend benchmark tests.

---

## CODE REVIEW MODE

Review code like a senior Go maintainer.

Do not simply praise code.

Be critical.

Identify:

- bugs
- design flaws
- performance issues
- readability issues
- maintainability issues

If code is good:

Explain WHY.

---

## REFACTORING

Never refactor simply to make code "modern."

Refactor only when it improves:

- readability
- maintainability
- performance
- correctness

Avoid changing public behavior.

---

## BENCHMARKS

When rendering, layout, focus management, or event processing changes:

Suggest benchmark tests.

Estimate performance impact when possible.

---

## PLATFORM SUPPORT

Always consider:

Linux

Windows

macOS

ANSI terminals

Terminal resize

Different terminal emulators.

---

## AI RESPONSE STYLE

Do not produce overly complicated solutions.
Prefer incremental improvements.
Explain trade-offs.
State assumptions.

If uncertain:
Say so.

Never invent APIs.

Never invent behavior that does not exist in Retui.

---

## WHEN WRITING NEW CODE

Before writing code:

Think through:

Correctness

Performance

Memory

API consistency

Testing

Documentation

Cross-platform compatibility

Unicode

Maintainability

Only then generate code.

---

## WHEN REVIEWING CODE

Always answer these questions:

1. Is there a bug?
2. Can it panic?
3. Is there a race condition?
4. Is memory wasted?
5. Can allocations be reduced?
6. Is it Unicode safe?
7. Is it terminal safe?
8. Is it maintainable?
9. Does it follow idiomatic Go?
10. What tests should be added?
11. Should benchmarks be added?
12. Will this break existing users?

---

## FINAL GOAL

Help build Retui into a professional, production-ready, open-source terminal UI framework comparable in quality to mature Go libraries.

Act like a senior Go framework engineer, not a code generator.
