// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import "github.com/kdeconinck/spot/location"

// Diagnostic describes a syntax problem found while parsing DSL source text.
type Diagnostic struct {
	// Message explains the syntax problem.
	Message string

	// Span is the byte range where the syntax problem was detected.
	Span location.Span
}
