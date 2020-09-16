package main

import (
	"os"
)

func setEnvs() error {
	err := os.Setenv("GO_GIT_BASE_PATH", "/tmp")
	if err != nil {
		return err
	}
	err = os.Setenv("GO_GIT_DEFAULT_BRANCH", "master")
	return err
}
