package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// Tests

// setEnvs

func TestSetEnvs(t *testing.T) {

	err := setEnvs()

	assert.Equal(t, nil, err)
	assert.Equal(t, getGoGitBasePathEnv(), "/tmp")
	assert.Equal(t, getGoGitDefaultBranchEnv(), "master")
}

func TestSetEnvsDefaultValues(t *testing.T) {

	assert.Equal(t, getGoGitBasePathEnv(), "/tmp")
	assert.Equal(t, getGoGitDefaultBranchEnv(), "master")
}

func setEnvs() error {
	err := os.Setenv("GO_GIT_BASE_PATH", "/tmp")
	if err != nil {
		return err
	}
	err = os.Setenv("GO_GIT_DEFAULT_BRANCH", "master")
	if err != nil {
		return err
	}
	return err
}
