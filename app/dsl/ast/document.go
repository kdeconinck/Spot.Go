// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package ast defines data structures that represent the AST (Abstract Syntax Tree) for Spot DSL syntax.
package ast

import "github.com/kdeconinck/spot/location"

// Document is the root syntax node for a Spot DSL file.
//
// Document stores parsed syntax in flat slices rather than as a pointer-linked tree of nested structs.
// Section fields such as Scope and Definitions describe where a section lives, while the corresponding slice fields
// store the actual section contents in source order.
//
// This layout is chosen for parser and traversal performance:
//   - Parsing allocates a small number of large slices instead of many tiny heap objects.
//   - Syntax data stays densely packed, which improves cache locality during validation and compilation.
//   - The parser can size every slice exactly before materialization, which avoids growth reallocations.
//
// This layout is intentionally split in two:
//   - Section descriptors store spans plus index/count pairs.
//   - Shared flat slices store the concrete entries and declarations.
//
// Expressions follow the same pattern. Definitions and token declarations store a DefinitionExpressionID, and the
// actual expression graph lives in Expressions. Nodes reference children by index through
// DefinitionExpressionArena.ChildIDs instead of through pointers.
//
// Example:
//
//	scope {
//	    include "**/*.go"
//	    exclude "vendor/**"
//	}
//
// is stored as:
//
//	Document.Scope:
//	    FirstElementIdx = 0
//	    AmountOfElements = 2
//
//	Document.ScopeEntries:
//	    [0] include "**/*.go"
//	    [1] exclude "vendor/**"
//
// So the scope section reads its entries from ScopeEntries[0:2].
//
// Expression nodes work the same way. The definition:
//
//	value = ('a' | 'b')+
//
// is stored as:
//
//	DefinitionList[n].Expression = 5
//
//	Expressions.Nodes:
//	    [1] Character('a')
//	    [2] Character('b')
//	    [3] Alternation  FirstChildIndex=0 ChildCount=2
//	    [4] Group        FirstChildIndex=2 ChildCount=1
//	    [5] Repetition+  FirstChildIndex=3 ChildCount=1
//
//	Expressions.ChildIDs:
//	    [0] 1
//	    [1] 2
//	    [2] 3
//	    [3] 4
//
// So node 3 reads ChildIDs[0:2] and gets children [1, 2], node 4 reads ChildIDs[2:3] and gets child [3], and node
// 5 reads ChildIDs[3:4] and gets child [4].
//
// The parser builds this layout in two passes:
//   - A sizing pass counts exact capacities for every flat slice.
//   - The real parse fills those slices without growth reallocations.
type Document struct {
	// Scope is the parsed scope section.
	Scope ScopeSection

	// ScopeEntries stores every parsed scope entry in source order.
	// Scope.FirstEntry and Scope.EntryCount describe which range belongs to the scope section.
	ScopeEntries []ScopeEntry

	// Definitions is the parsed definitions section.
	Definitions DefinitionsSection

	// DefinitionList stores every parsed definition in source order.
	// Definitions.FirstDefinition and Definitions.DefinitionCount describe which range belongs to the definitions section.
	DefinitionList []Definition

	// Tokens is the parsed tokens section.
	Tokens TokensSection

	// TokenList stores every parsed token definition in source order.
	// Tokens.FirstToken and Tokens.TokenCount describe which range belongs to the tokens section.
	TokenList []TokenDefinition

	// Syntax is the parsed syntax section.
	Syntax SyntaxSection

	// SyntaxNodeList stores every parsed syntax node declaration in source order.
	// Syntax.FirstElementIdx and Syntax.AmountOfElements describe which range belongs to the syntax section.
	SyntaxNodeList []SyntaxNode

	// SyntaxExpressions stores every parsed syntax node expression.
	// Syntax nodes refer into this arena by SyntaxExpressionID.
	SyntaxExpressions SyntaxExpressionArena

	// Rules is the parsed rules section.
	Rules RulesSection

	// RuleList stores every parsed rule in source order.
	// Rules.FirstRule and Rules.RuleCount describe which range belongs to the rules section.
	RuleList []Rule

	// Expressions stores every parsed definition and token expression node.
	// Definitions and tokens refer into this arena by DefinitionExpressionID.
	Expressions DefinitionExpressionArena

	// Span is the byte range covered by the document.
	Span location.Span
}

// ScopeSectionEntries returns the scope entries declared in section.
func (document Document) ScopeSectionEntries(section ScopeSection) []ScopeEntry {
	return document.ScopeEntries[section.FirstElementIdx : section.FirstElementIdx+section.AmountOfElements]
}

// SectionDefinitions returns the definitions declared in section.
func (document Document) SectionDefinitions(section DefinitionsSection) []Definition {
	return document.DefinitionList[section.FirstElementsIdx : section.FirstElementsIdx+section.AmountOfElements]
}

// SectionTokens returns the token definitions declared in section.
func (document Document) SectionTokens(section TokensSection) []TokenDefinition {
	return document.TokenList[section.FirstElementIdx : section.FirstElementIdx+section.AmountOfElements]
}

// SectionSyntaxNodes returns the syntax node declarations declared in section.
func (document Document) SectionSyntaxNodes(section SyntaxSection) []SyntaxNode {
	return document.SyntaxNodeList[section.FirstElementIdx : section.FirstElementIdx+section.AmountOfElements]
}

// SectionRules returns the rules declared in section.
func (document Document) SectionRules(section RulesSection) []Rule {
	return document.RuleList[section.FirstElementIdx : section.FirstElementIdx+section.AmountOfElements]
}
