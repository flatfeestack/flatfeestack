package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGetPlatformInformationNotGithub(t *testing.T)  {
	var defaultTime time.Time

	issues, pullRequests, err := getPlatformInformation("https://gitlab.com/fdroid/fdroidclient.git", defaultTime, defaultTime)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "repository is not on a platform that is supported", err.Error())
	assert.Equal(t, []GQLIssue{}, issues)
	assert.Equal(t, []GQLPullRequest{}, pullRequests)
}

func TestGetPlatformInformationFromUserNotGithub(t *testing.T)  {
	info, err := getPlatformInformationFromUser("https://gitlab.com/fdroid/fdroidclient.git", []GQLIssue{}, []GQLPullRequest{}, "test@gmail.com")
	assert.NotEqual(t, nil, err)
	assert.Equal(t, "repository is not on a platform that is supported", err.Error())
	assert.Equal(t, PlatformUserInformation{}, info)
}
