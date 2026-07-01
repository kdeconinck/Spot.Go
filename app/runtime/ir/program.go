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

	// SyntaxNodes are the compiled syntax node definitions in source order.
	SyntaxNodes []SyntaxNode

	// SyntaxFields are the named child captures referenced by compiled syntax trees and rule paths.
	SyntaxFields []string

	// SyntaxExpressions stores compiled syntax node expressions in flat slices.
	SyntaxExpressions SyntaxExpressionArena

	// SyntaxRoot is the source-order syntax node index that represents the single file-level syntax root.
	// It is -1 when the compiled program does not define a unique root.
	SyntaxRoot int

	// Rules are the compiled rule definitions in source order.
	Rules []Rule
}

// Token is a compiled token definition.
type Token struct {
	// Name is the emitted token name.
	Name string

	// Expression is the root compiled token expression.
	Expression ExpressionID

	// Fallback reports whether this token consumes one otherwise-unmatched byte.
	Fallback bool

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

// SyntaxNode is a compiled syntax node definition.
type SyntaxNode struct {
	// Name is the declared syntax node name.
	Name string

	// Expression is the root compiled syntax expression.
	Expression SyntaxExpressionID
}

// SyntaxExpressionKind identifies the form of a compiled syntax expression.
type SyntaxExpressionKind uint8

const (
	// SyntaxExpressionReference is a token or syntax node reference.
	SyntaxExpressionReference SyntaxExpressionKind = iota

	// SyntaxExpressionAny matches one emitted token of any kind.
	SyntaxExpressionAny

	// SyntaxExpressionCapture labels direct child syntax nodes produced by the wrapped expression.
	SyntaxExpressionCapture

	// SyntaxExpressionConcatenation is a sequence of syntax expressions.
	SyntaxExpressionConcatenation

	// SyntaxExpressionAlternation is a list of alternative syntax expressions.
	SyntaxExpressionAlternation

	// SyntaxExpressionGroup is a parenthesized syntax expression.
	SyntaxExpressionGroup

	// SyntaxExpressionRepetition is a repeated syntax expression.
	SyntaxExpressionRepetition
)

// SyntaxReferenceKind identifies the target kind of a syntax reference.
type SyntaxReferenceKind uint8

const (
	// SyntaxReferenceToken targets a compiled token definition.
	SyntaxReferenceToken SyntaxReferenceKind = iota

	// SyntaxReferenceNode targets a compiled syntax node definition.
	SyntaxReferenceNode
)

// SyntaxExpressionID identifies a node in a SyntaxExpressionArena.
type SyntaxExpressionID uint32

// SyntaxExpressionArena stores compiled syntax expressions in flat slices.
//
// Nodes contains the actual syntax expression records. ChildIDs stores the adjacency data for nodes that have
// children, such as alternations, concatenations, groups, and repetitions. A node's FirstElementIdx and
// AmountOfElements describe which segment of ChildIDs belongs to that node.
type SyntaxExpressionArena struct {
	// Nodes contains every compiled syntax expression node referenced by the program.
	Nodes []SyntaxExpressionNode

	// ChildIDs stores child node identifiers for branch nodes.
	ChildIDs []SyntaxExpressionID
}

// Node returns the syntax expression node identified by id.
func (arena SyntaxExpressionArena) Node(id SyntaxExpressionID) SyntaxExpressionNode {
	return arena.Nodes[id]
}

// Children returns the child syntax expression identifiers for node.
func (arena SyntaxExpressionArena) Children(node SyntaxExpressionNode) []SyntaxExpressionID {
	return arena.ChildIDs[node.FirstElementIdx : node.FirstElementIdx+node.AmountOfElements]
}

// SyntaxExpressionNode is one compiled syntax expression node.
type SyntaxExpressionNode struct {
	// Kind identifies the form of expression.
	Kind SyntaxExpressionKind

	// ReferenceKind identifies whether Reference targets a token or syntax node.
	ReferenceKind SyntaxReferenceKind

	// Reference identifies the target token or syntax node in a reference expression.
	Reference uint32

	// FieldID identifies the named child capture in a capture expression.
	FieldID uint32

	// FirstElementIdx is the start offset of this node's children in ChildIDs.
	FirstElementIdx uint32

	// AmountOfElements is the number of children stored for this node.
	AmountOfElements uint32

	// Repetition identifies the repetition operator.
	Repetition RepetitionKind
}

// RuleMatchKind identifies what a compiled rule matches.
type RuleMatchKind uint8

const (
	// RuleMatchToken matches one emitted token.
	RuleMatchToken RuleMatchKind = iota

	// RuleMatchSyntaxNode matches one runtime syntax node.
	RuleMatchSyntaxNode
)

// RuleMatchScopeKind identifies whether a syntax-node rule constrains ancestor syntax nodes.
type RuleMatchScopeKind uint8

const (
	// RuleMatchScopeNone does not constrain ancestor syntax nodes.
	RuleMatchScopeNone RuleMatchScopeKind = iota

	// RuleMatchScopeParent requires the matched syntax node to have the named direct parent syntax node.
	RuleMatchScopeParent

	// RuleMatchScopeInside requires the matched syntax node to be inside the named ancestor syntax node.
	RuleMatchScopeInside

	// RuleMatchScopeParentOutside requires the matched syntax node to not have the named direct parent syntax node.
	RuleMatchScopeParentOutside

	// RuleMatchScopeOutside requires the matched syntax node to be outside the named ancestor syntax node.
	RuleMatchScopeOutside
)

// RuleMatchRelationKind identifies whether a compiled rule matches one syntax node or a relation between two syntax nodes.
type RuleMatchRelationKind uint8

const (
	// RuleMatchRelationNone matches a single token or syntax node.
	RuleMatchRelationNone RuleMatchRelationKind = iota

	// RuleMatchRelationAdjacentSibling matches two adjacent sibling syntax nodes.
	RuleMatchRelationAdjacentSibling
)

// Rule is a compiled diagnostic rule.
type Rule struct {
	// Name is the DSL rule name.
	Name string

	// MatchKind identifies whether the rule matches a token or a syntax node.
	MatchKind RuleMatchKind

	// MatchIndex is the source-order token or syntax node index matched by the rule.
	MatchIndex int

	// RelationKind identifies whether the rule matches one node or a syntax-node relation.
	RelationKind RuleMatchRelationKind

	// RelatedMatchIndex is the source-order syntax node index for the related syntax node in a relational match.
	RelatedMatchIndex int

	// MatchScopeKind identifies whether the syntax-node match constrains ancestor syntax nodes.
	MatchScopeKind RuleMatchScopeKind

	// MatchScopeIndex is the source-order syntax node index used by the inside/outside constraint.
	MatchScopeIndex int

	// Where is the optional compiled condition.
	Where Condition

	// Report is the compiled diagnostic report.
	Report Report
}

// ConditionSubjectKind identifies which runtime value a condition reads.
type ConditionSubjectKind uint8

const (
	// ConditionSubjectNone marks a rule without a where clause.
	ConditionSubjectNone ConditionSubjectKind = iota

	// ConditionSubjectMatch reads the primary matched token or syntax node.
	ConditionSubjectMatch

	// ConditionSubjectRelatedMatch reads the related syntax node in a relational selector.
	ConditionSubjectRelatedMatch

	// ConditionSubjectGap reads the source gap between two adjacent matched syntax nodes.
	ConditionSubjectGap
)

// ConditionProperty identifies the property read from a condition subject.
type ConditionProperty uint8

const (
	// ConditionPropertyNone marks a rule without a where clause.
	ConditionPropertyNone ConditionProperty = iota

	// ConditionPropertyText reads the matched text.
	ConditionPropertyText

	// ConditionPropertyLength reads the matched text length.
	ConditionPropertyLength

	// ConditionPropertyBlankLines reads the number of blank lines in the source gap.
	ConditionPropertyBlankLines
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

	// ConditionOperatorStartsWith compares whether the left string starts with the right string.
	ConditionOperatorStartsWith
)

// Condition is a compiled rule where clause.
type Condition struct {
	// LeftSubject identifies the runtime value read on the left-hand side.
	LeftSubject ConditionSubjectKind

	// LeftPath is the named syntax-field path traversed from LeftSubject.
	LeftPath []uint32

	// Property identifies the matched token property read by the condition.
	LeftProperty ConditionProperty

	// Operator identifies the comparison performed by the condition.
	Operator ConditionOperator

	// RightSubject identifies the optional runtime value read on the right-hand side.
	RightSubject ConditionSubjectKind

	// RightPath is the named syntax-field path traversed from RightSubject.
	RightPath []uint32

	// RightProperty identifies the optional right-hand property.
	RightProperty ConditionProperty

	// String is the right-hand string value when RightSubject is ConditionSubjectNone and the right-hand side is textual.
	String string

	// Integer is the right-hand integer value when RightSubject is ConditionSubjectNone and the right-hand side is numeric.
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

	// TargetKind identifies whether the report span comes from a token or a syntax node.
	TargetKind RuleMatchKind

	// TargetIndex is the source-order token or syntax node index whose span receives the diagnostic.
	TargetIndex int

	// Message is the emitted diagnostic message.
	Message string
}
