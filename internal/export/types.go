package export

import (
	"time"

	"github.com/bonyuta0204/pr-analyzer/pkg/models"
)

// ExportPullRequest represents a PR optimized for export
type ExportPullRequest struct {
	Type               string                  `json:"type"`
	Number             int                     `json:"number"`
	Title              string                  `json:"title"`
	State              string                  `json:"state"`
	Author             models.User             `json:"author"`
	Assignees          []models.User           `json:"assignees,omitempty"`
	RequestedReviewers []models.User           `json:"requested_reviewers,omitempty"`
	Labels             []models.Label          `json:"labels,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedAt          time.Time               `json:"updated_at"`
	MergedAt           *time.Time              `json:"merged_at,omitempty"`
	Stats              models.PullRequestStats `json:"stats"`
	Files              []models.File           `json:"files,omitempty"`
	Reviews            []models.Review         `json:"reviews,omitempty"`
	Comments           []ExportComment         `json:"comments,omitempty"`
}

// ExportComment represents a comment optimized for export with additional context
type ExportComment struct {
	Type        string              `json:"type"`
	PRNumber    int                 `json:"pr_number"`
	CommentID   int64               `json:"comment_id"`
	Author      models.User         `json:"author"`
	Body        string              `json:"body"`
	FilePath    string              `json:"file_path,omitempty"`
	Line        *int                `json:"line,omitempty"`
	Side        string              `json:"side,omitempty"`
	CodeContext *models.CodeContext `json:"code_context,omitempty"`
	Reactions   map[string]int      `json:"reactions,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   *time.Time          `json:"updated_at,omitempty"`
}

// Exporter interface for different export formats
type Exporter interface {
	Export(prs []*models.PullRequest) error
	GetFileSize() (int64, error)
}

// ExportOptions contains options for export
type ExportOptions struct {
	Format       string
	Filename     string
	IncludeDiffs bool
}

// NewExporter creates an exporter based on format
func NewExporter(opts ExportOptions) Exporter {
	switch opts.Format {
	case "csv":
		return NewCSVExporter(opts.Filename, opts.IncludeDiffs)
	case "jsonl":
		fallthrough
	default:
		return NewJSONLExporter(opts.Filename, opts.IncludeDiffs)
	}
}
