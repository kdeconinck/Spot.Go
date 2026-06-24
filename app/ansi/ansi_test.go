// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the ansi package.
//
// Tests in this package are written against the exported API only.
// This ensures that validation behavior is tested through the same surface that external consumers would use.
package ansi_test

import (
	"strings"
	"testing"

	"github.com/kdeconinck/spot/ansi"
	"github.com/kdeconinck/spot/qa/claim"
)

// Labels, used in the different assertion methods.
const (
	CodeLabel          = "ANSI Code"
	LenLabel           = "Len"
	AnsiEscapeSequence = "Value"
)

func Test_Code_String(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inCode ansi.Code
		want   string
	}{
		"When using the 'reset' ANSI code, the returned value is correct.": {
			inCode: ansi.Reset,
			want:   "0",
		},
		"When using the 'red' ANSI code, the returned value is correct.": {
			inCode: ansi.Red,
			want:   "31",
		},
		"When using the 'green' ANSI code, the returned value is correct.": {
			inCode: ansi.Green,
			want:   "32",
		},
		"When using an unknown ANSI code, the returned value is correct.": {
			inCode: ansi.Code(128),
			want:   "128",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := tc.inCode.String()

			// Assert.
			claim.Equal(t, tcName, tc.want, got, CodeLabel)
		})
	}
}

func Test_OpenLen(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inCodes []ansi.Code
		want    int
	}{
		// NOTE: When using NO ANSI code(s), Len returns the length of value.
		"When using NO ANSI code(s), the returned value is correct.": {
			want: 0,
		},
		// NOTE: When using a single ANSI code, Len returns:
		// - Length of the 'ANSI Escape prefix'.
		// - Length of the value of the 'ANSI Escape code'.
		// - Length of the character 'm'.
		"When using a single ANSI code(s), the returned value is correct.": {
			inCodes: []ansi.Code{
				ansi.Green,
			},
			want: len("\x1b[") + len(ansi.Green.String()) + 1,
		},
		"When using the 'reset' ANSI code, the returned value is correct.": {
			inCodes: []ansi.Code{
				ansi.Reset,
			},
			want: len("\x1b[") + len(ansi.Reset.String()) + 1,
		},
		// NOTE: When using multiple ANSI codes, Len returns:
		// - Length of the 'ANSI Escape prefix'.
		// - Length of the value of the 'ANSI Escape code'.
		// - For each additional code:
		//   (*) Length of the character ';'.
		//   (*) Length of the value of the 'ANSI Escape code'.
		// - Length of the character 'm'.
		"When using a multiple ANSI code(s), the returned value is correct.": {
			inCodes: []ansi.Code{
				ansi.Green, ansi.Red,
			},
			want: len("\x1b[") + len(ansi.Green.String()) + 1 + len(ansi.Red.String()) + 1,
		},
		"When using unknown ANSI codes, the returned value is correct.": {
			inCodes: []ansi.Code{
				ansi.Code(9),
				ansi.Code(42),
				ansi.Code(128),
			},
			want: len("\x1b[") + len("9") + 1 + len("42") + 1 + len("128") + 1,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := ansi.OpenLen(tc.inCodes)

			// Assert.
			claim.Equal(t, tcName, tc.want, got, LenLabel)
		})
	}
}

func Test_WriteOpenTo(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inCodes []ansi.Code
		want    string
	}{
		"When using NO ANSI code(s), nothing is appended.": {
			want: "",
		},
		"When using an empty ANSI code slice, nothing is appended.": {
			inCodes: []ansi.Code{},
			want:    "",
		},
		"When using a single ANSI code, the correct value is appended": {
			inCodes: []ansi.Code{
				ansi.Green,
			},
			want: "\x1b[32m",
		},
		"When using the 'reset' ANSI code, the correct value is appended.": {
			inCodes: []ansi.Code{
				ansi.Reset,
			},
			want: "\x1b[0m",
		},
		"When using multiple ANSI codes, the correct value is appended.": {
			inCodes: []ansi.Code{
				ansi.Green,
				ansi.Red,
			},
			want: "\x1b[32;31m",
		},
		"When using unknown ANSI codes, the correct value is appended.": {
			inCodes: []ansi.Code{
				ansi.Code(9),
				ansi.Code(42),
				ansi.Code(128),
			},
			want: "\x1b[9;42;128m",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			var sb strings.Builder

			// Act.
			seq := ansi.NewSequence(tc.inCodes)
			seq.WriteOpenTo(&sb)

			// Assert.
			claim.Equal(t, tcName, tc.want, sb.String(), AnsiEscapeSequence)
		})
	}
}

