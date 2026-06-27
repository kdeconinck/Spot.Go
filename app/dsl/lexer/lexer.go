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

// Next returns the next token from the source text.
func (lexer *Lexer) Next() token.Token {
	lexer.skipTrivia()

	if lexer.offset >= len(lexer.src) {
		return lexer.makeToken(token.TokenEOF, lexer.offset, lexer.offset)
	}

	start := lexer.offset
	character := lexer.src[lexer.offset]

	switch character {
	case '"':
		return lexer.scanString(start)

	case '\'':
		return lexer.scanCharacter(start)
	}

	if tok, ok := lexer.scanSymbol(start, character); ok {
		return tok
	}

	if isIdentifierStart(character) {
		return lexer.scanIdentifier(start)
	}

	if isDigit(character) {
		return lexer.scanInteger(start)
	}

	lexer.offset++

	return lexer.makeToken(token.TokenInvalid, start, lexer.offset)
}

func (lexer *Lexer) scanDot(start int) token.Token {
	lexer.offset++

	if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '.' {
		lexer.offset++

		return lexer.makeToken(token.TokenDotDot, start, lexer.offset)
	}

	return lexer.makeToken(token.TokenDot, start, lexer.offset)
}

func (lexer *Lexer) scanEqual(start int) token.Token {
	lexer.offset++

	if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '=' {
		lexer.offset++

		return lexer.makeToken(token.TokenEqualEqual, start, lexer.offset)
	}

	return lexer.makeToken(token.TokenEqual, start, lexer.offset)
}

func (lexer *Lexer) scanBang(start int) token.Token {
	lexer.offset++

	if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '=' {
		lexer.offset++

		return lexer.makeToken(token.TokenBangEqual, start, lexer.offset)
	}

	return lexer.makeToken(token.TokenInvalid, start, lexer.offset)
}

func (lexer *Lexer) scanLess(start int) token.Token {
	lexer.offset++

	if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '=' {
		lexer.offset++

		return lexer.makeToken(token.TokenLessEqual, start, lexer.offset)
	}

	return lexer.makeToken(token.TokenLess, start, lexer.offset)
}

func (lexer *Lexer) scanGreater(start int) token.Token {
	lexer.offset++

	if lexer.offset < len(lexer.src) && lexer.src[lexer.offset] == '=' {
		lexer.offset++

		return lexer.makeToken(token.TokenGreaterEqual, start, lexer.offset)
	}

	return lexer.makeToken(token.TokenGreater, start, lexer.offset)
}

func (lexer *Lexer) scanSingleCharacter(start int, kind token.TokenKind) token.Token {
	lexer.offset++

	return lexer.makeToken(kind, start, lexer.offset)
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

func (lexer Lexer) hasValidCharacterEscape() bool {
	switch lexer.peek(1) {
	case '\\', '\'', 'n', 'r', 't':
		return true

	default:
		return false
	}
}

func (lexer Lexer) hasValidStringEscape() bool {
	switch lexer.peek(1) {
	case '\\', '"', 'n', 'r', 't':
		return true

	default:
		return false
	}
}

func (lexer Lexer) makeToken(kind token.TokenKind, start, end int) token.Token {
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
