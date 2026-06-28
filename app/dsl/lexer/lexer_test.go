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
		"When source starts with a space, the space is skipped.": {
			inSource: " scope",
			want: []token.Token{
				makeToken(token.TokenScope, 1, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When source starts with a tab, the tab is skipped.": {
			inSource: "\tscope",
			want: []token.Token{
				makeToken(token.TokenScope, 1, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When source starts with a carriage return, the carriage return is skipped.": {
			inSource: "\rscope",
			want: []token.Token{
				makeToken(token.TokenScope, 1, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When source starts with a newline, the newline is skipped.": {
			inSource: "\nscope",
			want: []token.Token{
				makeToken(token.TokenScope, 1, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When source starts with a line comment, the comment is skipped.": {
			inSource: "// comment\nscope",
			want: []token.Token{
				makeToken(token.TokenScope, 11, 16),
				makeToken(token.TokenEOF, 16, 16),
			},
		},
		"When source is only a line comment, EOF is returned after the comment.": {
			inSource: "// comment",
			want: []token.Token{
				makeToken(token.TokenEOF, 10, 10),
			},
		},
		"When source starts with a slash that is not a comment, the slash is not skipped.": {
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

func Test_Lexer_Next_ReturnsEOF(t *testing.T) {
	t.Parallel()

	// Arrange.
	lex := lexer.New("")

	// Act.
	got, want := lex.Next(), makeToken(token.TokenEOF, 0, 0)

	// Assert.
	claim.Equal(t, "When lexing empty source, EOF is returned.", want, got, "Token")
}

func Test_Lexer_Next_ScansString(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When lexing a string literal with plain text, a string token is returned.": {
			inSource: `"abc"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing a string literal with an escaped backslash, a string token is returned.": {
			inSource: `"\\"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped quote, a string token is returned.": {
			inSource: `"\""`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped newline, a string token is returned.": {
			inSource: `"\n"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped carriage return, a string token is returned.": {
			inSource: `"\r"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an escaped tab, a string token is returned.": {
			inSource: `"\t"`,
			want: []token.Token{
				makeToken(token.TokenString, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with an invalid escape, an invalid token is returned.": {
			inSource: `"\x"`,
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenIdentifier, 2, 3),
				makeToken(token.TokenInvalid, 3, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a string literal with a newline, an invalid token is returned.": {
			inSource: "\"abc\n\"",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 4),
				makeToken(token.TokenInvalid, 5, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing a string literal with a carriage return, an invalid token is returned.": {
			inSource: "\"abc\r\"",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 4),
				makeToken(token.TokenInvalid, 5, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing a string literal without a closing quote, an invalid token is returned.": {
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
		"When lexing a character literal with a plain character, a character token is returned.": {
			inSource: `'a'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 3),
				makeToken(token.TokenEOF, 3, 3),
			},
		},
		"When lexing a character literal with an escaped backslash, a character token is returned.": {
			inSource: `'\\'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped quote, a character token is returned.": {
			inSource: `'\''`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped newline, a character token is returned.": {
			inSource: `'\n'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped carriage return, a character token is returned.": {
			inSource: `'\r'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an escaped tab, a character token is returned.": {
			inSource: `'\t'`,
			want: []token.Token{
				makeToken(token.TokenCharacter, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with an invalid escape, an invalid token is returned.": {
			inSource: `'\x'`,
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenIdentifier, 2, 3),
				makeToken(token.TokenInvalid, 3, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with a newline, an invalid token is returned.": {
			inSource: "'a\n'",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenInvalid, 3, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal with a carriage return, an invalid token is returned.": {
			inSource: "'a\r'",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenInvalid, 3, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing a character literal whose first character is a newline, an invalid token is returned.": {
			inSource: "'\n'",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 1),
				makeToken(token.TokenInvalid, 2, 3),
				makeToken(token.TokenEOF, 3, 3),
			},
		},
		"When lexing a character literal whose first character is a carriage return, an invalid token is returned.": {
			inSource: "'\r'",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 1),
				makeToken(token.TokenInvalid, 2, 3),
				makeToken(token.TokenEOF, 3, 3),
			},
		},
		"When lexing a character literal without a closing quote, an invalid token is returned.": {
			inSource: `'a`,
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing an empty character literal, an invalid token is returned.": {
			inSource: `''`,
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 2),
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

func Test_Lexer_Next_ScansSymbols(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When lexing an equal sign, an equal token is returned.": {
			inSource: "=",
			want: []token.Token{
				makeToken(token.TokenEqual, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing an equality operator, an equal-equal token is returned.": {
			inSource: "==",
			want: []token.Token{
				makeToken(token.TokenEqualEqual, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing an exclamation mark without an equals sign, an invalid token is returned.": {
			inSource: "!",
			want: []token.Token{
				makeToken(token.TokenInvalid, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing an inequality operator, a bang-equal token is returned.": {
			inSource: "!=",
			want: []token.Token{
				makeToken(token.TokenBangEqual, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing a less-than comparison operator, a less token is returned.": {
			inSource: "<",
			want: []token.Token{
				makeToken(token.TokenLess, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a less-than-or-equal comparison operator, a less-equal token is returned.": {
			inSource: "<=",
			want: []token.Token{
				makeToken(token.TokenLessEqual, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing a greater-than comparison operator, a greater token is returned.": {
			inSource: ">",
			want: []token.Token{
				makeToken(token.TokenGreater, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a greater-than-or-equal comparison operator, a greater-equal token is returned.": {
			inSource: ">=",
			want: []token.Token{
				makeToken(token.TokenGreaterEqual, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing a dot, a dot token is returned.": {
			inSource: ".",
			want: []token.Token{
				makeToken(token.TokenDot, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a dot-dot operator, a dot-dot token is returned.": {
			inSource: "..",
			want: []token.Token{
				makeToken(token.TokenDotDot, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing a left parenthesis, a left-parenthesis token is returned.": {
			inSource: "(",
			want: []token.Token{
				makeToken(token.TokenLeftParen, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a right parenthesis, a right-parenthesis token is returned.": {
			inSource: ")",
			want: []token.Token{
				makeToken(token.TokenRightParen, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a left brace, a left-brace token is returned.": {
			inSource: "{",
			want: []token.Token{
				makeToken(token.TokenLeftBrace, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a right brace, a right-brace token is returned.": {
			inSource: "}",
			want: []token.Token{
				makeToken(token.TokenRightBrace, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a pipe, a pipe token is returned.": {
			inSource: "|",
			want: []token.Token{
				makeToken(token.TokenPipe, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a question mark, a question token is returned.": {
			inSource: "?",
			want: []token.Token{
				makeToken(token.TokenQuestion, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a star, a star token is returned.": {
			inSource: "*",
			want: []token.Token{
				makeToken(token.TokenStar, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing a plus, a plus token is returned.": {
			inSource: "+",
			want: []token.Token{
				makeToken(token.TokenPlus, 0, 1),
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

func Test_Lexer_Next_ScansIdentifiers(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource string
		want     []token.Token
	}{
		"When lexing the scope keyword, a scope token is returned.": {
			inSource: "scope",
			want: []token.Token{
				makeToken(token.TokenScope, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing the include keyword, an include token is returned.": {
			inSource: "include",
			want: []token.Token{
				makeToken(token.TokenInclude, 0, 7),
				makeToken(token.TokenEOF, 7, 7),
			},
		},
		"When lexing the exclude keyword, an exclude token is returned.": {
			inSource: "exclude",
			want: []token.Token{
				makeToken(token.TokenExclude, 0, 7),
				makeToken(token.TokenEOF, 7, 7),
			},
		},
		"When lexing the definitions keyword, a definitions token is returned.": {
			inSource: "definitions",
			want: []token.Token{
				makeToken(token.TokenDefinitions, 0, 11),
				makeToken(token.TokenEOF, 11, 11),
			},
		},
		"When lexing the tokens keyword, a tokens token is returned.": {
			inSource: "tokens",
			want: []token.Token{
				makeToken(token.TokenTokens, 0, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing the skip keyword, a skip token is returned.": {
			inSource: "skip",
			want: []token.Token{
				makeToken(token.TokenSkip, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the rules keyword, a rules token is returned.": {
			inSource: "rules",
			want: []token.Token{
				makeToken(token.TokenRules, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing the rule keyword, a rule token is returned.": {
			inSource: "rule",
			want: []token.Token{
				makeToken(token.TokenRule, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the match keyword, a match token is returned.": {
			inSource: "match",
			want: []token.Token{
				makeToken(token.TokenMatch, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing the where keyword, a where token is returned.": {
			inSource: "where",
			want: []token.Token{
				makeToken(token.TokenWhere, 0, 5),
				makeToken(token.TokenEOF, 5, 5),
			},
		},
		"When lexing the report keyword, a report token is returned.": {
			inSource: "report",
			want: []token.Token{
				makeToken(token.TokenReport, 0, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing the info keyword, an info token is returned.": {
			inSource: "info",
			want: []token.Token{
				makeToken(token.TokenInfo, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the warn keyword, a warn token is returned.": {
			inSource: "warn",
			want: []token.Token{
				makeToken(token.TokenWarn, 0, 4),
				makeToken(token.TokenEOF, 4, 4),
			},
		},
		"When lexing the err keyword, an err token is returned.": {
			inSource: "err",
			want: []token.Token{
				makeToken(token.TokenErr, 0, 3),
				makeToken(token.TokenEOF, 3, 3),
			},
		},
		"When lexing the at keyword, an at token is returned.": {
			inSource: "at",
			want: []token.Token{
				makeToken(token.TokenAt, 0, 2),
				makeToken(token.TokenEOF, 2, 2),
			},
		},
		"When lexing a lowercase identifier, an identifier token is returned.": {
			inSource: "letter",
			want: []token.Token{
				makeToken(token.TokenIdentifier, 0, 6),
				makeToken(token.TokenEOF, 6, 6),
			},
		},
		"When lexing an uppercase identifier, an identifier token is returned.": {
			inSource: "Identifier",
			want: []token.Token{
				makeToken(token.TokenIdentifier, 0, 10),
				makeToken(token.TokenEOF, 10, 10),
			},
		},
		"When lexing an identifier with digits, an identifier token is returned.": {
			inSource: "letter1",
			want: []token.Token{
				makeToken(token.TokenIdentifier, 0, 7),
				makeToken(token.TokenEOF, 7, 7),
			},
		},
		"When lexing an identifier with underscores, an identifier token is returned.": {
			inSource: "identifier_part",
			want: []token.Token{
				makeToken(token.TokenIdentifier, 0, 15),
				makeToken(token.TokenEOF, 15, 15),
			},
		},
		"When lexing an identifier that starts with a keyword, an identifier token is returned.": {
			inSource: "scopex",
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
		"When lexing a single digit, an integer token is returned.": {
			inSource: "1",
			want: []token.Token{
				makeToken(token.TokenInteger, 0, 1),
				makeToken(token.TokenEOF, 1, 1),
			},
		},
		"When lexing multiple digits, an integer token is returned.": {
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

func Test_Lexer_Next_ReturnsInvalidmakeToken(t *testing.T) {
	t.Parallel()

	// Arrange.
	lex := lexer.New("@")

	// Act.
	got, want := lex.Next(), makeToken(token.TokenInvalid, 0, 1)

	// Assert.
	claim.Equal(t, "When lexing an unknown symbol, an invalid token is returned.", want, got, "Token")
}

func Test_Lexer_Next_PreservesSpans(t *testing.T) {
	t.Parallel()

	// Arrange.
	inSource := "scope { include \"**/*.go\" exclude \"vendor/**\" }\ndefinitions { letter = 'a'..'z' | 'A'..'Z' }\ntokens { Whitespace = ' '+ skip }\nrules { rule PublicIdentifier { match Identifier where Identifier.length > 1 report warn at Identifier \"Public identifier found\" } }"

	want := []token.Token{
		makeToken(token.TokenScope, 0, 5),
		makeToken(token.TokenLeftBrace, 6, 7),
		makeToken(token.TokenInclude, 8, 15),
		makeToken(token.TokenString, 16, 25),
		makeToken(token.TokenExclude, 26, 33),
		makeToken(token.TokenString, 34, 45),
		makeToken(token.TokenRightBrace, 46, 47),
		makeToken(token.TokenDefinitions, 48, 59),
		makeToken(token.TokenLeftBrace, 60, 61),
		makeToken(token.TokenIdentifier, 62, 68),
		makeToken(token.TokenEqual, 69, 70),
		makeToken(token.TokenCharacter, 71, 74),
		makeToken(token.TokenDotDot, 74, 76),
		makeToken(token.TokenCharacter, 76, 79),
		makeToken(token.TokenPipe, 80, 81),
		makeToken(token.TokenCharacter, 82, 85),
		makeToken(token.TokenDotDot, 85, 87),
		makeToken(token.TokenCharacter, 87, 90),
		makeToken(token.TokenRightBrace, 91, 92),
		makeToken(token.TokenTokens, 93, 99),
		makeToken(token.TokenLeftBrace, 100, 101),
		makeToken(token.TokenIdentifier, 102, 112),
		makeToken(token.TokenEqual, 113, 114),
		makeToken(token.TokenCharacter, 115, 118),
		makeToken(token.TokenPlus, 118, 119),
		makeToken(token.TokenSkip, 120, 124),
		makeToken(token.TokenRightBrace, 125, 126),
		makeToken(token.TokenRules, 127, 132),
		makeToken(token.TokenLeftBrace, 133, 134),
		makeToken(token.TokenRule, 135, 139),
		makeToken(token.TokenIdentifier, 140, 156),
		makeToken(token.TokenLeftBrace, 157, 158),
		makeToken(token.TokenMatch, 159, 164),
		makeToken(token.TokenIdentifier, 165, 175),
		makeToken(token.TokenWhere, 176, 181),
		makeToken(token.TokenIdentifier, 182, 192),
		makeToken(token.TokenDot, 192, 193),
		makeToken(token.TokenIdentifier, 193, 199),
		makeToken(token.TokenGreater, 200, 201),
		makeToken(token.TokenInteger, 202, 203),
		makeToken(token.TokenReport, 204, 210),
		makeToken(token.TokenWarn, 211, 215),
		makeToken(token.TokenAt, 216, 218),
		makeToken(token.TokenIdentifier, 219, 229),
		makeToken(token.TokenString, 230, 255),
		makeToken(token.TokenRightBrace, 256, 257),
		makeToken(token.TokenRightBrace, 258, 259),
		makeToken(token.TokenEOF, 259, 259),
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

	inputData :=
		"scope {\n" +
			strings.Repeat("    include \"**/*.go\"\n    exclude \"vendor/**\"\n", size) +
			"}\n" +
			"definitions {\n" +
			strings.Repeat("    letter = 'a'..'z' | 'A'..'Z'\n    value = ('a' | 'b')+\n", size) +
			"}\n" +
			"tokens {\n" +
			strings.Repeat("    Whitespace = ' '+ skip\n", size) +
			"}\n" +
			"rules {\n" +
			strings.Repeat("    rule PublicIdentifier { match Identifier where Identifier.length > 1 report warn at Identifier \"Public identifier found\" }\n", size) +
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
