package clients

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGitHubFetch(t *testing.T) {
	t.Skip("This is for manual testing, we are calling coingecko here")
	repos, err := fetchGithubRepoSearch("tomp2p")
	assert.Nil(t, err)
	assert.True(t, len(repos) > 0)
	assert.Equal(t, "tomp2p/TomP2P", repos[0].Name)
	repo, err := fetchGithubRepoById(3114475)
	assert.Nil(t, err)
	assert.Equal(t, "tomp2p/TomP2P", repo.Name)
}
