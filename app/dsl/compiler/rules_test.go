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
				      MatchToken 0
				      Where none
				      Report info at 0 "first"
				    Rule Second
				      MatchToken 1
				      Where length >= 2
				      Report err at 1 "second"
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
				      MatchToken 0
				      Where text != "public"
				      Report warn at 0 "message"
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
				      MatchToken 0
				      Where length < 10
				      Report warn at 0 "message"
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
				      MatchToken 0
				      Where length <= 10
				      Report warn at 0 "message"
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
				      MatchToken 0
				      Where length > 10
				      Report warn at 0 "message"
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
