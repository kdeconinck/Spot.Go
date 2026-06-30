// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package scanner tokenizes source text using compiled Spot token definitions.
package scanner

import (
	"strconv"

	"github.com/kdeconinck/spot/location"
	"github.com/kdeconinck/spot/runtime/ir"
)

// Scanner tokenizes source text into Spot tokens.
type Scanner struct {
	tokens       []tokenDefinition
	states       []state
	start        int
	startClosure []int
	closures     [][]int
	src          string
	offset       int
	scratch      scanScratch
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

// New returns a scanner backed by an NFA compiled from program and initialized for src.
func New(program ir.Program, src string) Scanner {
	builder := machineBuilder{
		tokens: make([]tokenDefinition, 0, len(program.Tokens)),
		arena:  program.Expressions,
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
		src:          src,
		scratch: scanScratch{
			nextSeen: make([]bool, len(builder.states)),
		},
	}
}

// Reset prepares the scanner to tokenize a new source string.
func (scanner *Scanner) Reset(src string) {
	scanner.src = src
	scanner.offset = 0
	scanner.scratch.reset()
}

// Next returns the next emitted token or scanning diagnostic.
//
// The returned boolean reports whether a token or diagnostic was produced.
// When it is false, the scanner has reached the end of the source text.
func (scanner *Scanner) Next() (Token, Diagnostic, bool) {
	for scanner.offset < len(scanner.src) {
		match, ok := scanner.match(scanner.offset)

		if !ok {
			diagnostic := Diagnostic{
				Message: "No token matched at byte offset " + strconv.Itoa(scanner.offset) + ".",
				Span: location.Span{
					Start: location.Position(scanner.offset),
					End:   location.Position(scanner.offset + 1),
				},
			}
			scanner.offset = len(scanner.src)

			return Token{}, diagnostic, true
		}

		start := scanner.offset
		scanner.offset = match.end

		definition := scanner.tokens[match.tokenIndex]

		if definition.skip {
			continue
		}

		return Token{
			Name: definition.name,
			Text: scanner.src[start:match.end],
			Span: location.Span{
				Start: location.Position(start),
				End:   location.Position(match.end),
			},
		}, Diagnostic{}, true
	}

	return Token{}, Diagnostic{}, false
}

type tokenDefinition struct {
	name string
	skip bool
}

type match struct {
	tokenIndex int
	end        int
}

func (scanner *Scanner) match(start int) (match, bool) {
	best := match{
		tokenIndex: -1,
		end:        start,
	}
	active := scanner.startClosure

	// Active always contains the epsilon-closed set of states reachable at the
	// current offset. That lets the main loop focus on only two operations:
	// consume one byte, then re-expand through epsilon transitions.
	scanner.recordAccepts(active, start, &best)

	for offset := start; offset < len(scanner.src); offset++ {
		next := scanner.step(active, scanner.src[offset], &scanner.scratch)

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
