package main

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestSetEnvs(t *testing.T) {

	err := setEnvs()

	assert.Equal(t, nil, err)
	assert.Equal(t, os.Getenv("GO_GIT_BASE_PATH"), "/tmp")
	assert.Equal(t, os.Getenv("GO_GIT_DEFAULT_BRANCH"), "master")
}
