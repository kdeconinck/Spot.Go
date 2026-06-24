# Documentation

The following documents are authoritative:

| Document               | Purpose                                                |
| ---------------------- | ------------------------------------------------------ |
| `README.md`            | Project overview and high-level concepts.              |
| `docs/dsl.md`          | DSL syntax, semantics, validation rules, and examples. |
| `docs/architecture.md` | Pipeline stages, responsibilities, and data flow.      |
| `docs/roadmap.md`      | Milestones, priorities, and future work.               |

Before introducing behavior, verify whether it should first be reflected in one of the project documents.
Do not introduce DSL features, architectural changes, or significant behavior that is not documented appropriately.
When implementation and documentation diverge, update one of them so they are consistent.

# Core Principles

1. Performance is a feature.
2. Measure before optimizing.
3. Simplicity beats flexibility.
4. Build for current requirements, not hypothetical future requirements.
5. Write idiomatic Go.
6. Prefer explicit code over abstraction.
7. Prefer concrete types over unnecessary interfaces.
8. Prefer removing code over adding code.

When multiple solutions are possible, choose the simplest solution that correctly satisfies the current requirements.

# Performance

Performance is a primary project requirement.
However, performance decisions must be guided by measurement, not assumptions.

Before introducing an optimization:

1. Measure the current implementation.
2. Implement the change.
3. Measure again.
4. Keep the change only if the improvement is meaningful.

Avoid speculative optimization.
Prefer simple code with benchmark evidence over complex code based on intuition.

# Benchmarks

Performance-sensitive code must have benchmarks.

Examples include:

* DSL parsing.
* validation.
* compilation.
* scanning.
* tokenization.
* rule evaluation.
* diagnostic rendering.

When modifying performance-sensitive code, add or update benchmarks as necessary.
Benchmark results are more important than assumptions.

# Concurrency

Do not introduce concurrency prematurely.
Single-threaded performance must be excellent before considering parallel execution.

Before introducing concurrency:

1. Identify a measurable bottleneck.
2. Demonstrate the bottleneck with benchmarks.
3. Demonstrate that concurrency improves performance.
4. Verify that the additional complexity is justified.

Correctness and simplicity take priority over parallelism.

# Idiomatic Go

Write idiomatic Go.
Follow standard Go conventions and practices.

Prefer:

* Simple packages.
* Simple types.
* Explicit control flow.
* Composition.
* Small functions.
* Clear naming.

Avoid:

* Java-style architecture.
* Enterprise design patterns.
* Dependency injection frameworks.
* Excessive abstraction.
* Unnecessary indirection.

When uncertain, prefer the approach commonly found in the Go standard library.

# Documentation

All exported declarations must be documented.

This includes:

* Constants.
* Variables.
* Types.
* Struct fields.
* Functions.
* Methods.

Documentation must follow Go conventions.

Good:

```go
// Position represents a byte offset within a source file.
type Position struct {
    // Offset is the zero-based byte offset within the source file.
    Offset int
}
```

Bad:

```go
// Position stores a position.
type Position struct {
    Offset int
}
```

Comments should explain purpose and behavior.
Comments should not merely restate names.

# Package Design

Packages must have a single, clearly defined responsibility.
Avoid catch-all packages such as:

* Util.
* Utils.
* Helper.
* Helpers.
* Common.
* Shared.
* Misc.

Package names should communicate ownership and responsibility.
Dependencies between packages should remain explicit and easy to understand.

# Functions

Functions should:

* Be small and easy to read.
* Prefer straightforward control flow.
* Avoid deep nesting.
* Generally perform one task.
* Optimized for readability first.

Extract helper functions when doing so improves readability.

# Types

Do not introduce types before they are needed.
Do not create abstractions for anticipated future requirements.

Every type must solve a current problem.

Bad:

```go
type Position struct {
    Offset int
    Line   int
    Column int
}
```

when only `Offset` is currently required.

Good:

```go
type Position struct {
    Offset int
}
```

Add fields only when they serve a real and current purpose.
The same principle applies to:

* Structs.
* Fields.
* Methods.
* Packages.
* Configuration options.
* Extension points.

Build what is needed now.
Add complexity only when a concrete requirement justifies it.

# Interfaces

Do not introduce interfaces without a demonstrated need.
Do not create interfaces solely for future flexibility.
Do not create interfaces solely for testing.
Prefer concrete types until abstraction becomes necessary.
Interfaces should generally be owned by consumers rather than producers.
A single implementation is usually not sufficient justification for introducing an interface.

# Memory and Allocations

Avoid unnecessary allocations.
Avoid unnecessary string copying.
Avoid unnecessary temporary objects.
Avoid converting between strings and byte slices unless required.
Reuse memory where it improves performance without significantly increasing complexity.
When allocation behavior is unclear, benchmark it.

# Dependencies

Prefer the Go standard library.

Before introducing a dependency, verify that:

* It solves a real problem.
* The problem cannot reasonably be solved locally.
* The maintenance cost is justified.
* The performance impact is acceptable.

A small amount of local code is often preferable to an additional dependency.

# Error Handling

Errors should be explicit.
Return errors rather than hiding failures.
Error messages should provide useful context.

Good:

```text
failed to parse token definition "Identifier": unexpected character ')'
```

Bad:

```text
parse error
```

Errors should help users understand what happened and where it happened.

# Testing

Aim for 100% code coverage.
Coverage is not the goal by itself. The goal is confidence that every reachable path and observable behavior is tested.
Tests should generally be written in an external `_test` package.

Example:

```go
package scanner_test
```

instead of:

```go
package scanner
```

External tests ensure that behavior is verified through the public API rather than internal implementation details.
Avoid testing unexported functions directly.
Avoid exposing APIs solely to support testing.

## Reachability

