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
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"smithery/forge/internal/clients/git"
	"smithery/forge/internal/clients/github"
	"smithery/forge/internal/clients/gitlab"
	"smithery/forge/internal/clients/httpclient"
	"smithery/forge/internal/common"
	"smithery/forge/internal/config"
	"smithery/forge/internal/deployer"
	"smithery/forge/internal/observer"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	UNSPECIFIED_PATH = ""
	LOG_FMT_TEXT     = "text"
	LOG_FMT_JSON     = "json"
)

func init() {
	var isDev bool
	env := os.Getenv("ENV")
	isDev = env != "" && strings.ToLower(env) == "dev"

	if !isDev {
		return
	}

	if err := godotenv.Load(); err != nil {
		panic(fmt.Errorf("Failed to load .env file: %v", err))
	}
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	// config init
	var (
		isConfigGenerate bool
		dir              string
		logFmt           string
		slogHandler      slog.Handler
	)

	flag.BoolVar(&isConfigGenerate, "g", false, "generate config.yaml (must be used with '-d')")
	flag.StringVar(&dir, "d", UNSPECIFIED_PATH, "directory to config.yaml")
	flag.StringVar(&logFmt, "fmt", LOG_FMT_TEXT, "log format (json/text; default: text)")
	flag.Parse()

	if isConfigGenerate && dir == UNSPECIFIED_PATH {
		return errors.New("cannot generate config.yaml: no target path specified")
	} else if isConfigGenerate {
		err := config.Generate(dir)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		return nil
	}

	if dir == UNSPECIFIED_PATH {
		return errors.New("No config file specified")
	}
	cfg := config.MustParse(dir)

	// init directories
	mustInitDir(cfg.LogOutputDir, cfg.CloneDir)

	// slog setup
	logLevel, err := mapLogLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		return err
	}

	opts := slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	logName := fmt.Sprintf("%s.%s", time.Now().Format(time.DateOnly), "log")
	filePath := fmt.Sprintf("%s/%s", cfg.LogOutputDir, logName)
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to write to log file: %w", err)
	}
	defer logFile.Close()

	w := io.MultiWriter(logFile, os.Stdout)
	if strings.ToLower(logFmt) == LOG_FMT_JSON {
		slogHandler = slog.NewJSONHandler(w, &opts)
	} else {
		slogHandler = slog.NewTextHandler(w, &opts)
	}

	logger := slog.New(slogHandler)
	slog.SetDefault(logger)
	slog.Debug("slog initialised")

	// http client init
	ctx := context.Background()
	httpclient := httpclient.New(cfg.HTTPTimeout * time.Second)
	slog.Debug("http client initialised")

	// git init
	gitParams := git.GitClientParams{
		Repository:  cfg.Repository,
		AccessToken: cfg.AccessToken,
		HttpClient:  httpclient,
	}

	var git git.IGitClient
	switch cfg.Repository.Hostname() {
	case config.GITHUB_HOST:
		git, err = github.New(gitParams)
	case config.GITLAB_HOST:
		git, err = gitlab.New(gitParams)
	default:
		return fmt.Errorf("git client is not specified for host %s", cfg.Repository.Hostname())
	}
	if err != nil {
		return err
	}
	if err := git.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping repository: %w", err)
	}
	slog.Debug("git client initialised")

	// deployer init
	d := deployer.New(w, cfg.CloneDir, git)
	slog.Debug("deployer initialised")

	isEmpty, err := common.IsDirEmpty(cfg.CloneDir)
	if err != nil {
		return fmt.Errorf("failed to check whether the dir (%s) is empty: %w", cfg.CloneDir, err)
	}

	if isEmpty {
		slog.Debug("clone dir is empty")
		err := d.Deploy(ctx)
		if errors.Is(err, deployer.ErrDockerfileNotExist) {
			slog.Warn("failed initial repo cloning", "error", err.Error())
		} else if err != nil {
			return fmt.Errorf("failed initial repo cloning: %w", err)
		}
	}

	// observer init & observe
	params := observer.ObserverParams{
		Git:      git,
		Interval: time.Duration(cfg.ObserverInterval) * time.Second,
		Subscriptions: []func(context.Context) error{
			d.Deploy,
		},
	}

	o := observer.New(params)
	slog.Debug("observer created",
		slog.String("git_repository", params.Git.GetRawRepoURL()),
		slog.Int("interval", int(cfg.ObserverInterval)),
		slog.Int("subscription_length", len(params.Subscriptions)),
	)
	if err := o.Observe(ctx, cfg.Repository); err != nil {
		return fmt.Errorf("failed to observe: %w", err)
	}

	return nil
}

func mustInitDir(dirs ...string) {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			panic(fmt.Errorf("failed to init directory (%s): %w", dir, err))
		}
	}
}

func mapLogLevel(env string) (slog.Level, error) {
	if env == "" {
		return slog.LevelError, nil
	}

	switch strings.ToLower(env) {
	case "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	}

	return 0, fmt.Errorf("invalid slog level option (%s)", env)
}
