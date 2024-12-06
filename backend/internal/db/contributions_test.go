package db

import (
	"math/big"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestContributionInsert(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	userSponsor := insertTestUser(t, "sponsor")
	userContrib := insertTestUser(t, "contrib")
	r := insertTestRepo(t)

	err := InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", time.Time{}, time.Time{})
	assert.Nil(t, err)
}

func TestMultiContributionInsert(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	userSponsor := insertTestUser(t, "sponsor")
	userContrib := insertTestUser(t, "contrib")
	r := insertTestRepo(t)

	err := InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", time.Time{}, time.Time{})
	assert.Nil(t, err)
	err = InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", time.Time{}, time.Time{})
	assert.NotNil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(-3), "XBTC", time.Time{}.Add(1), time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(6), "XBTC", time.Time{}.Add(2), time.Time{})
	assert.Nil(t, err)

	m, err := FindSumDailyBalanceByRepoId(r.Id)
	assert.Nil(t, err)
	assert.Equal(t, big.NewInt(5), m["XBTC"])
}

/*func TestGetFoundationsFromDailyContributions(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	userSponsor := insertTestUser(t, "sponsor")
	userFoundation := insertTestUser(t, "foundation")
	userContrib := insertTestUser(t, "contrib")
	r := insertTestRepo(t)
	r2 := insertTestRepo(t)
	r3 := insertTestRepo(t)
	r4 := insertTestRepo(t)

	currentTime := time.Now()
	dateNowStr := currentTime.Format("2006-01-02")
	dateNow, _ := time.Parse("2006-01-02", dateNowStr)

	err := InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)
	err = InsertContribution(userSponsor.Id, userContrib.Id, r2.Id, big.NewInt(2), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r3.Id, big.NewInt(4), "USD", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r4.Id, big.NewInt(6), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	m1 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          userFoundation.Id,
		RepoId:       r.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	m2 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          userFoundation.Id,
		RepoId:       r2.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	m3 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          userFoundation.Id,
		RepoId:       r3.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	m4 := MultiplierEvent{
		Id:           uuid.New(),
		Uid:          userFoundation.Id,
		RepoId:       r4.Id,
		EventType:    Active,
		MultiplierAt: &t001,
	}

	err = InsertOrUpdateMultiplierRepo(&m1)
	assert.Nil(t, err)

	err = InsertOrUpdateMultiplierRepo(&m2)
	assert.Nil(t, err)

	err = InsertOrUpdateMultiplierRepo(&m3)
	assert.Nil(t, err)

	err = InsertOrUpdateMultiplierRepo(&m4)
	assert.Nil(t, err)

	_, err = GetFoundationsFromDailyContributions(dateNow)

	assert.Nil(t, err)

	// Build the expected result
	// expectedSponsorAmount := big.NewInt(12)
	// expectedResult := []FoundationCurrencyRepos{
	// 	{
	// 		UserId:        userFoundation.Id,
	// 		Currency:      "XBTC",
	// 		SponsorAmount: *expectedSponsorAmount,
	// 		RepoIds:       []uuid.UUID{r.Id},
	// 	},
	// }

	// Assert the result
	//assert.Equal(t, expectedResult, res)

}*/

func TestGetUserDonationReposEmpty(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	userSponsor := insertTestUser(t, "sponsor")

	currentTime := time.Now()
	dateNowStr := currentTime.Format("2006-01-02")
	dateNow, _ := time.Parse("2006-01-02", dateNowStr)

	res, err := GetUserDonationRepos(userSponsor.Id, dateNow)
	assert.Nil(t, err)

	expected := map[uuid.UUID][]UserDonationRepo{}

	assert.Equal(t, expected, res)
}

func TestGetUserDonationReposOneTrusted(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	userSponsor := insertTestUser(t, "sponsor")
	adminSponsor := insertTestUser(t, "admin")
	userContrib := insertTestUser(t, "contrib")
	r := insertTestRepo(t)

	currentTime := time.Now()
	dateNowStr := currentTime.Format("2006-01-02")
	dateNow, _ := time.Parse("2006-01-02", dateNowStr)

	err := InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       adminSponsor.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t1,
	}

	_ = InsertOrUpdateTrustRepo(&tr1)

	res, err := GetUserDonationRepos(userSponsor.Id, dateNow)
	assert.Nil(t, err)

	expected := map[uuid.UUID][]UserDonationRepo{
		userSponsor.Id: {
			{
				Currency:              "XBTC",
				SponsorAmount:         *big.NewInt(2),
				TrustedRepoSelected:   []uuid.UUID{r.Id},
				UntrustedRepoSelected: []uuid.UUID{},
			},
		},
	}

	assert.Equal(t, expected, res)
}

func TestGetUserDonationReposOneUntrusted(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()
	userSponsor := insertTestUser(t, "sponsor")
	userContrib := insertTestUser(t, "contrib")
	r := insertTestRepo(t)

	currentTime := time.Now()
	dateNowStr := currentTime.Format("2006-01-02")
	dateNow, _ := time.Parse("2006-01-02", dateNowStr)

	err := InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	res, err := GetUserDonationRepos(userSponsor.Id, dateNow)
	assert.Nil(t, err)

	expected := map[uuid.UUID][]UserDonationRepo{
		userSponsor.Id: {
			{
				Currency:              "XBTC",
				SponsorAmount:         *big.NewInt(2),
				TrustedRepoSelected:   []uuid.UUID{},
				UntrustedRepoSelected: []uuid.UUID{r.Id},
			},
		},
	}

	assert.Equal(t, expected, res)
}

