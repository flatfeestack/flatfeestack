package main

import (
	"fmt"
	"os"
)

func setEnvs() error {
	err := os.Setenv("GO_GIT_BASE_PATH", "/tmp")
	if err != nil {
		return err
	}
	err = os.Setenv("GO_GIT_DEFAULT_BRANCH", "master-3.x")
	fmt.Println("envs set")
	return err
}
