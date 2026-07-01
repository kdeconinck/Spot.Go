// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Verify the public API of the syntax package.
//
// Tests in this package are written against the exported API only.
// This ensures that behavior is tested through the same surface that external consumers would use.
package syntax_test

import (
	"strconv"
	"strings"
	"testing"

	"github.com/kdeconinck/spot/dsl/compiler"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/qa/claim"
	"github.com/kdeconinck/spot/runtime/ir"
	"github.com/kdeconinck/spot/runtime/scanner"
	"github.com/kdeconinck/spot/runtime/syntax"
)

func Test_New(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inRootNode string
		wantErr    string
	}{
		"When the root syntax node exists, no error is returned.": {
			inRootNode: "Root",
			wantErr:    "",
		},
		"When the root syntax node does not exist, an error is returned.": {
			inRootNode: "Missing",
			wantErr:    `syntax node "Missing" is not declared`,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			program := compileProgram(t, rootSyntaxDSL())

			// Act.
			_, gotErr := syntax.New(program, tc.inRootNode)

			// Assert.
			claim.Equal(t, tcName, tc.wantErr, errorText(gotErr), "Error")
		})
	}
}

func Test_Parse(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inDSL      string
		inRootNode string
		inSource   string
		wantOK     bool
		wantTree   string
	}{
		"When parsing a full syntax tree, a flat tree is returned.": {
			inDSL:      rootSyntaxDSL(),
			inRootNode: "Root",
			inSource:   "internal public id public id id internal",
			wantOK:     true,
			wantTree: normalizeMultilineLiteral(`
				Tree
				  Node Root [0:7]
				    optional: Node OptionalWord [0:1]
				    pair: Node WordPair [1:3]
				      left: Node Word [1:2]
				      right: Node Word [2:3]
				    list: Node WordList [3:6]
				      values: Node Word [3:4]
				      values: Node Word [4:5]
				      values: Node Word [5:6]
				    unknown: Node UnknownStatement [6:7]
				    tail: Node WordTail [7:7]
			`),
		},
		"When the root node does not consume the entire token slice, parsing fails.": {
			inDSL:      rootSyntaxDSL(),
			inRootNode: "WordPair",
			inSource:   "public id id",
			wantOK:     false,
		},
		"When the token stream does not match the syntax root, parsing fails.": {
			inDSL:      rootSyntaxDSL(),
			inRootNode: "Root",
			inSource:   "public internal",
			wantOK:     false,
		},
		"When parsing a partially modeled file with any, parsing succeeds.": {
			inDSL: strings.Join([]string{
				`scope { include "**/*.cs" }`,
				`definitions {`,
				`    lower = 'a'..'z'`,
				`    upper = 'A'..'Z'`,
				`    digit = '0'..'9'`,
				`    identifierStart = lower | upper | '_'`,
				`    identifierPart = identifierStart | digit`,
				`}`,
				`tokens {`,
				`    Whitespace = (' ' | '\t' | '\n' | '\r')+ skip`,
				`    KeywordUsing = "using"`,
				`    KeywordNamespace = "namespace"`,
				`    Identifier = identifierStart identifierPart*`,
				`    Dot = "."`,
				`    Semicolon = ";"`,
				`    LeftBrace = "{"`,
				`    RightBrace = "}"`,
				`    Unknown = fallback`,
				`}`,
				`syntax {`,
				`    node QualifiedIdentifierTail { Dot Identifier }`,
				`    node QualifiedIdentifier { head: Identifier tail*: QualifiedIdentifierTail }`,
				`    node UsingDirective { KeywordUsing name: QualifiedIdentifier Semicolon }`,
				`    node NamespaceBody { LeftBrace members*: oneOf { UsingDirective any } RightBrace }`,
				`    node NamespaceDeclaration { KeywordNamespace name: QualifiedIdentifier body: NamespaceBody }`,
				`    node Root { members*: oneOf { UsingDirective NamespaceDeclaration any } }`,
				`}`,
			}, "\n"),
			inRootNode: "Root",
			inSource: strings.Join([]string{
				`using System;`,
				``,
				`namespace Example {`,
				`    using System.Text;`,
				``,
				`    public static void Main(string[] args) {`,
				`        Console.WriteLine("Hello, World!");`,
				`    }`,
				`}`,
			}, "\n"),
			wantOK: true,
		},
		"When parsing named captures, the tree preserves field labels on child edges.": {
			inDSL: strings.Join([]string{
				`scope { include "**/*.go" }`,
				`tokens { Identifier = "id" Whitespace = ' '+ skip }`,
				`syntax {`,
				`    node QualifiedIdentifier { Identifier }`,
				`    node UsingDirective { name: QualifiedIdentifier }`,
				`    node Root { values: UsingDirective+ }`,
				`}`,
			}, "\n"),
			inRootNode: "Root",
			inSource:   "id id",
			wantOK:     true,
			wantTree: normalizeMultilineLiteral(`
				Tree
				  Node Root [0:2]
				    values: Node UsingDirective [0:1]
				      name: Node QualifiedIdentifier [0:1]
				    values: Node UsingDirective [1:2]
				      name: Node QualifiedIdentifier [1:2]
			`),
		},
		"When parsing structured syntax nodes, the tree preserves named fields.": {
			inDSL: strings.Join([]string{
				`scope { include "**/*.go" }`,
				`tokens { Identifier = "id" KeywordUsing = "using" Semicolon = ";" Whitespace = ' '+ skip }`,
				`syntax {`,
				`    node IdentifierName {`,
				`        Identifier`,
				`    }`,
				`    node QualifiedIdentifier {`,
				`        head: IdentifierName`,
				`    }`,
				`    node UsingDirective {`,
				`        KeywordUsing`,
				`        name: QualifiedIdentifier`,
				`        Semicolon`,
				`    }`,
				`    node Root {`,
				`        members*: UsingDirective`,
				`    }`,
				`}`,
			}, "\n"),
			inRootNode: "Root",
			inSource:   "using id;",
			wantOK:     true,
			wantTree: normalizeMultilineLiteral(`
				Tree
				  Node Root [0:3]
				    members: Node UsingDirective [0:3]
				      name: Node QualifiedIdentifier [1:2]
				        head: Node IdentifierName [1:2]
			`),
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			program := compileProgram(t, tc.inDSL)
			syntaxParser, parserErr := syntax.New(program, tc.inRootNode)
			tokens := scanTokens(t, program, tc.inSource)

			// Act.
			gotTree, gotOK := syntaxParser.Parse(tokens)

			// Assert.
			claim.Equal(t, tcName, error(nil), parserErr, "Parser Error")
			claim.Equal(t, tcName, tc.wantOK, gotOK, "OK")

			if tc.wantTree != "" {
				claim.Equal(t, tcName, tc.wantTree, renderTree(program, gotTree), "Tree")
			}
		})
	}
}

