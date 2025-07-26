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
	"github.com/docker/docker/api/types/container"
	"github.com/moby/moby/client"
)

type DockerfileDeployer struct {
	cli *client.Client
}

func NewDockerfileDeployer(cli *client.Client) *DockerfileDeployer {
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

func (df *DockerfileDeployer) Deploy(ctx context.Context) (err error) {
	var containers []container.Summary

	opts := container.ListOptions{}
	containers, err = df.cli.ContainerList(ctx, opts)
	if err != nil {
		return err
	}

	// for _, c := range containers {
	//     c.Names
	//     slices.Contains(c.Names, )
	// }
	_ = containers
	return nil
}
