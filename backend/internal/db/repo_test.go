package db

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func insertTestRepo(t *testing.T) *Repo {
	return insertTestRepoGitUrl(t, "git-url")
}

func insertTestRepoGitUrl(t *testing.T, gitUrl string) *Repo {
	rid := uuid.New()
	r := Repo{
		Id:          rid,
		Url:         stringPointer("url"),
		GitUrl:      stringPointer(gitUrl),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := InsertOrUpdateRepo(&r)
	assert.Nil(t, err)
	r2, err := FindRepoById(r.Id)
	assert.Nil(t, err)
	return r2
}

func TestRepoNotFound(t *testing.T) {
	setup()
	defer teardown()
	insertTestRepo(t)

	r2, err := FindRepoById(uuid.New())
	assert.Nil(t, err)
	assert.Nil(t, r2)
}

func TestRepoFound(t *testing.T) {
	setup()
	defer teardown()
	r := insertTestRepo(t)

	r2, err := FindRepoById(r.Id)
	assert.Nil(t, err)
	assert.NotNil(t, r2)
}
