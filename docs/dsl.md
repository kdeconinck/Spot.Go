# DSL Specification

This document is the authoritative specification for the Spot DSL.
The DSL defines how Spot selects files, tokenizes source text, evaluates rules, and emits diagnostics.
This document describes the current DSL design. It should remain small, explicit, and implementable.

# Design Goals

The DSL should be:

* Small.
* Explicit.
* Deterministic.
* Easy to parse.
* Easy to validate.
* Efficient to execute.
* Readable without advanced language knowledge.

The DSL should not try to be a general-purpose programming language.

# Top-Level Structure

A Spot DSL file is composed of top-level sections.
The current sections are:

```spot
scope {
    include "**/*.go"
    exclude "vendor/**"
}

definitions {
    letter = 'a'..'z' | 'A'..'Z'
    digit = '0'..'9'
}

tokens {
    Whitespace = (' ' | '\t' | '\r' | '\n')+ skip
    Identifier = letter (letter | digit | '_')*
}

syntax {
    node PackageName {
        Identifier
    }

    node PackageClause {
        KeywordPackage
        name: PackageName
    }

    node Root {
        package: PackageClause
    }
}

rules {
    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warning at Identifier "Public identifier found"
    }
}
```

Sections may appear at most once.
The recommended section order is:

1. `scope`
2. `definitions`
3. `tokens`
4. `syntax`
5. `rules`

# Comments

Line comments start with `//` and continue until the end of the line.

```spot
// Analyze Go source files.
scope {
    include "**/*.go"
}
```

Block comments are not part of the current DSL.

# Identifiers

Identifiers are used for definitions, tokens, and rules.
Definition names should use lower camel case:

```spot
letter
digit
identifierStart
```

Token names should use upper camel case:

```spot
Identifier
Whitespace
StringLiteral
```

Rule names should use upper camel case:

```spot
PublicIdentifier
MultipleSpaces
```

Identifiers must start with a letter and may contain letters, digits, and underscores.

The following words are reserved and therefore cannot be used as user-defined names:

* `scope`
* `include`
* `exclude`
* `definitions`
* `tokens`
* `rules`
* `syntax`
* `node`
* `oneOf`
* `any`
* `not`
* `startsWith`
* `rule`
* `match`
* `where`
* `inside`
* `outside`
* `report`
* `skip`
* `fallback`
* `info`
* `warn`
* `err`
* `at`

`syntax` and `node` are reserved for the syntax-tree construction layer.

# Literals

## Character Literals

Character literals use single quotes.

```spot
'a'
'0'
'_'
'\n'
'\t'
'\r'
```

Supported escapes are:

| Escape | Meaning         |
| ------ | --------------- |
| `\\`   | Backslash       |
| `\'`   | Single quote    |
| `\n`   | Newline         |
| `\r`   | Carriage return |
| `\t`   | Tab             |

A character literal represents exactly one character.

## String Literals

String literals use double quotes.

```spot
"public"
"Multiple spaces found"
```

Supported escapes are:

| Escape | Meaning         |
| ------ | --------------- |
| `\\`   | Backslash       |
| `\"`   | Double quote    |
| `\n`   | Newline         |
| `\r`   | Carriage return |
| `\t`   | Tab             |

String literals are used for exact text matches and diagnostic messages.

## Integer Literals

Integer literals use decimal digits.

```spot
0
1
123
```

Integer literals are used for numeric rule conditions, such as token length comparisons.

# Scope Section

The `scope` section defines which files are analyzed.

```spot
scope {
    include "**/*.go"
    exclude "vendor/**"
    exclude "**/*_generated.go"
}
```

## Include

`include` adds files to the analysis set.

```spot
include "**/*.go"
```

At least one `include` entry is required.

## Exclude

`exclude` removes files from the analysis set.

```spot
exclude "vendor/**"
```

`exclude` entries are optional.

## Scope Semantics

A file is analyzed when:

1. It matches at least one `include` pattern.
2. It does not match any `exclude` pattern.

Patterns are matched against the file path relative to the analyzed root directory.
Paths are normalized to use `/` as the separator, even on operating systems that use a different path separator.

