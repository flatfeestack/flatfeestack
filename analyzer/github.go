package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var GClientWrapper GithubClientWrapper

type GithubClientWrapper interface {
	Query(query string) ([]byte, error)
}

type GithubClientWrapperClient struct {
	GitHubURL string
}

// getGithubPlatformInformation collects ALL platform information (issues & pull requests) from github
func getGithubPlatformInformation(src string, since time.Time, until time.Time) ([]GQLIssue, []GQLPullRequest, error) {

	// Check if repository is on Github
	if !strings.Contains(src, "github.com") {
		return []GQLIssue{}, []GQLPullRequest{}, errors.New("repository is not on github")
	}

	repositoryOwner, repositoryName, err := getOwnerAndNameOfGithubUrl(src)
	if err != nil {
		return nil, nil, err
	}

	//var repository GQLIssueConnection
	repositoryIssues, err := getGithubRepositoryIssues(repositoryOwner, repositoryName, since, until)
	if err != nil {
		fmt.Println(err)
	}

	//var repository GQLIssueConnection
	repositoryPullRequests, err := getGithubRepositoryPullRequests(repositoryOwner, repositoryName, since, until)
	if err != nil {
		fmt.Println(err)
	}

	return repositoryIssues, repositoryPullRequests, nil
}

// getGithubRepositoryIssues fetches ALL issues from github
func getGithubRepositoryIssues(repositoryOwner string, repositoryName string, since time.Time, until time.Time) ([]GQLIssue, error) {
	var timeZeroValue time.Time

	sinceFilterBy := ""
	pageLength := 100

	if since != timeZeroValue {
		sinceFilterBy = `since: "` + since.Format(time.RFC3339) + `"`
	}

	query := fmt.Sprintf(
		`{
			repository(owner:"%s", name:"%s") {
				issues(first:%d, filterBy: {%s}) {
					pageInfo {
						endCursor
						hasNextPage
					}
					nodes {
						title
						number
						author {
							login
						}
						comments(first: %d) {
							pageInfo {
								endCursor
								hasNextPage
							}						
							nodes {
								author {
									login
								}
								updatedAt
							}
						}
						updatedAt
					}
				}
			}
		}`, repositoryOwner, repositoryName, pageLength, sinceFilterBy, pageLength)

	resp, err := GClientWrapper.Query(query)
	if err != nil {
		return []GQLIssue{}, err
	}
	var response RequestGQLRepositoryInformation
	if err := json.Unmarshal(resp, &response); err != nil {
		return []GQLIssue{}, err
	}

	// ----------
	// Pagination
	// ----------

	// Fetch the missing issues

	response, err = fetchPaginationIssues(response, repositoryOwner, repositoryName, pageLength, sinceFilterBy)
	if err != nil {
		return []GQLIssue{}, err
	}
	//Fetch the missing IssueComments

	response, err = fetchPaginationIssueComments(response, repositoryOwner, repositoryName, pageLength)
	if err != nil {
		return []GQLIssue{}, err
	}

	// ---------
	// Filtering
	// ---------

	// Filtering because there is no until filter in gql
	response.Data.Repository.Issues.Nodes = filterIssuesByDate(response.Data.Repository.Issues.Nodes, since, until)
	for index := range response.Data.Repository.Issues.Nodes {
		response.Data.Repository.Issues.Nodes[index].Comments.Nodes = filterIssueCommentsByDate(response.Data.Repository.Issues.Nodes[index].Comments.Nodes, since, until)
	}

	return response.Data.Repository.Issues.Nodes, nil
}

