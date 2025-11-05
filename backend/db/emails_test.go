package db

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindGitEmailsByUserId(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail1 := "git1@example.com"
	gitEmail2 := "git2@example.com"
	token1 := "token1"
	token2 := "token2"
	now := time.Now()

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail1, &token1, now))
	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail2, &token2, now))

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	assert.Len(t, emails, 2)

	emailAddresses := []string{emails[0].Email, emails[1].Email}
	assert.Contains(t, emailAddresses, gitEmail1)
	assert.Contains(t, emailAddresses, gitEmail2)
}

func TestFindGitEmailsByUserId_EmptyResult(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	assert.Empty(t, emails)
}

func TestCountExistingOrConfirmedGitEmail_ExistingUnconfirmed(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "git@example.com"
	token := "token123"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	count, err := db.CountExistingOrConfirmedGitEmail(user.Id, gitEmail)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCountExistingOrConfirmedGitEmail_Confirmed(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "git@example.com"
	token := "token123"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))
	require.NoError(t, db.ConfirmGitEmail(gitEmail, token, time.Now()))

	count, err := db.CountExistingOrConfirmedGitEmail(user.Id, gitEmail)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCountExistingOrConfirmedGitEmail_ConfirmedByOtherUser(t *testing.T) {
	TruncateAll(db, t)

	user1 := createTestUser(t, db, "user1@example.com")
	user2 := createTestUser(t, db, "user2@example.com")
	gitEmail := "shared@example.com"
	token := "token123"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user1.Id, gitEmail, &token, time.Now()))
	require.NoError(t, db.ConfirmGitEmail(gitEmail, token, time.Now()))

	count, err := db.CountExistingOrConfirmedGitEmail(user2.Id, gitEmail)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestInsertGitEmail_WithToken(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "newgit@example.com"
	token := "verification_token_123"
	now := time.Now()

	err := db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, now)
	require.NoError(t, err)

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	assert.Len(t, emails, 1)
	assert.Equal(t, gitEmail, emails[0].Email)
	assert.Nil(t, emails[0].ConfirmedAt)
}

func TestInsertGitEmail_WithoutToken(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "notokenneeded@example.com"
	now := time.Now()

	err := db.InsertGitEmail(uuid.New(), user.Id, gitEmail, nil, now)
	require.NoError(t, err)

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	assert.Len(t, emails, 1)
}

func TestConfirmGitEmail_Success(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "confirm@example.com"
	token := "token_to_confirm"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	confirmAt := time.Now().Add(time.Hour)
	err := db.ConfirmGitEmail(gitEmail, token, confirmAt)
	require.NoError(t, err)

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	require.Len(t, emails, 1)
	require.NotNil(t, emails[0].ConfirmedAt)
	assert.WithinDuration(t, confirmAt, *emails[0].ConfirmedAt, time.Second)
}

func TestConfirmGitEmail_WrongToken(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "confirm@example.com"
	token := "correct_token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	err := db.ConfirmGitEmail(gitEmail, "wrong_token", time.Now())
	require.NoError(t, err)

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	require.Len(t, emails, 1)
	assert.Nil(t, emails[0].ConfirmedAt)
}

func TestDeleteGitEmail_Success(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail1 := "keep@example.com"
	gitEmail2 := "delete@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail1, &token, time.Now()))
	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail2, &token, time.Now()))

	err := db.DeleteGitEmail(user.Id, gitEmail2)
	require.NoError(t, err)

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	assert.Len(t, emails, 1)
	assert.Equal(t, gitEmail1, emails[0].Email)
}

func TestDeleteGitEmailFromUserEmailsSent_Success(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "git@example.com"
	emailType := "add-git" + gitEmail

	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, gitEmail, emailType, time.Now()))

	count, err := db.CountEmailSentById(user.Id, emailType)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	err = db.DeleteGitEmailFromUserEmailsSent(user.Id, gitEmail)
	require.NoError(t, err)

	count, err = db.CountEmailSentById(user.Id, emailType)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestFindUserByGitEmail_Found(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "unique@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	foundUserId, err := db.FindUserByGitEmail(gitEmail)
	require.NoError(t, err)
	require.NotNil(t, foundUserId)
	assert.Equal(t, user.Id, *foundUserId)
}

func TestFindUserByGitEmail_NotFound(t *testing.T) {
	TruncateAll(db, t)

	foundUserId, err := db.FindUserByGitEmail("nonexistent@example.com")
	require.NoError(t, err)
	assert.Nil(t, foundUserId)
}

