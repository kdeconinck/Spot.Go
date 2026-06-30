// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package ast defines data structures that represent the AST (Abstract Syntax Tree) for Spot DSL syntax.
package ast

import (
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/location"
)

// RulesSection is a parsed rules section.
type RulesSection struct {
	// FirstElementIdx is the index of the section's first rule in Document.RuleList.
	FirstElementIdx uint32

	// AmountOfElements is the number of rules in the section.
	AmountOfElements uint32

	// Span is the byte range covered by the rules section.
	Span location.Span
}

// RuleMatchKind identifies what a rule matches.
type RuleMatchKind uint8

const (
	// RuleMatchToken matches one emitted token.
	RuleMatchToken RuleMatchKind = iota

	// RuleMatchNode matches one runtime syntax node.
	RuleMatchNode
)

// RuleMatchScopeKind identifies whether a syntax-node rule constrains ancestor nodes.
type RuleMatchScopeKind uint8

const (
	// RuleMatchScopeNone does not constrain ancestor nodes.
	RuleMatchScopeNone RuleMatchScopeKind = iota

	// RuleMatchScopeInside requires the matched syntax node to be inside the named ancestor syntax node.
	RuleMatchScopeInside

	// RuleMatchScopeOutside requires the matched syntax node to be outside the named ancestor syntax node.
	RuleMatchScopeOutside
)

// Rule is a diagnostic declaration over matched tokens or syntax nodes.
type Rule struct {
	// Name is the identifier token naming the rule.
	Name token.Token

	// Match is the token match statement for the rule.
	Match RuleMatch

	// Where is the optional condition statement for the rule.
	Where RuleCondition

	// Report is the diagnostic report statement for the rule.
	Report RuleReport

	// Span is the byte range covered by the rule.
	Span location.Span
}

// RuleMatch is a match statement inside a rule.
type RuleMatch struct {
	// Kind identifies whether the rule matches a token or a syntax node.
	Kind RuleMatchKind

	// Target is the identifier token naming the matched token or syntax node.
	Target token.Token

	// ScopeKind identifies whether the match constrains ancestor syntax nodes.
	ScopeKind RuleMatchScopeKind

	// ScopeTarget is the identifier token naming the ancestor syntax node for inside/outside constraints.
	ScopeTarget token.Token

	// Span is the byte range covered by the match statement.
	Span location.Span
}

// RuleCondition is a comparison condition inside a rule.
type RuleCondition struct {
	// Subject is the identifier token naming the token being inspected.
	Subject token.Token

	// Property is the identifier token naming the inspected token property.
	Property token.Token

	// Operator is the comparison operator token.
	Operator token.Token

	// Value is the literal token compared against the property.
	Value token.Token

	// Span is the byte range covered by the where statement.
	Span location.Span
}

// RuleReport is a diagnostic report statement inside a rule.
type RuleReport struct {
	// Severity is the diagnostic severity token.
	Severity token.Token

	// Target is the identifier token whose span receives the diagnostic.
	Target token.Token

	// Message is the string literal token containing the diagnostic text.
	Message token.Token

	// Span is the byte range covered by the report statement.
	Span location.Span
}
