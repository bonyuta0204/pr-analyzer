package ui

import (
	"fmt"
	"strings"
	"time"
)

func GenerateFilename(repo, format string, options FilenameOptions) string {
	// Clean repo name for filename
	cleanRepo := strings.ReplaceAll(repo, "/", "-")
	
	base := cleanRepo + "-prs"
	
	// Add modifiers based on options
	var parts []string
	
	if options.PRNumber > 0 {
		return fmt.Sprintf("%s-pr-%d.%s", cleanRepo, options.PRNumber, format)
	}
	
	if options.Limit > 0 && !options.All {
		parts = append(parts, fmt.Sprintf("recent-%d", options.Limit))
	}
	
	if options.Since != "" {
		// Convert date to short format
		if date, err := time.Parse("2006-01-02", options.Since); err == nil {
			parts = append(parts, date.Format("2006-01"))
		} else {
			parts = append(parts, "filtered")
		}
	}
	
	if len(parts) > 0 {
		base = base + "-" + strings.Join(parts, "-")
	}
	
	return fmt.Sprintf("%s.%s", base, format)
}

type FilenameOptions struct {
	PRNumber int
	Limit    int
	All      bool
	Since    string
}

func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	if bytes < KB {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < MB {
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	} else if bytes < GB {
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	} else {
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	}
}