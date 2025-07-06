// Forge - Automated Docker container deployment tool for VPS environments.
// Monitors Git repositories and redeploys containers on new commits.
// Copyright (C) 2025 Artemii Fedotov
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.
package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	GITHUB_HOST = "github.com"
	GITLAB_HOST = "gitlab.com"
)

type Config struct {
	ObserverInterval time.Duration
	HTTPTimeout      time.Duration
	Repository       *url.URL
	CloneDir         string
	LogOutputDir     string
	AccessToken      string
}

type configFile struct {
	Config struct {
		Repository   string         `yaml:"repository_url"`
		LogOutputDir string         `yaml:"log_output_dir"`
		Git          gitConfig      `yaml:"git"`
		Observer     observerConfig `yaml:"observer"`
		HttpClient   httpConfig     `yaml:"http_client"`
	} `yaml:"config"`
}

type gitConfig struct {
	CloneDir string `yaml:"clone_dir"`
}

type observerConfig struct {
	Interval int `yaml:"interval"`
}

type httpConfig struct {
	Timeout int `yaml:"timeout"`
}

func configFileDefaults() *configFile {
	cfg := configFile{}
	cfg.Config.Repository = "https://github.com/makefolder/forge"
	cfg.Config.Git.CloneDir = "~/.forge/clone_dir"
	cfg.Config.LogOutputDir = "~/.forge/logs"
	cfg.Config.Observer.Interval = 30 // 30 seconds
	cfg.Config.HttpClient.Timeout = 2 // 2 seconds
	return &cfg
}

func MustParse(dir string) *Config {
	var cfg configFile
	accessToken := os.Getenv("ACCESS_TOKEN")
	if len(accessToken) == 0 {
		panic("No git access token provided (ACCESS_TOKEN environment variable)")
	}

	file, err := os.ReadFile(dir)
	if err != nil {
		panic(fmt.Errorf("Failed to read config file: %w", err))
	}

	if err := yaml.Unmarshal(file, &cfg); err != nil {
		panic(fmt.Errorf("Failed to unmarshal config file: %w", err))
	}

	if len(cfg.Config.LogOutputDir) == 0 {
		panic("Invalid log output directory")
	}

	repo, err := url.Parse(cfg.Config.Repository)
	if err != nil || repo == nil {
		panic(fmt.Errorf("Failed to parse repository URL: %w", err))
	}

	if repo.String() == "" {
		panic("Invalid repo URL")
	}

	switch repo.Hostname() {
	case GITHUB_HOST:
	case GITLAB_HOST:
	default:
		panic("Invalid git host (supported: `github.com` or `gitlab.com`)")
	}

	if cfg.Config.Git.CloneDir == "" {
		panic("Invalid git clone directory")
	}

	if cfg.Config.Observer.Interval == 0 {
		panic("Invalid observer interval")
	}

	if cfg.Config.HttpClient.Timeout == 0 {
		panic("Invalid http client timeout")
	}

	cfg.Config.Git.CloneDir = strings.TrimRight(cfg.Config.Git.CloneDir, "/")
	if strings.HasPrefix(cfg.Config.Git.CloneDir, "~") {
		cfg.Config.Git.CloneDir = expandTilde(cfg.Config.Git.CloneDir)
	}

	cfg.Config.LogOutputDir = strings.TrimRight(cfg.Config.LogOutputDir, "/")
	if strings.HasPrefix(cfg.Config.LogOutputDir, "~") {
		cfg.Config.LogOutputDir = expandTilde(cfg.Config.LogOutputDir)
	}

	return &Config{
		ObserverInterval: time.Duration(cfg.Config.Observer.Interval),
		HTTPTimeout:      time.Duration(cfg.Config.HttpClient.Timeout),
		CloneDir:         cfg.Config.Git.CloneDir,
		LogOutputDir:     cfg.Config.LogOutputDir,
		Repository:       repo,
		AccessToken:      accessToken,
	}
}

func expandTilde(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("failed to get home directory: %w", err))
	}

	if len(path) > 1 {
		return filepath.Join(home, path[1:]) // Strip just the "~"
	}
	return home // If just "~", use home directly
}
