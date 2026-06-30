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

	// Expressions stores compiled token and definition expressions in flat slices.
	Expressions ExpressionArena

	// Rules are the compiled rule definitions in source order.
	Rules []Rule
}

// Token is a compiled token definition.
type Token struct {
	// Name is the emitted token name.
	Name string

	// Expression is the root compiled token expression.
	Expression ExpressionID

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

	// ExpressionReference is a reference to a compiled definition expression.
	ExpressionReference

	// ExpressionConcatenation is a sequence of expressions.
	ExpressionConcatenation

	// ExpressionAlternation is a list of alternative expressions.
	ExpressionAlternation

	// ExpressionGroup is a parenthesized expression.
	ExpressionGroup

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

// ExpressionID identifies a node in an ExpressionArena.
type ExpressionID uint32

// ExpressionArena stores compiled token expressions in flat slices.
//
// Nodes contains the actual expression records. ChildIDs stores the adjacency data for nodes that have children,
// such as alternations, concatenations, and repetitions. A node's FirstElementIdx and AmountOfElements describe which
// segment of ChildIDs belongs to that node.
type ExpressionArena struct {
	// Nodes contains every compiled expression node referenced by the program.
	Nodes []ExpressionNode

	// ChildIDs stores child node identifiers for branch nodes.
	ChildIDs []ExpressionID

	// Strings stores exact string literals used by compiled string expressions.
	Strings []string
}

// Node returns the expression node identified by id.
func (arena ExpressionArena) Node(id ExpressionID) ExpressionNode {
	return arena.Nodes[id]
}

// Children returns the child expression identifiers for node.
func (arena ExpressionArena) Children(node ExpressionNode) []ExpressionID {
	return arena.ChildIDs[node.FirstElementIdx : node.FirstElementIdx+node.AmountOfElements]
}

// String returns the string literal identified by id.
func (arena ExpressionArena) String(id uint32) string {
	return arena.Strings[id]
}

// ExpressionNode is one compiled token expression node.
type ExpressionNode struct {
	// Kind identifies the form of expression.
	Kind ExpressionKind

	// Character is the matched byte in a character expression.
	Character byte

	// RangeStart is the inclusive start of a range expression.
	RangeStart byte

	// RangeEnd is the inclusive end of a range expression.
	RangeEnd byte

	// Reference identifies the target definition root in a reference expression.
	Reference ExpressionID

	// StringID identifies the exact string literal in a string expression.
	StringID uint32

	// FirstElementIdx is the start offset of this node's children in ChildIDs.
	FirstElementIdx uint32

	// AmountOfElements is the number of children stored for this node.
	AmountOfElements uint32

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
