package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type GitAnalyzationChannel struct {
	Result map[Contributor]Contribution
	Reason error
}

type PlatformInformationChannel struct {
	ResultIssues       []GQLIssue
	ResultPullRequests []GQLPullRequest
	Reason             error
}

func getAllContributions(w http.ResponseWriter, r *http.Request) {
	var err error

	// get the repository url from the request
	repositoryUrl, err := getRepositoryFromRequest(r)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusNotFound)
		return
	}

	// detect whether the platformInformation flag in the request was set
	analyzePlatformInformation := getShouldAnalyzePlatformInformation(r)

	commitsSince, commitsUntil, err := getTimeRange(r)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// make the channels for both go routines (analyze repo / platform information)
	gitAnalyzationChannel := make(chan GitAnalyzationChannel)
	platformInformationChannel := make(chan PlatformInformationChannel)

	// if we don't have to analyze the platform, close the channel again since we don't need it
	if !analyzePlatformInformation {
		close(platformInformationChannel)
	}

	// go routine to analyze the repository using git independently from main thread
	go func() {
		routineContributionMap, routineErr := analyzeRepository(repositoryUrl, commitsSince, commitsUntil)
		gitAnalyzationChannel <- GitAnalyzationChannel{
			Result: routineContributionMap,
			Reason: routineErr,
		}
		close(gitAnalyzationChannel)
	}()

	// execute go routine to fetch the platform information only when the platformInformation flag is set
	if analyzePlatformInformation {
		go func() {
			routineIssues, routinePullRequests, routineErr := getPlatformInformation(repositoryUrl, commitsSince, commitsUntil)
			platformInformationChannel <- PlatformInformationChannel{
				ResultIssues: routineIssues,
				ResultPullRequests: routinePullRequests,
				Reason: routineErr,
			}
			close(platformInformationChannel)
		}()
	}

	// set the openness of the to the default value
	chanel1Open := true
	chanel2Open := analyzePlatformInformation

	// initialize the return variables for the go routines
	var contributionMap map[Contributor]Contribution
	var issues []GQLIssue
	var pullRequests []GQLPullRequest

	// wait for the results of both go routines
	for chanel1Open || chanel2Open {
		select {
		case msg1, ok1 := <-gitAnalyzationChannel:
			if !ok1 {
				// if the channel is closed set the flag to false
				chanel1Open = false
			} else if msg1.Reason != nil {
				// error handling
				if strings.Contains(msg1.Reason.Error(), "authentication") {
					makeHttpStatusErr(w, msg1.Reason.Error(), http.StatusUnauthorized)
				} else {
					makeHttpStatusErr(w, msg1.Reason.Error(), http.StatusInternalServerError)
				}
				return
			} else {
				// save the return value to the initialized variable
				contributionMap = msg1.Result
			}
		case msg2, ok2 := <-platformInformationChannel:
			if !ok2 {
				// if the channel is closed set the flag to false
				chanel2Open = false
			} else if msg2.Reason != nil {
				// error handling
				makeHttpStatusErr(w, msg2.Reason.Error(), http.StatusInternalServerError)
				return
			} else {
				// save the return value to the initialized variable
				issues = msg2.ResultIssues
				pullRequests = msg2.ResultPullRequests
			}
		}
	}

	if analyzePlatformInformation {
		// if platform information is desired, filter out the platform information per git user and
		// return the platform information and the git contribution
		var contributions []ContributionWithPlatformInformation
		for k, v := range contributionMap {
			userInformation, err := getPlatformInformationFromUser(repositoryUrl, issues, pullRequests, k.Email)
			if err != nil {
				makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
				return
			}
			contributions = append(contributions, ContributionWithPlatformInformation{
				GitInformation:      v,
				PlatformInformation: userInformation,
			})
		}
		jsonErr := json.NewEncoder(w).Encode(contributions)
		if jsonErr != nil {
			fmt.Println("Could not encode to json", jsonErr)
		}
	} else {
		// if platform information is not desired convert the map into an array and return it
		var contributions []Contribution
		for _, v := range contributionMap {
			contributions = append(contributions, v)
		}
		jsonErr := json.NewEncoder(w).Encode(contributions)
		if jsonErr != nil {
			fmt.Println("Could not encode to json", jsonErr)
		}
	}
}