func Test_ParseInto(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inDSL      string
		inRootNode string
		inSource   string
		wantOK     bool
		wantTree   string
	}{
		"When parsing a full syntax tree into a reused buffer, a flat tree is returned.": {
			inDSL:      rootSyntaxDSL(),
			inRootNode: "Root",
			inSource:   "internal public id public id id internal",
			wantOK:     true,
			wantTree: normalizeMultilineLiteral(`
				Tree
				  Node Root [0:7]
				    optional: Node OptionalWord [0:1]
				    pair: Node WordPair [1:3]
				      left: Node Word [1:2]
				      right: Node Word [2:3]
				    list: Node WordList [3:6]
				      values: Node Word [3:4]
				      values: Node Word [4:5]
				      values: Node Word [5:6]
				    unknown: Node UnknownStatement [6:7]
				    tail: Node WordTail [7:7]
			`),
		},
		"When parsing fails, the reused buffer is cleared.": {
			inDSL:      rootSyntaxDSL(),
			inRootNode: "Root",
			inSource:   "public internal",
			wantOK:     false,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Arrange.
			program := compileProgram(t, tc.inDSL)
			syntaxParser, parserErr := syntax.New(program, tc.inRootNode)
			tokens := scanTokens(t, program, tc.inSource)
			tree := syntax.Tree{
				Nodes:      make([]syntax.Node, 0, 64),
				ChildEdges: make([]syntax.ChildEdge, 0, 64),
			}

			// Act.
			gotOK := syntaxParser.ParseInto(tokens, &tree)

			// Assert.
			claim.Equal(t, tcName, error(nil), parserErr, "Parser Error")
			claim.Equal(t, tcName, tc.wantOK, gotOK, "OK")

			if tc.wantTree != "" {
				claim.Equal(t, tcName, tc.wantTree, renderTree(program, tree), "Tree")
			}

			if !tc.wantOK {
				claim.Equal(t, tcName, 0, len(tree.Nodes), "Node Count")
				claim.Equal(t, tcName, 0, len(tree.ChildEdges), "Child Count")
			}
		})
	}
}

