# Design Principles

The Spot architecture is built around a small number of principles.

1. Data flows in one direction.
2. Each stage has a single responsibility.
3. Stages communicate through explicit data structures.
4. Validation occurs before execution.
5. Invalid configurations should never reach execution stages.
6. Performance is important but should not compromise clarity.
7. Compilation should happen once and execution should happen many times.
8. Earlier stages must not depend on later stages.

Architecture should remain simple.
New stages, abstractions, and data structures should only be introduced when they solve a current problem.

# System Overview

Spot is implemented as a linear pipeline.

```text
DSL Source
    ↓
Parser
    ↓
Syntax Tree
    ↓
Resolver
    ↓
Resolved Syntax
    ↓
Validator
    ↓
Configuration
    ↓
Compiler
    ↓
Compiled Configuration
    ↓
Scanner
    ↓
Token Stream
    ↓
Rule Engine
    ↓
Diagnostics
```

Each stage consumes the output of the previous stage and produces output for the next stage.
Stages should not bypass other stages.
Stages should not communicate through hidden state.
Stages should not reach backwards into earlier stages.

# Pipeline Stages

## Parser

Input:

```text
DSL source text
```

Output:

```text
Syntax tree
```

Responsibilities:

* Parse DSL syntax.
* Produce syntax nodes.
* Preserve source locations.
* Detect syntax errors.
* Stop at the first syntax error and return it.
* Materialize syntax data in parser-owned flat storage.

Non-responsibilities:

* Semantic validation.
* Rule execution.
* Tokenization.
* Analysis.
* Optimization.

The parser answers the question:

> Is this DSL syntactically valid?

Implementation notes:

* The current parser stores syntax data in flat slices rather than as a pointer-linked object graph.
* Definition and token expressions are stored in a shared expression arena and referenced by indices.
* The parser performs a counting pass before materialization so flat storage can be allocated with exact capacities.
* This layout reduces allocation churn, keeps syntax data densely packed in memory, and makes parser cost more
  predictable during later traversal by the validator and compiler.

## Validator

Input:

```text
Resolved syntax
```

Output:

```text
Configuration
```

Responsibilities:

* Validate references.
* Validate section contents.
* Validate naming rules.
* Validate token definitions.
* Validate rule definitions.
* Validate DSL semantics.

Non-responsibilities:

* Execution.
* Compilation.
* Tokenization.
* Analysis.

The validator answers the question:

> Is this DSL semantically valid?

## Resolver

Input:

```text
Syntax tree
```

Output:

```text
Resolved syntax
```

Responsibilities:

* Expose parsed sections in traversal order.
* Build reusable name-to-declaration lookups.
* Preserve first-declaration indices for duplicate checks.
* Share resolution data with later stages.

Non-responsibilities:

* Semantic validation.
* Execution.
* Compilation.
* Tokenization.

The resolver answers the question:

> Which parsed declarations does each name refer to?

Implementation notes:

* The resolver is a cheap indexing pass over parser output.
* It does not copy parser-owned flat syntax storage.
* It builds lookup tables once so the validator and compiler do not rebuild temporary maps independently.

## Compiler

Input:

```text
Resolved syntax
```

Output:

```text
Compiled configuration
```

Responsibilities:

* Prepare token definitions for execution.
* Prepare rule definitions for execution.
* Produce efficient runtime structures.

Non-responsibilities:

* Reading source files.
* Tokenization.
* Rule evaluation.
* Diagnostic generation.

The compiler answers the question:

> Can this configuration be executed efficiently?

Implementation note:

* The compiler assumes semantic validation has already succeeded and reuses resolver output for declaration lookups.

## Scanner

Input:

```text
Compiled token definitions
Source file
```

Output:

```text
Token stream
```

Responsibilities:

* Read source text.
* Match token definitions.
* Produce tokens.
* Preserve source spans.
* Report tokenization failures.

Non-responsibilities:

* Rule evaluation.
* Diagnostics.
* Semantic analysis.

The scanner answers the question:

> What tokens exist in this source file?

Implementation note:

* Spot's current runtime scanner is implemented as an Nfa-backed matcher.
* The scanner design and construction details are documented in `docs/nfa.md`.
* The current rule engine consumes scanner output as a stream rather than materializing the full token stream first.

## Rule Engine

Input:

```text
Compiled rules
Token stream
```

Output:

```text
Diagnostics
```

Responsibilities:

* Evaluate rules.
* Apply rule conditions.
* Produce diagnostics.

