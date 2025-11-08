package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertOrUpdateRepo(t *testing.T) {
	TruncateAll(db, t)

	gitUrl := "https://github.com/test/repo"
	url := "https://github.com/test/repo"
	name := "test-repo"
	desc := "Test repository"
	source := "github"

	repo := &Repo{
		Id:          uuid.New(),
		GitUrl:      &gitUrl,
		Url:         &url,
		Name:        &name,
		Description: &desc,
		Source:      &source,
		CreatedAt:   time.Now(),
	}

	err := db.InsertOrUpdateRepo(repo)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, repo.Id)

	// Verify insert
	found, err := db.FindRepoById(repo.Id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, *repo.GitUrl, *found.GitUrl)
	assert.Equal(t, *repo.Name, *found.Name)
}

func TestInsertOrUpdateRepo_UpdateOnConflict(t *testing.T) {
	TruncateAll(db, t)

	gitUrl := "https://github.com/test/conflict-repo"
	name1 := "original-name"
	name2 := "updated-name"

	repo1 := &Repo{
		Id:        uuid.New(),
		GitUrl:    &gitUrl,
		Name:      &name1,
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.InsertOrUpdateRepo(repo1))

	repo2 := &Repo{
		Id:        uuid.New(),
		GitUrl:    &gitUrl,
		Name:      &name2,
		CreatedAt: time.Now(),
	}
	err := db.InsertOrUpdateRepo(repo2)
	require.NoError(t, err)
	
	// Should return same ID and update fields
	assert.Equal(t, repo1.Id, repo2.Id)
	
	found, err := db.FindRepoById(repo1.Id)
	require.NoError(t, err)
	assert.Equal(t, *repo2.Name, *found.Name)
}

func TestFindRepoById_NotFound(t *testing.T) {
	TruncateAll(db, t)

	found, err := db.FindRepoById(uuid.New())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestFindReposByName(t *testing.T) {
	TruncateAll(db, t)

	name := "common-name"
	gitUrl1 := "https://github.com/test/repo1"
	gitUrl2 := "https://github.com/test/repo2"

	repo1 := &Repo{
		Id:        uuid.New(),
		GitUrl:    &gitUrl1,
		Name:      &name,
		CreatedAt: time.Now(),
	}
	repo2 := &Repo{
		Id:        uuid.New(),
		GitUrl:    &gitUrl2,
		Name:      &name,
		CreatedAt: time.Now(),
	}

	require.NoError(t, db.InsertOrUpdateRepo(repo1))
	require.NoError(t, db.InsertOrUpdateRepo(repo2))

	repos, err := db.FindReposByName(name)
	require.NoError(t, err)
	assert.Len(t, repos, 2)
	
	// Verify both repos are present
	foundIds := []uuid.UUID{repos[0].Id, repos[1].Id}
	assert.Contains(t, foundIds, repo1.Id)
	assert.Contains(t, foundIds, repo2.Id)
}

func TestFindReposByName_NotFound(t *testing.T) {
	TruncateAll(db, t)

	repos, err := db.FindReposByName("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, repos)
}