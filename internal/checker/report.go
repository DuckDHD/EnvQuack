package checker

import (
	"fmt"
	"strings"

	"github.com/DuckDHD/EnvQuack/internal/quack"
)

// ReportOptions controls report formatting
type ReportOptions struct {
	ShowDuck bool
	Colorize bool
	Verbose  bool
}

// DefaultReportOptions returns sensible defaults
func DefaultReportOptions() *ReportOptions {
	return &ReportOptions{
		ShowDuck: true,
		Colorize: true,
		Verbose:  false,
	}
}

// GenerateReport creates a formatted report from the diff result
func GenerateReport(result *DiffResult, opts *ReportOptions) string {
	if opts == nil {
		opts = DefaultReportOptions()
	}

	var report strings.Builder

	if !result.HasIssues() {
		report.WriteString("✅ All envs aligned.\n")
		if opts.ShowDuck {
			report.WriteString("(Your gopher-duck is calm and happy.)\n")
		}
		return report.String()
	}

	// Header with duck
	if opts.ShowDuck {
		report.WriteString(quack.GetAngryDuck() + "\n")
		report.WriteString("QUACK! 🦆 Environment issues detected:\n\n")
	}

	// Missing variables
	if len(result.Missing) > 0 {
		if opts.Colorize {
			report.WriteString("🔴 Missing variables (present in .env.example but not in .env):\n")
		} else {
			report.WriteString("Missing variables:\n")
		}

		for _, key := range result.Missing {
			report.WriteString(fmt.Sprintf("  - %s\n", key))
		}
		report.WriteString("\n")
	}

	// Extra variables
	if len(result.Extra) > 0 {
		if opts.Colorize {
			report.WriteString("🟡 Extra variables (present in .env but not in .env.example):\n")
		} else {
			report.WriteString("Extra variables:\n")
		}

		for _, key := range result.Extra {
			report.WriteString(fmt.Sprintf("  - %s\n", key))
		}
		report.WriteString("\n")
	}

	// Footer with duck message
	if opts.ShowDuck {
		report.WriteString("(Your gopher-duck is angry. Fix your .env!)\n")
	}

	return report.String()
}

// GenerateSummary creates a brief summary of issues
func GenerateSummary(result *DiffResult) string {
	if !result.HasIssues() {
		return "No issues found"
	}

	parts := []string{}
	if len(result.Missing) > 0 {
		parts = append(parts, fmt.Sprintf("%d missing", len(result.Missing)))
	}
	if len(result.Extra) > 0 {
		parts = append(parts, fmt.Sprintf("%d extra", len(result.Extra)))
	}

	return strings.Join(parts, ", ")
}
