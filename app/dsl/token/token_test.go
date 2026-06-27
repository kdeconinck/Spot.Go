// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the syntax package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package token_test

import (
	"testing"

	"github.com/kdeconinck/spot/dsl/token"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Lookup(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inValue string
		want    token.TokenKind
	}{
		"When using 'scope', the correct value is returned.": {
			inValue: "scope",
			want:    token.TokenScope,
		},
		"When using 'include', the correct value is returned.": {
			inValue: "include",
			want:    token.TokenInclude,
		},
		"When using 'exclude', the correct value is returned.": {
			inValue: "exclude",
			want:    token.TokenExclude,
		},
		"When using 'definitions', the correct value is returned.": {
			inValue: "definitions",
			want:    token.TokenDefinitions,
		},
		"When using 'tokens', the correct value is returned.": {
			inValue: "tokens",
			want:    token.TokenTokens,
		},
		"When using 'skip', the correct value is returned.": {
			inValue: "skip",
			want:    token.TokenSkip,
		},
		"When using 'rules', the correct value is returned.": {
			inValue: "rules",
			want:    token.TokenRules,
		},
		"When using 'rule', the correct value is returned.": {
			inValue: "rule",
			want:    token.TokenRule,
		},
		"When using 'match', the correct value is returned.": {
			inValue: "match",
			want:    token.TokenMatch,
		},
		"When using 'where', the correct value is returned.": {
			inValue: "where",
			want:    token.TokenWhere,
		},
		"When using 'report', the correct value is returned.": {
			inValue: "report",
			want:    token.TokenReport,
		},
		"When using 'info', the correct value is returned.": {
			inValue: "info",
			want:    token.TokenInfo,
		},
		"When using 'warn', the correct value is returned.": {
			inValue: "warn",
			want:    token.TokenWarn,
		},
		"When using 'err', the correct value is returned.": {
			inValue: "err",
			want:    token.TokenErr,
		},
		"When using 'at', the correct value is returned.": {
			inValue: "at",
			want:    token.TokenAt,
		},
		"When using 'unsupported', the correct value is returned.": {
			inValue: "unsupported",
			want:    token.TokenIdentifier,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := token.LookupTokenKind(tc.inValue)

			// Assert.
			claim.Equal(t, tcName, tc.want, got, "Token Kind")
		})
	}
}

func Test_TokenKind_String(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inKind token.TokenKind
		want   string
	}{
		"When using an invalid token kind, the returned value is correct.": {
			inKind: token.TokenInvalid,
			want:   "invalid",
		},
		"When using an EOF token kind, the returned value is correct.": {
			inKind: token.TokenEOF,
			want:   "EOF",
		},
		"When using a scope token kind, the returned value is correct.": {
			inKind: token.TokenScope,
			want:   "scope",
		},
		"When using a definitions token kind, the returned value is correct.": {
			inKind: token.TokenDefinitions,
			want:   "definitions",
		},
		"When using a tokens token kind, the returned value is correct.": {
			inKind: token.TokenTokens,
			want:   "tokens",
		},
		"When using a rules token kind, the returned value is correct.": {
			inKind: token.TokenRules,
			want:   "rules",
		},
		"When using a skip token kind, the returned value is correct.": {
			inKind: token.TokenSkip,
			want:   "skip",
		},
		"When using a rule token kind, the returned value is correct.": {
			inKind: token.TokenRule,
			want:   "rule",
		},
		"When using a match token kind, the returned value is correct.": {
			inKind: token.TokenMatch,
			want:   "match",
		},
		"When using a where token kind, the returned value is correct.": {
			inKind: token.TokenWhere,
			want:   "where",
		},
		"When using a report token kind, the returned value is correct.": {
			inKind: token.TokenReport,
			want:   "report",
		},
		"When using an at token kind, the returned value is correct.": {
			inKind: token.TokenAt,
			want:   "at",
		},
		"When using an info token kind, the returned value is correct.": {
			inKind: token.TokenInfo,
			want:   "info",
		},
		"When using a warn token kind, the returned value is correct.": {
			inKind: token.TokenWarn,
			want:   "warn",
		},
		"When using an err token kind, the returned value is correct.": {
			inKind: token.TokenErr,
			want:   "err",
		},
		"When using an include token kind, the returned value is correct.": {
			inKind: token.TokenInclude,
			want:   "include",
		},
		"When using an exclude token kind, the returned value is correct.": {
			inKind: token.TokenExclude,
			want:   "exclude",
		},
		"When using a left-brace token kind, the returned value is correct.": {
			inKind: token.TokenLeftBrace,
			want:   "{",
		},
		"When using a right-brace token kind, the returned value is correct.": {
			inKind: token.TokenRightBrace,
			want:   "}",
		},
		"When using a left-parenthesis token kind, the returned value is correct.": {
			inKind: token.TokenLeftParen,
			want:   "(",
		},
		"When using a right-parenthesis token kind, the returned value is correct.": {
			inKind: token.TokenRightParen,
			want:   ")",
		},
		"When using a string token kind, the returned value is correct.": {
			inKind: token.TokenString,
			want:   "string",
		},
		"When using an integer token kind, the returned value is correct.": {
			inKind: token.TokenInteger,
			want:   "integer",
		},
		"When using an identifier token kind, the returned value is correct.": {
			inKind: token.TokenIdentifier,
			want:   "identifier",
		},
		"When using an equal token kind, the returned value is correct.": {
			inKind: token.TokenEqual,
			want:   "=",
		},
		"When using an equal-equal token kind, the returned value is correct.": {
			inKind: token.TokenEqualEqual,
			want:   "==",
		},
		"When using a bang-equal token kind, the returned value is correct.": {
			inKind: token.TokenBangEqual,
			want:   "!=",
		},
		"When using a less token kind, the returned value is correct.": {
			inKind: token.TokenLess,
			want:   "<",
		},
		"When using a less-equal token kind, the returned value is correct.": {
			inKind: token.TokenLessEqual,
			want:   "<=",
		},
		"When using a greater token kind, the returned value is correct.": {
			inKind: token.TokenGreater,
			want:   ">",
		},
		"When using a greater-equal token kind, the returned value is correct.": {
			inKind: token.TokenGreaterEqual,
			want:   ">=",
		},
		"When using a dot token kind, the returned value is correct.": {
			inKind: token.TokenDot,
			want:   ".",
		},
		"When using a dot-dot token kind, the returned value is correct.": {
			inKind: token.TokenDotDot,
			want:   "..",
		},
		"When using a pipe token kind, the returned value is correct.": {
			inKind: token.TokenPipe,
			want:   "|",
		},
		"When using a question token kind, the returned value is correct.": {
			inKind: token.TokenQuestion,
			want:   "?",
		},
		"When using a star token kind, the returned value is correct.": {
			inKind: token.TokenStar,
			want:   "*",
		},
		"When using a plus token kind, the returned value is correct.": {
			inKind: token.TokenPlus,
			want:   "+",
		},
		"When using a character token kind, the returned value is correct.": {
			inKind: token.TokenCharacter,
			want:   "character",
		},
		"When using an unknown token kind, the returned value is correct.": {
			inKind: token.TokenKind(255),
			want:   "unknown",
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := tc.inKind.String()

			// Assert.
			claim.Equal(t, tcName, tc.want, got, "Token Kind")
		})
	}
}
