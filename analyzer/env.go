package main

import (
	"os"
)

func getGoGitBasePathEnv() string {
	return GetEnvVar("GO_GIT_BASE_PATH", "/tmp")
}

func getGoGitDefaultBranchEnv() string {
	return GetEnvVar("GO_GIT_DEFAULT_BRANCH", "main")
}

// GetEnvVar get an environment variable value, or returns a fallback value
func GetEnvVar(key, fallback string) string {
	value, _ := GetEnvVarAndIfExists(key, fallback)
	return value
}

// GetEnvVarAndIfExists Retrieves an environment variable value, or returns a default (fallback) value
// It also returns true or false if the env variable exists or not
func GetEnvVarAndIfExists(key, fallback string) (string, bool) {
	value, exists := os.LookupEnv(key)
	if len(value) == 0 {
		return fallback, exists
	}
	return value, exists
}
