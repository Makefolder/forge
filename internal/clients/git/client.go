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
	"log/slog"
	"os"
	"time"

	"github.com/go-git/go-git/v5"
)

type IGitClient interface {
	Ping(context.Context) error
	GetRepository(context.Context) (*Repository, error)
	Clone(ctx context.Context, cloneDir, repoURL string) error
	GetRawRepoURL() string
}

type Git struct{}

func (g *Git) Clone(ctx context.Context, cloneDir, repoURL string) error {
	slog.Info("Clonning repository", "clone_dir", cloneDir, "repo_url", repoURL)
	repo, err := git.PlainCloneContext(ctx, cloneDir, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}
	slog.Info("Cloned repository successfully",
		"clone_dir", cloneDir, "repo_url", repoURL, "repository", repo)
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
