# PR Analyzer

A CLI tool for fetching and analyzing GitHub pull request data. It caches PR data locally and exports it in various formats for analysis.

## Features

- Fetch pull requests, reviews, comments, and file changes from GitHub
- Cache data locally using SQLite for fast access
- Export data in JSONL or CSV format
- Incremental sync to fetch only updated PRs
- Support for GitHub Enterprise
- Bot detection for filtering automated accounts
- Configurable rate limiting and batch processing

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

## Quick Start

1. **Initialize configuration with your GitHub token:**

```bash
pr-analyzer init --token <your-github-token>
```

2. **Fetch PR data from a repository:**

```bash
# Fetch all PRs
pr-analyzer fetch --repo owner/repo

# Fetch PRs updated since a specific date
pr-analyzer fetch --repo owner/repo --since 2024-01-01

# Fetch a specific PR
pr-analyzer fetch --repo owner/repo --pr 123
```

3. **Export data for analysis:**

```bash
# Export to JSONL (default)
pr-analyzer export --repo owner/repo

# Export to CSV
pr-analyzer export --repo owner/repo --format csv --output ./data

# Include file diffs in export
pr-analyzer export --repo owner/repo --include-diffs
```

## Configuration

The default configuration file is located at `~/.pr-analyzer/config.yaml`:

```yaml
github:
  token: ${GITHUB_TOKEN}  # Can use environment variable
  api_url: https://api.github.com  # For GitHub Enterprise
  
cache:
  location: ~/.pr-analyzer
  max_age_days: 90
  
export:
  default_format: jsonl
  include_raw_json: false
  
fetch:
  batch_size: 100
  rate_limit_buffer: 100  # Reserve API requests
```

## Commands

### `pr-analyzer init`
Initialize the configuration and cache directory.

### `pr-analyzer fetch`
Fetch PR data from GitHub and store in local cache.

Options:
- `--repo owner/repo` - Repository to fetch from (required)
- `--since YYYY-MM-DD` - Fetch PRs updated since this date
- `--pr NUMBER` - Fetch a specific PR
- `--full` - Perform full sync instead of incremental

### `pr-analyzer export`
Export cached PR data to files.

Options:
- `--repo owner/repo` - Repository to export (required)
- `--format jsonl|csv` - Export format (default: jsonl)
- `--output PATH` - Output directory (default: ./)
- `--include-diffs` - Include file diffs in export
- `--split-files` - Split output into multiple files

### `pr-analyzer cache`
Manage the local cache.

Subcommands:
- `cache stats [--repo owner/repo]` - Show cache statistics
- `cache clear [--repo owner/repo]` - Clear cache data

### `pr-analyzer config`
Manage configuration settings.

Subcommands:
- `config set KEY VALUE` - Set a configuration value
- `config get KEY` - Get a configuration value

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
-- Load data
CREATE TABLE pr_data AS 
SELECT * FROM read_json_auto('pr_data.jsonl');

-- Find most active reviewers
SELECT 
    json_extract_string(reviewer, '$.login') as reviewer,
    COUNT(*) as review_count
FROM (
    SELECT unnest(json_extract(data, '$.reviews')) as reviewer
    FROM pr_data
)
GROUP BY reviewer
ORDER BY review_count DESC;
```

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

## License

MIT
