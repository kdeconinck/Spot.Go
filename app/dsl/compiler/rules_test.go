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
	"testing"

	"github.com/kdeconinck/spot/dsl/compiler"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
)

func Test_Compile_Rules(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inSource    string
		wantProgram string
	}{
		"When compiling rules, source order is preserved.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" Number = ('0'..'9')+ } rules {
		rule First { match Identifier report info at Identifier "first" }
		rule Second { match Number where Number.length >= 2 report err at Number "second" }
	}`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				    Token Number
				      Repetition +
				        Range '0' '9'
				  Rules
				    Rule First
				      MatchToken Identifier
				      Where none
				      Report info at Identifier "first"
				    Rule Second
				      MatchToken Number
				      Where length >= 2
				      Report err at Number "second"
			`),
		},
		"When compiling a text inequality rule, the text condition is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule NotPublic { match Identifier where Identifier.text != "public" report warn at Identifier "message" } }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Rules
				    Rule NotPublic
				      MatchToken Identifier
				      Where text != "public"
				      Report warn at Identifier "message"
			`),
		},
		"When compiling a length less-than rule, the numeric condition is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule Short { match Identifier where Identifier.length < 10 report warn at Identifier "message" } }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Rules
				    Rule Short
				      MatchToken Identifier
				      Where length < 10
				      Report warn at Identifier "message"
			`),
		},
		"When compiling a length less-than-or-equal rule, the numeric condition is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule ShortOrEqual { match Identifier where Identifier.length <= 10 report warn at Identifier "message" } }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Rules
				    Rule ShortOrEqual
				      MatchToken Identifier
				      Where length <= 10
				      Report warn at Identifier "message"
			`),
		},
		"When compiling a length greater-than rule, the numeric condition is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } rules { rule Long { match Identifier where Identifier.length > 10 report warn at Identifier "message" } }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Rules
				    Rule Long
				      MatchToken Identifier
				      Where length > 10
				      Report warn at Identifier "message"
			`),
		},
		"When compiling a syntax-node rule, the node match is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } syntax { node Word { Identifier } node Root { values: Word+ } } rules { rule RootRule { match node Root where Root.length > 0 report warn at Root "message" } }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Syntax
				    Node Word
				      Token Identifier
				    Node Root
				      Capture values
				        Repetition +
				          Node Word
				  Rules
				    Rule RootRule
				      MatchNode Root
				      Where length > 0
				      Report warn at Root "message"
			`),
		},
		"When compiling a syntax-node ancestor constraint, the scope is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } syntax { node Using { Identifier } node Namespace { Using } node Root { Namespace } } rules { rule UsingOutsideNamespace { match node Using outside Namespace report warn at Using "message" } }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Syntax
				    Node Using
				      Token Identifier
				    Node Namespace
				      Node Using
				    Node Root
				      Node Namespace
				  Rules
				    Rule UsingOutsideNamespace
				      MatchNode Using outside Namespace
				      Where none
				      Report warn at Using "message"
			`),
		},
		"When compiling a selector rule with a direct parent query, the scope is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } syntax { node Using { Identifier } node Namespace { Using } node Root { Namespace } } rules { info "message" : Namespace > Using }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Syntax
				    Node Using
				      Token Identifier
				    Node Namespace
				      Node Using
				    Node Root
				      Node Namespace
				  Rules
				    Rule
				      MatchNode Using parent Namespace
				      Where none
				      Report info at Using "message"
			`),
		},
		"When compiling a selector rule with a negated direct parent query, the scope is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } syntax { node Using { Identifier } node Namespace { Using } node Root { Namespace } } rules { warn "message" : Using:not(Namespace > *) }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Syntax
				    Node Using
				      Token Identifier
				    Node Namespace
				      Node Using
				    Node Root
				      Node Namespace
				  Rules
				    Rule
				      MatchNode Using outside parent Namespace
				      Where none
				      Report warn at Using "message"
			`),
		},
		"When compiling a selector rule with an adjacent-sibling condition, the comparison is compiled.": {
			inSource: `scope { include "**/*.go" } tokens { Identifier = "id" } syntax { node Using { Identifier } node Root { values: Using+ } } rules { warn "message" : Using + Using where left.text > right.text }`,
			wantProgram: normalizeMultilineLiteral(`
				Program
				  Tokens
				    Token Identifier
				      String "id"
				  Syntax
				    Node Using
				      Token Identifier
				    Node Root
				      Capture values
				        Repetition +
				          Node Using
				  Rules
				    Rule
				      Match Using + Using
				      Where left.text > text
				      Report warn at Using "message"
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
