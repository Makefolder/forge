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
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const CONFIG_NAME = "default.yaml"

func Generate(dir string) error {
	if dir == "" {
		return errors.New("Invalid path")
	}

	dir = strings.TrimRight(dir, "/")
	if strings.HasPrefix(dir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		if len(dir) > 1 {
			dir = filepath.Join(home, dir[1:]) // Strip just the "~"
		} else {
			dir = home // If just "~", use home directly
		}
	} else if strings.HasPrefix(dir, ".") {
		absPath, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		if len(dir) > 1 {
			dir = filepath.Join(absPath, dir[1:])
		} else {
			dir = absPath
		}
	}

	cfg := configFileDefaults()
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	fileDir := fmt.Sprintf("%s/%s", dir, CONFIG_NAME)
	if err := os.WriteFile(fileDir, b, 0666); err != nil {
		return err
	}
	return nil
}
