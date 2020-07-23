package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getPlatformInformation() {
	query :=
		`{
			repository(owner:"go-git", name:"go-git") {
				issues(last:100, filterBy: {since: "2020-06-01T00:00:00+0000"}) {
					edges {
						cursor
						node {
							title
							number
							author {
								login
							}
							comments(first:100) {
								totalCount
								edges {
									cursor
									node {
										author {
											login
										}
									}
								}
							}
						}
					}
				}
			}
		}`

	resp, err := manualGQL(query)
	if err != nil {
		fmt.Println(err)
	}
	var response RequestGQLRepositoryInformation
	if err := json.Unmarshal(resp, &response); err != nil {
		fmt.Println(err)
	}
	fmt.Println(response.Data.Repository)
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
