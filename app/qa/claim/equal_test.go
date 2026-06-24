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
	"errors"
	"testing"

	"github.com/kdeconinck/spot/qa/claim"
)

type typedNilError struct {
	msg string
}

func (err *typedNilError) Error() string {
	return err.msg
}

type valueError struct {
	msg string
}

func (err valueError) Error() string {
	return err.msg
}

func Test_Equal(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		testName                         string
		label                            string
		wantInput, gotInput              bool
		wantMsg                          string
		wantHelperCalls, wantFatalfCalls int
	}{
		"When the compared values are equal, no failure is reported.": {
			testName:  "Bool equality.",
			label:     "OK",
			wantInput: true, gotInput: true,
			wantMsg:         "",
			wantHelperCalls: 1, wantFatalfCalls: 0,
		},
		"When the compared values are not equal, a failure is reported.": {
			testName:  "Bool equality.",
			label:     "OK",
			wantInput: false, gotInput: true,
			wantHelperCalls: 1, wantFatalfCalls: 1,
			wantMsg: "\n\nTest name:     Bool equality.\n\033[32mExpected (OK): false\033[0m\n\033[31mActual (OK):   true\033[0m\n\n",
		},
		"When the label is empty, the field names do not include a suffix.": {
			testName:  "Bool equality.",
			label:     "",
			wantInput: false, gotInput: true,
			wantHelperCalls: 1, wantFatalfCalls: 1,
			wantMsg: "\n\nTest name: Bool equality.\n\033[32mExpected:  false\033[0m\n\033[31mActual:    true\033[0m\n\n",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			spy := new(tbSpy)

			// Act.
			claim.Equal(spy, tc.testName, tc.wantInput, tc.gotInput, tc.label)

			// Assert.
			spy.verifyFailure(t, tc.wantMsg, tc.wantHelperCalls, tc.wantFatalfCalls)
		})
	}
}

func Test_Equal_RenderValue(t *testing.T) {
	t.Parallel()

	type customComparable struct {
		N int
	}

	for tcName, tc := range map[string]struct {
		fn   func(spy *tbSpy)
		want string
	}{
		"When rendering a string, the value is used as-is.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", "hello", "world", "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): \"hello\"\033[0m\n\033[31mActual (Label):   \"world\"\033[0m\n\n",
		},
		"When rendering a bool, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", false, true, "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): false\033[0m\n\033[31mActual (Label):   true\033[0m\n\n",
		},
		"When rendering an int, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", 1, 2, "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering an int8, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", int8(1), int8(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering an int16, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", int16(1), int16(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering an int32, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", int32(1), int32(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering an int64, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", int64(1), int64(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering a uint, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", uint(1), uint(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering a uint8, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", uint8(1), uint8(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering a uint16, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", uint16(1), uint16(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering a uint32, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", uint32(1), uint32(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering a uint64, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", uint64(1), uint64(2), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1\033[0m\n\033[31mActual (Label):   2\033[0m\n\n",
		},
		"When rendering a float32, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", float32(1.5), float32(2.5), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1.5\033[0m\n\033[31mActual (Label):   2.5\033[0m\n\n",
		},
		"When rendering a float64, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", 1.5, 2.5, "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): 1.5\033[0m\n\033[31mActual (Label):   2.5\033[0m\n\n",
		},
		"When rendering an error, the error message is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", errors.New("want error"), errors.New("got error"), "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): want error\033[0m\n\033[31mActual (Label):   got error\033[0m\n\n",
		},
		"When rendering a typed-nil error, the value is formatted without panicking.": {
			fn: func(spy *tbSpy) {
				var want error = (*typedNilError)(nil)
				got := errors.New("got error")

				claim.Equal(spy, "Test Name", want, got, "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): *claim_test.typedNilError(nil)\033[0m\n\033[31mActual (Label):   got error\033[0m\n\n",
		},
		"When rendering a value error, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				claim.Equal(spy, "Test Name", valueError{msg: "want error"}, valueError{msg: "got error"}, "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): want error\033[0m\n\033[31mActual (Label):   got error\033[0m\n\n",
		},
		"When rendering an unknown type, the value is formatted correctly.": {
			fn: func(spy *tbSpy) {
				want := customComparable{
					N: 1,
				}

				got := customComparable{
					N: 2,
				}

				claim.Equal(spy, "Test Name", want, got, "Label")
			},
			want: "\n\nTest name:        Test Name\n\033[32mExpected (Label): {1}\033[0m\n\033[31mActual (Label):   {2}\033[0m\n\n",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			spy := new(tbSpy)

			// Act.
			tc.fn(spy)

			// Assert.
			if spy.failureMsg != tc.want {
				t.Fatalf("Failure message = %q, want %q.", spy.failureMsg, tc.want)
			}
		})
	}
}

func benchmark_EqualFailure(b *testing.B) {
	b.Helper()

	tb := new(tbNoop)

	for b.Loop() {
		claim.Equal(tb, "Benchmark equality.", false, true, "Label")
	}
}

func benchmark_EqualSuccess(b *testing.B) {
	b.Helper()

	tb := new(tbNoop)

	for b.Loop() {
		claim.Equal(tb, "Benchmark equality.", true, true, "Label")
	}
}

func Benchmark_Equal_Failure(b *testing.B) { benchmark_EqualFailure(b) }
func Benchmark_Equal_Success(b *testing.B) { benchmark_EqualSuccess(b) }