// fetchPaginationIssues fetch missing issues repeatedly until github indicates no further pages
func fetchPaginationIssues(response RequestGQLRepositoryInformation, repositoryOwner string, repositoryName string, pageLength int, sinceFilterBy string) (RequestGQLRepositoryInformation, error) {
	issuesAfter := ""

	for ok0 := true; ok0; ok0 = response.Data.Repository.Issues.PageInfo.HasNextPage {
		if response.Data.Repository.Issues.PageInfo.HasNextPage {
			issuesAfter = response.Data.Repository.Issues.PageInfo.EndCursor
			issueRefetchQuery := fmt.Sprintf(
				`{
				repository(owner:"%s", name:"%s") {
					issues(first:%d, filterBy: {%s}, after: "%s") {
						pageInfo {
							endCursor
							hasNextPage
						}
						nodes {
							title
							number
							author {
								login
							}
							comments(first: %d) {
								pageInfo {
									endCursor
									hasNextPage
								}						
								nodes {
									author {
										login
									}
									updatedAt
								}
							}
							updatedAt
						}
					}
				}
			}`, repositoryOwner, repositoryName, pageLength, sinceFilterBy, issuesAfter, pageLength)
			resp, err := GClientWrapper.Query(issueRefetchQuery)
			if err != nil {
				return response, err
			}
			var refetchResponse RequestGQLRepositoryInformation
			if err := json.Unmarshal(resp, &refetchResponse); err != nil {
				return response, err
			}
			response.Data.Repository.Issues.Nodes = append(response.Data.Repository.Issues.Nodes, refetchResponse.Data.Repository.Issues.Nodes...)
			response.Data.Repository.Issues.PageInfo = refetchResponse.Data.Repository.Issues.PageInfo
		}
	}

	return response, nil
}

// fetchPaginationIssues fetch missing issue comments repeatedly until github indicates no further pages
func fetchPaginationIssueComments(response RequestGQLRepositoryInformation, repositoryOwner string, repositoryName string, pageLength int) (RequestGQLRepositoryInformation, error) {
	issueCommentsAfter := ""
	var issueToRefetch int

	for index := range response.Data.Repository.Issues.Nodes {
		issueToRefetch = response.Data.Repository.Issues.Nodes[index].Number
		for ok1 := true; ok1; ok1 = response.Data.Repository.Issues.Nodes[index].Comments.PageInfo.HasNextPage {
			if response.Data.Repository.Issues.Nodes[index].Comments.PageInfo.HasNextPage {
				issueCommentsAfter = response.Data.Repository.Issues.Nodes[index].Comments.PageInfo.EndCursor
				specificIssueQuery := fmt.Sprintf(
					`{
					repository(owner:"%s", name:"%s") {
						issue(number: %d) {
							title
							number
							author {
								login
							}
							comments(first: %d, after: "%s") {
								pageInfo {
									endCursor
									hasNextPage
								}
								nodes {
									author {
										login
									}
									updatedAt
								}
							}
							updatedAt
						}
					}
				}`, repositoryOwner, repositoryName, issueToRefetch, pageLength, issueCommentsAfter)
				resp, err := GClientWrapper.Query(specificIssueQuery)
				if err != nil {
					return response, nil
				}
				var refetchResponse RequestGQLRepositoryInformation
				if err := json.Unmarshal(resp, &refetchResponse); err != nil {
					return response, nil
				}
				response.Data.Repository.Issues.Nodes[index].Comments.Nodes = append(response.Data.Repository.Issues.Nodes[index].Comments.Nodes, refetchResponse.Data.Repository.Issue.Comments.Nodes...)
				response.Data.Repository.Issues.Nodes[index].Comments.PageInfo = refetchResponse.Data.Repository.Issue.Comments.PageInfo
			}
		}
	}

	return response, nil
}

