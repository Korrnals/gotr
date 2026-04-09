// internal/selftest/types.go
// Types and interfaces for the self-test command.
package selftest

import (
	"fmt"
	"time"
)

// Result represents the status of a check.
type Result string

const (
	ResultPass Result = "PASS" // passed
	ResultFail Result = "FAIL" // failed
	ResultWarn Result = "WARN" // warning
	ResultSkip Result = "SKIP" // skipped
)

// String returns the emoji representation of the result.
func (r Result) String() string {
	switch r {
	case ResultPass:
		return "✓"
	case ResultFail:
		return "✗"
	case ResultWarn:
		return "⚠"
	case ResultSkip:
		return "⊘"
	default:
		return "?"
	}
}

// Color returns the ANSI color code for the result.
func (r Result) Color() string {
	switch r {
	case ResultPass:
		return "\033[32m" // green
	case ResultFail:
		return "\033[31m" // red
	case ResultWarn:
		return "\033[33m" // yellow
	case ResultSkip:
		return "\033[90m" // gray
	default:
		return "\033[0m"
	}
}

// ResetColor returns the ANSI reset sequence.
func (r Result) ResetColor() string {
	return "\033[0m"
}

// CheckResult represents the result of a single check.
type CheckResult struct {
	Name       string        `json:"name"`        // check name
	Category   string        `json:"category"`    // category (config, api, tests, etc.)
	Result     Result        `json:"result"`      // status
	Message    string        `json:"message"`     // human-readable message
	Details    string        `json:"details"`     // details (optional)
	Duration   time.Duration `json:"duration"`    // execution time
	Error      error         `json:"-"`           // error (not serialized)
	CanFix     bool          `json:"can_fix"`     // whether auto-fix is possible
	FixCommand string        `json:"fix_command"` // command to fix the issue
}

// Report represents a complete self-test report.
type Report struct {
	Timestamp   time.Time     `json:"timestamp"`  // run time
	Version     string        `json:"version"`    // gotr version
	Commit      string        `json:"commit"`     // git commit
	GoVersion   string        `json:"go_version"` // Go version
	Platform    string        `json:"platform"`   // OS/Arch
	Checks      []CheckResult `json:"checks"`     // all checks
	Duration    time.Duration `json:"duration"`   // total duration
	Health      Result        `json:"health"`     // overall status
	TotalPassed int           `json:"total_passed"`
	TotalFailed int           `json:"total_failed"`
	TotalWarn   int           `json:"total_warn"`
	TotalSkip   int           `json:"total_skip"`
}

// CalculateHealth computes the overall health status.
func (r *Report) CalculateHealth() {
	hasFail := false
	hasWarn := false

	for _, c := range r.Checks {
		switch c.Result {
		case ResultPass:
			r.TotalPassed++
		case ResultFail:
			r.TotalFailed++
			hasFail = true
		case ResultWarn:
			r.TotalWarn++
			hasWarn = true
		case ResultSkip:
			r.TotalSkip++
		}
	}

	if hasFail {
		r.Health = ResultFail
	} else if hasWarn {
		r.Health = ResultWarn
	} else {
		r.Health = ResultPass
	}
}

// OverallStatus returns the health status as a human-readable string.
func (r *Report) OverallStatus() string {
	switch r.Health {
	case ResultPass:
		return "Healthy"
	case ResultWarn:
		return "Degraded"
	case ResultFail:
		return "Unhealthy"
	default:
		return "Unknown"
	}
}

// Checker is the interface for self-test checks.
type Checker interface {
	Name() string
	Category() string
	Check() CheckResult
}

// Runner executes all registered checks.
type Runner struct {
	checkers []Checker
}

// NewRunner creates a new Runner.
func NewRunner() *Runner {
	return &Runner{
		checkers: make([]Checker, 0),
	}
}

// Register adds a check to the runner.
func (r *Runner) Register(c Checker) {
	r.checkers = append(r.checkers, c)
}

// Run executes all registered checks and returns a report.
func (r *Runner) Run() *Report {
	report := &Report{
		Timestamp: time.Now(),
		Checks:    make([]CheckResult, 0, len(r.checkers)),
	}

	start := time.Now()

	for _, checker := range r.checkers {
		checkStart := time.Now()
		result := checker.Check()
		result.Duration = time.Since(checkStart)
		result.Name = checker.Name()
		result.Category = checker.Category()
		report.Checks = append(report.Checks, result)
	}

	report.Duration = time.Since(start)
	report.CalculateHealth()

	return report
}

// PrintHuman prints the report in human-readable format.
func (r *Report) PrintHuman() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    gotr Self-Test Report                     ║")
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Version:  %-50s║\n", r.Version)
	fmt.Printf("║  Commit:   %-50s║\n", r.Commit)
	fmt.Printf("║  Go:       %-50s║\n", r.GoVersion)
	fmt.Printf("║  Platform: %-50s║\n", r.Platform)
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")

	currentCategory := ""
	for _, check := range r.Checks {
		if check.Category != currentCategory {
			currentCategory = check.Category
			fmt.Printf("\n[%s]\n", currentCategory)
		}
		fmt.Printf("  %s %s %s\n",
			check.Result.Color()+check.Result.String()+check.Result.ResetColor(),
			check.Name,
			formatDetails(check),
		)
		if check.Details != "" {
			fmt.Printf("      %s\n", check.Details)
		}
	}

	fmt.Println()
	fmt.Println("╠══════════════════════════════════════════════════════════════╣")
	fmt.Printf("║  Results:  %s %d passed  %s %d failed  %s %d warn  %s %d skip   ║\n",
		ResultPass.String(), r.TotalPassed,
		ResultFail.String(), r.TotalFailed,
		ResultWarn.String(), r.TotalWarn,
		ResultSkip.String(), r.TotalSkip,
	)
	fmt.Printf("║  Overall:  %s%-56s%s║\n",
		r.Health.Color(), r.OverallStatus(), r.Health.ResetColor())
	fmt.Printf("║  Duration: %-50s║\n", r.Duration.Round(time.Millisecond))
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func formatDetails(c CheckResult) string {
	if c.Error != nil {
		return "— " + c.Error.Error()
	}
	if c.Message != "" {
		return "— " + c.Message
	}
	return ""
}
