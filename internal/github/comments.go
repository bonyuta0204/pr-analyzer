package github

import (
	"context"
	"fmt"

	"github.com/bonyuta0204/pr-analyzer/pkg/models"
	"github.com/google/go-github/v50/github"
)

func (c *Client) FetchComments(ctx context.Context, prNumber int) error {
	// Fetch issue comments (general PR comments)
	if err := c.fetchIssueComments(ctx, prNumber); err != nil {
		return fmt.Errorf("fetching issue comments: %w", err)
	}

	// Fetch review comments (code comments)
	if err := c.fetchReviewComments(ctx, prNumber); err != nil {
		return fmt.Errorf("fetching review comments: %w", err)
	}

	return nil
}

func (c *Client) fetchIssueComments(ctx context.Context, prNumber int) error {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		comments, resp, err := c.client.Issues.ListComments(ctx, c.owner, c.repo, prNumber, opts)
		if err != nil {
			return c.handleError(err, resp.Response)
		}

		for _, comment := range comments {
			convertedComment := c.convertIssueComment(comment, prNumber)
			if err := c.cache.SaveComment(convertedComment); err != nil {
				return fmt.Errorf("saving comment %d: %w", comment.GetID(), err)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

func (c *Client) fetchReviewComments(ctx context.Context, prNumber int) error {
	opts := &github.PullRequestListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		comments, resp, err := c.client.PullRequests.ListComments(ctx, c.owner, c.repo, prNumber, opts)
		if err != nil {
			return c.handleError(err, resp.Response)
		}

		for _, comment := range comments {
			convertedComment := c.convertReviewComment(comment, prNumber)
			if err := c.cache.SaveComment(convertedComment); err != nil {
				return fmt.Errorf("saving review comment %d: %w", comment.GetID(), err)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

func (c *Client) convertIssueComment(comment *github.IssueComment, prNumber int) *models.Comment {
	result := &models.Comment{
		ID:         comment.GetID(),
		PullNumber: prNumber,
		Body:       comment.GetBody(),
		CreatedAt:  comment.GetCreatedAt().Time,
		UpdatedAt:  comment.GetUpdatedAt().Time,
		Reactions:  c.convertReactions(comment.Reactions),
	}

	if comment.User != nil {
		result.Author = models.User{
			Login: comment.User.GetLogin(),
			Type:  comment.User.GetType(),
		}
	}

	return result
}

func (c *Client) convertReviewComment(comment *github.PullRequestComment, prNumber int) *models.Comment {
	result := &models.Comment{
		ID:         comment.GetID(),
		PullNumber: prNumber,
		Body:       comment.GetBody(),
		Path:       comment.GetPath(),
		DiffHunk:   comment.GetDiffHunk(),
		CreatedAt:  comment.GetCreatedAt().Time,
		UpdatedAt:  comment.GetUpdatedAt().Time,
		Reactions:  c.convertReactions(comment.Reactions),
	}

	if comment.Line != nil {
		result.Line = comment.Line
	}

	if comment.Side != nil {
		result.Side = *comment.Side
	}

	if comment.InReplyTo != nil {
		result.InReplyToID = comment.InReplyTo
	}

	if comment.PullRequestReviewID != nil {
		result.ReviewID = comment.PullRequestReviewID
	}

	if comment.User != nil {
		result.Author = models.User{
			Login: comment.User.GetLogin(),
			Type:  comment.User.GetType(),
		}
	}

	return result
}

func (c *Client) convertReactions(reactions *github.Reactions) map[string]int {
	if reactions == nil {
		return nil
	}

	result := make(map[string]int)
	if reactions.GetPlusOne() > 0 {
		result["+1"] = reactions.GetPlusOne()
	}
	if reactions.GetMinusOne() > 0 {
		result["-1"] = reactions.GetMinusOne()
	}
	if reactions.GetLaugh() > 0 {
		result["laugh"] = reactions.GetLaugh()
	}
	if reactions.GetConfused() > 0 {
		result["confused"] = reactions.GetConfused()
	}
	if reactions.GetHeart() > 0 {
		result["heart"] = reactions.GetHeart()
	}
	if reactions.GetHooray() > 0 {
		result["hooray"] = reactions.GetHooray()
	}
	if reactions.GetRocket() > 0 {
		result["rocket"] = reactions.GetRocket()
	}
	if reactions.GetEyes() > 0 {
		result["eyes"] = reactions.GetEyes()
	}

	if len(result) == 0 {
		return nil
	}
	return result
}
