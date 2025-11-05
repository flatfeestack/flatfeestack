package db

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertContribution(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	err := db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "USD",
		time.Now(), time.Now(), false,
	)
	require.NoError(t, err)
}

func TestFindContributions_AsContributor(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1500), "USD", day, time.Now(), false,
	))

	contributions, err := db.FindContributions(contributor.Id, true)
	require.NoError(t, err)
	assert.Len(t, contributions, 1)
	assert.Equal(t, big.NewInt(1500), contributions[0].Balance)
	assert.Equal(t, "USD", contributions[0].Currency)
	assert.False(t, contributions[0].FoundationPayment)
}

func TestFindContributions_AsSponsor(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(2500), "EUR", day, time.Now(), false,
	))

	contributions, err := db.FindContributions(sponsor.Id, false)
	require.NoError(t, err)
	assert.Len(t, contributions, 1)
	assert.Equal(t, big.NewInt(2500), contributions[0].Balance)
	assert.Equal(t, "EUR", contributions[0].Currency)
}

func TestFindContributions_MultipleContributions(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo1 := createTestRepo(t, db, "https://github.com/test/repo1")
	repo2 := createTestRepo(t, db, "https://github.com/test/repo2")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo1.Id,
		big.NewInt(1000), "USD", day, time.Now(), false,
	))
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo2.Id,
		big.NewInt(2000), "USD", day.AddDate(0, 0, 1), time.Now(), false,
	))

	contributions, err := db.FindContributions(contributor.Id, true)
	require.NoError(t, err)
	assert.Len(t, contributions, 2)
}

func TestInsertFutureContribution(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	futureDay := time.Now().AddDate(0, 0, 7)
	err := db.InsertFutureContribution(
		sponsor.Id, repo.Id, big.NewInt(5000),
		"USD", futureDay, time.Now(), false,
	)
	require.NoError(t, err)
}

func TestInsertOrUpdateFutureContribution_Insert(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	futureDay := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	err := db.InsertOrUpdateFutureContribution(
		sponsor.Id, repo.Id, big.NewInt(3000),
		"USD", futureDay, time.Now(), false,
	)
	require.NoError(t, err)

	balances, err := db.FindSumFutureBalanceByRepoId(repo.Id)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(3000), balances["USD"])
}

func TestInsertOrUpdateFutureContribution_Update(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	futureDay := time.Now().AddDate(0, 0, 7).Truncate(24 * time.Hour)
	require.NoError(t, db.InsertOrUpdateFutureContribution(
		sponsor.Id, repo.Id, big.NewInt(2000),
		"USD", futureDay, time.Now(), false,
	))

	require.NoError(t, db.InsertOrUpdateFutureContribution(
		sponsor.Id, repo.Id, big.NewInt(1000),
		"USD", futureDay, time.Now(), false,
	))

	balances, err := db.FindSumFutureBalanceByRepoId(repo.Id)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(3000), balances["USD"])
}

func TestFindSumDailyContributors(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "USD", day, time.Now(), false,
	))
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(500), "USD", day, time.Now(), false,
	))
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(2000), "EUR", day, time.Now(), false,
	))

	balances, err := db.FindSumDailyContributors(contributor.Id)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(1500), balances["USD"])
	assert.Equal(t, big.NewInt(2000), balances["EUR"])
}

func TestFindSumDailyContributors_NoContributions(t *testing.T) {
	TruncateAll(db, t)

	contributor := createTestUser(t, db, "contributor@example.com")

	balances, err := db.FindSumDailyContributors(contributor.Id)
	require.NoError(t, err)
	assert.Empty(t, balances)
}

func TestFindSumDailySponsors(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(3000), "USD", day, time.Now(), false,
	))
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "EUR", day, time.Now(), false,
	))

	balances, err := db.FindSumDailySponsors(sponsor.Id)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(3000), balances["USD"])
	assert.Equal(t, big.NewInt(1000), balances["EUR"])
}

func TestFindSumDailySponsors_ExcludesFoundationPayments(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(2000), "USD", day, time.Now(), false,
	))
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "USD", day, time.Now(), true,
	))

	balances, err := db.FindSumDailySponsors(sponsor.Id)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(2000), balances["USD"])
}

func TestFindSumDailySponsorsFromFoundation(t *testing.T) {
	TruncateAll(db, t)

	foundation := createTestUser(t, db, "foundation@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		foundation.Id, contributor.Id, repo.Id,
		big.NewInt(5000), "USD", day, time.Now(), true,
	))
	require.NoError(t, db.InsertContribution(
		foundation.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "USD", day, time.Now(), false,
	))

	balances, err := db.FindSumDailySponsorsFromFoundation(foundation.Id)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(5000), balances["USD"])
}

