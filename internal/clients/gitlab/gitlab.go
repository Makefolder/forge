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
package gitlab

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"smithery/forge/internal/clients/git"
	"smithery/forge/internal/clients/httpclient"
	"strings"
)

var errUnimplemented = errors.New("unimplemented")

type GitLabClient struct {
	git.Git
	base        *url.URL
	author      string
	repo        string
	accessToken string
	httpclient  *httpclient.HttpClient
}

func New(params git.GitClientParams) (git.IGitClient, error) {
	if err := git.ValidateParams(params); err != nil {
		return nil, err
	}

	repoPath := strings.TrimPrefix(params.Repository.Path, "/")
	s := strings.Split(repoPath, "/")
	if len(s) != 2 {
		return nil, git.ErrInvalidRepoURL
	}

	base := url.URL{
		Scheme: "https",
		Host:   "api.gitlab.com",
	}

	return &GitLabClient{
		base:        &base,
		accessToken: params.AccessToken,
		author:      s[0],
		repo:        s[1],
		httpclient:  params.HttpClient,
	}, nil
}

func (gl *GitLabClient) Ping(_ context.Context) error {
	return errUnimplemented
}

func (gl *GitLabClient) GetRepository(_ context.Context) (*git.Repository, error) {
	return nil, errUnimplemented
}

func (gl *GitLabClient) GetRawRepoURL() string {
	return fmt.Sprintf("https://gitlab.com/%s/%s", gl.author, gl.repo)
}

func (gl *GitLabClient) GetAccessToken() string { return gl.accessToken }
func (gl *GitLabClient) GetRepoName() string    { return gl.repo }
func (gl *GitLabClient) GetRepoAuthor() string  { return gl.author }
