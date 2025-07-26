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
package common

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type DeployerType int

const (
	Dockerfile DeployerType = iota
	DockerCompose
	Kubernetes
	Podman
	Buildah
	UnknownContainerTool
)

const getAllDirNames int = -1

var deployerSignatures = map[DeployerType]string{
	Dockerfile:    "Dockerfile",
	DockerCompose: "docker-compose",
}

func IsOK(res *http.Response) bool {
	return res != nil &&
		res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices
}

func IsDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

func CleanDir(dir string) error {
	var err error
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(getAllDirNames)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func GetDeployerType(dir string) (DeployerType, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return UnknownContainerTool, err
	}

	if len(entries) == 0 {
		return UnknownContainerTool, fmt.Errorf("no entries found in %s", dir)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()

		for deployer, pattern := range deployerSignatures {
			if strings.Contains(name, pattern) {
				return deployer, nil
			}
		}
	}

	return UnknownContainerTool, nil
}
