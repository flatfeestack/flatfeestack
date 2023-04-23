package main

import (
	"backend/api"
	"backend/clients"
	db "backend/db"
	"backend/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

var (
	day1  = time.Time{}
	day11 = time.Time{}.Add(time.Duration(1) * time.Second)
	day2  = time.Time{}.Add(time.Duration(1*24) * time.Hour)
	day3  = time.Time{}.Add(time.Duration(2*24) * time.Hour)
	day4  = time.Time{}.Add(time.Duration(3*24) * time.Hour)
	day5  = time.Time{}.Add(time.Duration(4*24) * time.Hour)
)

func TestDailyRunnerOneContributor(t *testing.T) {
	setup()
	defer teardown()

	//we have 5 sponsors, but only the first sponsor added funds
	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId := setupFunds(t, *sponsors[0], "USD", 365, 120, nil, day1)

	//we have 4 contributors, 1 has added a git email
	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")

	//we have 4 repos, ste is contribution with 0.5 to tomp2p
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day1, day3, []string{"ste@ste.ste"}, []float64{0.5})

	//tom is sponsoring the repo tomp2p at time 0
	err := setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)
	//tom is unsponsoring the repo tomp2p at time 0, no error should occur
	err = setupUnSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)
	//tom is sponsoring again the repo at time 0, no error should occur
	err = setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Error(t, err)
	//tom is sponsoring for the first second
	err = setupSponsor(t, sponsors[0], repos[0], day11)
	assert.Nil(t, err)

	//run the daily runner for the second day, so that means we have a full day 2 that needs to be processed
	err = dailyRunner(day3)
	assert.Nil(t, err)

	//now check the daily_contribution
	m1, err := db.FindSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	//120 / 365 * 1'000'000
	assert.Equal(t, m1["USD"].Balance.String(), "328767")

	m2, err := db.FindSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	assert.Equal(t, m2["USD"].Balance.String(), "328767")

	m3, err := db.FindSumDailyContributors(*contributors[1])
	assert.Nil(t, err)
	assert.Equal(t, 0, len(m3))

}

func TestDailyRunnerOneContributorLowFunds(t *testing.T) {
	setup()
	defer teardown()

	//we have 5 sponsors, but only the first sponsor added funds
	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	setupFunds(t, *sponsors[0], "USD", 2, 2, nil, day1)

	// we have 4 contributors
	setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")

	// we have 4 repos, ste is contribution with 0.5 to tomp2p
	// ste@ste.ch will receive an invitation mail
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day1, day3, []string{"ste@ste.ste"}, []float64{0.5})

	//tom is sponsoring the repo tomp2p at time 0
	err := setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)
	//tom is unsponsoring the repo tomp2p at time 0, no error should occur
	err = setupUnSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)
	//tom is sponsoring again the repo at time 0, no error should occur
	err = setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Error(t, err)
	//tom is sponsoring for the first second
	err = setupSponsor(t, sponsors[0], repos[0], day11)
	assert.Nil(t, err)

	//run the daily runner for the second day, so that means we have a full day 2 that needs to be processed
	err = dailyRunner(day3)
	assert.Nil(t, err)
	//check send
	assert.Equal(t, 1, clients.EmailNotifications)
}

func TestDailyRunnerOneFuture(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId := setupFunds(t, *sponsors[0], "USD", 365, 120, nil, day1)

	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day1, day1, []string{"hello@example.com"}, []float64{1.0})

	err := setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)

	err = dailyRunner(day2)
	assert.Nil(t, err)

	//now check the daily_contribution

	m1, err := db.FindSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	if assert.NotEmpty(t, m1) {
		assert.Equal(t, m1["USD"].Balance.String(), "328767")
	}
}

