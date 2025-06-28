# PR Analyzer

A modern CLI tool for fetching and analyzing GitHub pull request data. It features a beautiful terminal UI with progress indicators, transparent caching, and exports data in analysis-ready formats.

## Features

- 🚀 **Single command interface** - Just `pr-analyzer owner/repo`
- 🎨 **Beautiful progress display** - Animated spinners, emojis, and fancy box drawing
- 💾 **Transparent caching** - SQLite-based local cache for instant subsequent runs
- 📊 **Multiple export formats** - JSONL for DuckDB analysis, CSV for spreadsheets
- 🤖 **Bot detection** - Automatically identifies and flags bot accounts
- 🔄 **Incremental updates** - Fetches only new/updated PRs after initial sync
- 🏢 **GitHub Enterprise support** - Configure custom API endpoints
- 🎯 **Flexible filtering** - Limit by count, date range, or specific PR numbers

## Requirements

- Go 1.21 or higher
- GitHub personal access token with `repo` scope (or `public_repo` for public repositories)
- [DuckDB](https://duckdb.org/) (optional, for data analysis)

## Installation

```bash
# Clone the repository
git clone https://github.com/bonyuta0204/pr-analyzer
cd pr-analyzer

# Build from source
make build

# Or install directly
make install
```

### GitHub Token Setup

Create a personal access token at https://github.com/settings/tokens with the following scopes:
- `repo` - Full control of private repositories (or `public_repo` for public repos only)
- `read:org` - Read org and team membership (optional, for better user info)

## Quick Start

1. **Set up your GitHub token:**

```bash
# Create a .env file with your GitHub token
cp .env.example .env
# Edit .env and add your GitHub token

# Or use environment variable
export GITHUB_TOKEN=your_github_token_here
```

2. **Analyze a repository:**

```bash
# Basic usage - exports recent 100 PRs to JSONL
pr-analyzer microsoft/vscode

# Limit to recent 50 PRs
pr-analyzer microsoft/vscode --limit 50

# Export to CSV format
pr-analyzer facebook/react --format csv

# Fetch all PRs (no limit)
pr-analyzer kubernetes/kubernetes --all

# Fetch PRs since a specific date
pr-analyzer golang/go --since 2024-01-01

# Fetch a specific PR
pr-analyzer torvalds/linux --pr 12345
```

3. **Analyze the exported data:**

```bash
# Query with DuckDB
duckdb -c "SELECT * FROM 'microsoft-vscode-prs.jsonl' WHERE state = 'open'"

# Or open in your favorite data analysis tool
```

## Output Examples

### Terminal UI

```
┌─ 🔍 Analyzing microsoft/vscode
│
├─ 📋 Checking cache...
│  ✓ Found 145 PRs cached (last sync: 2 hours ago)
│
├─ 📡 Fetching from GitHub...
│  ├─ Recent PRs................ ✓ 55 new
│  ├─ Reviews................... ✓ 1,234 items
│  ├─ Comments.................. ✓ 5,678 items
│  └─ Files..................... ✓ 2,345 items
│
└─ ✅ Exported 200 PRs → microsoft-vscode-prs.jsonl (2.4 MB)

🎉 Analysis ready! Try: duckdb -c "SELECT * FROM 'microsoft-vscode-prs.jsonl'"
```

## Command Reference

```bash
pr-analyzer <owner/repo> [flags]
```

### Flags

- `--format string` - Export format: jsonl, csv (default "jsonl")
- `--limit int` - Fetch recent N PRs (default 100, use --all for unlimited)
- `--all` - Fetch all PRs (overrides --limit)
- `--include-diffs` - Include file diffs in export
- `--refetch` - Force refetch all data (ignore cache)
- `--since string` - Fetch PRs updated since date (YYYY-MM-DD)
- `--pr int` - Fetch specific PR number only
- `--output string` - Custom output filename
- `-h, --help` - Help for pr-analyzer

### Environment Variables

- `GITHUB_TOKEN` - GitHub personal access token (required)
- `GITHUB_API_URL` - GitHub Enterprise API URL (optional)

## Data Formats

### JSONL Export

Each line contains a complete PR with all associated data:

```json
{
  "type": "pull",
  "number": 123,
  "title": "Add feature X",
  "author": {"login": "user1", "type": "User", "is_bot": false},
  "created_at": "2024-01-01T10:00:00Z",
  "stats": {"additions": 150, "deletions": 30, "changed_files": 5},
  "files": [...],
  "reviews": [...],
  "comments": [...]
}
```

### CSV Export

Exports are split into multiple CSV files:
- `pulls.csv` - Pull request metadata
- `reviews.csv` - Review data
- `comments.csv` - Comment data
- `files.csv` - File change data

## Analyzing Data with DuckDB

The JSONL format is optimized for analysis with DuckDB:

```sql
-- Basic queries
SELECT COUNT(*) as total_prs FROM 'repo-prs.jsonl';

SELECT state, COUNT(*) as count 
FROM 'repo-prs.jsonl' 
GROUP BY state;

-- Find most active reviewers
SELECT 
    reviewer.login as reviewer,
    COUNT(*) as review_count
FROM 'repo-prs.jsonl' t,
    UNNEST(t.reviews) as review,
    LATERAL (SELECT review.reviewer) as reviewer
GROUP BY reviewer.login
ORDER BY review_count DESC
LIMIT 10;

-- Analyze PR merge times
SELECT 
    author.login as author,
    AVG(EPOCH(merged_at) - EPOCH(created_at))/86400 as avg_days_to_merge,
    COUNT(*) as merged_prs
FROM 'repo-prs.jsonl'
WHERE merged_at IS NOT NULL
GROUP BY author.login
HAVING COUNT(*) > 5
ORDER BY avg_days_to_merge;

-- Find files with most comments
SELECT 
    comment.file_path,
    COUNT(*) as comment_count,
    COUNT(DISTINCT number) as pr_count
FROM 'repo-prs.jsonl' t,
    UNNEST(t.comments) as comment
WHERE comment.file_path IS NOT NULL
GROUP BY comment.file_path
ORDER BY comment_count DESC
LIMIT 20;
```

## Output Filenames

The tool generates smart filenames based on your query:

- `owner-repo-prs.jsonl` - Default output
- `owner-repo-prs-recent-50.jsonl` - When using --limit
- `owner-repo-pr-123.jsonl` - When fetching specific PR
- `owner-repo-prs-2024-01.jsonl` - When using --since

## Performance

- **Caching**: After initial fetch, subsequent runs are instant (reads from local cache)
- **Incremental updates**: Only fetches new/updated PRs
- **Rate limiting**: Respects GitHub API rate limits automatically
- **Typical performance**:
  - Small repos (< 1000 PRs): 1-2 minutes initial fetch
  - Medium repos (1000-5000 PRs): 5-10 minutes initial fetch
  - Large repos (> 5000 PRs): Use --limit or --since for faster results

## Troubleshooting

### "GITHUB_TOKEN environment variable is required"
Create a `.env` file with your token or export it:
```bash
export GITHUB_TOKEN=your_token_here
```

### "repository not found"
- Check the repository name format: `owner/repo`
- Ensure your token has access to the repository
- For private repos, ensure `repo` scope is enabled

### Timeout issues
- Use `--limit` to fetch fewer PRs
- Use `--since` to fetch only recent PRs
- Check your internet connection

## Development

```bash
# Setup development environment
make dev-setup

# Run tests
make test

# Run linter
make lint

# Build for development
make dev
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT
