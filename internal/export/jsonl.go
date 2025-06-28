package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bonyuta0204/pr-analyzer/pkg/models"
)

type JSONLExporter struct {
	filename     string
	includeDiffs bool
}

func NewJSONLExporter(filename string, includeDiffs bool) *JSONLExporter {
	return &JSONLExporter{
		filename:     filename,
		includeDiffs: includeDiffs,
	}
}

func (e *JSONLExporter) Export(prs []*models.PullRequest) error {
	file, err := os.Create(e.filename)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer file.Close()

	for _, pr := range prs {
		// Create export record
		exportPR := e.transformPR(pr)
		
		// Marshal to JSON
		jsonData, err := json.Marshal(exportPR)
		if err != nil {
			return fmt.Errorf("marshaling PR %d: %w", pr.Number, err)
		}
		
		// Write JSONL line
		if _, err := file.Write(jsonData); err != nil {
			return fmt.Errorf("writing PR %d: %w", pr.Number, err)
		}
		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("writing newline for PR %d: %w", pr.Number, err)
		}
	}

	return nil
}

func (e *JSONLExporter) transformPR(pr *models.PullRequest) ExportPullRequest {
	exportPR := ExportPullRequest{
		Type:               "pull",
		Number:             pr.Number,
		Title:              pr.Title,
		State:              pr.State,
		Author:             pr.Author,
		Assignees:          pr.Assignees,
		RequestedReviewers: pr.RequestedReviewers,
		Labels:             pr.Labels,
		CreatedAt:          pr.CreatedAt,
		UpdatedAt:          pr.UpdatedAt,
		MergedAt:           pr.MergedAt,
		Stats:              pr.Stats,
		Reviews:            pr.Reviews,
		Comments:           e.transformComments(pr.Comments),
	}

	// Include files based on includeDiffs setting
	if e.includeDiffs {
		exportPR.Files = pr.Files
	} else {
		// Include files without patch content
		exportPR.Files = make([]models.File, len(pr.Files))
		for i, file := range pr.Files {
			exportPR.Files[i] = models.File{
				Filename:  file.Filename,
				Status:    file.Status,
				Additions: file.Additions,
				Deletions: file.Deletions,
				// Omit Patch field
			}
		}
	}

	return exportPR
}

func (e *JSONLExporter) transformComments(comments []models.Comment) []ExportComment {
	var exportComments []ExportComment
	
	for _, comment := range comments {
		exportComment := ExportComment{
			Type:      e.getCommentType(comment),
			PRNumber:  comment.PullNumber,
			CommentID: comment.ID,
			Author:    comment.Author,
			Body:      comment.Body,
			CreatedAt: comment.CreatedAt,
			Reactions: comment.Reactions,
		}

		// Add code-specific fields if it's a review comment
		if comment.Path != "" {
			exportComment.FilePath = comment.Path
			exportComment.Line = comment.Line
			exportComment.Side = comment.Side
			
			if e.includeDiffs && comment.DiffHunk != "" {
				exportComment.CodeContext = &models.CodeContext{
					DiffHunk: comment.DiffHunk,
				}
			}
		}

		// Add updated time if different from created
		if !comment.UpdatedAt.Equal(comment.CreatedAt) {
			exportComment.UpdatedAt = &comment.UpdatedAt
		}

		exportComments = append(exportComments, exportComment)
	}

	return exportComments
}

func (e *JSONLExporter) getCommentType(comment models.Comment) string {
	if comment.Path != "" {
		return "code_comment"
	}
	return "issue_comment"
}

func (e *JSONLExporter) GetFileSize() (int64, error) {
	if stat, err := os.Stat(e.filename); err != nil {
		return 0, err
	} else {
		return stat.Size(), nil
	}
}