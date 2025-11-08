package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertOrUpdateTrust_NewTrust(t *testing.T) {
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

	err := db.InsertOrUpdateTrust(trustEvent)
	require.NoError(t, err)
}

func TestInsertOrUpdateTrust_Untrust(t *testing.T) {
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
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent))

	untrustAt := time.Now().Add(time.Hour)
	untrustEvent := &TrustEvent{
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Inactive,
		UnTrustAt: &untrustAt,
	}

	err := db.InsertOrUpdateTrust(untrustEvent)
	require.NoError(t, err)
}

func TestInsertOrUpdateTrust_ErrorUntrustNotTrusting(t *testing.T) {
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

	err := db.InsertOrUpdateTrust(untrustEvent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot untrust")
}

func TestInsertOrUpdateTrust_ErrorAlreadyTrusting(t *testing.T) {
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
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent))

	trustAt2 := time.Now().Add(time.Hour)
	trustEvent2 := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &trustAt2,
	}

	err := db.InsertOrUpdateTrust(trustEvent2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already trusting")
}

func TestInsertOrUpdateTrust_ReTrust(t *testing.T) {
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
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent))

	untrustAt := time.Now().Add(time.Hour)
	untrustEvent := &TrustEvent{
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Inactive,
		UnTrustAt: &untrustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrust(untrustEvent))

	reTrustAt := time.Now().Add(2 * time.Hour)
	reTrustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   &reTrustAt,
	}

	err := db.InsertOrUpdateTrust(reTrustEvent)
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
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent))

	id, trust, untrust, err := db.FindLastEventTrustedRepo(user.Id, repo.Id)
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.NotNil(t, trust)
	assert.Nil(t, untrust)
}

func TestFindLastEventTrustedRepo_NotFound(t *testing.T) {
	TruncateAll(db, t)

	id, trust, untrust, err := db.FindLastEventTrustedRepo(uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Nil(t, id)
	assert.Nil(t, trust)
	assert.Nil(t, untrust)
}

func TestFindTrustedReposByUserId(t *testing.T) {
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
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent1))

	trustAt2 := time.Now()
	trustEvent2 := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo2.Id,
		EventType: Active,
		TrustAt:   &trustAt2,
	}
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent2))

	repos, err := db.FindTrustedReposByUserId(user.Id)
	require.NoError(t, err)
	assert.Len(t, repos, 2)
}

func TestFindTrustedReposByUserId_ExcludesUntrusted(t *testing.T) {
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
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent1))

	trustAt2 := time.Now()
	trustEvent2 := &TrustEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    untrustedRepo.Id,
		EventType: Active,
		TrustAt:   &trustAt2,
	}
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent2))

	untrustAt := time.Now().Add(time.Hour)
	untrustEvent := &TrustEvent{
		Uid:       user.Id,
		RepoId:    untrustedRepo.Id,
		EventType: Inactive,
		UnTrustAt: &untrustAt,
	}
	require.NoError(t, db.InsertOrUpdateTrust(untrustEvent))

	repos, err := db.FindTrustedReposByUserId(user.Id)
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
	require.NoError(t, db.InsertOrUpdateTrust(trustEvent))

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