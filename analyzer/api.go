package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
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

// getAllContributions controls the whole process of the /contribution endpoint
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

	// detect branch to analyze
	branch := getBranchToAnalyze(r)

	commitsSince, commitsUntil, err := getTimeRange(r)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("\n\n---> contribution analysis request for repository %s on branch %s \n", repositoryUrl, branch)

	// make the channels for both go routines (analyze repo / platform information)
	gitAnalyzationChannel := make(chan GitAnalyzationChannel)
	platformInformationChannel := make(chan PlatformInformationChannel)

	// if we don't have to analyze the platform, close the channel again since we don't need it
	if !analyzePlatformInformation {
		close(platformInformationChannel)
	}

	// go routine to analyze the repository using git independently from main thread
	go func() {
		routineContributionMap, routineErr := analyzeRepositoryFromString(repositoryUrl, commitsSince, commitsUntil, branch)
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
				ResultIssues:       routineIssues,
				ResultPullRequests: routinePullRequests,
				Reason:             routineErr,
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

		platformInformationMappingStart := time.Now()
		var contributions []ContributionWithPlatformInformation
		takenUsernames := make(map[string]string)
		for k, v := range contributionMap {
			userInformation, err := getPlatformInformationFromUser(repositoryUrl, issues, pullRequests, k.Email)
			if err != nil {
				fmt.Printf("COULD_NOT_GET_PLATFORMINFORMATION_FROM_USER: %s; %s\n", k.Email, err.Error())
				contributions = append(contributions, ContributionWithPlatformInformation{
					GitInformation: v,
				})
				continue
			}
			if _, found := takenUsernames[userInformation.UserName]; !found {
				takenUsernames[userInformation.UserName] = userInformation.UserName
				contributions = append(contributions, ContributionWithPlatformInformation{
					GitInformation:      v,
					PlatformInformation: userInformation,
				})
			} else {
				contributions = append(contributions, ContributionWithPlatformInformation{
					GitInformation: v,
				})
			}
		}
		platformInformationMappingEnd := time.Now()
		fmt.Printf("---> platform information mapping in %dms\n", platformInformationMappingEnd.Sub(platformInformationMappingStart).Milliseconds())

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

// getContributionWeights controls the whole process of the /weights endpoint
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

	// detect branch to analyze
	branch := getBranchToAnalyze(r)

	commitsSince, commitsUntil, err := getTimeRange(r)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Printf("\n\n---> weight contributions request for repository %s on branch %s \n", repositoryUrl, branch)

	// make the channels for both go routines (analyze repo / platform information)
	gitAnalyzationChannel := make(chan GitAnalyzationChannel)
	platformInformationChannel := make(chan PlatformInformationChannel)

	// if we don't have to analyze the platform, close the channel again since we don't need it
	if !analyzePlatformInformation {
		close(platformInformationChannel)
	}

	// go routine to analyze the repository using git independently from main thread
	go func() {
		routineContributionMap, routineErr := analyzeRepositoryFromString(repositoryUrl, commitsSince, commitsUntil, branch)
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
				ResultIssues:       routineIssues,
				ResultPullRequests: routinePullRequests,
				Reason:             routineErr,
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

		platformInformationMappingStart := time.Now()
		contributionWithPlatformInformation := make(map[Contributor]ContributionWithPlatformInformation)
		takenUsernames := make(map[string]string)
		for k, v := range contributionMap {
			userInformation, err := getPlatformInformationFromUser(repositoryUrl, issues, pullRequests, k.Email)
			if err != nil {
				fmt.Printf("COULD_NOT_GET_PLATFORMINFORMATION_FROM_USER: %s; %s\n", k.Email, err.Error())
				contributionWithPlatformInformation[v.Contributor] = ContributionWithPlatformInformation{
					GitInformation: v,
				}
				continue
			}
			if _, found := takenUsernames[userInformation.UserName]; !found {
				takenUsernames[userInformation.UserName] = userInformation.UserName
				contributionWithPlatformInformation[v.Contributor] = ContributionWithPlatformInformation{
					GitInformation:      v,
					PlatformInformation: userInformation,
				}
			} else {
				contributionWithPlatformInformation[v.Contributor] = ContributionWithPlatformInformation{
					GitInformation: v,
				}
			}
		}
		platformInformationMappingEnd := time.Now()
		fmt.Printf("---> platform information mapping in %dms\n", platformInformationMappingEnd.Sub(platformInformationMappingStart).Milliseconds())

		weightsMap, err := weightContributionsWithPlatformInformation(contributionWithPlatformInformation)

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
	} else {
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
}

func analyzeRepository(w http.ResponseWriter, r *http.Request) {
	var request WebhookRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusBadRequest)
	}

	fmt.Println(request)

	if len(request.RepositoryUrl) == 0 {
		makeHttpStatusErr(w, "no required repository_url provided", http.StatusBadRequest)
	}

	var branch string
	if len(request.Branch) > 0 {
		branch = request.Branch
	} else {
		branch = os.Getenv("GO_GIT_DEFAULT_BRANCH")
	}

	var commitsSince time.Time
	if len(request.Since) > 0 {
		commitsSince, err = convertTimestampStringToTime(request.Since)
		if err != nil {
			makeHttpStatusErr(w, err.Error(), http.StatusBadRequest)
		}
	}

	var commitsUntil time.Time
	if len(request.Until) > 0 {
		commitsUntil, err = convertTimestampStringToTime(request.Until)
		if err != nil {
			makeHttpStatusErr(w, err.Error(), http.StatusBadRequest)
		}
	}

	requestId := uuid.New().String()

	err = json.NewEncoder(w).Encode(WebhookResponse{RequestId: requestId})
	if err != nil {
		makeHttpStatusErr(w, err.Error(), http.StatusInternalServerError)

	}

	go analyzeForWebhookInBackground(requestId, request, branch, commitsSince, commitsUntil)
	fmt.Println("is analyzing")

}

