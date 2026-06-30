// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package parser parses Spot DSL source text into syntax data structures.
package parser

import (
	"strconv"

	"github.com/kdeconinck/spot/location"
)

// Diagnostic describes a syntax problem found while parsing DSL source text.
//
// Diagnostic implements error so the parser can fail fast and return the first syntax problem directly.
type Diagnostic struct {
	// Message explains the syntax problem.
	Message string

	// Span is the byte range where the syntax problem was detected.
	Span location.Span
}

// Error returns the diagnostic message with its source span.
func (diagnostic Diagnostic) Error() string {
	return diagnostic.Message +
		". [" +
		strconv.Itoa(int(diagnostic.Span.Start)) +
		":" +
		strconv.Itoa(int(diagnostic.Span.End)) +
		"]"
}
