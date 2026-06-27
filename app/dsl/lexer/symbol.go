// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package lexer converts Spot DSL source text into syntax tokens.
package lexer

import "github.com/kdeconinck/spot/dsl/token"

func (lexer *Lexer) scanSymbol(start int, character byte) (token.Token, bool) {
	switch character {
	case '.':
		return lexer.scanDot(start), true

	case '=':
		return lexer.scanEqual(start), true

	case '!':
		return lexer.scanBang(start), true

	case '<':
		return lexer.scanLess(start), true

	case '>':
		return lexer.scanGreater(start), true

	case '(':
		return lexer.scanSingleCharacter(start, token.TokenLeftParen), true

	case ')':
		return lexer.scanSingleCharacter(start, token.TokenRightParen), true

	case '{':
		return lexer.scanSingleCharacter(start, token.TokenLeftBrace), true

	case '}':
		return lexer.scanSingleCharacter(start, token.TokenRightBrace), true

	case '|':
		return lexer.scanSingleCharacter(start, token.TokenPipe), true

	case '?':
		return lexer.scanSingleCharacter(start, token.TokenQuestion), true

	case '*':
		return lexer.scanSingleCharacter(start, token.TokenStar), true

	case '+':
		return lexer.scanSingleCharacter(start, token.TokenPlus), true

	default:
		return token.Token{}, false
	}
}
