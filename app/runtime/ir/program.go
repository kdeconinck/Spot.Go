// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package ir defines the runtime-oriented intermediate representation produced by the Spot compiler.
package ir

// Program is a compiled Spot program.
type Program struct {
	// Tokens are the compiled token definitions in source order.
	Tokens []Token

	// Rules are the compiled rule definitions in source order.
	Rules []Rule
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

// Rule is a compiled diagnostic rule.
type Rule struct {
	// Name is the DSL rule name.
	Name string

	// MatchToken is the source-order token index matched by the rule.
	MatchToken int

	// Where is the optional compiled condition.
	Where Condition

	// Report is the compiled diagnostic report.
	Report Report
}

// ConditionProperty identifies the token property read by a rule condition.
type ConditionProperty uint8

const (
	// ConditionPropertyNone marks a rule without a where clause.
	ConditionPropertyNone ConditionProperty = iota

	// ConditionPropertyText reads the matched token text.
	ConditionPropertyText

	// ConditionPropertyLength reads the matched token length.
	ConditionPropertyLength
)

// ConditionOperator identifies the comparison performed by a rule condition.
type ConditionOperator uint8

const (
	// ConditionOperatorEqual compares values for equality.
	ConditionOperatorEqual ConditionOperator = iota

	// ConditionOperatorNotEqual compares values for inequality.
	ConditionOperatorNotEqual

	// ConditionOperatorLess compares whether the left value is less than the right value.
	ConditionOperatorLess

	// ConditionOperatorLessEqual compares whether the left value is less than or equal to the right value.
	ConditionOperatorLessEqual

	// ConditionOperatorGreater compares whether the left value is greater than the right value.
	ConditionOperatorGreater

	// ConditionOperatorGreaterEqual compares whether the left value is greater than or equal to the right value.
	ConditionOperatorGreaterEqual
)

// Condition is a compiled rule where clause.
type Condition struct {
	// Property identifies the matched token property read by the condition.
	Property ConditionProperty

	// Operator identifies the comparison performed by the condition.
	Operator ConditionOperator

	// String is the right-hand string value when Property is ConditionPropertyText.
	String string

	// Integer is the right-hand integer value when Property is ConditionPropertyLength.
	Integer int
}

// Severity identifies the diagnostic severity emitted by a rule.
type Severity uint8

const (
	// SeverityInfo emits an informational diagnostic.
	SeverityInfo Severity = iota

	// SeverityWarn emits a warning diagnostic.
	SeverityWarn

	// SeverityErr emits an error diagnostic.
	SeverityErr
)

// Report is a compiled diagnostic report clause.
type Report struct {
	// Severity identifies the diagnostic severity to emit.
	Severity Severity

	// TargetToken is the source-order token index whose span receives the diagnostic.
	TargetToken int

	// Message is the emitted diagnostic message.
	Message string
}
