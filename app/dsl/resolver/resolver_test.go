// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the resolver package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package resolver_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Resolve_ReturnsSectionData(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := `scope { include "**/*.go" } definitions { letter = 'a' digit = '0' } tokens { Identifier = letter } rules { rule PublicIdentifier { match Identifier report warn at Identifier "x" } }`
	document, parseErr := parser.Parse(source)

	// Act.
	resolution := resolver.Resolve(source, document)

	// Assert.
	claim.Equal(t, "When resolving a full DSL source, no parse error is returned.", error(nil), parseErr, "Parse Error")
	claim.Equal(t, "When resolving a full DSL source, scope entries are preserved.", 1, len(resolution.ScopeEntries), "Scope Entry Count")
	claim.Equal(t, "When resolving a full DSL source, definitions are preserved.", 2, len(resolution.Definitions), "Definition Count")
	claim.Equal(t, "When resolving a full DSL source, tokens are preserved.", 1, len(resolution.Tokens), "Token Count")
	claim.Equal(t, "When resolving a full DSL source, rules are preserved.", 1, len(resolution.Rules), "Rule Count")
}

func Test_Resolve_IndexesFirstDeclarations(t *testing.T) {
	t.Parallel()

	// Arrange.
	source := `scope { include "**/*.go" } definitions { value = 'a' value = 'b' } tokens { Identifier = "id" Identifier = "other" } rules { rule PublicIdentifier { match Identifier report warn at Identifier "x" } rule PublicIdentifier { match Identifier report warn at Identifier "x" } }`
	document, parseErr := parser.Parse(source)

	// Act.
	resolution := resolver.Resolve(source, document)
	definitionIndex, definitionOK := resolution.DefinitionIndex("value")
	tokenIndex, tokenOK := resolution.TokenIndex("Identifier")
	ruleIndex, ruleOK := resolution.RuleIndex("PublicIdentifier")

	// Assert.
	claim.Equal(t, "When resolving duplicated declarations, no parse error is returned.", error(nil), parseErr, "Parse Error")
	claim.Equal(t, "When resolving duplicated definitions, the first declaration index is stored.", 0, definitionIndex, "Definition Index")
	claim.Equal(t, "When resolving duplicated tokens, the first declaration index is stored.", 0, tokenIndex, "Token Index")
	claim.Equal(t, "When resolving duplicated rules, the first declaration index is stored.", 0, ruleIndex, "Rule Index")
	claim.Equal(t, "When resolving duplicated definitions, the name is found.", true, definitionOK, "Definition Found")
	claim.Equal(t, "When resolving duplicated tokens, the name is found.", true, tokenOK, "Token Found")
	claim.Equal(t, "When resolving duplicated rules, the name is found.", true, ruleOK, "Rule Found")
}

func Benchmark_Resolve_DSL_0(b *testing.B)    { benchmark_Resolve_DSL(b, 0) }
func Benchmark_Resolve_DSL_1(b *testing.B)    { benchmark_Resolve_DSL(b, 1) }
func Benchmark_Resolve_DSL_10(b *testing.B)   { benchmark_Resolve_DSL(b, 10) }
func Benchmark_Resolve_DSL_100(b *testing.B)  { benchmark_Resolve_DSL(b, 100) }
func Benchmark_Resolve_DSL_1000(b *testing.B) { benchmark_Resolve_DSL(b, 1000) }

func benchmark_Resolve_DSL(b *testing.B, size int) {
	b.Helper()

	source := resolverDSL(size)
	document, parseErr := parser.Parse(source)
	claim.Equal(b, "Resolve benchmark parse error.", error(nil), parseErr, "Parse Error")

	for b.Loop() {
		_ = resolver.Resolve(source, document)
	}
}

func resolverDSL(size int) string {
	var builder strings.Builder

	builder.WriteString("scope {\n")
	builder.WriteString("    include \"**/*.go\"\n")
	builder.WriteString("    exclude \"vendor/**\"\n")
	builder.WriteString("}\n")
	builder.WriteString("definitions {\n")

	for idx := 0; idx <= size; idx++ {
		suffix := strconv.Itoa(idx)

		builder.WriteString("    letter")
		builder.WriteString(suffix)
		builder.WriteString(" = 'a'..'z' | 'A'..'Z'\n")
		builder.WriteString("    digit")
		builder.WriteString(suffix)
		builder.WriteString(" = '0'..'9'\n")
		builder.WriteString("    identifierStart")
		builder.WriteString(suffix)
		builder.WriteString(" = letter")
		builder.WriteString(suffix)
		builder.WriteString(" | '_'\n")
		builder.WriteString("    identifierPart")
		builder.WriteString(suffix)
		builder.WriteString(" = identifierStart")
		builder.WriteString(suffix)
		builder.WriteString(" | digit")
		builder.WriteString(suffix)
		builder.WriteString("\n")
	}

	builder.WriteString("}\n")
	builder.WriteString("tokens {\n")

	for idx := 0; idx <= size; idx++ {
		suffix := strconv.Itoa(idx)

		builder.WriteString("    Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" = identifierStart")
		builder.WriteString(suffix)
		builder.WriteString(" identifierPart")
		builder.WriteString(suffix)
		builder.WriteString("*\n")
	}

	builder.WriteString("}\n")
	builder.WriteString("rules {\n")

	for idx := 0; idx <= size; idx++ {
		suffix := strconv.Itoa(idx)

		builder.WriteString("    rule PublicIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("        report warn at Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" \"x\"\n")
		builder.WriteString("    }\n")
	}

	builder.WriteString("}")

	return builder.String()
}
