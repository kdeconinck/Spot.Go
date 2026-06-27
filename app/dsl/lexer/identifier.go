// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package lexer converts Spot DSL source text into syntax tokens.
package lexer

import "github.com/kdeconinck/spot/dsl/token"

func (lexer *Lexer) scanIdentifier(start int) token.Token {
	lexer.offset++

	for lexer.offset < len(lexer.src) && isIdentifierPart(lexer.src[lexer.offset]) {
		lexer.offset++
	}

	tokKind := token.LookupTokenKind(lexer.src[start:lexer.offset])

	return lexer.makeToken(tokKind, start, lexer.offset)
}
