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
	Pattern token.Token

	// Span is the byte range covered by the scope entry.
	Span location.Span
}
