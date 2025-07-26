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
package deployer

import (
	"context"
	"errors"
	"log/slog"
	"smithery/forge/internal/clients/git"
	"smithery/forge/internal/common"
)

var ErrDockerfileNotExist = errors.New("dockerfile is not in the project's root directory")

type IDeployer interface {
	Deploy(context.Context, DeployParams) error
}

type DeployInvoker struct {
	deployer IDeployer
	git      git.IGitClient
	cloneDir string
}

type DeployParams struct {
	ContainerName string
}

type DIParams struct {
	Deployer IDeployer
	Git      git.IGitClient
	CloneDir string
}

func NewDeployInvoker(params DIParams) *DeployInvoker {
	return &DeployInvoker{
		deployer: params.Deployer,
		git:      params.Git,
		cloneDir: params.CloneDir,
	}
}

func (di *DeployInvoker) Deploy(ctx context.Context) error {
	slog.Debug("deploy triggered")
	isEmpty, err := common.IsDirEmpty(di.cloneDir)
	if err != nil {
		return err
	}
	if !isEmpty {
		if err := common.CleanDir(di.cloneDir); err != nil {
			return err
		}
		slog.Debug("directory emptied", "clone_dir", di.cloneDir)
	}

	accessToken := di.git.GetAccessToken()
	repo := di.git.GetRawRepoURL()
	if err := di.git.Clone(ctx, di.cloneDir, accessToken, repo); err != nil {
		return err
	}

	return di.deployer.Deploy(ctx, DeployParams{
		ContainerName: di.git.GetRepoName(),
	})
}
