package main

import (
	"errors"
	"strings"
	"time"
)

func getPlatformInformation(src string, since time.Time, until time.Time) ([]GQLIssue, error) {
	// Check if repository is on Github
	if strings.Contains(src, "github.com") {
		return getGithubPlatformInformation(src, since, until)
	} else {
		return []GQLIssue{}, errors.New("repository is not on a platform that is supported ")
	}
}

func getPlatformInformationFromUser(src string, issues []GQLIssue, userEmail string) (IssueUserInformation, error) {
	// Check if repository is on Github
	if strings.Contains(src, "github.com") {
		return getGithubPlatformInformationFromUser(src, issues, userEmail)
	} else {
		return IssueUserInformation{}, errors.New("repository is not on a platform that is supported ")
	}
}
