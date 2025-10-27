package errors

import (
	"strings"
	"testing"
)

func TestParseError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		contains []string // Strings that should appear in output
	}{
		{
			name: "basic error with line number",
			err:  NewParseError(".env", 15, "Missing '=' separator"),
			contains: []string{
				".env",
				"Line 15",
				"Missing '=' separator",
			},
		},
		{
			name: "error with hint",
			err: NewParseError(".env", 10, "Invalid syntax").
				WithHint("Each line should be in format KEY=VALUE"),
			contains: []string{
				"Line 10",
				"Invalid syntax",
				"Hint:",
				"KEY=VALUE",
			},
		},
		{
			name: "error with context",
			err: func() *ParseError {
				lines := []string{
					"DATABASE_URL=postgres://localhost",
					"API_KEY=secret",
					"BROKEN LINE WITHOUT EQUALS",
					"PORT=3000",
					"DEBUG=true",
				}
				return NewParseError(".env", 3, "Missing '=' separator").
					WithContext(lines, 3, 2)
			}(),
			contains: []string{
				"Line 3",
				"DATABASE_URL",    // context before
				"API_KEY",         // context before
				"BROKEN LINE",     // error line
				"PORT=3000",       // context after
				"DEBUG=true",      // context after
			},
		},
		{
			name: "error with column pointer",
			err: NewParseError(".env", 5, "Unexpected character").
				WithColumn(10),
			contains: []string{
				"Line 5",
				"^", // Column pointer
			},
		},
		{
			name: "error with all features",
			err: func() *ParseError {
				lines := []string{
					"# Configuration file",
					"DATABASE_URL=postgres://localhost",
					"API_KEY=\"unclosed quote",
					"PORT=3000",
				}
				return NewParseError("config/.env", 3, "Unclosed double quote").
					WithContext(lines, 3, 1).
					WithColumn(9).
					WithHint("Quotes must be balanced: KEY=\"value\" or KEY='value'")
			}(),
			contains: []string{
				"config/.env",
				"Line 3",
				"Unclosed double quote",
				"DATABASE_URL",
				"API_KEY",
				"PORT=3000",
				"^",
				"Hint:",
				"balanced",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := tt.err.Error()

			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Error output missing '%s'\nGot:\n%s", expected, output)
				}
			}
		})
	}
}

func TestParseError_FormattingReadability(t *testing.T) {
	lines := []string{
		"# Configuration",
		"DATABASE_URL=postgres://localhost",
		"API_KEY=sk_live_abc123",
		"BROKEN=LINE=WITH=TOO=MANY=EQUALS",
		"PORT=3000",
	}

	err := NewParseError("config/.env", 4, "Multiple '=' separators found").
		WithContext(lines, 4, 2).
		WithHint("Use quotes if value contains '=': KEY=\"value=with=equals\"")

	output := err.Error()

	// Verify formatting
	t.Logf("Error output:\n%s", output)

	// Check structure
	if !strings.Contains(output, "Line 4") {
		t.Error("Missing line number")
	}
	if !strings.Contains(output, "Hint:") {
		t.Error("Missing hint section")
	}
	// Check for DATABASE_URL and API_KEY which are context before the error line
	if !strings.Contains(output, "DATABASE_URL") {
		t.Error("Missing context before error (DATABASE_URL)")
	}
	if !strings.Contains(output, "API_KEY") {
		t.Error("Missing context before error (API_KEY)")
	}
	if !strings.Contains(output, "PORT=3000") {
		t.Error("Missing context after error")
	}
}

func TestParseError_WithContext(t *testing.T) {
	lines := []string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
	}

	tests := []struct {
		name              string
		errorLine         int
		contextSize       int
		wantContextBefore []string
		wantErrorLine     string
		wantContextAfter  []string
	}{
		{
			name:              "middle of file",
			errorLine:         4,
			contextSize:       2,
			wantContextBefore: []string{"line2", "line3"},
			wantErrorLine:     "line4",
			wantContextAfter:  []string{"line5", "line6"},
		},
		{
			name:              "near start",
			errorLine:         2,
			contextSize:       2,
			wantContextBefore: []string{"line1"},
			wantErrorLine:     "line2",
			wantContextAfter:  []string{"line3", "line4"},
		},
		{
			name:              "near end",
			errorLine:         6,
			contextSize:       2,
			wantContextBefore: []string{"line4", "line5"},
			wantErrorLine:     "line6",
			wantContextAfter:  []string{"line7"},
		},
		{
			name:              "at start",
			errorLine:         1,
			contextSize:       2,
			wantContextBefore: []string{},
			wantErrorLine:     "line1",
			wantContextAfter:  []string{"line2", "line3"},
		},
		{
			name:              "at end",
			errorLine:         7,
			contextSize:       2,
			wantContextBefore: []string{"line5", "line6"},
			wantErrorLine:     "line7",
			wantContextAfter:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewParseError("test.env", tt.errorLine, "test error").
				WithContext(lines, tt.errorLine, tt.contextSize)

			// Check context before
			if len(err.ContextBefore) != len(tt.wantContextBefore) {
				t.Errorf("ContextBefore length = %d, want %d", len(err.ContextBefore), len(tt.wantContextBefore))
			}
			for i, line := range tt.wantContextBefore {
				if i >= len(err.ContextBefore) || err.ContextBefore[i] != line {
					t.Errorf("ContextBefore[%d] = %q, want %q", i, err.ContextBefore[i], line)
				}
			}

			// Check error line
			if err.ErrorLine != tt.wantErrorLine {
				t.Errorf("ErrorLine = %q, want %q", err.ErrorLine, tt.wantErrorLine)
			}

			// Check context after
			if len(err.ContextAfter) != len(tt.wantContextAfter) {
				t.Errorf("ContextAfter length = %d, want %d", len(err.ContextAfter), len(tt.wantContextAfter))
			}
			for i, line := range tt.wantContextAfter {
				if i >= len(err.ContextAfter) || err.ContextAfter[i] != line {
					t.Errorf("ContextAfter[%d] = %q, want %q", i, err.ContextAfter[i], line)
				}
			}
		})
	}
}

