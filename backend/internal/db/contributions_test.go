package db

import (
	"backend/pkg/util"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

func TestContributionInsert(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

	userSponsor := insertTestUser(t, "sponsor")
	userContrib := insertTestUser(t, "contrib")
	r := insertTestRepo(t)

	err := InsertContribution(userSponsor.Id, userContrib.Id, r.Id, big.NewInt(2), "XBTC", time.Time{}, time.Time{})
	assert.Nil(t, err)
}

func TestMultiContributionInsert(t *testing.T) {
	util.SetupTestData()
	defer util.TeardownTestData()

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
