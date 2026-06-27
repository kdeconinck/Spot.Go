// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package lexer converts Spot DSL source text into syntax tokens.
package lexer

import (
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/location"
)

// Lexer reads Spot DSL source text one token at a time.
type Lexer struct {
	src    string
	offset int
}

// New returns a lexer for src.
func New(src string) Lexer {
	return Lexer{
		src: src,
	}
}

// Next returns the next syntax token from the source text.
func (lexer *Lexer) Next() token.Token {
	lexer.skipTrivia()

	if lexer.offset >= len(lexer.src) {
		return lexer.token(token.TokenEOF, lexer.offset, lexer.offset)
	}

	start := lexer.offset

	switch lexer.src[lexer.offset] {
	case '"':
		return lexer.scanString(start)

	case '\'':
		return lexer.scanCharacter(start)

	case '.':
		lexer.offset++

		if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '.' {
			lexer.offset++

			return lexer.token(token.TokenDotDot, start, lexer.offset)
		}

		return lexer.token(token.TokenDot, start, lexer.offset)

	case '=':
		lexer.offset++

		if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '=' {
			lexer.offset++

			return lexer.token(token.TokenEqualEqual, start, lexer.offset)
		}

		return lexer.token(token.TokenEqual, start, lexer.offset)

	case '!':
		lexer.offset++

		if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '=' {
			lexer.offset++

			return lexer.token(token.TokenBangEqual, start, lexer.offset)
		}

		return lexer.token(token.TokenInvalid, start, lexer.offset)

	case '<':
		lexer.offset++

		if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '=' {
			lexer.offset++

			return lexer.token(token.TokenLessEqual, start, lexer.offset)
		}

		return lexer.token(token.TokenLess, start, lexer.offset)

	case '>':
		lexer.offset++

		if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '=' {
			lexer.offset++

			return lexer.token(token.TokenGreaterEqual, start, lexer.offset)
		}

		return lexer.token(token.TokenGreater, start, lexer.offset)

	case '(':
		lexer.offset++

		return lexer.token(token.TokenLeftParen, start, lexer.offset)

	case ')':
		lexer.offset++

		return lexer.token(token.TokenRightParen, start, lexer.offset)

	case '{':
		lexer.offset++

		return lexer.token(token.TokenLeftBrace, start, lexer.offset)

	case '}':
		lexer.offset++

		return lexer.token(token.TokenRightBrace, start, lexer.offset)

	case '|':
		lexer.offset++

		return lexer.token(token.TokenPipe, start, lexer.offset)

	case '?':
		lexer.offset++

		return lexer.token(token.TokenQuestion, start, lexer.offset)

	case '*':
		lexer.offset++

		return lexer.token(token.TokenStar, start, lexer.offset)

	case '+':
		lexer.offset++

		return lexer.token(token.TokenPlus, start, lexer.offset)
	}

	if isIdentifierStart(lexer.src[lexer.offset]) {
		return lexer.scanIdentifier(start)
	}

	if isDigit(lexer.src[lexer.offset]) {
		return lexer.scanInteger(start)
	}

	lexer.offset++

	return lexer.token(token.TokenInvalid, start, lexer.offset)
}

func (lexer *Lexer) skipTrivia() {
	for lexer.offset < len(lexer.src) {
		switch lexer.src[lexer.offset] {
		case ' ', '\t', '\r', '\n':
			lexer.offset++

		case '/':
			if lexer.peek(1) != '/' {
				return
			}

			lexer.skipLineComment()

		default:
			return
		}
	}
}

func (lexer *Lexer) skipLineComment() {
	lexer.offset += 2

	for lexer.offset < len(lexer.src) && lexer.src[lexer.offset] != '\n' {
		lexer.offset++
	}
}

func (lexer Lexer) peek(delta int) byte {
	offset := lexer.offset + delta

	if offset >= len(lexer.src) {
		return 0
	}

	return lexer.src[offset]
}

func (lexer *Lexer) scanIdentifier(start int) token.Token {
	lexer.offset++

	for lexer.offset < len(lexer.src) && isIdentifierPart(lexer.src[lexer.offset]) {
		lexer.offset++
	}

	tokKind := token.LookupTokenKind(lexer.src[start:lexer.offset])

	return lexer.token(tokKind, start, lexer.offset)
}

func (lexer *Lexer) scanInteger(start int) token.Token {
	lexer.offset++

	for lexer.offset < len(lexer.src) && isDigit(lexer.src[lexer.offset]) {
		lexer.offset++
	}

	return lexer.token(token.TokenInteger, start, lexer.offset)
}

func (lexer *Lexer) scanCharacter(start int) token.Token {
	lexer.offset++

	if lexer.offset >= len(lexer.src) {
		return lexer.token(token.TokenInvalid, start, lexer.offset)
	}

	switch lexer.src[lexer.offset] {
	case '\\':
		if !lexer.hasValidCharacterEscape() {
			lexer.offset++

			return lexer.token(token.TokenInvalid, start, lexer.offset)
		}

		lexer.offset += 2

	case '\n', '\r':
		return lexer.token(token.TokenInvalid, start, lexer.offset)

	default:
		lexer.offset++
	}

	if lexer.offset >= len(lexer.src) || lexer.src[lexer.offset] != '\'' {
		return lexer.token(token.TokenInvalid, start, lexer.offset)
	}

	lexer.offset++

	return lexer.token(token.TokenCharacter, start, lexer.offset)
}

func (lexer Lexer) hasValidCharacterEscape() bool {
	switch lexer.peek(1) {
	case '\\', '\'', 'n', 'r', 't':
		return true

	default:
		return false
	}
}

func (lexer *Lexer) scanString(start int) token.Token {
	lexer.offset++

	for lexer.offset < len(lexer.src) {
		switch lexer.src[lexer.offset] {
		case '"':
			lexer.offset++

			return lexer.token(token.TokenString, start, lexer.offset)

		case '\\':
			if !lexer.hasValidStringEscape() {
				lexer.offset++

				return lexer.token(token.TokenInvalid, start, lexer.offset)
			}

			lexer.offset += 2

		case '\n', '\r':
			return lexer.token(token.TokenInvalid, start, lexer.offset)

		default:
			lexer.offset++
		}
	}

	return lexer.token(token.TokenInvalid, start, lexer.offset)
}

func (lexer Lexer) hasValidStringEscape() bool {
	switch lexer.peek(1) {
	case '\\', '"', 'n', 'r', 't':
		return true

	default:
		return false
	}
}

func (lexer Lexer) token(kind token.TokenKind, start, end int) token.Token {
	return token.Token{
		Kind: kind,
		Text: lexer.src[start:end],
		Span: location.Span{
			Start: location.Position(start),
			End:   location.Position(end),
		},
	}
}

func isIdentifierStart(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

func isIdentifierPart(ch byte) bool {
	return isIdentifierStart(ch) || ('0' <= ch && ch <= '9') || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