// getGithubRepositoryPullRequests fetches ALL pull requests from github
func getGithubRepositoryPullRequests(repositoryOwner string, repositoryName string, since time.Time, until time.Time) ([]GQLPullRequest, error) {
	pageLength := 100

	query := fmt.Sprintf(
		`{
			repository(owner:"%s", name:"%s") {
				pullRequests(first:%d) {
					pageInfo {
						endCursor
						hasNextPage
					}
					nodes {
						title
						number
						author {
							login
						}
						state
						reviews(first: %d) {
							pageInfo {
								endCursor
								hasNextPage
							}						
							nodes {
								author {
									login
								}
								state
								updatedAt
							}
						}
						updatedAt
					}
				}
			}
		}`, repositoryOwner, repositoryName, pageLength, pageLength)

	resp, err := GClientWrapper.Query(query)
	if err != nil {
		return []GQLPullRequest{}, err
	}
	var response RequestGQLRepositoryInformation
	if err := json.Unmarshal(resp, &response); err != nil {
		return []GQLPullRequest{}, err
	}

	// ----------
	// Pagination
	// ----------

	// Fetch the missing pullRequests

	response, err = fetchPaginationPullRequests(response, repositoryOwner, repositoryName, pageLength)

	//Fetch the missing PullRequestReviews

	response, err = fetchPaginationPullRequestReviews(response, repositoryOwner, repositoryName, pageLength)

	// ---------
	// Filtering
	// ---------

	// Filtering because there is no until filter in gql
	response.Data.Repository.PullRequests.Nodes = filterPullRequestsByDate(response.Data.Repository.PullRequests.Nodes, since, until)
	for index := range response.Data.Repository.PullRequests.Nodes {
		response.Data.Repository.PullRequests.Nodes[index].Reviews.Nodes = filterPullRequestReviewsByDate(response.Data.Repository.PullRequests.Nodes[index].Reviews.Nodes, since, until)
	}

	return response.Data.Repository.PullRequests.Nodes, nil
}

// fetchPaginationPullRequests fetch missing pull requests repeatedly until github indicates no further pages
func fetchPaginationPullRequests(response RequestGQLRepositoryInformation, repositoryOwner string, repositoryName string, pageLength int) (RequestGQLRepositoryInformation, error) {
	pullRequestsAfter := ""

	for ok0 := true; ok0; ok0 = response.Data.Repository.PullRequests.PageInfo.HasNextPage {
		if response.Data.Repository.PullRequests.PageInfo.HasNextPage {
			pullRequestsAfter = response.Data.Repository.PullRequests.PageInfo.EndCursor
			pullRequestRefetchQuery := fmt.Sprintf(
				`{
				repository(owner:"%s", name:"%s") {
					pullRequests(first:%d, after: "%s") {
						pageInfo {
							endCursor
							hasNextPage
						}
						nodes {
							title
							number
							author {
								login
							}
							state
							reviews(first: %d) {
								pageInfo {
									endCursor
									hasNextPage
								}						
								nodes {
									author {
										login
									}
									state
									updatedAt
								}
							}
							updatedAt
						}
					}
				}
			}`, repositoryOwner, repositoryName, pageLength, pullRequestsAfter, pageLength)
			resp, err := GClientWrapper.Query(pullRequestRefetchQuery)
			if err != nil {
				return response, err
			}
			var refetchResponse RequestGQLRepositoryInformation
			if err := json.Unmarshal(resp, &refetchResponse); err != nil {
				return response, err
			}
			response.Data.Repository.PullRequests.Nodes = append(response.Data.Repository.PullRequests.Nodes, refetchResponse.Data.Repository.PullRequests.Nodes...)
			response.Data.Repository.PullRequests.PageInfo = refetchResponse.Data.Repository.PullRequests.PageInfo
		}
	}
	return response, nil
}

// fetchPaginationIssues fetch missing pull request reviews repeatedly until github indicates no further pages
func fetchPaginationPullRequestReviews(response RequestGQLRepositoryInformation, repositoryOwner string, repositoryName string, pageLength int) (RequestGQLRepositoryInformation, error) {
	pullRequestReviewsAfter := ""
	var pullRequestToRefech int

	for index := range response.Data.Repository.PullRequests.Nodes {
		pullRequestToRefech = response.Data.Repository.PullRequests.Nodes[index].Number
		for ok1 := true; ok1; ok1 = response.Data.Repository.PullRequests.Nodes[index].Reviews.PageInfo.HasNextPage {
			if response.Data.Repository.PullRequests.Nodes[index].Reviews.PageInfo.HasNextPage {
				pullRequestReviewsAfter = response.Data.Repository.PullRequests.Nodes[index].Reviews.PageInfo.EndCursor
				specificPullRequestQuery := fmt.Sprintf(
					`{
					repository(owner:"%s", name:"%s") {
						pullRequest(number: %d) {
							title
							number
							author {
								login
							}
							state
							reviews(first: %d, after: "%s") {
								pageInfo {
									endCursor
									hasNextPage
								}
								nodes {
									author {
										login
									}
									state
									updatedAt
								}
							}
							updatedAt
						}
					}
				}`, repositoryOwner, repositoryName, pullRequestToRefech, pageLength, pullRequestReviewsAfter)
				resp, err := GClientWrapper.Query(specificPullRequestQuery)
				if err != nil {
					return response, err
				}
				var refetchResponse RequestGQLRepositoryInformation
				if err := json.Unmarshal(resp, &refetchResponse); err != nil {
					return response, err
				}
				response.Data.Repository.PullRequests.Nodes[index].Reviews.Nodes = append(response.Data.Repository.PullRequests.Nodes[index].Reviews.Nodes, refetchResponse.Data.Repository.PullRequest.Reviews.Nodes...)
				response.Data.Repository.PullRequests.Nodes[index].Reviews.PageInfo = refetchResponse.Data.Repository.PullRequest.Reviews.PageInfo
			}
		}
	}
	return response, nil
}

