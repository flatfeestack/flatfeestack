package db

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindAllEmails(t *testing.T) {
	TruncateAll(db, t)

	user1 := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "test1@example.com",
			Name:      "Test User 1",
			CreatedAt: time.Now(),
		},
	}
	user2 := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "test2@example.com",
			Name:      "Test User 2",
			CreatedAt: time.Now(),
		},
	}

	require.NoError(t, db.InsertUser(user1))
	require.NoError(t, db.InsertUser(user2))

	emails, err := db.FindAllEmails()
	require.NoError(t, err)
	assert.Len(t, emails, 2)
	assert.Contains(t, emails, "test1@example.com")
	assert.Contains(t, emails, "test2@example.com")
}

func TestFindUserByEmail(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "findme@example.com",
			Name:      "Find Me",
			CreatedAt: time.Now(),
		},
		Seats: 5,
		Freq:  30,
		Multiplier: true,
		MultiplierDailyLimit: 1000,
	}

	require.NoError(t, db.InsertUser(user))

	found, err := db.FindUserByEmail("findme@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, user.Id, found.Id)
	assert.Equal(t, "findme@example.com", found.Email)
	assert.Equal(t, "Find Me", found.Name)
}

func TestFindUserByEmail_NotFound(t *testing.T) {
	TruncateAll(db, t)

	found, err := db.FindUserByEmail("nonexistent@example.com")
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestFindUserById(t *testing.T) {
	TruncateAll(db, t)

	userId := uuid.New()
	user := &UserDetail{
		User: User{
			Id:        userId,
			Email:     "testid@example.com",
			Name:      "Test ID",
			CreatedAt: time.Now(),
		},
	}

	require.NoError(t, db.InsertUser(user))

	found, err := db.FindUserById(userId)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, userId, found.Id)
	assert.Equal(t, "testid@example.com", found.Email)
}

func TestFindUserById_NotFound(t *testing.T) {
	TruncateAll(db, t)

	found, err := db.FindUserById(uuid.New())
	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestFindPublicUserById(t *testing.T) {
	TruncateAll(db, t)

	userId := uuid.New()
	name := "Public User"
	image := "https://example.com/avatar.png"
	user := &UserDetail{
		User: User{
			Id:        userId,
			Email:     "public@example.com",
			Name:      name,
			CreatedAt: time.Now(),
		},
		Image: &image,
	}

	require.NoError(t, db.InsertUser(user))
	require.NoError(t, db.UpdateUserImage(userId, image))

	found, err := db.FindPublicUserById(userId)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, userId, found.Id)
	assert.NotNil(t, found.Name)
	assert.Equal(t, name, *found.Name)
	assert.NotNil(t, found.Image)
	assert.Equal(t, image, *found.Image)
}

func TestInsertFoundation(t *testing.T) {
	TruncateAll(db, t)

	stripeId := "stripe_foundation_123"
	foundation := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "foundation@example.com",
			CreatedAt: time.Now(),
		},
		StripeId: &stripeId,
		Multiplier: true,
		MultiplierDailyLimit: 10000,
	}

	err := db.InsertFoundation(foundation)
	require.NoError(t, err)

	found, err := db.FindUserByEmail("foundation@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.True(t, found.Multiplier)
	assert.Equal(t, 10000, found.MultiplierDailyLimit)
}

func TestUpdateStripe(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "stripe@example.com",
			Name:      "Stripe User",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	stripeId := "stripe_123"
	paymentMethod := "pm_123"
	last4 := "4242"
	user.StripeId = &stripeId
	user.PaymentMethod = &paymentMethod
	user.Last4 = &last4

	err := db.UpdateStripe(user)
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	require.NotNil(t, found.StripeId)
	assert.Equal(t, stripeId, *found.StripeId)
	require.NotNil(t, found.PaymentMethod)
	assert.Equal(t, paymentMethod, *found.PaymentMethod)
	require.NotNil(t, found.Last4)
	assert.Equal(t, last4, *found.Last4)
}

func TestUpdateUserName(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "name@example.com",
			Name:      "Old Name",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	err := db.UpdateUserName(user.Id, "New Name")
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	assert.Equal(t, "New Name", found.Name)
}

func TestUpdateUserImage(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "image@example.com",
			Name:      "Image User",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	imageData := "base64encodedimage"
	err := db.UpdateUserImage(user.Id, imageData)
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	require.NotNil(t, found.Image)
	assert.Equal(t, imageData, *found.Image)
}

func TestDeleteUserImage(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "deleteimg@example.com",
			Name:      "Delete Image",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	imageData := "imagetodeleted"
	require.NoError(t, db.UpdateUserImage(user.Id, imageData))

	err := db.DeleteUserImage(user.Id)
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	assert.Nil(t, found.Image)
}

