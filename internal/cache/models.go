package cache

import (
	"time"

	"gorm.io/gorm"
)

type Pull struct {
	ID                 int64     `gorm:"primaryKey"`
	Number             int       `gorm:"uniqueIndex"`
	Title              string
	State              string
	Author             string
	AuthorType         string
	AuthorIsBot        bool
	Assignees          string    // JSON array
	RequestedReviewers string    // JSON array
	Labels             string    // JSON array
	CreatedAt          time.Time
	UpdatedAt          time.Time
	MergedAt           *time.Time
	LastFetchedAt      time.Time
	RawJSON            string
}

type Review struct {
	ID           int64 `gorm:"primaryKey"`
	PullNumber   int   `gorm:"index"`
	Reviewer     string
	ReviewerType string
	ReviewerIsBot bool
	State        string
	SubmittedAt  time.Time
	RawJSON      string
}

type Comment struct {
	ID          int64 `gorm:"primaryKey"`
	PullNumber  int   `gorm:"index"`
	ReviewID    *int64
	Author      string
	AuthorType  string
	AuthorIsBot bool
	Body        string `gorm:"type:text"`
	Path        string
	Line        *int
	Side        string
	DiffHunk    string `gorm:"type:text"`
	Reactions   string // JSON object
	CreatedAt   time.Time
	UpdatedAt   time.Time
	InReplyToID *int64
	RawJSON     string `gorm:"type:text"`
}

type File struct {
	PullNumber int    `gorm:"primaryKey;autoIncrement:false"`
	Filename   string `gorm:"primaryKey"`
	Status     string
	Additions  int
	Deletions  int
	Patch      string `gorm:"type:text"`
	RawJSON    string `gorm:"type:text"`
}

type SyncMetadata struct {
	Repo         string `gorm:"primaryKey"`
	LastSyncAt   time.Time
	LastPRNumber int
	TotalPRs     int
	OpenPRs      int
	ClosedPRs    int
}

type BotPattern struct {
	Pattern     string `gorm:"primaryKey"`
	Description string
}

func Migrate(db *gorm.DB) error {
	// Auto migrate all tables
	if err := db.AutoMigrate(
		&Pull{},
		&Review{},
		&Comment{},
		&File{},
		&SyncMetadata{},
		&BotPattern{},
	); err != nil {
		return err
	}

	// Add initial bot patterns
	botPatterns := []BotPattern{
		{Pattern: "dependabot", Description: "Dependency updates"},
		{Pattern: "renovate", Description: "Dependency updates"},
		{Pattern: "snyk", Description: "Security scanning"},
		{Pattern: "codecov", Description: "Code coverage"},
		{Pattern: "github-actions", Description: "GitHub Actions bot"},
		{Pattern: "vercel", Description: "Vercel deployment bot"},
		{Pattern: "netlify", Description: "Netlify deployment bot"},
	}

	for _, pattern := range botPatterns {
		db.FirstOrCreate(&pattern, BotPattern{Pattern: pattern.Pattern})
	}

	return nil
}