// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package claim provides simple test helpers for writing concise assertions.
// It defines minimal assertion functions that integrate with testing.TB and report failures with descriptive messages.
package claim

import "strings"

// Equal reports a test failure if got is not equal to want.
// Both testName and label parameters are used to construct a descriptive failure message.
// Equal itself is marked as a helper, so failures are reported at the caller site.
func Equal[V comparable](tb TB, testName string, want, got V, label string) {
	tb.Helper()

	if got != want {
		var sb strings.Builder

		fmtWant, fmtGot := fmtValue(want), fmtValue(got)

		writeFailureMessage(&sb, testName, label, fmtWant, fmtGot)

		tb.Fatalf("%s", sb.String())
	}
}
