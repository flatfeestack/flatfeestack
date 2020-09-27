package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type TestGithubClientWrapperClient struct {
	RequestRepoInfo *RequestGQLRepositoryInformation
}

func (g *TestGithubClientWrapperClient) Query(query string) ([]byte, error) {
	return json.Marshal(g.RequestRepoInfo)
}

func parseRFC3339WithoutError(date string) time.Time {
	golangTime, _ := time.Parse(time.RFC3339, date)
	return golangTime
}

func TestGetGithubRepositoryIssuesFiltering(t *testing.T) {

	resRepo := RequestGQLRepositoryInformation{
		Data: GQLData{
			Repository: GQLRepository{
				Issues: GQLIssueConnection{
					Nodes: []GQLIssue{
						{
							Title:  "Issue 1",
							Number: 0,
							Author: GQLActor{
								Login: "octocat",
							},
							Comments: GQLIssueCommentConnection{
								Nodes: []GQLIssueComment{
									{
										Author: GQLActor{
											Login: "octocat",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-01-19T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-01-20T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-01-22T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-01-22T12:00:00Z"),
						},
						{
							Title:  "Issue 2",
							Number: 1,
							Author: GQLActor{
								Login: "octodog",
							},
							Comments: GQLIssueCommentConnection{
								Nodes: []GQLIssueComment{
									{
										Author: GQLActor{
											Login: "octocat",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-02-19T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-02-20T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
						},
						{
							Title:  "Issue 3",
							Number: 2,
							Author: GQLActor{
								Login: "octodog",
							},
							Comments: GQLIssueCommentConnection{
								Nodes: []GQLIssueComment{
									{
										Author: GQLActor{
											Login: "octocat",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-03-19T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-03-20T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-03-22T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-03-22T12:00:00Z"),
						},
					},
					PageInfo: GQLPageInfo{
						EndCursor:   "",
						HasNextPage: false,
					},
				},
			},
		},
	}

	GClientWrapper = &TestGithubClientWrapperClient{
		RequestRepoInfo: &resRepo,
	}

	startTime := parseRFC3339WithoutError("2020-02-21T12:00:00Z")
	var endTime time.Time

	issues, err := getGithubRepositoryIssues("neow3j", "neow3j", startTime, endTime)

	expectedOutcome := []GQLIssue{
		{
			Title:  "Issue 2",
			Number: 1,
			Author: GQLActor{
				Login: "octodog",
			},
			Comments: GQLIssueCommentConnection{
				Nodes: []GQLIssueComment{
					{
						Author: GQLActor{
							Login: "octoduck",
						},
						UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
					},
				},
				PageInfo: GQLPageInfo{
					EndCursor:   "",
					HasNextPage: false,
				},
			},
			UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
		},
		{
			Title:  "Issue 3",
			Number: 2,
			Author: GQLActor{
				Login: "octodog",
			},
			Comments: GQLIssueCommentConnection{
				Nodes: []GQLIssueComment{
					{
						Author: GQLActor{
							Login: "octocat",
						},
						UpdatedAt: parseRFC3339WithoutError("2020-03-19T12:00:00Z"),
					},
					{
						Author: GQLActor{
							Login: "octodog",
						},
						UpdatedAt: parseRFC3339WithoutError("2020-03-20T12:00:00Z"),
					},
					{
						Author: GQLActor{
							Login: "octoduck",
						},
						UpdatedAt: parseRFC3339WithoutError("2020-03-22T12:00:00Z"),
					},
				},
				PageInfo: GQLPageInfo{
					EndCursor:   "",
					HasNextPage: false,
				},
			},
			UpdatedAt: parseRFC3339WithoutError("2020-03-22T12:00:00Z"),
		},
	}

	assert.Equal(t, expectedOutcome, issues)
	assert.Equal(t, nil, err)
}

func TestGetGithubRepositoryIssuesNoFiltering(t *testing.T) {

	resRepo := RequestGQLRepositoryInformation{
		Data: GQLData{
			Repository: GQLRepository{
				Issues: GQLIssueConnection{
					Nodes: []GQLIssue{
						{
							Title:  "Issue 1",
							Number: 0,
							Author: GQLActor{
								Login: "octocat",
							},
							Comments: GQLIssueCommentConnection{
								Nodes: []GQLIssueComment{
									{
										Author: GQLActor{
											Login: "octocat",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-01-19T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-01-20T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-01-22T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-01-22T12:00:00Z"),
						},
						{
							Title:  "Issue 2",
							Number: 1,
							Author: GQLActor{
								Login: "octodog",
							},
							Comments: GQLIssueCommentConnection{
								Nodes: []GQLIssueComment{
									{
										Author: GQLActor{
											Login: "octocat",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-02-19T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-02-20T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
						},
						{
							Title:  "Issue 3",
							Number: 2,
							Author: GQLActor{
								Login: "octodog",
							},
							Comments: GQLIssueCommentConnection{
								Nodes: []GQLIssueComment{
									{
										Author: GQLActor{
											Login: "octocat",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-03-19T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-03-20T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										UpdatedAt: parseRFC3339WithoutError("2020-03-22T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-03-22T12:00:00Z"),
						},
					},
					PageInfo: GQLPageInfo{
						EndCursor:   "",
						HasNextPage: false,
					},
				},
			},
		},
	}

	GClientWrapper = &TestGithubClientWrapperClient{
		RequestRepoInfo: &resRepo,
	}

	var defaultTime time.Time

	issues, err := getGithubRepositoryIssues("neow3j", "neow3j", defaultTime, defaultTime)

	assert.Equal(t, resRepo.Data.Repository.Issues.Nodes, issues)
	assert.Equal(t, nil, err)
}

func TestGetGithubRepositoryPullRequestsFiltering(t *testing.T) {

	resRepo := RequestGQLRepositoryInformation{
		Data: GQLData{
			Repository: GQLRepository{
				PullRequests: GQLPullRequestConnection{
					Nodes: []GQLPullRequest{
						{
							Title:  "PR1",
							Number: 0,
							Author: GQLActor{
								Login: "octocat",
							},
							State: "MERGED",
							Reviews: GQLPullRequestReviewConnection{
								Nodes: []GQLPullRequestReview{
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "COMMENTED",
										UpdatedAt: parseRFC3339WithoutError("2020-02-21T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										State:     "CODE REVIEW",
										UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										State:     "APPROVED",
										UpdatedAt: parseRFC3339WithoutError("2020-02-23T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-02-23T12:00:00Z"),
						},
						{
							Title:  "PR2",
							Number: 1,
							Author: GQLActor{
								Login: "octodog",
							},
							State: "MERGED",
							Reviews: GQLPullRequestReviewConnection{
								Nodes: []GQLPullRequestReview{
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "COMMENTED",
										UpdatedAt: parseRFC3339WithoutError("2020-03-21T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octocat",
										},
										State:     "CODE REVIEW",
										UpdatedAt: parseRFC3339WithoutError("2020-04-22T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "APPROVED",
										UpdatedAt: parseRFC3339WithoutError("2020-05-23T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-05-23T12:00:00Z"),
						},
						{
							Title:  "PR3",
							Number: 2,
							Author: GQLActor{
								Login: "octodog",
							},
							State: "MERGED",
							Reviews: GQLPullRequestReviewConnection{
								Nodes: []GQLPullRequestReview{
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "COMMENTED",
										UpdatedAt: parseRFC3339WithoutError("2020-05-21T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octocat",
										},
										State:     "CODE REVIEW",
										UpdatedAt: parseRFC3339WithoutError("2020-06-22T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "APPROVED",
										UpdatedAt: parseRFC3339WithoutError("2020-06-23T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-06-23T12:00:00Z"),
						},
					},
					PageInfo: GQLPageInfo{
						EndCursor:   "",
						HasNextPage: false,
					},
				},
			},
		},
	}

	GClientWrapper = &TestGithubClientWrapperClient{
		RequestRepoInfo: &resRepo,
	}

	startTime := parseRFC3339WithoutError("2020-03-23T12:00:00Z")
	var endTime time.Time

	issues, err := getGithubRepositoryPullRequests("neow3j", "neow3j", startTime, endTime)

	expectedOutcome := []GQLPullRequest{
		{
			Title:  "PR2",
			Number: 1,
			Author: GQLActor{
				Login: "octodog",
			},
			State: "MERGED",
			Reviews: GQLPullRequestReviewConnection{
				Nodes: []GQLPullRequestReview{
					{
						Author: GQLActor{
							Login: "octocat",
						},
						State:     "CODE REVIEW",
						UpdatedAt: parseRFC3339WithoutError("2020-04-22T12:00:00Z"),
					},
					{
						Author: GQLActor{
							Login: "octoduck",
						},
						State:     "APPROVED",
						UpdatedAt: parseRFC3339WithoutError("2020-05-23T12:00:00Z"),
					},
				},
				PageInfo: GQLPageInfo{
					EndCursor:   "",
					HasNextPage: false,
				},
			},
			UpdatedAt: parseRFC3339WithoutError("2020-05-23T12:00:00Z"),
		},
		{
			Title:  "PR3",
			Number: 2,
			Author: GQLActor{
				Login: "octodog",
			},
			State: "MERGED",
			Reviews: GQLPullRequestReviewConnection{
				Nodes: []GQLPullRequestReview{
					{
						Author: GQLActor{
							Login: "octoduck",
						},
						State:     "COMMENTED",
						UpdatedAt: parseRFC3339WithoutError("2020-05-21T12:00:00Z"),
					},
					{
						Author: GQLActor{
							Login: "octocat",
						},
						State:     "CODE REVIEW",
						UpdatedAt: parseRFC3339WithoutError("2020-06-22T12:00:00Z"),
					},
					{
						Author: GQLActor{
							Login: "octoduck",
						},
						State:     "APPROVED",
						UpdatedAt: parseRFC3339WithoutError("2020-06-23T12:00:00Z"),
					},
				},
				PageInfo: GQLPageInfo{
					EndCursor:   "",
					HasNextPage: false,
				},
			},
			UpdatedAt: parseRFC3339WithoutError("2020-06-23T12:00:00Z"),
		},
	}

	assert.Equal(t, expectedOutcome, issues)
	assert.Equal(t, nil, err)
}

func TestGetGithubRepositoryPullRequestsNoFiltering(t *testing.T) {

	resRepo := RequestGQLRepositoryInformation{
		Data: GQLData{
			Repository: GQLRepository{
				PullRequests: GQLPullRequestConnection{
					Nodes: []GQLPullRequest{
						{
							Title:  "PR1",
							Number: 0,
							Author: GQLActor{
								Login: "octocat",
							},
							State: "MERGED",
							Reviews: GQLPullRequestReviewConnection{
								Nodes: []GQLPullRequestReview{
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "COMMENTED",
										UpdatedAt: parseRFC3339WithoutError("2020-02-21T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										State:     "CODE REVIEW",
										UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octodog",
										},
										State:     "APPROVED",
										UpdatedAt: parseRFC3339WithoutError("2020-02-23T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-02-23T12:00:00Z"),
						},
						{
							Title:  "PR2",
							Number: 1,
							Author: GQLActor{
								Login: "octodog",
							},
							State: "MERGED",
							Reviews: GQLPullRequestReviewConnection{
								Nodes: []GQLPullRequestReview{
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "COMMENTED",
										UpdatedAt: parseRFC3339WithoutError("2020-03-21T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octocat",
										},
										State:     "CODE REVIEW",
										UpdatedAt: parseRFC3339WithoutError("2020-04-22T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "APPROVED",
										UpdatedAt: parseRFC3339WithoutError("2020-05-23T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-05-23T12:00:00Z"),
						},
						{
							Title:  "PR3",
							Number: 2,
							Author: GQLActor{
								Login: "octodog",
							},
							State: "MERGED",
							Reviews: GQLPullRequestReviewConnection{
								Nodes: []GQLPullRequestReview{
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "COMMENTED",
										UpdatedAt: parseRFC3339WithoutError("2020-05-21T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octocat",
										},
										State:     "CODE REVIEW",
										UpdatedAt: parseRFC3339WithoutError("2020-06-22T12:00:00Z"),
									},
									{
										Author: GQLActor{
											Login: "octoduck",
										},
										State:     "APPROVED",
										UpdatedAt: parseRFC3339WithoutError("2020-06-23T12:00:00Z"),
									},
								},
								PageInfo: GQLPageInfo{
									EndCursor:   "",
									HasNextPage: false,
								},
							},
							UpdatedAt: parseRFC3339WithoutError("2020-06-23T12:00:00Z"),
						},
					},
					PageInfo: GQLPageInfo{
						EndCursor:   "",
						HasNextPage: false,
					},
				},
			},
		},
	}

	GClientWrapper = &TestGithubClientWrapperClient{
		RequestRepoInfo: &resRepo,
	}

	var defaultTime time.Time

	pullRequests, err := getGithubRepositoryPullRequests("neow3j", "neow3j", defaultTime, defaultTime)

	assert.Equal(t, resRepo.Data.Repository.PullRequests.Nodes, pullRequests)
	assert.Equal(t, nil, err)
}

func TestGetOwnerAndNameOfGithubUrl(t *testing.T) {
	owner, name := getOwnerAndNameOfGithubUrl("https://github.com/ownerName/repoName.git")
	assert.Equal(t, "ownerName", owner)
	assert.Equal(t, "repoName", name)
}

func TestGetGithubUsernameFromGitEmailFound(t *testing.T) {
	resRepo := RequestGQLRepositoryInformation{
		Data: GQLData{
			Repository: GQLRepository{
				Ref: GQLRef{
					Target: GQLGitObject{
						History: GQLCommitHistoryConnection{
							Nodes: []GQLCommit{
								{
									Author: GQLGitActor{
										User: GQLActor{
											Login: "octodog",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	GClientWrapper = &TestGithubClientWrapperClient{
		RequestRepoInfo: &resRepo,
	}

	name, err := getGithubUsernameFromGitEmail("octocat", "sampleProject", "octodog@github.com")
	assert.Equal(t, "octodog", name)
	assert.Equal(t, nil, err)
}

func TestGetGithubUsernameFromGitEmailNotFound(t *testing.T) {
	resRepo := RequestGQLRepositoryInformation{
		Data: GQLData{
			Repository: GQLRepository{
				Ref: GQLRef{
					Target: GQLGitObject{
						History: GQLCommitHistoryConnection{
							Nodes: []GQLCommit{
							},
						},
					},
				},
			},
		},
	}

	GClientWrapper = &TestGithubClientWrapperClient{
		RequestRepoInfo: &resRepo,
	}

	name, err := getGithubUsernameFromGitEmail("octocat", "sampleProject", "octodog@github.com")
	assert.Equal(t, "", name)
	assert.NotEqual(t,nil, err)
	assert.Equal(t, "could not find user", err.Error())
}

func TestGetPullRequestReviewStateArray(t *testing.T)  {
	pullRequest := GQLPullRequest{
		Title:  "PR1",
		Number: 0,
		Author: GQLActor{
			Login: "octocat",
		},
		State: "MERGED",
		Reviews: GQLPullRequestReviewConnection{
			Nodes: []GQLPullRequestReview{
				{
					Author: GQLActor{
						Login: "octoduck",
					},
					State:     "COMMENTED",
					UpdatedAt: parseRFC3339WithoutError("2020-02-21T12:00:00Z"),
				},
				{
					Author: GQLActor{
						Login: "octodog",
					},
					State:     "CODE REVIEW",
					UpdatedAt: parseRFC3339WithoutError("2020-02-22T12:00:00Z"),
				},
				{
					Author: GQLActor{
						Login: "octodog",
					},
					State:     "APPROVED",
					UpdatedAt: parseRFC3339WithoutError("2020-02-23T12:00:00Z"),
				},
			},
			PageInfo: GQLPageInfo{
				EndCursor:   "",
				HasNextPage: false,
			},
		},
		UpdatedAt: parseRFC3339WithoutError("2020-02-23T12:00:00Z"),
	}
	activity := getPullRequestReviewStateArray(pullRequest)
	assert.Equal(t, []string{"COMMENTED", "CODE REVIEW", "APPROVED"}, activity)
}

func TestGetPullRequestReviewStateArrayNoReviews(t *testing.T)  {
	pullRequest := GQLPullRequest{
		Title:  "PR1",
		Number: 0,
		Author: GQLActor{
			Login: "octocat",
		},
		State: "MERGED",
		Reviews: GQLPullRequestReviewConnection{
			Nodes: []GQLPullRequestReview{
			},
			PageInfo: GQLPageInfo{
				EndCursor:   "",
				HasNextPage: false,
			},
		},
		UpdatedAt: parseRFC3339WithoutError("2020-02-23T12:00:00Z"),
	}
	activity := getPullRequestReviewStateArray(pullRequest)
	var emptyStrings []string
	assert.Equal(t, emptyStrings, activity)
}