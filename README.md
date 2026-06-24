# Spot

Spot is a high-performance, language-agnostic static analysis engine written in Go.
Spot analyzes source files by transforming text into tokens and evaluating rules over the resulting token stream.
The project emphasizes explicit parsing, predictable performance, precise diagnostics, and a small, well-defined DSL.

Unlike many lightweight linters, Spot does not rely on regular expressions as its primary analysis mechanism.
Instead, it uses dedicated parsing and scanning stages with clearly defined responsibilities and data flow.

> **Note:** README DSL snippets are illustrative. The DSL specification in `docs/dsl.md` is the authoritative source for
syntax, semantics, and validation rules.

## Project Intent

Spot exists to provide a fast, deterministic, and understandable foundation for static analysis.

The project prioritizes:

* Predictable performance.
* Explicit parsing and scanning.
* Minimal dependency usage.
* Precise source locations.
* Clear ownership of responsibilities.
* Simple and maintainable implementation.
* Language-agnostic analysis through configuration rather than hardcoded language support.

## Design Principles

1. Explicit is preferred over implicit.
2. Parsing is preferred over regular-expression matching.
3. Data flows in a single direction through the analysis pipeline.
4. Each stage owns one responsibility.
5. Configuration is validated before execution.
6. Diagnostics must reference precise source locations.
7. Performance improvements must not significantly reduce clarity.
8. Dependencies should be introduced only when they provide substantial value.
9. Simple data structures are preferred over abstraction-heavy designs.

## Documentation

The repository documentation is organized as follows:

| Document               | Purpose                                                |
| ---------------------- | ------------------------------------------------------ |
| `README.md`            | Project overview and high-level concepts.              |
| `AGENTS.md`            | Instructions for contributors and AI agents.           |
| `docs/dsl.md`          | DSL syntax, semantics, validation rules, and examples. |
| `docs/architecture.md` | Pipeline stages, responsibilities, and data flow.      |
| `docs/roadmap.md`      | Milestones, priorities, and future work.               |

When documentation conflicts:

1. `docs/dsl.md` is authoritative for DSL behavior.
2. `docs/architecture.md` is authoritative for architecture and stage ownership.
3. `README.md` provides a high-level overview and may omit implementation details.

## Pipeline

Spot is designed as a linear processing pipeline.

```text
DSL Configuration
        ↓
Parser
        ↓
Validated Configuration
        ↓
Compiler
        ↓
Compiled Definitions
        ↓
Scanner
        ↓
Token Stream
        ↓
Rule Engine
        ↓
Diagnostics
```

Each stage consumes explicit input and produces explicit output for the next stage.

## DSL Overview

The DSL is organized around reusable definitions, token declarations, and analysis rules.

```spot
definitions {
    letter = 'a'..'z' | 'A'..'Z'
    digit = '0'..'9'
}

tokens {
    Whitespace = (' ' | '\t' | '\n' | '\r')+ skip
    Identifier = letter (letter | digit)*
}

rules {
    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warning at Identifier "Public identifier found"
    }
}
```

### Definitions

Definitions describe reusable character-level expressions.

```spot
definitions {
    letter = 'a'..'z' | 'A'..'Z'
    digit = '0'..'9'
}
```

Definitions can be referenced by token declarations and other definitions.

### Tokens

Tokens describe how source text is transformed into a stream of tokens.

```spot
tokens {
    Identifier = letter (letter | digit)*
}
```

The scanner evaluates token definitions and emits tokens with associated source locations.

### Rules

Rules inspect tokens and emit diagnostics when conditions are met.

```spot
rules {
    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warning at Identifier "Public identifier found"
    }
}
```

Rules operate on the token stream rather than directly on source text.

## Current Capabilities

The project currently focuses on the following capabilities:

* DSL parsing.
* Configuration validation.
* Definition compilation.
* Token definition compilation.
* Source scanning.
* Token stream generation.
* Token-based rule evaluation.
* Diagnostic generation and rendering.

Advanced features such as semantic analysis, AST construction, and cross-file reasoning are outside the current scope.

## Agent Guidance

When contributing to Spot:

### Dependencies

* Prefer the Go standard library.
* Avoid third-party dependencies unless there is a strong technical justification.
* Do not introduce frameworks for problems that can be solved with standard Go.

### Architecture

* Preserve the pipeline architecture.
* Keep stage responsibilities clearly separated.
* Prefer composition over deep abstraction layers.
* Avoid premature generalization.
* Avoid reflection unless explicitly required.

### Performance

* Use byte offsets for source locations.
* Avoid unnecessary allocations.
* Avoid unnecessary string copying.
* Prefer measurable optimizations over speculative optimizations.
* Keep memory usage predictable.

### Code Style

* Keep packages focused and cohesive.
* Prefer straightforward code over clever code.
* Favor explicit data flow.
* Keep public APIs small.
* Write code that is easy to reason about and debug.

### Documentation

* Update documentation when behavior changes.
* Treat `docs/dsl.md` as the source of truth for DSL behavior.
* Keep terminology consistent across the repository.

## Terminology

| Term             | Meaning                                                                                |
| ---------------- | -------------------------------------------------------------------------------------- |
| Definition       | A reusable character-level expression.                                                 |
| Token Definition | A declaration describing how source text becomes a token.                              |
| Token            | A scanner output with associated source location information.                          |
| Scanner          | Component responsible for converting source text into tokens.                          |
| Rule             | Analysis logic executed against a token stream.                                        |
| Diagnostic       | A warning, error, or informational message produced by analysis.                       |
| Span             | A byte range within a source file.                                                     |
| Configuration    | The validated representation of a DSL file.                                            |
| Compiler         | Component that converts validated DSL structures into executable analysis definitions. |

## Project Documents

* `docs/dsl.md` — DSL syntax, semantics, validation rules, and examples.
* `docs/architecture.md` — Architecture, package ownership, data flow, and implementation decisions.
* `docs/roadmap.md` — Planned milestones, future capabilities, and project direction.

## Status

Spot is under active development.

The current objective is to establish a complete end-to-end analysis pipeline from DSL configuration to diagnostic
output while maintaining a simple, understandable, and high-performance implementation.
