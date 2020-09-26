package main

import (
	"encoding/json"
	"testing"
	"time"
)

type TestGithubClientWrapperClient struct {
	RequestRepoInfo *RequestGQLRepositoryInformation
}

func (g *TestGithubClientWrapperClient) Query(query string) ([]byte, error) {
	return json.Marshal(g.RequestRepoInfo)
}

func TestGithubClientWrapperClient_Query(t *testing.T) {

	resRepo := RequestGQLRepositoryInformation{
		Data: GQLData{
			Repository: GQLRepository{
				Issues:       GQLIssueConnection{
					Nodes:    nil,
					PageInfo: GQLPageInfo{
						EndCursor:   "",
						HasNextPage: false,
					},
				},
				Issue:        GQLIssue{
					Title:     "title",
					Number:    0,
					Author:    GQLActor{
						Login: "guil",
					},
					Comments:  GQLIssueCommentConnection{
						Nodes:    []GQLIssueComment{
							GQLIssueComment{
								Author:    GQLActor{
									Login: "thomas",
								},
								UpdatedAt: time.Now(),
							},
						},
						PageInfo: GQLPageInfo{},
					},
					UpdatedAt: time.Time{},
				},
				PullRequests: GQLPullRequestConnection{},
				PullRequest:  GQLPullRequest{},
				Ref:          GQLRef{},
			},
		},
	}

	GClientWrapper = &TestGithubClientWrapperClient{
		RequestRepoInfo: &resRepo,
	}
	_, _ = getGithubRepositoryIssues("neow3j", "neow3j", time.Now(), time.Now())
}