// Query abstracts the requests to the github graphql api
func (g *GithubClientWrapperClient) Query(query string) ([]byte, error) {
	jsonData := map[string]string{
		"query": query,
	}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", g.GitHubURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "bearer 534321b71c6cd86f0d87a26445f982954c0a8594")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(response.Body)
}

// getOwnerAndNameOfGithubUrl parses the src string to extract the owner and name of the repository
func getOwnerAndNameOfGithubUrl(src string) (string, string, error) {
	partAfterDomain := strings.Split(src[0:len(src)-4], "github.com/")
	if len(partAfterDomain) < 2 {
		return "", "", errors.New("incorrect github repository url")
	}
	ownerAndName := strings.Split(partAfterDomain[1], "/")
	if len(ownerAndName) < 2 {
		return "", "", errors.New("incorrect github repository url")
	}
	return ownerAndName[0], ownerAndName[1], nil
}

// filterIssueCommentsByDate filters the issue comments so that only they between since and until are kept
func filterIssueCommentsByDate(comments []GQLIssueComment, since time.Time, until time.Time) []GQLIssueComment {
	var timeZeroValue time.Time
	var filteredIssueComments []GQLIssueComment
	for index := range comments {
		if (comments[index].UpdatedAt.After(since) || since == timeZeroValue) && (comments[index].UpdatedAt.Before(until) || until == timeZeroValue) {
			filteredIssueComments = append(filteredIssueComments, comments[index])
		}
	}
	return filteredIssueComments
}

// filterIssuesByDate filters the issues so that only they between since and until are kept
func filterIssuesByDate(issueEdges []GQLIssue, since time.Time, until time.Time) []GQLIssue {
	var timeZeroValue time.Time
	var filteredIssues []GQLIssue
	for index := range issueEdges {
		if (issueEdges[index].UpdatedAt.After(since) || since == timeZeroValue) && (issueEdges[index].UpdatedAt.Before(until) || until == timeZeroValue) {
			filteredIssues = append(filteredIssues, issueEdges[index])
		}
	}
	return filteredIssues
}

// filterPullRequestsByDate filters the pull requests so that only they between since and until are kept
func filterPullRequestsByDate(pullRequestEdges []GQLPullRequest, since time.Time, until time.Time) []GQLPullRequest {
	var timeZeroValue time.Time
	var filteredPullRequests []GQLPullRequest
	for index := range pullRequestEdges {
		if (pullRequestEdges[index].UpdatedAt.After(since) || since == timeZeroValue) && (pullRequestEdges[index].UpdatedAt.Before(until) || until == timeZeroValue) {
			filteredPullRequests = append(filteredPullRequests, pullRequestEdges[index])
		}
	}
	return filteredPullRequests
}

// filterPullRequestReviewsByDate filters the pull requests reviews so that only they between since and until are kept
func filterPullRequestReviewsByDate(reviews []GQLPullRequestReview, since time.Time, until time.Time) []GQLPullRequestReview {
	var timeZeroValue time.Time
	var filteredPullRequestReviews []GQLPullRequestReview
	for index := range reviews {
		if (reviews[index].UpdatedAt.After(since) || since == timeZeroValue) && (reviews[index].UpdatedAt.Before(until) || until == timeZeroValue) {
			filteredPullRequestReviews = append(filteredPullRequestReviews, reviews[index])
		}
	}
	return filteredPullRequestReviews
}

