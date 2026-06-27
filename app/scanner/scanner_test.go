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
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/compiler"
	"github.com/kdeconinck/spot/ir"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/parser"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/scanner"
	"github.com/kdeconinck/spot/validator"
)

func Test_Scanner_Scan(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inDSL           string
		inSource        string
		wantTokens      []scanner.Token
		wantDiagnostics []scanner.Diagnostic
	}{
		{
			name: "When the input contains identifiers and whitespace, skipped tokens are not emitted.",
			inDSL: `scope { include "**/*.go" }
definitions {
    letter = 'a'..'z' | 'A'..'Z'
    digit = '0'..'9'
    identifierStart = letter | '_'
    identifierPart = identifierStart | digit
}
tokens {
    Whitespace = (' ' | '\t' | '\n')+ skip
    Identifier = identifierStart identifierPart*
}`,
			inSource: "alpha beta_2",
			wantTokens: []scanner.Token{
				token("Identifier", "alpha", 0, 5),
				token("Identifier", "beta_2", 6, 12),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name: "When two tokens match the same text, the earlier token wins.",
			inDSL: `scope { include "**/*.go" }
definitions {
    letter = 'a'..'z'
    identifier = letter+
}
tokens {
    Keyword = "public"
    Identifier = identifier
}`,
			inSource: "public",
			wantTokens: []scanner.Token{
				token("Keyword", "public", 0, 6),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name: "When a token contains an optional prefix, the scanner accepts both present and skipped branches.",
			inDSL: `scope { include "**/*.go" }
tokens {
    OptionalPrefix = '-'? ('0'..'9')+
}`,
			inSource: "-12 34",
			wantTokens: []scanner.Token{
				token("OptionalPrefix", "-12", 0, 3),
			},
			wantDiagnostics: []scanner.Diagnostic{
				diagnostic("No token matched at byte offset 3.", 3, 4),
			},
		},
		{
			name: "When optional branches reconverge before consuming input, the scanner still reaches the following token tail.",
			inDSL: `scope { include "**/*.go" }
tokens {
    Value = ('a'? | 'a'?) 'b'
}`,
			inSource: "b",
			wantTokens: []scanner.Token{
				token("Value", "b", 0, 1),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name: "When an optional expression repeats before a required suffix, epsilon cycles are handled without revisiting states forever.",
			inDSL: `scope { include "**/*.go" }
tokens {
    Value = ('a'?)* 'b'
}`,
			inSource: "aaab",
			wantTokens: []scanner.Token{
				token("Value", "aaab", 0, 4),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name: "When tokens match different lengths, the longest match wins.",
			inDSL: `scope { include "**/*.go" }
tokens {
    Equal = "="
    EqualEqual = "=="
}`,
			inSource: "==",
			wantTokens: []scanner.Token{
				token("EqualEqual", "==", 0, 2),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name: "When two string tokens share a prefix, the longer string token wins.",
			inDSL: `scope { include "**/*.go" }
tokens {
    A = "a"
    Spot = "spot"
    Spotted = "spotted"
}`,
			inSource: "spotted",
			wantTokens: []scanner.Token{
				token("Spotted", "spotted", 0, 7),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name: "When two branches consume the same bytes before reconverging, duplicate next-state closures are ignored.",
			inDSL: `scope { include "**/*.go" }
tokens {
    Value = "ab" | "ab"
}`,
			inSource: "ab",
			wantTokens: []scanner.Token{
				token("Value", "ab", 0, 2),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name: "When a token uses alternation and repetition, the resulting spans are preserved.",
			inDSL: `scope { include "**/*.go" }
tokens {
    Whitespace = ' '+ skip
    Number = ('0'..'9')+
    Identifier = ('a'..'z' | 'A'..'Z')+
}`,
			inSource: "abc 123",
			wantTokens: []scanner.Token{
				token("Identifier", "abc", 0, 3),
				token("Number", "123", 4, 7),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name: "When alternatives reconverge before more input is consumed, the scanner still produces one token.",
			inDSL: `scope { include "**/*.go" }
tokens {
    Value = ('a' | 'a') 'b'
}`,
			inSource: "ab",
			wantTokens: []scanner.Token{
				token("Value", "ab", 0, 2),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name:     "When no token matches at an offset, a diagnostic is returned.",
			inDSL:    `scope { include "**/*.go" } tokens { Identifier = "a" }`,
			inSource: "ab",
			wantTokens: []scanner.Token{
				token("Identifier", "a", 0, 1),
			},
			wantDiagnostics: []scanner.Diagnostic{
				diagnostic("No token matched at byte offset 1.", 1, 2),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			program := compileProgram(t, tc.inDSL)
			scan := scanner.New(program, tc.inSource)
			gotTokens := make([]scanner.Token, 0, len(tc.wantTokens))
			gotDiagnostics := make([]scanner.Diagnostic, 0, len(tc.wantDiagnostics))

			// Act.
			for {
				token, diagnostic, ok := scan.Next()

				if !ok {
					break
				}

				if diagnostic.Message != "" {
					gotDiagnostics = append(gotDiagnostics, diagnostic)
					break
				}

				gotTokens = append(gotTokens, token)
			}

			// Assert.
			claim.DeepEqual(t, tc.name, tc.wantTokens, gotTokens, "Token")
			claim.DeepEqual(t, tc.name, tc.wantDiagnostics, gotDiagnostics, "Diagnostic")
		})
	}
}

func Test_Scanner_Next(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name            string
		inDSL           string
		inSource        string
		wantTokens      []scanner.Token
		wantDiagnostics []scanner.Diagnostic
	}{
		{
			name: "When the scanner streams over skipped tokens, only emitted tokens are returned.",
			inDSL: `scope { include "**/*.go" }
definitions {
    letter = 'a'..'z'
}
tokens {
    Whitespace = ' '+ skip
    Identifier = letter+
}`,
			inSource: "a b",
			wantTokens: []scanner.Token{
				token("Identifier", "a", 0, 1),
				token("Identifier", "b", 2, 3),
			},
			wantDiagnostics: []scanner.Diagnostic{},
		},
		{
			name:     "When the scanner reaches an invalid byte while streaming, a diagnostic is returned and scanning stops.",
			inDSL:    `scope { include "**/*.go" } tokens { Identifier = "a" }`,
			inSource: "ab",
			wantTokens: []scanner.Token{
				token("Identifier", "a", 0, 1),
			},
			wantDiagnostics: []scanner.Diagnostic{
				diagnostic("No token matched at byte offset 1.", 1, 2),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			program := compileProgram(t, tc.inDSL)
			scan := scanner.New(program, tc.inSource)
			gotTokens := make([]scanner.Token, 0, len(tc.wantTokens))
			gotDiagnostics := make([]scanner.Diagnostic, 0, len(tc.wantDiagnostics))

			// Act.
			for {
				token, diagnostic, ok := scan.Next()

				if !ok {
					break
				}

				if diagnostic.Message != "" {
					gotDiagnostics = append(gotDiagnostics, diagnostic)
					break
				}

				gotTokens = append(gotTokens, token)
			}

			// Assert.
			claim.DeepEqual(t, tc.name, tc.wantTokens, gotTokens, "Token")
			claim.DeepEqual(t, tc.name, tc.wantDiagnostics, gotDiagnostics, "Diagnostic")
		})
	}
}

func Benchmark_Scanner_Scan_0(b *testing.B)    { benchmark_Scanner_Scan(b, 0) }
func Benchmark_Scanner_Scan_1(b *testing.B)    { benchmark_Scanner_Scan(b, 1) }
func Benchmark_Scanner_Scan_10(b *testing.B)   { benchmark_Scanner_Scan(b, 10) }
func Benchmark_Scanner_Scan_100(b *testing.B)  { benchmark_Scanner_Scan(b, 100) }
func Benchmark_Scanner_Scan_1000(b *testing.B) { benchmark_Scanner_Scan(b, 1000) }

func benchmark_Scanner_Scan(b *testing.B, size int) {
	b.Helper()

	program := compileProgram(b, scannerDSL())
	scan := scanner.New(program, "")
	source := scannerInput(size)

	for b.Loop() {
		scan.Reset(source)

		for {
			_, _, ok := scan.Next()

			if !ok {
				break
			}
		}
	}
}

func compileProgram(tb testing.TB, source string) ir.Program {
	tb.Helper()

	document, parseDiagnostics := parser.Parse(source)
	validationDiagnostics := validator.Validate(document)

	if len(parseDiagnostics) != 0 {
		tb.Fatalf("scanner test parse diagnostics: got %d, want 0", len(parseDiagnostics))
	}

	if len(validationDiagnostics) != 0 {
		tb.Fatalf("scanner test validation diagnostics: got %d, want 0", len(validationDiagnostics))
	}

	return compiler.Compile(document)
}

func scannerDSL() string {
	return `scope { include "**/*.go" }
definitions {
    letter = 'a'..'z' | 'A'..'Z'
    digit = '0'..'9'
    identifierStart = letter | '_'
    identifierPart = identifierStart | digit
}
tokens {
    Whitespace = (' ' | '\n' | '\t')+ skip
    Identifier = identifierStart identifierPart*
    Number = digit+
}`
}

func scannerInput(size int) string {
	var sb strings.Builder

	sb.WriteString("name0 0")

	for idx := 1; idx <= size; idx++ {
		sb.WriteByte(' ')
		sb.WriteString("name")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteByte(' ')
		sb.WriteString(strconv.Itoa(idx))
	}

	return sb.String()
}

func token(name, text string, start, end location.Position) scanner.Token {
	return scanner.Token{
		Name: name,
		Text: text,
		Span: location.Span{
			Start: start,
			End:   end,
		},
	}
}

func diagnostic(message string, start, end location.Position) scanner.Diagnostic {
	return scanner.Diagnostic{
		Message: message,
		Span: location.Span{
			Start: start,
			End:   end,
		},
	}
}
