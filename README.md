# OSS Issue Finder

A CLI tool that helps developers find suitable GitHub issues to contribute to based on their skills and experience.

## How It Works

1. The tool searches GitHub for issues matching your criteria (language, labels, experience level)
2. Uses Claude AI to evaluate each issue against your developer profile
3. Returns the best 3-5 matches with specific reasons why they're good fits
4. Results are saved to a markdown file for easy reference

## Setup

### Prerequisites

- Go 1.23 or higher
- GitHub Personal Access Token
- Anthropic API Key (for Claude)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the binary:
   ```bash
   go build -o issue-finder
   ```

### Configuration

Create a configuration file at `~/.issue-finder.yaml`:

```yaml
github_token: "your-github-token"
anthropic_api_key: "your-anthropic-api-key"
```

Or set environment variables:
```bash
export GITHUB_TOKEN="your-github-token"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
```

## Usage

Run the search command with your preferences:

```bash
./issue-finder search \
  --skills "Go,Python,PostgreSQL" \
  --interests "Backend,Databases" \
  --experience 5
```

### Parameters

- `--skills`: Your technical skills (comma-separated, required)
- `--interests`: Your areas of interest (comma-separated, optional)
- `--experience`: Years of experience (affects complexity matching)
- `--output`: Output file path (default: ~/contributions.md)
- `--state`: State file path to track processed issues (default: ~/.issue-finder-state.json)
- `--no-notify`: Disable desktop notifications

### Output

Results are saved to `~/contributions.md` by default, or to the path specified with `--output`.

Each match includes:
- Issue title and link
- Repository information
- Specific reason why it matches your profile
- Estimated effort (small/medium/large)
- Labels and creation date
