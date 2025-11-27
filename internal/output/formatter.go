package output

import (
	"fmt"
	"time"
)

type ProgressFormatter struct {
	quiet       bool
	currentStep int
	totalSteps  int
	startTime   time.Time
}

func NewProgressFormatter(quiet bool) *ProgressFormatter {
	return &ProgressFormatter{
		quiet:      quiet,
		totalSteps: 4,
		startTime:  time.Now(),
	}
}

func (pf *ProgressFormatter) PrintHeader(profile string, skills []string, interests []string, experience int) {
	if pf.quiet {
		return
	}

	fmt.Println("Issue Finder v1.0.0")
	fmt.Println()
	fmt.Println("Profile:")
	fmt.Printf("  Skills: %s\n", joinStrings(skills, ", "))
	if len(interests) > 0 {
		fmt.Printf("  Interests: %s\n", joinStrings(interests, ", "))
	}
	fmt.Printf("  Experience: %d years\n", experience)
	fmt.Println()
}

func (pf *ProgressFormatter) Step(step int, message string) {
	if pf.quiet {
		return
	}

	pf.currentStep = step
	fmt.Printf("[%d/%d] %s\n", step, pf.totalSteps, message)
}

func (pf *ProgressFormatter) Detail(message string) {
	if pf.quiet {
		return
	}

	fmt.Printf("      %s\n", message)
}

func (pf *ProgressFormatter) EmptyLine() {
	if pf.quiet {
		return
	}

	fmt.Println()
}

func (pf *ProgressFormatter) Success(message string) {
	if pf.quiet {
		return
	}

	elapsed := time.Since(pf.startTime)
	fmt.Printf("\n%s\n", message)
	fmt.Printf("Completed in %.1f seconds\n", elapsed.Seconds())
}

func (pf *ProgressFormatter) Error(message string) {
	fmt.Printf("Error: %s\n", message)
}

func (pf *ProgressFormatter) Warning(message string) {
	if pf.quiet {
		return
	}

	fmt.Printf("Warning: %s\n", message)
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}

	return result
}
