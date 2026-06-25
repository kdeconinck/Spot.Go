// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the scanner package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package scanner_test

import (
	"testing"

	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/scanner"
	"github.com/kdeconinck/spot/syntax"
)

func Test_Scanner_Next(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []syntax.Token
	}{
		"When scanning an empty scope block, the expected tokens are returned.": {
			inSource: "scope {}",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 0, 5),
				token(syntax.TokenLeftBrace, "{", 6, 7),
				token(syntax.TokenRightBrace, "}", 7, 8),
				token(syntax.TokenEOF, "", 8, 8),
			},
		},
		"When scanning whitespace around an empty scope block, whitespace is skipped.": {
			inSource: "\n scope \t{\r\n}\n",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 2, 7),
				token(syntax.TokenLeftBrace, "{", 9, 10),
				token(syntax.TokenRightBrace, "}", 12, 13),
				token(syntax.TokenEOF, "", 14, 14),
			},
		},
		"When scanning a comment before an empty scope block, the comment is skipped.": {
			inSource: "// comment\nscope {}",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 11, 16),
				token(syntax.TokenLeftBrace, "{", 17, 18),
				token(syntax.TokenRightBrace, "}", 18, 19),
				token(syntax.TokenEOF, "", 19, 19),
			},
		},
		"When scanning a comment inside an empty scope block, the comment is skipped.": {
			inSource: "scope {// comment\n}",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 0, 5),
				token(syntax.TokenLeftBrace, "{", 6, 7),
				token(syntax.TokenRightBrace, "}", 18, 19),
				token(syntax.TokenEOF, "", 19, 19),
			},
		},
		"When scanning unknown text, an invalid token is returned.": {
			inSource: "x",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "x", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		"When scanning an unknown symbol, an invalid token is returned.": {
			inSource: "@",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "@", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		"When scanning a slash that does not start a comment, an invalid token is returned.": {
			inSource: "/",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "/", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		"When scanning an identifier that starts with a keyword, an invalid token is returned.": {
			inSource: "scopex",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "scopex", 0, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			scan := scanner.New(tc.inSource)

			// Act and assert.
			for idx := range tc.want {
				got := scan.Next()

				claim.Equal(t, tcName, tc.want[idx], got, "Token")
			}
		})
	}
}

func benchmark_Scanner_Next(b *testing.B, source string) {
	b.Helper()

	for b.Loop() {
		scan := scanner.New(source)

		for tok := scan.Next(); tok.Kind != syntax.TokenEOF; tok = scan.Next() {
			// NOTE: Intentionally left blank.
		}
	}
}

func Benchmark_Scanner_Next_EmptyScopeBlock(b *testing.B) {
	benchmark_Scanner_Next(b, "scope {}")
}

func token(kind syntax.TokenKind, text string, start, end location.Position) syntax.Token {
	return syntax.Token{
		Kind: kind,
		Text: text,
		Span: location.Span{
			Start: start,
			End:   end,
		},
	}
}
