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
}

// New returns an analysis engine for program.
func New(program ir.Program) Engine {
	rulesByTokenName := make(map[string][]ir.Rule, len(program.Tokens))

	for idx := range program.Rules {
		rule := program.Rules[idx]
		tokenName := program.Tokens[rule.MatchToken].Name
		rulesByTokenName[tokenName] = append(rulesByTokenName[tokenName], rule)
	}

	return Engine{
		rulesByTokenName: rulesByTokenName,
	}
}

// Analyze scans src and evaluates compiled rules in one pass.
func (engine Engine) Analyze(program ir.Program, src string, options Options) []Diagnostic {
	scan := scanner.New(program, src)
	var diagnostics []Diagnostic

	for {
		token, scanDiagnostic, ok := scan.Next()

		if !ok {
			return diagnostics
		}

		if scanDiagnostic.Message != "" {
			diagnostics = append(diagnostics, Diagnostic{
				Severity: SeverityErr,
				Message:  scanDiagnostic.Message,
				Span:     scanDiagnostic.Span,
			})

			return diagnostics
		}

		rules := engine.rulesByTokenName[token.Name]

		for idx := range rules {
			rule := rules[idx]

			if !matchesCondition(rule.Where, token) {
				continue
			}

			diagnostics = append(diagnostics, Diagnostic{
				Severity: severity(rule.Report.Severity),
				Message:  rule.Report.Message,
				Span:     token.Span,
			})

			if options.StopOnFirstDiagnostic {
				return diagnostics
			}
		}
	}
}

func matchesCondition(condition ir.Condition, token scanner.Token) bool {
	switch condition.Property {
	case ir.ConditionPropertyNone:
		return true

	case ir.ConditionPropertyText:
		return compareStrings(token.Text, condition.String, condition.Operator)

	default:
		return compareIntegers(len(token.Text), condition.Integer, condition.Operator)
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
