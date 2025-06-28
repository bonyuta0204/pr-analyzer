package github

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bonyuta0204/pr-analyzer/internal/cache"
	"github.com/bonyuta0204/pr-analyzer/internal/config"
	"github.com/bonyuta0204/pr-analyzer/pkg/models"
	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

type Client struct {
	client *github.Client
	cache  *cache.Store
	config *config.Config
	owner  string
	repo   string
}

func NewClient(cfg *config.Config, store *cache.Store, repoPath string) (*Client, error) {
	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository format: %s (expected owner/repo)", repoPath)
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.GitHub.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	githubClient := github.NewClient(tc)
	if cfg.GitHub.APIURL != "" && cfg.GitHub.APIURL != "https://api.github.com" {
		baseURL := cfg.GitHub.APIURL
		if !strings.HasSuffix(baseURL, "/") {
			baseURL += "/"
		}
		var err error
		githubClient.BaseURL, err = githubClient.BaseURL.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("parsing GitHub API URL: %w", err)
		}
	}

	return &Client{
		client: githubClient,
		cache:  store,
		config: cfg,
		owner:  parts[0],
		repo:   parts[1],
	}, nil
}

func (c *Client) GetRepository() string {
	return c.owner + "/" + c.repo
}

func (c *Client) FetchPullRequests(ctx context.Context, since time.Time, prNumber int) error {
	if prNumber > 0 {
		return c.fetchSinglePR(ctx, prNumber)
	}

	opts := &github.PullRequestListOptions{
		State:     "all",
		Sort:      "updated",
		Direction: "desc",
		ListOptions: github.ListOptions{
			PerPage: c.config.Fetch.BatchSize,
		},
	}

	for {
		prs, resp, err := c.client.PullRequests.List(ctx, c.owner, c.repo, opts)
		if err != nil {
			return c.handleError(err, resp.Response)
		}

		for i, pr := range prs {
			// Skip if PR hasn't been updated since the given time
			if !since.IsZero() && pr.UpdatedAt.Before(since) {
				// Since results are sorted by updated desc, we can stop here
				return nil
			}

			pullRequest := c.convertPullRequest(pr)
			if err := c.cache.SavePullRequest(pullRequest); err != nil {
				return fmt.Errorf("saving PR %d: %w", *pr.Number, err)
			}

			// Fetch additional data for each PR
			if err := c.fetchPRDetails(ctx, *pr.Number); err != nil {
				return fmt.Errorf("fetching details for PR %d: %w", *pr.Number, err)
			}

			// Simple progress feedback - print to show we're making progress
			if (i+1)%5 == 0 || i == len(prs)-1 {
				fmt.Printf("\r│  ├─ Recent PRs................ ⠋ %d fetched", i+1)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	// Update sync metadata
	meta := &models.SyncMetadata{
		Repo:       c.GetRepository(),
		LastSyncAt: time.Now(),
	}
	return c.cache.SaveSyncMetadata(meta)
}

func (c *Client) fetchSinglePR(ctx context.Context, number int) error {
	pr, resp, err := c.client.PullRequests.Get(ctx, c.owner, c.repo, number)
	if err != nil {
		return c.handleError(err, resp.Response)
	}

	pullRequest := c.convertPullRequest(pr)
	if err := c.cache.SavePullRequest(pullRequest); err != nil {
		return fmt.Errorf("saving PR %d: %w", number, err)
	}

	return c.fetchPRDetails(ctx, number)
}

func (c *Client) fetchPRDetails(ctx context.Context, number int) error {
	// Fetch reviews
	if err := c.FetchReviews(ctx, number); err != nil {
		return fmt.Errorf("fetching reviews: %w", err)
	}

	// Fetch comments
	if err := c.FetchComments(ctx, number); err != nil {
		return fmt.Errorf("fetching comments: %w", err)
	}

	// Fetch files
	if err := c.FetchFiles(ctx, number); err != nil {
		return fmt.Errorf("fetching files: %w", err)
	}

	return nil
}

func (c *Client) convertPullRequest(pr *github.PullRequest) *models.PullRequest {
	result := &models.PullRequest{
		ID:        pr.GetID(),
		Number:    pr.GetNumber(),
		Title:     pr.GetTitle(),
		State:     pr.GetState(),
		CreatedAt: pr.GetCreatedAt().Time,
		UpdatedAt: pr.GetUpdatedAt().Time,
		Stats: models.PullRequestStats{
			Additions:      pr.GetAdditions(),
			Deletions:      pr.GetDeletions(),
			ChangedFiles:   pr.GetChangedFiles(),
			Comments:       pr.GetComments(),
			ReviewComments: pr.GetReviewComments(),
		},
	}

	if pr.MergedAt != nil {
		result.MergedAt = &pr.MergedAt.Time
	}

	// Author
	if pr.User != nil {
		result.Author = models.User{
			Login: pr.User.GetLogin(),
			Type:  pr.User.GetType(),
		}
	}

	// Assignees
	for _, assignee := range pr.Assignees {
		result.Assignees = append(result.Assignees, models.User{
			Login: assignee.GetLogin(),
			Type:  assignee.GetType(),
		})
	}

	// Requested reviewers
	for _, reviewer := range pr.RequestedReviewers {
		result.RequestedReviewers = append(result.RequestedReviewers, models.User{
			Login: reviewer.GetLogin(),
			Type:  reviewer.GetType(),
		})
	}

	// Labels
	for _, label := range pr.Labels {
		result.Labels = append(result.Labels, models.Label{
			Name:  label.GetName(),
			Color: label.GetColor(),
		})
	}

	return result
}

func (c *Client) handleError(err error, resp *http.Response) error {
	if resp != nil {
		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("GitHub authentication failed: please check your token")
		}
		if resp.StatusCode == http.StatusForbidden {
			return fmt.Errorf("GitHub API rate limit exceeded or forbidden access")
		}
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("repository not found: %s/%s", c.owner, c.repo)
		}
	}
	return fmt.Errorf("GitHub API error: %w", err)
}
