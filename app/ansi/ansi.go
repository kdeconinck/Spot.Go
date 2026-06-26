// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package ansi provides helpers for building ANSI SGR escape sequences.
package ansi

import (
	"strconv"
	"strings"
)

// Code is a single ANSI SGR formatting code.
type Code uint8

const (
	ansiEscapePrefix = "\x1b["
	ansiResetSuffix  = "\x1b[0m"
)

// Supported ANSI formatting codes.
const (
	Reset Code = 0
	Red   Code = 31
	Green Code = 32
)

// Sequence is a precomputed ANSI SGR opening escape sequence.
// It is safe to reuse across calls and avoids rebuilding the opening sequence each time the caller appends it.
type Sequence struct {
	open string
}

// NewSequence returns a reusable ANSI SGR opening sequence for codes.
func NewSequence(codes []Code) Sequence {
	if len(codes) == 0 {
		return Sequence{
			// NOTE: Intentionally left blank.
		}
	}

	var sb strings.Builder

	sb.Grow(OpenLen(codes))
	sb.WriteString(ansiEscapePrefix)

	for i, code := range codes {
		if i > 0 {
			sb.WriteByte(';')
		}

		sb.WriteString(code.String())
	}

	sb.WriteByte('m')

	return Sequence{
		open: sb.String(),
	}
}

// OpenLen returns the number of bytes in the ANSI opening escape sequence.
func (seq Sequence) OpenLen() int { return len(seq.open) }

// ResetLen returns the number of bytes in the ANSI reset escape sequence.
func (seq Sequence) ResetLen() int {
	if seq.open == "" {
		return 0
	}

	return len(ansiResetSuffix)
}

// WriteOpenTo writes the ANSI opening escape sequence to builder.
func (seq Sequence) WriteOpenTo(builder *strings.Builder) {
	if seq.open == "" {
		return
	}

	builder.WriteString(seq.open)
}

// WriteResetTo appends the ANSI reset escape sequence to builder.
func (seq Sequence) WriteResetTo(builder *strings.Builder) {
	if seq.open == "" {
		return
	}

	builder.WriteString(ansiResetSuffix)
}

// String returns the decimal SGR code for code.
// The returned string is the numeric portion of an ANSI SGR sequence, such as "31" for red.
func (code Code) String() string {
	switch code {
	case Reset:
		return "0"

	case Red:
		return "31"

	case Green:
		return "32"

	default:
		return strconv.Itoa(int(code))
	}
}

// OpenLen returns the number of bytes in the ANSI opening escape sequence for the provided styling codes.
// If no codes are provided, OpenLen returns 0.
func OpenLen(codes []Code) int {
	if len(codes) == 0 {
		return 0
	}

	n := len(ansiEscapePrefix)

	for i, code := range codes {
		if i > 0 {
			n++
		}

		n += codeLen(code)
	}

	n++

	return n
}

func codeLen(code Code) int {
	switch code {
	case Reset:
		return 1

	case Red, Green:
		return 2

	default:
		switch {
		case code < 10:
			return 1

		case code < 100:
			return 2

		default:
			return 3
		}
	}
}
