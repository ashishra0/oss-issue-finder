package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/ashishra0/issue-finder/pkg/types"
)

type Evaluator struct {
	client *anthropic.Client
}

func NewEvaluator(apiKey string) *Evaluator {
	return &Evaluator{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
	}
}

// EvaluateIssues sends issues to Claude for evaluation and returns matches
func (e *Evaluator) EvaluateIssues(profile types.UserProfile, issues []map[string]interface{}) []types.IssueMatch {
	ctx := context.Background()

	profileJSON, _ := json.Marshal(profile)
	issuesJSON, _ := json.Marshal(issues)

	prompt := fmt.Sprintf(`You are helping find GitHub OSS contribution opportunities for a developer with %d years of experience.

Developer Profile:
%s

GitHub Issues to evaluate:
%s

Your task: Carefully evaluate each issue and return ONLY the best 3-5 matches that would be genuinely good first contributions.

Selection criteria (ALL must be met):
1. Clear scope: The issue has a well-defined problem and expected outcome
2. Skill match: Requires skills the developer has
3. Appropriate complexity: Not trivial, but achievable in a few hours to a day
4. Active project: The issue has recent activity and the project seems maintained
5. Welcoming: Issue description is friendly and provides context
6. Realistic: Avoid issues that are too vague, too large, or require deep domain knowledge

For each match, provide a SPECIFIC reason explaining:
- What skill(s) from their profile apply
- Why the complexity level is appropriate
- What makes this a good first contribution to this project

Return ONLY matching issues in this JSON format:
{
  "matches": [
    {
      "repo": "owner/repo-name",
      "issue_number": 123,
      "title": "Issue title",
      "url": "https://github.com/...",
      "match_reason": "Specific explanation: which skills apply, why it's good scope, what makes it welcoming",
      "estimated_effort": "small|medium|large",
      "labels": ["label1", "label2"],
      "created_at": "2024-01-01"
    }
  ]
}

Be VERY selective - quality over quantity. Return at most 5 matches. If no issues are genuinely good fits, return empty matches array.`,
		profile.ExperienceYears, string(profileJSON), string(issuesJSON))

	response, err := e.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.F("claude-sonnet-4-5-20250929"),
		MaxTokens: anthropic.F(int64(4096)),
		Messages: anthropic.F([]anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		}),
	})

	if err != nil {
		log.Printf("Error calling Claude: %v", err)
		return []types.IssueMatch{}
	}

	contentText := response.Content[0].Text

	jsonText := extractJSON(contentText)

	var result struct {
		Matches []types.IssueMatch `json:"matches"`
	}

	err = json.Unmarshal([]byte(jsonText), &result)
	if err != nil {
		log.Printf("Error parsing Claude response: %v", err)
		log.Printf("Response text: %s", contentText)
		return []types.IssueMatch{}
	}

	now := time.Now().Format("2006-01-02 15:04")
	for i := range result.Matches {
		result.Matches[i].FoundAt = now
	}

	return result.Matches
}

// extractJSON removes markdown code block wrapping from JSON responses
func extractJSON(text string) string {
	text = strings.TrimSpace(text)

	if strings.HasPrefix(text, "```") {
		startIdx := strings.Index(text, "\n")
		if startIdx == -1 {
			return text
		}

		endIdx := strings.LastIndex(text, "```")
		if endIdx == -1 || endIdx <= startIdx {
			return text
		}

		return strings.TrimSpace(text[startIdx+1 : endIdx])
	}

	return text
}
