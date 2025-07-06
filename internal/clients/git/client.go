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
package git

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"os"
	"smithery/forge/internal/clients/httpclient"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

var (
	ErrNilRepoURL       = errors.New("repository URL cannot be nil")
	ErrNilHttpClient    = errors.New("HTTP client cannot be nil")
	ErrEmptyAccessToken = errors.New("access token cannot be empty")
	ErrInvalidRepoURL   = errors.New("invalid repository URL")
)

type IGitClient interface {
	Ping(context.Context) error
	GetRepository(context.Context) (*Repository, error)
	Clone(ctx context.Context, cloneDir, repoURL, accessToken string) error
	GetRawRepoURL() string
	GetRepoName() string
	GetRepoAuthor() string
	GetAccessToken() string
}

type GitClientParams struct {
	Repository  *url.URL
	AccessToken string
	HttpClient  *httpclient.HttpClient
}

type Git struct{}

func (g *Git) Clone(ctx context.Context, cloneDir, accessToken, repoURL string) error {
	if accessToken == "" {
		return errors.New("access token cannot be empty")
	}

	if cloneDir == "" {
		return errors.New("clone dir cannot be empty")
	}

	if repoURL == "" {
		return errors.New("repo url cannot be empty")
	}

	slog.Info("cloning repository", "clone_dir", cloneDir, "repo_url", repoURL)
	auth := &http.BasicAuth{
		Username: "bearer",
		Password: accessToken,
	}

	repo, err := git.PlainCloneContext(ctx, cloneDir, false, &git.CloneOptions{
		Auth:     auth,
		URL:      repoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}
	slog.Info("repository cloned successfully",
		"clone_dir", cloneDir, "repo_url", repoURL, "repository", repo)
	return nil
}

func ValidateParams(params GitClientParams) error {
	if params.Repository == nil {
		return ErrNilRepoURL
	}

	if params.HttpClient == nil {
		return ErrNilHttpClient
	}

	if params.AccessToken == "" {
		return ErrEmptyAccessToken
	}
	return nil
}

type Repository struct {
	Id          int64     `json:"id"`
	Name        string    `json:"name"`
	Fullname    string    `json:"full_name"`
	Description *string   `json:"description,omitempty"`
	Private     bool      `json:"private"`
	PushedAt    time.Time `json:"pushed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
