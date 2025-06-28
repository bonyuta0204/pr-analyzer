package ui

import (
	"testing"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		repo     string
		format   string
		options  FilenameOptions
		expected string
	}{
		{
			name:     "basic filename",
			repo:     "owner/repo",
			format:   "jsonl",
			options:  FilenameOptions{},
			expected: "owner-repo-prs.jsonl",
		},
		{
			name:     "with limit",
			repo:     "owner/repo",
			format:   "csv",
			options:  FilenameOptions{Limit: 50},
			expected: "owner-repo-prs-recent-50.csv",
		},
		{
			name:     "specific PR",
			repo:     "owner/repo",
			format:   "jsonl",
			options:  FilenameOptions{PRNumber: 123},
			expected: "owner-repo-pr-123.jsonl",
		},
		{
			name:     "with since date",
			repo:     "owner/repo",
			format:   "jsonl",
			options:  FilenameOptions{Since: "2024-01-15"},
			expected: "owner-repo-prs-2024-01.jsonl",
		},
		{
			name:     "all PRs",
			repo:     "owner/repo",
			format:   "jsonl",
			options:  FilenameOptions{All: true},
			expected: "owner-repo-prs.jsonl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateFilename(tt.repo, tt.format, tt.options)
			if got != tt.expected {
				t.Errorf("GenerateFilename() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    500,
			expected: "500 B",
		},
		{
			name:     "kilobytes",
			bytes:    1536,
			expected: "1.5 KB",
		},
		{
			name:     "megabytes",
			bytes:    2621440,
			expected: "2.5 MB",
		},
		{
			name:     "gigabytes",
			bytes:    5368709120,
			expected: "5.0 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatFileSize(tt.bytes)
			if got != tt.expected {
				t.Errorf("FormatFileSize() = %v, want %v", got, tt.expected)
			}
		})
	}
}
