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

func getGithubPlatformInformation(src string, since time.Time, until time.Time) ([]GQLIssue, []GQLPullRequest, error) {

	// Check if repository is on Github
	if !strings.Contains(src, "github.com") {
		return []GQLIssue{}, []GQLPullRequest{}, errors.New("repository is not on github")
	}

	repositoryOwner, repositoryName := getOwnerAndNameOfGithubUrl(src)

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
				fmt.Println(err)
			}
			var refetchResponse RequestGQLRepositoryInformation
			if err := json.Unmarshal(resp, &refetchResponse); err != nil {
				fmt.Println(err)
			}
			response.Data.Repository.Issues.Nodes = append(response.Data.Repository.Issues.Nodes, refetchResponse.Data.Repository.Issues.Nodes...)
			response.Data.Repository.Issues.PageInfo = refetchResponse.Data.Repository.Issues.PageInfo
		}
	}

	//Fetch the missing IssueComments

	issueCommentsAfter := ""
	var issueToRefech int

	for index := range response.Data.Repository.Issues.Nodes {
		issueToRefech = response.Data.Repository.Issues.Nodes[index].Number
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
				}`, repositoryOwner, repositoryName, issueToRefech, pageLength, issueCommentsAfter)
				resp, err := GClientWrapper.Query(specificIssueQuery)
				if err != nil {
					fmt.Println(err)
				}
				var refetchResponse RequestGQLRepositoryInformation
				if err := json.Unmarshal(resp, &refetchResponse); err != nil {
					fmt.Println(err)
				}
				response.Data.Repository.Issues.Nodes[index].Comments.Nodes = append(response.Data.Repository.Issues.Nodes[index].Comments.Nodes, refetchResponse.Data.Repository.Issue.Comments.Nodes...)
				response.Data.Repository.Issues.Nodes[index].Comments.PageInfo = refetchResponse.Data.Repository.Issue.Comments.PageInfo
			}
		}
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
				fmt.Println(err)
			}
			var refetchResponse RequestGQLRepositoryInformation
			if err := json.Unmarshal(resp, &refetchResponse); err != nil {
				fmt.Println(err)
			}
			response.Data.Repository.PullRequests.Nodes = append(response.Data.Repository.PullRequests.Nodes, refetchResponse.Data.Repository.PullRequests.Nodes...)
			response.Data.Repository.PullRequests.PageInfo = refetchResponse.Data.Repository.PullRequests.PageInfo
		}
	}

	//Fetch the missing PullRequestReviews

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
					fmt.Println(err)
				}
				var refetchResponse RequestGQLRepositoryInformation
				if err := json.Unmarshal(resp, &refetchResponse); err != nil {
					fmt.Println(err)
				}
				response.Data.Repository.PullRequests.Nodes[index].Reviews.Nodes = append(response.Data.Repository.PullRequests.Nodes[index].Reviews.Nodes, refetchResponse.Data.Repository.PullRequest.Reviews.Nodes...)
				response.Data.Repository.PullRequests.Nodes[index].Reviews.PageInfo = refetchResponse.Data.Repository.PullRequest.Reviews.PageInfo
			}
		}
	}

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

func (g *GithubClientWrapperClient) Query(query string) ([]byte, error) {
	jsonData := map[string]string{
		"query": query,
	}
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", g.GitHubURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "bearer d4598c799e5085885405e23e873606d5795e19c8")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(response.Body)
}

func getOwnerAndNameOfGithubUrl(src string) (string, string) {
	ownerAndName := strings.Split(strings.Split(src[0:len(src)-4], "github.com/")[1], "/")
	return ownerAndName[0], ownerAndName[1]
}

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

func getGithubPlatformInformationFromUser(src string, issues []GQLIssue, pullRequests []GQLPullRequest, userEmail string) (PlatformUserInformation, error) {

	repoOwner, repoName := getOwnerAndNameOfGithubUrl(src)

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

func getPullRequestReviewStateArray(pullRequest GQLPullRequest) []string {
	var reviews []string
	for i := range pullRequest.Reviews.Nodes {
		reviews = append(reviews, pullRequest.Reviews.Nodes[i].State)
	}
	return reviews
}
