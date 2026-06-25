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
	"strings"
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
		"When scanning an empty definitions block, the expected tokens are returned.": {
			inSource: "definitions {}",
			want: []syntax.Token{
				token(syntax.TokenDefinitions, "definitions", 0, 11),
				token(syntax.TokenLeftBrace, "{", 12, 13),
				token(syntax.TokenRightBrace, "}", 13, 14),
				token(syntax.TokenEOF, "", 14, 14),
			},
		},
		"When scanning a definitions block with a character definition, the expected tokens are returned.": {
			inSource: "definitions { letter = 'a' }",
			want: []syntax.Token{
				token(syntax.TokenDefinitions, "definitions", 0, 11),
				token(syntax.TokenLeftBrace, "{", 12, 13),
				token(syntax.TokenIdentifier, "letter", 14, 20),
				token(syntax.TokenEqual, "=", 21, 22),
				token(syntax.TokenCharacter, "'a'", 23, 26),
				token(syntax.TokenRightBrace, "}", 27, 28),
				token(syntax.TokenEOF, "", 28, 28),
			},
		},
		"When scanning a definitions block with a character range, the expected tokens are returned.": {
			inSource: "definitions { letter = 'a'..'z' }",
			want: []syntax.Token{
				token(syntax.TokenDefinitions, "definitions", 0, 11),
				token(syntax.TokenLeftBrace, "{", 12, 13),
				token(syntax.TokenIdentifier, "letter", 14, 20),
				token(syntax.TokenEqual, "=", 21, 22),
				token(syntax.TokenCharacter, "'a'", 23, 26),
				token(syntax.TokenDotDot, "..", 26, 28),
				token(syntax.TokenCharacter, "'z'", 28, 31),
				token(syntax.TokenRightBrace, "}", 32, 33),
				token(syntax.TokenEOF, "", 33, 33),
			},
		},
		"When scanning a scope block with include and exclude entries, the expected tokens are returned.": {
			inSource: "scope { include \"**/*.go\" exclude \"vendor/**\" }",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 0, 5),
				token(syntax.TokenLeftBrace, "{", 6, 7),
				token(syntax.TokenInclude, "include", 8, 15),
				token(syntax.TokenString, "\"**/*.go\"", 16, 25),
				token(syntax.TokenExclude, "exclude", 26, 33),
				token(syntax.TokenString, "\"vendor/**\"", 34, 45),
				token(syntax.TokenRightBrace, "}", 46, 47),
				token(syntax.TokenEOF, "", 47, 47),
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
		"When scanning a string literal, a string token is returned.": {
			inSource: "\"**/*.go\"",
			want: []syntax.Token{
				token(syntax.TokenString, "\"**/*.go\"", 0, 9),
				token(syntax.TokenEOF, "", 9, 9),
			},
		},
		"When scanning a string literal with valid escapes, a string token is returned.": {
			inSource: `"` + `\\` + `\"` + `\n` + `\r` + `\t` + `"`,
			want: []syntax.Token{
				token(syntax.TokenString, `"`+`\\`+`\"`+`\n`+`\r`+`\t`+`"`, 0, 12),
				token(syntax.TokenEOF, "", 12, 12),
			},
		},
		"When scanning a string literal with an invalid escape, an invalid token is returned.": {
			inSource: "\"\\x\"",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "\"\\", 0, 2),
				token(syntax.TokenIdentifier, "x", 2, 3),
				token(syntax.TokenInvalid, "\"", 3, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		"When scanning a string literal without a closing quote, an invalid token is returned.": {
			inSource: "\"abc",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "\"abc", 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		"When scanning a string literal with a newline, an invalid token is returned.": {
			inSource: "\"abc\n\"",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "\"abc", 0, 4),
				token(syntax.TokenInvalid, "\"", 5, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		"When scanning identifier text, an identifier token is returned.": {
			inSource: "x",
			want: []syntax.Token{
				token(syntax.TokenIdentifier, "x", 0, 1),
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
		"When scanning a single dot, an invalid token is returned.": {
			inSource: ".",
			want: []syntax.Token{
				token(syntax.TokenInvalid, ".", 0, 1),
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
		"When scanning an identifier that starts with a keyword, an identifier token is returned.": {
			inSource: "scopex",
			want: []syntax.Token{
				token(syntax.TokenIdentifier, "scopex", 0, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		"When scanning a character literal with a valid escape, a character token is returned.": {
			inSource: `'\\'`,
			want: []syntax.Token{
				token(syntax.TokenCharacter, `'\\'`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		"When scanning a character literal with an invalid escape, an invalid token is returned.": {
			inSource: `'\x'`,
			want: []syntax.Token{
				token(syntax.TokenInvalid, `'\`, 0, 2),
				token(syntax.TokenIdentifier, "x", 2, 3),
				token(syntax.TokenInvalid, `'`, 3, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		"When scanning a character literal without a closing quote, an invalid token is returned.": {
			inSource: "'a",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "'a", 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
		"When scanning a character literal with a newline, an invalid token is returned.": {
			inSource: "'\n'",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "'", 0, 1),
				token(syntax.TokenInvalid, "'", 2, 3),
				token(syntax.TokenEOF, "", 3, 3),
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
	b.SetBytes(int64(len(source)))

	for b.Loop() {
		scan := scanner.New(source)

		for tok := scan.Next(); tok.Kind != syntax.TokenEOF; tok = scan.Next() {
			// NOTE: Intentionally left blank.
		}
	}
}

func Benchmark_Scanner_Next_DSL_0(b *testing.B)    { benchmark_Scanner_Next_DSL(b, 0, 0) }
func Benchmark_Scanner_Next_DSL_1(b *testing.B)    { benchmark_Scanner_Next_DSL(b, 1, 1) }
func Benchmark_Scanner_Next_DSL_10(b *testing.B)   { benchmark_Scanner_Next_DSL(b, 10, 10) }
func Benchmark_Scanner_Next_DSL_100(b *testing.B)  { benchmark_Scanner_Next_DSL(b, 100, 100) }
func Benchmark_Scanner_Next_DSL_1000(b *testing.B) { benchmark_Scanner_Next_DSL(b, 1000, 1000) }

func benchmark_Scanner_Next_DSL(b *testing.B, scopeEntryPairCount, definitionCount int) {
	b.Helper()

	inputData :=
		"scope {\n" +
			strings.Repeat("    include \"**/*.go\"\n", scopeEntryPairCount) +
			strings.Repeat("    exclude \"vendor/**\"\n", scopeEntryPairCount) +
			"}\n" +
			"definitions {\n" +
			strings.Repeat("    letter = 'a'\n", definitionCount) +
			"}"

	benchmark_Scanner_Next(b, inputData)
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
