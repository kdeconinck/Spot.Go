// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package resolver builds reusable declaration lookups over parsed Spot DSL syntax.
package resolver

import "github.com/kdeconinck/spot/dsl/ast"

// Resolution stores parser output together with declaration lookups derived from source order.
//
// Resolution does not validate semantics. It only indexes already-parsed syntax so later stages can reuse the same
// lookups instead of rebuilding temporary maps.
type Resolution struct {
	// Document is the parsed syntax tree being resolved.
	Document ast.Document

	// ScopeEntries stores the parsed scope entries in section order.
	ScopeEntries []ast.ScopeEntry

	// Definitions stores the parsed definitions in section order.
	Definitions []ast.Definition

	// Tokens stores the parsed token definitions in section order.
	Tokens []ast.TokenDefinition

	// SyntaxNodes stores the parsed syntax node declarations in section order.
	SyntaxNodes []ast.SyntaxNode

	// Rules stores the parsed rules in section order.
	Rules []ast.Rule

	// DefinitionIndexes maps each definition name to its first declaration index.
	DefinitionIndexes map[string]int

	// TokenIndexes maps each token name to its first declaration index.
	TokenIndexes map[string]int

	// SyntaxNodeIndexes maps each syntax node name to its first declaration index.
	SyntaxNodeIndexes map[string]int

	// RuleIndexes maps each rule name to its first declaration index.
	RuleIndexes map[string]int
}

// Resolve builds reusable declaration lookups over document.
func Resolve(source string, document ast.Document) Resolution {
	definitions := document.SectionDefinitions(document.Definitions)
	tokens := document.SectionTokens(document.Tokens)
	syntaxNodes := document.SectionSyntaxNodes(document.Syntax)
	rules := document.SectionRules(document.Rules)
	resolution := Resolution{
		Document:          document,
		ScopeEntries:      document.ScopeSectionEntries(document.Scope),
		Definitions:       definitions,
		Tokens:            tokens,
		SyntaxNodes:       syntaxNodes,
		Rules:             rules,
		DefinitionIndexes: make(map[string]int, len(definitions)),
		TokenIndexes:      make(map[string]int, len(tokens)),
		SyntaxNodeIndexes: make(map[string]int, len(syntaxNodes)),
		RuleIndexes:       make(map[string]int, len(rules)),
	}

	for idx := range definitions {
		name := definitions[idx].Name.Value(source)

		if _, ok := resolution.DefinitionIndexes[name]; !ok {
			resolution.DefinitionIndexes[name] = idx
		}
	}

	for idx := range tokens {
		name := tokens[idx].Name.Value(source)

		if _, ok := resolution.TokenIndexes[name]; !ok {
			resolution.TokenIndexes[name] = idx
		}
	}

	for idx := range syntaxNodes {
		name := syntaxNodes[idx].Name.Value(source)

		if _, ok := resolution.SyntaxNodeIndexes[name]; !ok {
			resolution.SyntaxNodeIndexes[name] = idx
		}
	}

	for idx := range rules {
		name := rules[idx].Name.Value(source)

		if _, ok := resolution.RuleIndexes[name]; !ok {
			resolution.RuleIndexes[name] = idx
		}
	}

	return resolution
}

// DefinitionIndex returns the first declaration index for name.
func (resolution Resolution) DefinitionIndex(name string) (int, bool) {
	idx, ok := resolution.DefinitionIndexes[name]

	return idx, ok
}

// TokenIndex returns the first declaration index for name.
func (resolution Resolution) TokenIndex(name string) (int, bool) {
	idx, ok := resolution.TokenIndexes[name]

	return idx, ok
}

// SyntaxNodeIndex returns the first declaration index for name.
func (resolution Resolution) SyntaxNodeIndex(name string) (int, bool) {
	idx, ok := resolution.SyntaxNodeIndexes[name]

	return idx, ok
}

// RuleIndex returns the first declaration index for name.
func (resolution Resolution) RuleIndex(name string) (int, bool) {
	idx, ok := resolution.RuleIndexes[name]

	return idx, ok
}
