# Rust Plan

This document describes a clean Rust implementation plan for Spot.
It is intentionally designed from product requirements, not from the current Go code layout.

The plan has four goals:

1. Keep feedback fast.
2. Keep each commit small and reviewable.
3. Keep the architecture data-oriented and performance-focused.
4. Reach the same functional place as Spot is targeting now: DSL parsing, validation, compilation, scanning, syntax-tree construction, rule evaluation, diagnostics, and AST printing.

The plan assumes there is no time pressure and that learning Rust well is a real goal.
That changes the tradeoff: we can afford to build the system carefully instead of racing toward a port.

## Core Position

Rust is a good fit for Spot because Spot is fundamentally a systems project:

* it transforms text through several explicit stages
* it cares about spans and diagnostics
* it benefits from compact memory layout
* it has hot paths where allocations and cache locality matter
* it should be deterministic and predictable

At the same time, Rust can slow you down if you start with too much abstraction.
This plan therefore prefers:

* flat storage over pointer-heavy graphs
* explicit indices over shared ownership
* small concrete types over traits
* one-way pipeline stages
* benchmarks at the hot boundaries

## Guiding Design

### 1. Build vertical slices early

Do not spend months building infrastructure before the tool does anything useful.
The first milestones should produce:

* a parsed DSL
* a compiled token matcher
* a scanner
* one executable rule
* one CLI command

That gives fast product feedback and fast Rust feedback.

### 2. Use byte offsets everywhere

All source locations should be half-open byte spans:

* `start` is inclusive
* `end` is exclusive

This keeps the core pipeline fast and simple.
Line and column mapping should be a separate layer used for rendering diagnostics.

### 3. Prefer flat arenas and index-based graphs

The parser AST, compiled expressions, runtime syntax tree, and query programs should all use flat vectors with integer indices.

That buys:

* compact memory layout
* low allocation churn
* predictable traversal cost
* easy serialization if wanted later
* no pointer chasing

### 4. Fail fast where failure invalidates the stage

Parsing the DSL should stop at the first syntax error.
Validation should collect semantic diagnostics because that is useful to users.
Scanning source text should stop at the first tokenization failure for a file.
Syntax parsing should stop at the first root mismatch for a file.

### 5. Separate what is stable from what is display-oriented

Stable internal forms:

* byte offsets
* indices
* compact enums
* interned strings

Display-oriented forms:

* line and column numbers
* rendered source snippets
* pretty-printed trees

This prevents presentation concerns from polluting hot code paths.

## Recommended Workspace Layout

Start with a Cargo workspace and a small set of crates:

* `crates/spot-source`
  Purpose: source text, spans, line mapping, file ids.
* `crates/spot-diagnostic`
  Purpose: diagnostics and rendering.
* `crates/spot-dsl`
  Purpose: DSL lexer, parser, AST, resolver, validator.
* `crates/spot-compile`
  Purpose: compile validated DSL into runtime programs.
* `crates/spot-runtime`
  Purpose: scanner, syntax parser, query engine, rule execution.
* `crates/spot-cli`
  Purpose: command-line entry point.

This is enough separation to keep responsibilities clear, but not so much that the project becomes fragmented.

## Shared Type Conventions

Use these conventions from the start:

* `u32` for offsets and arena ids where practical
* `NonZeroU32` only when there is a measured reason
* `Vec<T>` as the default storage
* `#[repr(u8)]` or `#[repr(u16)]` on dense enums used in hot paths
* newtype ids for clarity:
  * `FileId`
  * `NodeId`
  * `ExprId`
  * `TokenId`
  * `RuleId`
* half-open spans: `[start, end)`

Do not introduce lifetimes into the whole pipeline unless they earn their keep.
Owned source text plus indices is usually the simpler and faster choice for this kind of tool.

## Benchmark Strategy

Add benchmarks early, but only for hot stages:

* DSL lexing
* DSL parsing
* validation
* token compilation
* scanning
* syntax parsing
* rule evaluation
* full engine execution

Use one benchmark harness consistently.
`criterion` as a dev-dependency is justified here because performance is a product feature.

