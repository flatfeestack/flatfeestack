package db

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// invite.go tests
func TestInsertInvite(t *testing.T) {
	TruncateAll(db, t)

	id := uuid.New()
	fromEmail := "inviter@example.com"
	toEmail := "invitee@example.com"
	now := time.Now()

	err := db.InsertInvite(id, fromEmail, toEmail, now)
	require.NoError(t, err)
}

func TestFindInvitationsByAnyEmail(t *testing.T) {
	TruncateAll(db, t)

	fromEmail := "sender@example.com"
	toEmail := "receiver@example.com"
	now := time.Now()

	require.NoError(t, db.InsertInvite(uuid.New(), fromEmail, toEmail, now))

	invitations, err := db.FindInvitationsByAnyEmail(fromEmail)
	require.NoError(t, err)
	assert.Len(t, invitations, 1)
	assert.Equal(t, fromEmail, invitations[0].Email)
	assert.Equal(t, toEmail, invitations[0].InviteEmail)

	invitations, err = db.FindInvitationsByAnyEmail(toEmail)
	require.NoError(t, err)
	assert.Len(t, invitations, 1)
}

func TestFindMyInvitations(t *testing.T) {
	TruncateAll(db, t)

	fromEmail := "inviter@example.com"
	toEmail1 := "invitee1@example.com"
	toEmail2 := "invitee2@example.com"
	now := time.Now()

	require.NoError(t, db.InsertInvite(uuid.New(), fromEmail, toEmail1, now))
	require.NoError(t, db.InsertInvite(uuid.New(), fromEmail, toEmail2, now))
	require.NoError(t, db.InsertInvite(uuid.New(), "other@example.com", toEmail1, now))

	invitations, err := db.FindMyInvitations(fromEmail)
	require.NoError(t, err)
	assert.Len(t, invitations, 2)
}

func TestUpdateConfirmInviteAt(t *testing.T) {
	TruncateAll(db, t)

	fromEmail := "inviter@example.com"
	toEmail := "invitee@example.com"
	createdAt := time.Now()

	require.NoError(t, db.InsertInvite(uuid.New(), fromEmail, toEmail, createdAt))

	confirmAt := time.Now().Add(time.Hour)
	err := db.UpdateConfirmInviteAt(fromEmail, toEmail, confirmAt)
	require.NoError(t, err)

	invitations, err := db.FindInvitationsByAnyEmail(fromEmail)
	require.NoError(t, err)
	require.Len(t, invitations, 1)
	require.NotNil(t, invitations[0].ConfirmedAt)
}

func TestDeleteInvite(t *testing.T) {
	TruncateAll(db, t)

	fromEmail := "inviter@example.com"
	toEmail := "invitee@example.com"
	now := time.Now()

	require.NoError(t, db.InsertInvite(uuid.New(), fromEmail, toEmail, now))

	err := db.DeleteInvite(fromEmail, toEmail)
	require.NoError(t, err)

	invitations, err := db.FindInvitationsByAnyEmail(fromEmail)
	require.NoError(t, err)
	assert.Empty(t, invitations)
}

// emails.go tests
func TestInsertGitEmail(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "git@example.com"
	token := "verification_token"

	err := db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now())
	require.NoError(t, err)
}

func TestCountExistingOrConfirmedGitEmail(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "git@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	count, err := db.CountExistingOrConfirmedGitEmail(user.Id, gitEmail)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestCountExistingOrConfirmedGitEmail_NotFound(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")

	count, err := db.CountExistingOrConfirmedGitEmail(user.Id, "nonexistent@example.com")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestConfirmGitEmail(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "git@example.com"
	token := "verification_token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	confirmAt := time.Now().Add(time.Hour)
	err := db.ConfirmGitEmail(gitEmail, token, confirmAt)
	require.NoError(t, err)

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	require.Len(t, emails, 1)
	require.NotNil(t, emails[0].ConfirmedAt)
}

func TestDeleteGitEmail(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "git@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	err := db.DeleteGitEmail(user.Id, gitEmail)
	require.NoError(t, err)

	emails, err := db.FindGitEmailsByUserId(user.Id)
	require.NoError(t, err)
	assert.Empty(t, emails)
}

func TestFindUserByGitEmail(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	gitEmail := "git@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	userId, err := db.FindUserByGitEmail(gitEmail)
	require.NoError(t, err)
	require.NotNil(t, userId)
	assert.Equal(t, user.Id, *userId)
}

