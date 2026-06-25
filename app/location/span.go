// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package location defines source-location primitives shared by Spot pipeline stages.
package location

// Position represents a zero-based byte offset in source text.
type Position int

// Span represents a half-open byte range in source text.
type Span struct {
	// Start is the zero-based byte offset where the span begins.
	Start Position

	// End is the zero-based byte offset immediately after the span.
	End Position
}
