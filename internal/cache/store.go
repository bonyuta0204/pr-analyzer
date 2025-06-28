package cache

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bonyuta0204/pr-analyzer/pkg/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	db         *gorm.DB
	botPatterns []string
}

func NewStore(dbPath string) (*Store, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if err := Migrate(db); err != nil {
		return nil, fmt.Errorf("migrating database: %w", err)
	}

	store := &Store{db: db}
	if err := store.loadBotPatterns(); err != nil {
		return nil, fmt.Errorf("loading bot patterns: %w", err)
	}

	return store, nil
}

func (s *Store) loadBotPatterns() error {
	var patterns []BotPattern
	if err := s.db.Find(&patterns).Error; err != nil {
		return err
	}

	s.botPatterns = make([]string, len(patterns))
	for i, p := range patterns {
		s.botPatterns[i] = p.Pattern
	}

	return nil
}

func (s *Store) isBot(username string) bool {
	lowerUsername := strings.ToLower(username)
	for _, pattern := range s.botPatterns {
		if strings.Contains(lowerUsername, pattern) {
			return true
		}
	}
	return strings.HasSuffix(lowerUsername, "[bot]")
}

func (s *Store) SavePullRequest(pr *models.PullRequest) error {
	rawJSON, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("marshaling raw JSON: %w", err)
	}

	assignees, _ := json.Marshal(pr.Assignees)
	reviewers, _ := json.Marshal(pr.RequestedReviewers)
	labels, _ := json.Marshal(pr.Labels)

	cachePR := &Pull{
		ID:                 pr.ID,
		Number:             pr.Number,
		Title:              pr.Title,
		State:              pr.State,
		Author:             pr.Author.Login,
		AuthorType:         pr.Author.Type,
		AuthorIsBot:        s.isBot(pr.Author.Login),
		Assignees:          string(assignees),
		RequestedReviewers: string(reviewers),
		Labels:             string(labels),
		CreatedAt:          pr.CreatedAt,
		UpdatedAt:          pr.UpdatedAt,
		MergedAt:           pr.MergedAt,
		LastFetchedAt:      time.Now(),
		RawJSON:            string(rawJSON),
	}

	return s.db.Save(cachePR).Error
}

func (s *Store) SaveReview(review *models.Review) error {
	rawJSON, err := json.Marshal(review)
	if err != nil {
		return fmt.Errorf("marshaling raw JSON: %w", err)
	}

	cacheReview := &Review{
		ID:            review.ID,
		PullNumber:    review.PullNumber,
		Reviewer:      review.Reviewer.Login,
		ReviewerType:  review.Reviewer.Type,
		ReviewerIsBot: s.isBot(review.Reviewer.Login),
		State:         review.State,
		SubmittedAt:   review.SubmittedAt,
		RawJSON:       string(rawJSON),
	}

	return s.db.Save(cacheReview).Error
}

func (s *Store) SaveComment(comment *models.Comment) error {
	rawJSON, err := json.Marshal(comment)
	if err != nil {
		return fmt.Errorf("marshaling raw JSON: %w", err)
	}

	reactions, _ := json.Marshal(comment.Reactions)

	cacheComment := &Comment{
		ID:          comment.ID,
		PullNumber:  comment.PullNumber,
		ReviewID:    comment.ReviewID,
		Author:      comment.Author.Login,
		AuthorType:  comment.Author.Type,
		AuthorIsBot: s.isBot(comment.Author.Login),
		Body:        comment.Body,
		Path:        comment.Path,
		Line:        comment.Line,
		Side:        comment.Side,
		DiffHunk:    comment.DiffHunk,
		Reactions:   string(reactions),
		CreatedAt:   comment.CreatedAt,
		UpdatedAt:   comment.UpdatedAt,
		InReplyToID: comment.InReplyToID,
		RawJSON:     string(rawJSON),
	}

	return s.db.Save(cacheComment).Error
}

func (s *Store) SaveFile(file *models.File) error {
	rawJSON, err := json.Marshal(file)
	if err != nil {
		return fmt.Errorf("marshaling raw JSON: %w", err)
	}

	cacheFile := &File{
		PullNumber: file.PullNumber,
		Filename:   file.Filename,
		Status:     file.Status,
		Additions:  file.Additions,
		Deletions:  file.Deletions,
		Patch:      file.Patch,
		RawJSON:    string(rawJSON),
	}

	return s.db.Save(cacheFile).Error
}

func (s *Store) GetPullRequest(number int) (*models.PullRequest, error) {
	var pull Pull
	if err := s.db.First(&pull, "number = ?", number).Error; err != nil {
		return nil, err
	}

	var pr models.PullRequest
	if err := json.Unmarshal([]byte(pull.RawJSON), &pr); err != nil {
		return nil, fmt.Errorf("unmarshaling PR: %w", err)
	}

	return &pr, nil
}

