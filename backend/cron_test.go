package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

var (
	day0  = time.Time{}
	day01 = time.Time{}.Add(time.Duration(1) * time.Second)
	day1  = time.Time{}.Add(time.Duration(1*24) * time.Hour)
	day2  = time.Time{}.Add(time.Duration(2*24) * time.Hour)
	day3  = time.Time{}.Add(time.Duration(3*24) * time.Hour)
	day4  = time.Time{}.Add(time.Duration(4*24) * time.Hour)
)

func TestDailyRunnerOneContributor(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId := setupFunds(t, *sponsors[0], "USD", 365, 120, nil)

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day0, day2, []string{"ste@ste.ste"}, []float64{0.5})

	err := setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)
	err = setupUnSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Error(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day01)
	assert.Nil(t, err)

	err = dailyRunner(day2)
	assert.Nil(t, err)

	//now check the daily_contribution
	m1, err := findSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	assert.Equal(t, m1["USD"].Balance.String(), "328767")

	m2, err := findSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	assert.Equal(t, m2["USD"].Balance.String(), "328767")

	m3, err := findSumDailyContributors(*contributors[1])
	assert.Nil(t, err)
	assert.Equal(t, 0, len(m3))

}

func TestDailyRunnerOneFuture(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId := setupFunds(t, *sponsors[0], "USD", 365, 120, nil)

	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")

	err := setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)

	err = dailyRunner(day2)
	assert.Nil(t, err)

	//now check the daily_contribution

	m1, err := findSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	assert.Equal(t, m1["USD"].Balance.String(), "328767")

}

func TestDailyRunnerThreeContributors(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId1 := setupFunds(t, *sponsors[0], "USD", 365, 120, nil)
	payId2 := setupFunds(t, *sponsors[1], "USD", 365, 120, nil)
	payId3 := setupFunds(t, *sponsors[2], "USD", 365, 120, nil)

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	setupGitEmail(t, *contributors[1], "pea@pea.pea")
	setupGitEmail(t, *contributors[2], "luc@luc.luc")

	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day0, day2, []string{"ste@ste.ste", "pea@pea.pea"}, []float64{0.3, 0.3})
	setupContributor(t, *repos[1], day0, day2, []string{"luc@luc.luc", "pea@pea.pea"}, []float64{0.4, 0.2})

	err := setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)
	err = setupUnSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Error(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day01)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[1], repos[0], day0)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[1], repos[1], day0)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[2], repos[0], day0)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[2], repos[1], day0)
	assert.Nil(t, err)

	err = dailyRunner(day2)
	assert.Nil(t, err)

	//the distribution needs to be as follows:
	//full amount: 328767
	//1/2: 164383
	//1/4: 82191
	//1/8: 41095
	//ste gets from tom 1/2, from mic 1/4, from arm 1/4 for repo tomp2p
	//pea gets from tom 1/2, from mic 1/4, from arm 1/4 for repo tomp2p

	//neow3j repo has 328767 funds, it is diveded by
	//328767 / 0.6 * 0.4 for luc, 328767 / 0.6 * 0.2 for pea
	//pea gets from mic 1/8, from arm 1/8 for repo neow3j
	//luc gets from mic 1/4, from arm 1/4 for repo neow3j

	//now check the daily_contribution
	m1, err := findSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	assert.Equal(t, m1["USD"].Balance.String(), "328765")

	m2, err := findSumDailyContributors(*contributors[1])
	assert.Nil(t, err)
	assert.Equal(t, m2["USD"].Balance.String(), "438353")

	m3, err := findSumDailyContributors(*contributors[2])
	assert.Nil(t, err)
	assert.Equal(t, m3["USD"].Balance.String(), "219176")

	mtot1 := m1["USD"].Balance.Int64() + m2["USD"].Balance.Int64() + m3["USD"].Balance.Int64()

	m4, err := findSumDailySponsors(*sponsors[0], *payId1)
	assert.Nil(t, err)
	assert.Equal(t, m4["USD"].Balance.String(), "328766")

	m5, err := findSumDailySponsors(*sponsors[1], *payId2)
	assert.Nil(t, err)
	assert.Equal(t, m5["USD"].Balance.String(), "328764")

	m6, err := findSumDailySponsors(*sponsors[2], *payId3)
	assert.Nil(t, err)
	assert.Equal(t, m6["USD"].Balance.String(), "328764")

	mtot2 := m4["USD"].Balance.Int64() + m5["USD"].Balance.Int64() + m6["USD"].Balance.Int64()
	assert.Equal(t, mtot1, mtot2)

}

