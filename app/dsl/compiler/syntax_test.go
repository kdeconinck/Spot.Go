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
				`    node Word { oneOf { Identifier KeywordPublic } }`,
				`    node WordPair { left: Word right: Word }`,
				`    node OptionalWord { value?: oneOf { Word KeywordInternal } }`,
				`    node WordList { values: Word+ }`,
				`    node UnknownStatement { values: any+ }`,
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
				        Capture left
				          Node Word
				        Capture right
				          Node Word
				    Node OptionalWord
				      Capture value
				        Repetition ?
				          Alternation
				            Node Word
				            Token KeywordInternal
				    Node WordList
				      Capture values
				        Repetition +
				          Node Word
				    Node UnknownStatement
				      Capture values
				        Repetition +
				          Any
				  Rules
			`),
		},
		"When compiling syntax nodes with named captures, field metadata is preserved.": {
			inSource: strings.Join([]string{
				`scope { include "**/*.go" }`,
				`tokens { Identifier = "id" }`,
				`syntax {`,
				`    node QualifiedIdentifier { Identifier }`,
				`    node UsingDirective { name: QualifiedIdentifier }`,
				`}`,
			}, "\n"),
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Syntax
				    Node QualifiedIdentifier
				      Token Identifier
				    Node UsingDirective
				      Capture name
				        Node QualifiedIdentifier
				  Rules
			`),
		},
		"When compiling structured syntax nodes, field captures and oneOf are preserved.": {
			inSource: strings.Join([]string{
				`scope { include "**/*.go" }`,
				`tokens { Identifier = "id" KeywordUsing = "using" Semicolon = ";" }`,
				`syntax {`,
				`    node QualifiedIdentifier {`,
				`        head: Identifier`,
				`    }`,
				`    node UsingDirective {`,
				`        KeywordUsing`,
				`        name: QualifiedIdentifier`,
				`        Semicolon`,
				`    }`,
				`    node Root {`,
				`        members*: oneOf {`,
				`            UsingDirective`,
				`            any`,
				`        }`,
				`    }`,
				`}`,
			}, "\n"),
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				    Token KeywordUsing
				      String "using"
				    Token Semicolon
				      String ";"
				  Syntax
				    Node QualifiedIdentifier
				      Capture head
				        Token Identifier
				    Node UsingDirective
				      Concatenation
				        Token KeywordUsing
				        Capture name
				          Node QualifiedIdentifier
				        Token Semicolon
				    Node Root
				      Capture members
				        Repetition *
				          Alternation
				            Node UsingDirective
				            Any
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
	builder.WriteString("    node Word {\n")
	builder.WriteString("        oneOf {\n")
	builder.WriteString("            Identifier\n")
	builder.WriteString("            KeywordPublic\n")
	builder.WriteString("        }\n")
	builder.WriteString("    }\n")
	builder.WriteString("    node WordPair {\n")
	builder.WriteString("        left: Word\n")
	builder.WriteString("        right: Word\n")
	builder.WriteString("    }\n")
	builder.WriteString("    node OptionalWord {\n")
	builder.WriteString("        value?: oneOf {\n")
	builder.WriteString("            Word\n")
	builder.WriteString("            KeywordInternal\n")
	builder.WriteString("        }\n")
	builder.WriteString("    }\n")
	builder.WriteString("    node WordList {\n")
	builder.WriteString("        values: Word+\n")
	builder.WriteString("    }\n")
	builder.WriteString("    node UnknownStatement {\n")
	builder.WriteString("        values: any+\n")
	builder.WriteString("    }\n")

	for idx := 1; idx <= size; idx++ {
		builder.WriteString("    node Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" {\n")
		builder.WriteString("        oneOf {\n")
		builder.WriteString("            Identifier")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("\n")
		builder.WriteString("            KeywordPublic")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("\n")
		builder.WriteString("        }\n")
		builder.WriteString("    }\n")
		builder.WriteString("    node WordPair")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" {\n")
		builder.WriteString("        left: Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("\n")
		builder.WriteString("        right: Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("\n")
		builder.WriteString("    }\n")
		builder.WriteString("    node OptionalWord")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" {\n")
		builder.WriteString("        value?: oneOf {\n")
		builder.WriteString("            Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("\n")
		builder.WriteString("            KeywordInternal")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("\n")
		builder.WriteString("        }\n")
		builder.WriteString("    }\n")
		builder.WriteString("    node WordList")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" {\n")
		builder.WriteString("        values: Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("+\n")
		builder.WriteString("    }\n")
		builder.WriteString("    node UnknownStatement")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" {\n")
		builder.WriteString("        values: any+\n")
		builder.WriteString("    }\n")
	}

	builder.WriteString("}")

	return builder.String()
}
