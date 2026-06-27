// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package lexer converts Spot DSL source text into syntax tokens.
package lexer

import "github.com/kdeconinck/spot/dsl/token"

func (lexer *Lexer) scanCharacter(start int) token.Token {
	lexer.offset++

	if lexer.offset >= len(lexer.src) {
		return lexer.makeToken(token.TokenInvalid, start, lexer.offset)
	}

	switch lexer.src[lexer.offset] {
	case '\\':
		if !lexer.hasValidCharacterEscape() {
			lexer.offset++

			return lexer.makeToken(token.TokenInvalid, start, lexer.offset)
		}

		lexer.offset += 2

	case '\n', '\r':
		return lexer.makeToken(token.TokenInvalid, start, lexer.offset)

	default:
		lexer.offset++
	}

	if lexer.offset >= len(lexer.src) || lexer.src[lexer.offset] != '\'' {
		return lexer.makeToken(token.TokenInvalid, start, lexer.offset)
	}

	lexer.offset++

	return lexer.makeToken(token.TokenCharacter, start, lexer.offset)
}
