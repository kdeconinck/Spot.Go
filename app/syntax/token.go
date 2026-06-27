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

	// TokenInclude marks a scope include entry.
	TokenInclude

	// TokenExclude marks a scope exclude entry.
	TokenExclude

	// TokenDefinitions marks the 'definitions' section keyword.
	TokenDefinitions

	// TokenTokens marks the 'tokens' section keyword.
	TokenTokens

	// TokenSkip marks a skipped token declaration.
	TokenSkip

	// TokenRules marks the 'rules' section keyword.
	TokenRules

	// TokenRule marks a rule declaration.
	TokenRule

	// TokenMatch marks a rule token match statement.
	TokenMatch

	// TokenWhere marks a rule condition statement.
	TokenWhere

	// TokenReport marks a rule report statement.
	TokenReport

	// TokenInfo marks an informational diagnostic severity.
	TokenInfo

	// TokenWarn marks a warning diagnostic severity.
	TokenWarn

	// TokenErr marks an error diagnostic severity.
	TokenErr

	// TokenAt marks a rule report target.
	TokenAt

	// TokenString marks a double-quoted string literal.
	TokenString

	// TokenInteger marks a decimal integer literal.
	TokenInteger

	// TokenIdentifier marks a user-defined DSL name.
	TokenIdentifier

	// TokenCharacter marks a single-quoted character literal.
	TokenCharacter

	// TokenLeftBrace marks the start of a section block.
	TokenLeftBrace

	// TokenRightBrace marks the end of a section block.
	TokenRightBrace

	// TokenLeftParen marks the start of an expression group.
	TokenLeftParen

	// TokenRightParen marks the end of an expression group.
	TokenRightParen

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
)

var keywordMap = map[string]TokenKind{
	"scope":       TokenScope,
	"include":     TokenInclude,
	"exclude":     TokenExclude,
	"definitions": TokenDefinitions,
	"tokens":      TokenTokens,
	"skip":        TokenSkip,
	"rules":       TokenRules,
	"rule":        TokenRule,
	"match":       TokenMatch,
	"where":       TokenWhere,
	"report":      TokenReport,
	"info":        TokenInfo,
	"warn":        TokenWarn,
	"err":         TokenErr,
	"at":          TokenAt,
}

// LookupTokenKind returns the TokenKind corresponding to value of TokenIdentifier if value doesn't match any TokenKind.
func LookupTokenKind(value string) TokenKind {
	if v, ok := keywordMap[value]; ok {
		return v
	}

	return TokenIdentifier
}

// Token is a lexical token from a Spot DSL source file.
type Token struct {
	// Kind identifies the syntactic role of the token.
	Kind TokenKind

	// Text is the exact source text covered by the token.
	Text string

	// Span is the byte range covered by the token.
	Span location.Span
}

// String returns a stable display name for kind.
func (kind TokenKind) String() string {
	switch kind {
	case TokenInvalid:
		return "invalid"

	case TokenEOF:
		return "EOF"

	case TokenScope:
		return "scope"

	case TokenInclude:
		return "include"

	case TokenExclude:
		return "exclude"

	case TokenDefinitions:
		return "definitions"

	case TokenTokens:
		return "tokens"

	case TokenSkip:
		return "skip"

	case TokenRules:
		return "rules"

	case TokenRule:
		return "rule"

	case TokenMatch:
		return "match"

	case TokenWhere:
		return "where"

	case TokenReport:
		return "report"

	case TokenInfo:
		return "info"

	case TokenWarn:
		return "warn"

	case TokenErr:
		return "err"

	case TokenAt:
		return "at"

	case TokenString:
		return "string"

	case TokenInteger:
		return "integer"

	case TokenIdentifier:
		return "identifier"

	case TokenCharacter:
		return "character"

	case TokenLeftBrace:
		return "{"

	case TokenRightBrace:
		return "}"

	case TokenLeftParen:
		return "("

	case TokenRightParen:
		return ")"

	case TokenEqual:
		return "="

	case TokenEqualEqual:
		return "=="

	case TokenBangEqual:
		return "!="

	case TokenLess:
		return "<"

	case TokenLessEqual:
		return "<="

	case TokenGreater:
		return ">"

	case TokenGreaterEqual:
		return ">="

	case TokenDot:
		return "."

	case TokenDotDot:
		return ".."

	case TokenPipe:
		return "|"

	case TokenQuestion:
		return "?"

	case TokenStar:
		return "*"

	case TokenPlus:
		return "+"

	default:
		return "unknown"
	}
}