// getGithubUsernameFromGitEmail check whether there exists a github user assigned to a commit that was created by a git user with this email
func getGithubUsernameFromGitEmail(repositoryOwner string, repositoryName string, email string) (string, error) {
	query := fmt.Sprintf(
		`{
			repository(owner:"%s", name:"%s") {
				ref(qualifiedName: "master") {
				  	target {
						... on Commit {
						  	history(first: 1, author: {emails: "%s"} ) {
								nodes {
									author {
										name
										email
										date
										user {
											login
										}
									}
								}
						  	}
						}
					}
				}
			}
      	}`, repositoryOwner, repositoryName, email)

	resp, err := GClientWrapper.Query(query)
	if err != nil {
		return "", err
	}
	var response RequestGQLRepositoryInformation
	if err := json.Unmarshal(resp, &response); err != nil {
		return "", err
	}
	if len(response.Data.Repository.Ref.Target.History.Nodes) < 1 {
		return "", errors.New("could not find user")
	}
	return response.Data.Repository.Ref.Target.History.Nodes[0].Author.User.Login, nil
}

// getGithubPlatformInformationFromUser extract the platform information for a specific user from the collected platform metrics
func getGithubPlatformInformationFromUser(src string, issues []GQLIssue, pullRequests []GQLPullRequest, userEmail string) (PlatformUserInformation, error) {

	repoOwner, repoName, err := getOwnerAndNameOfGithubUrl(src)
	if err != nil {
		return PlatformUserInformation{}, err
	}

	login, err := getGithubUsernameFromGitEmail(repoOwner, repoName, userEmail)

	if err != nil {
		return PlatformUserInformation{}, err
	}

	// list issues where user is author, where the values are the amount of comments inside that issue
	var issuesWhereUserIsAuthor []int

	// amount of comments user wrote in issues
	amountOfIssueCommenter := 0

	for i := 0; i < len(issues); i++ {
		if issues[i].Author.Login == login {
			issuesWhereUserIsAuthor = append(issuesWhereUserIsAuthor, len(issues[i].Comments.Nodes))
		}
		for j := 0; j < len(issues[i].Comments.Nodes); j++ {
			if issues[i].Comments.Nodes[j].Author.Login == login {
				amountOfIssueCommenter++
			}
		}
	}

	// list issues where user is author, where the values are the amount of comments inside that issue
	var pullRequestsWhereUserIsAuthor []PullRequestInformation

	// amount of comments user wrote in issues
	amountOfPullRequestReviewer := 0

	for i := 0; i < len(pullRequests); i++ {
		if pullRequests[i].Author.Login == login {
			pullRequestsWhereUserIsAuthor = append(pullRequestsWhereUserIsAuthor, PullRequestInformation{
				State:   pullRequests[i].State,
				Reviews: getPullRequestReviewStateArray(pullRequests[i]),
			})
		}
		for j := 0; j < len(pullRequests[i].Reviews.Nodes); j++ {
			if pullRequests[i].Reviews.Nodes[j].Author.Login == login {
				amountOfPullRequestReviewer++
			}
		}
	}

	return PlatformUserInformation{
		UserName: login,
		PullRequestInformation: PullRequestUserInformation{
			Author:   pullRequestsWhereUserIsAuthor,
			Reviewer: amountOfPullRequestReviewer,
		},
		IssueInformation: IssueUserInformation{
			Author:    issuesWhereUserIsAuthor,
			Commenter: amountOfIssueCommenter,
		},
	}, nil
}

// getPullRequestReviewStateArray creates an array containing all the events happened in the reviewing process on a pull request
func getPullRequestReviewStateArray(pullRequest GQLPullRequest) []string {
	var reviews []string
	for i := range pullRequest.Reviews.Nodes {
		reviews = append(reviews, pullRequest.Reviews.Nodes[i].State)
	}
	return reviews
}