func (s *Store) GetPullRequests(repo string, since time.Time) ([]*models.PullRequest, error) {
	var pulls []Pull
	query := s.db.Order("updated_at DESC")
	
	if !since.IsZero() {
		query = query.Where("updated_at >= ?", since)
	}

	if err := query.Find(&pulls).Error; err != nil {
		return nil, err
	}

	prs := make([]*models.PullRequest, len(pulls))
	for i, pull := range pulls {
		var pr models.PullRequest
		if err := json.Unmarshal([]byte(pull.RawJSON), &pr); err != nil {
			return nil, fmt.Errorf("unmarshaling PR %d: %w", pull.Number, err)
		}
		prs[i] = &pr
	}

	return prs, nil
}

func (s *Store) GetReviews(prNumber int) ([]*models.Review, error) {
	var reviews []Review
	if err := s.db.Where("pull_number = ?", prNumber).Find(&reviews).Error; err != nil {
		return nil, err
	}

	result := make([]*models.Review, len(reviews))
	for i, review := range reviews {
		var r models.Review
		if err := json.Unmarshal([]byte(review.RawJSON), &r); err != nil {
			return nil, fmt.Errorf("unmarshaling review %d: %w", review.ID, err)
		}
		result[i] = &r
	}

	return result, nil
}

func (s *Store) GetComments(prNumber int) ([]*models.Comment, error) {
	var comments []Comment
	if err := s.db.Where("pull_number = ?", prNumber).Order("created_at").Find(&comments).Error; err != nil {
		return nil, err
	}

	result := make([]*models.Comment, len(comments))
	for i, comment := range comments {
		var c models.Comment
		if err := json.Unmarshal([]byte(comment.RawJSON), &c); err != nil {
			return nil, fmt.Errorf("unmarshaling comment %d: %w", comment.ID, err)
		}
		result[i] = &c
	}

	return result, nil
}

func (s *Store) GetFiles(prNumber int) ([]*models.File, error) {
	var files []File
	if err := s.db.Where("pull_number = ?", prNumber).Find(&files).Error; err != nil {
		return nil, err
	}

	result := make([]*models.File, len(files))
	for i, file := range files {
		var f models.File
		if err := json.Unmarshal([]byte(file.RawJSON), &f); err != nil {
			return nil, fmt.Errorf("unmarshaling file %s: %w", file.Filename, err)
		}
		result[i] = &f
	}

	return result, nil
}

func (s *Store) GetSyncMetadata(repo string) (*models.SyncMetadata, error) {
	var meta SyncMetadata
	if err := s.db.First(&meta, "repo = ?", repo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &models.SyncMetadata{
		Repo:         meta.Repo,
		LastSyncAt:   meta.LastSyncAt,
		LastPRNumber: meta.LastPRNumber,
		TotalPRs:     meta.TotalPRs,
		OpenPRs:      meta.OpenPRs,
		ClosedPRs:    meta.ClosedPRs,
	}, nil
}

func (s *Store) SaveSyncMetadata(meta *models.SyncMetadata) error {
	cacheMeta := &SyncMetadata{
		Repo:         meta.Repo,
		LastSyncAt:   meta.LastSyncAt,
		LastPRNumber: meta.LastPRNumber,
		TotalPRs:     meta.TotalPRs,
		OpenPRs:      meta.OpenPRs,
		ClosedPRs:    meta.ClosedPRs,
	}

	return s.db.Save(cacheMeta).Error
}

func (s *Store) GetStats(repo string) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count PRs
	var prCount int64
	s.db.Model(&Pull{}).Count(&prCount)
	stats["total_prs"] = prCount

	// Count by state
	var openCount, closedCount int64
	s.db.Model(&Pull{}).Where("state = ?", "open").Count(&openCount)
	s.db.Model(&Pull{}).Where("state = ?", "closed").Count(&closedCount)
	stats["open_prs"] = openCount
	stats["closed_prs"] = closedCount

	// Count reviews
	var reviewCount int64
	s.db.Model(&Review{}).Count(&reviewCount)
	stats["total_reviews"] = reviewCount

	// Count comments
	var commentCount int64
	s.db.Model(&Comment{}).Count(&commentCount)
	stats["total_comments"] = commentCount

	// Count files
	var fileCount int64
	s.db.Model(&File{}).Count(&fileCount)
	stats["total_files"] = fileCount

	// Get sync metadata
	meta, _ := s.GetSyncMetadata(repo)
	if meta != nil {
		stats["last_sync"] = meta.LastSyncAt
	}

	return stats, nil
}

func (s *Store) Clear(repo string) error {
	// If repo is specified, only clear data for that repo
	// For now, we'll clear all data since we don't track repo in all tables
	return s.db.Transaction(func(tx *gorm.DB) error {
		tables := []string{"pulls", "reviews", "comments", "files"}
		for _, table := range tables {
			if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
				return err
			}
		}

		if repo != "" {
			return tx.Where("repo = ?", repo).Delete(&SyncMetadata{}).Error
		}
		return tx.Exec("DELETE FROM sync_metadata").Error
	})
}

func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}