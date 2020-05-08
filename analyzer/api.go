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

	var contributions []Contribution
	for k, v := range contributionMap {
		contributions = append(contributions, Contribution{Contributor: k, Changes: v})
	}
	json.NewEncoder(w).Encode(contributions)
}

func convertTimestampStringToTime(stamp string) (time.Time, error) {
	commitsSinceInt, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(commitsSinceInt, 0), nil
}