func TestGetUserDonationReposMany(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	userSponsor := insertTestUser(t, "sponsor")
	adminSponsor := insertTestUser(t, "admin")
	userContrib := insertTestUser(t, "contrib")
	userContrib2 := insertTestUser(t, "contrib2")
	r := insertTestRepo(t)
	r2 := insertTestRepo(t)
	r3 := insertTestRepo(t)
	r4 := insertTestRepo(t)

	currentTime := time.Now()
	dateNowStr := currentTime.Format("2006-01-02")
	dateNow, _ := time.Parse("2006-01-02", dateNowStr)

	err := InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r2.Id, big.NewInt(2), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib2.Id, r2.Id, big.NewInt(7), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r3.Id, big.NewInt(4), "USD", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib2.Id, r3.Id, big.NewInt(10), "USD", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r4.Id, big.NewInt(6), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       adminSponsor.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t1,
	}

	tr2 := TrustEvent{
		Id:        uuid.New(),
		Uid:       adminSponsor.Id,
		RepoId:    r4.Id,
		EventType: Active,
		TrustAt:   &t2,
	}

	_ = InsertOrUpdateTrustRepo(&tr1)
	_ = InsertOrUpdateTrustRepo(&tr2)

	res, err := GetUserDonationRepos(userSponsor.Id, dateNow)
	assert.Nil(t, err)

	expected := map[uuid.UUID][]UserDonationRepo{
		userSponsor.Id: {
			{
				Currency:              "XBTC",
				SponsorAmount:         *big.NewInt(17),
				TrustedRepoSelected:   []uuid.UUID{r.Id, r4.Id},
				UntrustedRepoSelected: []uuid.UUID{r2.Id},
			},
			{
				Currency:              "USD",
				SponsorAmount:         *big.NewInt(14),
				TrustedRepoSelected:   []uuid.UUID{},
				UntrustedRepoSelected: []uuid.UUID{r3.Id},
			},
		},
	}

	assert.Equal(t, expected, res)
}

func TestGetUserDonationReposManyDynamic(t *testing.T) {
	SetupTestData()
	defer TeardownTestData()

	userSponsor := insertTestUser(t, "sponsor")
	adminSponsor := insertTestUser(t, "admin")
	userContrib := insertTestUser(t, "contrib")
	r := insertTestRepo(t)
	r2 := insertTestRepo(t)
	r3 := insertTestRepo(t)
	r4 := insertTestRepo(t)

	currentTime := time.Now()
	dateNowStr := currentTime.Format("2006-01-02")
	dateNow, _ := time.Parse("2006-01-02", dateNowStr)

	err := InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)
	err = InsertContribution(userSponsor.Id, userContrib.Id, r2.Id, big.NewInt(2), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r3.Id, big.NewInt(4), "USD", dateNow, time.Time{})
	assert.Nil(t, err)

	err = InsertContribution(userSponsor.Id, userContrib.Id, r4.Id, big.NewInt(6), "XBTC", dateNow, time.Time{})
	assert.Nil(t, err)

	tr1 := TrustEvent{
		Id:        uuid.New(),
		Uid:       adminSponsor.Id,
		RepoId:    r.Id,
		EventType: Active,
		TrustAt:   &t1,
	}

	idTrustEvent := uuid.New()

	tr2 := TrustEvent{
		Id:        idTrustEvent,
		Uid:       adminSponsor.Id,
		RepoId:    r4.Id,
		EventType: Active,
		TrustAt:   &t2,
	}

	_ = InsertOrUpdateTrustRepo(&tr1)
	_ = InsertOrUpdateTrustRepo(&tr2)

	res, err := GetUserDonationRepos(userSponsor.Id, dateNow)
	assert.Nil(t, err)

	expected := map[uuid.UUID][]UserDonationRepo{
		userSponsor.Id: {
			{
				Currency:              "XBTC",
				SponsorAmount:         *big.NewInt(10),
				TrustedRepoSelected:   []uuid.UUID{r.Id, r4.Id},
				UntrustedRepoSelected: []uuid.UUID{r2.Id},
			},
			{
				Currency:              "USD",
				SponsorAmount:         *big.NewInt(4),
				TrustedRepoSelected:   []uuid.UUID{},
				UntrustedRepoSelected: []uuid.UUID{r3.Id},
			},
		},
	}

	assert.Equal(t, expected, res)

	tr3 := TrustEvent{
		Id:        idTrustEvent,
		Uid:       adminSponsor.Id,
		RepoId:    r4.Id,
		EventType: Inactive,
		UnTrustAt: &t3,
	}

	_ = InsertOrUpdateTrustRepo(&tr3)

	res, err = GetUserDonationRepos(userSponsor.Id, dateNow)
	assert.Nil(t, err)

	expected = map[uuid.UUID][]UserDonationRepo{
		userSponsor.Id: {
			{
				Currency:              "XBTC",
				SponsorAmount:         *big.NewInt(10),
				TrustedRepoSelected:   []uuid.UUID{r.Id},
				UntrustedRepoSelected: []uuid.UUID{r2.Id, r4.Id},
			},
			{
				Currency:              "USD",
				SponsorAmount:         *big.NewInt(4),
				TrustedRepoSelected:   []uuid.UUID{},
				UntrustedRepoSelected: []uuid.UUID{r3.Id},
			},
		},
	}

	assert.Equal(t, expected, res)
}
