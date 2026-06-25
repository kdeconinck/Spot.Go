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

	// Expression is the character literal token assigned to the definition.
	Expression Token

	// Span is the byte range covered by the definition.
	Span location.Span
}
