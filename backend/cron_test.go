package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
	"time"
)

type IDNumberDay struct {
	Id     uuid.UUID
	Number int
	Day    time.Time
}

type EmailNumberDay struct {
	Email  string
	Number int
	Day    time.Time
}

var (
	day0 = time.Time{}
	day1 = time.Time{}.Add(time.Duration(1*24) * time.Hour)
	day2 = time.Time{}.Add(time.Duration(2*24) * time.Hour)
	day3 = time.Time{}.Add(time.Duration(3*24) * time.Hour)
	day4 = time.Time{}.Add(time.Duration(4*24) * time.Hour)
)

func TestDailyRunner1(t *testing.T) {
	setup()
	defer teardown()

	sponsors := setupUsers(t, "tom@tom.tom s1", "mic@mic.mic s2", "arm@arm.arm s3", "gui@gui.gui s4", "mar@mar.mar s5")
	setupFundsUSD(t, *sponsors[0], 120)

	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	setupGitEmail(t, *contributors[0], "ste@ste.ste")
	repos := setupRepos(t, "tomp2p r1", "neow3j r2", "sql r3", "linux r4")
	setupContributor(t, *repos[0], day0, day1, []string{"ste@ste.ste"}, []float64{0.5})

	setupSponsor(t, sponsors[0], repos[0], day0)
	setupUnsponsor(t, sponsors[0], repos[0], day0)
	setupSponsor(t, sponsors[0], repos[0], day0)

	err := dailyRunner(day2)
	assert.Nil(t, err)

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

func setupFundsUSD(t *testing.T, uid uuid.UUID, balance int64) {
	paymentCycleId, err := insertNewPaymentCycle(uid, 1, 365, timeNow())
	assert.Nil(t, err)
	err = paymentSuccess(uid, uuid.UUID{}, *paymentCycleId, big.NewInt(balance*1_000_000), "USD", 1, 365, big.NewInt(0))
	assert.Nil(t, err)
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

func setupSponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day time.Time) {
	e := SponsorEvent{
		Id:        uuid.New(),
		Uid:       *userId,
		RepoId:    *repoId,
		EventType: Active,
		SponsorAt: day,
	}
	err1, err2 := insertOrUpdateSponsor(&e)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func setupUnsponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day time.Time) {
	e := SponsorEvent{
		Id:          uuid.New(),
		Uid:         *userId,
		RepoId:      *repoId,
		EventType:   Inactive,
		UnsponsorAt: day,
	}
	err1, err2 := insertOrUpdateSponsor(&e)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func setupUser(email string) (*uuid.UUID, error) {
	u := User{
		Id:       uuid.New(),
		StripeId: stringPointer("strip-id"),
		Email:    email,
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

/*func schedSQL(t *testing.T, col string, table string, order string) []IDNumberDay {
	//sql := "SELECT repo_hours FROM daily_repo_hours"
	sql := "SELECT " + col + " FROM " + table + " ORDER by " + order
	rows, err := db.Query(sql)
	defer rows.Close()
	assert.Nil(t, err)

	var rhs []IDNumberDay
	for rows.Next() {
		var rh IDNumberDay
		err = rows.Scan(&rh.Id, &rh.Number, &rh.Day)
		assert.Nil(t, err)
		rhs = append(rhs, rh)
	}
	return rhs
}

func schedSQLEmail(t *testing.T) []EmailNumberDay {
	sql := "SELECT email, balance, day FROM daily_email_payout ORDER by balance"
	rows, err := db.Query(sql)
	defer rows.Close()
	assert.Nil(t, err)

	var rhs []EmailNumberDay
	for rows.Next() {
		var rh EmailNumberDay
		err = rows.Scan(&rh.Email, &rh.Number, &rh.Day)
		assert.Nil(t, err)
		rhs = append(rhs, rh)
	}
	return rhs
}*/