## End-State Feature Target

The Rust implementation should end with support for:

* scope, definitions, tokens, syntax, and rules sections
* exact spans on DSL declarations
* token expressions with concatenation, alternation, repetition, groups, strings, chars, ranges, references, skip, and fallback
* syntax nodes with block-style syntax only
* named captures and repeated fields
* `oneOf` and `any`
* exactly one syntax root
* token scanning
* runtime syntax-tree construction
* token-based and syntax-based rules
* selector-style rule matching
* simple ancestry and adjacency relationships
* basic `where` conditions on text, numeric properties, and gaps
* CLI execution
* AST printing

The plan below reaches that target in small commits.

---

# Commit Plan

## Commit 1: Create the Rust workspace

Add:

* workspace `Cargo.toml`
* crates listed above
* top-level `README` note for the Rust implementation

Purpose:

* establish structure
* make every later commit land in the right place

Keep all crates compiling with placeholder APIs.

## Commit 2: Introduce `Position`

Add in `spot-source`:

```rust
pub struct Position {
    pub offset: u32,
}
```

Meaning:

* `Position` represents one byte offset into a source file
* it does not know line or column
* it does not know file id by itself

Behavior:

* cheap to copy
* comparable and sortable
* no hidden conversions

Why this type exists:

* raw `u32` is too anonymous
* source locations become self-documenting
* later APIs can clearly say whether they want a position or a span

Tests:

* construction
* ordering
* debug formatting if customized

## Commit 3: Introduce `Span`

Add in `spot-source`:

```rust
pub struct Span {
    pub start: Position,
    pub end: Position,
}
```

Meaning:

* `Span` is a half-open range `[start, end)`
* it can represent empty spans

Behavior:

* `len()`
* `is_empty()`
* `contains(Position)`
* `join(Span) -> Span`

Rules:

* `start <= end`
* constructors should enforce or debug-assert the invariant

Why this type exists:

* spans are the currency of diagnostics
* they are also the currency of token and node boundaries

Tests:

* empty span
* non-empty span
* joining spans
* containment

## Commit 4: Introduce `LineColumn`, `LineMap`, and `SourceText`

Add in `spot-source`:

* `LineColumn { line: u32, column: u32 }`
* `LineMap`
* `SourceText`

Meaning:

* `SourceText` owns the file contents
* `LineMap` stores line-start byte offsets
* `LineColumn` is only for human-readable rendering

Behavior:

* `LineMap::new(text: &str)`
* `line_column(offset: u32) -> LineColumn`
* `line_span(line: u32) -> Span`
* `slice(span) -> &str` on `SourceText`

Why these types exist:

* the core engine should work in offsets
* humans need line and column
* line mapping should not be recomputed per diagnostic

Performance rule:

* `LineMap` is built once per file
* lookup should be binary search over line starts

Tests:

* empty file
* single line
* multiple lines
* trailing newline
* span slicing

## Commit 5: Introduce `FileId` and `SourceFile`

Add in `spot-source`:

* `FileId`
* `SourceFile`

`SourceFile` should contain:

* `id: FileId`
* `path: PathBuf`
* `text: String`
* `line_map: LineMap`

Why:

* the runtime will analyze many files
* diagnostics need both file identity and spans

Tests:

* create a source file
* slice by span
* convert offset to line and column

## Commit 6: Introduce diagnostics

Add in `spot-diagnostic`:

* `Severity`
* `Diagnostic`
* `Label`
* basic render function

Keep it simple:

* one primary span per diagnostic is enough at first
* secondary labels can come later if needed

`Diagnostic` should contain:

* file id
* span
* severity
* message

Tests:

* display formatting
* rendering with line and column
* rendering a line snippet

## Commit 7: Add the CLI skeleton

Add in `spot-cli`:

* parse arguments
* load one DSL file
* load one directory path
* print placeholder output

Do not implement the engine yet.
This commit is just for command shape and quick feedback.

Support from the start:

* `spot <dsl-file> <path>`
* `spot --print-ast <dsl-file> <path>`

## Commit 8: Add string interning

