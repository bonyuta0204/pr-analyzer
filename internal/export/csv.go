package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bonyuta0204/pr-analyzer/pkg/models"
)

type CSVExporter struct {
	filename     string
	includeDiffs bool
}

func NewCSVExporter(filename string, includeDiffs bool) *CSVExporter {
	return &CSVExporter{
		filename:     filename,
		includeDiffs: includeDiffs,
	}
}

func (e *CSVExporter) Export(prs []*models.PullRequest) error {
	file, err := os.Create(e.filename)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"number", "title", "state", "author", "author_type", "author_is_bot",
		"assignees", "requested_reviewers", "labels",
		"created_at", "updated_at", "merged_at",
		"additions", "deletions", "changed_files", "comments", "review_comments", "reviews",
		"review_states", "comment_authors", "file_paths",
	}

	if e.includeDiffs {
		header = append(header, "diff_summary")
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("writing CSV header: %w", err)
	}

	// Write data rows
	for _, pr := range prs {
		row := e.transformPRToRow(pr)
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("writing PR %d: %w", pr.Number, err)
		}
	}

	return nil
}

func (e *CSVExporter) transformPRToRow(pr *models.PullRequest) []string {
	// Basic PR fields
	row := []string{
		strconv.Itoa(pr.Number),
		e.escapeCsvValue(pr.Title),
		pr.State,
		pr.Author.Login,
		pr.Author.Type,
		strconv.FormatBool(pr.Author.IsBot),
		e.serializeUsers(pr.Assignees),
		e.serializeUsers(pr.RequestedReviewers),
		e.serializeLabels(pr.Labels),
		pr.CreatedAt.Format("2006-01-02T15:04:05Z"),
		pr.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		e.formatOptionalTime(pr.MergedAt),
		strconv.Itoa(pr.Stats.Additions),
		strconv.Itoa(pr.Stats.Deletions),
		strconv.Itoa(pr.Stats.ChangedFiles),
		strconv.Itoa(pr.Stats.Comments),
		strconv.Itoa(pr.Stats.ReviewComments),
		strconv.Itoa(pr.Stats.Reviews),
		e.getReviewStates(pr.Reviews),
		e.getCommentAuthors(pr.Comments),
		e.getFilePaths(pr.Files),
	}

	// Add diff summary if requested
	if e.includeDiffs {
		row = append(row, e.getDiffSummary(pr.Files))
	}

	return row
}

func (e *CSVExporter) escapeCsvValue(value string) string {
	// Remove newlines and tabs, truncate if too long
	cleaned := strings.ReplaceAll(value, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")
	cleaned = strings.ReplaceAll(cleaned, "\r", " ")

	if len(cleaned) > 200 {
		cleaned = cleaned[:200] + "..."
	}

	return cleaned
}

func (e *CSVExporter) serializeUsers(users []models.User) string {
	if len(users) == 0 {
		return ""
	}

	var usernames []string
	for _, user := range users {
		usernames = append(usernames, user.Login)
	}
	return strings.Join(usernames, ";")
}

func (e *CSVExporter) serializeLabels(labels []models.Label) string {
	if len(labels) == 0 {
		return ""
	}

	var labelNames []string
	for _, label := range labels {
		labelNames = append(labelNames, label.Name)
	}
	return strings.Join(labelNames, ";")
}

func (e *CSVExporter) formatOptionalTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02T15:04:05Z")
}

func (e *CSVExporter) getReviewStates(reviews []models.Review) string {
	if len(reviews) == 0 {
		return ""
	}

	stateCount := make(map[string]int)
	for _, review := range reviews {
		stateCount[review.State]++
	}

	var states []string
	for state, count := range stateCount {
		states = append(states, fmt.Sprintf("%s:%d", state, count))
	}
	return strings.Join(states, ";")
}

func (e *CSVExporter) getCommentAuthors(comments []models.Comment) string {
	if len(comments) == 0 {
		return ""
	}

	authorSet := make(map[string]bool)
	for _, comment := range comments {
		authorSet[comment.Author.Login] = true
	}

	var authors []string
	for author := range authorSet {
		authors = append(authors, author)
	}
	return strings.Join(authors, ";")
}

func (e *CSVExporter) getFilePaths(files []models.File) string {
	if len(files) == 0 {
		return ""
	}

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Filename)
	}
	return strings.Join(paths, ";")
}

func (e *CSVExporter) getDiffSummary(files []models.File) string {
	if len(files) == 0 {
		return ""
	}

	var summary []string
	for _, file := range files {
		if file.Patch != "" {
			lines := strings.Split(file.Patch, "\n")
			summary = append(summary, fmt.Sprintf("%s:%d_lines", file.Filename, len(lines)))
		}
	}
	return strings.Join(summary, ";")
}

func (e *CSVExporter) GetFileSize() (int64, error) {
	if stat, err := os.Stat(e.filename); err != nil {
		return 0, err
	} else {
		return stat.Size(), nil
	}
}
