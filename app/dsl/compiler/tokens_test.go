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
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/runtime/ir"
)

func Test_Compile_Tokens_PreservesSourceOrder(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := `scope { include "**/*.go" } tokens { First = "a" Second = "b" Third = "c" }`
	document, parseErr := parser.Parse(source)
	validationDiagnostics := validator.Validate(source, document)
	wantProgram := ir.Program{
		Tokens: []ir.Token{
			{Name: "First", Expression: ir.Expression{Kind: ir.ExpressionString, String: "a"}},
			{Name: "Second", Expression: ir.Expression{Kind: ir.ExpressionString, String: "b"}},
			{Name: "Third", Expression: ir.Expression{Kind: ir.ExpressionString, String: "c"}},
		},
		Rules: []ir.Rule{},
	}

	// Act.
	gotProgram := compiler.Compile(source, document)

	// Assert.
	claim.Equal(t, "When compiling tokens, no parse error is returned.", error(nil), parseErr, "Parse Error")
	claim.Equal(t, "When compiling tokens, validation diagnostics are not returned.", 0, len(validationDiagnostics), "Validation Diagnostic Count")
	claim.DeepEqual(t, "When compiling tokens, source order is preserved.", wantProgram, gotProgram, "Program")
}

func Test_Compile_Tokens_ResolvesDefinitionReferences(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := `scope { include "**/*.go" } definitions { letter = 'a'..'z' identifier = letter (letter | '_')* } tokens { Identifier = identifier }`
	document, parseErr := parser.Parse(source)
	validationDiagnostics := validator.Validate(source, document)
	wantProgram := ir.Program{
		Tokens: []ir.Token{
			{
				Name: "Identifier",
				Expression: ir.Expression{
					Kind: ir.ExpressionConcatenation,
					Terms: []ir.Expression{
						{Kind: ir.ExpressionRange, RangeStart: 'a', RangeEnd: 'z'},
						{
							Kind: ir.ExpressionRepetition,
							Inner: pointer(ir.Expression{
								Kind: ir.ExpressionAlternation,
								Terms: []ir.Expression{
									{Kind: ir.ExpressionRange, RangeStart: 'a', RangeEnd: 'z'},
									{Kind: ir.ExpressionCharacter, Character: '_'},
								},
							}),
							Repetition: ir.RepetitionZeroOrMore,
						},
					},
				},
			},
		},
		Rules: []ir.Rule{},
	}

	// Act.
	gotProgram := compiler.Compile(source, document)

	// Assert.
	claim.Equal(t, "When compiling tokens with definition references, no parse error is returned.", error(nil), parseErr, "Parse Error")
	claim.Equal(t, "When compiling tokens with definition references, validation diagnostics are not returned.", 0, len(validationDiagnostics), "Validation Diagnostic Count")
	claim.DeepEqual(t, "When compiling tokens with definition references, expressions are resolved.", wantProgram, gotProgram, "Program")
}

func Test_Compile_Tokens_UnescapesLiterals(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := "scope { include \"**/*.go\" } tokens { Newline = '\\n' Text = \"a\\tb\\n\\\"c\\\\\" }"
	document, parseErr := parser.Parse(source)
	validationDiagnostics := validator.Validate(source, document)
	wantProgram := ir.Program{
		Tokens: []ir.Token{
			{Name: "Newline", Expression: ir.Expression{Kind: ir.ExpressionCharacter, Character: '\n'}},
			{Name: "Text", Expression: ir.Expression{Kind: ir.ExpressionString, String: "a\tb\n\"c\\"}},
		},
		Rules: []ir.Rule{},
	}

	// Act.
	gotProgram := compiler.Compile(source, document)

	// Assert.
	claim.Equal(t, "When compiling literal tokens, no parse error is returned.", error(nil), parseErr, "Parse Error")
	claim.Equal(t, "When compiling literal tokens, validation diagnostics are not returned.", 0, len(validationDiagnostics), "Validation Diagnostic Count")
	claim.DeepEqual(t, "When compiling literal tokens, escape sequences are unescaped.", wantProgram, gotProgram, "Program")
}

func Test_Compile_Tokens_UnescapesCharacterEscapes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		inSource    string
		wantProgram ir.Program
	}{
		{
			name:     "When compiling a backslash character literal, the literal is unescaped.",
			inSource: `scope { include "**/*.go" } tokens { Backslash = '\\' }`,
			wantProgram: ir.Program{
				Tokens: []ir.Token{
					{Name: "Backslash", Expression: ir.Expression{Kind: ir.ExpressionCharacter, Character: '\\'}},
				},
				Rules: []ir.Rule{},
			},
		},
		{
			name:     "When compiling a single quote character literal, the literal is unescaped.",
			inSource: `scope { include "**/*.go" } tokens { Quote = '\'' }`,
			wantProgram: ir.Program{
				Tokens: []ir.Token{
					{Name: "Quote", Expression: ir.Expression{Kind: ir.ExpressionCharacter, Character: '\''}},
				},
				Rules: []ir.Rule{},
			},
		},
		{
			name:     "When compiling a carriage return character literal, the literal is unescaped.",
			inSource: `scope { include "**/*.go" } tokens { CarriageReturn = '\r' }`,
			wantProgram: ir.Program{
				Tokens: []ir.Token{
					{Name: "CarriageReturn", Expression: ir.Expression{Kind: ir.ExpressionCharacter, Character: '\r'}},
				},
				Rules: []ir.Rule{},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			document, parseErr := parser.Parse(tc.inSource)
			validationDiagnostics := validator.Validate(tc.inSource, document)

			// Act.
			gotProgram := compiler.Compile(tc.inSource, document)

			// Assert.
			claim.Equal(t, tc.name, error(nil), parseErr, "Parse Error")
			claim.Equal(t, tc.name, 0, len(validationDiagnostics), "Validation Diagnostic Count")
			claim.DeepEqual(t, tc.name, tc.wantProgram, gotProgram, "Program")
		})
	}
}

