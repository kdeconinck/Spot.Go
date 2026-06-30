// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the engine package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package engine_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/compiler"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/runtime/engine"
	"github.com/kdeconinck/spot/runtime/ir"
)

func Test_Engine_Analyze(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inDSL           string
		inSource        string
		inOptions       engine.Options
		wantDiagnostics []engine.Diagnostic
	}{
		"When a rule matches token text, a warning diagnostic is returned.": {
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
}
rules {
    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warn at Identifier "Public identifier found"
    }
}`,
			inSource: "private public internal",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityWarn, "Public identifier found", 8, 14),
			},
		},
		"When a rule matches token length, an error diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
definitions {
    digit = '0'..'9'
}
tokens {
    Whitespace = ' '+ skip
    Number = digit+
}
rules {
    rule LargeNumber {
        match Number
        where Number.length >= 3
        report err at Number "Number too large"
    }
}`,
			inSource: "9 42 123 7",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityErr, "Number too large", 5, 8),
			},
		},
		"When a text inequality rule matches, an informational diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
definitions {
    letter = 'a'..'z'
}
tokens {
    Whitespace = ' '+ skip
    Identifier = letter+
}
rules {
    rule NonPublicIdentifier {
        match Identifier
        where Identifier.text != "public"
        report info at Identifier "non-public"
    }
}`,
			inSource: "public private",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityInfo, "non-public", 7, 14),
			},
		},
		"When a length equality rule matches, an informational diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
definitions {
    digit = '0'..'9'
}
tokens {
    Whitespace = ' '+ skip
    Number = digit+
}
rules {
    rule ExactLength {
        match Number
        where Number.length == 2
        report info at Number "exact"
    }
}`,
			inSource: "1 22 333",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityInfo, "exact", 2, 4),
			},
		},
		"When a length inequality rule matches, an informational diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
definitions {
    digit = '0'..'9'
}
tokens {
    Whitespace = ' '+ skip
    Number = digit+
}
rules {
    rule NotLengthTwo {
        match Number
        where Number.length != 2
        report info at Number "not-two"
    }
}`,
			inSource: "1 22",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityInfo, "not-two", 0, 1),
			},
		},
		"When a length less-than rule matches, an informational diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
definitions {
    digit = '0'..'9'
}
tokens {
    Whitespace = ' '+ skip
    Number = digit+
}
rules {
    rule Short {
        match Number
        where Number.length < 3
        report info at Number "short"
    }
}`,
			inSource: "1 222",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityInfo, "short", 0, 1),
			},
		},
		"When a length less-than-or-equal rule matches, an informational diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
definitions {
    digit = '0'..'9'
}
tokens {
    Whitespace = ' '+ skip
    Number = digit+
}
rules {
    rule ShortOrEqual {
        match Number
        where Number.length <= 2
        report info at Number "short-or-equal"
    }
}`,
			inSource: "1 22 333",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityInfo, "short-or-equal", 0, 1),
				diagnostic(engine.SeverityInfo, "short-or-equal", 2, 4),
			},
		},
		"When a length greater-than rule matches, an informational diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
definitions {
    digit = '0'..'9'
}
tokens {
    Whitespace = ' '+ skip
    Number = digit+
}
rules {
    rule Long {
        match Number
        where Number.length > 2
        report info at Number "long"
    }
}`,
			inSource: "1 222",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityInfo, "long", 2, 5),
			},
		},
		"When multiple rules match and early exit is enabled, only the first diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
definitions {
    letter = 'a'..'z'
}
tokens {
    Whitespace = ' '+ skip
    Identifier = letter+
}
rules {
    rule Foo {
        match Identifier
        where Identifier.text == "foo"
        report warn at Identifier "foo"
    }
    rule Bar {
        match Identifier
        where Identifier.text == "bar"
        report warn at Identifier "bar"
    }
}`,
			inSource: "foo bar",
			inOptions: engine.Options{
				StopOnFirstDiagnostic: true,
			},
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityWarn, "foo", 0, 3),
			},
		},
		"When scanning fails, an error diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
tokens {
    Identifier = "a"
}
rules {
    rule IdentifierRule {
        match Identifier
        report info at Identifier "identifier"
    }
}`,
			inSource: "ab",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityInfo, "identifier", 0, 1),
				diagnostic(engine.SeverityErr, "No token matched at byte offset 1.", 1, 2),
			},
		},
		"When a syntax-node rule matches, a diagnostic is returned for the matched node span.": {
			inDSL: `scope { include "**/*.go" }
tokens {
    Whitespace = ' '+ skip
    KeywordPackage = "package"
    Identifier = "main"
}
syntax {
    node PackageClause = KeywordPackage Identifier
    node Root = PackageClause
}
rules {
    rule PackageRule {
        match node PackageClause
        where PackageClause.text == "package main"
        report warn at PackageClause "package clause found"
    }
}`,
			inSource: "package main",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityWarn, "package clause found", 0, 12),
			},
		},
		"When a syntax-node inside constraint matches, a diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