func TestFindUsersByGitEmails_MultipleUsers(t *testing.T) {
	TruncateAll(db, t)

	user1 := createTestUser(t, db, "user1@example.com")
	user2 := createTestUser(t, db, "user2@example.com")
	user3 := createTestUser(t, db, "user3@example.com")

	gitEmail1 := "git1@example.com"
	gitEmail2 := "git2@example.com"
	gitEmail3 := "git3@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user1.Id, gitEmail1, &token, time.Now()))
	require.NoError(t, db.InsertGitEmail(uuid.New(), user2.Id, gitEmail2, &token, time.Now()))
	require.NoError(t, db.InsertGitEmail(uuid.New(), user3.Id, gitEmail3, &token, time.Now()))

	userIds, err := db.FindUsersByGitEmails([]string{gitEmail1, gitEmail3, "nonexistent@example.com"})
	require.NoError(t, err)
	assert.Len(t, userIds, 2)
	assert.Contains(t, userIds, user1.Id)
	assert.Contains(t, userIds, user3.Id)
	assert.NotContains(t, userIds, user2.Id)
}

func TestFindUsersByGitEmails_EmptyInput(t *testing.T) {
	TruncateAll(db, t)

	userIds, err := db.FindUsersByGitEmails([]string{})
	require.NoError(t, err)
	assert.Nil(t, userIds)
}

func TestFindUsersByGitEmails_SameUserMultipleEmails(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail1 := "git1@example.com"
	gitEmail2 := "git2@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail1, &token, time.Now()))
	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail2, &token, time.Now()))

	userIds, err := db.FindUsersByGitEmails([]string{gitEmail1, gitEmail2})
	require.NoError(t, err)
	assert.Len(t, userIds, 1)
	assert.Equal(t, user.Id, userIds[0])
}

func TestGetRepoEmails(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/repo")

	analysisRequest := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repo.Id,
		DateFrom: time.Now().AddDate(0, 0, -30),
		DateTo:   time.Now(),
		GitUrl:   "https://github.com/test/repo",
	}
	require.NoError(t, db.InsertAnalysisRequest(analysisRequest, time.Now()))

	email1 := "contributor1@example.com"
	email2 := "contributor2@example.com"
	names := []string{"Contributor"}

	require.NoError(t, db.InsertAnalysisResponse(analysisRequest.Id, repo.Id, email1, names, 0.5, time.Now()))
	require.NoError(t, db.InsertAnalysisResponse(analysisRequest.Id, repo.Id, email2, names, 0.5, time.Now()))
	require.NoError(t, db.InsertAnalysisResponse(analysisRequest.Id, repo.Id, email1, names, 0.3, time.Now()))

	emails, err := db.GetRepoEmails(repo.Id)
	require.NoError(t, err)
	assert.Len(t, emails, 2)
	assert.Contains(t, emails, email1)
	assert.Contains(t, emails, email2)
}

func TestGetRepoEmails_NoEmails(t *testing.T) {
	TruncateAll(db, t)

	repo := createTestRepo(t, db, "https://github.com/test/empty-repo")

	emails, err := db.GetRepoEmails(repo.Id)
	require.NoError(t, err)
	assert.Empty(t, emails)
}

func TestInsertEmailSent_WithUserId(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	email := "notification@example.com"
	emailType := "welcome"
	now := time.Now()

	err := db.InsertEmailSent(uuid.New(), &user.Id, email, emailType, now)
	require.NoError(t, err)

	count, err := db.CountEmailSentById(user.Id, emailType)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestInsertEmailSent_WithoutUserId(t *testing.T) {
	TruncateAll(db, t)

	email := "anonymous@example.com"
	emailType := "marketing"
	now := time.Now()

	err := db.InsertEmailSent(uuid.New(), nil, email, emailType, now)
	require.NoError(t, err)

	count, err := db.CountEmailSentByEmail(email, emailType)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCountEmailSentById_MultipleEmails(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	email := "test@example.com"
	emailType1 := "reminder"
	emailType2 := "notification"

	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, email, emailType1, time.Now()))
	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, email, emailType1, time.Now()))
	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, email, emailType2, time.Now()))

	count1, err := db.CountEmailSentById(user.Id, emailType1)
	require.NoError(t, err)
	assert.Equal(t, 2, count1)

	count2, err := db.CountEmailSentById(user.Id, emailType2)
	require.NoError(t, err)
	assert.Equal(t, 1, count2)
}

func TestCountEmailSentByEmail_MultipleTypes(t *testing.T) {
	TruncateAll(db, t)

	email := "test@example.com"
	emailType1 := "promo"
	emailType2 := "update"

	require.NoError(t, db.InsertEmailSent(uuid.New(), nil, email, emailType1, time.Now()))
	require.NoError(t, db.InsertEmailSent(uuid.New(), nil, email, emailType1, time.Now()))
	require.NoError(t, db.InsertEmailSent(uuid.New(), nil, email, emailType1, time.Now()))
	require.NoError(t, db.InsertEmailSent(uuid.New(), nil, email, emailType2, time.Now()))

	count1, err := db.CountEmailSentByEmail(email, emailType1)
	require.NoError(t, err)
	assert.Equal(t, 3, count1)

	count2, err := db.CountEmailSentByEmail(email, emailType2)
	require.NoError(t, err)
	assert.Equal(t, 1, count2)
}

