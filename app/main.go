// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package main implements 'Spot', a high-performance, language-agnostic static analysis engine.
package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kdeconinck/spot/dsl/ast"
	"github.com/kdeconinck/spot/dsl/compiler"
	"github.com/kdeconinck/spot/dsl/parser"
	"github.com/kdeconinck/spot/dsl/resolver"
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/runtime/engine"
	"github.com/kdeconinck/spot/runtime/ir"
	"github.com/kdeconinck/spot/runtime/scanner"
	"github.com/kdeconinck/spot/runtime/syntax"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	options, err := parseCLIOptions(args)

	if err != nil {
		fmt.Fprintln(stderr, err.Error())
		fmt.Fprintln(stderr, "usage: spot [-print-ast] <dsl-file> <directory>")

		return 2
	}

	dslPath := options.dslPath
	rootPath := options.rootPath
	source, err := os.ReadFile(dslPath)

	if err != nil {
		fmt.Fprintf(stderr, "failed to read DSL file %q: %v\n", dslPath, err)

		return 2
	}

	document, parseErr := parser.Parse(string(source))

	if parseErr != nil {
		diagnostic := parseErr.(parser.Diagnostic)
		writeSyntaxDiagnostic(stderr, dslPath, diagnostic.Span.Start, diagnostic.Span.End, diagnostic.Message)

		return 2
	}

	resolution := resolver.Resolve(string(source), document)
	validationDiagnostics := validator.Validate(string(source), resolution)

	if len(validationDiagnostics) != 0 {
		for idx := range validationDiagnostics {
			writeSyntaxDiagnostic(stderr, dslPath, validationDiagnostics[idx].Span.Start, validationDiagnostics[idx].Span.End, validationDiagnostics[idx].Message)
		}

		return 2
	}

	scope, err := compileScope(resolution.ScopeEntries, string(source))

	if err != nil {
		fmt.Fprintf(stderr, "failed to compile scope patterns: %v\n", err)

		return 2
	}

	rootInfo, err := os.Stat(rootPath)

	if err != nil {
		fmt.Fprintf(stderr, "failed to stat analysis directory %q: %v\n", rootPath, err)

		return 2
	}

	if !rootInfo.IsDir() {
		fmt.Fprintf(stderr, "analysis path %q is not a directory\n", rootPath)

		return 2
	}

	program := compiler.Compile(string(source), resolution)
	analysisEngine := engine.New(program)
	diagnosticCount := 0
	walkErr := filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(rootPath, path)

		if err != nil {
			return err
		}

		relativePath = filepath.ToSlash(relativePath)

		if !scope.matches(relativePath) {
			return nil
		}

		content, err := os.ReadFile(path)

		if err != nil {
			return err
		}

		if options.printAST {
			writeRuntimeSyntaxTree(stdout, relativePath, program, string(content))
		}

		diagnostics := analysisEngine.Analyze(program, string(content), engine.Options{})

		for idx := range diagnostics {
			diagnosticCount++
			fmt.Fprintf(stdout, "%s:%d-%d: %s: %s\n",
				relativePath,
				diagnostics[idx].Span.Start,
				diagnostics[idx].Span.End,
				severityName(diagnostics[idx].Severity),
				diagnostics[idx].Message,
			)
		}

		return nil
	})

	if walkErr != nil {
		fmt.Fprintf(stderr, "failed to analyze directory %q: %v\n", rootPath, walkErr)

		return 2
	}

	if diagnosticCount != 0 {
		return 1
	}

	return 0
}

type cliOptions struct {
	dslPath  string
	rootPath string
	printAST bool
}

func parseCLIOptions(args []string) (cliOptions, error) {
	options := cliOptions{}
	paths := make([]string, 0, 2)

	for idx := range args {
		switch args[idx] {
		case "-print-ast", "--print-ast":
			options.printAST = true

		default:
			if strings.HasPrefix(args[idx], "-") {
				return cliOptions{}, fmt.Errorf("unknown flag %q", args[idx])
			}

			paths = append(paths, args[idx])
		}
	}

	if len(paths) != 2 {
		return cliOptions{}, fmt.Errorf("expected DSL file path and analysis directory")
	}

	options.dslPath = paths[0]
	options.rootPath = paths[1]

	return options, nil
}

type scope struct {
	includes []string
	excludes []string
}

func compileScope(entries []ast.ScopeEntry, source string) (scope, error) {
	compiled := scope{
		includes: make([]string, 0, len(entries)),
		excludes: make([]string, 0, len(entries)),
	}

	for idx := range entries {
		pattern, err := strconv.Unquote(entries[idx].Pattern.Value(source))

		if err != nil {
			return scope{}, err
		}

		if entries[idx].Kind == ast.ScopeEntryInclude {
			compiled.includes = append(compiled.includes, pattern)
			continue
		}

		compiled.excludes = append(compiled.excludes, pattern)
	}

	return compiled, nil
}