func TestUpdateSeatsFreq(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "seats@example.com",
			Name:      "Seats User",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	err := db.UpdateSeatsFreq(user.Id, 10, 7)
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	assert.Equal(t, 10, found.Seats)
	assert.Equal(t, 7, found.Freq)
}

func TestFindSponsoredUserBalances(t *testing.T) {
	TruncateAll(db, t)

	inviter := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "inviter@example.com",
			Name:      "Inviter",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(inviter))

	invited1 := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "invited1@example.com",
			Name:      "Invited One",
			CreatedAt: time.Now(),
		},
		InvitedId: &inviter.Id,
	}
	invited2 := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "invited2@example.com",
			Name:      "Invited Two",
			CreatedAt: time.Now(),
		},
		InvitedId: &inviter.Id,
	}

	require.NoError(t, db.InsertUser(invited1))
	require.NoError(t, db.UpdateUserInviteId(invited1.Id, inviter.Id))
	require.NoError(t, db.InsertUser(invited2))
	require.NoError(t, db.UpdateUserInviteId(invited2.Id, inviter.Id))

	users, err := db.FindSponsoredUserBalances(inviter.Id)
	require.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestUpdateUserInviteId(t *testing.T) {
	TruncateAll(db, t)
	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "invite@example.com",
			Name:      "Invite Test",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	inviteId := uuid.New()
	err := db.UpdateUserInviteId(user.Id, inviteId)
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	require.NotNil(t, found.InvitedId)
	assert.Equal(t, inviteId, *found.InvitedId)
}

func TestUpdateClientSecret(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "secret@example.com",
			Name:      "Secret User",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	clientSecret := "secret_123"
	err := db.UpdateClientSecret(user.Id, clientSecret)
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	require.NotNil(t, found.StripeClientSecret)
	assert.Equal(t, clientSecret, *found.StripeClientSecret)
}

func TestUpdateMultiplier(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "multiplier@example.com",
			Name:      "Multiplier User",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	err := db.UpdateMultiplier(user.Id, true)
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	assert.True(t, found.Multiplier)
}

func TestUpdateMultiplierDailyLimit(t *testing.T) {
	TruncateAll(db, t)

	user := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "limit@example.com",
			Name:      "Limit User",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(user))

	err := db.UpdateMultiplierDailyLimit(user.Id, 5000)
	require.NoError(t, err)

	found, err := db.FindUserById(user.Id)
	require.NoError(t, err)
	assert.Equal(t, 5000, found.MultiplierDailyLimit)
}

func TestCheckDailyLimitStillAdheredTo_WithinLimit(t *testing.T) {
	TruncateAll(db, t)

	foundation := &Foundation{
		Id:                   uuid.New(),
		MultiplierDailyLimit: 10000,
	}

	amount := big.NewInt(5000)
	currency := "USD"
	yesterdayStart := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)

	result, err := db.CheckDailyLimitStillAdheredTo(foundation, amount, currency, yesterdayStart)
	require.NoError(t, err)
	assert.Equal(t, amount, result)
}

func TestCheckDailyLimitStillAdheredTo_ExceedsLimit(t *testing.T) {
	TruncateAll(db, t)

	foundationUser := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "foundation@example.com",
			CreatedAt: time.Now(),
		},
		Multiplier:           true,
		MultiplierDailyLimit: 10000,
	}
	require.NoError(t, db.InsertUser(foundationUser))

	contributor := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "contributor@example.com",
			Name:      "Contributor",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(contributor))

	repo := &Repo{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
	}
	gitUrl := "https://github.com/test/repo"
	repo.GitUrl = &gitUrl
	require.NoError(t, db.InsertOrUpdateRepo(repo))

	yesterdayStart := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		foundationUser.Id, contributor.Id, repo.Id,
		big.NewInt(8000), "USD", yesterdayStart, time.Now(), true))

	foundation := &Foundation{
		Id:                   foundationUser.Id,
		MultiplierDailyLimit: 10000,
	}

	result, err := db.CheckDailyLimitStillAdheredTo(foundation, big.NewInt(5000), "USD", yesterdayStart)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(2000), result)
}

