// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package compiler compiles validated Spot DSL syntax into runtime-oriented data structures.
package compiler

// Program is a compiled token program.
type Program struct {
	// Tokens are the compiled token definitions in source order.
	Tokens []Token
}

// Token is a compiled token definition.
type Token struct {
	// Name is the emitted token name.
	Name string

	// Expression is the compiled token expression.
	Expression Expression

	// Skip reports whether matches for this token should be emitted.
	Skip bool
}

// ExpressionKind identifies the form of a compiled expression.
type ExpressionKind uint8

const (
	// ExpressionCharacter is a single character match.
	ExpressionCharacter ExpressionKind = iota

	// ExpressionString is an exact string match.
	ExpressionString

	// ExpressionRange is a single character range match.
	ExpressionRange

	// ExpressionConcatenation is a sequence of expressions.
	ExpressionConcatenation

	// ExpressionAlternation is a list of alternative expressions.
	ExpressionAlternation

	// ExpressionRepetition is a repeated expression.
	ExpressionRepetition
)

// RepetitionKind identifies the form of a repetition expression.
type RepetitionKind uint8

const (
	// RepetitionZeroOrOne matches the inner expression zero or one time.
	RepetitionZeroOrOne RepetitionKind = iota

	// RepetitionZeroOrMore matches the inner expression zero or more times.
	RepetitionZeroOrMore

	// RepetitionOneOrMore matches the inner expression one or more times.
	RepetitionOneOrMore
)

// Expression is a compiled token expression.
type Expression struct {
	// Kind identifies the form of expression.
	Kind ExpressionKind

	// Character is the matched byte in a character expression.
	Character byte

	// String is the matched text in a string expression.
	String string

	// RangeStart is the inclusive start of a range expression.
	RangeStart byte

	// RangeEnd is the inclusive end of a range expression.
	RangeEnd byte

	// Terms are the child expressions in a concatenation or alternation expression.
	Terms []Expression

	// Inner is the repeated expression in a repetition expression.
	Inner *Expression

	// Repetition identifies the repetition operator.
	Repetition RepetitionKind
}
