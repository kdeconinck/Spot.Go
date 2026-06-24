// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package claim provides simple test helpers for writing concise assertions.
// It defines minimal assertion functions that integrate with testing.TB and report failures with descriptive messages.
package claim

// TB is the minimal interface required by this package for reporting test failures.
// It matches the subset of methods from testing.TB used by this package and is satisfied by types such as *testing.T
// and *testing.B.
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
}
