// =====================================================================================================================
// == LICENSE:                 Copyright (c) 2026 Kevin De Coninck.
// == SPDX-License-Identifier: LicenseRef-PolyForm-Noncommercial-1.0.0
// =====================================================================================================================

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kdeconinck/spot/qa/claim"
)

func Test_matchPattern(t *testing.T) {
	t.Parallel()

	for tcName, tc := range map[string]struct {
		inPattern string
		inName    string
		want      bool
	}{
		"When a recursive glob matches nested Go files, the pattern matches.": {
			inPattern: "**/*.go",
			inName:    "internal/app/main.go",
			want:      true,
		},
		"When a recursive glob matches a top-level Go file, the pattern matches.": {
			inPattern: "**/*.go",
			inName:    "main.go",
			want:      true,
		},
		"When an exclude pattern matches a vendor file, the pattern matches.": {
			inPattern: "vendor/**",
			inName:    "vendor/pkg/file.go",
			want:      true,
		},
		"When a single-segment wildcard would need to cross a directory boundary, the pattern does not match.": {
			inPattern: "*.go",
			inName:    "pkg/main.go",
			want:      false,
		},
		"When a question mark matches one character in a segment, the pattern matches.": {
			inPattern: "file?.go",
			inName:    "file1.go",
			want:      true,
		},
		"When a literal segment differs, the pattern does not match.": {
			inPattern: "cmd/*.go",
			inName:    "pkg/main.go",
			want:      false,
		},
	} {
		t.Run(tcName, func(t *testing.T) {
			t.Parallel()

			// Act.
			got := matchPattern(tc.inPattern, tc.inName)

			// Assert.
			claim.Equal(t, tcName, tc.want, got, "Match")
		})
	}
}

func Test_run_AnalyzesScopedFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dslPath := filepath.Join(root, "spot.dsl")
	includePath := filepath.Join(root, "main.go")
	excludeDir := filepath.Join(root, "vendor")
	excludePath := filepath.Join(excludeDir, "ignored.go")
	otherPath := filepath.Join(root, "notes.txt")

	err := os.Mkdir(excludeDir, 0o755)
	claim.Equal(t, "When creating the excluded directory, no error is returned.", nil, err, "MakeDir Error")

	err = os.WriteFile(dslPath, []byte(`scope {
    include "**/*.go"
    exclude "vendor/**"
}
definitions {
    letter = 'a'..'z'
}
tokens {
    Whitespace = ' '+ skip
    Identifier = letter+
}
rules {
    rule PublicIdentifier {
        match Identifier
        where Identifier.text == "public"
        report warn at Identifier "Public identifier found"
    }
}`), 0o644)
	claim.Equal(t, "When writing the DSL file, no error is returned.", nil, err, "DSL Write Error")

	err = os.WriteFile(includePath, []byte("public"), 0o644)
	claim.Equal(t, "When writing the included file, no error is returned.", nil, err, "Include Write Error")

	err = os.WriteFile(excludePath, []byte("public"), 0o644)
	claim.Equal(t, "When writing the excluded file, no error is returned.", nil, err, "Exclude Write Error")

	err = os.WriteFile(otherPath, []byte("public"), 0o644)
	claim.Equal(t, "When writing the non-matching file, no error is returned.", nil, err, "Other Write Error")

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	// Act.
	exitCode := run([]string{dslPath, root}, &stdout, &stderr)

	// Assert.
	claim.Equal(t, "When a scoped file produces a diagnostic, the CLI exits with status 1.", 1, exitCode, "Exit Code")
	claim.Equal(t, "When scoped analysis succeeds, stderr remains empty.", "", stderr.String(), "Stderr")
	claim.Equal(t, "When analyzing a directory, only included non-excluded files are reported.", "main.go:0-6: warn: Public identifier found\n", stdout.String(), "Stdout")
}
