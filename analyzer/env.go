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
	if err != nil {
		return err
	}
	//err = os.Setenv("WEBHOOK_CALLBACK_URL","https://webhook.site/f6f2c48c-03d2-4f28-a82e-10a970f22aa0")
	return err
}