## Scope Pattern Syntax

Scope patterns support a small glob syntax.

### `*`

`*` matches zero or more characters within a single path segment.
It does not cross `/`.

Examples:

```text
*.go
file*
*_generated.go
```

### `?`

`?` matches exactly one character within a single path segment.
It does not cross `/`.

Examples:

```text
file?.go
v?
```

### `**`

`**` matches zero or more complete path segments.
It may cross `/`.
`**` only has this recursive meaning when it occupies an entire path segment.

Examples:

```text
**/*.go
vendor/**
internal/**/testdata/*.json
```

## Scope Examples

```spot
scope {
    include "**/*.go"
    exclude "vendor/**"
    exclude "**/*_generated.go"
}
```

With these rules:

* `cmd/spot/main.go` is included.
* `vendor/example/main.go` is excluded.
* `pkg/model/user_generated.go` is excluded.

# Definitions Section

The `definitions` section declares reusable character-level expressions.

```spot
definitions {
    letter = 'a'..'z' | 'A'..'Z'
    digit = '0'..'9'
    identifierStart = letter | '_'
    identifierPart = letter | digit | '_'
}
```

Definitions do not emit tokens.
Definitions exist only to make token declarations reusable and readable.

## Definition Expressions

Definitions may contain:

* Character literals.
* Character ranges.
* References to earlier or later definitions.
* Grouping.
* Concatenation.
* Alternation.
* Repetition.

Examples:

```spot
letter = 'a'..'z' | 'A'..'Z'
digit = '0'..'9'
identifierStart = letter | '_'
identifierPart = identifierStart | digit
```

Definition expressions use the following precedence, from highest to lowest:

| Precedence | Expression     | Example              |
| ---------- | -------------- | -------------------- |
| 1          | Grouping       | `('a' | 'b')`        |
| 2          | Repetition     | `letter*`            |
| 3          | Concatenation  | `letter digit`       |
| 4          | Alternation    | `letter digit | '_'` |

For example, `letter digit | '_'` is parsed as `(letter digit) | '_'`.
Use grouping when alternation should be part of a sequence: `letter (digit | '_')`.

## Character Ranges

Character ranges use `..`.

```spot
'a'..'z'
'0'..'9'
```

Both sides of a range must be character literals.
The start character must be less than or equal to the end character.

# Tokens Section

The `tokens` section declares how source text is converted into tokens.

```spot
tokens {
    Whitespace = (' ' | '\t' | '\r' | '\n')+ skip
    Identifier = identifierStart identifierPart*
    Number = digit+
}
```

Each token declaration has a name and an expression.

```spot
TokenName = expression
```

A token may optionally be marked as `skip`.

```spot
Whitespace = (' ' | '\t' | '\r' | '\n')+ skip
```

Skipped tokens are recognized by the scanner but are not emitted into the token stream.

A token may also be declared as a fallback token.

```spot
Unknown = fallback
```

Fallback tokens do not use normal token expressions. A fallback token consumes exactly one byte of source text only
when no other token matches at the current byte offset.

Fallback tokens follow these rules:

* At most one fallback token may be declared.
* A fallback token loses to every normal token.
* A fallback token always consumes exactly one byte.
* A fallback token may be combined with `skip`.

## Token Expressions

Token expressions may contain:

* Character literals.
* String literals.
* Character ranges.
* Definition references.
* Grouping.
* Alternation.
* Repetition.

Examples:

```spot
Identifier = identifierStart identifierPart*
Number = digit+
Equals = "="
Whitespace = (' ' | '\t' | '\r' | '\n')+ skip
Unknown = fallback
```

## Repetition

The DSL supports three repetition operators.

| Operator | Meaning       |
| -------- | ------------- |
| `?`      | Zero or one.  |
| `*`      | Zero or more. |
| `+`      | One or more.  |

Examples:

```spot
digit+
letter*
'-'?
```

# Scanner Semantics

The scanner reads source text from left to right.
At each byte offset, it evaluates token definitions and chooses the best match.

Selection rules:

