// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package scanner converts Spot DSL source text into syntax tokens.
package scanner

import (
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/syntax"
)

var keywordMap = map[string]syntax.TokenKind{
	"scope":       syntax.TokenScope,
	"include":     syntax.TokenInclude,
	"exclude":     syntax.TokenExclude,
	"definitions": syntax.TokenDefinitions,
}

// Scanner reads Spot DSL source text one token at a time.
type Scanner struct {
	src    string
	offset int
}

// New returns a scanner for src.
func New(src string) Scanner {
	return Scanner{
		src: src,
	}
}

// Next returns the next syntax token from the source text.
func (scanner *Scanner) Next() syntax.Token {
	scanner.skipTrivia()

	if scanner.offset >= len(scanner.src) {
		return scanner.token(syntax.TokenEOF, scanner.offset, scanner.offset)
	}

	start := scanner.offset

	switch scanner.src[scanner.offset] {
	case '"':
		return scanner.scanString(start)

	case '\'':
		return scanner.scanCharacter(start)

	case '.':
		scanner.offset++

		if scanner.offset < len(scanner.src) && scanner.src[scanner.offset] == '.' {
			scanner.offset++

			return scanner.token(syntax.TokenDotDot, start, scanner.offset)
		}

		return scanner.token(syntax.TokenInvalid, start, scanner.offset)

	case '=':
		scanner.offset++

		return scanner.token(syntax.TokenEqual, start, scanner.offset)

	case '(':
		scanner.offset++

		return scanner.token(syntax.TokenLeftParen, start, scanner.offset)

	case ')':
		scanner.offset++

		return scanner.token(syntax.TokenRightParen, start, scanner.offset)

	case '{':
		scanner.offset++

		return scanner.token(syntax.TokenLeftBrace, start, scanner.offset)

	case '}':
		scanner.offset++

		return scanner.token(syntax.TokenRightBrace, start, scanner.offset)

	case '|':
		scanner.offset++

		return scanner.token(syntax.TokenPipe, start, scanner.offset)
	}

	if isIdentifierStart(scanner.src[scanner.offset]) {
		return scanner.scanIdentifier(start)
	}

	scanner.offset++

	return scanner.token(syntax.TokenInvalid, start, scanner.offset)
}

func (scanner *Scanner) skipTrivia() {
	for scanner.offset < len(scanner.src) {
		switch scanner.src[scanner.offset] {
		case ' ', '\t', '\r', '\n':
			scanner.offset++

		case '/':
			if scanner.peek(1) != '/' {
				return
			}

			scanner.skipLineComment()

		default:
			return
		}
	}
}

func (scanner *Scanner) skipLineComment() {
	scanner.offset += 2

	for scanner.offset < len(scanner.src) && scanner.src[scanner.offset] != '\n' {
		scanner.offset++
	}
}

func (scanner Scanner) peek(delta int) byte {
	offset := scanner.offset + delta

	if offset >= len(scanner.src) {
		return 0
	}

	return scanner.src[offset]
}

func (scanner *Scanner) scanIdentifier(start int) syntax.Token {
	scanner.offset++

	for scanner.offset < len(scanner.src) && isIdentifierPart(scanner.src[scanner.offset]) {
		scanner.offset++
	}

	kind, ok := keywordMap[scanner.src[start:scanner.offset]]

	if !ok {
		kind = syntax.TokenIdentifier
	}

	return scanner.token(kind, start, scanner.offset)
}

func (scanner *Scanner) scanCharacter(start int) syntax.Token {
	scanner.offset++

	if scanner.offset >= len(scanner.src) {
		return scanner.token(syntax.TokenInvalid, start, scanner.offset)
	}

	switch scanner.src[scanner.offset] {
	case '\\':
		if !scanner.hasValidCharacterEscape() {
			scanner.offset++

			return scanner.token(syntax.TokenInvalid, start, scanner.offset)
		}

		scanner.offset += 2

	case '\n', '\r':
		return scanner.token(syntax.TokenInvalid, start, scanner.offset)

	default:
		scanner.offset++
	}

	if scanner.offset >= len(scanner.src) || scanner.src[scanner.offset] != '\'' {
		return scanner.token(syntax.TokenInvalid, start, scanner.offset)
	}

	scanner.offset++

	return scanner.token(syntax.TokenCharacter, start, scanner.offset)
}

func (scanner Scanner) hasValidCharacterEscape() bool {
	switch scanner.peek(1) {
	case '\\', '\'', 'n', 'r', 't':
		return true

	default:
		return false
	}
}

func (scanner *Scanner) scanString(start int) syntax.Token {
	scanner.offset++

	for scanner.offset < len(scanner.src) {
		switch scanner.src[scanner.offset] {
		case '"':
			scanner.offset++

			return scanner.token(syntax.TokenString, start, scanner.offset)

		case '\\':
			if !scanner.hasValidStringEscape() {
				scanner.offset++

				return scanner.token(syntax.TokenInvalid, start, scanner.offset)
			}

			scanner.offset += 2

		case '\n', '\r':
			return scanner.token(syntax.TokenInvalid, start, scanner.offset)

		default:
			scanner.offset++
		}
	}

	return scanner.token(syntax.TokenInvalid, start, scanner.offset)
}

func (scanner Scanner) hasValidStringEscape() bool {
	switch scanner.peek(1) {
	case '\\', '"', 'n', 'r', 't':
		return true

	default:
		return false
	}
}

func (scanner Scanner) token(kind syntax.TokenKind, start, end int) syntax.Token {
	return syntax.Token{
		Kind: kind,
		Text: scanner.src[start:end],
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
