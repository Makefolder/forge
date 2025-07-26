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
	"fmt"
	"log/slog"
	"slices"

	"github.com/docker/docker/api/types/container"
	"github.com/moby/moby/client"
)

type DockerfileDeployer struct {
	cli *client.Client
}

func NewDockerfileDeployer(cli *client.Client) IDeployer {
	return &DockerfileDeployer{
		cli: cli,
	}
}

// Note to self:
// FromEnv uses the following environment variables:
//
//   - DOCKER_HOST ([EnvOverrideHost]) to set the URL to the docker server.
//   - DOCKER_API_VERSION ([EnvOverrideAPIVersion]) to set the version of the
//     API to use, leave empty for latest.
//   - DOCKER_CERT_PATH ([EnvOverrideCertPath]) to specify the directory from
//     which to load the TLS certificates ("ca.pem", "cert.pem", "key.pem').
//   - DOCKER_TLS_VERIFY ([EnvTLSVerify]) to enable or disable TLS verification
//     (off by default).

func (df *DockerfileDeployer) Deploy(ctx context.Context, params DeployParams) error {
	var containers []container.Summary
	containers, err := df.cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return err
	}

	if err := df.safeRemoveContainer(ctx, containers, params.ContainerName); err != nil {
		return err
	}

	res, err := df.cli.ContainerCreate(ctx, nil, nil, nil, nil, params.ContainerName)
	if err != nil {
		return err
	}

	if len(res.Warnings) > 0 {
		warnMsg := fmt.Sprintf("warning occured during %s container deployment",
			params.ContainerName)

		for _, warn := range res.Warnings {
			slog.Warn(warnMsg, "msg", warn)
		}
	}

	return nil
}

// Removes container if exists
func (df *DockerfileDeployer) safeRemoveContainer(
	ctx context.Context,
	containers []container.Summary,
	containerName string,
) error {
	for _, c := range containers {
		if slices.Contains(c.Names, containerName) {
			if isStoppable(c.State) {
				copts := container.StopOptions{}
				err := df.cli.ContainerStop(ctx, c.ID, copts)
				if err != nil {
					return err
				}
			}
			err := df.cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{
				RemoveVolumes: true,
				RemoveLinks:   true,
				Force:         false,
			})
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func isStoppable(state container.ContainerState) bool {
	switch state {
	case container.StateRunning, container.StatePaused, container.StateRestarting:
		return true
	}
	return false
}
