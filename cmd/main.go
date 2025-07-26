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
	"github.com/moby/moby/client"
)

const (
	unspecifiedPath = ""
	logFmtText      = "text"
	logFmtJSON      = "json"
)

func init() {
	var isDev bool
	env := os.Getenv("ENV")
	isDev = env != "" && strings.ToLower(env) == "dev"

	if !isDev {
		return
	}

	if err := godotenv.Load(); err != nil {
		panic(fmt.Errorf("failed to load .env file: %v", err))
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
	flag.StringVar(&dir, "d", unspecifiedPath, "directory to config.yaml")
	flag.StringVar(&logFmt, "fmt", logFmtText, "log format (json/text; default: text)")
	flag.Parse()

	if isConfigGenerate && dir == unspecifiedPath {
		return errors.New("cannot generate config.yaml: no target path specified")
	} else if isConfigGenerate {
		err := config.Generate(dir)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}
		return nil
	}

	if dir == unspecifiedPath {
		return errors.New("no config file specified")
	}
	cfg := config.MustParse(dir)

	// init directories
	if err := initDir(cfg.LogOutputDir, cfg.CloneDir); err != nil {
		return err
	}

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
	if strings.ToLower(logFmt) == logFmtJSON {
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
	case config.GithubHost:
		git, err = github.New(gitParams)
	case config.GitlabHost:
		git, err = gitlab.New(gitParams)
	default:
		return fmt.Errorf("git client is not specified for host %s", cfg.Repository.Hostname())
	}
	if err != nil {
		return fmt.Errorf("failed to initialise git client: %w", err)
	}
	if err := git.Ping(ctx); err != nil {
		return fmt.Errorf("failed to ping repository: %w", err)
	}
	slog.Debug("git client initialised")

	// deployer init
	dockerClient, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to initialise docker client: %w", err)
	}

	defer func() {
		if err == nil {
			err = dockerClient.Close()
		}
	}()

	// todo!: Determine deployer type
	// common.GetDeployerType()
	d := deployer.NewDockerfileDeployer(dockerClient)

	diParams := deployer.DIParams{
		Deployer: d,
		Git:      git,
		CloneDir: cfg.CloneDir,
	}

	di := deployer.NewDeployInvoker(diParams)
	slog.Debug("deploy invoker initialised")

	isEmpty, err := common.IsDirEmpty(cfg.CloneDir)
	if err != nil {
		return fmt.Errorf("failed to check whether the dir (%s) is empty: %w", cfg.CloneDir, err)
	}

	if isEmpty {
		slog.Debug("clone dir is empty")
		err := di.Deploy(ctx)
		if errors.Is(err, deployer.ErrDockerfileNotExist) {
			// at this point, deployment is not going to happen but notifications will be sent
			slog.Warn("failed initial deployment", "error", err.Error())
		} else if err != nil {
			return fmt.Errorf("failed initial deployment: %w", err)
		}
	}

	// observer init & observe
	params := observer.ObserverParams{
		Git:      git,
		Interval: time.Duration(cfg.ObserverInterval) * time.Second,
		Subscriptions: []func(context.Context) error{
			di.Deploy,
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

func initDir(dirs ...string) error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to init directory (%s): %w", dir, err)
		}
	}
	return nil
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