func TestInsertUnclaimed_Success(t *testing.T) {
	TruncateAll(db, t)

	email := "unclaimed@example.com"
	repo := createTestRepo(t, db, "https://github.com/test/repo")
	balance := big.NewInt(10000)
	currency := "USD"
	day := time.Now().Truncate(24 * time.Hour)
	now := time.Now()

	err := db.InsertUnclaimed(uuid.New(), email, repo.Id, balance, currency, day, now)
	require.NoError(t, err)
}

func TestFindMarketingEmails_SingleCurrency(t *testing.T) {
	TruncateAll(db, t)

	email := "marketing@example.com"
	repo := createTestRepo(t, db, "https://github.com/test/repo")
	day := time.Now().Truncate(24 * time.Hour)

	require.NoError(t, db.InsertUnclaimed(uuid.New(), email, repo.Id, big.NewInt(1000), "USD", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), email, repo.Id, big.NewInt(500), "USD", day, time.Now()))

	marketingEmails, err := db.FindMarketingEmails()
	require.NoError(t, err)
	assert.Len(t, marketingEmails, 1)
	assert.Equal(t, email, marketingEmails[0].Email)
	assert.Equal(t, big.NewInt(1500), marketingEmails[0].Balances["USD"])
}

func TestFindMarketingEmails_MultipleCurrencies(t *testing.T) {
	TruncateAll(db, t)

	email := "multicurrency@example.com"
	repo := createTestRepo(t, db, "https://github.com/test/repo")
	day := time.Now().Truncate(24 * time.Hour)

	require.NoError(t, db.InsertUnclaimed(uuid.New(), email, repo.Id, big.NewInt(1000), "USD", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), email, repo.Id, big.NewInt(2000), "EUR", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), email, repo.Id, big.NewInt(3000), "GBP", day, time.Now()))

	marketingEmails, err := db.FindMarketingEmails()
	require.NoError(t, err)
	assert.Len(t, marketingEmails, 1)
	assert.Equal(t, big.NewInt(1000), marketingEmails[0].Balances["USD"])
	assert.Equal(t, big.NewInt(2000), marketingEmails[0].Balances["EUR"])
	assert.Equal(t, big.NewInt(3000), marketingEmails[0].Balances["GBP"])
}

func TestFindMarketingEmails_MultipleEmails(t *testing.T) {
	TruncateAll(db, t)

	email1 := "marketing1@example.com"
	email2 := "marketing2@example.com"
	email3 := "marketing3@example.com"
	repo := createTestRepo(t, db, "https://github.com/test/repo")
	day := time.Now().Truncate(24 * time.Hour)

	require.NoError(t, db.InsertUnclaimed(uuid.New(), email1, repo.Id, big.NewInt(1000), "USD", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), email2, repo.Id, big.NewInt(2000), "EUR", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), email3, repo.Id, big.NewInt(3000), "GBP", day, time.Now()))

	marketingEmails, err := db.FindMarketingEmails()
	require.NoError(t, err)
	assert.Len(t, marketingEmails, 3)

	emailMap := make(map[string]Marketing)
	for _, me := range marketingEmails {
		emailMap[me.Email] = me
	}

	assert.Contains(t, emailMap, email1)
	assert.Contains(t, emailMap, email2)
	assert.Contains(t, emailMap, email3)
}

func TestFindMarketingEmails_ExcludesRegisteredGitEmails(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "registered@example.com")
	registeredEmail := "registered@example.com"
	unregisteredEmail := "unregistered@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, registeredEmail, &token, time.Now()))

	repo := createTestRepo(t, db, "https://github.com/test/repo")
	day := time.Now().Truncate(24 * time.Hour)

	require.NoError(t, db.InsertUnclaimed(uuid.New(), registeredEmail, repo.Id, big.NewInt(1000), "USD", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), unregisteredEmail, repo.Id, big.NewInt(2000), "USD", day, time.Now()))

	marketingEmails, err := db.FindMarketingEmails()
	require.NoError(t, err)
	assert.Len(t, marketingEmails, 1)
	assert.Equal(t, unregisteredEmail, marketingEmails[0].Email)
	assert.Equal(t, big.NewInt(2000), marketingEmails[0].Balances["USD"])
}

func TestFindMarketingEmails_NoUnclaimedBalances(t *testing.T) {
	TruncateAll(db, t)

	marketingEmails, err := db.FindMarketingEmails()
	require.NoError(t, err)
	assert.Empty(t, marketingEmails)
}