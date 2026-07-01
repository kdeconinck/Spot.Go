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
	rulesByTokenName       map[string][]ir.Rule
	rulesBySyntaxID        [][]ir.Rule
	adjacentRulesByRightID [][]ir.Rule
	syntaxParser           syntax.Parser
	hasSyntaxParser        bool
}

// New returns an analysis engine for program.
func New(program ir.Program) Engine {
	rulesByTokenName := make(map[string][]ir.Rule, len(program.Tokens))
	rulesBySyntaxID := make([][]ir.Rule, len(program.SyntaxNodes))
	adjacentRulesByRightID := make([][]ir.Rule, len(program.SyntaxNodes))

	for idx := range program.Rules {
		rule := program.Rules[idx]

		if rule.MatchKind == ir.RuleMatchToken {
			tokenName := program.Tokens[rule.MatchIndex].Name
			rulesByTokenName[tokenName] = append(rulesByTokenName[tokenName], rule)
			continue
		}

		if rule.RelationKind == ir.RuleMatchRelationAdjacentSibling {
			adjacentRulesByRightID[rule.MatchIndex] = append(adjacentRulesByRightID[rule.MatchIndex], rule)
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
		rulesByTokenName:       rulesByTokenName,
		rulesBySyntaxID:        rulesBySyntaxID,
		adjacentRulesByRightID: adjacentRulesByRightID,
		syntaxParser:           syntaxParser,
		hasSyntaxParser:        hasSyntaxParser,
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

			if !matchesCondition(rule.Where, conditionContext{
				matchText:   scannedToken.Text,
				matchLength: len(scannedToken.Text),
			}) {
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

		if !matchesCondition(rule.Where, conditionContext{
			program:      program,
			src:          src,
			tree:         tree,
			matchText:    nodeText,
			matchLength:  len(nodeText),
			matchNodeID:  nodeID,
			hasMatchNode: true,
		}) {
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

	childEdges := tree.Children(node)

	for _, childEdge := range childEdges {
		if engine.evaluateSyntaxNode(program, src, tree, childEdge.ChildID, ancestors, diagnostics, options) {
			return true
		}
	}

	for idx := 1; idx < len(childEdges); idx++ {
		if engine.evaluateAdjacentSyntaxPair(program, src, tree, childEdges[idx-1].ChildID, childEdges[idx].ChildID, ancestors, diagnostics, options) {
			return true
		}
	}

	return false
}

func (engine Engine) evaluateAdjacentSyntaxPair(program ir.Program, src string, tree syntax.Tree, leftID, rightID syntax.NodeID, ancestors []uint32, diagnostics *[]Diagnostic, options Options) bool {
	leftNode := tree.Node(leftID)
	rightNode := tree.Node(rightID)
	rules := engine.adjacentRulesByRightID[rightNode.Kind]

	if len(rules) == 0 {
		return false
	}

	leftText, _ := syntaxNodeTextAndSpan(src, tree, leftNode)
	rightText, rightSpan := syntaxNodeTextAndSpan(src, tree, rightNode)
	context := conditionContext{
		program:        program,
		src:            src,
		tree:           tree,
		matchText:      rightText,
		matchLength:    len(rightText),
		relatedText:    leftText,
		relatedLength:  len(leftText),
		gapBlankLines:  blankLinesBetween(src, tree, leftNode, rightNode),
		matchNodeID:    rightID,
		relatedNodeID:  leftID,
		hasMatchNode:   true,
		hasRelatedNode: true,
	}

	for idx := range rules {
		rule := rules[idx]

		if int(leftNode.Kind) != rule.RelatedMatchIndex {
			continue
		}

		if !matchesSyntaxScope(rule, ancestors) {
			continue
		}

		if !matchesCondition(rule.Where, context) {
			continue
		}

		*diagnostics = append(*diagnostics, Diagnostic{
			Severity: severity(rule.Report.Severity),
			Message:  rule.Report.Message,
			Span:     rightSpan,
		})

		if options.StopOnFirstDiagnostic {
			return true
		}
	}

	return false
}

func matchesSyntaxScope(rule ir.Rule, ancestors []uint32) bool {
	switch rule.MatchScopeKind {
	case ir.RuleMatchScopeParent:
		if len(ancestors) == 0 {
			return false
		}

		return int(ancestors[len(ancestors)-1]) == rule.MatchScopeIndex

	case ir.RuleMatchScopeInside:
		for idx := range ancestors {
			if int(ancestors[idx]) == rule.MatchScopeIndex {
				return true
			}
		}

		return false

	case ir.RuleMatchScopeParentOutside:
		if len(ancestors) == 0 {
			return true
		}

		return int(ancestors[len(ancestors)-1]) != rule.MatchScopeIndex

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

type conditionContext struct {
	program        ir.Program
	src            string
	tree           syntax.Tree
	matchText      string
	matchLength    int
	relatedText    string
	relatedLength  int
	gapBlankLines  int
	matchNodeID    syntax.NodeID
	relatedNodeID  syntax.NodeID
	hasMatchNode   bool
	hasRelatedNode bool
}

func matchesCondition(condition ir.Condition, context conditionContext) bool {
	if condition.LeftProperty == ir.ConditionPropertyNone {
		return true
	}

	if condition.RightSubject != ir.ConditionSubjectNone {
		leftString, leftInteger := conditionValue(context, condition.LeftSubject, condition.LeftPath, condition.LeftProperty)
		rightString, rightInteger := conditionValue(context, condition.RightSubject, condition.RightPath, condition.RightProperty)

		if condition.LeftProperty == ir.ConditionPropertyText {
			return compareStrings(leftString, rightString, condition.Operator)
		}

		return compareIntegers(leftInteger, rightInteger, condition.Operator)
	}

	if condition.LeftProperty == ir.ConditionPropertyText {
		return compareStrings(conditionValueString(context, condition.LeftSubject, condition.LeftPath, condition.LeftProperty), condition.String, condition.Operator)
	}

	return compareIntegers(conditionValueInteger(context, condition.LeftSubject, condition.LeftPath, condition.LeftProperty), condition.Integer, condition.Operator)
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

	case ir.ConditionOperatorNotEqual:
		return left != right

	case ir.ConditionOperatorLess:
		return left < right

	case ir.ConditionOperatorLessEqual:
		return left <= right

	case ir.ConditionOperatorGreater:
		return left > right

	case ir.ConditionOperatorGreaterEqual:
		return left >= right

	default:
		return len(left) >= len(right) && left[:len(right)] == right
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

func conditionValue(context conditionContext, subject ir.ConditionSubjectKind, path []uint32, property ir.ConditionProperty) (string, int) {
	if property == ir.ConditionPropertyText {
		return conditionValueString(context, subject, path, property), 0
	}

	return "", conditionValueInteger(context, subject, path, property)
}

func conditionValueString(context conditionContext, subject ir.ConditionSubjectKind, path []uint32, property ir.ConditionProperty) string {
	if property != ir.ConditionPropertyText {
		return ""
	}

	if len(path) > 0 {
		node, ok := resolveConditionNode(context, subject, path)

		if !ok {
			return ""
		}

		text, _ := syntaxNodeTextAndSpan(context.src, context.tree, node)

		return text
	}

	switch subject {
	case ir.ConditionSubjectRelatedMatch:
		return context.relatedText

	default:
		return context.matchText
	}
}

func conditionValueInteger(context conditionContext, subject ir.ConditionSubjectKind, path []uint32, property ir.ConditionProperty) int {
	if len(path) > 0 {
		node, ok := resolveConditionNode(context, subject, path)

		if !ok {
			return 0
		}

		if property == ir.ConditionPropertyLength {
			text, _ := syntaxNodeTextAndSpan(context.src, context.tree, node)

			return len(text)
		}

		return 0
	}

	switch subject {
	case ir.ConditionSubjectGap:
		return context.gapBlankLines

	case ir.ConditionSubjectRelatedMatch:
		return context.relatedLength

	default:
		return context.matchLength
	}
}

func resolveConditionNode(context conditionContext, subject ir.ConditionSubjectKind, path []uint32) (syntax.Node, bool) {
	currentID, ok := conditionNodeID(context, subject)

	if !ok {
		return syntax.Node{}, false
	}

	current := context.tree.Node(currentID)

	for idx := range path {
		childEdge, found := context.tree.ChildByField(current, path[idx])

		if !found {
			return syntax.Node{}, false
		}

		current = context.tree.Node(childEdge.ChildID)
	}

	return current, true
}

func conditionNodeID(context conditionContext, subject ir.ConditionSubjectKind) (syntax.NodeID, bool) {
	switch subject {
	case ir.ConditionSubjectRelatedMatch:
		return context.relatedNodeID, context.hasRelatedNode

	case ir.ConditionSubjectMatch:
		return context.matchNodeID, context.hasMatchNode

	default:
		return 0, false
	}
}

func blankLinesBetween(src string, tree syntax.Tree, leftNode, rightNode syntax.Node) int {
	if leftNode.AmountOfTokens == 0 || rightNode.AmountOfTokens == 0 {
		return 0
	}

	leftEnd := tree.Tokens[leftNode.FirstTokenIndex+leftNode.AmountOfTokens-1].Span.End
	rightStart := tree.Tokens[rightNode.FirstTokenIndex].Span.Start
	gap := src[leftEnd:rightStart]
	newlines := 0

	for idx := 0; idx < len(gap); idx++ {
		if gap[idx] == '\n' {
			newlines++
		}
	}

	if newlines == 0 {
		return 0
	}

	return newlines - 1
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