Add a simple `StringInterner` in `spot-dsl` or a tiny shared crate if clearly justified.

Purpose:

* user-defined identifiers repeat often
* interned ids make AST and compiled forms smaller

Behavior:

* `intern(&str) -> Symbol`
* `resolve(Symbol) -> &str`

Do not over-engineer this.
A simple `HashMap<String, Symbol>` plus `Vec<String>` is enough initially.

Tests:

* same string gets same symbol
* different strings get different symbols

## Commit 9: Define DSL token kinds

Add:

* `TokenKind`
* reserved keyword table
* punctuation and operator kinds

Include from the start:

* section keywords
* `node`
* `oneOf`
* `any`
* `skip`
* `fallback`
* `match`
* `where`
* `report`
* `inside`
* `outside`
* selector punctuation
* literals

Keep the enum dense and explicit.

## Commit 10: Implement the DSL lexer

Responsibilities:

* lex identifiers
* lex keywords
* lex integer, character, and string literals
* lex punctuation and operators
* skip whitespace and comments
* produce exact spans
* fail fast on malformed tokens

Output:

* flat `Vec<Token>`

`Token` should contain:

* kind
* span
* optional symbol id or literal payload reference

Tests:

* all token kinds
* escaped literals
* comments
* malformed strings and chars

Benchmarks:

* small DSL
* large happy-path DSL

## Commit 11: Introduce the parser-facing AST ids and arenas

Add in `spot-dsl`:

* `ExprId`
* `SyntaxExprId`
* `RuleExprId`
* `Document`
* flat arenas for parsed expressions

Design:

* no boxed recursive enums for the main AST
* store nodes in `Vec<NodeRecord>`
* relationships are stored as child index ranges

This is the key architectural move.
The parser should target the final storage model from the start.

## Commit 12: Parse the top-level document skeleton

Implement:

* top-level section loop
* section uniqueness checks at parse time only if syntactic
* fail-fast syntax errors

At this point, parse empty sections:

* `scope {}`
* `definitions {}`
* `tokens {}`
* `syntax {}`
* `rules {}`

Tests:

* empty document
* each empty section
* duplicate braces and missing braces

## Commit 13: Parse the `scope` section

Implement:

* `include`
* `exclude`

Storage:

* flat `Vec<ScopeEntry>`

Tests:

* multiple includes and excludes
* missing string literal
* unexpected token inside section

## Commit 14: Parse definition expressions

Implement character-level expressions:

* character literal
* string literal if definitions allow it in final grammar
* range
* reference
* group
* concatenation
* alternation
* repetition: `?`, `*`, `+`

This is the first real expression parser.

Keep it explicit:

* parse primary
* parse postfix repetition
* parse concatenation
* parse alternation

Tests:

* all happy-path constructs
* precedence
* spans
* syntax errors

## Commit 15: Parse the `definitions` section

Implement:

* named definition declarations
* expression storage in the flat arena

Tests:

* multiple definitions
* spans
* invalid declaration forms

## Commit 16: Parse token expressions and token flags

Implement token declarations:

* token name
* expression
* optional `skip`
* optional `fallback`

Design choice:

* represent token flags explicitly, not as synthetic expression nodes

Tests:

* skip token
* fallback token
* strings, ranges, references, groups, operators
* invalid flag placement

## Commit 17: Parse the `tokens` section

This commit wires token declarations into the document.

Add tests for:

* multiple tokens
* mixed token forms
* spans
* fail-fast behavior

## Commit 18: Parse the `syntax` section with block-style nodes only

Do not support compact `node X = ...`.
Go directly to the cleaner syntax.

Implement:

* `node Name { ... }`
* unnamed entries
* named captures: `name: Child`
* repeated fields: `items*: Child`
* optional fields: `value?: Child`
* `oneOf { ... }`
* `any`

Storage:

* separate syntax-expression arena from token-expression arena if that keeps the model clearer

Tests:

* one-node syntax
* nested nodes
* named captures
* repeated fields
* `oneOf`
* `any`
* syntax errors and spans

## Commit 19: Parse legacy rule blocks

Implement first:

* `rule Name { ... }`
* `match Identifier`
* `match node NodeName`
* optional `where`
* `report severity at Target "message"`

Why start here:

* it gives a small, obvious execution path
* it is easier to validate and compile than query syntax

Tests:

* token match rules
* node match rules
* where clauses
* reports

## Commit 20: Introduce the selector-rule AST

Add parsed structures for the cleaner rule language:

```spot
warn "Message"
    : Selector
    where ...
```

Include:

* selector sequence
* negation
* ancestry markers
* bound names like `left` and `right`

Do not execute it yet.
This commit only parses and stores it.

Why:

* this is likely the long-term rule surface
* it should exist as a first-class model, not as an ad hoc extension

## Commit 21: Parse selector-style rules

Implement parsing for:

* severity-first selector rules
* selector after `:`
* optional `where`

Keep both rule forms temporarily if needed for migration experiments, but choose one canonical form in docs.
If you want the cleanest final system, make selector rules the only public rule surface and keep old block rules out.

My recommendation:

* support selector rules as the target surface
* only keep block rules if they are necessary as a temporary bootstrap

## Commit 22: Introduce the resolver

Responsibilities:

* build name lookup tables
* preserve declaration order
* expose first declaration for duplicate diagnostics

Lookups needed:

* definitions by name
* tokens by name
* syntax nodes by name
* rules by name

This should be a cheap indexing pass over parser output.

Tests:

* successful resolution
* duplicates found by name

## Commit 23: Add validator infrastructure

Add:

* `ValidationError`
* `ValidationResult`
* semantic diagnostic collection

This is where behavior changes from fail-fast to collect-many.

The validator should walk:

* scope
* definitions
* tokens
* syntax
* rules

## Commit 24: Validate section-level rules

Validate:

* section appears at most once
* required sections for execution
* at least one scope include

Keep section-order rules as style choices unless they are semantically required.

## Commit 25: Validate definitions

Validate:

* references exist
* cycles are rejected if they can match without consuming input incorrectly
* names are unique

Decide clearly whether recursive definitions are allowed.
For a first robust system, my recommendation is:

* reject recursive definitions unless there is a concrete need

That makes both reasoning and compilation much simpler.

## Commit 26: Validate tokens

Validate:

* token references resolve
* exactly one fallback token at most
* fallback token has no ordinary matcher body if that is the chosen design
* skip usage is valid
* token names are unique

Also validate tokenizer safety:

* no token expression may match empty input unless explicitly and safely designed for it

This matters a lot for scanner correctness.

## Commit 27: Validate syntax nodes

Validate:

* referenced tokens and nodes exist
* exactly one root syntax node exists
* repeated syntax expressions do not match empty input
* node names are unique
* field names are unique within a node if that is the chosen rule

This is one of the most important correctness commits in the plan.

## Commit 28: Validate rules

Validate:

* all referenced tokens and nodes exist
* referenced properties are legal for the selected subject type
* `inside` and `outside` target node types only
* selector variables and captures are bound correctly
* `where` expressions are type-correct

At the end of this commit, the DSL should be semantically trustworthy.

## Commit 29: Define compiled token IR

In `spot-compile`, add the runtime program representation for token matching.

Prefer:

* flat instruction array
* explicit transitions
* no heap graph

This IR is the compiled form of definitions and tokens.
It should be built once and reused for many files.

## Commit 30: Compile definition and token expressions

Compile parsed expressions into token IR.

Include:

* literal char
* literal string
* range
* reference expansion
* concatenation
* alternation
* repetition

This is the first commit where the configuration becomes executable.

Tests:

* compile all happy-path forms
* compare compiled shapes where useful

Benchmarks:

* token compilation throughput

## Commit 31: Implement the scanner

In `spot-runtime`, implement source scanning.

Input:

* compiled token program
* source file text

Output:

* flat token stream

`ScannedToken` should contain:

* token kind or compiled token id
* span

Behavior:

* longest valid token wins
* tie-breaking rules must be explicit and documented
* skip tokens are omitted from the final stream
* fallback token consumes one unit when no other token matches