Do not add defensive code for impossible states.
If a value is guaranteed by design, do not add unreachable safeguards for it.

Bad:

```go
func Parse(tokens []Token) error {
    if tokens == nil {
        return errors.New("tokens cannot be nil")
    }

    // ...
}
```

when `tokens` is only produced by a scanner that always returns a non-nil token slice.
Prefer making invariants explicit in the design instead of adding unreachable checks.
Unreachable code should not exist merely to make the implementation feel safer.
If a branch cannot be reached through the public API, reconsider whether the branch should exist.

## Path Coverage

Every reachable path should be tested.

If the same code can be reached in multiple meaningful ways, test each way separately.
For example, if numeric-character validation can be reached through four different parser paths, write four tests.
Do not rely on one test path to cover behavior that is reachable through another path.
Each public behavior should have its own test, even when it exercises code already covered elsewhere.

## Test Ordering

Test cases should follow the order of the production code.

If the implementation checks conditions in this order:

1. Input length.
2. Numeric character validation.
3. Range validation.

Then tests should be ordered the same way:

1. Invalid length.
2. Non-numeric character.
3. Out-of-range value.

This makes tests easier to compare against the implementation and easier to audit for missing branches.
For table-driven tests, order cases according to the order in which the implementation evaluates the conditions.

## Test Style

Tests should be easy to read and follow a consistent structure.
Use explicit sections:

```go
// Arrange.

// Act.

// Assert.
```

The separation between setup, execution, and verification should remain obvious.

## Parallel Execution

Tests and subtests should run in parallel whenever possible.

Use:

```go
t.Parallel()
```

Serial execution should be the exception rather than the default.

## Table-Driven Tests

Use table-driven tests when multiple cases verify the same behavior.
Table-driven tests should use descriptive test case names that explain the expected behavior.
Prefer maps keyed by the test case description.

Example:

```go
for tcName, tc := range map[string]struct {
    // Test case fields.
}{
    "When ...": {},
} {
    t.Run(tcName, func(t *testing.T) {
        t.Parallel()

        // Arrange.

        // Act.

        // Assert.
    })
}
```

The exact contents of the test case structure are less important than readability and maintainability.

## Test Names

Test names should describe behavior rather than implementation details.

Good:

```text
When line is 0, an error is returned.
When column is -1, an error is returned.
When the input is empty, no tokens are returned.
```

Bad:

```text
invalid line
invalid column
empty input
```

A reader should understand the intent of the test without reading the implementation.

## Test Granularity

Prefer multiple focused tests over a single large test.
Do not combine unrelated behaviors into one test.
If a behavior deserves its own explanation, it deserves its own test.
Tests should verify observable behavior rather than implementation details.

## Internal Tests

External `_test` package tests are the default.

Internal package tests should be rare.
Only use internal tests when there is a clear reason that cannot reasonably be solved through the public API, such as:

* validating a low-level algorithm that is intentionally not exported
* benchmarking a performance-sensitive internal function
* testing behavior that cannot be reached through public contracts

Internal tests must be justified by the code being tested.
Do not use internal tests as a shortcut around poor API design.

## Benchmark Style

Benchmarks should avoid duplicating benchmark logic.
Prefer one unexported, parameterized benchmark helper that contains the actual benchmark implementation.
Exported benchmark functions should call that helper with specific parameters.

Example:

```go
func benchmark_Scan(b *testing.B, inputSize int) {
    b.Helper()

    input := makeInput(inputSize)

    for b.Loop() {
        _ = scanner.Scan(input)
    }
}

func Benchmark_Scan_1KB(b *testing.B)   { benchmark_Scan(b, 1_024) }
func Benchmark_Scan_10KB(b *testing.B)  { benchmark_Scan(b, 10_240) }
func Benchmark_Scan_100KB(b *testing.B) { benchmark_Scan(b, 102_400) }
```

Benchmark helpers should:

* Be unexported.
* Call `b.Helper()`.
* Prepare input outside the measured loop.
* Contain the measured operation inside the benchmark loop.
* Avoid unnecessary allocations inside the measured loop unless those allocations are part of the behavior being
  measured.

Use benchmark names that clearly describe the scenario being measured.

# Change Guidelines

Before adding code, ask:

* Does this solve a current problem?
* Is this the simplest solution?
* Can this be implemented with fewer abstractions?
* Is the added complexity justified?

Before introducing a new:

* Package,
* Type,
* Field,
* Method,
* Interface,
* Configuration option,

Verify that it serves a real and immediate purpose.
Do not add code because it may be useful later.

# Change Size

Changes should be as small as possible while still delivering meaningful project value.
Prefer small, focused commits.
Each change should represent one coherent step from a business or product perspective.

Good examples:

* Add `FileID`.
* Add `Position`.
* Add `Span`.
* Add scanner support for source spans.
* Add diagnostic rendering for source spans.

Do not combine unrelated changes.

However, do not split a change so far that the repository ends up in an artificial or unusable state.
If adding a new token requires coordinated updates to the scanner, parser, token model, and tests, it is acceptable to
update those components together when that produces one coherent behavior change.

Split the work only when a part is independently meaningful or when the combined implementation becomes too large or
difficult to review.

Use judgment.
Smaller is preferred, but coherence matters more than mechanical size.

# Code Review Checklist

Before considering a change complete, verify:

* The code is idiomatic Go.
* All exported declarations are documented.
* The solution is as simple as possible.
* Every type serves a current purpose.
* Every field serves a current purpose.
* Every abstraction serves a current purpose.
* Performance-sensitive changes are benchmarked.
* Performance decisions are based on measurement.
* Concurrency is justified by evidence.
* Unnecessary code has been removed.
* Tests validate the expected behavior.

When in doubt, choose the simpler solution.