func TestDailyRunnerThreeContributors(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId1 := setupFunds(t, *sponsors[0], "USD", 365, 120, nil, day1)
	payId2 := setupFunds(t, *sponsors[1], "USD", 365, 120, nil, day1)
	payId3 := setupFunds(t, *sponsors[2], "USD", 365, 120, nil, day1)

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	setupGitEmail(t, *contributors[1], "pea@pea.pea")
	setupGitEmail(t, *contributors[2], "luc@luc.luc")

	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day1, day2, []string{"ste@ste.ste", "pea@pea.pea"}, []float64{0.3, 0.3})
	setupContributor(t, *repos[1], day1, day2, []string{"luc@luc.luc", "pea@pea.pea"}, []float64{0.4, 0.2})

	err := setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)
	err = setupUnSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Error(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day11)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[1], repos[0], day1)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[1], repos[1], day1)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[2], repos[0], day1)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[2], repos[1], day1)
	assert.Nil(t, err)

	err = dailyRunner(day3)
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
	m1, err := db.FindSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	assert.Equal(t, m1["USD"].Balance.String(), "328765")

	m2, err := db.FindSumDailyContributors(*contributors[1])
	assert.Nil(t, err)
	assert.Equal(t, m2["USD"].Balance.String(), "438353")

	m3, err := db.FindSumDailyContributors(*contributors[2])
	assert.Nil(t, err)
	assert.Equal(t, m3["USD"].Balance.String(), "219176")

	mtot1 := m1["USD"].Balance.Int64() + m2["USD"].Balance.Int64() + m3["USD"].Balance.Int64()

	m4, err := db.FindSumDailySponsors(*sponsors[0], *payId1)
	assert.Nil(t, err)
	assert.Equal(t, m4["USD"].Balance.String(), "328766")

	m5, err := db.FindSumDailySponsors(*sponsors[1], *payId2)
	assert.Nil(t, err)
	assert.Equal(t, m5["USD"].Balance.String(), "328764")

	m6, err := db.FindSumDailySponsors(*sponsors[2], *payId3)
	assert.Nil(t, err)
	assert.Equal(t, m6["USD"].Balance.String(), "328764")

	mtot2 := m4["USD"].Balance.Int64() + m5["USD"].Balance.Int64() + m6["USD"].Balance.Int64()
	assert.Equal(t, mtot1, mtot2)

}

func TestDailyRunnerThreeContributorsTwice(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId1 := setupFunds(t, *sponsors[0], "USD", 365, 120, nil, day1)
	payId2 := setupFunds(t, *sponsors[1], "USD", 365, 120, nil, day1)
	payId3 := setupFunds(t, *sponsors[2], "USD", 365, 120, nil, day1)

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	setupGitEmail(t, *contributors[1], "pea@pea.pea")
	setupGitEmail(t, *contributors[2], "luc@luc.luc")

	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day1, day3, []string{"ste@ste.ste", "pea@pea.pea"}, []float64{0.3, 0.3})
	setupContributor(t, *repos[1], day1, day3, []string{"luc@luc.luc", "pea@pea.pea"}, []float64{0.4, 0.2})

	err := setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)
	err = setupUnSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Error(t, err)
	err = setupSponsor(t, sponsors[0], repos[0], day11)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[1], repos[0], day1)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[1], repos[1], day1)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[2], repos[0], day1)
	assert.Nil(t, err)
	err = setupSponsor(t, sponsors[2], repos[1], day1)
	assert.Nil(t, err)

	err = dailyRunner(day3)
	assert.Nil(t, err)

	err = dailyRunner(day4)
	assert.Nil(t, err)

	err = dailyRunner(day5)
	assert.Nil(t, err)

	//now check the daily_contribution
	_, err = db.FindSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	// assert.Equal(t, "657530", m1["USD"].Balance.String())

	_, err = db.FindSumDailyContributors(*contributors[1])
	assert.Nil(t, err)
	// assert.Equal(t, "876706", m2["USD"].Balance.String())

	_, err = db.FindSumDailyContributors(*contributors[2])
	assert.Nil(t, err)
	// assert.Equal(t, "438352", m3["USD"].Balance.String())

	// mtot1 := m1["USD"].Balance.Int64() + m2["USD"].Balance.Int64() + m3["USD"].Balance.Int64()

	_, err = db.FindSumDailySponsors(*sponsors[0], *payId1)
	assert.Nil(t, err)
	// assert.Equal(t, "986299", m4["USD"].Balance.String())

	_, err = db.FindSumDailySponsors(*sponsors[1], *payId2)
	assert.Nil(t, err)
	// assert.Equal(t, "986294", m5["USD"].Balance.String())

	_, err = db.FindSumDailySponsors(*sponsors[2], *payId3)
	assert.Nil(t, err)
	// assert.Equal(t, "986294", m6["USD"].Balance.String())

	// mtot2 := m4["USD"].Balance.Int64() + m5["USD"].Balance.Int64() + m6["USD"].Balance.Int64()
	// assert.Equal(t, mtot2-(3*(328767))+2, mtot1) //-2 is rounding diff
}

