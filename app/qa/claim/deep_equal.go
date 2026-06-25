// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

package claim

import (
	"reflect"
	"strings"
)

// DeepEqual reports a test failure if got is not deeply equal to want.
// Both testName and label parameters are used to construct a descriptive failure message.
// DeepEqual itself is marked as a helper, so failures are reported at the caller site.
func DeepEqual[V any](tb TB, testName string, want, got V, label string) {
	tb.Helper()

	if !reflect.DeepEqual(got, want) {
		var sb strings.Builder

		fmtWant, fmtGot := fmtValue(want), fmtValue(got)

		writeFailureMessage(&sb, testName, label, fmtWant, fmtGot)

		tb.Fatalf("%s", sb.String())
	}
}
