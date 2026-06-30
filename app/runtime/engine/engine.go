// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package engine evaluates compiled Spot rules over scanned source text.
package engine

import (
	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/runtime/ir"
	"github.com/kdeconinck/spot/runtime/scanner"
	"github.com/kdeconinck/spot/runtime/syntax"
)

// Severity identifies the severity of a runtime diagnostic.
type Severity uint8

const (
	// SeverityInfo is an informational diagnostic.
	SeverityInfo Severity = iota

	// SeverityWarn is a warning diagnostic.
	SeverityWarn

	// SeverityErr is an error diagnostic.
	SeverityErr
)

// Diagnostic is a runtime analysis diagnostic.
type Diagnostic struct {
	// Severity identifies the diagnostic severity.
	Severity Severity

	// Message is the diagnostic message.
	Message string

	// Span is the byte range highlighted by the diagnostic.
	Span location.Span
}

// Options configures runtime analysis behavior.
type Options struct {
	// StopOnFirstDiagnostic reports only the first diagnostic encountered.
	StopOnFirstDiagnostic bool
}

// Engine evaluates compiled rules over source text.
type Engine struct {
	rulesByTokenName map[string][]ir.Rule
	rulesBySyntaxID  [][]ir.Rule
	syntaxParser     syntax.Parser
	hasSyntaxParser  bool
}

// New returns an analysis engine for program.
func New(program ir.Program) Engine {
	rulesByTokenName := make(map[string][]ir.Rule, len(program.Tokens))
	rulesBySyntaxID := make([][]ir.Rule, len(program.SyntaxNodes))

	for idx := range program.Rules {
		rule := program.Rules[idx]

		if rule.MatchKind == ir.RuleMatchToken {
			tokenName := program.Tokens[rule.MatchIndex].Name
			rulesByTokenName[tokenName] = append(rulesByTokenName[tokenName], rule)
			continue
		}

		rulesBySyntaxID[rule.MatchIndex] = append(rulesBySyntaxID[rule.MatchIndex], rule)
	}

	syntaxParser := syntax.Parser{}
	hasSyntaxParser := false

	if len(program.SyntaxNodes) > 0 && program.SyntaxRoot >= 0 {
		var err error
		syntaxParser, err = syntax.New(program, program.SyntaxNodes[program.SyntaxRoot].Name)

		if err == nil {
			hasSyntaxParser = true
		}
	}

	return Engine{
		rulesByTokenName: rulesByTokenName,
		rulesBySyntaxID:  rulesBySyntaxID,
		syntaxParser:     syntaxParser,
		hasSyntaxParser:  hasSyntaxParser,
	}
}

// Analyze scans src, builds the runtime syntax tree when configured, and evaluates compiled rules.
func (engine Engine) Analyze(program ir.Program, src string, options Options) []Diagnostic {
	scan := scanner.New(program, src)
	tokens := make([]scanner.Token, 0, len(src))
	var diagnostics []Diagnostic

	for {
		scannedToken, scanDiagnostic, ok := scan.Next()

		if !ok {
			break
		}

		if scanDiagnostic.Message != "" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: SeverityErr,
				Message:  scanDiagnostic.Message,
				Span:     scanDiagnostic.Span,
			})

			return diagnostics
		}

		tokens = append(tokens, scannedToken)
		rules := engine.rulesByTokenName[scannedToken.Name]

		for idx := range rules {
			rule := rules[idx]

			if !matchesCondition(rule.Where, scannedToken.Text, len(scannedToken.Text)) {
				continue
			}

			diagnostics = append(diagnostics, Diagnostic{
				Severity: severity(rule.Report.Severity),
				Message:  rule.Report.Message,
				Span:     scannedToken.Span,
			})

			if options.StopOnFirstDiagnostic {
				return diagnostics
			}
		}
	}

	if !engine.hasSyntaxParser || len(engine.rulesBySyntaxID) == 0 {
		return diagnostics
	}

	var tree syntax.Tree

	if !engine.syntaxParser.ParseInto(tokens, &tree) {
		diagnostics = append(diagnostics, Diagnostic{
			Severity: SeverityErr,
			Message:  `Source text does not match syntax root "` + program.SyntaxNodes[program.SyntaxRoot].Name + `".`,
			Span: location.Span{
				Start: 0,
				End:   location.Position(len(src)),
			},
		})

		return diagnostics
	}

	engine.evaluateSyntaxNode(program, src, tree, tree.Root, nil, &diagnostics, options)

	return diagnostics
}