1. The longest match wins.
2. If multiple tokens match the same length, the token declared first wins.
3. If no token matches and a fallback token exists, the fallback token consumes one byte.
4. If no token matches and no fallback token exists, scanning fails with a diagnostic.

This makes tokenization deterministic.

## Token Order

Token order matters when two token definitions can match the same text.

Example:

```spot
tokens {
    KeywordPublic = "public"
    Identifier = identifierStart identifierPart*
}
```

For the input:

```text
public
```

both `KeywordPublic` and `Identifier` match the same text.

Because `KeywordPublic` is declared first, it wins.

# Token Stream

A token has:

* Token name.
* Text.
* Source span.

The source span is represented using byte offsets.
Line and column information may be derived later for diagnostic rendering.
Skipped tokens are not included in the token stream.
A fallback token is only emitted when no regular token matches at the current byte offset.

# Rules Section

The `rules` section declares diagnostics over the token stream.

```spot
rules {
    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warning at Identifier "Public identifier found"
    }
}
```

A rule contains:

1. A token match.
2. Optional conditions.
3. A report statement.

# Syntax Section

The `syntax` section declares syntax node kinds and their structure for future syntax-tree construction.

```spot
syntax {
    node Word {
        oneOf {
            Identifier
            KeywordPublic
        }
    }

    node WordPair {
        left: Word
        right: Word
    }

    node WordList {
        items*: oneOf {
            Word
            any
        }
    }
}
```

Each declaration introduces one syntax node kind.
The preferred surface form is block-shaped and describes the resulting tree directly:

```spot
node UsingStatement {
    KeywordUsing
    name: QualifiedIdentifier
    Semicolon
}
```

Syntax expressions may contain:

* Token references.
* Syntax node references.
* Named child captures.
* `oneOf { ... }` variant blocks in structured syntax nodes.
* `any`, which matches one emitted token of any kind.
* Grouping.
* Concatenation.
* Alternation.
* Repetition.

Examples:

```spot
node Word {
    oneOf {
        Identifier
        KeywordPublic
    }
}

node WordPair {
    left: Word
    right: Word
}

node UsingDirective {
    KeywordUsing
    name: QualifiedIdentifier
    Semicolon
}
```

`any` is useful when the syntax section models only the node kinds that matter to the current rules. It consumes one
emitted token regardless of token name, so a file can still be fully materialized even when the grammar is only
partially described.

A named child capture labels the direct child syntax nodes produced by one syntax expression:

```spot
node UsingDirective {
    KeywordUsing
    name: QualifiedIdentifier
    Semicolon
}
```

Here `name` becomes a stable field on `UsingDirective`. Rules may later navigate through that field with
`UsingDirective.name`.

Structured syntax nodes support:

* Unnamed required entries such as `KeywordUsing`.
* Named child fields such as `name: QualifiedIdentifier`.
* Optional fields such as `value?: Expression`.
* Repeated fields such as `members*: Statement`.
* Variant blocks such as `oneOf { A B C }`.

Syntax expressions use the following precedence, from highest to lowest:

| Precedence | Expression    | Example                       |
| ---------- | ------------- | ----------------------------- |
| 1          | Grouping      | `(Word | KeywordInternal)`    |
| 2          | Repetition    | `Word*`                       |
| 3          | Concatenation | `KeywordUsing Identifier`     |
| 4          | Alternation   | `Identifier | KeywordPublic`  |

For example, `Word KeywordInternal | Identifier` is parsed as `(Word KeywordInternal) | Identifier`.
Use grouping when alternation should be part of a sequence: `Word (KeywordInternal | Identifier)`.

Today, the `syntax` section is intentionally limited:

* It describes syntax node structure.
* It can be compiled and matched against token streams.
* Rules may match syntax nodes and inspect `text`, `length`, and captured child paths.
* It does not yet model semantic information such as name resolution or types.

# Rule Match

Spot supports two rule styles:

* Block rules, which spell out `match`, optional `where`, and `report`.
* Selector rules, which use a compact query-like syntax for syntax-node matches.

Block rules match either a single token or a single syntax node.

