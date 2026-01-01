// Copyright 2025 GeekchanskiY
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	_ "embed"
	"fmt"
	"os"
	"path"
)

//go:embed templates/config_template.yml
var configTemplate string

func Initialize() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	_, err = os.Stat(path.Join(cwd, "migratigo"))
	if err == nil {
		return ErrAlreadyInitialized
	}

	err = os.Mkdir("migratigo", 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(cwd, "migratigo", "config.yaml"))
	if err != nil {
		return err
	}

	defer func() {
		err = f.Close()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, "error closing file (migratigo/config.yaml):", err)
		}
	}()

	_, err = f.WriteString(configTemplate)
	if err != nil {
		return err
	}

	return nil
}
