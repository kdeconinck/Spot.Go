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

	// Span is the byte range covered by the document.
	Span location.Span
}

// ScopeSection is a parsed scope section.
type ScopeSection struct {
	// Span is the byte range covered by the scope section.
	Span location.Span
}
