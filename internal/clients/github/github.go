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
package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"smithery/forge/internal/clients/git"
	"smithery/forge/internal/clients/httpclient"
	"smithery/forge/internal/common"
	"strings"
)

var errUnimplemented = errors.New("Unimplemented")

type GitHubClient struct {
	git.Git
	base        *url.URL
	author      string
	repo        string
	accessToken string
	httpclient  *httpclient.HttpClient
}

func New(repository *url.URL, accessToken string, httpclient *httpclient.HttpClient) git.IGitClient {
	if repository == nil {
		panic("Repository URL cannot be nil")
	}

	repoPath := strings.TrimPrefix(repository.Path, "/")
	s := strings.Split(repoPath, "/")
	if len(s) != 2 {
		panic("Invalid repository URL")
	}

	base := url.URL{
		Scheme: "https",
		Host:   "api.github.com",
	}

	return &GitHubClient{
		base:        &base,
		accessToken: accessToken,
		author:      s[0],
		repo:        s[1],
		httpclient:  httpclient,
	}
}

func (gh *GitHubClient) Ping(ctx context.Context) error {
	res, err := gh.httpclient.Get(ctx, gh.base.JoinPath("users", gh.author), nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if !common.IsOK(res) {
		return errors.New("API response was not OK")
	}
	return nil
}

func (gh *GitHubClient) GetRepository(ctx context.Context) (*git.Repository, error) {
	headers := make(map[string]string)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", gh.accessToken)

	url := gh.base.JoinPath("repos", gh.author, gh.repo)
	res, err := gh.httpclient.Get(ctx, url, headers)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if !common.IsOK(res) {
		return nil, errors.New("API response was not OK")
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	repo := &git.Repository{}
	if err := json.Unmarshal(data, repo); err != nil {
		return nil, err
	}
	return repo, nil
}

func (gh *GitHubClient) GetRawRepoURL() string {
	return fmt.Sprintf("https://github.com/%s/%s", gh.author, gh.repo)
}
