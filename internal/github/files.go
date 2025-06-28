package github

import (
	"context"
	"fmt"

	"github.com/bonyuta0204/pr-analyzer/pkg/models"
	"github.com/google/go-github/v50/github"
)

func (c *Client) FetchFiles(ctx context.Context, prNumber int) error {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		files, resp, err := c.client.PullRequests.ListFiles(ctx, c.owner, c.repo, prNumber, opts)
		if err != nil {
			return c.handleError(err, resp.Response)
		}

		for _, file := range files {
			f := c.convertFile(file, prNumber)
			if err := c.cache.SaveFile(f); err != nil {
				return fmt.Errorf("saving file %s: %w", file.GetFilename(), err)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

func (c *Client) convertFile(file *github.CommitFile, prNumber int) *models.File {
	return &models.File{
		PullNumber: prNumber,
		Filename:   file.GetFilename(),
		Status:     file.GetStatus(),
		Additions:  file.GetAdditions(),
		Deletions:  file.GetDeletions(),
		Patch:      file.GetPatch(),
	}
}