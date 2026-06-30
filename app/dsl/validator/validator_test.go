// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the validator package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package validator_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/qa/claim"
)

type markedSource struct {
	text  string
	spans []location.Span
}

type expectedDiagnostic struct {
	message   string
	spanIndex int
}

func Benchmark_Validate_DSL_0(b *testing.B)    { benchmark_Validate_DSL(b, 0) }
func Benchmark_Validate_DSL_1(b *testing.B)    { benchmark_Validate_DSL(b, 1) }
func Benchmark_Validate_DSL_10(b *testing.B)   { benchmark_Validate_DSL(b, 10) }
func Benchmark_Validate_DSL_100(b *testing.B)  { benchmark_Validate_DSL(b, 100) }
func Benchmark_Validate_DSL_1000(b *testing.B) { benchmark_Validate_DSL(b, 1000) }

func benchmark_Validate_DSL(b *testing.B, size int) {
	b.Helper()

	benchmark_Validate(b, fullHappyPathValidationDSL(size))
}

func benchmark_Validate(b *testing.B, source string) {
	b.Helper()

	resolution := parseResolution(b, "Benchmark source parses successfully.", source)

	for b.Loop() {
		_ = validator.Validate(source, resolution)
	}
}

func parseResolution(tb testing.TB, assertionName string, source string) resolver.Resolution {
	tb.Helper()

	document, parseErr := parser.Parse(source)

	claim.Equal(tb, assertionName, error(nil), parseErr, "Parse Error")

	return resolver.Resolve(source, document)
}

func normalizeMultilineLiteral(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	for idx := range lines {
		lines[idx] = strings.TrimRight(strings.TrimLeft(lines[idx], "\t"), " ")
	}

	return strings.Join(lines, "\n")
}

func markedMultilineLiteral(text string) markedSource {
	normalized := normalizeMultilineLiteral(text)

	var builder strings.Builder
	spans := []location.Span{}
	start := -1

	for idx := 0; idx < len(normalized); idx++ {
		if strings.HasPrefix(normalized[idx:], "[[") {
			if start != -1 {
				panic("nested span marker")
			}

			start = builder.Len()
			idx++

			continue
		}

		if strings.HasPrefix(normalized[idx:], "]]") {
			if start == -1 {
				panic("closing span marker without opening marker")
			}

			spans = append(spans, location.Span{
				Start: location.Position(start),
				End:   location.Position(builder.Len()),
			})
			start = -1
			idx++

			continue
		}

		builder.WriteByte(normalized[idx])
	}

	if start != -1 {
		panic("unclosed span marker")
	}

	return markedSource{
		text:  builder.String(),
		spans: spans,
	}
}

func expectDiagnostic(message string, spanIndex int) expectedDiagnostic {
	return expectedDiagnostic{
		message:   message,
		spanIndex: spanIndex,
	}
}

func realizeDiagnostics(source markedSource, diagnostics []expectedDiagnostic) []validator.Diagnostic {
	if len(diagnostics) == 0 {
		return nil
	}

	expected := make([]validator.Diagnostic, 0, len(diagnostics))

	for _, diagnostic := range diagnostics {
		if diagnostic.spanIndex == -1 {
			expected = append(expected, validator.Diagnostic{
				Message: diagnostic.message,
			})

			continue
		}

		if diagnostic.spanIndex < 0 || diagnostic.spanIndex >= len(source.spans) {
			panic(fmt.Sprintf("span index %d out of range", diagnostic.spanIndex))
		}

		expected = append(expected, validator.Diagnostic{
			Message: diagnostic.message,
			Span:    source.spans[diagnostic.spanIndex],
		})
	}

	return expected
}

func scopeHappyPathDSL(size int) string {
	var builder strings.Builder

	builder.WriteString("scope {\n")

	for idx := 0; idx <= size; idx++ {
		builder.WriteString("    include \"src/")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("/**/*.go\"\n")
		builder.WriteString("    exclude \"src/")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("/vendor/**\"\n")
	}

	builder.WriteString("}\n")
	builder.WriteString("tokens {\n")
	builder.WriteString("    Token = \"x\"\n")
	builder.WriteString("}")

	return builder.String()
}

