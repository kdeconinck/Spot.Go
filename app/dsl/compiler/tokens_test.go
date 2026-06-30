// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the compiler package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package compiler_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/compiler"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Compile_Tokens(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource    string
		wantProgram string
	}{
		"When compiling tokens, source order is preserved.": {
			inSource: `scope { include "**/*.go" } tokens { First = "a" Second = "b" Third = "c" }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token First
				      String "a"
				    Token Second
				      String "b"
				    Token Third
				      String "c"
				  Rules
			`),
		},
		"When compiling tokens with definition references, expressions are resolved.": {
			inSource: `scope { include "**/*.go" } definitions { letter = 'a'..'z' identifier = letter (letter | '_')* } tokens { Identifier = identifier }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      Concatenation
				        Range 'a' 'z'
				        Repetition *
				          Alternation
				            Range 'a' 'z'
				            Character '_'
				  Rules
			`),
		},
		"When compiling literal tokens, escape sequences are unescaped.": {
			inSource: "scope { include \"**/*.go\" } tokens { Newline = '\\n' Text = \"a\\tb\\n\\\"c\\\\\" }",
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Newline
				      Character '\n'
				    Token Text
				      String "a\tb\n\"c\\"
				  Rules
			`),
		},
		"When compiling a backslash character literal, the literal is unescaped.": {
			inSource: `scope { include "**/*.go" } tokens { Backslash = '\\' }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Backslash
				      Character '\\'
				  Rules
			`),
		},
		"When compiling a single quote character literal, the literal is unescaped.": {
			inSource: `scope { include "**/*.go" } tokens { Quote = '\'' }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Quote
				      Character '\''
				  Rules
			`),
		},
		"When compiling a carriage return character literal, the literal is unescaped.": {
			inSource: `scope { include "**/*.go" } tokens { CarriageReturn = '\r' }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token CarriageReturn
				      Character '\r'
				  Rules
			`),
		},
		"When compiling a string literal with backslash, quote, and carriage return escapes, the literal is unescaped.": {
			inSource: `scope { include "**/*.go" } tokens { Escapes = "\\\"\r" }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Escapes
				      String "\\\"\r"
				  Rules
			`),
		},
		"When compiling skipped tokens, skip flags are preserved.": {
			inSource: `scope { include "**/*.go" } definitions { whitespace = ' ' | '\t' } tokens { Whitespace = whitespace+ skip Identifier = "id" }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Whitespace
				      Repetition +
				        Alternation
				          Character ' '
				          Character '\t'
				      Skip
				    Token Identifier
				      String "id"
				  Rules
			`),
		},
		"When compiling a token with zero-or-one repetition inside a concatenation, the repetition kind is preserved.": {
			inSource: `scope { include "**/*.go" } tokens { Optional = "a"? "b" }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Optional
				      Concatenation
				        Repetition ?
				          String "a"
				        String "b"
				  Rules
			`),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			document, parseErr := parser.Parse(tc.inSource)
			resolution := resolver.Resolve(tc.inSource, document)
			validationDiagnostics := validator.Validate(tc.inSource, resolution)

			// Act.
			gotProgram := compiler.Compile(tc.inSource, resolution)

			// Assert.
			claim.Equal(t, tcName, error(nil), parseErr, "Parse Error")
			claim.Equal(t, tcName, 0, len(validationDiagnostics), "Validation Diagnostic Count")
			claim.Equal(t, tcName, tc.wantProgram, renderProgram(gotProgram), "Program")
		})
	}
}

func Benchmark_Compile_Tokens_0(b *testing.B)    { benchmark_Compile_Tokens(b, 0) }
func Benchmark_Compile_Tokens_1(b *testing.B)    { benchmark_Compile_Tokens(b, 1) }
func Benchmark_Compile_Tokens_10(b *testing.B)   { benchmark_Compile_Tokens(b, 10) }
func Benchmark_Compile_Tokens_100(b *testing.B)  { benchmark_Compile_Tokens(b, 100) }
func Benchmark_Compile_Tokens_1000(b *testing.B) { benchmark_Compile_Tokens(b, 1000) }

func benchmark_Compile_Tokens(b *testing.B, size int) {
	b.Helper()

	source := tokensDSL(size)
	document, parseErr := parser.Parse(source)
	resolution := resolver.Resolve(source, document)
	validationDiagnostics := validator.Validate(source, resolution)
	claim.Equal(b, "Compile tokens benchmark parse error.", error(nil), parseErr, "Parse Error")
	claim.Equal(b, "Compile tokens benchmark validation diagnostics.", 0, len(validationDiagnostics), "Validation Diagnostic Count")

	for b.Loop() {
		_ = compiler.Compile(source, resolution)
	}
}

func tokensDSL(size int) string {
	var sb strings.Builder

	sb.WriteString("scope { include \"**/*.go\" }\n")
	sb.WriteString("definitions {\n")
	sb.WriteString("    letter = 'a'..'z' | 'A'..'Z'\n")
	sb.WriteString("    digit = '0'..'9'\n")
	sb.WriteString("    identifier = letter (letter | digit | '_')*\n")
	sb.WriteString("}\n")
	sb.WriteString("tokens {\n")
	sb.WriteString("    Identifier = identifier\n")
	sb.WriteString("    KeywordPublic = \"public\"\n")
	sb.WriteString("    Whitespace = (' ' | '\\t')+ skip\n")

	for idx := 1; idx <= size; idx++ {
		sb.WriteString("    Identifier")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = identifier\n")
		sb.WriteString("    Keyword")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = \"public\"\n")
		sb.WriteString("    Whitespace")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(" = (' ' | '\\t')+ skip\n")
	}

	sb.WriteString("}")

	return sb.String()
}
