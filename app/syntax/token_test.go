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

func Test_Lookup(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inValue string
		want    syntax.TokenKind
	}{
		"When using 'scope', the correct value is returned.": {
			inValue: "scope",
			want:    syntax.TokenScope,
		},
		"When using 'include', the correct value is returned.": {
			inValue: "include",
			want:    syntax.TokenInclude,
		},
		"When using 'exclude', the correct value is returned.": {
			inValue: "exclude",
			want:    syntax.TokenExclude,
		},
		"When using 'definitions', the correct value is returned.": {
			inValue: "definitions",
			want:    syntax.TokenDefinitions,
		},
		"When using 'tokens', the correct value is returned.": {
			inValue: "tokens",
			want:    syntax.TokenTokens,
		},
		"When using 'skip', the correct value is returned.": {
			inValue: "skip",
			want:    syntax.TokenSkip,
		},
		"When using 'rules', the correct value is returned.": {
			inValue: "rules",
			want:    syntax.TokenRules,
		},
		"When using 'rule', the correct value is returned.": {
			inValue: "rule",
			want:    syntax.TokenRule,
		},
		"When using 'match', the correct value is returned.": {
			inValue: "match",
			want:    syntax.TokenMatch,
		},
		"When using 'where', the correct value is returned.": {
			inValue: "where",
			want:    syntax.TokenWhere,
		},
		"When using 'report', the correct value is returned.": {
			inValue: "report",
			want:    syntax.TokenReport,
		},
		"When using 'info', the correct value is returned.": {
			inValue: "info",
			want:    syntax.TokenInfo,
		},
		"When using 'warn', the correct value is returned.": {
			inValue: "warn",
			want:    syntax.TokenWarn,
		},
		"When using 'err', the correct value is returned.": {
			inValue: "err",
			want:    syntax.TokenErr,
		},
		"When using 'at', the correct value is returned.": {
			inValue: "at",
			want:    syntax.TokenAt,
		},
		"When using 'unsupported', the correct value is returned.": {
			inValue: "unsupported",
			want:    syntax.TokenIdentifier,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := syntax.LookupTokenKind(tc.inValue)

			// Assert.
			claim.Equal(t, tcName, tc.want, got, "Token Kind")
		})
	}
}

func Test_TokenKind_String(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inKind syntax.TokenKind
		want   string
	}{
		"When using an invalid token kind, the returned value is correct.": {
			inKind: syntax.TokenInvalid,
			want:   "invalid",
		},
		"When using an EOF token kind, the returned value is correct.": {
			inKind: syntax.TokenEOF,
			want:   "EOF",
		},
		"When using a scope token kind, the returned value is correct.": {
			inKind: syntax.TokenScope,
			want:   "scope",
		},
		"When using a definitions token kind, the returned value is correct.": {
			inKind: syntax.TokenDefinitions,
			want:   "definitions",
		},
		"When using a tokens token kind, the returned value is correct.": {
			inKind: syntax.TokenTokens,
			want:   "tokens",
		},
		"When using a rules token kind, the returned value is correct.": {
			inKind: syntax.TokenRules,
			want:   "rules",
		},
		"When using a skip token kind, the returned value is correct.": {
			inKind: syntax.TokenSkip,
			want:   "skip",
		},
		"When using a rule token kind, the returned value is correct.": {
			inKind: syntax.TokenRule,
			want:   "rule",
		},
		"When using a match token kind, the returned value is correct.": {
			inKind: syntax.TokenMatch,
			want:   "match",
		},
		"When using a where token kind, the returned value is correct.": {
			inKind: syntax.TokenWhere,
			want:   "where",
		},
		"When using a report token kind, the returned value is correct.": {
			inKind: syntax.TokenReport,
			want:   "report",
		},
		"When using an at token kind, the returned value is correct.": {
			inKind: syntax.TokenAt,
			want:   "at",
		},
		"When using an info token kind, the returned value is correct.": {
			inKind: syntax.TokenInfo,
			want:   "info",
		},
		"When using a warn token kind, the returned value is correct.": {
			inKind: syntax.TokenWarn,
			want:   "warn",
		},
		"When using an err token kind, the returned value is correct.": {
			inKind: syntax.TokenErr,
			want:   "err",
		},
		"When using an include token kind, the returned value is correct.": {
			inKind: syntax.TokenInclude,
			want:   "include",
		},
		"When using an exclude token kind, the returned value is correct.": {
			inKind: syntax.TokenExclude,
			want:   "exclude",
		},
		"When using a left-brace token kind, the returned value is correct.": {
			inKind: syntax.TokenLeftBrace,
			want:   "{",
		},
		"When using a right-brace token kind, the returned value is correct.": {
			inKind: syntax.TokenRightBrace,
			want:   "}",
		},
		"When using a left-parenthesis token kind, the returned value is correct.": {
			inKind: syntax.TokenLeftParen,
			want:   "(",
		},
		"When using a right-parenthesis token kind, the returned value is correct.": {
			inKind: syntax.TokenRightParen,
			want:   ")",
		},
		"When using a string token kind, the returned value is correct.": {
			inKind: syntax.TokenString,
			want:   "string",
		},
		"When using an integer token kind, the returned value is correct.": {
			inKind: syntax.TokenInteger,
			want:   "integer",
		},
		"When using an identifier token kind, the returned value is correct.": {
			inKind: syntax.TokenIdentifier,
			want:   "identifier",
		},
		"When using an equal token kind, the returned value is correct.": {
			inKind: syntax.TokenEqual,
			want:   "=",
		},
		"When using an equal-equal token kind, the returned value is correct.": {
			inKind: syntax.TokenEqualEqual,
			want:   "==",
		},
		"When using a bang-equal token kind, the returned value is correct.": {
			inKind: syntax.TokenBangEqual,
			want:   "!=",
		},
		"When using a less token kind, the returned value is correct.": {
			inKind: syntax.TokenLess,
			want:   "<",
		},
		"When using a less-equal token kind, the returned value is correct.": {
			inKind: syntax.TokenLessEqual,
			want:   "<=",
		},
		"When using a greater token kind, the returned value is correct.": {
			inKind: syntax.TokenGreater,
			want:   ">",
		},
		"When using a greater-equal token kind, the returned value is correct.": {
			inKind: syntax.TokenGreaterEqual,
			want:   ">=",
		},
		"When using a dot token kind, the returned value is correct.": {
			inKind: syntax.TokenDot,
			want:   ".",
		},
		"When using a dot-dot token kind, the returned value is correct.": {
			inKind: syntax.TokenDotDot,
			want:   "..",
		},
		"When using a pipe token kind, the returned value is correct.": {
			inKind: syntax.TokenPipe,
			want:   "|",
		},
		"When using a question token kind, the returned value is correct.": {
			inKind: syntax.TokenQuestion,
			want:   "?",
		},
		"When using a star token kind, the returned value is correct.": {
			inKind: syntax.TokenStar,
			want:   "*",
		},
		"When using a plus token kind, the returned value is correct.": {
			inKind: syntax.TokenPlus,
			want:   "+",
		},
		"When using a character token kind, the returned value is correct.": {
			inKind: syntax.TokenCharacter,
			want:   "character",
		},
		"When using an unknown token kind, the returned value is correct.": {
			inKind: syntax.TokenKind(255),
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