func TestDailyRunnerThreeContributorsTwice(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId1 := setupFunds(t, *sponsors[0], "USD", 365, 120, nil)
	payId2 := setupFunds(t, *sponsors[1], "USD", 365, 120, nil)
	payId3 := setupFunds(t, *sponsors[2], "USD", 365, 120, nil)

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	setupGitEmail(t, *contributors[1], "pea@pea.pea")
	setupGitEmail(t, *contributors[2], "luc@luc.luc")

	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day0, day3, []string{"ste@ste.ste", "pea@pea.pea"}, []float64{0.3, 0.3})
	setupContributor(t, *repos[1], day0, day3, []string{"luc@luc.luc", "pea@pea.pea"}, []float64{0.4, 0.2})

	err := setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)
	err = setupUnSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Error(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day01)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[1], repos[0], day0)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[1], repos[1], day0)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[2], repos[0], day0)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[2], repos[1], day0)
	assert.Nil(t, err)

	err = dailyRunner(day2)
	assert.Nil(t, err)

	err = dailyRunner(day2)
	assert.Nil(t, err)

	err = dailyRunner(day3)
	assert.Nil(t, err)

	err = dailyRunner(day4)
	assert.Nil(t, err)

	//now check the daily_contribution
	m1, err := findSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	assert.Equal(t, m1["USD"].Balance.String(), "657530")

	m2, err := findSumDailyContributors(*contributors[1])
	assert.Nil(t, err)
	assert.Equal(t, m2["USD"].Balance.String(), "876706")

	m3, err := findSumDailyContributors(*contributors[2])
	assert.Nil(t, err)
	assert.Equal(t, m3["USD"].Balance.String(), "438352")

	mtot1 := m1["USD"].Balance.Int64() + m2["USD"].Balance.Int64() + m3["USD"].Balance.Int64()

	m4, err := findSumDailySponsors(*sponsors[0], *payId1)
	assert.Nil(t, err)
	assert.Equal(t, m4["USD"].Balance.String(), "986299")

	m5, err := findSumDailySponsors(*sponsors[1], *payId2)
	assert.Nil(t, err)
	assert.Equal(t, m5["USD"].Balance.String(), "986294")

	m6, err := findSumDailySponsors(*sponsors[2], *payId3)
	assert.Nil(t, err)
	assert.Equal(t, m6["USD"].Balance.String(), "986294")

	mtot2 := m4["USD"].Balance.Int64() + m5["USD"].Balance.Int64() + m6["USD"].Balance.Int64()
	assert.Equal(t, mtot1, mtot2-(3*(328767))+2) //-2 is rounding diff
}

func TestDailyRunnerFutureContribution(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId := setupFunds(t, *sponsors[0], "USD", 365, 120, nil)
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")

	err := setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)

	err = dailyRunner(day2)
	assert.Nil(t, err)

	m2, err := findSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	assert.Equal(t, m2["USD"].Balance.String(), "328767")

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	setupContributor(t, *repos[0], day0, day3, []string{"ste@ste.ste"}, []float64{0.4})

	//this needs to fail, as we already processed this
	err = dailyRunner(day2)
	assert.NotNil(t, err)

	err = dailyRunner(day3)
	assert.Nil(t, err)

	m3, err := findSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	assert.Equal(t, m3["USD"].Balance.String(), "657534")

	m4, err := findSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	assert.Equal(t, m4["USD"].Balance.String(), "657534")
}