func (engine Engine) evaluateSyntaxNode(program ir.Program, src string, tree syntax.Tree, nodeID syntax.NodeID, ancestors []uint32, diagnostics *[]Diagnostic, options Options) bool {
	node := tree.Node(nodeID)
	nodeRules := engine.rulesBySyntaxID[node.Kind]
	nodeText, nodeSpan := syntaxNodeTextAndSpan(src, tree, node)

	for idx := range nodeRules {
		rule := nodeRules[idx]

		if !matchesSyntaxScope(rule, ancestors) {
			continue
		}

		if !matchesCondition(rule.Where, nodeText, len(nodeText)) {
			continue
		}

		*diagnostics = append(*diagnostics, Diagnostic{
			Severity: severity(rule.Report.Severity),
			Message:  rule.Report.Message,
			Span:     nodeSpan,
		})

		if options.StopOnFirstDiagnostic {
			return true
		}
	}

	ancestors = append(ancestors, node.Kind)

	for _, childID := range tree.Children(node) {
		if engine.evaluateSyntaxNode(program, src, tree, childID, ancestors, diagnostics, options) {
			return true
		}
	}

	return false
}

func matchesSyntaxScope(rule ir.Rule, ancestors []uint32) bool {
	switch rule.MatchScopeKind {
	case ir.RuleMatchScopeInside:
		for idx := range ancestors {
			if int(ancestors[idx]) == rule.MatchScopeIndex {
				return true
			}
		}

		return false

	case ir.RuleMatchScopeOutside:
		for idx := range ancestors {
			if int(ancestors[idx]) == rule.MatchScopeIndex {
				return false
			}
		}

		return true

	default:
		return true
	}
}

func matchesCondition(condition ir.Condition, text string, length int) bool {
	switch condition.Property {
	case ir.ConditionPropertyNone:
		return true

	case ir.ConditionPropertyText:
		return compareStrings(text, condition.String, condition.Operator)

	default:
		return compareIntegers(length, condition.Integer, condition.Operator)
	}
}

func syntaxNodeTextAndSpan(src string, tree syntax.Tree, node syntax.Node) (string, location.Span) {
	if node.AmountOfTokens == 0 {
		position := location.Position(0)

		if node.FirstTokenIndex < uint32(len(tree.Tokens)) {
			position = tree.Tokens[node.FirstTokenIndex].Span.Start
		} else if len(tree.Tokens) > 0 {
			position = tree.Tokens[len(tree.Tokens)-1].Span.End
		}

		return "", location.Span{
			Start: position,
			End:   position,
		}
	}

	startToken := tree.Tokens[node.FirstTokenIndex]
	endToken := tree.Tokens[node.FirstTokenIndex+node.AmountOfTokens-1]

	return src[startToken.Span.Start:endToken.Span.End], location.Span{
		Start: startToken.Span.Start,
		End:   endToken.Span.End,
	}
}

func compareStrings(left, right string, operator ir.ConditionOperator) bool {
	switch operator {
	case ir.ConditionOperatorEqual:
		return left == right

	default:
		return left != right
	}
}

func compareIntegers(left, right int, operator ir.ConditionOperator) bool {
	switch operator {
	case ir.ConditionOperatorEqual:
		return left == right

	case ir.ConditionOperatorNotEqual:
		return left != right

	case ir.ConditionOperatorLess:
		return left < right

	case ir.ConditionOperatorLessEqual:
		return left <= right

	case ir.ConditionOperatorGreater:
		return left > right

	default:
		return left >= right
	}
}

func severity(value ir.Severity) Severity {
	switch value {
	case ir.SeverityInfo:
		return SeverityInfo

	case ir.SeverityWarn:
		return SeverityWarn

	default:
		return SeverityErr
	}
}
