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

	var timeZeroValue time.Time

	sinceFilterBy := ""

	if since != timeZeroValue {
		sinceFilterBy = `since: "` + since.Format(time.RFC3339) + `"`
	}

	query := fmt.Sprintf(
		`{
			repository(owner:"%s", name:"%s") {
				issues(last:3, filterBy: {%s}) {
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
						comments(first:100) {
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
		}`, repositoryOwner, repositoryName, sinceFilterBy)

	resp, err := manualGQL(query)
	if err != nil {
		fmt.Println(err)
	}
	var response RequestGQLRepositoryInformation
	if err := json.Unmarshal(resp, &response); err != nil {
		fmt.Println(err)
	}

	// Filtering
	response.Data.Repository.Issues.Nodes = filterIssuesByDate(response.Data.Repository.Issues.Nodes, since, until)
	for index := range response.Data.Repository.Issues.Nodes {
		response.Data.Repository.Issues.Nodes[index].Comments.Nodes = filterIssueCommentsByDate(response.Data.Repository.Issues.Nodes[index].Comments.Nodes, since, until)
	}


	fmt.Println(response.Data.Repository.Issues.Nodes)
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