func TestDailyRunnerUSDandNEO(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId1 := setupFunds(t, *sponsors[0], "USD", 365, 120, nil)
	payId2 := setupFunds(t, *sponsors[0], "GAS", 360, 200, payId1)
	payId3 := setupFunds(t, *sponsors[1], "GAS", 365, 300, nil)
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")

	err := setupSponsor(t, sponsors[0], repos[0], day0)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[1], repos[0], day0)
	assert.Nil(t, err)

	err = dailyRunner(day2)
	assert.Nil(t, err)

	m2, err := findSumDailySponsors(*sponsors[1], *payId1)
	assert.Nil(t, err)
	assert.Nil(t, m2["USD"])

	m3, err := findSumDailySponsors(*sponsors[0], *payId2)
	assert.Equal(t, m3["USD"].Balance.String(), "328767")

	contributors := setupUsers(t, "ste@ste.ste c1")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	setupContributor(t, *repos[0], day0, day3, []string{"ste@ste.ste"}, []float64{0.4})

	err = dailyRunner(day3)
	assert.Nil(t, err)

	m4, err := findSumDailySponsors(*sponsors[0], *payId2)
	assert.Nil(t, err)
	assert.Equal(t, m4["USD"].Balance.String(), "657534")

	m5, err := findSumDailySponsors(*sponsors[1], *payId3)
	assert.Nil(t, err)
	assert.Equal(t, m5["GAS"].Balance.String(), "164383560")

	m6, err := findSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	assert.Equal(t, m6["GAS"].Balance.String(), "164383560")
	assert.Equal(t, m6["USD"].Balance.String(), "657534")
}

func setupGitEmail(t *testing.T, user uuid.UUID, email string) {
	now := timeNow()
	err := insertGitEmail(user, email, nil, now)
	assert.Nil(t, err)
}

func setupContributor(t *testing.T, repo uuid.UUID, from time.Time, to time.Time, email []string, weight []float64) {
	aid := uuid.New()
	now := timeNow()
	//id uuid.UUID, repo_id uuid.UUID, date_from time.Time, date_to time.Time, branch string
	err := insertAnalysisRequest(aid, repo, from, to, "test", now)
	assert.Nil(t, err)
	for k, v := range email {
		w1 := FlatFeeWeight{
			Email:  v,
			Name:   v,
			Weight: weight[k],
		}
		err = insertAnalysisResponse(aid, &w1, now)
		assert.Nil(t, err)
	}
}

func setupFunds(t *testing.T, uid uuid.UUID, currency string, freq int64, balance int64, oldPaymentCycleId *uuid.UUID) *uuid.UUID {
	pow, err := getFactorInt(currency)
	assert.Nil(t, err)
	paymentCycleId, err := insertNewPaymentCycleIn(1, freq, timeNow())
	assert.Nil(t, err)
	err = paymentSuccess(uid, oldPaymentCycleId, paymentCycleId, big.NewInt(balance*pow), currency, 1, freq, big.NewInt(0))
	assert.Nil(t, err)
	return paymentCycleId
}

func setupUsers(t *testing.T, userNames ...string) []*uuid.UUID {
	var users []*uuid.UUID
	for _, v := range userNames {
		userId1, err := setupUser(v)
		assert.Nil(t, err)
		users = append(users, userId1)
	}
	return users
}

func setupRepos(t *testing.T, repoNames ...string) []*uuid.UUID {
	var repos []*uuid.UUID
	for _, v := range repoNames {
		userId1, err := setupRepo(v)
		assert.Nil(t, err)
		repos = append(repos, userId1)
	}
	return repos
}

func setupSponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day time.Time) error {
	e := SponsorEvent{
		Id:        uuid.New(),
		Uid:       *userId,
		RepoId:    *repoId,
		EventType: Active,
		SponsorAt: day,
	}
	return insertOrUpdateSponsor(&e)
}

func setupUnSponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day time.Time) error {
	e := SponsorEvent{
		Id:          uuid.New(),
		Uid:         *userId,
		RepoId:      *repoId,
		EventType:   Inactive,
		UnSponsorAt: &day,
	}
	return insertOrUpdateSponsor(&e)
}

func setupUser(email string) (*uuid.UUID, error) {
	payOutId := uuid.New()
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		PaymentCycleOutId: &payOutId,
		Email:             email,
	}

	err := insertUser(&u)
	if err != nil {
		return nil, err
	}
	return &u.Id, nil
}

func setupRepo(url string) (*uuid.UUID, error) {
	r := Repo{
		Id:          uuid.New(),
		OrigId:      0,
		Url:         stringPointer(url),
		GitUrl:      stringPointer(url),
		Branch:      stringPointer("main"),
		Source:      stringPointer("github"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
		CreatedAt:   time.Time{},
	}
	id, err := insertOrUpdateRepo(&r)
	if err != nil {
		return nil, err
	}
	return id, nil
}
