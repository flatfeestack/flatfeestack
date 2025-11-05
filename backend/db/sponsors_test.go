package db

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sponsors.go tests
func TestInsertOrUpdateSponsor_NewSponsor(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/sponsor-repo")

	sponsorAt := time.Now()
	sponsorEvent := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		SponsorAt: &sponsorAt,
	}

	err := db.InsertOrUpdateSponsor(sponsorEvent)
	require.NoError(t, err)
}

func TestInsertOrUpdateSponsor_Unsponsor(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/unsponsor-repo")

	sponsorAt := time.Now()
	sponsorEvent := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		SponsorAt: &sponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent))

	unsponsorAt := time.Now().Add(time.Hour)
	unsponsorEvent := &SponsorEvent{
		Uid:           user.Id,
		RepoId:        repo.Id,
		EventType:     Inactive,
		UnSponsorAt:   &unsponsorAt,
	}

	err := db.InsertOrUpdateSponsor(unsponsorEvent)
	require.NoError(t, err)
}

func TestInsertOrUpdateSponsor_ErrorUnsponsorNotSponsoring(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/error-repo")

	unsponsorAt := time.Now()
	unsponsorEvent := &SponsorEvent{
		Id:          uuid.New(),
		Uid:         user.Id,
		RepoId:      repo.Id,
		EventType:   Inactive,
		UnSponsorAt: &unsponsorAt,
	}

	err := db.InsertOrUpdateSponsor(unsponsorEvent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unsponsor")
}

func TestInsertOrUpdateSponsor_ErrorAlreadySponsoring(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/double-sponsor")

	sponsorAt := time.Now()
	sponsorEvent := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		SponsorAt: &sponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent))

	sponsorAt2 := time.Now().Add(time.Hour)
	sponsorEvent2 := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		SponsorAt: &sponsorAt2,
	}

	err := db.InsertOrUpdateSponsor(sponsorEvent2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already sponsoring")
}

func TestInsertOrUpdateSponsor_ReSponsor(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/responsor-repo")

	sponsorAt := time.Now()
	sponsorEvent := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		SponsorAt: &sponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent))

	unsponsorAt := time.Now().Add(time.Hour)
	unsponsorEvent := &SponsorEvent{
		Uid:           user.Id,
		RepoId:        repo.Id,
		EventType:     Inactive,
		UnSponsorAt:   &unsponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(unsponsorEvent))

	reSponsorAt := time.Now().Add(2 * time.Hour)
	reSponsorEvent := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		SponsorAt: &reSponsorAt,
	}

	err := db.InsertOrUpdateSponsor(reSponsorEvent)
	require.NoError(t, err)
}

func TestFindLastEventSponsoredRepo(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/last-event")

	sponsorAt := time.Now()
	sponsorEvent := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo.Id,
		EventType: Active,
		SponsorAt: &sponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent))

	id, sponsor, unsponsor, err := db.FindLastEventSponsoredRepo(user.Id, repo.Id)
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.NotNil(t, sponsor)
	assert.Nil(t, unsponsor)
}

func TestFindLastEventSponsoredRepo_NotFound(t *testing.T) {
	TruncateAll(db, t)

	id, sponsor, unsponsor, err := db.FindLastEventSponsoredRepo(uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Nil(t, id)
	assert.Nil(t, sponsor)
	assert.Nil(t, unsponsor)
}

func TestFindSponsoredReposByUserId(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "sponsor@example.com")
	repo1 := createTestRepo(t, db, "https://github.com/test/sponsored1")
	repo2 := createTestRepo(t, db, "https://github.com/test/sponsored2")

	sponsorAt1 := time.Now()
	sponsorEvent1 := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo1.Id,
		EventType: Active,
		SponsorAt: &sponsorAt1,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent1))

	sponsorAt2 := time.Now()
	sponsorEvent2 := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    repo2.Id,
		EventType: Active,
		SponsorAt: &sponsorAt2,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent2))

	repos, err := db.FindSponsoredReposByUserId(user.Id)
	require.NoError(t, err)
	assert.Len(t, repos, 2)
}