func TestDailyRunnerFutureContribution(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId := setupFunds(t, *sponsors[0], "USD", 365, 120, nil, day1)
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day1, day1, []string{"hello@example.com"}, []float64{1.0})

	err := setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)

	err = dailyRunner(day3)
	assert.Nil(t, err)

	m2, err := db.FindSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	if assert.NotEmpty(t, m2) {
		assert.Equal(t, m2["USD"].Balance.String(), "328767")
	}

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	setupContributor(t, *repos[0], day1, day4, []string{"ste@ste.ste"}, []float64{0.4})

	//this needs to fail, as we already processed this
	err = dailyRunner(day3)
	assert.NotNil(t, err)

	err = dailyRunner(day4)
	assert.Nil(t, err)

	m3, err := db.FindSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	if assert.NotEmpty(t, m3) {
		assert.Equal(t, m3["USD"].Balance.String(), "657534")
	}

	m4, err := db.FindSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	if assert.NotEmpty(t, m4) {
		assert.Equal(t, m4["USD"].Balance.String(), "657534")
	}
}

func TestDailyRunnerUSDandNEO(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId1 := setupFunds(t, *sponsors[0], "USD", 365, 120, nil, day1)
	payId2 := setupFunds(t, *sponsors[0], "GAS", 360, 200, payId1, day1)
	payId3 := setupFunds(t, *sponsors[1], "GAS", 365, 300, nil, day1)
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")

	contributors := setupUsers(t, "ste@ste.ste c1")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	setupContributor(t, *repos[0], day1, day4, []string{"ste@ste.ste"}, []float64{0.4})

	err := setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[1], repos[0], day1)
	assert.Nil(t, err)

	err = dailyRunner(day3)
	assert.Nil(t, err)

	m2, err := db.FindSumDailySponsors(*sponsors[1], *payId1)
	assert.Nil(t, err)
	assert.Nil(t, m2["USD"])

	m3, err := db.FindSumDailySponsors(*sponsors[0], *payId2)
	// no contributor registered so far
	if assert.NotEmpty(t, m3) {
		assert.Equal(t, m3["USD"].Balance.String(), "328767")
	}

	err = dailyRunner(day4)
	assert.Nil(t, err)

	m4, err := db.FindSumDailySponsors(*sponsors[0], *payId2)
	assert.Nil(t, err)
	if assert.NotEmpty(t, m4) {
		assert.Equal(t, "657534", m4["USD"].Balance.String())
	}

	m5, err := db.FindSumDailySponsors(*sponsors[1], *payId3)
	assert.Nil(t, err)
	if assert.NotEmpty(t, m5) {
		assert.Equal(t, "164383560", m5["GAS"].Balance.String())
	}

	m6, err := db.FindSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	if assert.NotEmpty(t, m6) {
		assert.Equal(t, "164383560", m6["GAS"].Balance.String())
		assert.Equal(t, "657534", m6["USD"].Balance.String())
	}
}

