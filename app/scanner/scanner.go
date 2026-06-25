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
	"scope": syntax.TokenScope,
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
	scanner.skipWhitespace()

	if scanner.offset >= len(scanner.src) {
		return scanner.token(syntax.TokenEOF, scanner.offset, scanner.offset)
	}

	start := scanner.offset

	switch scanner.src[scanner.offset] {
	case '{':
		scanner.offset++

		return scanner.token(syntax.TokenLeftBrace, start, scanner.offset)

	case '}':
		scanner.offset++

		return scanner.token(syntax.TokenRightBrace, start, scanner.offset)
	}

	if isIdentifierStart(scanner.src[scanner.offset]) {
		return scanner.scanIdentifier(start)
	}

	scanner.offset++

	return scanner.token(syntax.TokenInvalid, start, scanner.offset)
}

func (scanner *Scanner) skipWhitespace() {
	for scanner.offset < len(scanner.src) {
		switch scanner.src[scanner.offset] {
		case ' ', '\t', '\r', '\n':
			scanner.offset++

		default:
			return
		}
	}
}

func (scanner *Scanner) scanIdentifier(start int) syntax.Token {
	scanner.offset++

	for scanner.offset < len(scanner.src) && isIdentifierPart(scanner.src[scanner.offset]) {
		scanner.offset++
	}

	kind, ok := keywordMap[scanner.src[start:scanner.offset]]

	if !ok {
		kind = syntax.TokenInvalid
	}

	return scanner.token(kind, start, scanner.offset)
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
