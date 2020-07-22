package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

func getPlatformInformation(src string, since time.Time, until time.Time) {

	if !strings.Contains(src, "github.com") {
		fmt.Println("Repository is not on github")
		return
	}

	repositoryOwner, repositoryName := getOwnerAndNameOfGithubUrl(src)

	//var repository GQLIssueConnection
	repositoryIssues, err := getRepositoryIssues(repositoryOwner, repositoryName, since, until)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(len(repositoryIssues.Nodes))
}

func getRepositoryIssues(repositoryOwner string, repositoryName string, since time.Time, until time.Time) (GQLIssueConnection, error) {
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

	resp, err := manualGQL(query)
	if err != nil {
		return GQLIssueConnection{}, err
	}
	var response RequestGQLRepositoryInformation
	if err := json.Unmarshal(resp, &response); err != nil {
		return GQLIssueConnection{}, err
	}

	// Pagination

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
			resp, err := manualGQL(issueRefetchQuery)
			if err != nil {
				fmt.Println(err)
			}
			var refetchResponse RequestGQLRepositoryInformation
			if err := json.Unmarshal(resp, &refetchResponse); err != nil {
				fmt.Println(err)
			}
			response.Data.Repository.Issues.Nodes = append(response.Data.Repository.Issues.Nodes, refetchResponse.Data.Repository.Issues.Nodes...)
			response.Data.Repository.Issues.PageInfo = refetchResponse.Data.Repository.Issues.PageInfo
			fmt.Println(len(response.Data.Repository.Issues.Nodes))
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
				resp, err := manualGQL(specificIssueQuery)
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

	// Filtering
	response.Data.Repository.Issues.Nodes = filterIssuesByDate(response.Data.Repository.Issues.Nodes, since, until)
	for index := range response.Data.Repository.Issues.Nodes {
		response.Data.Repository.Issues.Nodes[index].Comments.Nodes = filterIssueCommentsByDate(response.Data.Repository.Issues.Nodes[index].Comments.Nodes, since, until)
	}

	return response.Data.Repository.Issues, nil
}

func manualGQL(query string) ([]byte, error) {
	jsonData := map[string]string{
		"query": query,
	}
	fmt.Println("Making request with this query", query)
	jsonValue, _ := json.Marshal(jsonData)
	request, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(jsonValue))
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
