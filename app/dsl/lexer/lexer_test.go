// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the lexer package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package lexer_test

import (
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/lexer"
	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Lexer_Next_SkipsTrivia(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When source starts with a space, the returned value is correct.": {
			inSource: " scope",
			want: []token.Token{
				makeToken(token.TokenScope, 1, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When source starts with a tab, the returned value is correct.": {
			inSource: "\tscope",
			want: []token.Token{
				makeToken(token.TokenScope, 1, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When source starts with a carriage return, the returned value is correct.": {
			inSource: "\rscope",
			want: []token.Token{
				makeToken(token.TokenScope, 1, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When source starts with a newline, the returned value is correct.": {
			inSource: "\nscope",
			want: []token.Token{
				makeToken(token.TokenScope, 1, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When source starts with a line comment, the returned value is correct.": {
			inSource: "// comment\nscope",
			want: []token.Token{
				makeToken(token.TokenScope, 11, 16),
				makeToken(token.TokenEOF, 16, 16),
			},
		},
		"When source is only a line comment, the returned value is correct.": {
			inSource: "// comment",
			want: []token.Token{
				makeToken(token.TokenEOF, 10, 10),
			},
		},
		"When source starts with a slash that is not a comment, the returned value is correct.": {
			inSource: "/",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tcName, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansSymbols(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When lexing a left parenthesis, the returned value is correct.": {
			inSource: "(",
			want: []token.Token{
				makeToken(token.TokenLeftParen, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a right parenthesis, the returned value is correct.": {
			inSource: ")",
			want: []token.Token{
				makeToken(token.TokenRightParen, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a colon, the returned value is correct.": {
			inSource: ":",
			want: []token.Token{
				makeToken(token.TokenColon, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a left brace, the returned value is correct.": {
			inSource: "{",
			want: []token.Token{
				makeToken(token.TokenLeftBrace, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a right brace, the returned value is correct.": {
			inSource: "}",
			want: []token.Token{
				makeToken(token.TokenRightBrace, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a pipe, the returned value is correct.": {
			inSource: "|",
			want: []token.Token{
				makeToken(token.TokenPipe, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a question mark, the returned value is correct.": {
			inSource: "?",
			want: []token.Token{
				makeToken(token.TokenQuestion, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a star, the returned value is correct.": {
			inSource: "*",
			want: []token.Token{
				makeToken(token.TokenStar, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a plus, the returned value is correct.": {
			inSource: "+",
			want: []token.Token{
				makeToken(token.TokenPlus, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a dot, the returned value is correct.": {
			inSource: ".",
			want: []token.Token{
				makeToken(token.TokenDot, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a dot-dot operator, the returned value is correct.": {
			inSource: "..",
			want: []token.Token{
				makeToken(token.TokenDotDot, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing an equal sign, the returned value is correct.": {
			inSource: "=",
			want: []token.Token{
				makeToken(token.TokenEqual, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing an equality operator, the returned value is correct.": {
			inSource: "==",
			want: []token.Token{
				makeToken(token.TokenEqualEqual, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing an exclamation mark without an equals sign, the returned value is correct.": {
			inSource: "!",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing an inequality operator, the returned value is correct.": {
			inSource: "!=",
			want: []token.Token{
				makeToken(token.TokenBangEqual, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing a less-than comparison operator, the returned value is correct.": {
			inSource: "<",
			want: []token.Token{
				makeToken(token.TokenLess, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a less-than-or-equal comparison operator, the returned value is correct.": {
			inSource: "<=",
			want: []token.Token{
				makeToken(token.TokenLessEqual, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing a greater-than comparison operator, the returned value is correct.": {
			inSource: ">",
			want: []token.Token{
				makeToken(token.TokenGreater, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a greater-than-or-equal comparison operator, the returned value is correct.": {
			inSource: ">=",
			want: []token.Token{
				makeToken(token.TokenGreaterEqual, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tcName, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansString(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When lexing a string literal with plain text, the returned value is correct.": {
			inSource: `"abc"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing a string literal with an invalid escape, the returned value is correct.": {
			inSource: `"\x"`,
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenIdentifier, 2, 3),
				makeToken(token.TokenInvalid, 3, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped backslash, the returned value is correct.": {
			inSource: `"\\"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped quote, the returned value is correct.": {
			inSource: `"\""`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped newline, the returned value is correct.": {
			inSource: `"\n"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped carriage return, the returned value is correct.": {
			inSource: `"\r"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped tab, the returned value is correct.": {
			inSource: `"\t"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with a newline, the returned value is correct.": {
			inSource: "\"abc\n\"",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 4),
				makeToken(token.TokenInvalid, 5, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing a string literal with a carriage return, the returned value is correct.": {
			inSource: "\"abc\r\"",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 4),
				makeToken(token.TokenInvalid, 5, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing a string literal without a closing quote, the returned value is correct.": {
			inSource: `"abc`,
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tcName, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansCharacter(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When lexing a character literal that's not closed, the returned value is correct.": {
			inSource: `'`,
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a character literal with an invalid escape, the returned value is correct.": {
			inSource: `'\x'`,
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenIdentifier, 2, 3),
				makeToken(token.TokenInvalid, 3, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped backslash, the returned value is correct.": {
			inSource: `'\\'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped quote, the returned value is correct.": {
			inSource: `'\''`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped newline, the returned value is correct.": {
			inSource: `'\n'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped carriage return, the returned value is correct.": {
			inSource: `'\r'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped tab, the returned value is correct.": {
			inSource: `'\t'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with a newline, the returned value is correct.": {
			inSource: "'a\n'",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenInvalid, 3, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with a carriage return, the returned value is correct.": {
			inSource: "'a\r'",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenInvalid, 3, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal whose first character is a newline, the returned value is correct.": {
			inSource: "'\n'",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 1),
				makeToken(token.TokenInvalid, 2, 3),
				makeToken(token.TokenEOF, 3, 3),
			},
		},
		"When lexing a character literal whose first character is a carriage return, the returned value is correct.": {
			inSource: "'\r'",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 1),
				makeToken(token.TokenInvalid, 2, 3),
				makeToken(token.TokenEOF, 3, 3),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tcName, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansIdentifiers(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When lexing the scope keyword, the returned value is correct.": {
			inSource: "scope",
			want: []token.Token{
				makeToken(token.TokenScope, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing the include keyword, the returned value is correct.": {
			inSource: "include",
			want: []token.Token{
				makeToken(token.TokenInclude, 0, 7),
				makeToken(token.TokenEOF, 7, 7),
			},
		},
		"When lexing the exclude keyword, the returned value is correct.": {
			inSource: "exclude",
			want: []token.Token{
				makeToken(token.TokenExclude, 0, 7),
				makeToken(token.TokenEOF, 7, 7),
			},
		},
		"When lexing the definitions keyword, the returned value is correct.": {
			inSource: "definitions",
			want: []token.Token{
				makeToken(token.TokenDefinitions, 0, 11),
				makeToken(token.TokenEOF, 11, 11),
			},
		},
		"When lexing the tokens keyword, the returned value is correct.": {
			inSource: "tokens",
			want: []token.Token{
				makeToken(token.TokenTokens, 0, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing the skip keyword, the returned value is correct.": {
			inSource: "skip",
			want: []token.Token{
				makeToken(token.TokenSkip, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the rules keyword, the returned value is correct.": {
			inSource: "rules",
			want: []token.Token{
				makeToken(token.TokenRules, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing the syntax keyword, the returned value is correct.": {
			inSource: "syntax",
			want: []token.Token{
				makeToken(token.TokenSyntax, 0, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing the node keyword, the returned value is correct.": {
			inSource: "node",
			want: []token.Token{
				makeToken(token.TokenNode, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the rule keyword, the returned value is correct.": {
			inSource: "rule",
			want: []token.Token{
				makeToken(token.TokenRule, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the match keyword, the returned value is correct.": {
			inSource: "match",
			want: []token.Token{
				makeToken(token.TokenMatch, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing the where keyword, the returned value is correct.": {
			inSource: "where",
			want: []token.Token{
				makeToken(token.TokenWhere, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing the report keyword, the returned value is correct.": {
			inSource: "report",
			want: []token.Token{
				makeToken(token.TokenReport, 0, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing the info keyword, the returned value is correct.": {
			inSource: "info",
			want: []token.Token{
				makeToken(token.TokenInfo, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the warn keyword, the returned value is correct.": {
			inSource: "warn",
			want: []token.Token{
				makeToken(token.TokenWarn, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the err keyword, the returned value is correct.": {
			inSource: "err",
			want: []token.Token{
				makeToken(token.TokenErr, 0, 3),
				makeToken(token.TokenEOF, 3, 3),
			},
		},
		"When lexing the at keyword, the returned value is correct.": {
			inSource: "at",
			want: []token.Token{
				makeToken(token.TokenAt, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing an identifier, the returned value is correct.": {
			inSource: "letter",
			want: []token.Token{
				makeToken(token.TokenIdentifier, 0, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tcName, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansIntegers(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When lexing a single digit, the returned value is correct.": {
			inSource: "1",
			want: []token.Token{
				makeToken(token.TokenInteger, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing multiple digits, the returned value is correct.": {
			inSource: "123",
			want: []token.Token{
				makeToken(token.TokenInteger, 0, 3),
				makeToken(token.TokenEOF, 3, 3),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tcName, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ReturnsInvalid(t *testing.T) {
	t.Parallel()

	// Arrange.
	lex := lexer.New("@")

	// Act.
	got, want := lex.Next(), makeToken(token.TokenInvalid, 0, 1)

	// Assert.
	claim.Equal(t, "When lexing an unknown symbol, the returned value is correct.", want, got, "Token")
}

func Test_Lexer_Next_PreservesSpans(t *testing.T) {
	t.Parallel()

	// Arrange.
	inSource := `scope {
  include "**/*.go"
  exclude "vendor/**"
}
definitions {
  letter = 'a'..'z' | 'A'..'Z'
}
tokens {
  Whitespace = ' '+ skip
}
rules {
  rule PublicIdentifier {
    match Identifier
    where Identifier.length > 1
    report warn at Identifier "Public identifier found"
  }
}`

	want := []token.Token{
		makeToken(token.TokenScope, 0, 5),
		makeToken(token.TokenLeftBrace, 6, 7),
		makeToken(token.TokenInclude, 10, 17),
		makeToken(token.TokenString, 18, 27),
		makeToken(token.TokenExclude, 30, 37),
		makeToken(token.TokenString, 38, 49),
		makeToken(token.TokenRightBrace, 50, 51),
		makeToken(token.TokenDefinitions, 52, 63),
		makeToken(token.TokenLeftBrace, 64, 65),
		makeToken(token.TokenIdentifier, 68, 74),
		makeToken(token.TokenEqual, 75, 76),
		makeToken(token.TokenCharacter, 77, 80),
		makeToken(token.TokenDotDot, 80, 82),
		makeToken(token.TokenCharacter, 82, 85),
		makeToken(token.TokenPipe, 86, 87),
		makeToken(token.TokenCharacter, 88, 91),
		makeToken(token.TokenDotDot, 91, 93),
		makeToken(token.TokenCharacter, 93, 96),
		makeToken(token.TokenRightBrace, 97, 98),
		makeToken(token.TokenTokens, 99, 105),
		makeToken(token.TokenLeftBrace, 106, 107),
		makeToken(token.TokenIdentifier, 110, 120),
		makeToken(token.TokenEqual, 121, 122),
		makeToken(token.TokenCharacter, 123, 126),
		makeToken(token.TokenPlus, 126, 127),
		makeToken(token.TokenSkip, 128, 132),
		makeToken(token.TokenRightBrace, 133, 134),
		makeToken(token.TokenRules, 135, 140),
		makeToken(token.TokenLeftBrace, 141, 142),
		makeToken(token.TokenRule, 145, 149),
		makeToken(token.TokenIdentifier, 150, 166),
		makeToken(token.TokenLeftBrace, 167, 168),
		makeToken(token.TokenMatch, 173, 178),
		makeToken(token.TokenIdentifier, 179, 189),
		makeToken(token.TokenWhere, 194, 199),
		makeToken(token.TokenIdentifier, 200, 210),
		makeToken(token.TokenDot, 210, 211),
		makeToken(token.TokenIdentifier, 211, 217),
		makeToken(token.TokenGreater, 218, 219),
		makeToken(token.TokenInteger, 220, 221),
		makeToken(token.TokenReport, 226, 232),
		makeToken(token.TokenWarn, 233, 237),
		makeToken(token.TokenAt, 238, 240),
		makeToken(token.TokenIdentifier, 241, 251),
		makeToken(token.TokenString, 252, 277),
		makeToken(token.TokenRightBrace, 280, 281),
		makeToken(token.TokenRightBrace, 282, 283),
		makeToken(token.TokenEOF, 283, 283),
	}

	lex := lexer.New(inSource)

	// Act & assert.
	for idx := range want {
		got := lex.Next()

		claim.Equal(t, "When lexing realistic DSL source, token spans are preserved.", want[idx], got, "Token")
	}
}

func benchmark_Lexer_Next(b *testing.B, source string) {
	b.Helper()

	for b.Loop() {
		lex := lexer.New(source)

		for tok := lex.Next(); tok.Kind != token.TokenEOF; tok = lex.Next() {
			// NOTE: Intentionally left blank.
		}
	}
}

func Benchmark_Lexer_Next_DSL_0(b *testing.B)    { benchmark_Lexer_Next_DSL(b, 0) }
func Benchmark_Lexer_Next_DSL_1(b *testing.B)    { benchmark_Lexer_Next_DSL(b, 1) }
func Benchmark_Lexer_Next_DSL_10(b *testing.B)   { benchmark_Lexer_Next_DSL(b, 10) }
func Benchmark_Lexer_Next_DSL_100(b *testing.B)  { benchmark_Lexer_Next_DSL(b, 100) }
func Benchmark_Lexer_Next_DSL_1000(b *testing.B) { benchmark_Lexer_Next_DSL(b, 1000) }

func benchmark_Lexer_Next_DSL(b *testing.B, size int) {
	b.Helper()

	const scopeBlock = `  include "**/*.go"
  exclude "vendor/**"
`
	const definitionsBlock = `  lower = 'a'..'z'
  upper = 'A'..'Z'
  digit = '0'..'9'
  letter = lower | upper
  alphanumeric = letter | digit
  word = (letter | digit | '_')+
  optionalSign = ('+' | '-')?
  padding = (' ' | '\t')*
`
	const tokensBlock = `  Whitespace = ' '+ skip
  Identifier = letter (alphanumeric | '_')*
  SignedInteger = optionalSign digit+
`
	const rulesBlock = `  rule PublicIdentifier {
    match Identifier
    where Identifier.length == 1
    where Identifier.length != 2
    where Identifier.length < 3
    where Identifier.length <= 4
    where Identifier.length > 5
    where Identifier.length >= 6
    report info at Identifier "short identifier"
    report warn at Identifier "public identifier found"
    report err at Identifier "identifier too long"
  }
`

	inputData :=
		"scope {\n" +
			strings.Repeat(scopeBlock, size) +
			"}\n" +
			"definitions {\n" +
			strings.Repeat(definitionsBlock, size) +
			"}\n" +
			"tokens {\n" +
			strings.Repeat(tokensBlock, size) +
			"}\n" +
			"rules {\n" +
			strings.Repeat(rulesBlock, size) +
			"}"

	benchmark_Lexer_Next(b, inputData)
}

func makeToken(kind token.TokenKind, start, end location.Position) token.Token {
	return token.Token{
		Kind: kind,
		Span: location.Span{
			Start: start,
			End:   end,
		},
	}
}