tokens {
    Whitespace = ' '+ skip
    KeywordNamespace = "namespace"
    KeywordUsing = "using"
    Identifier = "x"
    LeftBrace = "{"
    RightBrace = "}"
}
syntax {
    node UsingDirective = KeywordUsing Identifier
    node NamespaceBody = LeftBrace UsingDirective RightBrace
    node NamespaceDeclaration = KeywordNamespace Identifier NamespaceBody
    node Root = NamespaceDeclaration
}
rules {
    rule UsingInsideNamespace {
        match node UsingDirective inside NamespaceDeclaration
        report warn at UsingDirective "using inside namespace"
    }
}`,
			inSource: "namespace x { using x }",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityWarn, "using inside namespace", 14, 21),
			},
		},
		"When a syntax-node outside constraint matches, a diagnostic is returned.": {
			inDSL: `scope { include "**/*.go" }
tokens {
    Whitespace = ' '+ skip
    KeywordNamespace = "namespace"
    KeywordUsing = "using"
    Identifier = "x"
    LeftBrace = "{"
    RightBrace = "}"
}
syntax {
    node UsingDirective = KeywordUsing Identifier
    node NamespaceBody = LeftBrace RightBrace
    node NamespaceDeclaration = KeywordNamespace Identifier NamespaceBody
    node Root = UsingDirective NamespaceDeclaration
}
rules {
    rule UsingOutsideNamespace {
        match node UsingDirective outside NamespaceDeclaration
        report warn at UsingDirective "using outside namespace"
    }
}`,
			inSource: "using x namespace x { }",
			wantDiagnostics: []engine.Diagnostic{
				diagnostic(engine.SeverityWarn, "using outside namespace", 0, 7),
			},
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			program := compileProgram(t, tc.inDSL)
			analysisEngine := engine.New(program)

			// Act.
			gotDiagnostics := analysisEngine.Analyze(program, tc.inSource, tc.inOptions)

			// Assert.
			claim.DeepEqual(t, tcName, tc.wantDiagnostics, gotDiagnostics, "Diagnostic")
		})
	}
}

func Benchmark_Engine_Analyze_0(b *testing.B)    { benchmark_Engine_Analyze(b, 0) }
func Benchmark_Engine_Analyze_1(b *testing.B)    { benchmark_Engine_Analyze(b, 1) }
func Benchmark_Engine_Analyze_10(b *testing.B)   { benchmark_Engine_Analyze(b, 10) }
func Benchmark_Engine_Analyze_100(b *testing.B)  { benchmark_Engine_Analyze(b, 100) }
func Benchmark_Engine_Analyze_1000(b *testing.B) { benchmark_Engine_Analyze(b, 1000) }

func benchmark_Engine_Analyze(b *testing.B, size int) {
	b.Helper()

	program := compileProgram(b, engineDSL())
	analysisEngine := engine.New(program)
	source := engineInput(size)

	for b.Loop() {
		_ = analysisEngine.Analyze(program, source, engine.Options{})
	}
}

func compileProgram(tb testing.TB, source string) ir.Program {
	tb.Helper()

	document, parseErr := parser.Parse(source)
	resolution := resolver.Resolve(source, document)
	validationDiagnostics := validator.Validate(source, resolution)

	if parseErr != nil {
		tb.Fatalf("engine test parse error: got %v, want nil", parseErr)
	}

	if len(validationDiagnostics) != 0 {
		tb.Fatalf("engine test validation diagnostics: got %d, want 0", len(validationDiagnostics))
	}

	return compiler.Compile(source, resolution)
}

func engineDSL() string {
	return `scope { include "**/*.go" }
definitions {
    letter = 'a'..'z' | 'A'..'Z'
    digit = '0'..'9'
    identifierStart = letter | '_'
    identifierPart = identifierStart | digit
}
tokens {
    Whitespace = (' ' | '\t' | '\n')+ skip
    Identifier = identifierStart identifierPart*
    Number = digit+
}
syntax {
    node NumberNode = Number
    node IdentifierNode = Identifier
    node Statement = IdentifierNode NumberNode
    node Root = Statement+
}
rules {
    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warn at Identifier "Public identifier found"
    }
    rule LargeNumber {
        match Number
        where Number.length >= 3
        report err at Number "Number too large"
    }
    rule StatementRule {
        match node Statement
        where Statement.length > 0
        report info at Statement "statement"
    }
}`
}

func engineInput(size int) string {
	var sb strings.Builder

	sb.WriteString("public 123")

	for idx := 1; idx <= size; idx++ {
		sb.WriteByte(' ')
		sb.WriteString("name")
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteByte(' ')
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(strconv.Itoa(idx))
		sb.WriteString(strconv.Itoa(idx))
	}

	return sb.String()
}

func diagnostic(severity engine.Severity, message string, start, end location.Position) engine.Diagnostic {
	return engine.Diagnostic{
		Severity: severity,
		Message:  message,
		Span: location.Span{
			Start: start,
			End:   end,
		},
	}
}
