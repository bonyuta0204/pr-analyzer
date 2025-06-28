package analyzer

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bonyuta0204/pr-analyzer/internal/cache"
	"github.com/bonyuta0204/pr-analyzer/internal/config"
	"github.com/bonyuta0204/pr-analyzer/internal/export"
	"github.com/bonyuta0204/pr-analyzer/internal/github"
	"github.com/bonyuta0204/pr-analyzer/internal/ui"
	"github.com/bonyuta0204/pr-analyzer/pkg/models"
)

type Service struct {
	config   *config.Config
	cache    *cache.Store
	github   *github.Client
	progress *ui.ProgressDisplay
}

type AnalyzeOptions struct {
	Repo         string
	Format       string
	Limit        int
	All          bool
	IncludeDiffs bool
	Refetch      bool
	Since        string
	PRNumber     int
	Output       string
}

func NewService() (*Service, error) {
	// Load configuration
	cfg := config.DefaultConfig()
	
	// Validate GitHub token
	if cfg.GitHub.Token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	
	// Create cache directory if it doesn't exist
	if err := ensureCacheDir(cfg.Cache.Location); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}
	
	// Initialize cache
	cacheStore, err := cache.NewStore(cfg.CacheDB())
	if err != nil {
		return nil, fmt.Errorf("initializing cache: %w", err)
	}
	
	return &Service{
		config:   cfg,
		cache:    cacheStore,
		progress: ui.NewProgressDisplay(),
	}, nil
}

func (s *Service) Analyze(ctx context.Context, opts AnalyzeOptions) error {
	// Start analysis display
	s.progress.StartSection("🔍", fmt.Sprintf("Analyzing %s", opts.Repo))
	
	// Initialize GitHub client for this repo
	githubClient, err := github.NewClient(s.config, s.cache, opts.Repo)
	if err != nil {
		s.progress.ShowError(fmt.Errorf("initializing GitHub client: %w", err))
		return err
	}
	s.github = githubClient
	
	// Check cache status
	if err := s.showCacheStatus(opts.Repo, opts.Refetch); err != nil {
		s.progress.ShowError(err)
		return err
	}
	
	// Fetch data from GitHub
	if err := s.fetchData(ctx, opts); err != nil {
		s.progress.ShowError(err)
		return err
	}
	
	// Load data from cache
	prs, err := s.loadDataFromCache(opts)
	if err != nil {
		s.progress.ShowError(err)
		return err
	}
	
	// Export data
	filename, fileSize, err := s.exportData(prs, opts)
	if err != nil {
		s.progress.ShowError(err)
		return err
	}
	
	// Show success
	s.progress.ShowSuccess(len(prs), filename, ui.FormatFileSize(fileSize))
	
	return nil
}

func (s *Service) showCacheStatus(repo string, refetch bool) error {
	if refetch {
		s.progress.ShowCacheStatus(0, "forced refresh")
		return nil
	}
	
	// Get sync metadata
	meta, err := s.cache.GetSyncMetadata(repo)
	if err != nil {
		s.progress.ShowCacheStatus(0, "no cache")
		return nil
	}
	
	if meta == nil {
		s.progress.ShowCacheStatus(0, "no cache")
		return nil
	}
	
	// Calculate relative time
	elapsed := time.Since(meta.LastSyncAt)
	var timeStr string
	if elapsed < time.Hour {
		timeStr = fmt.Sprintf("%d minutes ago", int(elapsed.Minutes()))
	} else if elapsed < 24*time.Hour {
		timeStr = fmt.Sprintf("%d hours ago", int(elapsed.Hours()))
	} else {
		timeStr = fmt.Sprintf("%d days ago", int(elapsed.Hours()/24))
	}
	
	s.progress.ShowCacheStatus(meta.TotalPRs, timeStr)
	return nil
}