Non-responsibilities:

* Tokenization.
* DSL parsing.
* Validation.
* Source file loading.

The rule engine answers the question:

> Which diagnostics should be produced?

# Pipeline Data

The architecture revolves around a small number of conceptual data structures.
These concepts are architectural concepts rather than implementation requirements.
The existence of a concept in this document does not imply the existence of a specific Go type.

## Syntax Tree

Represents parsed DSL syntax.

Properties:

* Preserves source locations.
* Mirrors DSL syntax.
* May contain invalid references.
* May contain invalid semantics.
* May be stored in flat parser-owned arenas and section-indexed slices.
* Is optimized for compact storage and fast sequential traversal rather than in-place tree mutation.

The syntax tree exists solely to represent parsed syntax.

## Configuration

Represents validated DSL data.

Properties:

* Semantically valid.
* References resolved.

## Resolved Syntax

Represents parsed DSL syntax together with reusable declaration lookups.

Properties:

* Preserves source locations.
* Reuses parser-owned flat syntax storage.
* Maps names to first declaration indices.
* May still contain semantic errors.
* Exists to avoid rebuilding the same lookup structures in multiple later stages.
* Executable after compilation.

A configuration represents a valid Spot program.

## Compiled Configuration

Represents executable analysis structures.

Properties:

* Optimized for execution.
* Free from validation concerns.
* Reusable across many source files.

Compilation should happen once.
Execution should happen many times.

## Token

Represents scanner output.

Properties:

* Token kind.
* Token text.
* Source span.

Tokens should be treated as immutable after creation.

## Diagnostic

Represents a reported problem.

Properties:

* Severity.
* Message.
* Source span.

Diagnostics are the primary output of analysis.

# Source Locations

Source locations are important throughout the entire pipeline.
Every stage should preserve source location information whenever practical.
Locations should be represented using byte offsets.
Line and column information should be derived when rendering diagnostics rather than stored throughout the pipeline.
This minimizes memory usage and avoids repeated bookkeeping.

# Validation Philosophy

Validation should occur as early as possible.
Invalid configurations should never reach execution stages.
The compiler, scanner, and rule engine should be able to assume that validated inputs satisfy all DSL requirements.
This reduces defensive programming and improves performance.

Bad:

```text
Parser
    ↓
Compiler
    ↓
Runtime validation
```

Good:

```text
Parser
    ↓
Validator
    ↓
Compiler
```

Errors should be discovered before execution whenever possible.

# Execution Philosophy

Configuration processing and source analysis are separate concerns.

The DSL should be:

1. Parsed
2. Validated
3. Compiled

once.

The resulting compiled configuration should then be reused across many source files.

Example:

```text
Parse DSL
Validate DSL
Compile DSL

Analyze file A
Analyze file B
Analyze file C
Analyze file D
```

Expensive configuration work should not be repeated unnecessarily.

# Error Ownership

Errors should be reported by the stage that discovers them.
Ownership should remain clear and stable.

| Problem                      | Owner       |
| ---------------------------- | ----------- |
| Missing closing brace        | Parser      |
| Unexpected token             | Parser      |
| Unknown definition reference | Validator   |
| Unknown token reference      | Validator   |
| Recursive definition         | Validator   |
| Invalid rule                 | Validator   |
| Scanner match failure        | Scanner     |
| Rule evaluation failure      | Rule Engine |

A stage should not report errors belonging to another stage.

# Dependency Rules

Dependencies should flow in the same direction as the pipeline.

Good:

```text
Parser
    ↓
Validator
    ↓
Compiler
    ↓
Scanner
    ↓
Rule Engine
```

Bad:

```text
Parser
    ↓
Validator
    ↑
Compiler
```

Circular dependencies are not permitted.
Earlier stages must never depend on later stages.

# Package Independence

Architecture and package structure are separate concerns.
This document defines responsibilities and boundaries.
It does not define package names, package counts, or directory layouts.
Packages should emerge naturally from implementation requirements.
Avoid designing package structures before there is sufficient implementation experience.

# Future Evolution

New capabilities should be introduced by extending existing stages whenever practical.
A new stage should only be introduced when it owns a genuinely distinct responsibility.
Before introducing a new stage, verify:

1. The responsibility cannot reasonably belong to an existing stage.
2. The new stage improves clarity.
3. The new stage improves maintainability.
4. The added complexity is justified.

Architecture should evolve slowly.
Simplicity is preferred over flexibility.
A stage should not be introduced merely because it may be useful in the future.