func definitionsHappyPathDSL(size int) string {
	var builder strings.Builder

	builder.WriteString("scope {\n")
	builder.WriteString("    include \"**/*.go\"\n")
	builder.WriteString("}\n")
	builder.WriteString("definitions {\n")

	for idx := 0; idx <= size; idx++ {
		suffix := strconv.Itoa(idx)

		builder.WriteString("    lower")
		builder.WriteString(suffix)
		builder.WriteString(" = 'a'..'z'\n")
		builder.WriteString("    upper")
		builder.WriteString(suffix)
		builder.WriteString(" = 'A'..'Z'\n")
		builder.WriteString("    digit")
		builder.WriteString(suffix)
		builder.WriteString(" = '0'..'9'\n")
		builder.WriteString("    underscore")
		builder.WriteString(suffix)
		builder.WriteString(" = '_'\n")
		builder.WriteString("    letter")
		builder.WriteString(suffix)
		builder.WriteString(" = lower")
		builder.WriteString(suffix)
		builder.WriteString(" | upper")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("    identifierStart")
		builder.WriteString(suffix)
		builder.WriteString(" = letter")
		builder.WriteString(suffix)
		builder.WriteString(" | underscore")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("    identifierPart")
		builder.WriteString(suffix)
		builder.WriteString(" = letter")
		builder.WriteString(suffix)
		builder.WriteString(" | digit")
		builder.WriteString(suffix)
		builder.WriteString(" | underscore")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("    optionalSign")
		builder.WriteString(suffix)
		builder.WriteString(" = ('+' | '-')?\n")
		builder.WriteString("    repeatedLetter")
		builder.WriteString(suffix)
		builder.WriteString(" = letter")
		builder.WriteString(suffix)
		builder.WriteString("*\n")
		builder.WriteString("    value")
		builder.WriteString(suffix)
		builder.WriteString(" = letter")
		builder.WriteString(suffix)
		builder.WriteString(" ('a' | 'b')+\n")
	}

	builder.WriteString("}\n")
	builder.WriteString("tokens {\n")
	builder.WriteString("    Token = \"x\"\n")
	builder.WriteString("}")

	return builder.String()
}

