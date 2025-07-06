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
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"smithery/forge/internal/clients/git"
	"smithery/forge/internal/clients/github"
	"smithery/forge/internal/clients/gitlab"
	"smithery/forge/internal/clients/httpclient"
	"smithery/forge/internal/config"
	"smithery/forge/internal/deployer"
	"smithery/forge/internal/observer"
	"strings"
	"time"
)

const (
	UNSPECIFIED_PATH = ""
	LOG_FMT_TEXT     = "text"
	LOG_FMT_JSON     = "json"
)

func main() {
	// config init
	var (
		isConfigGenerate bool
		dir              string
		logFmt           string
		slogHandler      slog.Handler
	)

	flag.BoolVar(&isConfigGenerate, "g", false, "generate config.yaml (should be used with '-d')")
	flag.StringVar(&dir, "d", UNSPECIFIED_PATH, "directory to config.yaml")
	flag.StringVar(&logFmt, "fmt", LOG_FMT_TEXT, "log format (json/text; default: text)")
	flag.Parse()

	if isConfigGenerate && dir == UNSPECIFIED_PATH {
		panic("Cannot generate config.yaml: no target path specified")
	} else if isConfigGenerate {
		err := config.Generate(dir)
		if err != nil {
			msg := fmt.Errorf("Failed to generate config: %w", err)
			panic(msg)
		}
		return
	}

	if dir == UNSPECIFIED_PATH {
		panic("No config file specified")
	}
	cfg := config.MustParse(dir)

	// slog setup
	opts := slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
	}

	logName := fmt.Sprintf("%s.%s", time.Now().Format(time.DateOnly), "log")
	filePath := fmt.Sprintf("%s/%s", cfg.LogOutputDir, logName)
	logFile, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(fmt.Errorf("Failed to write to log file: %w", err))
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

	// http client init
	ctx := context.Background()
	httpclient := httpclient.New(cfg.HTTPTimeout * time.Second)

	// git init
	var git git.IGitClient
	switch cfg.Repository.Hostname() {
	case config.GITHUB_HOST:
		git = github.New(cfg.Repository, cfg.AccessToken, httpclient)
	case config.GITLAB_HOST:
		git = gitlab.New(cfg.Repository, cfg.AccessToken, httpclient)
	default:
		panic(fmt.Sprintf("git client is not specified for host %s", cfg.Repository.Hostname()))
	}
	if err := git.Ping(ctx); err != nil {
		panic(fmt.Errorf("Failed to ping repository: %w", err))
	}

	// deployer init
	d := deployer.New(git.GetRawRepoURL(), git)

	// observer init & observe
	params := observer.ObserverParams{
		Git:      git,
		Interval: time.Duration(cfg.ObserverInterval),
		Subscriptions: []func(context.Context) error{
			d.Deploy,
		},
	}

	o := observer.New(params)
	if err := o.Observe(ctx, cfg.Repository); err != nil {
		panic(fmt.Errorf("Failed to observe: %w", err))
	}
}
