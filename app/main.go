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
	"github.com/kdeconinck/spot/dsl/validator"
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/runtime/engine"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		fmt.Fprintln(stderr, "usage: spot <dsl-file> <directory>")

		return 2
	}

	dslPath := args[0]
	rootPath := args[1]
	source, err := os.ReadFile(dslPath)

	if err != nil {
		fmt.Fprintf(stderr, "failed to read DSL file %q: %v\n", dslPath, err)

		return 2
	}

	document, parseDiagnostics := parser.Parse(string(source))

	if len(parseDiagnostics) != 0 {
		for idx := range parseDiagnostics {
			writeSyntaxDiagnostic(stderr, dslPath, parseDiagnostics[idx].Span.Start, parseDiagnostics[idx].Span.End, parseDiagnostics[idx].Message)
		}

		return 2
	}

	validationDiagnostics := validator.Validate(string(source), document)

	if len(validationDiagnostics) != 0 {
		for idx := range validationDiagnostics {
			writeSyntaxDiagnostic(stderr, dslPath, validationDiagnostics[idx].Span.Start, validationDiagnostics[idx].Span.End, validationDiagnostics[idx].Message)
		}

		return 2
	}

	scope, err := compileScope(document.Scope, string(source))

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

	program := compiler.Compile(string(source), document)
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

type scope struct {
	includes []string
	excludes []string
}

func compileScope(section ast.ScopeSection, source string) (scope, error) {
	compiled := scope{
		includes: make([]string, 0, len(section.Entries)),
		excludes: make([]string, 0, len(section.Entries)),
	}

	for idx := range section.Entries {
		pattern, err := strconv.Unquote(section.Entries[idx].Pattern.Value(source))

		if err != nil {
			return scope{}, err
		}

		if section.Entries[idx].Kind == ast.ScopeEntryInclude {
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