func getContributionWeights(w http.ResponseWriter, r *http.Request) {
	var err error

	// get the repository url from the request
	repositoryUrl, err := getRepositoryFromRequest(r)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusNotFound)
		return
	}

	// detect whether the platformInformation flag in the request was set
	analyzePlatformInformation := getShouldAnalyzePlatformInformation(r)

	commitsSince, commitsUntil, err := getTimeRange(r)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// make the channels for both go routines (analyze repo / platform information)
	gitAnalyzationChannel := make(chan GitAnalyzationChannel)
	platformInformationChannel := make(chan PlatformInformationChannel)

	// if we don't have to analyze the platform, close the channel again since we don't need it
	if !analyzePlatformInformation {
		close(platformInformationChannel)
	}

	// go routine to analyze the repository using git independently from main thread
	go func() {
		routineContributionMap, routineErr := analyzeRepository(repositoryUrl, commitsSince, commitsUntil)
		gitAnalyzationChannel <- GitAnalyzationChannel{
			Result: routineContributionMap,
			Reason: routineErr,
		}
		close(gitAnalyzationChannel)
	}()

	// execute go routine to fetch the platform information only when the platformInformation flag is set
	if analyzePlatformInformation {
		go func() {
			routineIssues, routinePullRequests, routineErr := getPlatformInformation(repositoryUrl, commitsSince, commitsUntil)
			platformInformationChannel <- PlatformInformationChannel{
				ResultIssues: routineIssues,
				ResultPullRequests: routinePullRequests,
				Reason: routineErr,
			}
			close(platformInformationChannel)
		}()
	}

	// set the openness of the to the default value
	chanel1Open := true
	chanel2Open := analyzePlatformInformation

	// initialize the return variables for the go routines
	var contributionMap map[Contributor]Contribution
	var issues []GQLIssue
	var pullRequests []GQLPullRequest

	// wait for the results of both go routines
	for chanel1Open || chanel2Open {
		select {
		case msg1, ok1 := <-gitAnalyzationChannel:
			if !ok1 {
				// if the channel is closed set the flag to false
				chanel1Open = false
			} else if msg1.Reason != nil {
				// error handling
				if strings.Contains(msg1.Reason.Error(), "authentication") {
					makeHttpStatusErr(w, msg1.Reason.Error(), http.StatusUnauthorized)
				} else {
					makeHttpStatusErr(w, msg1.Reason.Error(), http.StatusInternalServerError)
				}
				return
			} else {
				// save the return value to the initialized variable
				contributionMap = msg1.Result
			}
		case msg2, ok2 := <-platformInformationChannel:
			if !ok2 {
				// if the channel is closed set the flag to false
				chanel2Open = false
			} else if msg2.Reason != nil {
				// error handling
				makeHttpStatusErr(w, msg2.Reason.Error(), http.StatusInternalServerError)
				return
			} else {
				// save the return value to the initialized variable
				issues = msg2.ResultIssues
				pullRequests = msg2.ResultPullRequests
			}
		}
	}

	// only temporary since we don't include platform information in the weights yet
	fmt.Println(issues)
	fmt.Println(pullRequests)

	weightsMap, err := weightContributions(contributionMap)

	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var contributionWeights []FlatFeeWeight
	for _, v := range weightsMap {
		contributionWeights = append(contributionWeights, v)
	}
	jsonErr := json.NewEncoder(w).Encode(contributionWeights)
	if jsonErr != nil {
		fmt.Println("Could not encode to json", jsonErr)
	}
}

func getRepositoryFromRequest(r *http.Request) (string, error) {
	repositoryUrl := r.URL.Query()["repositoryUrl"]
	if len(repositoryUrl) < 1 {
		return "", errors.New("repository not found")
	}
	return repositoryUrl[0], nil
}

// returns the time range in the format since, until, error from the request with time in rfc3339 format
func getTimeRange(r *http.Request) (time.Time, time.Time, error) {
	var err error

	// convert since RFC3339 into golang time
	commitsSinceString := r.URL.Query()["since"]
	var commitsSince time.Time
	if len(commitsSinceString) > 0 {
		commitsSince, err = convertTimestampStringToTime(commitsSinceString[0])
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	// convert until RFC3339 into golang time
	commitsUntilString := r.URL.Query()["until"]
	var commitsUntil time.Time
	if len(commitsUntilString) > 0 {
		commitsUntil, err = convertTimestampStringToTime(commitsUntilString[0])
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	return commitsSince, commitsUntil, nil
}

func makeHttpStatusErr(w http.ResponseWriter, errString string, httpStatusError int) {
	w.WriteHeader(http.StatusInternalServerError)
	_, fmtErr := fmt.Fprintf(w, errString)
	if fmtErr != nil {
		fmt.Println("Could not format", fmtErr)
	}
}

func getShouldAnalyzePlatformInformation(r *http.Request) bool {
	platformInformationUrlParam := r.URL.Query()["platformInformation"]
	return len(platformInformationUrlParam) > 0 && platformInformationUrlParam[0] == "true"
}

func convertTimestampStringToTime(rfc3339time string) (time.Time, error) {
	commitsSinceTime, err := time.Parse(time.RFC3339, rfc3339time)
	if err != nil {
		return time.Time{}, err
	}
	return commitsSinceTime, nil
}
