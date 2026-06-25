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

	// TokenLeftBrace marks the start of a section block.
	TokenLeftBrace

	// TokenRightBrace marks the end of a section block.
	TokenRightBrace

	// TokenString marks a double-quoted string literal.
	TokenString
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
