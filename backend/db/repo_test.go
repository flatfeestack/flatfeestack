package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repo.go tests
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
		Score:       100,
		CreatedAt:   time.Now(),
	}

	err := db.InsertOrUpdateRepo(repo)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, repo.Id)
}

func TestInsertOrUpdateRepo_Conflict(t *testing.T) {
	TruncateAll(db, t)

	gitUrl := "https://github.com/test/conflict-repo"
	name := "conflict-repo"

	repo1 := &Repo{
		Id:        uuid.New(),
		GitUrl:    &gitUrl,
		Name:      &name,
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.InsertOrUpdateRepo(repo1))

	repo2 := &Repo{
		Id:        uuid.New(),
		GitUrl:    &gitUrl,
		Name:      &name,
		CreatedAt: time.Now(),
	}
	err := db.InsertOrUpdateRepo(repo2)
	require.NoError(t, err)
	assert.Equal(t, repo1.Id, repo2.Id)
}

func TestFindRepoById(t *testing.T) {
	TruncateAll(db, t)

	gitUrl := "https://github.com/test/find-repo"
	name := "find-repo"
	repo := &Repo{
		Id:        uuid.New(),
		GitUrl:    &gitUrl,
		Name:      &name,
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.InsertOrUpdateRepo(repo))

	found, err := db.FindRepoById(repo.Id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, repo.Id, found.Id)
	assert.Equal(t, gitUrl, *found.GitUrl)
}

func TestFindRepoById_NotFound(t *testing.T) {
	TruncateAll(db, t)

	found, err := db.FindRepoById(uuid.New())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestFindRepoWithTrustDateById(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitUrl := "https://github.com/test/trusted-repo"
	repo := createTestRepo(t, db, gitUrl)

	trustAt := time.Now()
	trustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &trustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent))

	found, err := db.FindRepoWithTrustDateById(repo.Id)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, repo.Id, found.Id)
	assert.False(t, found.TrustAt.IsZero())
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
}

// trusted_repos.go tests
func TestInsertOrUpdateTrustRepo_NewTrust(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/trust-repo")

	trustAt := time.Now()
	trustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &trustAt,
	}

	err := db.InsertOrUpdateTrustRepo(trustEvent)
	require.NoError(t, err)
}

func TestInsertOrUpdateTrustRepo_Untrust(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/untrust-repo")

	trustAt := time.Now()
	trustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &trustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent))

	untrustAt := time.Now().Add(time.Hour)
	untrustEvent := &TrustEvent{
		Uid:         user.Id,
		RepoId:      repo.Id,
		EventType:   Inactive,
		UnTrustAt:   &untrustAt,
	}

	err := db.InsertOrUpdateTrustRepo(untrustEvent)
	require.NoError(t, err)
}

func TestInsertOrUpdateTrustRepo_ErrorUntrustNotTrusted(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/error-repo")

	untrustAt := time.Now()
	untrustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Inactive,
		UnTrustAt: &untrustAt,
	}

	err := db.InsertOrUpdateTrustRepo(untrustEvent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot untrust")
}

func TestInsertOrUpdateTrustRepo_ErrorAlreadyTrusted(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/double-trust")

	trustAt := time.Now()
	trustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &trustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent))

	trustAt2 := time.Now().Add(time.Hour)
	trustEvent2 := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &trustAt2,
	}

	err := db.InsertOrUpdateTrustRepo(trustEvent2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already trusting")
}

func TestInsertOrUpdateTrustRepo_ReTrust(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/retrust-repo")

	trustAt := time.Now()
	trustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &trustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent))

	untrustAt := time.Now().Add(time.Hour)
	untrustEvent := &TrustEvent{
		Uid:         user.Id,
		RepoId:      repo.Id,
		EventType:   Inactive,
		UnTrustAt:   &untrustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(untrustEvent))

	reTrustAt := time.Now().Add(2 * time.Hour)
	reTrustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &reTrustAt,
	}

	err := db.InsertOrUpdateTrustRepo(reTrustEvent)
	require.NoError(t, err)
}

func TestFindLastEventTrustedRepo(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/last-event")

	trustAt := time.Now()
	trustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &trustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent))

	id, trust, untrust, err := db.FindLastEventTrustedRepo(repo.Id)
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.NotNil(t, trust)
	assert.Nil(t, untrust)
}

func TestFindLastEventTrustedRepo_NotFound(t *testing.T) {
	TruncateAll(db, t)

	id, trust, untrust, err := db.FindLastEventTrustedRepo(uuid.New())
	require.NoError(t, err)
	assert.Nil(t, id)
	assert.Nil(t, trust)
	assert.Nil(t, untrust)
}

func TestFindTrustedRepos(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	repo1 := createTestRepo(t, db, "https://github.com/test/trusted1")
	repo2 := createTestRepo(t, db, "https://github.com/test/trusted2")

	trustAt1 := time.Now()
	trustEvent1 := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo1.Id,
		EventType: Active,
		TrustAt:   &trustAt1,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent1))

	trustAt2 := time.Now()
	trustEvent2 := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo2.Id,
		EventType: Active,
		TrustAt:   &trustAt2,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent2))

	repos, err := db.FindTrustedRepos()
	require.NoError(t, err)
	assert.Len(t, repos, 2)
}

func TestFindTrustedRepos_ExcludesUntrusted(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	trustedRepo := createTestRepo(t, db, "https://github.com/test/trusted")
	untrustedRepo := createTestRepo(t, db, "https://github.com/test/untrusted")

	trustAt1 := time.Now()
	trustEvent1 := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    trustedRepo.Id,
		EventType: Active,
		TrustAt:   &trustAt1,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent1))

	trustAt2 := time.Now()
	trustEvent2 := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    untrustedRepo.Id,
		EventType: Active,
		TrustAt:   &trustAt2,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent2))

	untrustAt := time.Now().Add(time.Hour)
	untrustEvent := &TrustEvent{
		Uid:         user.Id,
		RepoId:      untrustedRepo.Id,
		EventType:   Inactive,
		UnTrustAt:   &untrustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(untrustEvent))

	repos, err := db.FindTrustedRepos()
	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, trustedRepo.Id, repos[0].Id)
}

func TestGetTrustedReposFromList(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	trustedRepo := createTestRepo(t, db, "https://github.com/test/trusted")
	untrustedRepo := createTestRepo(t, db, "https://github.com/test/untrusted")

	trustAt := time.Now()
	trustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    trustedRepo.Id,
		EventType: Active,
		TrustAt:   &trustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent))

	repoIds := []uuid.UUID{trustedRepo.Id, untrustedRepo.Id}
	trustedIds, err := db.GetTrustedReposFromList(repoIds)
	require.NoError(t, err)
	assert.Len(t, trustedIds, 1)
	assert.Contains(t, trustedIds, trustedRepo.Id)
}

func TestGetTrustedReposFromList_EmptyInput(t *testing.T) {
	TruncateAll(db, t)

	trustedIds, err := db.GetTrustedReposFromList([]uuid.UUID{})
	require.NoError(t, err)
	assert.Nil(t, trustedIds)
}

func TestCountReposForUsers_EmptyList(t *testing.T) {
	TruncateAll(db, t)

	count, err := db.CountReposForUsers([]uuid.UUID{}, 3)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCountReposForUsers_NegativeMonths(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")

	count, err := db.CountReposForUsers([]uuid.UUID{sponsor.Id}, -1)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}