```spot
match Identifier
```

```spot
match node PackageClause
```

```spot
match node UsingDirective outside NamespaceDeclaration
```

The matched token or syntax node is then referenced by name in `where` and `report`.

Example:

```spot
rule PublicIdentifier {
    match Identifier
    where Identifier.text == "public"
    report warning at Identifier "Public identifier found"
}

rule PackageClauseRule {
    match node PackageClause
    where PackageClause.text == "package main"
    report warn at PackageClause "Package clause found"
}

rule UsingOutsideNamespace {
    match node UsingDirective outside NamespaceDeclaration
    report warn at UsingDirective "Using directive must live inside a namespace"
}
```

When a rule matches a syntax node, Spot first builds one full-file syntax tree and then evaluates the rule against
every node of the requested kind in that tree.

Syntax-node matches may optionally constrain ancestor nodes:

* `inside NamespaceDeclaration` means the matched node must have a `NamespaceDeclaration` ancestor.
* `outside NamespaceDeclaration` means the matched node must not have a `NamespaceDeclaration` ancestor.

Ancestor constraints are only valid on `match node ...` rules.

Selector rules are syntax-node rules written in a compact form:

```spot
warn "Using directive is outside a namespace."
    : UsingDirective:not(NamespaceDeclaration > *)

info "Using directive is inside a namespace."
    : NamespaceDeclaration > UsingDirective
```

Selector rules currently support:

* `Node`, which matches every node of that kind.
* `Ancestor Node`, which matches `Node` anywhere inside `Ancestor`.
* `Parent > Node`, which matches `Node` only when its direct parent is `Parent`.
* `Left + Right`, which matches `Right` when it is the adjacent sibling of `Left`.
* `Node:not(Ancestor *)`, which matches `Node` only when it is outside `Ancestor`.
* `Node:not(Parent > *)`, which matches `Node` only when its direct parent is not `Parent`.

Selector rules are always syntax-node rules. They do not currently support token matches, but they may use `where`
clauses.

For adjacent-sibling rules, the `where` clause may reference:

* `left`, which is the left adjacent sibling.
* `right`, which is the matched node on the right.
* `gap`, which is the source text between `left` and `right`.

Supported generic properties are:

* `text`
* `length`
* `blankLines`

When a syntax node declares named child captures, a `where` clause may navigate through them before reading a
property:

```spot
where left.name.text > right.name.text
```

Each path segment must match a named capture declared by the syntax node reached at the previous step.

Examples:

```spot
warn "Using directives must be alphabetical."
    : UsingDirective + UsingDirective
    where left.text > right.text

warn "A blank line must separate using groups."
    : UsingDirective + UsingDirective
    where gap.blankLines == 0
```

The current DSL does not yet support boolean operators such as `and` or `or`, so multi-part ordering rules must still
be expressed as separate rules.

# Rule Conditions

A `where` clause filters the matched token or syntax node.

```spot
where Identifier.text == "public"
```

The current condition model supports comparisons against match properties.

Supported match properties:

| Property | Meaning                                   |
| -------- | ----------------------------------------- |
| `text`   | The exact source text covered by the match. |
| `length` | The byte length of the matched source text. |

Supported comparison operators depend on the property:

| Property | Operators                        |
| -------- | -------------------------------- |
| `text`   | `==`, `!=`, `<`, `<=`, `>`, `>=`, `startsWith` |
| `length` | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| `blankLines` | `==`, `!=`, `<`, `<=`, `>`, `>=` |

Examples:

```spot
where Identifier.text == "public"
where PackageClause.length > 0
where left.text > right.text
where left.name.text > right.name.text
where gap.blankLines == 1
```

Only one `where` clause is supported for now.
Boolean expressions such as `and`, `or`, and `not` are outside the current DSL.

# Reports

A `report` statement emits a diagnostic.

```spot
report warning at Identifier "Public identifier found"
```

Report syntax:

```spot
report severity at matchName "message"
```

Supported severities:

* Informational (`info`).
* Warning (`warn`).
* Error (`err`).