func (scope scope) matches(name string) bool {
	if !matchesAny(scope.includes, name) {
		return false
	}

	return !matchesAny(scope.excludes, name)
}

func matchesAny(patterns []string, name string) bool {
	for idx := range patterns {
		if matchPattern(patterns[idx], name) {
			return true
		}
	}

	return false
}

func matchPattern(pattern, name string) bool {
	return matchPatternSegments(splitSegments(pattern), splitSegments(name))
}

func splitSegments(value string) []string {
	if value == "" {
		return []string{""}
	}

	return strings.Split(value, "/")
}

func matchPatternSegments(patternSegments, nameSegments []string) bool {
	if len(patternSegments) == 0 {
		return len(nameSegments) == 0
	}

	if patternSegments[0] == "**" {
		if matchPatternSegments(patternSegments[1:], nameSegments) {
			return true
		}

		if len(nameSegments) == 0 {
			return false
		}

		return matchPatternSegments(patternSegments, nameSegments[1:])
	}

	if len(nameSegments) == 0 {
		return false
	}

	if !matchSegment(patternSegments[0], nameSegments[0]) {
		return false
	}

	return matchPatternSegments(patternSegments[1:], nameSegments[1:])
}

func matchSegment(pattern, name string) bool {
	patternIndex := 0
	nameIndex := 0
	star := -1
	match := 0

	for nameIndex < len(name) {
		switch {
		case patternIndex < len(pattern) && pattern[patternIndex] == '?':
			patternIndex++
			nameIndex++

		case patternIndex < len(pattern) && pattern[patternIndex] == '*':
			star = patternIndex
			match = nameIndex
			patternIndex++

		case patternIndex < len(pattern) && pattern[patternIndex] == name[nameIndex]:
			patternIndex++
			nameIndex++

		case star != -1:
			patternIndex = star + 1
			match++
			nameIndex = match

		default:
			return false
		}
	}

	for patternIndex < len(pattern) && pattern[patternIndex] == '*' {
		patternIndex++
	}

	return patternIndex == len(pattern)
}

func writeSyntaxDiagnostic(writer io.Writer, path string, start, end location.Position, message string) {
	fmt.Fprintf(writer, "%s:%d-%d: error: %s\n", path, start, end, message)
}

func severityName(severity engine.Severity) string {
	switch severity {
	case engine.SeverityInfo:
		return "info"

	case engine.SeverityWarn:
		return "warn"

	default:
		return "err"
	}
}

func writeRuntimeSyntaxTree(stdout io.Writer, relativePath string, program ir.Program, src string) {
	if program.SyntaxRoot < 0 || len(program.SyntaxNodes) == 0 {
		return
	}

	tokens, ok := scanSource(program, src)

	if !ok {
		return
	}

	syntaxParser, err := syntax.New(program, program.SyntaxNodes[program.SyntaxRoot].Name)

	if err != nil {
		return
	}

	var tree syntax.Tree

	if !syntaxParser.ParseInto(tokens, &tree) {
		return
	}

	fmt.Fprintf(stdout, "%s\n%s\n", relativePath, renderRuntimeSyntaxTree(program, tree))
}

func scanSource(program ir.Program, src string) ([]scanner.Token, bool) {
	scan := scanner.New(program, src)
	tokens := make([]scanner.Token, 0, len(src))

	for {
		token, diagnostic, ok := scan.Next()

		if !ok {
			return tokens, true
		}

		if diagnostic.Message != "" {
			return nil, false
		}

		tokens = append(tokens, token)
	}
}

func renderRuntimeSyntaxTree(program ir.Program, tree syntax.Tree) string {
	var builder strings.Builder

	builder.WriteString("Tree\n")
	appendRuntimeSyntaxNode(&builder, program, tree, tree.Root, 1)

	return strings.TrimSpace(builder.String())
}

func appendRuntimeSyntaxNode(builder *strings.Builder, program ir.Program, tree syntax.Tree, nodeID syntax.NodeID, depth int) {
	node := tree.Node(nodeID)
	end := node.FirstTokenIndex + node.AmountOfTokens

	appendIndentedLine(builder, depth, "Node "+program.SyntaxNodes[node.Kind].Name+" ["+strconv.Itoa(int(node.FirstTokenIndex))+":"+strconv.Itoa(int(end))+"]")

	for _, childID := range tree.Children(node) {
		appendRuntimeSyntaxNode(builder, program, tree, childID, depth+1)
	}
}

func appendIndentedLine(builder *strings.Builder, depth int, text string) {
	builder.WriteString(strings.Repeat("  ", depth))
	builder.WriteString(text)
	builder.WriteByte('\n')
}
