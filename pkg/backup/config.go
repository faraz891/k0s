//go:build !windows
// +build !windows

/*
Copyright 2022 k0s authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package backup

import (
	"fmt"
	"io"
	"os"
	"path"

	"github.com/sirupsen/logrus"

	"github.com/k0sproject/k0s/internal/pkg/file"
)

type configurationStep struct {
	path               string
	restoredConfigPath string
}

func newConfigurationStep(path string, restoredConfigPath string) *configurationStep {
	return &configurationStep{
		path:               path,
		restoredConfigPath: restoredConfigPath,
	}
}

func (c configurationStep) Name() string {
	return c.path
}

func (c configurationStep) Backup() (StepResult, error) {
	_, err := os.Stat(c.path)
	if os.IsNotExist(err) {
		logrus.Warn("default k0s.yaml is used, do not back it up")
		return StepResult{}, nil
	}
	if err != nil {
		return StepResult{}, fmt.Errorf("can't backup `%s`: %v", c.path, err)
	}
	return StepResult{filesForBackup: []string{c.path}}, nil
}

func (c configurationStep) Restore(restoreFrom, restoreTo string) error {
	objectPathInArchive := path.Join(restoreFrom, "k0s.yaml")

	if !file.Exists(objectPathInArchive) {
		logrus.Debugf("%s does not exist in the backup file", objectPathInArchive)
		return nil
	}
	logrus.Infof("Previously used k0s.yaml saved under the data directory `%s`", restoreTo)

	if c.restoredConfigPath == "-" {
		f, err := os.Open(objectPathInArchive)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(f, os.Stdout)
		return err
	}
	logrus.Infof("restoring from `%s` to `%s`", objectPathInArchive, c.restoredConfigPath)
	return file.Copy(objectPathInArchive, c.restoredConfigPath)
}