func TestFindSumDailySponsorsFromFoundationByCurrency(t *testing.T) {
	TruncateAll(db, t)

	foundation := createTestUser(t, db, "foundation@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		foundation.Id, contributor.Id, repo.Id,
		big.NewInt(3000), "USD", day, time.Now(), true,
	))
	require.NoError(t, db.InsertContribution(
		foundation.Id, contributor.Id, repo.Id,
		big.NewInt(2000), "EUR", day, time.Now(), true,
	))

	balance, err := db.FindSumDailySponsorsFromFoundationByCurrency(foundation.Id, "USD")
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(3000), balance)
}

func TestFindSumDailySponsorsFromFoundationByCurrency_NoCurrency(t *testing.T) {
	TruncateAll(db, t)
	
	foundation := createTestUser(t, db, "foundation@example.com")

	balance, err := db.FindSumDailySponsorsFromFoundationByCurrency(foundation.Id, "JPY")
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0), balance)
}

func TestFindContributionsGroupedByCurrencyAndRepo(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo1 := createTestRepo(t, db, "https://github.com/test/repo1")
	repo2 := createTestRepo(t, db, "https://github.com/test/repo2")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo1.Id,
		big.NewInt(1000), "USD", day, time.Now(), false,
	))
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo2.Id,
		big.NewInt(2000), "USD", day, time.Now(), false,
	))

	contributions, err := db.FindContributionsGroupedByCurrencyAndRepo(sponsor.Id)
	require.NoError(t, err)
	require.NotNil(t, contributions["USD"])
	assert.Len(t, contributions["USD"], 2)
}

func TestFindSumFutureBalanceByRepoId(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	futureDay := time.Now().AddDate(0, 0, 7)
	require.NoError(t, db.InsertFutureContribution(
		sponsor.Id, repo.Id, big.NewInt(4000),
		"USD", futureDay, time.Now(), false,
	))

	balances, err := db.FindSumFutureBalanceByRepoId(repo.Id)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(4000), balances["USD"])
}

func TestFindSumDailyBalanceByRepoId(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(7000), "USD", day, time.Now(), false,
	))

	balances, err := db.FindSumDailyBalanceByRepoId(repo.Id)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(7000), balances["USD"])
}

func TestFindOwnContributionIds(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "USD", day, time.Now(), false,
	))
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(500), "USD", day, time.Now(), false,
	))

	ids, err := db.FindOwnContributionIds(contributor.Id, "USD")
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestSumTotalEarnedAmountForContributionIds(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "USD", day, time.Now(), false,
	))
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1500), "USD", day, time.Now(), false,
	))

	ids, err := db.FindOwnContributionIds(contributor.Id, "USD")
	require.NoError(t, err)

	total, err := db.SumTotalEarnedAmountForContributionIds(ids)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(2500), total)
}

func TestSumTotalEarnedAmountForContributionIds_EmptyList(t *testing.T) {
	TruncateAll(db, t)

	total, err := db.SumTotalEarnedAmountForContributionIds([]uuid.UUID{})
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(0), total)
}

func TestGetActiveSponsors(t *testing.T) {
	TruncateAll(db, t)

	sponsor := createTestUser(t, db, "sponsor@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	trustEvent := &TrustEvent{
		Id:        uuid.New(),
		Uid:       sponsor.Id,
		RepoId:    repo.Id,
		EventType: Active,
		TrustAt:   timePtr(time.Now()),
	}
	require.NoError(t, db.InsertOrUpdateTrustRepo(trustEvent))

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		sponsor.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "USD", day, time.Now(), false,
	))

	sponsors, err := db.GetActiveSponsors(3)
	require.NoError(t, err)
	assert.Contains(t, sponsors, sponsor.Id)
}

func TestFilterActiveUsers(t *testing.T) {
	TruncateAll(db, t)

	activeUser := createTestUser(t, db, "active@example.com")
	inactiveUser := createTestUser(t, db, "inactive@example.com")
	contributor := createTestUser(t, db, "contributor@example.com")
	repo := createTestRepo(t, db, "https://github.com/test/repo")

	day := time.Now().Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		activeUser.Id, contributor.Id, repo.Id,
		big.NewInt(1000), "USD", day, time.Now(), false,
	))

	userIds := []uuid.UUID{activeUser.Id, inactiveUser.Id}
	activeUsers, err := db.FilterActiveUsers(userIds, 3)
	require.NoError(t, err)
	assert.Len(t, activeUsers, 1)
	assert.Contains(t, activeUsers, activeUser.Id)
}

func TestFilterActiveUsers_EmptyList(t *testing.T) {
	TruncateAll(db, t)

	activeUsers, err := db.FilterActiveUsers([]uuid.UUID{}, 3)
	require.NoError(t, err)
	assert.Nil(t, activeUsers)
}

// Helper functions
func createTestUser(t *testing.T, db *DB, email string) *UserDetail {
	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     email,
			Name:      "Test User",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))
	return user
}

func createTestRepo(t *testing.T, db *DB, gitUrl string) *Repo {
	repo := &Repo{
		Id:        uuid.New(),
		GitUrl:    &gitUrl,
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.InsertOrUpdateRepo(repo))
	return repo
}

func timePtr(t time.Time) *time.Time {
	return &t
}