func Test_Compile_Tokens_UnescapesStringEscapes(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := `scope { include "**/*.go" } tokens { Escapes = "\\\"\r" }`
	document, parseErr := parser.Parse(source)
	validationDiagnostics := validator.Validate(source, document)
	wantProgram := ir.Program{
		Tokens: []ir.Token{
			{Name: "Escapes", Expression: ir.Expression{Kind: ir.ExpressionString, String: "\\\"\r"}},
		},
		Rules: []ir.Rule{},
	}

	// Act.
	gotProgram := compiler.Compile(source, document)

	// Assert.
	claim.Equal(t, "When compiling a string literal with backslash, quote, and carriage return escapes, no parse error is returned.", error(nil), parseErr, "Parse Error")
	claim.Equal(t, "When compiling a string literal with backslash, quote, and carriage return escapes, the literal is unescaped.", 0, len(validationDiagnostics), "Validation Diagnostic Count")
	claim.DeepEqual(t, "When compiling a string literal with backslash, quote, and carriage return escapes, the literal is unescaped.", wantProgram, gotProgram, "Program")
}

func Test_Compile_Tokens_PreservesSkipFlags(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := `scope { include "**/*.go" } definitions { whitespace = ' ' | '\t' } tokens { Whitespace = whitespace+ skip Identifier = "id" }`
	document, parseErr := parser.Parse(source)
	validationDiagnostics := validator.Validate(source, document)
	wantProgram := ir.Program{
		Tokens: []ir.Token{
			{
				Name: "Whitespace",
				Expression: ir.Expression{
					Kind: ir.ExpressionRepetition,
					Inner: pointer(ir.Expression{
						Kind: ir.ExpressionAlternation,
						Terms: []ir.Expression{
							{Kind: ir.ExpressionCharacter, Character: ' '},
							{Kind: ir.ExpressionCharacter, Character: '\t'},
						},
					}),
					Repetition: ir.RepetitionOneOrMore,
				},
				Skip: true,
			},
			{Name: "Identifier", Expression: ir.Expression{Kind: ir.ExpressionString, String: "id"}},
		},
		Rules: []ir.Rule{},
	}

	// Act.
	gotProgram := compiler.Compile(source, document)

	// Assert.
	claim.Equal(t, "When compiling skipped tokens, no parse error is returned.", error(nil), parseErr, "Parse Error")
	claim.Equal(t, "When compiling skipped tokens, validation diagnostics are not returned.", 0, len(validationDiagnostics), "Validation Diagnostic Count")
	claim.DeepEqual(t, "When compiling skipped tokens, skip flags are preserved.", wantProgram, gotProgram, "Program")
}

func Test_Compile_Tokens_PreservesZeroOrOneRepetition(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := `scope { include "**/*.go" } tokens { Optional = "a"? "b" }`
	document, parseErr := parser.Parse(source)
	validationDiagnostics := validator.Validate(source, document)
	wantProgram := ir.Program{
		Tokens: []ir.Token{
			{
				Name: "Optional",
				Expression: ir.Expression{
					Kind: ir.ExpressionConcatenation,
					Terms: []ir.Expression{
						{
							Kind:       ir.ExpressionRepetition,
							Inner:      pointer(ir.Expression{Kind: ir.ExpressionString, String: "a"}),
							Repetition: ir.RepetitionZeroOrOne,
						},
						{Kind: ir.ExpressionString, String: "b"},
					},
				},
			},
		},
		Rules: []ir.Rule{},
	}

	// Act.
	gotProgram := compiler.Compile(source, document)

	// Assert.
	claim.Equal(t, "When compiling a token with zero-or-one repetition, no parse error is returned.", error(nil), parseErr, "Parse Error")
	claim.Equal(t, "When compiling a token with zero-or-one repetition inside a concatenation, validation diagnostics are not returned.", 0, len(validationDiagnostics), "Validation Diagnostic Count")
	claim.DeepEqual(t, "When compiling a token with zero-or-one repetition inside a concatenation, the repetition kind is preserved.", wantProgram, gotProgram, "Program")
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
	validationDiagnostics := validator.Validate(source, document)
	claim.Equal(b, "Compile tokens benchmark parse error.", error(nil), parseErr, "Parse Error")
	claim.Equal(b, "Compile tokens benchmark validation diagnostics.", 0, len(validationDiagnostics), "Validation Diagnostic Count")

	for b.Loop() {
		_ = compiler.Compile(source, document)
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

func pointer(expression ir.Expression) *ir.Expression {
	return &expression
}
