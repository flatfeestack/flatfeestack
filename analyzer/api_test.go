package main

import (
	"net/http"
	"net/url"
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestGetShouldAnalyzePlatformInformation(t *testing.T)  {
	uri, _ := url.Parse("http://localhost:8080/api?platformInformation=true")
	req := http.Request{
		Method: "GET",
		URL:    uri,
	}
	info := getShouldAnalyzePlatformInformation(&req)
	assert.Equal(t, true, info, "they should be equal")
}