func Benchmark_Parse_Syntax_0(b *testing.B)    { benchmark_Parse_Syntax(b, 0) }
func Benchmark_Parse_Syntax_1(b *testing.B)    { benchmark_Parse_Syntax(b, 1) }
func Benchmark_Parse_Syntax_10(b *testing.B)   { benchmark_Parse_Syntax(b, 10) }
func Benchmark_Parse_Syntax_100(b *testing.B)  { benchmark_Parse_Syntax(b, 100) }
func Benchmark_Parse_Syntax_1000(b *testing.B) { benchmark_Parse_Syntax(b, 1000) }

func Benchmark_ParseInto_Syntax_0(b *testing.B)    { benchmark_ParseInto_Syntax(b, 0) }
func Benchmark_ParseInto_Syntax_1(b *testing.B)    { benchmark_ParseInto_Syntax(b, 1) }
func Benchmark_ParseInto_Syntax_10(b *testing.B)   { benchmark_ParseInto_Syntax(b, 10) }
func Benchmark_ParseInto_Syntax_100(b *testing.B)  { benchmark_ParseInto_Syntax(b, 100) }
func Benchmark_ParseInto_Syntax_1000(b *testing.B) { benchmark_ParseInto_Syntax(b, 1000) }

func benchmark_Parse_Syntax(b *testing.B, size int) {
	b.Helper()

	program := compileProgram(b, syntaxDSL(size))
	syntaxParser, parserErr := syntax.New(program, "Root")
	claim.Equal(b, "Syntax benchmark parser error.", error(nil), parserErr, "Parser Error")
	tokens := scanTokens(b, program, syntaxSource(size))

	for b.Loop() {
		_, _ = syntaxParser.Parse(tokens)
	}
}

func benchmark_ParseInto_Syntax(b *testing.B, size int) {
	b.Helper()

	program := compileProgram(b, syntaxDSL(size))
	syntaxParser, parserErr := syntax.New(program, "Root")
	claim.Equal(b, "Syntax benchmark parser error.", error(nil), parserErr, "Parser Error")
	tokens := scanTokens(b, program, syntaxSource(size))
	tree := syntax.Tree{
		Nodes:      make([]syntax.Node, 0, len(tokens)),
		ChildEdges: make([]syntax.ChildEdge, 0, len(tokens)),
	}

	for b.Loop() {
		_ = syntaxParser.ParseInto(tokens, &tree)
	}
}

func compileProgram(tb testing.TB, source string) ir.Program {
	tb.Helper()

	document, parseErr := parser.Parse(source)
	claim.Equal(tb, "Compile program parse error.", error(nil), parseErr, "Parse Error")
	resolution := resolver.Resolve(source, document)
	validationDiagnostics := validator.Validate(source, resolution)
	claim.Equal(tb, "Compile program validation diagnostics.", 0, len(validationDiagnostics), "Validation Diagnostic Count")

	return compiler.Compile(source, resolution)
}

func scanTokens(tb testing.TB, program ir.Program, source string) []scanner.Token {
	tb.Helper()

	scan := scanner.New(program, source)
	tokens := make([]scanner.Token, 0, len(source))

	for {
		token, diagnostic, ok := scan.Next()

		if !ok {
			return tokens
		}

		claim.Equal(tb, "Scan diagnostic.", "", diagnostic.Message, "Diagnostic")
		tokens = append(tokens, token)
	}
}

func renderTree(program ir.Program, tree syntax.Tree) string {
	var builder strings.Builder

	builder.WriteString("Tree\n")
	appendNode(&builder, program, tree, tree.Root, 1)

	return strings.TrimSpace(builder.String())
}

func appendNode(builder *strings.Builder, program ir.Program, tree syntax.Tree, nodeID syntax.NodeID, depth int) {
	node := tree.Node(nodeID)
	end := node.FirstTokenIndex + node.AmountOfTokens

	appendIndentedLine(
		builder,
		depth,
		"Node "+program.SyntaxNodes[node.Kind].Name+" ["+strconv.Itoa(int(node.FirstTokenIndex))+":"+strconv.Itoa(int(end))+"]",
	)

	for _, childEdge := range tree.Children(node) {
		appendTreeEdge(builder, program, tree, childEdge, depth+1)
	}
}