func TestFindUsersByGitEmails(t *testing.T) {
	TruncateAll(db, t)

	user1 := createTestUser(t, db, "user1@example.com")
	user2 := createTestUser(t, db, "user2@example.com")
	gitEmail1 := "git1@example.com"
	gitEmail2 := "git2@example.com"
	token := "token"

	require.NoError(t, db.InsertGitEmail(uuid.New(), user1.Id, gitEmail1, &token, time.Now()))
	require.NoError(t, db.InsertGitEmail(uuid.New(), user2.Id, gitEmail2, &token, time.Now()))

	userIds, err := db.FindUsersByGitEmails([]string{gitEmail1, gitEmail2, "nonexistent@example.com"})
	require.NoError(t, err)
	assert.Len(t, userIds, 2)
	assert.Contains(t, userIds, user1.Id)
	assert.Contains(t, userIds, user2.Id)
}

func TestFindUsersByGitEmails_EmptyList(t *testing.T) {
	TruncateAll(db, t)

	userIds, err := db.FindUsersByGitEmails([]string{})
	require.NoError(t, err)
	assert.Nil(t, userIds)
}

func TestInsertEmailSent(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	email := "sent@example.com"
	emailType := "welcome"

	err := db.InsertEmailSent(uuid.New(), &user.Id, email, emailType, time.Now())
	require.NoError(t, err)
}

func TestCountEmailSentById(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	email := "sent@example.com"
	emailType := "reminder"

	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, email, emailType, time.Now()))
	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, email, emailType, time.Now()))

	count, err := db.CountEmailSentById(user.Id, emailType)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestCountEmailSentById_NotFound(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")

	count, err := db.CountEmailSentById(user.Id, "nonexistent-type")
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestCountEmailSentByEmail(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	email := "sent@example.com"
	emailType := "notification"

	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, email, emailType, time.Now()))
	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, email, emailType, time.Now()))

	count, err := db.CountEmailSentByEmail(email, emailType)
	require.NoError(t, err)
	assert.Equal(t, 2, count)
}

func TestInsertUnclaimed(t *testing.T) {
	TruncateAll(db, t)

	email := "unclaimed@example.com"
	repo := createTestRepo(t, db, "https://github.com/test/repo")
	balance := big.NewInt(5000)
	currency := "USD"
	day := time.Now().Truncate(24 * time.Hour)

	err := db.InsertUnclaimed(uuid.New(), email, repo.Id, balance, currency, day, time.Now())
	require.NoError(t, err)
}

func TestFindMarketingEmails(t *testing.T) {
	TruncateAll(db, t)

	email1 := "marketing1@example.com"
	email2 := "marketing2@example.com"
	repo := createTestRepo(t, db, "https://github.com/test/repo")
	day := time.Now().Truncate(24 * time.Hour)

	require.NoError(t, db.InsertUnclaimed(uuid.New(), email1, repo.Id, big.NewInt(1000), "USD", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), email1, repo.Id, big.NewInt(500), "USD", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), email2, repo.Id, big.NewInt(2000), "EUR", day, time.Now()))

	marketingEmails, err := db.FindMarketingEmails()
	require.NoError(t, err)
	assert.Len(t, marketingEmails, 2)

	var email1Found, email2Found bool
	for _, me := range marketingEmails {
		if me.Email == email1 {
			email1Found = true
			assert.Equal(t, big.NewInt(1500), me.Balances["USD"])
		}
		if me.Email == email2 {
			email2Found = true
			assert.Equal(t, big.NewInt(2000), me.Balances["EUR"])
		}
	}
	assert.True(t, email1Found)
	assert.True(t, email2Found)
}

func TestFindMarketingEmails_ExcludesRegisteredUsers(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "registered@example.com")
	gitEmail := "registered@example.com"
	token := "token"
	require.NoError(t, db.InsertGitEmail(uuid.New(), user.Id, gitEmail, &token, time.Now()))

	unregisteredEmail := "unregistered@example.com"
	repo := createTestRepo(t, db, "https://github.com/test/repo")
	day := time.Now().Truncate(24 * time.Hour)

	require.NoError(t, db.InsertUnclaimed(uuid.New(), gitEmail, repo.Id, big.NewInt(1000), "USD", day, time.Now()))
	require.NoError(t, db.InsertUnclaimed(uuid.New(), unregisteredEmail, repo.Id, big.NewInt(2000), "USD", day, time.Now()))

	marketingEmails, err := db.FindMarketingEmails()
	require.NoError(t, err)
	assert.Len(t, marketingEmails, 1)
	assert.Equal(t, unregisteredEmail, marketingEmails[0].Email)
}

func TestDeleteGitEmailFromUserEmailsSent(t *testing.T) {
	TruncateAll(db, t)

	user := createTestUser(t, db, "user@example.com")
	email := "git@example.com"
	emailType := "add-git" + email

	require.NoError(t, db.InsertEmailSent(uuid.New(), &user.Id, email, emailType, time.Now()))

	err := db.DeleteGitEmailFromUserEmailsSent(user.Id, email)
	require.NoError(t, err)

	count, err := db.CountEmailSentById(user.Id, emailType)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}