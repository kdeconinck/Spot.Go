// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package compiler compiles validated Spot DSL syntax into runtime-oriented data structures.
package compiler

import (
	"github.com/kdeconinck/spot/ir"
	"github.com/kdeconinck/spot/syntax"
)

// Compile compiles a validated Spot DSL document into a token program.
func Compile(document syntax.Document) ir.Program {
	definitions := map[string]syntax.Definition{}

	for idx := range document.Definitions.Definitions {
		definition := document.Definitions.Definitions[idx]
		definitions[definition.Name.Text] = definition
	}

	program := ir.Program{
		Tokens: make([]ir.Token, 0, len(document.Tokens.Tokens)),
	}

	for idx := range document.Tokens.Tokens {
		token := document.Tokens.Tokens[idx]
		program.Tokens = append(program.Tokens, ir.Token{
			Name:       token.Name.Text,
			Expression: compileExpression(token.Expression, definitions),
			Skip:       token.Skip.Kind == syntax.TokenSkip,
		})
	}

	return program
}

func compileExpression(expression syntax.DefinitionExpression, definitions map[string]syntax.Definition) ir.Expression {
	switch expression.Kind {
	case syntax.DefinitionExpressionCharacter:
		return ir.Expression{
			Kind:      ir.ExpressionCharacter,
			Character: characterValue(expression.Start),
		}

	case syntax.DefinitionExpressionString:
		return ir.Expression{
			Kind:   ir.ExpressionString,
			String: stringValue(expression.Start),
		}

	case syntax.DefinitionExpressionRange:
		return ir.Expression{
			Kind:       ir.ExpressionRange,
			RangeStart: characterValue(expression.Start),
			RangeEnd:   characterValue(expression.End),
		}

	case syntax.DefinitionExpressionReference:
		return compileExpression(definitions[expression.Start.Text].Expression, definitions)

	case syntax.DefinitionExpressionConcatenation:
		return ir.Expression{
			Kind:  ir.ExpressionConcatenation,
			Terms: compileTerms(expression.Terms, definitions),
		}

	case syntax.DefinitionExpressionAlternation:
		return ir.Expression{
			Kind:  ir.ExpressionAlternation,
			Terms: compileTerms(expression.Terms, definitions),
		}

	case syntax.DefinitionExpressionGroup:
		return compileExpression(*expression.Inner, definitions)

	default:
		return ir.Expression{
			Kind:       ir.ExpressionRepetition,
			Inner:      pointer(compileExpression(*expression.Inner, definitions)),
			Repetition: repetitionKind(expression.Operator.Kind),
		}
	}
}

func compileTerms(expressions []syntax.DefinitionExpression, definitions map[string]syntax.Definition) []ir.Expression {
	terms := make([]ir.Expression, 0, len(expressions))

	for idx := range expressions {
		terms = append(terms, compileExpression(expressions[idx], definitions))
	}

	return terms
}

func repetitionKind(kind syntax.TokenKind) ir.RepetitionKind {
	switch kind {
	case syntax.TokenQuestion:
		return ir.RepetitionZeroOrOne

	case syntax.TokenStar:
		return ir.RepetitionZeroOrMore

	default:
		return ir.RepetitionOneOrMore
	}
}

func characterValue(token syntax.Token) byte {
	if token.Text[1] != '\\' {
		return token.Text[1]
	}

	switch token.Text[2] {
	case '\\':
		return '\\'

	case '\'':
		return '\''

	case 'n':
		return '\n'

	case 'r':
		return '\r'

	default:
		return '\t'
	}
}

func stringValue(token syntax.Token) string {
	value := make([]byte, 0, len(token.Text)-2)

	for idx := 1; idx < len(token.Text)-1; idx++ {
		character := token.Text[idx]

		if character != '\\' {
			value = append(value, character)
			continue
		}

		idx++

		switch token.Text[idx] {
		case '\\':
			value = append(value, '\\')

		case '"':
			value = append(value, '"')

		case 'n':
			value = append(value, '\n')

		case 'r':
			value = append(value, '\r')

		default:
			value = append(value, '\t')
		}
	}

	return string(value)
}

func pointer(expression ir.Expression) *ir.Expression {
	return &expression
}