func TestCommonErrors(t *testing.T) {
	tests := []struct {
		name      string
		errorFunc func() *ParseError
		wantLine  int
		wantHint  string
	}{
		{
			name:      "missing equals",
			errorFunc: func() *ParseError { return MissingEqualsError(".env", 10, "NOEQUALS") },
			wantLine:  10,
			wantHint:  "KEY=VALUE",
		},
		{
			name:      "invalid key",
			errorFunc: func() *ParseError { return InvalidKeyError(".env", 5, "123-invalid") },
			wantLine:  5,
			wantHint:  "A-Z, 0-9, and _",
		},
		{
			name:      "empty key",
			errorFunc: func() *ParseError { return EmptyKeyError(".env", 7) },
			wantLine:  7,
			wantHint:  "KEY=value, not =value",
		},
		{
			name:      "unclosed quote",
			errorFunc: func() *ParseError { return UnclosedQuoteError(".env", 8, 15) },
			wantLine:  8,
			wantHint:  "balanced",
		},
		{
			name:      "yaml indentation",
			errorFunc: func() *ParseError { return YAMLIndentationError("docker-compose.yml", 12) },
			wantLine:  12,
			wantHint:  "indentation",
		},
		{
			name:      "dockerfile instruction - ARGS typo",
			errorFunc: func() *ParseError { return DockerfileInstructionError("Dockerfile", 3, "ARGS") },
			wantLine:  3,
			wantHint:  "Did you mean 'ARG'?",
		},
		{
			name:      "dockerfile instruction - ENVS typo",
			errorFunc: func() *ParseError { return DockerfileInstructionError("Dockerfile", 5, "ENVS") },
			wantLine:  5,
			wantHint:  "Did you mean 'ENV'?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()
			output := err.Error()

			// Check line number
			if err.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", err.Line, tt.wantLine)
			}

			// Check hint
			if !strings.Contains(output, tt.wantHint) {
				t.Errorf("Missing hint '%s' in error:\n%s", tt.wantHint, output)
			}

			// Check that error is formatted properly
			if !strings.Contains(output, "Line") {
				t.Error("Error output missing 'Line' label")
			}
			if !strings.Contains(output, "Hint:") {
				t.Error("Error output missing 'Hint:' label")
			}
		})
	}
}

func TestParseError_WithColumn(t *testing.T) {
	err := NewParseError("test.env", 5, "Unexpected character").
		WithColumn(15)

	if err.Column != 15 {
		t.Errorf("Column = %d, want 15", err.Column)
	}

	output := err.Error()
	if !strings.Contains(output, "^") {
		t.Error("Error output missing column pointer '^'")
	}
}

func TestParseError_WithHint(t *testing.T) {
	hint := "This is a helpful hint"
	err := NewParseError("test.env", 1, "Some error").
		WithHint(hint)

	if err.Hint != hint {
		t.Errorf("Hint = %q, want %q", err.Hint, hint)
	}

	output := err.Error()
	if !strings.Contains(output, hint) {
		t.Errorf("Error output missing hint: %s", hint)
	}
}

func TestParseError_LineNumberAlignment(t *testing.T) {
	lines := []string{
		"line1",
		"line2",
		"line3",
		"line4",
		"line5",
		"line6",
		"line7",
		"line8",
		"line9",
		"line10",
	}

	err := NewParseError("test.env", 5, "test error").
		WithContext(lines, 5, 2)

	output := err.Error()

	// Check that line numbers are present and properly formatted
	// Line numbers should be aligned (e.g., "  3 |", "  4 |", "  5 |")
	expectedLines := []string{" 3 |", " 4 |", " 5 |", " 6 |", " 7 |"}
	for _, linePrefix := range expectedLines {
		if !strings.Contains(output, linePrefix) {
			t.Errorf("Missing line prefix '%s' in output:\n%s", linePrefix, output)
		}
	}
}

func TestParseError_NoContext(t *testing.T) {
	// Error without context should still work
	err := NewParseError("test.env", 5, "test error")

	output := err.Error()

	if !strings.Contains(output, "Line 5") {
		t.Error("Error output missing line number")
	}
	if !strings.Contains(output, "test error") {
		t.Error("Error output missing error message")
	}
}

func TestParseError_ErrorWithColor(t *testing.T) {
	lines := []string{
		"DATABASE_URL=postgres://localhost",
		"API_KEY=secret",
		"BROKEN LINE",
		"PORT=3000",
	}

	err := NewParseError(".env", 3, "Missing '=' separator").
		WithContext(lines, 3, 1).
		WithHint("Each line should be in format KEY=VALUE")

	// Test without color
	outputNoColor := err.Error()
	if strings.Contains(outputNoColor, "\033[") {
		t.Error("Error() should not contain ANSI color codes")
	}

	// Test with color
	outputWithColor := err.ErrorWithColor()
	if !strings.Contains(outputWithColor, "\033[") {
		t.Error("ErrorWithColor() should contain ANSI color codes")
	}

	// Both should contain the same text content
	if !strings.Contains(outputWithColor, "Line 3") {
		t.Error("Colored output missing line number")
	}
	if !strings.Contains(outputWithColor, "Missing '=' separator") {
		t.Error("Colored output missing error message")
	}
}
