// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

// Package claim provides simple test helpers for writing concise assertions.
// It defines minimal assertion functions that integrate with testing.TB and report failures with descriptive messages.
package claim

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/kdeconinck/spot/ansi"
)

// Performance: An optimized function to avoid the overhead of fmt.Sprint for common types such as string, bool, ...
func fmtValue[T any](v T) string {
	switch val := any(v).(type) {
	case string:
		return strconv.Quote(val)

	case bool:
		return strconv.FormatBool(val)

	case int:
		return strconv.Itoa(val)

	case int8:
		return strconv.FormatInt(int64(val), 10)

	case int16:
		return strconv.FormatInt(int64(val), 10)

	case int32:
		return strconv.FormatInt(int64(val), 10)

	case int64:
		return strconv.FormatInt(val, 10)

	case uint:
		return strconv.FormatUint(uint64(val), 10)

	case uint8:
		return strconv.FormatUint(uint64(val), 10)

	case uint16:
		return strconv.FormatUint(uint64(val), 10)

	case uint32:
		return strconv.FormatUint(uint64(val), 10)

	case uint64:
		return strconv.FormatUint(val, 10)

	case float32:
		return strconv.FormatFloat(float64(val), 'g', -1, 32)

	case float64:
		return strconv.FormatFloat(val, 'g', -1, 64)

	case error:
		if isNilValue(val) {
			return fmt.Sprintf("%T(nil)", val)
		}

		return val.Error()

	default:
		return fmt.Sprint(val)
	}
}

func isNilValue(v any) bool {
	rv := reflect.ValueOf(v)

	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return rv.IsNil()
	default:
		return false
	}
}

func writeFailureMessage(sb *strings.Builder, testName, label, want, got string) {
	rows := messageRows(testName, label, want, got)
	keyWidth := failureKeyWidth(rows)
	msgLen := failureMessageLen(rows, keyWidth)

	sb.Grow(msgLen) // Foresee enough room to write the actual failure message.

	sb.WriteByte('\n')
	sb.WriteByte('\n')

	for i := range rows {
		writeFailureRow(sb, rows[i], keyWidth)

		sb.WriteByte('\n')
	}

	sb.WriteByte('\n')
}

type messageRow struct {
	keyBase string
	label   string
	value   string
	style   ansi.Sequence
}

// The different styles applied to the messages.
var (
	wantRowStyle = ansi.NewSequence([]ansi.Code{
		ansi.Green,
	})

	gotRowStyle = ansi.NewSequence([]ansi.Code{
		ansi.Red,
	})
)

func messageRows(testName, label, want, got string) [3]messageRow {
	return [3]messageRow{
		{
			keyBase: "Test name",
			value:   testName,
		},
		{
			keyBase: "Expected",
			label:   label,
			value:   want,
			style:   wantRowStyle,
		},
		{
			keyBase: "Actual",
			label:   label,
			value:   got,
			style:   gotRowStyle,
		},
	}
}

func failureKeyWidth(rows [3]messageRow) int {
	width := 0

	for i := range rows {
		width = max(width, failureKeyLen(rows[i]))
	}

	return width
}

func failureMessageLen(rows [3]messageRow, keyWidth int) int {
	n := 2 + len(rows) + 1 // Leading blank line, one newline per row, and the trailing blank line.

	for i := range rows {
		n += failureRowLen(rows[i], keyWidth)
	}

	return n
}

func failureRowLen(row messageRow, keyWidth int) int {
	return row.style.OpenLen() + keyWidth + 1 + len(row.value) + row.style.ResetLen()
}

func writeFailureRow(sb *strings.Builder, row messageRow, keyWidth int) {
	row.style.WriteOpenTo(sb)

	writeFieldKey(sb, row.keyBase, row.label)
	writeKeyPadding(sb, keyWidth-failureKeyLen(row))

	sb.WriteByte(' ')
	sb.WriteString(row.value)

	row.style.WriteResetTo(sb)
}

func failureKeyLen(row messageRow) int {
	return fieldKeyLen(row.keyBase, row.label)
}

func fieldKeyLen(base, label string) int {
	n := len(base) + 1 // ':'

	if label != "" {
		n += 3 + len(label) // " (" + label + ")"
	}

	return n
}

func writeFieldKey(sb *strings.Builder, base, label string) {
	sb.WriteString(base)

	if label != "" {
		sb.WriteString(" (")
		sb.WriteString(label)
		sb.WriteByte(')')
	}

	sb.WriteByte(':')
}

func writeKeyPadding(sb *strings.Builder, n int) {
	for range n {
		sb.WriteByte(' ')
	}
}