Tests:

* token precedence
* skip behavior
* fallback behavior
* source spans

Benchmarks:

* large valid source
* fallback-heavy source

## Commit 32: Implement diagnostic reporting for scanner failures

If scanning fails without a valid fallback:

* return a diagnostic with exact span
* show the unexpected source region

Keep this separate from the scanning core so the hot path remains clean.

## Commit 33: Define compiled syntax IR

In `spot-compile`, add compiled syntax-node programs.

Compile syntax declarations into a compact IR that the runtime syntax parser can execute.

Represent:

* node declarations
* token references
* node references
* captures
* repetitions
* alternations
* `any`

Use indices, not pointers.

## Commit 34: Define the runtime syntax tree model

In `spot-runtime`, add a flat tree:

* `Tree`
* `NodeRecord`
* `ChildEdge`

Recommended shape:

* `Tree.nodes: Vec<NodeRecord>`
* `Tree.edges: Vec<ChildEdge>`

`NodeRecord` should contain:

* node kind id
* token start index
* token end index
* first child edge
* child edge count

`ChildEdge` should contain:

* parent node id
* child node id
* optional field/capture symbol

Why this shape:

* dense storage
* stable traversal
* child labels without pointer graphs

## Commit 35: Implement the runtime syntax parser

Input:

* compiled syntax IR
* token stream

Output:

* runtime `Tree`

Behavior:

* parse from one root
* succeed only if the root consumes the full token stream
* support repeated fields and captures
* preserve child labels on edges

Tests:

* full happy-path tree
* named captures
* repeated fields
* `oneOf`
* `any`
* root mismatch

Benchmarks:

* full happy-path syntax parsing

## Commit 36: Add AST printing

Add pretty-printers for:

* parsed DSL document
* runtime syntax tree

The runtime tree printer should show:

* node name
* token span
* labeled child edges

This is important both for users and for your own development speed.

## Commit 37: Implement token-based rule execution

Start with the simpler execution path:

* match a token kind
* evaluate simple properties like `.text` and `.length`
* emit a diagnostic

This gives the first full useful engine:

DSL -> compile -> scan -> evaluate -> diagnostic

Benchmarks:

* token rule throughput

## Commit 38: Define query/selector IR

Add a compiled representation for syntax-oriented selectors.

Support initially:

* node-type match
* adjacency: `A + B`
* negation
* ancestry conditions like `inside` and `outside`
* bound names like `left` and `right`

This should compile into a flat program too.

## Commit 39: Implement selector matching on the runtime tree

Execute selector rules against the flat tree.

Start with:

* single-node selectors
* parent/ancestor checks
* adjacent sibling or adjacent root-member checks, depending on the chosen meaning

Document every relationship very clearly.
Selectors become hard to reason about if semantics stay fuzzy.

## Commit 40: Implement gap analysis

Add support for the concepts needed by rules such as:

* `gap.blankLines`

This likely needs a lightweight derived view over token or source spans between two matched nodes.

Keep it cheap:

* compute from source spans and `LineMap`
* avoid allocating intermediate strings

## Commit 41: Implement `where` evaluation for selector bindings

Support conditions like:

* `left.name.text > right.name.text`
* `gap.blankLines != 1`

This commit should define:

* property access
* supported value types
* comparison semantics

Recommended first value types:

* string
* integer
* boolean

## Commit 42: Integrate syntax-based rules into the engine

The engine pipeline now becomes:

1. scan source text into tokens
2. optionally build the runtime syntax tree if syntax rules exist
3. run token rules
4. run syntax rules
5. collect diagnostics

Do not always build the tree if the program does not need it.
That would be wasted work for token-only configurations.

## Commit 43: Add file collection and scope matching

Implement:

* include and exclude matching
* path normalization
* file traversal

Keep the glob engine intentionally small.
Do not pull in a heavy dependency unless a benchmark justifies it.

## Commit 44: Make the CLI fully operational

Implement end-to-end execution:

* load DSL
* parse
* resolve
* validate
* compile
* walk files
* analyze files
* print diagnostics
* return non-zero on errors

