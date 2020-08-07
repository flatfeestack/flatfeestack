package main

import (
	"errors"
	"strings"
	"time"
)

func getPlatformInformation(src string, since time.Time, until time.Time) ([]GQLIssue, []GQLPullRequest, error) {
	// Check if repository is on Github
	if strings.Contains(src, "github.com") {
		return getGithubPlatformInformation(src, since, until)
	} else {
		return []GQLIssue{}, []GQLPullRequest{}, errors.New("repository is not on a platform that is supported ")
	}
}

func getPlatformInformationFromUser(src string, issues []GQLIssue, pullRequests []GQLPullRequest, userEmail string) (PlatformUserInformation, error) {
	// Check if repository is on Github
	if strings.Contains(src, "github.com") {
		return getGithubPlatformInformationFromUser(src, issues, pullRequests, userEmail)
	} else {
		return PlatformUserInformation{}, errors.New("repository is not on a platform that is supported ")
	}
}
