// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package validator validates parsed Spot DSL syntax.
package validator

import "github.com/kdeconinck/spot/location"

// Diagnostic describes a semantic problem found while validating DSL syntax.
type Diagnostic struct {
	// Message explains the validation problem.
	Message string

	// Span is the byte range where the validation problem was detected.
	Span location.Span
}
