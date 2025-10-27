package errors

import (
	"fmt"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/masking"
)

// ParseError represents a parsing error with location information
type ParseError struct {
	Filename     string
	Line         int
	Column       int
	Message      string
	Hint         string
	ContextBefore []string // Lines before the error
	ErrorLine    string   // The actual line with the error
	ContextAfter []string // Lines after the error
}

// Error implements the error interface
func (e *ParseError) Error() string {
	return e.format(false)
}

// ErrorWithColor returns the error message with color formatting
func (e *ParseError) ErrorWithColor() string {
	return e.format(true)
}

// format produces the formatted error output
func (e *ParseError) format(useColor bool) string {
	var sb strings.Builder

	// Color functions (no-op if colors disabled)
	red := func(s string) string { return s }
	yellow := func(s string) string { return s }
	cyan := func(s string) string { return s }
	gray := func(s string) string { return s }

	if useColor {
		// ANSI color codes
		red = func(s string) string { return fmt.Sprintf("\033[31m%s\033[0m", s) }
		yellow = func(s string) string { return fmt.Sprintf("\033[33m%s\033[0m", s) }
		cyan = func(s string) string { return fmt.Sprintf("\033[36m%s\033[0m", s) }
		gray = func(s string) string { return fmt.Sprintf("\033[90m%s\033[0m", s) }
	}

	// Header: filename and error type
	sb.WriteString(red(fmt.Sprintf("Failed to parse %s\n", e.Filename)))
	sb.WriteString("\n")

	// Line number and message
	sb.WriteString(yellow(fmt.Sprintf("  Line %d: ", e.Line)))
	sb.WriteString(e.Message)
	sb.WriteString("\n")
	sb.WriteString("\n")

	// Context: lines before error
	if len(e.ContextBefore) > 0 {
		startLine := e.Line - len(e.ContextBefore)
		for i, line := range e.ContextBefore {
			lineNum := startLine + i
			sb.WriteString(gray(fmt.Sprintf("  %3d | %s\n", lineNum, line)))
		}
	}

	// Error line with highlight (mask sensitive values)
	if e.ErrorLine != "" {
		displayLine := e.ErrorLine

		// If line contains KEY=VALUE, mask the value if sensitive
		if idx := strings.Index(displayLine, "="); idx > 0 {
			key := strings.TrimSpace(displayLine[:idx])
			value := strings.TrimSpace(displayLine[idx+1:])

			// Mask if key or value appears sensitive
			if masking.IsSensitive(key) || len(value) > 0 {
				maskedValue := masking.MaskIfSensitive(key, value)
				displayLine = key + "=" + maskedValue
			}
		}

		sb.WriteString(fmt.Sprintf("  %3d | %s\n", e.Line, displayLine))
	}

	// Column pointer (if specified)
	if e.Column > 0 {
		padding := 7 + e.Column // "  NNN | " = 7 chars + column position
		pointer := strings.Repeat(" ", padding) + "^"
		sb.WriteString(red(pointer))
		sb.WriteString("\n")
	}

	// Context: lines after error
	if len(e.ContextAfter) > 0 {
		startLine := e.Line + 1
		for i, line := range e.ContextAfter {
			lineNum := startLine + i
			sb.WriteString(gray(fmt.Sprintf("  %3d | %s\n", lineNum, line)))
		}
	}

	// Hint (if provided)
	if e.Hint != "" {
		sb.WriteString("\n")
		sb.WriteString(cyan("  Hint: "))
		sb.WriteString(e.Hint)
		sb.WriteString("\n")
	}

	return sb.String()
}

// NewParseError creates a new parse error
func NewParseError(filename string, line int, message string) *ParseError {
	return &ParseError{
		Filename: filename,
		Line:     line,
		Message:  message,
	}
}

// WithHint adds a helpful hint to the error
func (e *ParseError) WithHint(hint string) *ParseError {
	e.Hint = hint
	return e
}

// WithContext adds surrounding lines for context
func (e *ParseError) WithContext(allLines []string, errorLine int, contextSize int) *ParseError {
	// Extract lines around the error
	// errorLine is 1-indexed for user display
	errorIdx := errorLine - 1 // Convert to 0-indexed

	// Context before
	start := errorIdx - contextSize
	if start < 0 {
		start = 0
	}
	if start < errorIdx {
		e.ContextBefore = allLines[start:errorIdx]
	}

	// Error line itself
	if errorIdx >= 0 && errorIdx < len(allLines) {
		e.ErrorLine = allLines[errorIdx]
	}

	// Context after
	end := errorIdx + contextSize + 1
	if end > len(allLines) {
		end = len(allLines)
	}
	if errorIdx+1 < end {
		e.ContextAfter = allLines[errorIdx+1 : end]
	}

	return e
}

// WithColumn sets the column number
func (e *ParseError) WithColumn(col int) *ParseError {
	e.Column = col
	return e
}

// Common error constructors for typical parsing issues

// MissingEqualsError creates an error for missing '=' separator in key=value format
func MissingEqualsError(filename string, line int, content string) *ParseError {
	return NewParseError(filename, line, "Missing '=' separator").
		WithHint("Each line should be in format KEY=VALUE")
}

// InvalidKeyError creates an error for invalid variable names
func InvalidKeyError(filename string, line int, key string) *ParseError {
	return NewParseError(filename, line, fmt.Sprintf("Invalid variable name '%s'", key)).
		WithHint("Variable names should contain only A-Z, 0-9, and _ (uppercase recommended)")
}

// EmptyKeyError creates an error for empty variable names
func EmptyKeyError(filename string, line int) *ParseError {
	return NewParseError(filename, line, "Empty variable name before '='").
		WithHint("Format should be KEY=value, not =value")
}

// UnclosedQuoteError creates an error for unclosed quotes in values
func UnclosedQuoteError(filename string, line int, col int) *ParseError {
	return NewParseError(filename, line, "Unclosed quote").
		WithColumn(col).
		WithHint("Quotes must be balanced: KEY=\"value\" or KEY='value'")
}

// YAMLIndentationError creates an error for YAML indentation issues
func YAMLIndentationError(filename string, line int) *ParseError {
	return NewParseError(filename, line, "Invalid YAML indentation").
		WithHint("YAML requires consistent indentation (use 2 or 4 spaces, not tabs)")
}

// DockerfileInstructionError creates an error for invalid Dockerfile instructions
func DockerfileInstructionError(filename string, line int, instruction string) *ParseError {
	hint := "See Dockerfile reference: https://docs.docker.com/engine/reference/builder/"

	// Smart suggestions for common typos
	switch strings.ToUpper(instruction) {
	case "ARGS":
		hint = "Did you mean 'ARG'? " + hint
	case "ENVS":
		hint = "Did you mean 'ENV'? " + hint
	case "RUNS":
		hint = "Did you mean 'RUN'? " + hint
	case "COPYS":
		hint = "Did you mean 'COPY'? " + hint
	case "FROMS":
		hint = "Did you mean 'FROM'? " + hint
	}

	return NewParseError(filename, line, fmt.Sprintf("Invalid instruction '%s'", instruction)).
		WithHint(hint)
}