func TestCheckDailyLimitStillAdheredTo_LimitReached(t *testing.T) {
	TruncateAll(db, t)

	foundationUser := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "foundation2@example.com",
			CreatedAt: time.Now(),
		},
		Multiplier:           true,
		MultiplierDailyLimit: 10000,
	}
	require.NoError(t, db.InsertUser(foundationUser))

	contributor := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "contributor2@example.com",
			Name:      "Contributor",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(contributor))

	repo := &Repo{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
	}
	gitUrl := "https://github.com/test/repo2"
	repo.GitUrl = &gitUrl
	require.NoError(t, db.InsertOrUpdateRepo(repo))

	yesterdayStart := time.Now().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	require.NoError(t, db.InsertContribution(
		foundationUser.Id, contributor.Id, repo.Id,
		big.NewInt(10000), "USD", yesterdayStart, time.Now(), true))

	foundation := &Foundation{
		Id:                   foundationUser.Id,
		MultiplierDailyLimit: 10000,
	}

	result, err := db.CheckDailyLimitStillAdheredTo(foundation, big.NewInt(1000), "USD", yesterdayStart)
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(-1), result)
}

func TestCheckFondsAmountEnough_SufficientFunds(t *testing.T) {
	TruncateAll(db, t)

	foundationUser := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "foundation3@example.com",
			CreatedAt: time.Now(),
		},
		Multiplier:           true,
		MultiplierDailyLimit: 10000,
	}
	require.NoError(t, db.InsertUser(foundationUser))

	payInEvent := PayInEvent{
		Id:        uuid.New(),
		UserId:    foundationUser.Id,
		Balance:   big.NewInt(50000),
		Currency:  "USD",
		Status:    PayInSuccess,
		Seats:     1,
		Freq:      1,
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent))

	foundation := &Foundation{
		Id:                   foundationUser.Id,
		MultiplierDailyLimit: 10000,
	}

	result, err := db.CheckFondsAmountEnough(foundation, big.NewInt(10000), "USD")
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(10000), result)
}

func TestCheckFondsAmountEnough_InsufficientFunds(t *testing.T) {
	TruncateAll(db, t)

	foundationUser := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "foundation4@example.com",
			CreatedAt: time.Now(),
		},
		Multiplier:           true,
		MultiplierDailyLimit: 10000,
	}
	require.NoError(t, db.InsertUser(foundationUser))

	contributor := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "contributor4@example.com",
			Name:      "Contributor",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(contributor))

	repo := &Repo{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
	}
	gitUrl := "https://github.com/test/repo4"
	repo.GitUrl = &gitUrl
	require.NoError(t, db.InsertOrUpdateRepo(repo))

	payInEvent := PayInEvent{
		Id:        uuid.New(),
		UserId:    foundationUser.Id,
		Balance:   big.NewInt(10000),
		Currency:  "USD",
		Status:    PayInSuccess,
		Seats:     1,
		Freq:      1,
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent))

	require.NoError(t, db.InsertContribution(
		foundationUser.Id, contributor.Id, repo.Id,
		big.NewInt(8000), "USD", time.Now(), time.Now(), true))

	foundation := &Foundation{
		Id:                   foundationUser.Id,
		MultiplierDailyLimit: 10000,
	}

	result, err := db.CheckFondsAmountEnough(foundation, big.NewInt(5000), "USD")
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(2000), result)
}

func TestCheckFondsAmountEnough_NoFunds(t *testing.T) {
	TruncateAll(db, t)

	foundationUser := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "foundation5@example.com",
			CreatedAt: time.Now(),
		},
		Multiplier:           true,
		MultiplierDailyLimit: 10000,
	}
	require.NoError(t, db.InsertUser(foundationUser))

	contributor := &UserDetail{
		User: User{
			Id:        uuid.New(),
			Email:     "contributor5@example.com",
			Name:      "Contributor",
			CreatedAt: time.Now(),
		},
	}
	require.NoError(t, db.InsertUser(contributor))

	repo := &Repo{
		Id:        uuid.New(),
		CreatedAt: time.Now(),
	}
	gitUrl := "https://github.com/test/repo5"
	repo.GitUrl = &gitUrl
	require.NoError(t, db.InsertOrUpdateRepo(repo))

	payInEvent := PayInEvent{
		Id:        uuid.New(),
		UserId:    foundationUser.Id,
		Balance:   big.NewInt(10000),
		Currency:  "USD",
		Status:    PayInSuccess,
		Seats:     1,
		Freq:      1,
		CreatedAt: time.Now(),
	}
	require.NoError(t, db.InsertPayInEvent(payInEvent))

	require.NoError(t, db.InsertContribution(
		foundationUser.Id, contributor.Id, repo.Id,
		big.NewInt(10000), "USD", time.Now(), time.Now(), true))

	foundation := &Foundation{
		Id:                   foundationUser.Id,
		MultiplierDailyLimit: 10000,
	}

	result, err := db.CheckFondsAmountEnough(foundation, big.NewInt(1000), "USD")
	require.NoError(t, err)
	assert.Equal(t, big.NewInt(-1), result)
}