func tokensHappyPathDSL(size int) string {
	var builder strings.Builder

	builder.WriteString("scope {\n")
	builder.WriteString("    include \"**/*.go\"\n")
	builder.WriteString("}\n")
	builder.WriteString("definitions {\n")
	builder.WriteString("    lower = 'a'..'z'\n")
	builder.WriteString("    upper = 'A'..'Z'\n")
	builder.WriteString("    digit = '0'..'9'\n")
	builder.WriteString("    underscore = '_'\n")
	builder.WriteString("    letter = lower | upper\n")
	builder.WriteString("    identifierStart = letter | underscore\n")
	builder.WriteString("    identifierPart = letter | digit | underscore\n")
	builder.WriteString("    optionalSign = ('+' | '-')?\n")
	builder.WriteString("}\n")
	builder.WriteString("tokens {\n")

	for idx := 0; idx <= size; idx++ {
		suffix := strconv.Itoa(idx)

		builder.WriteString("    Plus")
		builder.WriteString(suffix)
		builder.WriteString(" = '+'\n")
		builder.WriteString("    Lower")
		builder.WriteString(suffix)
		builder.WriteString(" = 'a'..'z'\n")
		builder.WriteString("    Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" = identifierStart identifierPart*\n")
		builder.WriteString("    KeywordPublic")
		builder.WriteString(suffix)
		builder.WriteString(" = \"public\"\n")
		builder.WriteString("    SignedInteger")
		builder.WriteString(suffix)
		builder.WriteString(" = optionalSign digit+\n")
		builder.WriteString("    Whitespace")
		builder.WriteString(suffix)
		builder.WriteString(" = (' ' | '\\t' | '\\n' | '\\r')+ skip\n")
	}

	builder.WriteString("}")

	return builder.String()
}

func rulesHappyPathDSL(size int) string {
	var builder strings.Builder

	builder.WriteString("scope {\n")
	builder.WriteString("    include \"**/*.go\"\n")
	builder.WriteString("}\n")
	builder.WriteString("tokens {\n")
	builder.WriteString("    Identifier = \"id\"\n")
	builder.WriteString("}\n")
	builder.WriteString("rules {\n")

	for idx := 0; idx <= size; idx++ {
		suffix := strconv.Itoa(idx)

		builder.WriteString("    rule PublicIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier\n")
		builder.WriteString("        where Identifier.text == \"public\"\n")
		builder.WriteString("        report warn at Identifier \"Public identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule InternalIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier\n")
		builder.WriteString("        where Identifier.text != \"internal\"\n")
		builder.WriteString("        report info at Identifier \"Internal identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule ShortIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier\n")
		builder.WriteString("        where Identifier.length < 3\n")
		builder.WriteString("        report info at Identifier \"Short identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule MediumIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier\n")
		builder.WriteString("        where Identifier.length <= 4\n")
		builder.WriteString("        report warn at Identifier \"Medium identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule LongIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier\n")
		builder.WriteString("        where Identifier.length > 5\n")
		builder.WriteString("        report err at Identifier \"Long identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule VeryLongIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier\n")
		builder.WriteString("        where Identifier.length >= 6\n")
		builder.WriteString("        report err at Identifier \"Very long identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule AnyIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier\n")
		builder.WriteString("        report info at Identifier \"Identifier found\"\n")
		builder.WriteString("    }\n")
	}

	builder.WriteString("}")

	return builder.String()
}

func fullHappyPathValidationDSL(size int) string {
	var builder strings.Builder

	builder.WriteString("scope {\n")
	builder.WriteString("    include \"**/*.go\"\n")
	builder.WriteString("    exclude \"vendor/**\"\n")
	builder.WriteString("}\n")
	builder.WriteString("definitions {\n")

	for idx := 0; idx <= size; idx++ {
		suffix := strconv.Itoa(idx)

		builder.WriteString("    lower")
		builder.WriteString(suffix)
		builder.WriteString(" = 'a'..'z'\n")
		builder.WriteString("    upper")
		builder.WriteString(suffix)
		builder.WriteString(" = 'A'..'Z'\n")
		builder.WriteString("    digit")
		builder.WriteString(suffix)
		builder.WriteString(" = '0'..'9'\n")
		builder.WriteString("    underscore")
		builder.WriteString(suffix)
		builder.WriteString(" = '_'\n")
		builder.WriteString("    letter")
		builder.WriteString(suffix)
		builder.WriteString(" = lower")
		builder.WriteString(suffix)
		builder.WriteString(" | upper")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("    identifierStart")
		builder.WriteString(suffix)
		builder.WriteString(" = letter")
		builder.WriteString(suffix)
		builder.WriteString(" | underscore")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("    identifierPart")
		builder.WriteString(suffix)
		builder.WriteString(" = letter")
		builder.WriteString(suffix)
		builder.WriteString(" | digit")
		builder.WriteString(suffix)
		builder.WriteString(" | underscore")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("    optionalSign")
		builder.WriteString(suffix)
		builder.WriteString(" = ('+' | '-')?\n")
		builder.WriteString("    value")
		builder.WriteString(suffix)
		builder.WriteString(" = letter")
		builder.WriteString(suffix)
		builder.WriteString(" ('a' | 'b')+\n")
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
		builder.WriteString("    KeywordPublic")
		builder.WriteString(suffix)
		builder.WriteString(" = \"public\"\n")
		builder.WriteString("    KeywordInternal")
		builder.WriteString(suffix)
		builder.WriteString(" = \"internal\"\n")
		builder.WriteString("    SignedInteger")
		builder.WriteString(suffix)
		builder.WriteString(" = optionalSign")
		builder.WriteString(suffix)
		builder.WriteString(" digit")
		builder.WriteString(suffix)
		builder.WriteString("+\n")
		builder.WriteString("    Whitespace")
		builder.WriteString(suffix)
		builder.WriteString(" = (' ' | '\\t' | '\\n' | '\\r')+ skip\n")
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
		builder.WriteString("        where Identifier")
		builder.WriteString(suffix)
		builder.WriteString(".text == \"public\"\n")
		builder.WriteString("        report warn at Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" \"Public identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule InternalIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("        where Identifier")
		builder.WriteString(suffix)
		builder.WriteString(".text != \"internal\"\n")
		builder.WriteString("        report info at Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" \"Internal identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule ShortIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("        where Identifier")
		builder.WriteString(suffix)
		builder.WriteString(".length < 3\n")
		builder.WriteString("        report info at Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" \"Short identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule MediumIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("        where Identifier")
		builder.WriteString(suffix)
		builder.WriteString(".length <= 4\n")
		builder.WriteString("        report warn at Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" \"Medium identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule LongIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("        where Identifier")
		builder.WriteString(suffix)
		builder.WriteString(".length > 5\n")
		builder.WriteString("        report err at Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" \"Long identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule VeryLongIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("        where Identifier")
		builder.WriteString(suffix)
		builder.WriteString(".length >= 6\n")
		builder.WriteString("        report err at Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" \"Very long identifier found\"\n")
		builder.WriteString("    }\n")
		builder.WriteString("    rule AnyIdentifier")
		builder.WriteString(suffix)
		builder.WriteString(" {\n")
		builder.WriteString("        match Identifier")
		builder.WriteString(suffix)
		builder.WriteString("\n")
		builder.WriteString("        report info at Identifier")
		builder.WriteString(suffix)
		builder.WriteString(" \"Identifier found\"\n")
		builder.WriteString("    }\n")
	}

	builder.WriteString("}")

	return builder.String()
}
