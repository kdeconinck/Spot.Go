// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package scanner tokenizes source text using compiled Spot token definitions.
package scanner

import (
	"strconv"

	"github.com/kdeconinck/spot/ir"
	"github.com/kdeconinck/spot/location"
)

// Scanner tokenizes source text into Spot tokens.
type Scanner struct {
	tokens       []tokenDefinition
	states       []state
	start        int
	startClosure []int
	closures     [][]int
}

// Token is a scanner output token.
type Token struct {
	// Name is the DSL token name that matched the source text.
	Name string

	// Text is the exact source text matched by the token.
	Text string

	// Span is the byte range covered by the token.
	Span location.Span
}

// Diagnostic reports a scanning failure.
type Diagnostic struct {
	// Message explains why scanning failed.
	Message string

	// Span identifies the source range where scanning failed.
	Span location.Span
}

// New returns a scanner backed by an NFA compiled from program.
func New(program ir.Program) Scanner {
	builder := machineBuilder{
		tokens: make([]tokenDefinition, 0, len(program.Tokens)),
	}

	tokenStarts := make([]int, 0, len(program.Tokens))

	for idx := range program.Tokens {
		token := program.Tokens[idx]
		fragment := builder.buildExpression(token.Expression)
		accept := builder.newState(state{
			kind:       stateAccept,
			tokenIndex: idx,
		})

		builder.link(fragment.end, accept)
		builder.tokens = append(builder.tokens, tokenDefinition{
			name: token.Name,
			skip: token.Skip,
		})
		tokenStarts = append(tokenStarts, fragment.start)
	}

	start := tokenStarts[0]

	for idx := 1; idx < len(tokenStarts); idx++ {
		start = builder.newState(state{
			kind: stateSplit,
			next: tokenStarts[idx],
			alt:  start,
		})
	}

	closures := buildClosures(builder.states)

	return Scanner{
		tokens:       builder.tokens,
		states:       builder.states,
		start:        start,
		startClosure: closures[start],
		closures:     closures,
	}
}

// Scan tokenizes src and returns the emitted tokens and any scanning diagnostics.
func (scanner Scanner) Scan(src string) ([]Token, []Diagnostic) {
	tokens := make([]Token, 0, len(src))
	offset := 0

	for offset < len(src) {
		match, ok := scanner.match(src, offset)

		if !ok {
			return tokens, []Diagnostic{{
				Message: "No token matched at byte offset " + strconv.Itoa(offset) + ".",
				Span: location.Span{
					Start: location.Position(offset),
					End:   location.Position(offset + 1),
				},
			}}
		}

		definition := scanner.tokens[match.tokenIndex]

		if !definition.skip {
			tokens = append(tokens, Token{
				Name: definition.name,
				Text: src[offset:match.end],
				Span: location.Span{
					Start: location.Position(offset),
					End:   location.Position(match.end),
				},
			})
		}

		offset = match.end
	}

	return tokens, nil
}

type tokenDefinition struct {
	name string
	skip bool
}

type match struct {
	tokenIndex int
	end        int
}

func (scanner Scanner) match(src string, start int) (match, bool) {
	best := match{
		tokenIndex: -1,
		end:        start,
	}
	active := scanner.startClosure
	scratch := scanScratch{
		nextSeen: make([]bool, len(scanner.states)),
	}

	// Active always contains the epsilon-closed set of states reachable at the
	// current offset. That lets the main loop focus on only two operations:
	// consume one byte, then re-expand through epsilon transitions.
	scanner.recordAccepts(active, start, &best)

	for offset := start; offset < len(src); offset++ {
		next := scanner.step(active, src[offset], &scratch)

		if len(next) == 0 {
			break
		}

		scanner.recordAccepts(next, offset+1, &best)
		active = next
	}

	if best.tokenIndex == -1 {
		return match{}, false
	}

	return best, true
}

func (scanner Scanner) recordAccepts(active []int, end int, best *match) {
	for idx := range active {
		state := scanner.states[active[idx]]

		if state.kind != stateAccept {
			continue
		}

		if best.tokenIndex == -1 || end > best.end || (end == best.end && state.tokenIndex < best.tokenIndex) {
			best.tokenIndex = state.tokenIndex
			best.end = end
		}
	}
}

func (scanner Scanner) step(active []int, ch byte, scratch *scanScratch) []int {
	scratch.reset()

	for idx := range active {
		state := scanner.states[active[idx]]

		if !state.matches(ch) {
			continue
		}

		// A consuming transition advances to exactly one target state. The target
		// closure is then merged into the next active set so the following loop
		// iteration sees every epsilon-reachable continuation.
		scratch.addClosures(scanner.closures[state.next])
	}

	return scratch.next
}

type scanScratch struct {
	next     []int
	nextSeen []bool
}

func (scratch *scanScratch) reset() {
	for idx := range scratch.next {
		scratch.nextSeen[scratch.next[idx]] = false
	}

	scratch.next = scratch.next[:0]
}

func (scratch *scanScratch) addClosures(states []int) {
	for idx := range states {
		stateIndex := states[idx]

		if scratch.nextSeen[stateIndex] {
			continue
		}

		scratch.nextSeen[stateIndex] = true
		scratch.next = append(scratch.next, stateIndex)
	}
}
