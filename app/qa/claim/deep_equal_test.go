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
	"testing"

	"github.com/kdeconinck/spot/qa/claim"
)

func Test_DeepEqual(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		wantInput, gotInput              []int
		wantMsg                          string
		wantHelperCalls, wantFatalfCalls int
	}{
		"When the compared values are deeply equal, no failure is reported.": {
			wantInput: []int{1, 2}, gotInput: []int{1, 2},
			wantMsg:         "",
			wantHelperCalls: 1, wantFatalfCalls: 0,
		},
		"When the compared values are not deeply equal, a failure is reported.": {
			wantInput: []int{1, 2}, gotInput: []int{1, 3},
			wantHelperCalls: 1, wantFatalfCalls: 1,
			wantMsg: "\n\nTest name:          Deep equality.\n\033[32mExpected (Numbers): [1 2]\033[0m\n\033[31mActual (Numbers):   [1 3]\033[0m\n\n",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			spy := new(tbSpy)

			// Act.
			claim.DeepEqual(spy, "Deep equality.", tc.wantInput, tc.gotInput, "Numbers")

			// Assert.
			spy.verifyFailure(t, tc.wantMsg, tc.wantHelperCalls, tc.wantFatalfCalls)
		})
	}
}
