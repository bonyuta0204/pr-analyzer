package models

import (
	"encoding/json"
	"time"
)

type User struct {
	Login string `json:"login"`
	Type  string `json:"type"`
	IsBot bool   `json:"is_bot"`
}

type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type PullRequest struct {
	ID                 int64            `json:"id"`
	Number             int              `json:"number"`
	Title              string           `json:"title"`
	State              string           `json:"state"`
	Author             User             `json:"author"`
	Assignees          []User           `json:"assignees"`
	RequestedReviewers []User           `json:"requested_reviewers"`
	Labels             []Label          `json:"labels"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	MergedAt           *time.Time       `json:"merged_at,omitempty"`
	Stats              PullRequestStats `json:"stats"`
	Files              []File           `json:"files,omitempty"`
	Reviews            []Review         `json:"reviews,omitempty"`
	Comments           []Comment        `json:"comments,omitempty"`
	RawJSON            json.RawMessage  `json:"-"`
}

type PullRequestStats struct {
	Additions      int `json:"additions"`
	Deletions      int `json:"deletions"`
	ChangedFiles   int `json:"changed_files"`
	Comments       int `json:"comments"`
	ReviewComments int `json:"review_comments"`
	Reviews        int `json:"reviews"`
}

type File struct {
	PullNumber int             `json:"-"`
	Filename   string          `json:"filename"`
	Status     string          `json:"status"`
	Additions  int             `json:"additions"`
	Deletions  int             `json:"deletions"`
	Patch      string          `json:"patch,omitempty"`
	RawJSON    json.RawMessage `json:"-"`
}

type Review struct {
	ID          int64           `json:"id"`
	PullNumber  int             `json:"-"`
	Reviewer    User            `json:"reviewer"`
	State       string          `json:"state"`
	SubmittedAt time.Time       `json:"submitted_at"`
	Body        string          `json:"body,omitempty"`
	RawJSON     json.RawMessage `json:"-"`
}

type Comment struct {
	ID          int64           `json:"id"`
	PullNumber  int             `json:"pr_number"`
	ReviewID    *int64          `json:"review_id,omitempty"`
	Author      User            `json:"author"`
	Body        string          `json:"body"`
	Path        string          `json:"file_path,omitempty"`
	Line        *int            `json:"line,omitempty"`
	Side        string          `json:"side,omitempty"`
	DiffHunk    string          `json:"diff_hunk,omitempty"`
	InReplyToID *int64          `json:"in_reply_to_id,omitempty"`
	Reactions   map[string]int  `json:"reactions,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	RawJSON     json.RawMessage `json:"-"`
}

type CodeContext struct {
	DiffHunk    string   `json:"diff_hunk"`
	BeforeLines []string `json:"before_lines,omitempty"`
	TargetLine  string   `json:"target_line,omitempty"`
	AfterLines  []string `json:"after_lines,omitempty"`
}

type CommentThread struct {
	RootID  int64     `json:"root_id"`
	Replies []Comment `json:"replies,omitempty"`
}

type ExportComment struct {
	Type        string         `json:"type"`
	PRNumber    int            `json:"pr_number"`
	CommentID   int64          `json:"comment_id"`
	Author      User           `json:"author"`
	Body        string         `json:"body"`
	FilePath    string         `json:"file_path,omitempty"`
	Line        *int           `json:"line,omitempty"`
	Side        string         `json:"side,omitempty"`
	CodeContext *CodeContext   `json:"code_context,omitempty"`
	Thread      *CommentThread `json:"thread,omitempty"`
	Reactions   map[string]int `json:"reactions,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   *time.Time     `json:"updated_at,omitempty"`
}

type SyncMetadata struct {
	Repo         string    `json:"repo"`
	LastSyncAt   time.Time `json:"last_sync_at"`
	LastPRNumber int       `json:"last_pr_number"`
	TotalPRs     int       `json:"total_prs"`
	OpenPRs      int       `json:"open_prs"`
	ClosedPRs    int       `json:"closed_prs"`
}