func analyzeForWebhookInBackground(requestId string, request WebhookRequest, branch string, commitsSince time.Time, commitsUntil time.Time) {
	fmt.Printf("\n\n---> webhook request for repository %s on branch %s \n", request.RepositoryUrl, branch)
	fmt.Printf("Request id: %s\n", requestId)

	// make the channels for both go routines (analyze repo / platform information)
	gitAnalyzationChannel := make(chan GitAnalyzationChannel)
	platformInformationChannel := make(chan PlatformInformationChannel)

	// if we don't have to analyze the platform, close the channel again since we don't need it
	if !request.PlatformInformation {
		close(platformInformationChannel)
	}

	// go routine to analyze the repository using git independently from main thread
	go func() {
		routineContributionMap, routineErr := analyzeRepositoryFromString(request.RepositoryUrl, commitsSince, commitsUntil, branch)
		gitAnalyzationChannel <- GitAnalyzationChannel{
			Result: routineContributionMap,
			Reason: routineErr,
		}
		close(gitAnalyzationChannel)
	}()

	// execute go routine to fetch the platform information only when the platformInformation flag is set
	if request.PlatformInformation {
		go func() {
			routineIssues, routinePullRequests, routineErr := getPlatformInformation(request.RepositoryUrl, commitsSince, commitsUntil)
			platformInformationChannel <- PlatformInformationChannel{
				ResultIssues:       routineIssues,
				ResultPullRequests: routinePullRequests,
				Reason:             routineErr,
			}
			close(platformInformationChannel)
		}()
	}

	// set the openness of the to the default value
	chanel1Open := true
	chanel2Open := request.PlatformInformation

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
				callbackToWebhook(WebhookCallback{
					RequestId: requestId,
					Success:   false,
					Error:     msg1.Reason.Error(),
				})
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
				callbackToWebhook(WebhookCallback{
					RequestId: requestId,
					Success:   false,
					Error:     msg2.Reason.Error(),
				})
				return
			} else {
				// save the return value to the initialized variable
				issues = msg2.ResultIssues
				pullRequests = msg2.ResultPullRequests
			}
		}
	}

	if request.PlatformInformation {
		// if platform information is desired, filter out the platform information per git user and
		// return the platform information and the git contribution

		platformInformationMappingStart := time.Now()
		contributionWithPlatformInformation := make(map[Contributor]ContributionWithPlatformInformation)
		takenUsernames := make(map[string]string)
		for k, v := range contributionMap {
			userInformation, err := getPlatformInformationFromUser(request.RepositoryUrl, issues, pullRequests, k.Email)
			if err != nil {
				fmt.Printf("COULD_NOT_GET_PLATFORMINFORMATION_FROM_USER: %s; %s\n", k.Email, err.Error())
				contributionWithPlatformInformation[v.Contributor] = ContributionWithPlatformInformation{
					GitInformation: v,
				}
				continue
			}
			if _, found := takenUsernames[userInformation.UserName]; !found {
				takenUsernames[userInformation.UserName] = userInformation.UserName
				contributionWithPlatformInformation[v.Contributor] = ContributionWithPlatformInformation{
					GitInformation:      v,
					PlatformInformation: userInformation,
				}
			} else {
				contributionWithPlatformInformation[v.Contributor] = ContributionWithPlatformInformation{
					GitInformation: v,
				}
			}
		}
		platformInformationMappingEnd := time.Now()
		fmt.Printf("---> platform information mapping in %dms\n", platformInformationMappingEnd.Sub(platformInformationMappingStart).Milliseconds())

		weightsMap, err := weightContributionsWithPlatformInformation(contributionWithPlatformInformation)

		if err != nil {
			callbackToWebhook(WebhookCallback{
				RequestId: requestId,
				Success:   false,
				Error:     err.Error(),
			})
			return
		}

		var contributionWeights []FlatFeeWeight
		for _, v := range weightsMap {
			contributionWeights = append(contributionWeights, v)
		}

		callbackToWebhook(WebhookCallback{
			RequestId: requestId,
			Success:   true,
			Result:    contributionWeights,
		})
	} else {
		weightsMap, err := weightContributions(contributionMap)

		if err != nil {
			callbackToWebhook(
				WebhookCallback{
					RequestId: requestId,
					Success:   false,
					Error:     err.Error(),
				})
			return
		}

		var contributionWeights []FlatFeeWeight
		for _, v := range weightsMap {
			contributionWeights = append(contributionWeights, v)
		}
		callbackToWebhook(WebhookCallback{
			RequestId: requestId,
			Success:   true,
			Result:    contributionWeights,
		})
	}
	fmt.Printf("Finished request %s\n", requestId)
}

