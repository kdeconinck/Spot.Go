# Vision

Spot aims to provide a fast, deterministic, language-agnostic static analysis engine driven by a small DSL.
The project prioritizes:

* Predictable performance.
* Explicit behavior.
* Precise diagnostics.
* Minimal dependencies.
* Simple architecture.

The roadmap focuses on delivering a complete and usable analysis pipeline before expanding functionality.

# Current Focus

The immediate goal is to establish a complete end-to-end pipeline:

```text
DSL
    ↓
Parser
    ↓
Validator
    ↓
Compiler
    ↓
Scanner
    ↓
Rule Engine
    ↓
Diagnostics
```

The first usable version should prove that this pipeline works reliably from configuration input to diagnostic output.

# Milestone 1: DSL Parsing

Goal:

Parse DSL files into a syntax tree.

Capabilities:

* Parse all supported top-level sections.
* Parse definitions.
* Parse tokens.
* Parse rules.
* Preserve source locations.
* Report syntax errors.

Success criteria:

* Valid DSL files parse successfully.
* Invalid DSL files produce useful syntax diagnostics.
* Source locations are preserved.

# Milestone 2: Validation

Goal:

Validate parsed DSL files.

Capabilities:

* Validate references.
* Validate naming rules.
* Validate token definitions.
* Validate rule definitions.
* Detect invalid recursion.
* Produce validation diagnostics.

Success criteria:

* Invalid configurations cannot proceed to compilation.
* Validation errors are reported with accurate source locations.

# Milestone 3: Compilation

Goal:

Convert validated configurations into executable runtime structures.

Capabilities:

* Resolve definitions.
* Prepare token definitions.
* Prepare rules.
* Produce reusable compiled configurations.

Success criteria:

* Compilation occurs once.
* Compiled configurations can be reused across many source files.

# Milestone 4: Scanning

Goal:

Convert source text into tokens.

Capabilities:

* Match token definitions.
* Produce token streams.
* Support skipped tokens.
* Preserve source spans.
* Report scanner failures.

Success criteria:

* Source text can be tokenized deterministically.
* Token precedence rules behave as documented.

# Milestone 5: Rule Evaluation

Goal:

Evaluate rules against token streams.

Capabilities:

* Match tokens.
* Evaluate conditions.
* Produce diagnostics.
* Support all documented severities.

Success criteria:

* Rules execute correctly.
* Diagnostics are produced at the expected locations.

# Milestone 6: Diagnostic Rendering

Goal:

Present diagnostics to users.

Capabilities:

* Render severity.
* Render messages.
* Render source locations.
* Render source snippets when available.

Success criteria:

* Diagnostics are understandable and actionable.
* Source locations are accurate.

# Milestone 7: Command-Line Interface

Goal:

Provide a usable command-line experience.

Capabilities:

* Load DSL files.
* Analyze directories.
* Analyze individual files.
* Display diagnostics.
* Return meaningful exit codes.

Success criteria:

* Spot can be used without writing additional code.
* Analysis can be executed from the command line.

# Future Candidates

The following capabilities may be considered in the future.

They are not currently planned work.

## Multi-Token Rules

Rules that match sequences of tokens instead of a single token.

## Additional Rule Expressions

More expressive rule conditions.

## Custom Diagnostic Codes

User-defined diagnostic identifiers.

## Automatic Fixes

Rules capable of proposing source modifications.

## Cross-File Analysis

Analysis involving multiple files.

## AST-Based Analysis

Language-specific parsing and analysis.

## Incremental Analysis

Reusing previous analysis results.

## Language Packs

Reusable DSL configurations for common languages.

# Long-Term Vision

The current Spot engine is token-based.
It tokenizes source text, evaluates rules over those tokens, and emits diagnostics.
That is the first vertical slice.

The long-term ambition is broader:

* Spot should remain DSL-driven.
* The DSL should remain the source of truth.
* The engine should be able to analyze many kinds of text formats, including programming languages, configuration files, and documentation formats.

If Spot grows toward Sonar-style analysis, the DSL will likely need to describe more than tokens.
The likely long-term shape is a layered DSL:

1. Lexical layer.
   Tokens describe how source text is split into lexical units.
2. Structural layer.
   Nodes describe how tokens form larger syntactic constructs.
3. Binding layer.
   Symbols, scopes, declarations, and references describe how structure gains semantic meaning.
4. Rule layer.
   Rules describe diagnostics over tokens, nodes, symbols, or derived flow facts.

Conceptually, that future pipeline would look like:

```text
DSL
    ↓
Token Definitions
    ↓
Node Definitions
    ↓
Binding Definitions
    ↓
Compiled Analysis Model
    ↓
Source Text
    ↓
Tokenization
    ↓
Parsing
    ↓
Binding / Semantic Model
    ↓
Rule Evaluation
    ↓
Diagnostics
```

In that model, the DSL remains the only user-provided input.
However, the engine is still responsible for compiling the DSL into runtime machinery such as:

* scanners
* parsers
* symbol tables
* reference resolution
* control-flow or data-flow structures

The future rule model may therefore evolve beyond today's single-token rules.
Possible future rule targets include:

* tokens
* token sequences
* syntax nodes
* symbols
* references
* flow facts

Examples of rules this vision aims to support eventually:

* structural rules, such as ensuring a `using` directive appears inside a namespace declaration
* symbol rules, such as detecting unused parameters
* semantic rules, such as detecting broken `async` / `await` patterns

This is a long-term direction, not a current milestone.
The engine should only move toward it incrementally, with benchmark evidence and the simplest design that satisfies each new requirement.

# Non-Goals

The following are intentionally outside the current roadmap:

* Compiler construction.
* Type checking.
* Semantic language analysis.
* Language servers.
* IDE integrations.
* Build system integration.
* Code generation.

These may be reconsidered in the future if they align with the project goals.

# Roadmap Philosophy

The roadmap defines capabilities rather than implementation tasks.
Implementation details should emerge during development.
Capabilities should be completed vertically.

Prefer:

```text
DSL feature
    ↓
Validation
    ↓
Compilation
    ↓
Execution
    ↓
Tests
```

over partially implementing many unrelated capabilities.
The simplest complete solution is preferred over the most flexible solution.
