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
	"fmt"
	"log/slog"
	"os"
	"smithery/forge/internal/clients/git"
	"smithery/forge/internal/common"
)

type IDeployer interface {
	Deploy(context.Context) error
}

type Deployer struct {
	git      git.IGitClient
	cloneDir string
}

var ErrDockerfileNotExist = errors.New("dockerfile is not in the project's root directory")

func New(cloneDir string, git git.IGitClient) IDeployer {
	return &Deployer{
		cloneDir: cloneDir,
		git:      git,
	}
}

func (d *Deployer) Deploy(ctx context.Context) error {
	slog.Debug("deploy triggered")
	isEmpty, err := common.IsDirEmpty(d.cloneDir)
	if err != nil {
		return err
	}
	if !isEmpty {
		if err := common.CleanDir(d.cloneDir); err != nil {
			return err
		}
		slog.Debug("directory emptied", "clone_dir", d.cloneDir)
	}

	// step 1: clone repo
	accessToken := d.git.GetAccessToken()
	repo := d.git.GetRawRepoURL()
	if err := d.git.Clone(ctx, d.cloneDir, accessToken, repo); err != nil {
		return err
	}

	// step 2: check if there's dockerfile
	dockerfilePath := fmt.Sprintf("%s/%s/Dockerfile", d.cloneDir, d.git.GetRepoName())
	_, err = os.Stat(dockerfilePath)
	if errors.Is(err, os.ErrNotExist) {
		return ErrDockerfileNotExist
	} else if err != nil {
		return fmt.Errorf("failed to check for Dockerfile: %w", err)
	}

	// step 3: build dockerfile
	// step 4: create service file

	return nil
}
