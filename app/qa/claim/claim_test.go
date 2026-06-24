// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the claim package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package claim_test

import (
	"fmt"
	"testing"
)

// A strict test double for claim.TB.
// It implements only the methods required by the public API, which keeps it small and makes test intent explicit.
// Unlike a real testing object, it records calls instead of terminating the test, allowing inspections by the caller.
type tbSpy struct {
	failureMsg          string
	amountOfHelperCalls int
	amountOfFatalfCalls int
}

// Fatalf records the formatted failure message and increments the count of Fatalf calls.
func (spy *tbSpy) Fatalf(format string, args ...any) {
	spy.amountOfFatalfCalls++
	spy.failureMsg = fmt.Sprintf(format, args...)
}

// Helper the count of Helper calls.
func (spy *tbSpy) Helper() {
	spy.amountOfHelperCalls++
}

func (spy *tbSpy) verifyFailure(t *testing.T, wantMsg string, wantHelperCalls, wantFatalCalls int) {
	if spy.failureMsg != wantMsg {
		t.Fatalf("Failure message = %q, want %q.", spy.failureMsg, wantMsg)
	}

	if spy.amountOfHelperCalls != wantHelperCalls {
		t.Fatalf("Helper calls = %d, want %d.", spy.amountOfHelperCalls, wantHelperCalls)
	}

	if spy.amountOfFatalfCalls != wantFatalCalls {
		t.Fatalf("Fatalf calls = %d, want %d.", spy.amountOfFatalfCalls, wantFatalCalls)
	}
}

// A no-op implementation of claim.TB.
// It implements the same methods as tbSpy but does nothing, making it useful for tests that don't require inspection
// of TB interactions, such as benchmarks.
type tbNoop struct {
	// NOTE: Intentionally left blank.
}

func (tb *tbNoop) Helper()                   { /* NOTE: Intentionally left blank. */ }
func (tb *tbNoop) Fatalf(_ string, _ ...any) { /* NOTE: Intentionally left blank. */ }