func (s *Service) fetchData(ctx context.Context, opts AnalyzeOptions) error {
	s.progress.StartFetching()
	
	// Parse since date if provided
	var sinceTime time.Time
	if opts.Since != "" {
		var err error
		sinceTime, err = time.Parse("2006-01-02", opts.Since)
		if err != nil {
			return fmt.Errorf("invalid date format '%s': use YYYY-MM-DD", opts.Since)
		}
	}
	
	// Show progress for fetching PRs
	s.progress.ShowProgress("Recent PRs", 0, "fetching")
	
	// Fetch from GitHub
	err := s.github.FetchPullRequests(ctx, sinceTime, opts.PRNumber)
	if err != nil {
		return fmt.Errorf("fetching pull requests: %w", err)
	}
	
	// For demo purposes, show some progress updates
	// In real implementation, this would be driven by the GitHub client
	s.progress.StopProgress()
	
	// Simulate additional fetching steps
	s.progress.ShowProgress("Reviews", 0, "fetching")
	time.Sleep(100 * time.Millisecond) // Simulated work
	s.progress.StopProgress()
	
	s.progress.ShowProgress("Comments", 0, "fetching")
	time.Sleep(100 * time.Millisecond) // Simulated work
	s.progress.StopProgress()
	
	s.progress.ShowProgress("Files", 0, "fetching")
	time.Sleep(100 * time.Millisecond) // Simulated work
	s.progress.StopProgress()
	
	return nil
}

func (s *Service) loadDataFromCache(opts AnalyzeOptions) ([]*models.PullRequest, error) {
	// For now, load all PRs from cache
	// TODO: Apply limit and filtering
	prs, err := s.cache.GetPullRequests(opts.Repo, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("loading PRs from cache: %w", err)
	}
	
	// Apply limit if specified and not fetching all
	if !opts.All && opts.Limit > 0 && len(prs) > opts.Limit {
		prs = prs[:opts.Limit]
	}
	
	// Load associated data for each PR
	for _, pr := range prs {
		// Load reviews
		reviews, err := s.cache.GetReviews(pr.Number)
		if err != nil {
			return nil, fmt.Errorf("loading reviews for PR %d: %w", pr.Number, err)
		}
		// Convert slice of pointers to slice of values
		pr.Reviews = make([]models.Review, len(reviews))
		for i, review := range reviews {
			pr.Reviews[i] = *review
		}
		
		// Load comments
		comments, err := s.cache.GetComments(pr.Number)
		if err != nil {
			return nil, fmt.Errorf("loading comments for PR %d: %w", pr.Number, err)
		}
		// Convert slice of pointers to slice of values
		pr.Comments = make([]models.Comment, len(comments))
		for i, comment := range comments {
			pr.Comments[i] = *comment
		}
		
		// Load files
		files, err := s.cache.GetFiles(pr.Number)
		if err != nil {
			return nil, fmt.Errorf("loading files for PR %d: %w", pr.Number, err)
		}
		// Convert slice of pointers to slice of values
		pr.Files = make([]models.File, len(files))
		for i, file := range files {
			pr.Files[i] = *file
		}
	}
	
	return prs, nil
}

func (s *Service) exportData(prs []*models.PullRequest, opts AnalyzeOptions) (string, int64, error) {
	// Generate filename if not provided
	filename := opts.Output
	if filename == "" {
		filenameOpts := ui.FilenameOptions{
			PRNumber: opts.PRNumber,
			Limit:    opts.Limit,
			All:      opts.All,
			Since:    opts.Since,
		}
		filename = ui.GenerateFilename(opts.Repo, opts.Format, filenameOpts)
	}
	
	// Create exporter
	exportOpts := export.ExportOptions{
		Format:       opts.Format,
		Filename:     filename,
		IncludeDiffs: opts.IncludeDiffs,
	}
	exporter := export.NewExporter(exportOpts)
	
	// Export data
	if err := exporter.Export(prs); err != nil {
		return "", 0, fmt.Errorf("exporting data: %w", err)
	}
	
	// Get file size
	fileSize, err := exporter.GetFileSize()
	if err != nil {
		return filename, 0, fmt.Errorf("getting file size: %w", err)
	}
	
	return filename, fileSize, nil
}

func (s *Service) Close() error {
	if s.cache != nil {
		return s.cache.Close()
	}
	return nil
}

func ensureCacheDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}