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
    node FileHeader
    node PackageClause
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
* `rule`
* `match`
* `where`
* `report`
* `skip`
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
3. If no token matches, scanning fails with a diagnostic.

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
    node Word = Identifier | KeywordPublic
    node WordPair = Word Word
    node OptionalWord = (Word | KeywordInternal)?
    node WordList = Word+
}
```

Each declaration introduces one syntax node kind together with a syntax expression:

```spot
node UsingStatement = KeywordUsing QualifiedIdentifier Semicolon
```

Syntax expressions may contain:

* Token references.
* Syntax node references.
* Grouping.
* Concatenation.
* Alternation.
* Repetition.

Examples:

```spot
node Word = Identifier | KeywordPublic
node WordPair = Word Word
node OptionalWord = (Word | KeywordInternal)?
node WordList = Word+
```

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

* It only describes syntax node structure.
* It does not yet build runtime syntax trees from token streams.
* It does not yet introduce AST-based rules.

This means the section is parsed and validated today, but it does not yet affect scanning, compilation, or rule
execution.

# Rule Match

The current rule model matches a single token.

```spot
match Identifier
```

This binds the matched token to its token name.
The matched token may then be referenced in `where` and `report`.

Example:

```spot
rule PublicIdentifier {
    match Identifier
    where Identifier.text == "public"
    report warning at Identifier "Public identifier found"
}
```

Multi-token patterns are intentionally outside the current DSL.

# Rule Conditions

A `where` clause filters matched tokens.

```spot
where Identifier.text == "public"
```

The current condition model supports comparisons against token properties.
.
Supported token properties:

| Property | Meaning                             |
| -------- | ----------------------------------- |
| `text`   | The exact source text of the token. |
| `length` | The byte length of the token text.  |

Supported comparison operators depend on the property:

| Property | Operators                        |
| -------- | -------------------------------- |
| `text`   | `==`, `!=`                       |
| `length` | `==`, `!=`, `<`, `<=`, `>`, `>=` |

Examples:

```spot
where Identifier.text == "public"
where Whitespace.length > 1
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
report severity at tokenName "message"
```

Supported severities:

* Informational (`info`).
* Warning (`warn`).
* Error (`err`).

The `at` target must reference a token matched by the rule.
The diagnostic span is the span of the referenced token.

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

Invalid:

```spot
tokens {
    Empty = digit*
}
```

because it can match an empty string.

## Rule Validation

* A rule must contain exactly one `match`.
* A rule must contain exactly one `report`.
* A `where` clause may only reference the matched token.
* A `report` target must reference the matched token.
* A token property must exist.
* A comparison must use compatible value types.

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

The DSL now includes syntax node structure declarations, but runtime syntax-tree construction and AST-based rule
evaluation are still outside the current scope.

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
