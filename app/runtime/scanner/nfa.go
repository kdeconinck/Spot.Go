// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package scanner tokenizes source text using compiled Spot token definitions.
package scanner

import "github.com/kdeconinck/spot/runtime/ir"

const noState = -1

type stateKind uint8

const (
	stateEpsilon stateKind = iota
	stateSplit
	stateByte
	stateRange
	stateAccept
)

type state struct {
	kind       stateKind
	next       int
	alt        int
	start      byte
	end        byte
	tokenIndex int
}

func (state state) matches(ch byte) bool {
	switch state.kind {
	case stateByte:
		return state.start == ch

	case stateRange:
		return state.start <= ch && ch <= state.end

	default:
		return false
	}
}

type fragment struct {
	start int
	end   int
}

type machineBuilder struct {
	tokens []tokenDefinition
	states []state
}

func (builder *machineBuilder) buildExpression(expression ir.Expression) fragment {
	switch expression.Kind {
	case ir.ExpressionCharacter:
		return builder.buildByte(expression.Character)

	case ir.ExpressionString:
		return builder.buildString(expression.String)

	case ir.ExpressionRange:
		return builder.buildRange(expression.RangeStart, expression.RangeEnd)

	case ir.ExpressionConcatenation:
		return builder.buildConcatenation(expression.Terms)

	case ir.ExpressionAlternation:
		return builder.buildAlternation(expression.Terms)

	default:
		return builder.buildRepetition(*expression.Inner, expression.Repetition)
	}
}

func (builder *machineBuilder) buildByte(ch byte) fragment {
	end := builder.newState(state{
		kind: stateEpsilon,
		next: noState,
		alt:  noState,
	})

	start := builder.newState(state{
		kind:  stateByte,
		next:  end,
		start: ch,
	})

	return fragment{start: start, end: end}
}

func (builder *machineBuilder) buildString(text string) fragment {
	fragment := builder.buildByte(text[0])

	for idx := 1; idx < len(text); idx++ {
		fragment = builder.concatenate(fragment, builder.buildByte(text[idx]))
	}

	return fragment
}

func (builder *machineBuilder) buildRange(start, end byte) fragment {
	terminal := builder.newState(state{
		kind: stateEpsilon,
		next: noState,
		alt:  noState,
	})

	initial := builder.newState(state{
		kind:  stateRange,
		next:  terminal,
		start: start,
		end:   end,
	})

	return fragment{start: initial, end: terminal}
}

func (builder *machineBuilder) buildConcatenation(terms []ir.Expression) fragment {
	combined := builder.buildExpression(terms[0])

	for idx := 1; idx < len(terms); idx++ {
		// Each fragment owns a single open end state. Concatenation is therefore
		// just "connect the previous end to the next start".
		combined = builder.concatenate(combined, builder.buildExpression(terms[idx]))
	}

	return combined
}

func (builder *machineBuilder) concatenate(left, right fragment) fragment {
	builder.link(left.end, right.start)

	return fragment{
		start: left.start,
		end:   right.end,
	}
}

func (builder *machineBuilder) buildAlternation(terms []ir.Expression) fragment {
	combined := builder.buildExpression(terms[0])

	for idx := 1; idx < len(terms); idx++ {
		combined = builder.alternate(combined, builder.buildExpression(terms[idx]))
	}

	return combined
}

func (builder *machineBuilder) alternate(left, right fragment) fragment {
	end := builder.newState(state{
		kind: stateEpsilon,
		next: noState,
		alt:  noState,
	})
	start := builder.newState(state{
		kind: stateSplit,
		next: left.start,
		alt:  right.start,
	})

	builder.link(left.end, end)
	builder.link(right.end, end)

	// The split state is the NFA branch point. Either branch may proceed, and
	// both reconnect at the shared end state.
	return fragment{start: start, end: end}
}

func (builder *machineBuilder) buildRepetition(inner ir.Expression, repetition ir.RepetitionKind) fragment {
	switch repetition {
	case ir.RepetitionZeroOrOne:
		return builder.buildZeroOrOne(inner)

	case ir.RepetitionZeroOrMore:
		return builder.buildZeroOrMore(inner)

	default:
		return builder.buildOneOrMore(inner)
	}
}

func (builder *machineBuilder) buildZeroOrOne(inner ir.Expression) fragment {
	innerFragment := builder.buildExpression(inner)
	end := builder.newState(state{
		kind: stateEpsilon,
		next: noState,
		alt:  noState,
	})
	start := builder.newState(state{
		kind: stateSplit,
		next: innerFragment.start,
		alt:  end,
	})

	builder.link(innerFragment.end, end)

	return fragment{start: start, end: end}
}

func (builder *machineBuilder) buildZeroOrMore(inner ir.Expression) fragment {
	innerFragment := builder.buildExpression(inner)
	end := builder.newState(state{
		kind: stateEpsilon,
		next: noState,
		alt:  noState,
	})
	start := builder.newState(state{
		kind: stateSplit,
		next: innerFragment.start,
		alt:  end,
	})
	loop := builder.newState(state{
		kind: stateSplit,
		next: innerFragment.start,
		alt:  end,
	})

	builder.link(innerFragment.end, loop)

	return fragment{start: start, end: end}
}

func (builder *machineBuilder) buildOneOrMore(inner ir.Expression) fragment {
	innerFragment := builder.buildExpression(inner)
	end := builder.newState(state{
		kind: stateEpsilon,
		next: noState,
		alt:  noState,
	})
	loop := builder.newState(state{
		kind: stateSplit,
		next: innerFragment.start,
		alt:  end,
	})

	builder.link(innerFragment.end, loop)

	return fragment{start: innerFragment.start, end: end}
}

func (builder *machineBuilder) link(from, to int) {
	builder.states[from].next = to
}

func (builder *machineBuilder) newState(state state) int {
	builder.states = append(builder.states, state)

	return len(builder.states) - 1
}

func buildClosures(states []state) [][]int {
	closures := make([][]int, len(states))

	for idx := range states {
		closures[idx] = buildClosure(states, idx)
	}

	return closures
}

func buildClosure(states []state, start int) []int {
	stack := []int{start}
	seen := make([]bool, len(states))
	closure := make([]int, 0, 4)

	for len(stack) > 0 {
		last := len(stack) - 1
		stateIndex := stack[last]
		stack = stack[:last]

		if seen[stateIndex] {
			continue
		}

		seen[stateIndex] = true
		closure = append(closure, stateIndex)

		// Closures walk only the non-consuming edges. Byte and range states stay
		// in the closure as candidates for the next input byte.
		switch states[stateIndex].kind {
		case stateEpsilon:
			if states[stateIndex].next != noState {
				stack = append(stack, states[stateIndex].next)
			}

		case stateSplit:
			stack = append(stack, states[stateIndex].next, states[stateIndex].alt)
		}
	}

	return closure
}