func Test_Sequence_WriteOpenTo_Appends(t *testing.T) {
	t.Parallel()

	// Arrange.
	var sb strings.Builder
	sb.WriteString("prefix:")

	seq := ansi.NewSequence([]ansi.Code{
		ansi.Green,
	})

	// Act.
	seq.WriteOpenTo(&sb)

	// Assert.
	claim.Equal(t, "When appending a opening sequence to a non-empty builder, the correct value is appended.", "prefix:\x1b[32m", sb.String(), AnsiEscapeSequence)
}

func Test_WriteResetTo(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inCodes []ansi.Code
		want    string
	}{
		"When appending the ANSI reset sequence for a non-empty sequence, the correct value is appended.": {
			inCodes: []ansi.Code{
				ansi.Green,
			},
			want: "\x1b[0m",
		},
		"When appending the ANSI reset sequence for an empty sequence, nothing is appended.": {
			want: "",
		},
		"When appending the ANSI reset sequence for an empty ANSI code slice, nothing is appended.": {
			inCodes: []ansi.Code{},
			want:    "",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			var sb strings.Builder
			seq := ansi.NewSequence(tc.inCodes)

			// Act.
			seq.WriteResetTo(&sb)

			// Assert.
			claim.Equal(t, tcName, tc.want, sb.String(), AnsiEscapeSequence)
		})
	}
}

func Test_Sequence_WriteResetTo_Appends(t *testing.T) {
	t.Parallel()

	// Arrange.
	var sb strings.Builder
	sb.WriteString("prefix:")

	seq := ansi.NewSequence([]ansi.Code{
		ansi.Green,
	})

	// Act.
	seq.WriteResetTo(&sb)

	// Assert.
	claim.Equal(t, "When appending a reset sequence to a non-empty builder, the correct value is appended.", "prefix:\x1b[0m", sb.String(), AnsiEscapeSequence)
}

func Test_Sequence_OpenLen(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inCodes []ansi.Code
		want    int
	}{
		"When a sequence contains no codes, the opening length is correct.": {
			want: 0,
		},
		"When a sequence contains an empty code slice, the opening length is correct.": {
			inCodes: []ansi.Code{},
			want:    0,
		},
		"When a sequence contains codes, the opening length is correct.": {
			inCodes: []ansi.Code{
				ansi.Green,
			},
			want: len("\x1b[32m"),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			seq := ansi.NewSequence(tc.inCodes)
			got := seq.OpenLen()

			// Assert.
			claim.Equal(t, tcName, tc.want, got, LenLabel)
		})
	}
}

func Test_Sequence_ResetLen(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inCodes []ansi.Code
		want    int
	}{
		"When a sequence contains no codes, the reset length is correct.": {
			want: 0,
		},
		"When a sequence contains an empty code slice, the reset length is correct.": {
			inCodes: []ansi.Code{},
			want:    0,
		},
		"When a sequence contains codes, the reset length is correct.": {
			inCodes: []ansi.Code{
				ansi.Green,
			},
			want: len("\x1b[0m"),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			seq := ansi.NewSequence(tc.inCodes)
			got := seq.ResetLen()

			// Assert.
			claim.Equal(t, tcName, tc.want, got, LenLabel)
		})
	}
}

func benchmark_WriteOpenToFreshBuilder(b *testing.B, codeCount int) {
	b.Helper()

	codes := make([]ansi.Code, codeCount)

	for i := range codes {
		codes[i] = ansi.Green
	}

	totalLen := ansi.OpenLen(codes)
	seq := ansi.NewSequence(codes)

	for b.Loop() {
		var sb strings.Builder

		sb.Grow(totalLen)

		seq.WriteOpenTo(&sb)
	}
}

func Benchmark_WriteOpenToFreshBuilder_10Codes(b *testing.B) {
	benchmark_WriteOpenToFreshBuilder(b, 10)
}

func Benchmark_WriteOpenToFreshBuilder_100Codes(b *testing.B) {
	benchmark_WriteOpenToFreshBuilder(b, 100)
}

func Benchmark_WriteOpenToFreshBuilder_1000Codes(b *testing.B) {
	benchmark_WriteOpenToFreshBuilder(b, 1000)
}

func Benchmark_WriteOpenToFreshBuilder_10000Codes(b *testing.B) {
	benchmark_WriteOpenToFreshBuilder(b, 10_000)
}

func benchmark_NewSequence(b *testing.B, codeCount int) {
	b.Helper()

	codes := make([]ansi.Code, codeCount)

	for i := range codes {
		codes[i] = ansi.Green
	}

	for b.Loop() {
		_ = ansi.NewSequence(codes)
	}
}

func Benchmark_NewSequence_10Codes(b *testing.B)    { benchmark_NewSequence(b, 10) }
func Benchmark_NewSequence_100Codes(b *testing.B)   { benchmark_NewSequence(b, 100) }
func Benchmark_NewSequence_1000Codes(b *testing.B)  { benchmark_NewSequence(b, 1000) }
func Benchmark_NewSequence_10000Codes(b *testing.B) { benchmark_NewSequence(b, 10_000) }
