package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestGetRepositoryFromRequest_Valid(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions?repositoryUrl=https://github.com/neow3j/neow3j.git")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	repo, err := getRepositoryFromRequest(&req)
	assert.Equal(t, nil, err)
	assert.Equal(t, "https://github.com/neow3j/neow3j.git", repo[0])
}

func TestGetRepositoryFromRequest_MultiParam(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions?repositoryUrl=https://github.com/neow3j/neow3j.git&repositoryUrl=https://github.com/go-git/go-git.git")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	repo, err := getRepositoryFromRequest(&req)
	assert.Equal(t, nil, err, "err should be nil")
	assert.Equal(t, "https://github.com/neow3j/neow3j.git", repo[0], "they should be equal")
	assert.Equal(t, "https://github.com/go-git/go-git.git", repo[1], "they should be equal")
}

func TestGetRepositoryFromRequest_NoRepo(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	repo, err := getRepositoryFromRequest(&req)
	assert.NotEqual(t, nil, err, "err should not be nil")
	assert.Equal(t, "repository not found", err.Error(), "should throw repo not found error")
	assert.Nil(t, repo, "they should be equal")
}

// getTimeRange

func TestGetTimeRange_NoInput(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	since, until, err := getTimeRange(&req)
	var defaultTime time.Time
	assert.Equal(t, defaultTime, since, "should take default time")
	assert.Equal(t, defaultTime, until, "should take default time")
	assert.Equal(t, nil, err, "should take default time")
}

func TestGetTimeRange_SinceValid(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions?since=2020-01-22T15:04:05Z")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	since, until, err := getTimeRange(&req)
	var defaultTime time.Time
	sinceCorrect, _ := time.Parse(time.RFC3339, "2020-01-22T15:04:05Z")
	assert.Equal(t, sinceCorrect, since, "sould be equal")
	assert.Equal(t, defaultTime, until, "should take default time")
	assert.Equal(t, nil, err, "should take default time")
}

func TestGetTimeRange_UntilValid(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions?until=2020-01-22T15:04:05Z")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	since, until, err := getTimeRange(&req)
	var defaultTime time.Time
	untilCorrect, _ := time.Parse(time.RFC3339, "2020-01-22T15:04:05Z")
	assert.Equal(t, defaultTime, since)
	assert.Equal(t, untilCorrect, until)
	assert.Equal(t, nil, err)
}

func TestGetTimeRange_BothValid(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions?until=2020-01-22T15:04:05Z&since=2020-01-22T15:04:05Z")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	since, until, err := getTimeRange(&req)
	correctTime, _ := time.Parse(time.RFC3339, "2020-01-22T15:04:05Z")
	assert.Equal(t, correctTime, since)
	assert.Equal(t, correctTime, until)
	assert.Equal(t, nil, err)
}

func TestGetTimeRange_UntilInvalid(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions?until=2020-01-22T15:04:05:11Z&since=2020-01-22T15:04:05Z")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	since, until, err := getTimeRange(&req)
	var defaultTime time.Time
	expectedErr := errors.New("parsing time \"2020-01-22T15:04:05:11Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \":11Z\" as \"Z07:00\"")
	assert.Equal(t, defaultTime, since)
	assert.Equal(t, defaultTime, until)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
}

func TestGetTimeRange_SinceInvalid(t *testing.T) {
	uri, _ := url.Parse("http://localhost:8080/contributions?since=2020-01-22T15:04:05:11Z&until=2020-01-22T15:04:05Z")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	since, until, err := getTimeRange(&req)
	var defaultTime time.Time
	expectedErr := errors.New("parsing time \"2020-01-22T15:04:05:11Z\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \":11Z\" as \"Z07:00\"")
	assert.Equal(t, defaultTime, since)
	assert.Equal(t, defaultTime, until)
	assert.NotEqual(t, nil, err)
	assert.Equal(t, expectedErr.Error(), err.Error())
}
