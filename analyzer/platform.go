package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

func getPlatformInformation(src string, since time.Time, until time.Time) ([]GQLIssue, []GQLPullRequest, error) {
	// Check if repository is on Github
	if strings.Contains(src, "github.com") {
		platformInformationStart := time.Now()
		issues, pullRequests, err := getGithubPlatformInformation(src, since, until)
		platformInformationEnd := time.Now()
		fmt.Printf("---> platform information collection in %dms\n", platformInformationEnd.Sub(platformInformationStart).Milliseconds())
		return issues, pullRequests, err
	} else {
		return []GQLIssue{}, []GQLPullRequest{}, errors.New("repository is not on a platform that is supported")
	}
}

func getPlatformInformationFromUser(src string, issues []GQLIssue, pullRequests []GQLPullRequest, userEmail string) (PlatformUserInformation, error) {
	// Check if repository is on Github
	if strings.Contains(src, "github.com") {
		return getGithubPlatformInformationFromUser(src, issues, pullRequests, userEmail)
	} else {
		return PlatformUserInformation{}, errors.New("repository is not on a platform that is supported")
	}
}
