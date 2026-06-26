// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package compiler compiles validated Spot DSL syntax into runtime-oriented data structures.
package compiler

import "github.com/kdeconinck/spot/syntax"

// Compile compiles a validated Spot DSL document into a token program.
func Compile(document syntax.Document) Program {
	definitions := map[string]syntax.Definition{}

	for idx := range document.Definitions.Definitions {
		definition := document.Definitions.Definitions[idx]
		definitions[definition.Name.Text] = definition
	}

	program := Program{
		Tokens: make([]Token, 0, len(document.Tokens.Tokens)),
	}

	for idx := range document.Tokens.Tokens {
		token := document.Tokens.Tokens[idx]
		program.Tokens = append(program.Tokens, Token{
			Name:       token.Name.Text,
			Expression: compileExpression(token.Expression, definitions),
			Skip:       token.Skip.Kind == syntax.TokenSkip,
		})
	}

	return program
}

func compileExpression(expression syntax.DefinitionExpression, definitions map[string]syntax.Definition) Expression {
	switch expression.Kind {
	case syntax.DefinitionExpressionCharacter:
		return Expression{
			Kind:      ExpressionCharacter,
			Character: characterValue(expression.Start),
		}

	case syntax.DefinitionExpressionString:
		return Expression{
			Kind:   ExpressionString,
			String: stringValue(expression.Start),
		}

	case syntax.DefinitionExpressionRange:
		return Expression{
			Kind:       ExpressionRange,
			RangeStart: characterValue(expression.Start),
			RangeEnd:   characterValue(expression.End),
		}

	case syntax.DefinitionExpressionReference:
		return compileExpression(definitions[expression.Start.Text].Expression, definitions)

	case syntax.DefinitionExpressionConcatenation:
		return Expression{
			Kind:  ExpressionConcatenation,
			Terms: compileTerms(expression.Terms, definitions),
		}

	case syntax.DefinitionExpressionAlternation:
		return Expression{
			Kind:  ExpressionAlternation,
			Terms: compileTerms(expression.Terms, definitions),
		}

	case syntax.DefinitionExpressionGroup:
		return compileExpression(*expression.Inner, definitions)

	default:
		return Expression{
			Kind:       ExpressionRepetition,
			Inner:      pointer(compileExpression(*expression.Inner, definitions)),
			Repetition: repetitionKind(expression.Operator.Kind),
		}
	}
}

func compileTerms(expressions []syntax.DefinitionExpression, definitions map[string]syntax.Definition) []Expression {
	terms := make([]Expression, 0, len(expressions))

	for idx := range expressions {
		terms = append(terms, compileExpression(expressions[idx], definitions))
	}

	return terms
}

func repetitionKind(kind syntax.TokenKind) RepetitionKind {
	switch kind {
	case syntax.TokenQuestion:
		return RepetitionZeroOrOne

	case syntax.TokenStar:
		return RepetitionZeroOrMore

	default:
		return RepetitionOneOrMore
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

func pointer(expression Expression) *Expression {
	return &expression
}