func TestDailyRunnerSponsor(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	payId := setupFunds(t, *sponsors[0], "USD", 365, 120, nil, day1)
	setupSponsorUser(t, *sponsors[0], *sponsors[1])

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day1, day3, []string{"ste@ste.ste"}, []float64{0.5})
	setupContributor(t, *repos[1], day1, day3, []string{"ste@ste.ste", "pea@pea.pea"}, []float64{0.5, 0.5})

	err := setupSponsor(t, sponsors[0], repos[0], day1)
	assert.Nil(t, err)

	err = setupSponsor(t, sponsors[1], repos[1], day1)
	assert.Nil(t, err)

	err = dailyRunner(day3)
	assert.Nil(t, err)

	//now check the daily_contribution
	m1, err := db.FindSumDailyContributors(*contributors[0])
	assert.Nil(t, err)
	if assert.NotEmpty(t, m1) {
		assert.Equal(t, "657534", m1["USD"].Balance.String())
	}

	m2, err := db.FindSumDailySponsors(*sponsors[0], *payId)
	assert.Nil(t, err)
	if assert.NotEmpty(t, m2) {
		assert.Equal(t, "657534", m2["USD"].Balance.String())
	}

	m3, err := db.FindSumDailySponsors(*sponsors[1], *payId)
	assert.Nil(t, err)
	assert.Empty(t, m3)
}

func setupSponsorUser(t *testing.T, uid1 uuid.UUID, uid2 uuid.UUID) {
	err := db.UpdateUserInviteId(uid2, uid1)
	assert.Nil(t, err)
}

func setupGitEmail(t *testing.T, user uuid.UUID, email string) {
	now := utils.TimeNow()
	err := db.InsertGitEmail(user, email, nil, now)
	assert.Nil(t, err)
}

func setupContributor(t *testing.T, repoId uuid.UUID, from time.Time, to time.Time, email []string, weight []float64) {
	now := utils.TimeNow()
	//id uuid.UUID, repo_id uuid.UUID, date_from time.Time, date_to time.Time, branch string
	a := db.AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   repoId,
		DateFrom: from,
		DateTo:   to,
		GitUrl:   "test",
	}
	err := db.InsertAnalysisRequest(a, now)
	assert.Nil(t, err)
	for k, v := range email {
		err = db.InsertAnalysisResponse(a.Id, v, []string{v}, weight[k], now)
		assert.Nil(t, err)
	}
}

func setupFunds(t *testing.T, uid uuid.UUID, currency string, freq int64, balance int64, oldPaymentCycleId *uuid.UUID, now time.Time) *uuid.UUID {
	pow, err := utils.GetFactorInt(currency)
	assert.Nil(t, err)
	paymentCycleId, err := db.InsertNewPaymentCycleIn(1, freq, now)
	assert.Nil(t, err)
	err = api.PaymentSuccess(uid, oldPaymentCycleId, paymentCycleId, big.NewInt(balance*pow), currency, 1, freq, big.NewInt(0))
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
	e := db.SponsorEvent{
		Id:        uuid.New(),
		Uid:       *userId,
		RepoId:    *repoId,
		EventType: db.Active,
		SponsorAt: &day,
	}
	return db.InsertOrUpdateSponsor(&e)
}

func setupUnSponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day time.Time) error {
	e := db.SponsorEvent{
		Id:          uuid.New(),
		Uid:         *userId,
		RepoId:      *repoId,
		EventType:   db.Inactive,
		UnSponsorAt: &day,
	}
	return db.InsertOrUpdateSponsor(&e)
}

func setupUser(email string) (*uuid.UUID, error) {
	payOutId := uuid.New()
	u := db.User{
		Id:                uuid.New(),
		StripeId:          utils.StringPointer("strip-id"),
		PaymentCycleOutId: payOutId,
		Email:             email,
	}

	err := db.InsertUser(&u)
	if err != nil {
		return nil, err
	}
	return &u.Id, nil
}

func setupRepo(url string) (*uuid.UUID, error) {
	r := db.Repo{
		Id:          uuid.New(),
		Url:         utils.StringPointer(url),
		GitUrl:      utils.StringPointer(url),
		Source:      utils.StringPointer("github"),
		Name:        utils.StringPointer("name"),
		Description: utils.StringPointer("desc"),
		CreatedAt:   time.Time{},
	}
	err := db.InsertOrUpdateRepo(&r)
	if err != nil {
		return nil, err
	}
	return &r.Id, nil
}
