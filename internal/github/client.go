package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ashishra0/issue-finder/pkg/types"
)

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchRelevantIssues fetches issues based on profile - multiple targeted queries
func (gc *Client) FetchRelevantIssues(profile types.UserProfile) []map[string]any {
	queries := gc.buildSearchQueries(profile)

	allIssues := []map[string]any{}
	seenIssueURLs := make(map[string]bool)

	for i, query := range queries {
		if i > 0 {
			time.Sleep(2 * time.Second)
		}

		issues := gc.searchIssues(query)

		for _, issue := range issues {
			issueURL := issue.URL

			if seenIssueURLs[issueURL] {
				continue
			}

			seenIssueURLs[issueURL] = true

			labelNames := []string{}
			for _, label := range issue.Labels {
				labelNames = append(labelNames, label.Name)
			}

			repoName := gc.extractRepoName(issue.RepoURL)

			body := issue.Body
			if len(body) > 500 {
				body = body[:500] + "... [truncated]"
			}

			allIssues = append(allIssues, map[string]any{
				"repo":       repoName,
				"number":     issue.Number,
				"title":      issue.Title,
				"url":        issue.URL,
				"labels":     labelNames,
				"body":       body,
				"created_at": issue.CreatedAt.Format("2006-01-02"),
			})
		}
	}

	return allIssues
}

// buildSearchQueries builds targeted search queries based on skills and interests
func (gc *Client) buildSearchQueries(profile types.UserProfile) []string {
	queries := []string{}

	baseConstraints := "is:issue is:open no:assignee created:>2024-10-01 comments:>=1"
	goodLabels := []string{"good first issue", "help wanted"}

	for _, skill := range profile.Skills {
		skillLower := strings.ToLower(skill)
		languageOrTopic := gc.mapSkillToGitHub(skillLower)

		for _, label := range goodLabels {
			query := fmt.Sprintf("%s %s label:\"%s\"", baseConstraints, languageOrTopic, label)
			queries = append(queries, query)
		}
	}

	return queries
}

func (gc *Client) mapSkillToGitHub(skill string) string {
	mappings := map[string]string{
		"python":         "language:python",
		"ruby on rails":  "language:ruby topic:rails",
		"ruby":           "language:ruby",
		"go":             "language:go",
		"golang":         "language:go",
		"rust":           "language:rust",
		"javascript":     "language:javascript",
		"typescript":     "language:typescript",
		"java":           "language:java",
		"c++":            "language:c++",
		"c":              "language:c",
		"php":            "language:php",
		"swift":          "language:swift",
		"kotlin":         "language:kotlin",
		"postgresql":     "topic:postgresql",
		"postgres":       "topic:postgresql",
		"sqlite":         "topic:sqlite",
		"mysql":          "topic:mysql",
		"mongodb":        "topic:mongodb",
		"redis":          "topic:redis",
		"message queues": "topic:message-queue OR topic:rabbitmq OR topic:kafka",
		"event-driven":   "topic:event-driven",
	}

	mapped := mappings[skill]
	if mapped != "" {
		return mapped
	}

	return fmt.Sprintf("topic:%s", skill)
}

func (gc *Client) searchIssues(query string) []types.GitHubIssue {
	searchURL := "https://api.github.com/search/issues?q="

	params := url.Values{}
	params.Add("q", query)
	params.Add("sort", "created")
	params.Add("order", "desc")
	params.Add("per_page", "30")

	fullURL := fmt.Sprintf("%s?%s", searchURL, params.Encode())

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return []types.GitHubIssue{}
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", gc.token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		fmt.Printf("Error executing request: %v\n", err)
		return []types.GitHubIssue{}
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		fmt.Println("GitHub authentication failed (401 Unauthorized)")
		fmt.Println("Your GITHUB_TOKEN may be invalid or expired.")
		fmt.Println("Create a new token at: https://github.com/settings/tokens")
		return []types.GitHubIssue{}
	}

	if resp.StatusCode == 403 {
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)

		if strings.Contains(bodyStr, "secondary rate limit") {
			fmt.Println("Hit GitHub secondary rate limit (too many requests too quickly)")
			fmt.Println("Please wait 5-10 minutes before trying again")
		} else {
			rateLimitRemaining := resp.Header.Get("X-RateLimit-Remaining")
			if rateLimitRemaining == "0" {
				fmt.Println("GitHub primary rate limit exceeded")
				fmt.Printf("Resets at: %s\n", resp.Header.Get("X-RateLimit-Reset"))
			} else {
				fmt.Println("GitHub API access denied (403)")
			}
		}
		return []types.GitHubIssue{}
	}

	if resp.StatusCode != 200 {
		fmt.Printf("GitHub API error: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body))
		return []types.GitHubIssue{}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return []types.GitHubIssue{}
	}

	var result struct {
		Items []types.GitHubIssue `json:"items"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return []types.GitHubIssue{}
	}

	return result.Items
}

func (gc *Client) extractRepoName(repoURL string) string {
	parts := strings.Split(repoURL, "/")

	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1]
	}

	return "unknown"
}