func TestFindSponsoredReposByUserId_ExcludesUnsponsored(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "sponsor@example.com")
	sponsoredRepo := createTestRepo(t, db, "https://github.com/test/sponsored")
	unsponsoredRepo := createTestRepo(t, db, "https://github.com/test/unsponsored")

	sponsorAt1 := time.Now()
	sponsorEvent1 := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    sponsoredRepo.Id,
		EventType: Active,
		SponsorAt: &sponsorAt1,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent1))

	sponsorAt2 := time.Now()
	sponsorEvent2 := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user.Id,
		RepoId:    unsponsoredRepo.Id,
		EventType: Active,
		SponsorAt: &sponsorAt2,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent2))

	unsponsorAt := time.Now().Add(time.Hour)
	unsponsorEvent := &SponsorEvent{
		Uid:           user.Id,
		RepoId:        unsponsoredRepo.Id,
		EventType:     Inactive,
		UnSponsorAt:   &unsponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(unsponsorEvent))

	repos, err := db.FindSponsoredReposByUserId(user.Id)
	require.NoError(t, err)
	assert.Len(t, repos, 1)
	assert.Equal(t, sponsoredRepo.Id, repos[0].Id)
}

func TestFindSponsorsBetween(t *testing.T) {
	TruncateAll(db, t)

	user1 := createTestUser(t, db, "sponsor1@example.com")
	user2 := createTestUser(t, db, "sponsor2@example.com")
	repo1 := createTestRepo(t, db, "https://github.com/test/repo1")
	repo2 := createTestRepo(t, db, "https://github.com/test/repo2")

	baseTime := time.Now().Truncate(24 * time.Hour)

	sponsorAt := baseTime.AddDate(0, 0, -5)
	sponsorEvent1 := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user1.Id,
		RepoId:    repo1.Id,
		EventType: Active,
		SponsorAt: &sponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent1))

	sponsorEvent2 := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user1.Id,
		RepoId:    repo2.Id,
		EventType: Active,
		SponsorAt: &sponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent2))

	sponsorEvent3 := &SponsorEvent{
		Id:        uuid.New(),
		Uid:       user2.Id,
		RepoId:    repo1.Id,
		EventType: Active,
		SponsorAt: &sponsorAt,
	}
	require.NoError(t, db.InsertOrUpdateSponsor(sponsorEvent3))

	start := baseTime
	stop := baseTime.AddDate(0, 0, -10)

	results, err := db.FindSponsorsBetween(start, stop)
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

// multipliers.go tests
func TestInsertOrUpdateMultiplierRepo_NewMultiplier(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "multiplier@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/multiplier-repo")

	multiplierAt := time.Now()
	multiplierEvent := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          user.Id,
		RepoId:       repo.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}

	err := db.InsertOrUpdateMultiplierRepo(multiplierEvent)
	require.NoError(t, err)
}

func TestInsertOrUpdateMultiplierRepo_UnsetMultiplier(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "multiplier@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/unmultiplier-repo")

	multiplierAt := time.Now()
	multiplierEvent := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          user.Id,
		RepoId:       repo.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent))

	unmultiplierAt := time.Now().Add(time.Hour)
	unmultiplierEvent := &MultiplierEvent{
		Uid:              user.Id,
		RepoId:           repo.Id,
		EventType:        Inactive,
		UnMultiplierAt:   &unmultiplierAt,
	}

	err := db.InsertOrUpdateMultiplierRepo(unmultiplierEvent)
	require.NoError(t, err)
}

func TestInsertOrUpdateMultiplierRepo_ErrorUnsetNotSet(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "multiplier@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/error-repo")

	unmultiplierAt := time.Now()
	unmultiplierEvent := &MultiplierEvent{
		Id:             uuid.New(),
		Uid:            user.Id,
		RepoId:         repo.Id,
		EventType:      Inactive,
		UnMultiplierAt: &unmultiplierAt,
	}

	err := db.InsertOrUpdateMultiplierRepo(unmultiplierEvent)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot unset multiplier")
}

func TestInsertOrUpdateMultiplierRepo_ErrorAlreadyActive(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "multiplier@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/double-multiplier")

	multiplierAt := time.Now()
	multiplierEvent := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          user.Id,
		RepoId:       repo.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent))

	multiplierAt2 := time.Now().Add(time.Hour)
	multiplierEvent2 := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          user.Id,
		RepoId:       repo.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt2,
	}

	err := db.InsertOrUpdateMultiplierRepo(multiplierEvent2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "multiplier already active")
}

func TestFindLastEventMultiplierRepo(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "multiplier@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/last-event")

	multiplierAt := time.Now()
	multiplierEvent := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          user.Id,
		RepoId:       repo.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent))

	id, multiplier, unmultiplier, err := db.FindLastEventMultiplierRepo(user.Id, repo.Id)
	require.NoError(t, err)
	require.NotNil(t, id)
	assert.NotNil(t, multiplier)
	assert.Nil(t, unmultiplier)
}