// getRepositoryFromRequest extracts the repository from the route parameters
func getRepositoryFromRequest(r *http.Request) (string, error) {
	repositoryUrl := r.URL.Query()["repositoryUrl"]
	if len(repositoryUrl) < 1 {
		return "", errors.New("repository not found")
	}
	return repositoryUrl[0], nil
}

// getTimeRange returns the time range in the format since, until, error from the request with time in rfc3339 format
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

// makeHttpStatusErr writes an http status error with a specific message
func makeHttpStatusErr(w http.ResponseWriter, errString string, httpStatusError int) {
	w.WriteHeader(httpStatusError)
	_, fmtErr := fmt.Fprintf(w, errString)
	if fmtErr != nil {
		fmt.Println("Could not format", fmtErr)
	}
}

func callbackToWebhook(body WebhookCallback) {
	reqBody, _ := json.Marshal(body)
	_, _ = http.Post(os.Getenv("WEBHOOK_CALLBACK_URL"), "application/json", bytes.NewBuffer(reqBody))
}

// getShouldAnalyzePlatformInformation extracts whether platform information should be considered from the route parameters
func getShouldAnalyzePlatformInformation(r *http.Request) bool {
	platformInformationUrlParam := r.URL.Query()["platformInformation"]
	return len(platformInformationUrlParam) > 0 && platformInformationUrlParam[0] == "true"
}

// getBranchToAnalyze extracts from the route parameters and env variables the correct branch to analyze
func getBranchToAnalyze(r *http.Request) string {
	branchUrlParam := r.URL.Query()["branch"]
	// check whether the param was set. If it was return this branch name, else return the default one
	if len(branchUrlParam) > 0 {
		return branchUrlParam[0]
	} else {
		return os.Getenv("GO_GIT_DEFAULT_BRANCH")
	}
}

// convertTimestampStringToTime is a timestamp converter to the time interpretation of go
func convertTimestampStringToTime(rfc3339time string) (time.Time, error) {
	commitsSinceTime, err := time.Parse(time.RFC3339, rfc3339time)
	if err != nil {
		return time.Time{}, err
	}
	return commitsSinceTime, nil
}
