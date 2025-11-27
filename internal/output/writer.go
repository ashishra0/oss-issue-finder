package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ashishra0/issue-finder/pkg/types"
)

// WriteMarkdownFile writes matches to a markdown file
func WriteMarkdownFile(outputPath string, state types.State) error {
	var sb strings.Builder

	sb.WriteString("# GitHub OSS Contribution Opportunities\n\n")
	sb.WriteString(fmt.Sprintf("Last updated: %s\n\n", time.Now().Format("2006-01-02 15:04")))
	sb.WriteString(fmt.Sprintf("Total opportunities: %d\n\n", len(state.AllMatches)))
	sb.WriteString("---\n\n")

	if len(state.AllMatches) == 0 {
		sb.WriteString("No OSS opportunities found yet. Check back later!\n")
	} else {
		for _, match := range state.AllMatches {
			sb.WriteString(fmt.Sprintf("## [%s] %s\n\n", match.Repo, match.Title))

			sb.WriteString(fmt.Sprintf("- **URL**: %s\n", match.URL))
			sb.WriteString(fmt.Sprintf("- **Effort**: %s\n", match.Effort))
			sb.WriteString(fmt.Sprintf("- **Created**: %s\n", match.CreatedAt))
			sb.WriteString(fmt.Sprintf("- **Found**: %s\n", match.FoundAt))

			if len(match.Labels) > 0 {
				sb.WriteString(fmt.Sprintf("- **Labels**: %s\n", strings.Join(match.Labels, ", ")))
			}

			sb.WriteString(fmt.Sprintf("\n**Why this fits you:**\n%s\n\n", match.MatchReason))

			sb.WriteString("---\n\n")
		}
	}

	err := os.WriteFile(outputPath, []byte(sb.String()), 0644)
	if err != nil {
		return fmt.Errorf("error writing markdown file: %w", err)
	}

	return nil
}