Support:

* `--print-ast`

When enabled, print the runtime syntax tree for each successfully parsed file before or alongside diagnostics.

## Commit 45: Add golden end-to-end examples

Add examples for:

* a Go-like file checked by token and syntax rules
* a C# using-order example
* a file-scoped namespace example

Tests should run the CLI against example input and verify:

* diagnostics
* AST printing output where relevant

## Commit 46: Add parser benchmarks

Benchmark:

* DSL lexer
* DSL parser
* validator

Use complete happy-path inputs, not synthetic trivial cases only.

## Commit 47: Add runtime benchmarks

Benchmark:

* scanner
* syntax parser
* token rule engine
* selector rule engine
* full engine

At this point you finally have data on whether the Rust architecture is paying off.

## Commit 48: Tune storage using measurements

Only now start tuning hot storage details, for example:

* exact `Vec` reservations in compiler and runtime builders
* enum representation size
* field ordering in hot structs
* reducing duplicate symbols
* avoiding temporary allocations during `where` evaluation

This commit should be benchmark-driven, not aesthetic.

## Commit 49: Simplify any design that proved unnecessary

This is an important cleanup commit.

Remove:

* unused abstractions
* over-general enums
* redundant ids
* traits that bought nothing
* duplicate IR concepts

Rust code tends to drift into cleverness if left unchecked.
This commit keeps the project honest.

## Commit 50: Stabilize the public model and document the architecture

Write:

* DSL guide
* architecture guide
* benchmark guide
* contributor guide

Document especially:

* why flat arrays are used
* why spans are byte-based
* why the syntax tree stores labeled child edges
* which stages fail fast and which collect diagnostics

At this point the Rust implementation is a serious replacement candidate.

---

# Important Design Notes

## Why `Position`, `Span`, `LineMap`, and `LineColumn` matter

These types are worth introducing early because they separate concerns correctly.

### `Position`

Represents one exact byte offset.
It is the smallest meaningful source location unit.

### `Span`

Represents a region of text.
Everything user-visible eventually points to a span:

* lexer tokens
* parser nodes
* diagnostics
* scanned tokens
* runtime syntax nodes

### `LineMap`

Maps byte offsets to display coordinates.
This is not needed for parsing or execution.
It is needed for rendering useful diagnostics.

### `LineColumn`

Represents the human-readable form of a location.
This should never drive parsing logic.
It exists for output only.

That split keeps the engine fast and the UX good.

## Why flat arrays are the right default in Rust too

Rust gives you ownership safety, but it does not change hardware reality.
Pointer-rich trees still cost:

* more allocations
* worse locality
* more indirection
* more complex lifetimes if references are stored

Flat vectors with ids are better for Spot because the pipeline is naturally batch-oriented:

* parse one DSL file
* compile once
* analyze many files

That shape maps well to arenas and index-based relationships.

## What not to do

Avoid these traps:

* do not model the AST as `Box<Expr>` everywhere
* do not use `Rc<RefCell<_>>`
* do not introduce traits for every stage
* do not use parser generators until the hand-written parser clearly becomes the bottleneck
* do not optimize based on gut feeling alone

## What I would personally choose

If I were rebuilding Spot in Rust for both learning and performance, I would choose this exact order:

1. source and diagnostics
2. lexer
3. parser
4. validator
5. token compiler
6. scanner
7. token rules
8. syntax compiler
9. runtime syntax tree
10. selector rules
11. benchmarks and tuning

That order gives the fastest meaningful feedback while still leading to a serious architecture.

## Final Recommendation

Do not treat the Rust rewrite as a translation exercise.
Treat it as a redesign with strict product parity goals and measurable performance goals.

The rewrite is worth it if:

* you want to learn Rust deeply
* you want a data-oriented architecture
* you are willing to benchmark honestly
* you are willing to remove abstractions that Rust makes tempting

The rewrite is not worth it if:

* you mainly want to ship quickly
* you will rewrite everything before getting one working vertical slice
* you will assume Rust automatically makes weak designs fast

If done with discipline, Rust can make Spot both faster and conceptually stronger.