func appendTreeEdge(builder *strings.Builder, program ir.Program, tree syntax.Tree, edge syntax.ChildEdge, depth int) {
	if edge.FieldID == ^uint32(0) {
		appendNode(builder, program, tree, edge.ChildID, depth)

		return
	}

	node := tree.Node(edge.ChildID)
	end := node.FirstTokenIndex + node.AmountOfTokens
	appendIndentedLine(
		builder,
		depth,
		program.SyntaxFields[edge.FieldID]+": Node "+program.SyntaxNodes[node.Kind].Name+" ["+strconv.Itoa(int(node.FirstTokenIndex))+":"+strconv.Itoa(int(end))+"]",
	)

	for _, childEdge := range tree.Children(node) {
		appendTreeEdge(builder, program, tree, childEdge, depth+1)
	}
}

func appendIndentedLine(builder *strings.Builder, depth int, text string) {
	builder.WriteString(strings.Repeat("  ", depth))
	builder.WriteString(text)
	builder.WriteByte('\n')
}

func errorText(err error) string {
	if err == nil {
		return ""
	}

	return err.Error()
}

func normalizeMultilineLiteral(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	for idx := range lines {
		lines[idx] = strings.TrimRight(strings.TrimLeft(lines[idx], "\t"), " ")
	}

	return strings.Join(lines, "\n")
}

func rootSyntaxDSL() string {
	return strings.Join([]string{
		`scope { include "**/*.go" }`,
		`tokens {`,
		`    Identifier = "id"`,
		`    KeywordPublic = "public"`,
		`    KeywordInternal = "internal"`,
		`    Whitespace = ' '+ skip`,
		`}`,
		`syntax {`,
		`    node Word { oneOf { Identifier KeywordPublic } }`,
		`    node WordPair { left: Word right: Word }`,
		`    node OptionalWord { value?: oneOf { Word KeywordInternal } }`,
		`    node UnknownStatement { values: any+ }`,
		`    node WordTail { values*: Word }`,
		`    node WordList { values: Word+ }`,
		`    node Root { optional: OptionalWord pair: WordPair list: WordList unknown?: UnknownStatement tail: WordTail }`,
		`}`,
	}, "\n")
}

func syntaxDSL(size int) string {
	var builder strings.Builder

	builder.WriteString("scope { include \"**/*.go\" }\n")
	builder.WriteString("tokens {\n")
	builder.WriteString("    Identifier = \"id\"\n")
	builder.WriteString("    KeywordPublic = \"public\"\n")
	builder.WriteString("    KeywordInternal = \"internal\"\n")
	builder.WriteString("    Whitespace = ' '+ skip\n")
	builder.WriteString("}\n")
	builder.WriteString("syntax {\n")
	builder.WriteString("    node Word { oneOf { Identifier KeywordPublic } }\n")
	builder.WriteString("    node WordPair { left: Word right: Word }\n")
	builder.WriteString("    node OptionalWord { value?: oneOf { Word KeywordInternal } }\n")
	builder.WriteString("    node UnknownStatement { values: any+ }\n")
	builder.WriteString("    node WordTail { values*: Word }\n")
	builder.WriteString("    node WordList { values: Word+ }\n")
	builder.WriteString("    node Chunk { optional: OptionalWord pair: WordPair list: WordList unknown?: UnknownStatement tail: WordTail }\n")
	builder.WriteString("    node Root { values*: Chunk }\n")

	for idx := 1; idx <= size; idx++ {
		builder.WriteString("    node Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" { oneOf { Identifier KeywordPublic } }\n")
		builder.WriteString("    node WordPair")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" { left: Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" right: Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" }\n")
		builder.WriteString("    node OptionalWord")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" { value?: oneOf { Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" KeywordInternal } }\n")
		builder.WriteString("    node UnknownStatement")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" { values: any+ }\n")
		builder.WriteString("    node WordTail")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" { values*: Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" }\n")
		builder.WriteString("    node WordList")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" { values: Word")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString("+ }\n")
		builder.WriteString("    node Chunk")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" { optional: OptionalWord")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" pair: WordPair")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" list: WordList")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" unknown?: UnknownStatement")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" tail: WordTail")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" }\n")
		builder.WriteString("    node Root")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" { values*: Chunk")
		builder.WriteString(strconv.Itoa(idx))
		builder.WriteString(" }\n")
	}

	builder.WriteString("}")

	return builder.String()
}

func syntaxSource(size int) string {
	parts := make([]string, 0, 7+size*7)
	parts = append(parts, "internal", "public", "id", "public", "id", "id", "internal")

	for range size {
		parts = append(parts, "internal", "public", "id", "public", "id", "id", "internal")
	}

	return strings.Join(parts, " ")
}
