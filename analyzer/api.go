package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

	// get the repository url from the request
	repositoryUrl := r.URL.Query()["repositoryUrl"]
	if len(repositoryUrl) < 1 {
		w.WriteHeader(http.StatusNotFound)
		_, fmtErr := fmt.Fprintf(w, "Repository not found")
		if fmtErr != nil {
			fmt.Println("Could not format", fmtErr)
		}
		return
	}

	// detect whether the platformInformation flag in the request was set
	analyzePlatformInformation := false
	platformInformationUrlParam := r.URL.Query()["platformInformation"]
	if len(platformInformationUrlParam) > 0 && platformInformationUrlParam[0] == "true" {
		analyzePlatformInformation = true
	}

	// convert since timestamp into go time
	var err error
	commitsSinceString := r.URL.Query()["since"]
	var commitsSince time.Time
	if len(commitsSinceString) > 0 {
		commitsSince, err = convertTimestampStringToTime(commitsSinceString[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, fmtErr := fmt.Fprintf(w, err.Error())
			if fmtErr != nil {
				fmt.Println("Could not format", fmtErr)
			}
			return
		}

	}

	// convert until timestamp into go time
	commitsUntilString := r.URL.Query()["until"]
	var commitsUntil time.Time
	if len(commitsUntilString) > 0 {
		commitsUntil, err = convertTimestampStringToTime(commitsUntilString[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, fmtErr := fmt.Fprintf(w, err.Error())
			if fmtErr != nil {
				fmt.Println("Could not format", fmtErr)
			}
			return
		}
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
		routineContributionMap, routineErr := analyzeRepository(repositoryUrl[0], commitsSince, commitsUntil)
		gitAnalyzationChannel <- GitAnalyzationChannel{
			Result: routineContributionMap,
			Reason: routineErr,
		}
		close(gitAnalyzationChannel)
	}()

	// execute go routine to fetch the platform information only when the platformInformation flag is set
	if analyzePlatformInformation {
		go func() {
			routineIssues, routinePullRequests, routineErr := getPlatformInformation(repositoryUrl[0], commitsSince, commitsUntil)
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
					w.WriteHeader(http.StatusUnauthorized)
					_, fmtErr := fmt.Fprintf(w, msg1.Reason.Error())
					if fmtErr != nil {
						fmt.Println("Could not format", fmtErr)
					}
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					_, fmtErr := fmt.Fprintf(w, msg1.Reason.Error())
					if fmtErr != nil {
						fmt.Println("Could not format", fmtErr)
					}
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
				w.WriteHeader(http.StatusInternalServerError)
				_, fmtErr := fmt.Fprintf(w, msg2.Reason.Error())
				if fmtErr != nil {
					fmt.Println("Could not format", fmtErr)
				}
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
			userInformation, err := getPlatformInformationFromUser(repositoryUrl[0], issues, pullRequests, k.Email)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, fmtErr := fmt.Fprintf(w, err.Error())
				if fmtErr != nil {
					fmt.Println("Could not format", fmtErr)
				}
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
	repositoryUrl := r.URL.Query()["repositoryUrl"]
	if len(repositoryUrl) < 1 {
		w.WriteHeader(http.StatusNotFound)
		_, fmtErr := fmt.Fprintf(w, "Repository not found")
		if fmtErr != nil {
			fmt.Println("Could not format", fmtErr)
		}
		return
	}

	var err error

	commitsSinceString := r.URL.Query()["since"]
	var commitsSince time.Time
	if len(commitsSinceString) > 0 {
		commitsSince, err = convertTimestampStringToTime(commitsSinceString[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, fmtErr := fmt.Fprintf(w, err.Error())
			if fmtErr != nil {
				fmt.Println("Could not format", fmtErr)
			}
			return
		}

	}

	commitsUntilString := r.URL.Query()["until"]
	var commitsUntil time.Time
	if len(commitsUntilString) > 0 {
		commitsUntil, err = convertTimestampStringToTime(commitsUntilString[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, fmtErr := fmt.Fprintf(w, err.Error())
			if fmtErr != nil {
				fmt.Println("Could not format", fmtErr)
			}
			return
		}
	}

	contributionMap, err := analyzeRepository(repositoryUrl[0], commitsSince, commitsUntil)
	if err != nil {
		if strings.Contains(err.Error(), "authentication") {
			w.WriteHeader(http.StatusUnauthorized)
			_, fmtErr := fmt.Fprintf(w, err.Error())
			if fmtErr != nil {
				fmt.Println("Could not format", fmtErr)
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, fmtErr := fmt.Fprintf(w, err.Error())
			if fmtErr != nil {
				fmt.Println("Could not format", fmtErr)
			}
		}
		return
	}

	weightsMap, err := weightContributions(contributionMap)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, fmtErr := fmt.Fprintf(w, err.Error())
		if fmtErr != nil {
			fmt.Println("Could not format", fmtErr)
		}
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

func convertTimestampStringToTime(stamp string) (time.Time, error) {
	commitsSinceInt, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(commitsSinceInt, 0), nil
}
