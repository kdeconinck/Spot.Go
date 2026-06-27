// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package lexer converts Spot DSL source text into syntax tokens.
package lexer

import "github.com/kdeconinck/spot/dsl/token"

func (lexer *Lexer) scanString(start int) token.Token {
	lexer.offset++

	for lexer.offset < len(lexer.src) {
		switch lexer.src[lexer.offset] {
		case '"':
			lexer.offset++

			return lexer.makeToken(token.TokenString, start, lexer.offset)

		case '\\':
			if !lexer.hasValidStringEscape() {
				lexer.offset++

				return lexer.makeToken(token.TokenInvalid, start, lexer.offset)
			}

			lexer.offset += 2

		case '\n', '\r':
			return lexer.makeToken(token.TokenInvalid, start, lexer.offset)

		default:
			lexer.offset++
		}
	}

	return lexer.makeToken(token.TokenInvalid, start, lexer.offset)
}
