// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package syntax defines data structures that represent Spot DSL syntax.
package syntax

import "github.com/kdeconinck/spot/location"

// TokenKind identifies the syntactic role of a DSL token.
type TokenKind uint8

const (
	// TokenInvalid marks source text that cannot be tokenized as valid DSL syntax.
	TokenInvalid TokenKind = iota

	// TokenEOF marks the end of a token stream.
	TokenEOF

	// TokenScope marks the 'scope' section keyword.
	TokenScope

	// TokenDefinitions marks the 'definitions' section keyword.
	TokenDefinitions

	// TokenTokens marks the 'tokens' section keyword.
	TokenTokens

	// TokenRules marks the 'rules' section keyword.
	TokenRules

	// TokenSkip marks a skipped token declaration.
	TokenSkip

	// TokenRule marks a rule declaration.
	TokenRule

	// TokenMatch marks a rule token match statement.
	TokenMatch

	// TokenWhere marks a rule condition statement.
	TokenWhere

	// TokenReport marks a rule report statement.
	TokenReport

	// TokenAt marks a rule report target.
	TokenAt

	// TokenInfo marks an informational diagnostic severity.
	TokenInfo

	// TokenWarn marks a warning diagnostic severity.
	TokenWarn

	// TokenErr marks an error diagnostic severity.
	TokenErr

	// TokenInclude marks a scope include entry.
	TokenInclude

	// TokenExclude marks a scope exclude entry.
	TokenExclude

	// TokenLeftBrace marks the start of a section block.
	TokenLeftBrace

	// TokenRightBrace marks the end of a section block.
	TokenRightBrace

	// TokenLeftParen marks the start of an expression group.
	TokenLeftParen

	// TokenRightParen marks the end of an expression group.
	TokenRightParen

	// TokenString marks a double-quoted string literal.
	TokenString

	// TokenInteger marks a decimal integer literal.
	TokenInteger

	// TokenIdentifier marks a user-defined DSL name.
	TokenIdentifier

	// TokenEqual marks a definition assignment.
	TokenEqual

	// TokenEqualEqual marks an equality comparison operator.
	TokenEqualEqual

	// TokenBangEqual marks an inequality comparison operator.
	TokenBangEqual

	// TokenLess marks a less-than comparison operator.
	TokenLess

	// TokenLessEqual marks a less-than-or-equal comparison operator.
	TokenLessEqual

	// TokenGreater marks a greater-than comparison operator.
	TokenGreater

	// TokenGreaterEqual marks a greater-than-or-equal comparison operator.
	TokenGreaterEqual

	// TokenDot marks a token property access operator.
	TokenDot

	// TokenDotDot marks a character range operator.
	TokenDotDot

	// TokenPipe marks an alternation operator.
	TokenPipe

	// TokenQuestion marks zero-or-one repetition.
	TokenQuestion

	// TokenStar marks zero-or-more repetition.
	TokenStar

	// TokenPlus marks one-or-more repetition.
	TokenPlus

	// TokenCharacter marks a single-quoted character literal.
	TokenCharacter
)

// Token is a lexical token from a Spot DSL source file.
type Token struct {
	// Kind identifies the syntactic role of the token.
	Kind TokenKind

	// Text is the exact source text covered by the token.
	Text string

	// Span is the byte range covered by the token.
	Span location.Span
}
