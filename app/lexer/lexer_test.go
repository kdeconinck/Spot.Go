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

	"github.com/kdeconinck/spot/lexer"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/syntax"
)

func Test_Lexer_Next_SkipsTrivia(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		inSource string
		want     []syntax.Token
	}{
		{
			name:     "When source starts with a space, the space is skipped.",
			inSource: " scope",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 1, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When source starts with a tab, the tab is skipped.",
			inSource: "\tscope",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 1, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When source starts with a carriage return, the carriage return is skipped.",
			inSource: "\rscope",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 1, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When source starts with a newline, the newline is skipped.",
			inSource: "\nscope",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 1, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When source starts with a line comment, the comment is skipped.",
			inSource: "// comment\nscope",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 11, 16),
				token(syntax.TokenEOF, "", 16, 16),
			},
		},
		{
			name:     "When source is only a line comment, EOF is returned after the comment.",
			inSource: "// comment",
			want: []syntax.Token{
				token(syntax.TokenEOF, "", 10, 10),
			},
		},
		{
			name:     "When source starts with a slash that is not a comment, the slash is not skipped.",
			inSource: "/",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "/", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tc.name, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ReturnsEOF(t *testing.T) {
	t.Parallel()

	// Arrange.
	lex := lexer.New("")

	// Act.
	got, want := lex.Next(), token(syntax.TokenEOF, "", 0, 0)

	// Assert.
	claim.Equal(t, "When lexing empty source, EOF is returned.", want, got, "Token")
}

func Test_Lexer_Next_ScansString(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		inSource string
		want     []syntax.Token
	}{
		{
			name:     "When lexing a string literal with plain text, a string token is returned.",
			inSource: `"abc"`,
			want: []syntax.Token{
				token(syntax.TokenString, `"abc"`, 0, 5),
				token(syntax.TokenEOF, "", 5, 5),
			},
		},
		{
			name:     "When lexing a string literal with an escaped backslash, a string token is returned.",
			inSource: `"\\"`,
			want: []syntax.Token{
				token(syntax.TokenString, `"\\"`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a string literal with an escaped quote, a string token is returned.",
			inSource: `"\""`,
			want: []syntax.Token{
				token(syntax.TokenString, `"\""`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a string literal with an escaped newline, a string token is returned.",
			inSource: `"\n"`,
			want: []syntax.Token{
				token(syntax.TokenString, `"\n"`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a string literal with an escaped carriage return, a string token is returned.",
			inSource: `"\r"`,
			want: []syntax.Token{
				token(syntax.TokenString, `"\r"`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a string literal with an escaped tab, a string token is returned.",
			inSource: `"\t"`,
			want: []syntax.Token{
				token(syntax.TokenString, `"\t"`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a string literal with an invalid escape, an invalid token is returned.",
			inSource: `"\x"`,
			want: []syntax.Token{
				token(syntax.TokenInvalid, `"\`, 0, 2),
				token(syntax.TokenIdentifier, "x", 2, 3),
				token(syntax.TokenInvalid, `"`, 3, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a string literal with a newline, an invalid token is returned.",
			inSource: "\"abc\n\"",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "\"abc", 0, 4),
				token(syntax.TokenInvalid, `"`, 5, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When lexing a string literal with a carriage return, an invalid token is returned.",
			inSource: "\"abc\r\"",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "\"abc", 0, 4),
				token(syntax.TokenInvalid, `"`, 5, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When lexing a string literal without a closing quote, an invalid token is returned.",
			inSource: `"abc`,
			want: []syntax.Token{
				token(syntax.TokenInvalid, `"abc`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tc.name, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansCharacter(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		inSource string
		want     []syntax.Token
	}{
		{
			name:     "When lexing a character literal with a plain character, a character token is returned.",
			inSource: `'a'`,
			want: []syntax.Token{
				token(syntax.TokenCharacter, `'a'`, 0, 3),
				token(syntax.TokenEOF, "", 3, 3),
			},
		},
		{
			name:     "When lexing a character literal with an escaped backslash, a character token is returned.",
			inSource: `'\\'`,
			want: []syntax.Token{
				token(syntax.TokenCharacter, `'\\'`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a character literal with an escaped quote, a character token is returned.",
			inSource: `'\''`,
			want: []syntax.Token{
				token(syntax.TokenCharacter, `'\''`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a character literal with an escaped newline, a character token is returned.",
			inSource: `'\n'`,
			want: []syntax.Token{
				token(syntax.TokenCharacter, `'\n'`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a character literal with an escaped carriage return, a character token is returned.",
			inSource: `'\r'`,
			want: []syntax.Token{
				token(syntax.TokenCharacter, `'\r'`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a character literal with an escaped tab, a character token is returned.",
			inSource: `'\t'`,
			want: []syntax.Token{
				token(syntax.TokenCharacter, `'\t'`, 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a character literal with an invalid escape, an invalid token is returned.",
			inSource: `'\x'`,
			want: []syntax.Token{
				token(syntax.TokenInvalid, `'\`, 0, 2),
				token(syntax.TokenIdentifier, "x", 2, 3),
				token(syntax.TokenInvalid, `'`, 3, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a character literal with a newline, an invalid token is returned.",
			inSource: "'a\n'",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "'a", 0, 2),
				token(syntax.TokenInvalid, `'`, 3, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a character literal with a carriage return, an invalid token is returned.",
			inSource: "'a\r'",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "'a", 0, 2),
				token(syntax.TokenInvalid, `'`, 3, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing a character literal whose first character is a newline, an invalid token is returned.",
			inSource: "'\n'",
			want: []syntax.Token{
				token(syntax.TokenInvalid, `'`, 0, 1),
				token(syntax.TokenInvalid, `'`, 2, 3),
				token(syntax.TokenEOF, "", 3, 3),
			},
		},
		{
			name:     "When lexing a character literal whose first character is a carriage return, an invalid token is returned.",
			inSource: "'\r'",
			want: []syntax.Token{
				token(syntax.TokenInvalid, `'`, 0, 1),
				token(syntax.TokenInvalid, `'`, 2, 3),
				token(syntax.TokenEOF, "", 3, 3),
			},
		},
		{
			name:     "When lexing a character literal without a closing quote, an invalid token is returned.",
			inSource: `'a`,
			want: []syntax.Token{
				token(syntax.TokenInvalid, `'a`, 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
		{
			name:     "When lexing an empty character literal, an invalid token is returned.",
			inSource: `''`,
			want: []syntax.Token{
				token(syntax.TokenInvalid, `''`, 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tc.name, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansSymbols(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		inSource string
		want     []syntax.Token
	}{
		{
			name:     "When lexing an equal sign, an equal token is returned.",
			inSource: "=",
			want: []syntax.Token{
				token(syntax.TokenEqual, "=", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing an equality operator, an equal-equal token is returned.",
			inSource: "==",
			want: []syntax.Token{
				token(syntax.TokenEqualEqual, "==", 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
		{
			name:     "When lexing an exclamation mark without an equals sign, an invalid token is returned.",
			inSource: "!",
			want: []syntax.Token{
				token(syntax.TokenInvalid, "!", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing an inequality operator, a bang-equal token is returned.",
			inSource: "!=",
			want: []syntax.Token{
				token(syntax.TokenBangEqual, "!=", 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
		{
			name:     "When lexing a less-than comparison operator, a less token is returned.",
			inSource: "<",
			want: []syntax.Token{
				token(syntax.TokenLess, "<", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a less-than-or-equal comparison operator, a less-equal token is returned.",
			inSource: "<=",
			want: []syntax.Token{
				token(syntax.TokenLessEqual, "<=", 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
		{
			name:     "When lexing a greater-than comparison operator, a greater token is returned.",
			inSource: ">",
			want: []syntax.Token{
				token(syntax.TokenGreater, ">", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a greater-than-or-equal comparison operator, a greater-equal token is returned.",
			inSource: ">=",
			want: []syntax.Token{
				token(syntax.TokenGreaterEqual, ">=", 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
		{
			name:     "When lexing a dot, a dot token is returned.",
			inSource: ".",
			want: []syntax.Token{
				token(syntax.TokenDot, ".", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a dot-dot operator, a dot-dot token is returned.",
			inSource: "..",
			want: []syntax.Token{
				token(syntax.TokenDotDot, "..", 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
		{
			name:     "When lexing a left parenthesis, a left-parenthesis token is returned.",
			inSource: "(",
			want: []syntax.Token{
				token(syntax.TokenLeftParen, "(", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a right parenthesis, a right-parenthesis token is returned.",
			inSource: ")",
			want: []syntax.Token{
				token(syntax.TokenRightParen, ")", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a left brace, a left-brace token is returned.",
			inSource: "{",
			want: []syntax.Token{
				token(syntax.TokenLeftBrace, "{", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a right brace, a right-brace token is returned.",
			inSource: "}",
			want: []syntax.Token{
				token(syntax.TokenRightBrace, "}", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a pipe, a pipe token is returned.",
			inSource: "|",
			want: []syntax.Token{
				token(syntax.TokenPipe, "|", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a question mark, a question token is returned.",
			inSource: "?",
			want: []syntax.Token{
				token(syntax.TokenQuestion, "?", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a star, a star token is returned.",
			inSource: "*",
			want: []syntax.Token{
				token(syntax.TokenStar, "*", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing a plus, a plus token is returned.",
			inSource: "+",
			want: []syntax.Token{
				token(syntax.TokenPlus, "+", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tc.name, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansIdentifiers(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		inSource string
		want     []syntax.Token
	}{
		{
			name:     "When lexing the scope keyword, a scope token is returned.",
			inSource: "scope",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 0, 5),
				token(syntax.TokenEOF, "", 5, 5),
			},
		},
		{
			name:     "When lexing the include keyword, an include token is returned.",
			inSource: "include",
			want: []syntax.Token{
				token(syntax.TokenInclude, "include", 0, 7),
				token(syntax.TokenEOF, "", 7, 7),
			},
		},
		{
			name:     "When lexing the exclude keyword, an exclude token is returned.",
			inSource: "exclude",
			want: []syntax.Token{
				token(syntax.TokenExclude, "exclude", 0, 7),
				token(syntax.TokenEOF, "", 7, 7),
			},
		},
		{
			name:     "When lexing the definitions keyword, a definitions token is returned.",
			inSource: "definitions",
			want: []syntax.Token{
				token(syntax.TokenDefinitions, "definitions", 0, 11),
				token(syntax.TokenEOF, "", 11, 11),
			},
		},
		{
			name:     "When lexing the tokens keyword, a tokens token is returned.",
			inSource: "tokens",
			want: []syntax.Token{
				token(syntax.TokenTokens, "tokens", 0, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When lexing the skip keyword, a skip token is returned.",
			inSource: "skip",
			want: []syntax.Token{
				token(syntax.TokenSkip, "skip", 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing the rules keyword, a rules token is returned.",
			inSource: "rules",
			want: []syntax.Token{
				token(syntax.TokenRules, "rules", 0, 5),
				token(syntax.TokenEOF, "", 5, 5),
			},
		},
		{
			name:     "When lexing the rule keyword, a rule token is returned.",
			inSource: "rule",
			want: []syntax.Token{
				token(syntax.TokenRule, "rule", 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing the match keyword, a match token is returned.",
			inSource: "match",
			want: []syntax.Token{
				token(syntax.TokenMatch, "match", 0, 5),
				token(syntax.TokenEOF, "", 5, 5),
			},
		},
		{
			name:     "When lexing the where keyword, a where token is returned.",
			inSource: "where",
			want: []syntax.Token{
				token(syntax.TokenWhere, "where", 0, 5),
				token(syntax.TokenEOF, "", 5, 5),
			},
		},
		{
			name:     "When lexing the report keyword, a report token is returned.",
			inSource: "report",
			want: []syntax.Token{
				token(syntax.TokenReport, "report", 0, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When lexing the info keyword, an info token is returned.",
			inSource: "info",
			want: []syntax.Token{
				token(syntax.TokenInfo, "info", 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing the warn keyword, a warn token is returned.",
			inSource: "warn",
			want: []syntax.Token{
				token(syntax.TokenWarn, "warn", 0, 4),
				token(syntax.TokenEOF, "", 4, 4),
			},
		},
		{
			name:     "When lexing the err keyword, an err token is returned.",
			inSource: "err",
			want: []syntax.Token{
				token(syntax.TokenErr, "err", 0, 3),
				token(syntax.TokenEOF, "", 3, 3),
			},
		},
		{
			name:     "When lexing the at keyword, an at token is returned.",
			inSource: "at",
			want: []syntax.Token{
				token(syntax.TokenAt, "at", 0, 2),
				token(syntax.TokenEOF, "", 2, 2),
			},
		},
		{
			name:     "When lexing a lowercase identifier, an identifier token is returned.",
			inSource: "letter",
			want: []syntax.Token{
				token(syntax.TokenIdentifier, "letter", 0, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
		{
			name:     "When lexing an uppercase identifier, an identifier token is returned.",
			inSource: "Identifier",
			want: []syntax.Token{
				token(syntax.TokenIdentifier, "Identifier", 0, 10),
				token(syntax.TokenEOF, "", 10, 10),
			},
		},
		{
			name:     "When lexing an identifier with digits, an identifier token is returned.",
			inSource: "letter1",
			want: []syntax.Token{
				token(syntax.TokenIdentifier, "letter1", 0, 7),
				token(syntax.TokenEOF, "", 7, 7),
			},
		},
		{
			name:     "When lexing an identifier with underscores, an identifier token is returned.",
			inSource: "identifier_part",
			want: []syntax.Token{
				token(syntax.TokenIdentifier, "identifier_part", 0, 15),
				token(syntax.TokenEOF, "", 15, 15),
			},
		},
		{
			name:     "When lexing an identifier that starts with a keyword, an identifier token is returned.",
			inSource: "scopex",
			want: []syntax.Token{
				token(syntax.TokenIdentifier, "scopex", 0, 6),
				token(syntax.TokenEOF, "", 6, 6),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tc.name, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ScansIntegers(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		inSource string
		want     []syntax.Token
	}{
		{
			name:     "When lexing a single digit, an integer token is returned.",
			inSource: "1",
			want: []syntax.Token{
				token(syntax.TokenInteger, "1", 0, 1),
				token(syntax.TokenEOF, "", 1, 1),
			},
		},
		{
			name:     "When lexing multiple digits, an integer token is returned.",
			inSource: "123",
			want: []syntax.Token{
				token(syntax.TokenInteger, "123", 0, 3),
				token(syntax.TokenEOF, "", 3, 3),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tc.name, tc.want[idx], got, "Token")
			}
		})
	}
}

func Test_Lexer_Next_ReturnsInvalidToken(t *testing.T) {
	t.Parallel()

	// Arrange.
	lex := lexer.New("@")

	// Act.
	got, want := lex.Next(), token(syntax.TokenInvalid, "@", 0, 1)

	// Assert.
	claim.Equal(t, "When lexing an unknown symbol, an invalid token is returned.", want, got, "Token")
}

func Test_Lexer_Next_PreservesSpans(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		inSource string
		want     []syntax.Token
	}{
		{
			name:     "When lexing realistic DSL source, token spans are preserved.",
			inSource: "scope { include \"**/*.go\" exclude \"vendor/**\" }\ndefinitions { letter = 'a'..'z' | 'A'..'Z' }\ntokens { Whitespace = ' '+ skip }\nrules { rule PublicIdentifier { match Identifier where Identifier.length > 1 report warn at Identifier \"Public identifier found\" } }",
			want: []syntax.Token{
				token(syntax.TokenScope, "scope", 0, 5),
				token(syntax.TokenLeftBrace, "{", 6, 7),
				token(syntax.TokenInclude, "include", 8, 15),
				token(syntax.TokenString, "\"**/*.go\"", 16, 25),
				token(syntax.TokenExclude, "exclude", 26, 33),
				token(syntax.TokenString, "\"vendor/**\"", 34, 45),
				token(syntax.TokenRightBrace, "}", 46, 47),
				token(syntax.TokenDefinitions, "definitions", 48, 59),
				token(syntax.TokenLeftBrace, "{", 60, 61),
				token(syntax.TokenIdentifier, "letter", 62, 68),
				token(syntax.TokenEqual, "=", 69, 70),
				token(syntax.TokenCharacter, "'a'", 71, 74),
				token(syntax.TokenDotDot, "..", 74, 76),
				token(syntax.TokenCharacter, "'z'", 76, 79),
				token(syntax.TokenPipe, "|", 80, 81),
				token(syntax.TokenCharacter, "'A'", 82, 85),
				token(syntax.TokenDotDot, "..", 85, 87),
				token(syntax.TokenCharacter, "'Z'", 87, 90),
				token(syntax.TokenRightBrace, "}", 91, 92),
				token(syntax.TokenTokens, "tokens", 93, 99),
				token(syntax.TokenLeftBrace, "{", 100, 101),
				token(syntax.TokenIdentifier, "Whitespace", 102, 112),
				token(syntax.TokenEqual, "=", 113, 114),
				token(syntax.TokenCharacter, "' '", 115, 118),
				token(syntax.TokenPlus, "+", 118, 119),
				token(syntax.TokenSkip, "skip", 120, 124),
				token(syntax.TokenRightBrace, "}", 125, 126),
				token(syntax.TokenRules, "rules", 127, 132),
				token(syntax.TokenLeftBrace, "{", 133, 134),
				token(syntax.TokenRule, "rule", 135, 139),
				token(syntax.TokenIdentifier, "PublicIdentifier", 140, 156),
				token(syntax.TokenLeftBrace, "{", 157, 158),
				token(syntax.TokenMatch, "match", 159, 164),
				token(syntax.TokenIdentifier, "Identifier", 165, 175),
				token(syntax.TokenWhere, "where", 176, 181),
				token(syntax.TokenIdentifier, "Identifier", 182, 192),
				token(syntax.TokenDot, ".", 192, 193),
				token(syntax.TokenIdentifier, "length", 193, 199),
				token(syntax.TokenGreater, ">", 200, 201),
				token(syntax.TokenInteger, "1", 202, 203),
				token(syntax.TokenReport, "report", 204, 210),
				token(syntax.TokenWarn, "warn", 211, 215),
				token(syntax.TokenAt, "at", 216, 218),
				token(syntax.TokenIdentifier, "Identifier", 219, 229),
				token(syntax.TokenString, "\"Public identifier found\"", 230, 255),
				token(syntax.TokenRightBrace, "}", 256, 257),
				token(syntax.TokenRightBrace, "}", 258, 259),
				token(syntax.TokenEOF, "", 259, 259),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			lex := lexer.New(tc.inSource)

			// Act & assert.
			for idx := range tc.want {
				got := lex.Next()

				claim.Equal(t, tc.name, tc.want[idx], got, "Token")
			}
		})
	}
}

func benchmark_Lexer_Next(b *testing.B, source string) {
	b.Helper()

	for b.Loop() {
		lex := lexer.New(source)

		for tok := lex.Next(); tok.Kind != syntax.TokenEOF; tok = lex.Next() {
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
