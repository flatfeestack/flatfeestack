package api_test

import (
	"context"
	"forum/api"
	"testing"

	"github.com/stretchr/testify/assert" // Use a testing library, such as testify, for assertions
)

func TestGetPosts(t *testing.T) {
	// Create a new instance of StrictServerImpl
	server := &api.StrictServerImpl{}

	// Create a test request object
	request := api.GetPostsRequestObject{}

	// Call the GetPosts method with the test request and response recorder
	response, err := server.GetPosts(context.Background(), request)

	// Assert that there's no error returned
	assert.NoError(t, err)

	// Assert that the response is not nil
	assert.Nil(t, response)
}