The `at` target must reference the same token or syntax node matched by the rule.
The diagnostic span is the span of that matched token or syntax node.

# Complete Example

```spot
scope {
    include "**/*.go"
    exclude "vendor/**"
}

definitions {
    letter = 'a'..'z' | 'A'..'Z'
    digit = '0'..'9'
    identifierStart = letter | '_'
    identifierPart = identifierStart | digit
}

tokens {
    Whitespace = (' ' | '\t' | '\r' | '\n')+ skip
    Identifier = identifierStart identifierPart*
}

rules {
    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warning at Identifier "Public identifier found"
    }
}
```

# Validation Rules

A valid DSL file must satisfy all validation rules.

## Section Validation

* Each top-level section may appear at most once.
* Unknown top-level sections are invalid.
* `scope` must contain at least one `include`.
* `include` patterns must not be empty.
* `exclude` patterns must not be empty.
* `tokens` must contain at least one token declaration.

## Name Validation

* Definition names must be unique.
* Token names must be unique.
* Syntax node names must be unique.
* Rule names must be unique.
* A definition and a token may not share the same name.
* A syntax node and a token may not share the same name.
* Referenced definitions must exist.
* Referenced tokens must exist.
* Referenced syntax expressions must resolve to a declared token or syntax node.
* Tokens may contain at most one fallback token.
* Syntax rules require exactly one root syntax node for the file-level syntax tree.

## Definition Validation

* Character ranges must have a valid order.
* Recursive definitions are invalid.

Invalid:

```spot
definitions {
    a = b
    b = a
}
```

## Token Validation

* Token expressions must not match empty input.
* `skip` may only appear at the end of a token declaration.
* A fallback token consumes one otherwise-unmatched byte and is checked only after all normal tokens fail.

Invalid:

```spot
tokens {
    Empty = digit*
}
```

because it can match an empty string.

## Syntax Validation

* Syntax node references must resolve to declared tokens or syntax nodes.
* `any` always matches exactly one emitted token.
* Recursive syntax nodes are invalid.
* A syntax repetition expression must not repeat something that can match empty input.
* Named syntax-field paths in rules must follow declared captures.

Invalid:

```spot
syntax {
    node Word {
        value?: Identifier
    }

    node WordList {
        values: Word*
    }
}
```

because `Word` can match empty input, so repeating it is ambiguous and not executable safely.

## Rule Validation

* A block rule must contain exactly one `match`.
* A block rule must contain exactly one `report`.
* A token-rule `where` clause may only reference the matched token.
* A `report` target must reference the matched token.
* A token property must exist.
* A comparison must use compatible value types.
* A syntax rule must match a declared syntax node.
* A syntax rule's `where` clause may only reference the matched syntax node, `left`, `right`, or `gap` as allowed by the selector kind.
* A syntax-rule path segment must reference a declared named child capture on the selected syntax path.
* A syntax rule's `report` target must reference the matched syntax node.
* A syntax node property must exist.
* A syntax-node comparison must use compatible value types.
* Only syntax-node rules may use `inside` or `outside`.
* An `inside` or `outside` constraint must reference a declared syntax node.
* Selector rules may only reference declared syntax nodes.

# Current Non-Goals

The current DSL does not include:

* Import statements.
* Variables.
* Functions.
* Macros.
* Semantic analysis.
* Cross-file analysis.
* Multi-token rule patterns.
* Nested rule blocks.
* Boolean expressions.
* Custom diagnostic codes.
* Autofix support.
* Configuration inheritance.

The DSL now includes syntax node structure declarations, named child captures, and syntax-tree-based rule evaluation,
but semantic analysis and richer AST-style queries are still outside the current scope.

These may be added later if justified by concrete requirements.

# Evolution Rules

The DSL should grow incrementally.
Do not add syntax for future flexibility unless it is required by current behavior.

When adding DSL features:

1. Update this document first.
2. Define syntax.
3. Define semantics.
4. Define validation rules.
5. Add examples.
6. Implement parser support.
7. Implement validation support.
8. Add tests

A DSL feature is not complete until it is documented, parsed, validated, tested, and executable.
