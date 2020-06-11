package main

import "os"

func setEnvs() error {
	err := os.Setenv("GO_GIT_BASE_PATH", "/tmp");
	if err != nil {
		return err
	}
	return os.Setenv("GO_GIT_DEFAULT_BRANCH", "master");
}
