package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func getAllContributions(w http.ResponseWriter, r *http.Request) {
	repositoryUrl := r.URL.Query()["repositoryUrl"]
	if len(repositoryUrl) < 1 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Repository not found")
		return
	}

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
			fmt.Fprintf(w, err.Error())
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
			fmt.Fprintf(w, err.Error())
			return
		}
	}

	contributionMap, err := analyzeRepository(repositoryUrl[0], commitsSince, commitsUntil)
	if err != nil {
		if strings.Contains(err.Error(), "authentication") {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, err.Error())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
		}
		return
	}

	if analyzePlatformInformation {
		issues, err := getPlatformInformation(repositoryUrl[0], commitsSince, commitsUntil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
		}

		var contributions []ContributionWithPlatformInformation
		for k, v := range contributionMap {
			userInformation, err := getPlatformInformationFromUser(repositoryUrl[0], issues, k.Email)
			if err != nil {
				fmt.Println(err)
			}
			contributions = append(contributions, ContributionWithPlatformInformation{
				GitInformation:      v,
				PlatformInformation: userInformation,
			})
		}
		json.NewEncoder(w).Encode(contributions)
	} else {
		var contributions []Contribution
		for _, v := range contributionMap {
			contributions = append(contributions, v)
		}
		json.NewEncoder(w).Encode(contributions)
	}
}

func getContributionWeights(w http.ResponseWriter, r *http.Request) {
	repositoryUrl := r.URL.Query()["repositoryUrl"]
	if len(repositoryUrl) < 1 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Repository not found")
		return
	}

	var err error

	commitsSinceString := r.URL.Query()["since"]
	var commitsSince time.Time
	if len(commitsSinceString) > 0 {
		commitsSince, err = convertTimestampStringToTime(commitsSinceString[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}

	}

	commitsUntilString := r.URL.Query()["until"]
	var commitsUntil time.Time
	if len(commitsUntilString) > 0 {
		commitsUntil, err = convertTimestampStringToTime(commitsUntilString[0])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
			return
		}
	}

	contributionMap, err := analyzeRepository(repositoryUrl[0], commitsSince, commitsUntil)
	if err != nil {
		if strings.Contains(err.Error(), "authentication") {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, err.Error())
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, err.Error())
		}
		return
	}

	weightsMap, err := weightContributions(contributionMap)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, err.Error())
	}

	var contributionWeights []FlatFeeWeight
	for _, v := range weightsMap {
		contributionWeights = append(contributionWeights, v)
	}
	json.NewEncoder(w).Encode(contributionWeights)
}

func convertTimestampStringToTime(stamp string) (time.Time, error) {
	commitsSinceInt, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(commitsSinceInt, 0), nil
}
