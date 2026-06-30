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

func Test_Compile_Syntax(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource    string
		wantProgram string
	}{
		"When compiling syntax nodes, token and node references are resolved.": {
			inSource: strings.Join([]string{
				`scope { include "**/*.go" }`,
				`tokens { Identifier = "id" KeywordPublic = "public" KeywordInternal = "internal" }`,
				`syntax {`,
				`    node Word = Identifier | KeywordPublic`,
				`    node WordPair = Word Word`,
				`    node OptionalWord = (Word | KeywordInternal)?`,
				`    node WordList = Word+`,
				`}`,
			}, "\n"),
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				    Token KeywordPublic
				      String "public"
				    Token KeywordInternal
				      String "internal"
				  Syntax
				    Node Word
				      Alternation
				        Token Identifier
				        Token KeywordPublic
				    Node WordPair
				      Concatenation
				        Node Word
				        Node Word
				    Node OptionalWord
				      Repetition ?
				        Group
				          Alternation
				            Node Word
				            Token KeywordInternal
				    Node WordList
				      Repetition +
				        Node Word
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

func Benchmark_Compile_Syntax_0(b *testing.B)    { benchmark_Compile_Syntax(b, 0) }
func Benchmark_Compile_Syntax_1(b *testing.B)    { benchmark_Compile_Syntax(b, 1) }
func Benchmark_Compile_Syntax_10(b *testing.B)   { benchmark_Compile_Syntax(b, 10) }
func Benchmark_Compile_Syntax_100(b *testing.B)  { benchmark_Compile_Syntax(b, 100) }
func Benchmark_Compile_Syntax_1000(b *testing.B) { benchmark_Compile_Syntax(b, 1000) }

func benchmark_Compile_Syntax(b *testing.B, size int) {
	b.Helper()

	source := syntaxDSL(size)
	document, parseErr := parser.Parse(source)
	resolution := resolver.Resolve(source, document)
	validationDiagnostics := validator.Validate(source, resolution)
	claim.Equal(b, "Compile syntax benchmark parse error.", error(nil), parseErr, "Parse Error")
	claim.Equal(b, "Compile syntax benchmark validation diagnostics.", 0, len(validationDiagnostics), "Validation Diagnostic Count")

	for b.Loop() {
		_ = compiler.Compile(source, resolution)
	}
}

func syntaxDSL(size int) string {
	var builder strings.Builder

	builder.WriteString("scope { include \"**/*.go\" }\n")
	builder.WriteString("tokens {\n")
	builder.WriteString("    Identifier = \"id\"\n")
	builder.WriteString("    KeywordPublic = \"public\"\n")
	builder.WriteString("    KeywordInternal = \"internal\"\n")

	for idx := 1; idx <= size; idx++ {
		builder.WriteString("    Identifier")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" = \"id\"\n")
		builder.WriteString("    KeywordPublic")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" = \"public\"\n")
		builder.WriteString("    KeywordInternal")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" = \"internal\"\n")
	}

	builder.WriteString("}\n")
	builder.WriteString("syntax {\n")
	builder.WriteString("    node Word = Identifier | KeywordPublic\n")
	builder.WriteString("    node WordPair = Word Word\n")
	builder.WriteString("    node OptionalWord = (Word | KeywordInternal)?\n")
	builder.WriteString("    node WordList = Word+\n")

	for idx := 1; idx <= size; idx++ {
		builder.WriteString("    node Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" = Identifier")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" | KeywordPublic")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("\n")
		builder.WriteString("    node WordPair")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" = Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("\n")
		builder.WriteString("    node OptionalWord")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" = (Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" | KeywordInternal")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(")?\n")
		builder.WriteString("    node WordList")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" = Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("+\n")
	}

	builder.WriteString("}")

	return builder.String()
}
