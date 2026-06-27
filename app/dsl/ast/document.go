// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package ast defines data structures that represent the AST (Abstract Syntax Tree) for Spot DSL syntax.
package ast

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
