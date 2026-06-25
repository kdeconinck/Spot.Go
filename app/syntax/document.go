// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package syntax defines data structures that represent Spot DSL syntax.
package syntax

import "github.com/kdeconinck/spot/location"

// Document is the root syntax node for a Spot DSL file.
type Document struct {
	// Scope is the parsed scope section.
	Scope ScopeSection

	// Definitions is the parsed definitions section.
	Definitions DefinitionsSection

	// Tokens is the parsed tokens section.
	Tokens TokensSection

	// Rules is the parsed rules section.
	Rules RulesSection

	// Span is the byte range covered by the document.
	Span location.Span
}

// ScopeSection is a parsed scope section.
type ScopeSection struct {
	// Entries are the include and exclude declarations inside the scope section.
	Entries []ScopeEntry

	// Span is the byte range covered by the scope section.
	Span location.Span
}

// ScopeEntryKind identifies the kind of scope entry.
type ScopeEntryKind uint8

const (
	// ScopeEntryInclude is an include pattern entry.
	ScopeEntryInclude ScopeEntryKind = iota

	// ScopeEntryExclude is an exclude pattern entry.
	ScopeEntryExclude
)

// ScopeEntry is an include or exclude declaration inside a scope section.
type ScopeEntry struct {
	// Kind identifies whether the entry includes or excludes files.
	Kind ScopeEntryKind

	// Pattern is the string literal token containing the file pattern.
	Pattern Token

	// Span is the byte range covered by the scope entry.
	Span location.Span
}

// DefinitionsSection is a parsed definitions section.
type DefinitionsSection struct {
	// Definitions are the declarations inside the definitions section.
	Definitions []Definition

	// Span is the byte range covered by the definitions section.
	Span location.Span
}

// Definition is a reusable character-level expression declaration.
type Definition struct {
	// Name is the identifier token naming the definition.
	Name Token

	// Expression is the character-level expression assigned to the definition.
	Expression DefinitionExpression

	// Span is the byte range covered by the definition.
	Span location.Span
}

// TokensSection is a parsed tokens section.
type TokensSection struct {
	// Tokens are the declarations inside the tokens section.
	Tokens []TokenDefinition

	// Span is the byte range covered by the tokens section.
	Span location.Span
}

// TokenDefinition is a token expression declaration.
type TokenDefinition struct {
	// Name is the identifier token naming the emitted token.
	Name Token

	// Expression is the expression assigned to the token.
	Expression DefinitionExpression

	// Skip marks a token declaration whose matches are not emitted.
	Skip Token

	// Span is the byte range covered by the token definition.
	Span location.Span
}

// RulesSection is a parsed rules section.
type RulesSection struct {
	// Rules are the declarations inside the rules section.
	Rules []Rule

	// Span is the byte range covered by the rules section.
	Span location.Span
}

// Rule is a diagnostic declaration over matched tokens.
type Rule struct {
	// Name is the identifier token naming the rule.
	Name Token

	// Match is the token match statement for the rule.
	Match RuleMatch

	// Report is the diagnostic report statement for the rule.
	Report RuleReport

	// Span is the byte range covered by the rule.
	Span location.Span
}

// RuleMatch is a token match statement inside a rule.
type RuleMatch struct {
	// Token is the identifier token naming the token to match.
	Token Token

	// Span is the byte range covered by the match statement.
	Span location.Span
}

// RuleReport is a diagnostic report statement inside a rule.
type RuleReport struct {
	// Severity is the diagnostic severity token.
	Severity Token

	// Target is the identifier token whose span receives the diagnostic.
	Target Token

	// Message is the string literal token containing the diagnostic text.
	Message Token

	// Span is the byte range covered by the report statement.
	Span location.Span
}

// DefinitionExpressionKind identifies the form of a definition expression.
type DefinitionExpressionKind uint8

const (
	// DefinitionExpressionCharacter is a single character literal expression.
	DefinitionExpressionCharacter DefinitionExpressionKind = iota

	// DefinitionExpressionString is a string literal expression.
	DefinitionExpressionString

	// DefinitionExpressionRange is a character range expression.
	DefinitionExpressionRange

	// DefinitionExpressionReference is a reference to another definition.
	DefinitionExpressionReference

	// DefinitionExpressionConcatenation is a sequence of adjacent expressions.
	DefinitionExpressionConcatenation

	// DefinitionExpressionAlternation is a list of alternative expressions.
	DefinitionExpressionAlternation

	// DefinitionExpressionGroup is a parenthesized expression.
	DefinitionExpressionGroup

	// DefinitionExpressionRepetition is a repeated expression.
	DefinitionExpressionRepetition
)

// DefinitionExpression is a parsed character-level definition expression.
type DefinitionExpression struct {
	// Kind identifies the form of expression.
	Kind DefinitionExpressionKind

	// Start is the first token in the expression.
	Start Token

	// End is the final character literal in a range expression.
	End Token

	// Operator is the postfix operator token in a repetition expression.
	Operator Token

	// Terms are the child expressions in an alternation or concatenation expression.
	Terms []DefinitionExpression

	// Inner is the expression contained in a grouped or repetition expression.
	Inner *DefinitionExpression

	// Span is the byte range covered by the expression.
	Span location.Span
}