func TestFindLastEventMultiplierRepo_NotFound(t *testing.T) {
	TruncateAll(db, t)

	id, multiplier, unmultiplier, err := db.FindLastEventMultiplierRepo(uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.Nil(t, id)
	assert.Nil(t, multiplier)
	assert.Nil(t, unmultiplier)
}

func TestFindMultiplierRepoByUserId(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "multiplier@example.com")
	repo1 := createTestRepo(t, db, "https://github.com/test/multiplied1")
	repo2 := createTestRepo(t, db, "https://github.com/test/multiplied2")

	multiplierAt1 := time.Now()
	multiplierEvent1 := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          user.Id,
		RepoId:       repo1.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt1,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent1))

	multiplierAt2 := time.Now()
	multiplierEvent2 := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          user.Id,
		RepoId:       repo2.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt2,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent2))

	repos, err := db.FindMultiplierRepoByUserId(user.Id)
	require.NoError(t, err)
	assert.Len(t, repos, 2)
}

func TestGetFoundationsSupportingRepo(t *testing.T) {
	TruncateAll(db, t)

	foundation1 := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "foundation1@example.com",
			Name:      "Foundation 1",
			CreatedAt: time.Now(),
		},
		Multiplier:           true,
		MultiplierDailyLimit: 5000,
	}
	require.NoError(t, db.InsertUser(foundation1))
	require.NoError(t, db.UpdateMultiplier(foundation1.Id, true))
	require.NoError(t, db.UpdateMultiplierDailyLimit(foundation1.Id, 5000))

	repo := createTestRepo(t, db, "https://github.com/test/supported-repo")

	multiplierAt := time.Now()
	multiplierEvent := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          foundation1.Id,
		RepoId:       repo.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent))

	foundations, err := db.GetFoundationsSupportingRepo(repo.Id)
	require.NoError(t, err)
	assert.Len(t, foundations, 1)
	assert.Equal(t, foundation1.Id, foundations[0].Id)
	assert.Equal(t, 5000, foundations[0].MultiplierDailyLimit)
}

func TestGetAllFoundationsSupportingRepos(t *testing.T) {
	TruncateAll(db, t)

	foundation := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "foundation@example.com",
			Name:      "Foundation",
			CreatedAt: time.Now(),
		},
		Multiplier:           true,
		MultiplierDailyLimit: 10000,
	}
	require.NoError(t, db.InsertUser(foundation))
	require.NoError(t, db.UpdateMultiplier(foundation.Id, true))
	require.NoError(t, db.UpdateMultiplierDailyLimit(foundation.Id, 10000))

	repo1 := createTestRepo(t, db, "https://github.com/test/repo1")
	repo2 := createTestRepo(t, db, "https://github.com/test/repo2")

	multiplierAt := time.Now()
	multiplierEvent1 := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          foundation.Id,
		RepoId:       repo1.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent1))

	multiplierEvent2 := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          foundation.Id,
		RepoId:       repo2.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent2))

	foundations, totalCount, err := db.GetAllFoundationsSupportingRepos([]uuid.UUID{repo1.Id, repo2.Id})
	require.NoError(t, err)
	assert.Len(t, foundations, 1)
	assert.Equal(t, 2, totalCount)
	assert.Len(t, foundations[0].RepoIds, 2)
}

func TestGetAllFoundationsSupportingRepos_EmptyList(t *testing.T) {
	TruncateAll(db, t)

	foundations, totalCount, err := db.GetAllFoundationsSupportingRepos([]uuid.UUID{})
	require.NoError(t, err)
	assert.Nil(t, foundations)
	assert.Equal(t, 0, totalCount)
}

func TestGetMultiplierCount(t *testing.T) {
	TruncateAll(db, t)

	multiplier1 := createTestUser(t, db, "multiplier1@example.com")
	multiplier2 := createTestUser(t, db, "multiplier2@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/multiplied-repo")

	multiplierAt := time.Now()
	multiplierEvent1 := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          multiplier1.Id,
		RepoId:       repo.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent1))

	multiplierEvent2 := &MultiplierEvent{
		Id:           uuid.New(),
		Uid:          multiplier2.Id,
		RepoId:       repo.Id,
		EventType:    Active,
		MultiplierAt: &multiplierAt,
	}
	require.NoError(t, db.InsertOrUpdateMultiplierRepo(multiplierEvent2))

	count, err := db.GetMultiplierCount(repo.Id, []uuid.UUID{multiplier1.Id, multiplier2.Id})
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestGetMultiplierCount_EmptyList(t *testing.T) {
	TruncateAll(db, t)
	repo := createTestRepo(t, db, "https://github.com/test/no-multipliers")

	count, err := db.GetMultiplierCount(repo.Id, []uuid.UUID{})
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}