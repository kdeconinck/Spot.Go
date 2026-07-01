// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package token defines data structures that represent Spot DSL syntax.
package token

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

	// TokenFallback marks a token fallback declaration.
	TokenFallback

	// TokenRules marks the 'rules' section keyword.
	TokenRules

	// TokenSyntax marks the 'syntax' section keyword.
	TokenSyntax

	// TokenNode marks a syntax node declaration.
	TokenNode

	// TokenOneOf marks a syntax-node variant block.
	TokenOneOf

	// TokenAny marks a syntax expression that matches any emitted token.
	TokenAny

	// TokenNot marks selector negation.
	TokenNot

	// TokenStartsWith marks a string-prefix comparison operator in rule conditions.
	TokenStartsWith

	// TokenRule marks a rule declaration.
	TokenRule

	// TokenMatch marks a rule token match statement.
	TokenMatch

	// TokenWhere marks a rule condition statement.
	TokenWhere

	// TokenInside marks a syntax-rule ancestor inclusion constraint.
	TokenInside

	// TokenOutside marks a syntax-rule ancestor exclusion constraint.
	TokenOutside

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

	// TokenColon marks the start of a selector query.
	TokenColon

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

// LookupTokenKind returns the TokenKind corresponding to value of TokenIdentifier if value doesn't match any TokenKind.
func LookupTokenKind(value string) TokenKind {
	switch value {
	case "scope":
		return TokenScope

	case "include":
		return TokenInclude

	case "exclude":
		return TokenExclude

	case "definitions":
		return TokenDefinitions

	case "tokens":
		return TokenTokens

	case "skip":
		return TokenSkip

	case "fallback":
		return TokenFallback

	case "rules":
		return TokenRules

	case "syntax":
		return TokenSyntax

	case "node":
		return TokenNode

	case "oneOf":
		return TokenOneOf

	case "any":
		return TokenAny

	case "not":
		return TokenNot

	case "startsWith":
		return TokenStartsWith

	case "rule":
		return TokenRule

	case "match":
		return TokenMatch

	case "where":
		return TokenWhere

	case "inside":
		return TokenInside

	case "outside":
		return TokenOutside

	case "report":
		return TokenReport

	case "info":
		return TokenInfo

	case "warn":
		return TokenWarn

	case "err":
		return TokenErr

	case "at":
		return TokenAt

	default:
		return TokenIdentifier
	}
}

// Token is a lexical token from a Spot DSL source file.
type Token struct {
	// Kind identifies the syntactic role of the token.
	Kind TokenKind

	// Span is the byte range covered by the token.
	Span location.Span
}

// Value returns the slice of the original source code that this token covers.
func (tok Token) Value(source string) string {
	return source[tok.Span.Start:tok.Span.End]
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

	case TokenFallback:
		return "fallback"

	case TokenRules:
		return "rules"

	case TokenSyntax:
		return "syntax"

	case TokenNode:
		return "node"

	case TokenOneOf:
		return "oneOf"

	case TokenAny:
		return "any"

	case TokenNot:
		return "not"

	case TokenStartsWith:
		return "startsWith"

	case TokenRule:
		return "rule"

	case TokenMatch:
		return "match"

	case TokenWhere:
		return "where"

	case TokenInside:
		return "inside"

	case TokenOutside:
		return "outside"

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

	case TokenColon:
		return ":"

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
