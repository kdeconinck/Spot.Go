// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the syntax package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package syntax_test

import (
	"testing"

	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/syntax"
)

func Test_TokenKind_String(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		inKind syntax.TokenKind
		want   string
	}{
		{
			name:   "When using an invalid token kind, the returned value is correct.",
			inKind: syntax.TokenInvalid,
			want:   "invalid",
		},
		{
			name:   "When using an EOF token kind, the returned value is correct.",
			inKind: syntax.TokenEOF,
			want:   "EOF",
		},
		{
			name:   "When using a scope token kind, the returned value is correct.",
			inKind: syntax.TokenScope,
			want:   "scope",
		},
		{
			name:   "When using a definitions token kind, the returned value is correct.",
			inKind: syntax.TokenDefinitions,
			want:   "definitions",
		},
		{
			name:   "When using a tokens token kind, the returned value is correct.",
			inKind: syntax.TokenTokens,
			want:   "tokens",
		},
		{
			name:   "When using a rules token kind, the returned value is correct.",
			inKind: syntax.TokenRules,
			want:   "rules",
		},
		{
			name:   "When using a skip token kind, the returned value is correct.",
			inKind: syntax.TokenSkip,
			want:   "skip",
		},
		{
			name:   "When using a rule token kind, the returned value is correct.",
			inKind: syntax.TokenRule,
			want:   "rule",
		},
		{
			name:   "When using a match token kind, the returned value is correct.",
			inKind: syntax.TokenMatch,
			want:   "match",
		},
		{
			name:   "When using a where token kind, the returned value is correct.",
			inKind: syntax.TokenWhere,
			want:   "where",
		},
		{
			name:   "When using a report token kind, the returned value is correct.",
			inKind: syntax.TokenReport,
			want:   "report",
		},
		{
			name:   "When using an at token kind, the returned value is correct.",
			inKind: syntax.TokenAt,
			want:   "at",
		},
		{
			name:   "When using an info token kind, the returned value is correct.",
			inKind: syntax.TokenInfo,
			want:   "info",
		},
		{
			name:   "When using a warn token kind, the returned value is correct.",
			inKind: syntax.TokenWarn,
			want:   "warn",
		},
		{
			name:   "When using an err token kind, the returned value is correct.",
			inKind: syntax.TokenErr,
			want:   "err",
		},
		{
			name:   "When using an include token kind, the returned value is correct.",
			inKind: syntax.TokenInclude,
			want:   "include",
		},
		{
			name:   "When using an exclude token kind, the returned value is correct.",
			inKind: syntax.TokenExclude,
			want:   "exclude",
		},
		{
			name:   "When using a left-brace token kind, the returned value is correct.",
			inKind: syntax.TokenLeftBrace,
			want:   "{",
		},
		{
			name:   "When using a right-brace token kind, the returned value is correct.",
			inKind: syntax.TokenRightBrace,
			want:   "}",
		},
		{
			name:   "When using a left-parenthesis token kind, the returned value is correct.",
			inKind: syntax.TokenLeftParen,
			want:   "(",
		},
		{
			name:   "When using a right-parenthesis token kind, the returned value is correct.",
			inKind: syntax.TokenRightParen,
			want:   ")",
		},
		{
			name:   "When using a string token kind, the returned value is correct.",
			inKind: syntax.TokenString,
			want:   "string",
		},
		{
			name:   "When using an integer token kind, the returned value is correct.",
			inKind: syntax.TokenInteger,
			want:   "integer",
		},
		{
			name:   "When using an identifier token kind, the returned value is correct.",
			inKind: syntax.TokenIdentifier,
			want:   "identifier",
		},
		{
			name:   "When using an equal token kind, the returned value is correct.",
			inKind: syntax.TokenEqual,
			want:   "=",
		},
		{
			name:   "When using an equal-equal token kind, the returned value is correct.",
			inKind: syntax.TokenEqualEqual,
			want:   "==",
		},
		{
			name:   "When using a bang-equal token kind, the returned value is correct.",
			inKind: syntax.TokenBangEqual,
			want:   "!=",
		},
		{
			name:   "When using a less token kind, the returned value is correct.",
			inKind: syntax.TokenLess,
			want:   "<",
		},
		{
			name:   "When using a less-equal token kind, the returned value is correct.",
			inKind: syntax.TokenLessEqual,
			want:   "<=",
		},
		{
			name:   "When using a greater token kind, the returned value is correct.",
			inKind: syntax.TokenGreater,
			want:   ">",
		},
		{
			name:   "When using a greater-equal token kind, the returned value is correct.",
			inKind: syntax.TokenGreaterEqual,
			want:   ">=",
		},
		{
			name:   "When using a dot token kind, the returned value is correct.",
			inKind: syntax.TokenDot,
			want:   ".",
		},
		{
			name:   "When using a dot-dot token kind, the returned value is correct.",
			inKind: syntax.TokenDotDot,
			want:   "..",
		},
		{
			name:   "When using a pipe token kind, the returned value is correct.",
			inKind: syntax.TokenPipe,
			want:   "|",
		},
		{
			name:   "When using a question token kind, the returned value is correct.",
			inKind: syntax.TokenQuestion,
			want:   "?",
		},
		{
			name:   "When using a star token kind, the returned value is correct.",
			inKind: syntax.TokenStar,
			want:   "*",
		},
		{
			name:   "When using a plus token kind, the returned value is correct.",
			inKind: syntax.TokenPlus,
			want:   "+",
		},
		{
			name:   "When using a character token kind, the returned value is correct.",
			inKind: syntax.TokenCharacter,
			want:   "character",
		},
		{
			name:   "When using an unknown token kind, the returned value is correct.",
			inKind: syntax.TokenKind(255),
			want:   "unknown",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := tc.inKind.String()

			// Assert.
			claim.Equal(t, tc.name, tc.want, got, "Token Kind")
		})
	}
}
