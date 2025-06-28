package github

import (
	"context"
	"fmt"

	"github.com/bonyuta0204/pr-analyzer/pkg/models"
	"github.com/google/go-github/v50/github"
)

func (c *Client) FetchReviews(ctx context.Context, prNumber int) error {
	opts := &github.ListOptions{
		PerPage: 100,
	}

	for {
		reviews, resp, err := c.client.PullRequests.ListReviews(ctx, c.owner, c.repo, prNumber, opts)
		if err != nil {
			return c.handleError(err, resp.Response)
		}

		for _, review := range reviews {
			r := c.convertReview(review, prNumber)
			if err := c.cache.SaveReview(r); err != nil {
				return fmt.Errorf("saving review %d: %w", review.GetID(), err)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return nil
}

func (c *Client) convertReview(review *github.PullRequestReview, prNumber int) *models.Review {
	result := &models.Review{
		ID:          review.GetID(),
		PullNumber:  prNumber,
		State:       review.GetState(),
		Body:        review.GetBody(),
		SubmittedAt: review.GetSubmittedAt().Time,
	}

	if review.User != nil {
		result.Reviewer = models.User{
			Login: review.User.GetLogin(),
			Type:  review.User.GetType(),
		}
	}

	return result
}
