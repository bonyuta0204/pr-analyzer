package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	GitHub GitHubConfig `yaml:"github"`
	Cache  CacheConfig  `yaml:"cache"`
	Export ExportConfig `yaml:"export"`
	Fetch  FetchConfig  `yaml:"fetch"`
}

type GitHubConfig struct {
	Token  string `yaml:"token"`
	APIURL string `yaml:"api_url"`
}

type CacheConfig struct {
	Location   string `yaml:"location"`
	MaxAgeDays int    `yaml:"max_age_days"`
}

type ExportConfig struct {
	DefaultFormat  string `yaml:"default_format"`
	IncludeRawJSON bool   `yaml:"include_raw_json"`
}

type FetchConfig struct {
	BatchSize       int `yaml:"batch_size"`
	RateLimitBuffer int `yaml:"rate_limit_buffer"`
}

func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()

	return &Config{
		GitHub: GitHubConfig{
			Token:  os.Getenv("GITHUB_TOKEN"),
			APIURL: "https://api.github.com",
		},
		Cache: CacheConfig{
			Location:   filepath.Join(homeDir, ".pr-analyzer"),
			MaxAgeDays: 90,
		},
		Export: ExportConfig{
			DefaultFormat:  "jsonl",
			IncludeRawJSON: false,
		},
		Fetch: FetchConfig{
			BatchSize:       100,
			RateLimitBuffer: 100,
		},
	}
}

func Load(path string) (*Config, error) {
	config := DefaultConfig()

	data, err := os.ReadFile(filepath.Clean(path)) // #nosec G304 - path is validated by caller
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	// Override with environment variables if set
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.GitHub.Token = token
	}

	return config, nil
}

func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

func (c *Config) CacheDB() string {
	return filepath.Join(c.Cache.Location, "cache.db")
}

func (c *Config) ConfigPath() string {
	return filepath.Join(c.Cache.Location, "config.yaml")
}

func (c *Config) ShouldPruneCache(lastUpdate time.Time) bool {
	maxAge := time.Duration(c.Cache.MaxAgeDays) * 24 * time.Hour
	return time.Since(lastUpdate) > maxAge
}
