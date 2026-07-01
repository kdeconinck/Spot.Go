  ## Summary

  Create GO-REWRITE-PLAN.md as the authoritative clean-slate Go rewrite plan for Spot.

  The document must describe a language-agnostic static analyzer and formatter driven entirely by a DSL. It must avoid magic constructs. Every feature must have clear syntax, semantics, data flow, runtime representation, and test strategy.

  The implementation starts extremely small: a CLI that discovers files and logs them, then a tiny DSL lexer with only enough tokens to parse files { include "..." }, then gradually expands. Every commit must compile, test, and produce observable behavior.

  ## Core Architecture To Document

  The document must define the full flow explicitly:

  DSL file text
    -> DSL lexer
    -> DSL tokens
    -> DSL parser
    -> parsed DSL document
    -> resolver
    -> validator
    -> compiler
    -> compiled program

  source file text
    -> source scanner
    -> source tokens
    -> syntax parser
    -> runtime syntax tree
    -> query engine
    -> diagnostics
    -> formatter query engine
    -> text edits
    -> formatted output or changed files

  The document must explain each stage:

  - DSL lexer: converts Spot DSL text into DSL tokens.
  - DSL parser: converts DSL tokens into a parsed document.
  - Resolver: builds name indexes without validating semantics.
  - Validator: checks semantic correctness and collects errors.
  - Compiler: lowers names and expressions to dense runtime IDs.
  - Source scanner: converts analyzed source text into language tokens using the DSL-defined token rules.
  - Syntax parser: converts source tokens into a runtime syntax tree using the DSL-defined syntax model.
  - Query engine: evaluates DSL-defined structural queries against tokens or syntax nodes.
  - Formatter: runs formatting queries and emits deterministic text edits.

  ## Performance Design To Document

  The plan must explicitly describe the scanner strategy.

  Initial scanner:

  - Use Thompson NFA simulation for correctness and implementation speed.
  - Compile token expressions into an NFA once per DSL program.
  - Simulate active state sets over source bytes/runes.
  - Use longest match.
  - Break ties by token declaration order.
  - Skip tokens are recognized but not emitted.
  - Fallback token consumes exactly one rune when no normal token matches.

  Later optimization:

  - Add a DFA cache only after NFA benchmarks identify scanner simulation as a bottleneck.
  - Cache active-state-set transitions by input class.
  - Keep Thompson NFA as the canonical implementation because it is simpler and easier to validate.
  - Keep DFA as an optional acceleration layer, not as the first implementation.

  The plan must define benchmark checkpoints:

  - before scanner implementation
  - after Thompson NFA scanner
  - after runtime syntax parser
  - after query engine
  - after formatter
  - after optional DFA cache

  The document must also define runtime data rules:

  - Parsed DSL may use straightforward structs for readability.
  - Compiled/runtime structures must use flat slices and dense integer IDs.
  - Runtime tokens store token ID and span, not copied text.
  - Runtime syntax tree stores nodes and edges in flat arrays.
  - Query programs compile to small instruction-like structures.
  - Formatting emits text edits with span and replacement; edits are sorted and overlap-checked before application.

  ## DSL Specification To Include

  The plan must define this DSL direction, but as a specification, not as unexplained examples.

  ### Files

  files {
      include "**/*.cs"
      exclude "bin/**"
  }

  Semantics:

  - include adds candidate paths.
  - exclude removes paths.
  - Patterns are matched against normalized slash-separated relative paths.
  - The first working CLI milestone only logs discovered files.

  ### Characters

  chars {
      lower = 'a'..'z'
      upper = 'A'..'Z'
      digit = '0'..'9'
      identifierStart = lower | upper | '_'
      identifierPart = identifierStart | digit
  }

  Semantics:

  - chars defines reusable character-level expressions.
  - Expressions are used by token definitions.
  - Character expressions can reference other character definitions.
  - Character expressions must not recursively reference themselves.

  Supported operators:

  - 'a'
  - 'a'..'z'
  - "text"
  - name
  - (expr)
  - expr expr
  - expr | expr
  - expr?
  - expr*
  - expr+

  ### Tokens

  tokens {
      Whitespace = (' ' | '\t' | '\n' | '\r')+ skip
      Identifier = identifierStart identifierPart*
      Using = "using"
      Unknown = fallback
  }

  Semantics:

  - Tokens define how analyzed source files become source tokens.
  - skip means recognized but not emitted.
  - fallback means consume one rune only when nothing else matches.
  - Normal token expressions must not match empty input.
  - At most one fallback token is allowed.

  ### Syntax

  syntax {
      root SourceFile

      node QualifiedName {
          head: Identifier
          tail*: QualifiedNameTail
      }

      node QualifiedNameTail {
          Dot
          part: Identifier
      }

      node UsingDirective {
          Using
          name: QualifiedName
          Semicolon
      }

      node SourceFile {
          members*: oneOf {
              UsingDirective
              any
          }
      }
  }

  Semantics:

  - root names the only file-level syntax node.
  - node defines one runtime syntax node type.
  - Bare entries match token or node references and do not create labeled edges.
  - name: QualifiedName creates a labeled edge.
  - tail*: QualifiedNameTail captures zero or more children under field tail.
  - oneOf is ordered alternation.
  - any consumes exactly one emitted source token and creates no modeled syntax node unless captured through a field.
  - Syntax repetition must not repeat something that can match empty input.

  ### Query Rules

  The document must define rules through a real query system.

  Example:

  rules {
      warn "Using directive is outside a namespace"
          when UsingDirective
          where not ancestor(NamespaceDeclaration)

      warn "Using directives must be sorted"
          when left: UsingDirective + right: UsingDirective
          where left.name.text > right.name.text
  }

  Selector semantics:

  - UsingDirective matches a syntax node type.
  - left: UsingDirective binds the matched node as left.
  - A + B matches adjacent sibling nodes in the same parent child list.
  - A > B matches direct parent-child relationship.
  - A >> B matches ancestor-descendant relationship.
  - ancestor(NodeType) checks whether the current match is inside a node type.
  - not expr negates a boolean expression.
  - where filters matches after structural matching.

  Supported first-version properties:

  - .text: source text covered by the token or node span.
  - .length: byte length of .text.
  - .start: start byte offset.
  - .end: end byte offset.
  - .field: follows a named syntax edge.
  - gap(left, right).blankLines: blank lines between adjacent nodes.

  Supported comparisons:

  - ==
  - !=
  - <
  - <=
  - >
  - >=

  ### Formatter

  The formatter must be query-driven, not magic.

  Instead of:

  format {
      sort UsingDirective by name.text
      blankLine between UsingDirective groups by usingGroup(name.text)
  }

  Use explicit formatting rules:

  format {
      rewrite "Sort using directives"
          when group: siblings(UsingDirective)
          order group by item.name.text asc
          emit group joined by "\n"

      rewrite "Separate using groups"
          when left: UsingDirective + right: UsingDirective
          where usingGroup(left.name.text) != usingGroup(right.name.text)
          emit gap(left, right) as "\n\n"

      rewrite "Keep same using group tight"
          when left: UsingDirective + right: UsingDirective
          where usingGroup(left.name.text) == usingGroup(right.name.text)
          emit gap(left, right) as "\n"
  }

  Formatter semantics:

  - rewrite defines one formatting rule.
  - when uses the same query system as rules.
  - order group by ... sorts a matched group.
  - emit group joined by "\n" replaces the matched group span with sorted node texts joined by the separator.
  - emit gap(left, right) as "\n\n" replaces only the text between two nodes.
  - Formatter rules emit text edits.
  - Edits are rejected if they overlap.
  - --check reports that formatting would change files.
  - --write applies edits.

  The plan must explicitly say that helper functions like usingGroup are not built into Go for C#. The DSL must eventually define them, but the first formatter version may support only simple property comparisons and literal grouping. If usingGroup is included in examples, the plan must include a commit that implements user-defined DSL functions.

  ## Commit Plan Requirements

  Each commit in the document must use this structure:

  ## Commit N: Title

  Goal:
  Create:
  How:
  Why:
  Tests:
  Done When:

  Every commit must be small enough to implement in one focused change.

  ## Detailed Commit Sequence

  ### Commit 1: Create the Go module

  Goal:
  Create an empty, compilable Go project.

  Create:

  - go.mod
  - cmd/spot/main.go

  How:

  - main prints spot: no command implemented yet.
  - No DSL logic yet.

  Why:
  Establish a runnable binary before any architecture exists.

  Tests:

  - Add a CLI smoke test that runs the command and checks exit code 0.

  Done When:

  - go test ./... passes.
  - go run ./cmd/spot prints the placeholder message.

  ### Commit 2: Add CLI argument parsing for DSL path and target path

  Goal:
  Accept the command shape without executing analysis.

  Create:

  - argument parsing in cmd/spot

  How:

  - Accept spot <dsl-file> <path>.
  - If missing arguments, print usage and exit non-zero.

  Why:
  The tool’s outer workflow should exist before internals are built.

  Tests:

  - Missing args returns usage.
  - Two args are accepted.

  Done When:

  - CLI can receive a DSL file path and a target path.

  ### Commit 3: Add a small logger

  Goal:
  Make early vertical slices observable.

  Create:

  - internal/log

  How:

  - Add a tiny logger with Infof.
  - Keep it dependency-free.
  - CLI logs the received DSL path and target path.

  Why:
  Early commits should show behavior even before analysis exists.

  Tests:

  - CLI output includes both paths.

  Done When:

  - Running the CLI shows what it would analyze.

  ### Commit 4: Add file discovery for one literal file

  Goal:
  Discover a single file path passed as target.

  Create:

  - internal/source/file_discovery.go

  How:

  - If target path is a file, return that file.
  - Directories are not recursive yet.

  Why:
  The first file pipeline should work with one file before glob logic exists.

  Tests:

  - Existing file is returned.
  - Missing file returns an error.

  Done When:

  - CLI logs the single file it found.

  ### Commit 5: Add recursive directory discovery

  Goal:
  Discover files under a directory.

  Create:

  - recursive walk in internal/source

  How:

  - Use filepath.WalkDir.
  - Return normalized slash-separated relative paths.
  - Do not apply include/exclude yet.

  Why:
  File discovery is needed before files {} can be meaningful.

  Tests:

  - Directory with nested files returns normalized relative paths.

  Done When:

  - CLI logs files found under a directory.

  ### Commit 6: Add Position

  Goal:
  Represent byte offsets explicitly.

  Create:

  - internal/source/position.go

  How:

  type Position struct {
      Offset uint32
  }

  Why:
  Offsets should not be anonymous integers throughout the codebase.

  Tests:

  - Construction.
  - Ordering helpers if added.

  Done When:

  - Source package compiles and tests pass.

  ### Commit 7: Add Span

  Goal:
  Represent source ranges.

  Create:

  - internal/source/span.go

  How:

  type Span struct {
      Start Position
      End   Position
  }

  Add:

  - Len() uint32
  - Empty() bool
  - Contains(Position) bool
  - Join(Span) Span

  Why:
  Tokens, nodes, diagnostics, and edits all need byte ranges.

  Tests:

  - Empty span.
  - Non-empty span.
  - Join.
  - Contains.

  Done When:

  - Span behavior is fully tested.

  ### Commit 8: Add LineMap

  Goal:
  Convert byte offsets to line and column only when rendering.

  Create:

  - internal/source/line_map.go

  How:

  - Store line-start offsets in []uint32.
  - Build once from source text.
  - Use binary search for offset lookup.

  Why:
  The runtime should use byte spans, while users need line/column diagnostics.

  Tests:

  - Empty text.
  - One line.
  - Multiple lines.
  - Trailing newline.

  Done When:

  - Offset-to-line-column behavior is deterministic.

  ### Commit 9: Add SourceFile

  Goal:
  Bundle path, text, and line map.

  Create:

  - internal/source/source_file.go

  How:

  type SourceFile struct {
      Path    string
      Text    string
      LineMap LineMap
  }

  Add:

  - Slice(span Span) string

  Why:
  All later stages should pass one source context object.

  Tests:

  - Slicing by span.
  - Line lookup through file.

  Done When:

  - Source text can be sliced safely by byte span.

  ### Commit 10: Add diagnostics

  Goal:
  Represent user-facing errors and warnings.

  Create:

  - internal/diagnostic

  How:

  type Severity uint8
  type Diagnostic struct {
      Severity Severity
      Path     string
      Span     source.Span
      Message  string
  }

  Add renderer:

  - path:start-end: severity: message
  - line/column rendering can be added later.

  Why:
  Every stage needs one diagnostic shape.

  Tests:

  - Render info, warn, err.

  Done When:

  - Diagnostics render consistently.

  ### Commit 11: Create minimal DSL token model with EOF only

  Goal:
  Start the DSL lexer with the smallest possible token stream.

  Create:

  - internal/dsl/lexer/token.go

  How:

  type Kind uint8

  const (
      EOF Kind = iota
  )

  type Token struct {
      Kind Kind
      Span source.Span
  }

  Why:
  Build lexer infrastructure before adding real syntax.

  Tests:

  - Empty input returns EOF.

  Done When:

  - Lexer package can emit EOF.

  ### Commit 12: Add DSL identifiers

  Goal:
  Lex names like files and include.

  Create:

  - identifier scanning in internal/dsl/lexer

  How:

  - First char: letter or _.
  - Rest: letter, digit, _.
  - All identifiers initially produce Identifier.

  Why:
  Most DSL syntax is identifier-driven.

  Tests:

  - abc
  - _name
  - name1

  Done When:

  - Identifier token spans are correct.

  ### Commit 13: Add { and } DSL tokens

  Goal:
  Support block boundaries.

  Create:

  - left and right brace token kinds.

  How:

  - Recognize { and }.
  - Unknown characters produce lexer diagnostics.

  Why:
  The first parseable DSL shape is files {}.

  Tests:

  - files {} produces identifier, left brace, right brace, EOF.

  Done When:

  - Lexer can tokenize empty blocks.

  ### Commit 14: Add string literals

  Goal:
  Support include "*.go".

  Create:

  - string literal token kind.

  How:

  - Double quoted strings.
  - Escapes: \\, \", \n, \r, \t.
  - Store literal span only; decoding can happen in parser.

  Why:
  Files patterns and messages need strings.

  Tests:

  - Basic string.
  - Escaped quote.
  - Unterminated string diagnostic.

  Done When:

  - Strings are lexed with exact spans.

  ### Commit 15: Add keyword classification for files, include, exclude

  Goal:
  Parse file selection without a generic keyword table yet.

  Create:

  - keyword lookup for three words.

  How:

  - After scanning an identifier, map known words to keyword kinds.
  - Unknown names remain Identifier.

  Why:
  Keep token support narrow and useful.

  Tests:

  - files include exclude
  - Unknown identifier remains identifier.

  Done When:

  - First DSL section has dedicated token kinds.

  ### Commit 16: Parse files {}

  Goal:
  Create the first parsed DSL document.

  Create:

  - internal/dsl/ast
  - internal/dsl/parser

  How:

  - Parser owns token slice and cursor.
  - Parse exactly files { }.
  - Return a document containing an empty files section.

  Why:
  This creates the first DSL vertical slice.

  Tests:

  - Valid files {}.
  - Missing {.
  - Missing }.

  Done When:

  - Parser returns a document for files {}.

  ### Commit 17: Parse one include

  Goal:
  Support files { include "**/*.go" }.

  Create:

  - ast.FilePattern
  - parser logic for include entries.

  How:

  - Entry kind is include.
  - Pattern stores raw decoded string and span.

  Why:
  The DSL can now affect file discovery.

  Tests:

  - One include.
  - Missing string.
  - Unknown entry.

  Done When:

  - Parsed document contains one include pattern.

  ### Commit 18: Parse multiple includes and excludes

  Goal:
  Complete minimal files parsing.

  Create:

  - repeated file entries.

  How:

  - Loop until }.
  - Support include and exclude.

  Why:
  Real configurations need more than one pattern.

  Tests:

  - Multiple includes.
  - Multiple excludes.
  - Mixed order.

  Done When:

  - Files section parser is complete for v1.

  ### Commit 19: Apply literal include patterns

  Goal:
  Make files affect CLI output.

  Create:

  - internal/compile minimal program with file patterns.
  - CLI parse path to file discovery.

  How:

  - For now, support exact relative path match only.
  - Ignore glob syntax until later commits.

  Why:
  This creates the first end-to-end DSL behavior.

  Tests:

  - Include exact file.
  - Exclude exact file.

  Done When:

  - CLI logs only files selected by exact patterns.

  ### Commit 20: Add * glob support

  Goal:
  Support simple file patterns.

  Create:

  - glob matcher in internal/source

  How:

  - * matches within one path segment.
  - No ** yet.

  Why:
  Small glob behavior is useful and testable.

  Tests:

  - *.go
  - src/*
  - * does not cross /.

  Done When:

  - * matching is correct.

  ### Commit 21: Add ** glob support

  Goal:
  Support recursive include patterns.

  Create:

  - recursive segment matcher.

  How:

  - ** matches zero or more path segments.
  - Keep implementation local and explicit.

  Why:
  Most real configs need **/*.cs.

  Tests:

  - **/*.go
  - src/**
  - zero-segment and multi-segment matches.

  Done When:

  - files { include "**/*.go" } works.

  ### Commit 22: Add DSL comments and whitespace skipping

  Goal:
  Make DSL files readable.

  Create:

  - lexer skipping for whitespace and // comments.

  How:

  - Whitespace is skipped.
  - // skips until newline.

  Why:
  Readable DSL files need comments.

  Tests:

  - Comments between tokens.
  - Comments at EOF.

  Done When:

  - Comments never reach parser.

  ### Commit 23: Add char literals to DSL lexer

  Goal:
  Prepare for chars.

  Create:

  - char literal token kind.

  How:

  - Single quoted char.
  - Same escapes as strings where applicable.
  - Validate exactly one rune after decoding.

  Why:
  Character definitions are the base of tokenization.

  Tests:

  - 'a'
  - '\n'
  - invalid empty char.
  - invalid multi-rune char.

  Done When:

  - Char literals are lexed and validated.

  ### Commit 24: Add expression punctuation tokens

  Goal:
  Prepare for character expressions.

  Create:

  - =, |, (, ), ?, *, +, ..

  How:

  - Add each token kind separately.
  - Keep spans exact.

  Why:
  Expression parsing needs these operators.

  Tests:

  - Each operator tokenizes correctly.

  Done When:

  - Lexer supports expression punctuation.

  ### Commit 25: Parse empty chars {}

  Goal:
  Add second DSL section.

  Create:

  - parser support for chars {}.

  How:

  - Add chars keyword.
  - Document has optional files and chars sections.
  - Parse top-level sections in any order but preserve order.

  Why:
  Top-level parser must grow incrementally.

  Tests:

  - files {} chars {}
  - chars {} files {}
  - duplicate chars is accepted syntactically and rejected later.

  Done When:

  - Empty chars section parses.

  ### Commit 26: Parse one character definition

  Goal:
  Support lower = 'a'.

  Create:

  - ast.CharDefinition
  - ast.CharExpression

  How:

  - Parse Identifier = CharLiteral.
  - Store expression as flat arena node.

  Why:
  This starts reusable token building blocks.

  Tests:

  - One definition.
  - Missing name.
  - Missing =.
  - Missing expression.

  Done When:

  - chars { lower = 'a' } parses.

  ### Commit 27: Parse character ranges

  Goal:
  Support 'a'..'z'.

  Create:

  - range expression node.

  How:

  - In primary parser, detect CharLiteral DotDot CharLiteral.

  Why:
  Ranges are essential for identifiers and numbers.

  Tests:

  - Valid range.
  - Missing right char.
  - Reversed range is validation, not parsing.

  Done When:

  - Range expressions parse with spans.

  ### Commit 28: Parse character references

  Goal:
  Support identifierStart = lower.

  Create:

  - reference expression node.

  How:

  - Identifier in expression position becomes reference.

  Why:
  Definitions need reuse.

  Tests:

  - Reference expression.
  - Reference followed by operator.

  Done When:

  - Reference expressions parse.

  ### Commit 29: Parse groups and postfix repetition

  Goal:
  Support (lower | upper)+.

  Create:

  - group expression.
  - repetition expression.

  How:

  - Parse primary.
  - If next token is ?, *, or +, wrap expression.

  Why:
  Repetition is required for token definitions.

  Tests:

  - lower?
  - lower*
  - lower+
  - (lower)+

  Done When:

  - Postfix repetition parses.

  ### Commit 30: Parse concatenation

  Goal:
  Support identifierStart identifierPart*.

  Create:

  - concatenation expression.

  How:

  - Parse consecutive repetition expressions until a boundary token.
  - A single item stays as the item.

  Why:
  Token definitions depend on implicit sequence.

  Tests:

  - Two-item sequence.
  - Three-item sequence.
  - Concatenation stops before |, ), }.

  Done When:

  - Concatenation precedence is correct.

  ### Commit 31: Parse alternation

  Goal:
  Support lower | upper | '_'.

  Create:

  - alternation expression.

  How:

  - Parse concatenation.
  - Consume | and parse next concatenation.

  Why:
  Alternation is the main expression composition tool.

  Tests:

  - Two alternatives.
  - Three alternatives.
  - Missing right side.

  Done When:

  - Expression precedence is group/repetition, concatenation, alternation.

  ### Commit 32: Validate chars

  Goal:
  Reject semantic errors in character definitions.

  Create:

  - internal/validate

  How:

  - Check duplicate names.
  - Check unknown references.
  - Check reversed ranges.
  - Check recursion.

  Why:
  Bad character definitions must not reach token compilation.

  Tests:

  - Each validation path.
  - Valid multi-definition file.

  Done When:

  - Validator produces useful diagnostics.

  ### Commit 33: Add tokens {} lexer keywords and parser section

  Goal:
  Start token declarations.

  Create:

  - tokens, skip, fallback keywords.
  - Empty tokens section parser.

  How:

  - Parse top-level tokens {}.

  Why:
  Source scanning depends on token definitions.

  Tests:

  - Empty tokens section parses.

  Done When:

  - Document can contain files, chars, and tokens.

  ### Commit 34: Parse one token definition

  Goal:
  Support Identifier = identifierStart.

  Create:

  - ast.TokenDefinition

  How:

  - Token definition reuses expression parser.
  - Token names are identifiers in declaration position.

  Why:
  Tokens are executable scanner definitions.

  Tests:

  - One token.
  - Missing name.
  - Missing expression.

  Done When:

  - Basic token declarations parse.

  ### Commit 35: Parse string expressions in tokens

  Goal:
  Support keyword tokens.

  Create:

  - string expression node.

  How:

  - String literal in expression position creates string matcher.

  Why:
  Keywords and punctuation often use string literals.

  Tests:

  - Using = "using"
  - Dot = "."

  Done When:

  - Tokens can match exact strings.

  ### Commit 36: Parse skip

  Goal:
  Support skipped tokens.

  Create:

  - token flag on TokenDefinition.

  How:

  - Optional skip after expression.

  Why:
  Whitespace and comments are usually recognized but not emitted.

  Tests:

  - Skip token.
  - Duplicate skip keyword diagnostic if parser or validator chooses to reject it.

  Done When:

  - Token definitions can be marked skip.

  ### Commit 37: Parse fallback

  Goal:
  Support partial grammars.

  Create:

  - fallback token flag.

  How:

  - Unknown = fallback has no expression.
  - Store it as fallback token definition.

  Why:
  Users must be able to model only the parts of a language they care about.

  Tests:

  - Fallback token parses.
  - fallback inside normal expression is rejected.

  Done When:

  - Fallback token declarations parse.

  ### Commit 38: Validate tokens

  Goal:
  Reject unsafe token definitions.

  Create:

  - token validator.

  How:

  - Check duplicate names.
  - Check unknown char references.
  - Check normal token expressions do not match empty input.
  - Check at most one fallback.
  - Check fallback has no normal expression.

  Why:
  Scanner must terminate and choose tokens deterministically.

  Tests:

  - Unknown reference.
  - Empty token expression.
  - Two fallback tokens.
  - Valid identifier/whitespace/fallback setup.

  Done When:

  - Valid tokens can proceed to compilation.

  ### Commit 39: Compile names to IDs

  Goal:
  Create the first compiled program shape.

  Create:

  - internal/compile.Program
  - ID types: TokenID, CharDefID

  How:

  - Assign IDs in source order.
  - Store maps only during compilation.
  - Runtime program stores slices.

  Why:
  Runtime should compare dense integers, not strings.

  Tests:

  - Stable ID assignment.

  Done When:

  - Valid DSL compiles to a minimal program.

  ### Commit 40: Compile token expressions to flat arena

  Goal:
  Create executable token expression storage.

  Create:

  - TokenExprArena
  - TokenExprNode
  - TokenExprID

  How:

  - Store nodes in []TokenExprNode.
  - Store child IDs in []TokenExprID.
  - Node contains kind, span-independent payload, first child, child count.

  Why:
  Scanner construction needs compact expression traversal.

  Tests:

  - Compile char, range, string, ref, concat, alt, repetition.

  Done When:

  - All token expression kinds compile.

  ### Commit 41: Build Thompson NFA fragments

  Goal:
  Lower token expressions into NFA fragments.

  Create:

  - internal/runtime/scanner/nfa.go

  How:

  - States: char, range, split, jump, accept.
  - Fragment has start and end state indexes.
  - Concatenation links fragments.
  - Alternation creates split.
  - Repetition creates split/back edges.

  Why:
  Thompson NFA is simple, correct, and fast enough to measure.

  Tests:

  - Fragment shape for literal char.
  - Fragment shape for concatenation.
  - Fragment shape for alternation and repetition.

  Done When:

  - NFA can be built from compiled token expressions.

  ### Commit 42: Simulate the NFA for one token

  Goal:
  Match source text against one token expression.

  Create:

  - active state set simulation.

  How:

  - Compute epsilon closure.
  - Advance by rune.
  - Track accept position.
  - Return longest match length.

  Why:
  This proves the scanner core before multiple tokens exist.

  Tests:

  - Match literal.
  - Match range.
  - Match repetition.
  - No match.

  Done When:

  - One compiled token can match source text.

  ### Commit 43: Simulate all tokens with precedence

  Goal:
  Scan with multiple token definitions.

  Create:

  - scanner over full program.

  How:

  - Start all token NFAs at each source offset.
  - Track longest accepted token.
  - On equal length, lower token ID wins.
  - Emit token unless skip.

  Why:
  This is the source string to token stream stage.

  Tests:

  - Keyword before identifier by declaration order.
  - Longest match beats earlier shorter match.
  - Skip whitespace.

  Done When:

  - Source files become token streams.

  ### Commit 44: Add fallback scanning

  Goal:
  Support unknown source text.

  Create:

  - fallback handling in scanner.

  How:

  - If no normal token matches and fallback exists, consume one rune as fallback token.
  - If no fallback exists, return scanner diagnostic.

  Why:
  Partial grammars require forward progress.

  Tests:

  - Unknown char becomes fallback token.
  - Unknown char without fallback fails.

  Done When:

  - Scanner can tokenize partially modeled files.

  ### Commit 45: Add scanner benchmark

  Goal:
  Measure the first hot path.

  Create:

  - scanner benchmark.

  How:

  - Use a realistic DSL program and source text.
  - Benchmark only scanning, not parsing DSL.

  Why:
  Performance work must be measured early.

  Tests:

  - Benchmark compiles and runs.

  Done When:

  - go test ./internal/runtime/scanner -bench . -benchmem works.

  ### Commit 46: Add syntax { root Name }

  Goal:
  Start syntax modeling.

  Create:

  - syntax and root keywords.
  - parser support for root declaration.

  How:

  - Parse syntax { root SourceFile }.

  Why:
  Runtime tree construction needs an explicit root.

  Tests:

  - Root parses.
  - Missing root name.

  Done When:

  - Syntax section can declare a root.

  ### Commit 47: Parse empty syntax node

  Goal:
  Support node Name {} syntactically.

  Create:

  - node keyword.
  - ast.SyntaxNode

  How:

  - Parse node name and braces.
  - Empty node is syntactically valid but later validation rejects if needed.

  Why:
  Node structure should be introduced before expressions.

  Tests:

  - Empty node.
  - Missing name.
  - Missing brace.

  Done When:

  - Syntax nodes parse.

  ### Commit 48: Parse bare syntax entries

  Goal:
  Support token/node references inside nodes.

  Create:

  - syntax expression reference node.

  How:

  - Identifier inside node body becomes syntax reference.

  Why:
  Nodes are composed from token and node references.

  Tests:

  - node Using { Using Identifier }.

  Done When:

  - Bare entries produce concatenation when multiple exist.

  ### Commit 49: Parse named captures

  Goal:
  Support tree field labels.

  Create:

  - syntax capture expression.

  How:

  - name: QualifiedName wraps target expression with field name.

  Why:
  Queries and formatting need meaningful paths like name.text.

  Tests:

  - One capture.
  - Multiple captures.

  Done When:

  - Captures are preserved in parsed AST.

  ### Commit 50: Parse field repetition

  Goal:
  Support repeated child fields.

  Create:

  - syntax repetition for fields.

  How:

  - members*: Statement
  - item?: Node
  - items+: Node

  Why:
  Syntax trees need lists and optional parts.

  Tests:

  - ?, *, + field modifiers.

  Done When:

  - Field repetition parses.

  ### Commit 51: Parse oneOf

  Goal:
  Support syntax alternation.

  Create:

  - oneOf keyword and expression node.

  How:

  - oneOf { A B any } is ordered alternatives.
  - Each entry is one alternative.

  Why:
  Grammar choices must be readable.

  Tests:

  - Multiple alternatives.
  - Missing closing brace.

  Done When:

  - oneOf parses.

  ### Commit 52: Parse any

  Goal:
  Support partially modeled syntax.

  Create:

  - any syntax expression.

  How:

  - any matches one emitted token.

  Why:
  Users should not have to fully model a language upfront.

  Tests:

  - node Root { members*: oneOf { Known any } }.

  Done When:

  - any is represented in syntax AST.

  ### Commit 53: Validate syntax

  Goal:
  Reject invalid syntax models.

  Create:

  - syntax validator.

  How:

  - Check root exists.
  - Check node names unique.
  - Check references resolve to token or node.
  - Check repetitions cannot match empty input.
  - Check any only appears in syntax expressions.

  Why:
  Runtime parser must terminate and be predictable.

  Tests:

  - Missing root.
  - Unknown reference.
  - Unsafe repetition.
  - Valid partial grammar.

  Done When:

  - Valid syntax models pass validation.

  ### Commit 54: Compile syntax expressions

  Goal:
  Lower syntax model to runtime program.

  Create:

  - SyntaxExprArena
  - SyntaxNodeDef
  - IDs: SyntaxNodeID, FieldID

  How:

  - Token references compile to token IDs.
  - Node references compile to syntax node IDs.
  - Captures compile to field IDs.
  - any compiles to a dedicated op.

  Why:
  Runtime syntax parser should not resolve names.

  Tests:

  - Compile references, captures, repetition, oneOf, any.

  Done When:

  - Syntax program is compact and ID-based.

  ### Commit 55: Add runtime tree storage

  Goal:
  Represent materialized syntax trees.

  Create:

  - internal/runtime/syntax/tree.go

  How:

  type Tree struct {
      Nodes []Node
      Edges []Edge
  }

  type Node struct {
      Kind       SyntaxNodeID
      StartToken uint32
      EndToken   uint32
      FirstEdge  uint32
      EdgeCount  uint32
  }

  type Edge struct {
      Field FieldID
      Child NodeID
  }

  Why:
  Queries need fast traversal without pointer-heavy trees.

  Tests:

  - Add node.
  - Add edge.
  - Enumerate children.

  Done When:

  - Tree storage is tested independently.

  ### Commit 56: Parse one syntax node at runtime

  Goal:
  Match a compiled node against tokens.

  Create:

  - runtime syntax parser.

  How:

  - Function parseNode(nodeID, tokenIndex).
  - Return node ID and next token index on success.
  - Fail without diagnostics for backtracking.

  Why:
  This is the core syntax parser primitive.

  Tests:

  - Node matching one token.
  - Node mismatch.

  Done When:

  - One-node tree can be built.

  ### Commit 57: Parse concatenation and captures

  Goal:
  Build child edges.

  Create:

  - runtime evaluation for sequence and capture.

  How:

  - Concatenation parses entries in order.
  - Capture adds edge from current node to child node.
  - Bare token references do not create nodes.

  Why:
  Runtime tree shape must reflect named syntax fields.

  Tests:

  - Qualified name with tail captures.
  - Using directive with name capture.

  Done When:

  - Captured child nodes appear in tree.

  ### Commit 58: Parse repetition and oneOf

  Goal:
  Complete syntax parser basics.

  Create:

  - runtime evaluation for repetition and alternation.

  How:

  - oneOf tries alternatives in order.
  - Repetition loops until no progress.
  - If repeated expression makes no progress, fail validation earlier.

  Why:
  Real syntax models need lists and choices.

  Tests:

  - Repeated members.
  - Ordered alternatives.
  - Partial syntax using any.

  Done When:

  - Runtime parser can build a full source file tree.

  ### Commit 59: Add AST printing

  Goal:
  Make generated trees visible.

  Create:

  - tree renderer.

  How:

  - Print node name, token range, and field labels.
  - CLI flag --print-ast.

  Why:
  Users must understand what their DSL creates.

  Tests:

  - Stable printed tree.

  Done When:

  - CLI can print runtime syntax tree.

  ### Commit 60: Add query token model

  Goal:
  Start rule parsing in tiny steps.

  Create:

  - lexer keywords: rules, when, where, warn, info, err.

  How:

  - Add only required tokens for minimal rule syntax.

  Why:
  Rule language should grow incrementally.

  Tests:

  - Keywords tokenize.

  Done When:

  - Lexer supports minimal rules vocabulary.

  ### Commit 61: Parse one simple rule

  Goal:
  Support a node match rule.

  Create:

  - ast.Rule

  How:

  rules {
      warn "message" when UsingDirective
  }

  Why:
  This is the smallest syntax rule.

  Tests:

  - One rule.
  - Missing message.
  - Missing when.

  Done When:

  - Rule declarations parse.

  ### Commit 62: Parse selector bindings

  Goal:
  Support named matches.

  Create:

  - selector binding AST.

  How:

  - left: UsingDirective
  - If no binding is given, bind as match.

  Why:
  where and formatter rules need stable names.

  Tests:

  - Bound selector.
  - Unbound selector.

  Done When:

  - Selectors can introduce names.

  ### Commit 63: Parse adjacency selectors

  Goal:
  Support sibling relationships.

  Create:

  - selector relation AST.

  How:

  - left: UsingDirective + right: UsingDirective
  - + means adjacent siblings with same parent.

  Why:
  Ordering and spacing rules depend on adjacency.

  Tests:

  - Adjacent selector parses.
  - Missing right selector error.

  Done When:

  - Adjacency is parsed with bindings.

  ### Commit 64: Parse parent and ancestor selectors

  Goal:
  Support structural relationships.

  Create:

  - selector relation kinds.

  How:

  - A > B means direct parent.
  - A >> B means ancestor.

  Why:
  Rules need inside/outside structure.

  Tests:

  - Direct parent.
  - Ancestor.

  Done When:

  - Structural selectors parse.

  ### Commit 65: Parse where property paths

  Goal:
  Support rule filters.

  Create:

  - condition AST.

  How:

  - Parse left.name.text.
  - Path is subject plus field/property segments.

  Why:
  Queries need access to captured syntax data.

  Tests:

  - .text
  - .length
  - nested field path.

  Done When:

  - Property paths parse.

  ### Commit 66: Parse comparisons

  Goal:
  Support boolean filters.

  Create:

  - comparison condition AST.

  How:

  - Parse left.name.text > right.name.text.
  - Parse string and integer literal right-hand sides.

  Why:
  Sorting and style rules need comparisons.

  Tests:

  - All comparison operators.
  - String literal comparison.
  - Integer comparison.

  Done When:

  - Comparisons parse.

  ### Commit 67: Parse boolean conditions

  Goal:
  Support not, and, or.

  Create:

  - boolean expression AST.

  How:

  - Precedence: not, and, or.
  - Parentheses optional in later commit if needed.

  Why:
  Real rules need compound filters.

  Tests:

  - not ancestor(NamespaceDeclaration)
  - a and b
  - a or b

  Done When:

  - Boolean where expressions parse.

  ### Commit 68: Validate rules

  Goal:
  Reject invalid query rules.

  Create:

  - rule validator.

  How:

  - Check node names.
  - Check bindings exist.
  - Check property paths exist.
  - Check comparison types.

  Why:
  Runtime query engine should not handle bad rules.

  Tests:

  - Unknown node.
  - Unknown binding.
  - Unknown field path.
  - Valid adjacency rule.

  Done When:

  - Rule validation is complete for v1.

  ### Commit 69: Compile query programs

  Goal:
  Lower parsed selectors to runtime query programs.

  Create:

  - internal/compile/query.go

  How:

  - Compile node names to syntax IDs.
  - Compile field paths to field IDs.
  - Compile condition expressions to compact operations.

  Why:
  Runtime query evaluation should be ID-based.

  Tests:

  - Single node query.
  - Adjacent query.
  - Where condition.

  Done When:

  - Rules compile to query programs.

  ### Commit 70: Execute single-node queries

  Goal:
  Find matching syntax nodes.

  Create:

  - internal/runtime/query

  How:

  - Iterate tree nodes.
  - Match node kind ID.
  - Produce match bindings.

  Why:
  This is the smallest query engine.

  Tests:

  - Match one node.
  - Match multiple nodes.
  - Match none.

  Done When:

  - Simple rules can find nodes.

  ### Commit 71: Execute adjacency queries

  Goal:
  Support ordering and spacing rules.

  Create:

  - adjacent sibling matcher.

  How:

  - For each parent, inspect ordered child edges.
  - Match adjacent node kinds.
  - Bind left and right nodes.

  Why:
  Formatter and style rules often depend on order.

  Tests:

  - Adjacent match.
  - Non-adjacent no match.
  - Different parents no match.

  Done When:

  - A + B works.

  ### Commit 72: Execute parent and ancestor queries

  Goal:
  Support inside/outside rules.

  Create:

  - parent lookup or ancestor traversal.

  How:

  - Build parent indexes once per tree.
  - > checks direct parent.
  - >> walks parents.

  Why:
  Structural rules depend on containment.

  Tests:

  - Direct parent.
  - Deep ancestor.
  - No ancestor.

  Done When:

  - Parent and ancestor selectors work.

  ### Commit 73: Evaluate query conditions

  Goal:
  Filter matches.

  Create:

  - condition evaluator.

  How:

  - Resolve .text by slicing source from node span.
  - Resolve .length from span length.
  - Resolve field path by following child edges.
  - Compare string/int values.

  Why:
  Structural matches need semantic filtering.

  Tests:

  - Text comparison.
  - Length comparison.
  - Field path comparison.

  Done When:

  - where left.name.text > right.name.text works.

  ### Commit 74: Emit diagnostics from rules

  Goal:
  Complete analyzer loop.

  Create:

  - rule diagnostic generation.

  How:

  - Diagnostic span defaults to primary match binding.
  - Severity and message come from rule.
  - Later DSL can add explicit at.

  Why:
  Rules must produce user-visible output.

  Tests:

  - One diagnostic.
  - Multiple diagnostics.
  - No diagnostics.

  Done When:

  - End-to-end analyzer reports rule violations.

  ### Commit 75: Add full engine orchestration

  Goal:
  Run the whole analysis pipeline.

  Create:

  - internal/runtime/engine

  How:

  - Input: compiled program and source file.
  - Scan source.
  - Parse syntax if syntax exists.
  - Run queries.
  - Return diagnostics.

  Why:
  CLI should call one runtime entrypoint.

  Tests:

  - Full source file analysis.

  Done When:

  - CLI can analyze a selected file.

  ### Commit 76: Add formatter DSL section

  Goal:
  Start formatter syntax.

  Create:

  - format, rewrite, emit, joined, by, as keywords.

  How:

  - Parse empty format {} first.

  Why:
  Formatter is part of the product, but should use the query system.

  Tests:

  - Empty format section.

  Done When:

  - Format section parses.

  ### Commit 77: Parse rewrite rule with when

  Goal:
  Reuse query syntax for formatting.

  Create:

  - ast.FormatRule

  How:

  format {
      rewrite "Name"
          when left: UsingDirective + right: UsingDirective
  }

  Why:
  Formatting should select syntax using the same model as rules.

  Tests:

  - One rewrite rule.
  - Missing name.
  - Missing when.

  Done When:

  - Formatter can parse query-based rules.

  ### Commit 78: Parse gap replacement

  Goal:
  Support spacing normalization.

  Create:

  - formatter action AST.

  How:

  emit gap(left, right) as "\n\n"

  Why:
  Blank-line formatting should be explicit text replacement, not magic.

  Tests:

  - Gap emit parses.
  - Unknown binding is validation.

  Done When:

  - Gap replacement is represented.

  ### Commit 79: Parse group ordering

  Goal:
  Support sorted groups explicitly.

  Create:

  - group selector and ordering AST.

  How:

  rewrite "Sort usings"
      when group: siblings(UsingDirective)
      order group by item.name.text asc
      emit group joined by "\n"

  Define:

  - siblings(NodeType) returns contiguous siblings of the same node type under one parent.
  - item is the implicit variable for each group member during ordering.

  Why:
  Sorting is a general formatter operation over query-selected groups.

  Tests:

  - Group rule parses.
  - Order key parses.
  - Joined emit parses.

  Done When:

  - Sort formatter syntax is explicit and parsed.

  ### Commit 80: Validate formatter rules

  Goal:
  Reject invalid formatting declarations.

  Create:

  - formatter validator.

  How:

  - Reuse query validation.
  - Check gap bindings exist.
  - Check group binding exists.
  - Check order path resolves from group item.
  - Check emit target is valid.

  Why:
  Formatter runtime should receive valid commands only.

  Tests:

  - Unknown binding.
  - Invalid path.
  - Valid gap rule.
  - Valid group sort rule.

  Done When:

  - Formatter validation is complete for first version.

  ### Commit 81: Add text edit model

  Goal:
  Represent formatter output safely.

  Create:

  - internal/runtime/format/edit.go

  How:

  type Edit struct {
      Span source.Span
      Replacement string
  }

  Add:

  - sort edits by span.
  - reject overlaps.
  - apply edits back-to-front.

  Why:
  Formatting must be deterministic and safe.

  Tests:

  - Apply one edit.
  - Apply multiple edits.
  - Reject overlap.

  Done When:

  - Text edit application is tested.

  ### Commit 82: Execute gap formatter actions

  Goal:
  Normalize spacing between matched nodes.

  Create:

  - formatter runtime for gap emit.

  How:

  - Use node end span and next node start span.
  - Replace that source region with configured string.

  Why:
  This handles blank-line and spacing rules generally.

  Tests:

  - Replace no blank line with one.
  - Replace too many blank lines.
  - No-op if already correct.

  Done When:

  - Gap formatting produces edits.

  ### Commit 83: Execute group sort formatter actions

  Goal:
  Sort syntax node groups.

  Create:

  - formatter runtime for group ordering.

  How:

  - Evaluate order key for each group item.
  - Stable sort by key.
  - Replace full group span with node texts joined by separator.

  Why:
  Sorting is a general formatting primitive.

  Tests:

  - Sort three using directives.
  - Already sorted produces no edit.
  - Preserve stable order for equal keys.

  Done When:

  - Group sorting produces deterministic edits.

  ### Commit 84: Add CLI --check

  Goal:
  Report formatting drift without writing.

  Create:

  - CLI flag.

  How:

  - Run formatter.
  - If edits exist, print file path and exit non-zero.

  Why:
  CI needs check mode.

  Tests:

  - Clean file exits zero.
  - Dirty file exits non-zero.

  Done When:

  - --check works.

  ### Commit 85: Add CLI --write

  Goal:
  Apply formatter edits.

  Create:

  - CLI write mode.

  How:

  - Apply validated edits.
  - Write file only if content changes.

  Why:
  Formatter must be usable.

  Tests:

  - File content changes.
  - No-op does not rewrite.

  Done When:

  - --write works.

  ### Commit 86: Add full examples

  Goal:
  Provide executable documentation.

  Create:

  - examples/go-basic.spot
  - examples/csharp-usings.spot
  - source samples for each.

  How:

  - Examples must cover files, chars, tokens, syntax, rules, and format.

  Why:
  Examples show the real product surface.

  Tests:

  - Example integration tests.

  Done When:

  - Examples run in tests.

  ### Commit 87: Add parser benchmarks

  Goal:
  Measure DSL frontend.

  Create:

  - benchmarks for lexer, parser, resolver, validator, compiler.

  How:

  - Use full DSL input with all constructs.

  Why:
  DSL processing cost must be known.

  Done When:

  - Benchmarks run with -benchmem.

  ### Commit 88: Add runtime benchmarks

  Goal:
  Measure hot execution paths.

  Create:

  - scanner benchmark.
  - syntax parser benchmark.
  - query benchmark.
  - formatter benchmark.
  - full engine benchmark.

  How:

  - Use representative source files and compiled program.

  Why:
  Runtime performance is the selling point.

  Done When:

  - All hot stages are benchmarked.

  Goal:
  Optimize scanner only with evidence.

  Create:

  - active-state-set transition cache.

  How:

  - Key by active state set plus input class.
  - Reuse computed next state set.
  - Keep NFA simulation as source of truth.

  Why:
  Avoid speculative complexity.

  Tests:

  - Same scanner behavior as NFA.
  - Benchmark before/after.

  Done When:

  - Keep this commit only if benchmark improvement is meaningful.

  ### Commit 90: Final documentation pass

  Goal:
  Make the implementation explainable.

  Create:

  - final docs/dsl.md
  - final docs/architecture.md
  - final docs/performance.md

  How:

  - Document DSL grammar.
  - Document flow from string to tokens to syntax tree to queries to edits.
  - Document performance model and benchmark expectations.

  Why:
  The DSL and architecture are the product.

  Done When:

  - Docs match examples and tests.

  ## Acceptance Criteria

  The document is complete when:

  - It explains source text to tokens to syntax tree to queries to diagnostics/edits.
  - It defines scanner algorithm choices and optimization strategy.
  - It defines the query language enough to implement it.
  - It defines formatter semantics without magic commands.
  - Every commit is small, compilable, testable, and has observable value.
  - An engineer can implement the rewrite without making architectural decisions.
