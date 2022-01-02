package main

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

	setupContributionScenario1(t)

	err := dailyRunner(day1)
	assert.Nil(t, err)

	/*drh := schedSQL(t, "user_id, repo_hours, day", "daily_repo_hours", "repo_hours")
	assert.Equal(t, 1, len(drh))
	assert.Equal(t, 20, drh[0].Number)
	drb := schedSQL(t, "repo_id, balance, day", "daily_repo_balance", "balance")
	assert.Equal(t, 1, len(drb))
	assert.Equal(t, 13750, drb[0].Number)

	err = dailyRunner(day2)
	assert.Nil(t, err)
	drh = schedSQL(t, "user_id, repo_hours, day", "daily_repo_hours", "repo_hours")
	assert.Equal(t, 2, len(drh))
	assert.Equal(t, 1, drh[0].Number)
	drb = schedSQL(t, "repo_id, balance, day", "daily_repo_balance", "balance")
	assert.Equal(t, 2, len(drb))
	assert.Equal(t, 13750, drb[0].Number)*/
}

func TestDailyRunner2(t *testing.T) {
	setup()
	defer teardown()

	setupContributionScenario2(t)

	err := dailyRunner(day2)
	assert.Nil(t, err)

	/*drh := schedSQL(t, "user_id, repo_hours, day", "daily_repo_hours", "repo_hours")
	assert.Equal(t, 2, len(drh))
	assert.Equal(t, 1, drh[0].Number)
	assert.Equal(t, 19, drh[1].Number)
	drb := schedSQL(t, "repo_id, balance, day", "daily_repo_balance", "balance")
	assert.Equal(t, 2, len(drb))
	assert.Equal(t, 6512, drb[0].Number)
	assert.Equal(t, 20986, drb[1].Number)

	err = dailyRunner(day3)
	assert.Nil(t, err)
	drh = schedSQL(t, "user_id, repo_hours, day", "daily_repo_hours", "repo_hours")
	assert.Equal(t, 3, len(drh))
	assert.Equal(t, 24, drh[2].Number)
	drb = schedSQL(t, "repo_id, balance, day", "daily_repo_balance", "balance")
	assert.Equal(t, 3, len(drb))
	assert.Equal(t, 20262, drb[1].Number)

	err = dailyRunner(day3)
	assert.NotNil(t, err) //running twice should give an error due to constraints*/
}

func TestDailyRunner3(t *testing.T) {
	setup()
	defer teardown()

	users, repos := setupContributionScenario2(t)
	setupContributor(t, *repos[0], day0, day1, []string{"tom@tom.tom", "jon@jon.jon", "me@me.me"}, []float64{0.3, 0.5, 0.2}, day0)
	setupContributor(t, *repos[0], day1, day2, []string{"sam@sam.sam", "jon@jon.jon", "me@me.me"}, []float64{0.6, 0.3, 0.1}, day0)
	setupContributor(t, *repos[1], day2, day3, []string{"tom@tom.tom", "sam@sam.sam"}, []float64{0.1, 0.9}, day0)

	setupGitEmail(t, *users[0], "tom@tom.tom", day0)
	setupGitEmail(t, *users[1], "sam@sam.sam", day0)

	setupSponsor(t, users[0], repos[0], 2, 0) //tomp2p gets sponsoring starting on day 2

	err := dailyRunner(day1) //day 0-1 tom gets all,
	assert.Nil(t, err)

	err = dailyRunner(day2) //day 1-2 sam gets all
	assert.Nil(t, err)

	err = dailyRunner(day3) //day 2-3 tom and sam split
	assert.Nil(t, err)

	err = dailyRunner(day4) //day 3-4 tom and sam split
	assert.Nil(t, err)

	/*rb1 := schedSQL(t, "user_id, balance, day", "daily_user_payout", "balance")
	assert.Equal(t, 6, len(rb1))
	assert.Equal(t, 1375, rb1[0].Number)
	assert.Equal(t, 2026, rb1[1].Number)
	assert.Equal(t, 13750, rb1[2].Number)
	assert.Equal(t, 20986, rb1[3].Number)
	assert.Equal(t, 26125, rb1[4].Number)
	assert.Equal(t, 31985, rb1[5].Number)

	//repo1 has a balance of 43283, check if we split it correctly 0.3 (12984), 0.5 (21641), 0.2 (8656)
	rb2 := schedSQLEmail(t)
	assert.Equal(t, 14, len(rb2))*/
}

func setupGitEmail(t *testing.T, user uuid.UUID, email string, now time.Time) {
	err := insertGitEmail(user, email, nil, now)
	assert.Nil(t, err)
}

func setupContributor(t *testing.T, repo uuid.UUID, from time.Time, to time.Time, email []string, weight []float64, now time.Time) {
	aid := uuid.New()
	//id uuid.UUID, repo_id uuid.UUID, date_from time.Time, date_to time.Time, branch string
	err := insertAnalysisRequest(aid, repo, from, to, "test", now)
	assert.Nil(t, err)
	for k, v := range email {
		w1 := FlatFeeWeight{
			Email:  v,
			Name:   v,
			Weight: weight[k],
		}
		err = insertAnalysisResponse(aid, &w1, timeNow())
		assert.Nil(t, err)
	}
}

func setupContributionScenario1(t *testing.T) ([]*uuid.UUID, []*uuid.UUID) {
	/*sponsors := setupUsers(t, "tom@tom.tom sp1", "sam@sam.sam sp2", "arm@arm.arm sp3", "gui@gui.gui sp4", "mar@mar.mar sp5")
	contributors := setupUsers(t, "ste@ste.ste c1", "pea@pea.pea c2", "luc@luc.luc c3", "nic@nic.nic c4")
	repos := setupRepos(t, "tomp2p 1", "yaml 2", "sql 3", "linux 4")
	setupFundsUSD(sponsors[0], 120000000, 300000)
	setupSponsor(t, users[0], repos[0], 0, 0)
	setupUnsponsor(t, users[0], repos[0], 0, 12)
	setupSponsor(t, users[0], repos[0], 0, 12)
	setupUnsponsor(t, users[0], repos[0], 0, 20)
	setupSponsor(t, users[0], repos[0], 1, 0)
	setupUnsponsor(t, users[0], repos[0], 1, 1)
	//TomP2P, day 0: 20h from tom
	//TomP2P, day 1: 1h from tom
	//
	//repo balance for day1 (tomp2p total repo_hours day1 1)
	//13750
	//return users, repos*/
	return nil, nil
}

func setupFundsUSD(u *uuid.UUID, total uint32, daily uint32) {

}

func setupContributionScenario2(t *testing.T) ([]*uuid.UUID, []*uuid.UUID) {
	users, repos := setupContributionScenario1(t)
	setupSponsor(t, users[1], repos[0], 1, 0)
	setupUnsponsor(t, users[1], repos[0], 1, 10)
	setupSponsor(t, users[1], repos[1], 1, 8)
	setupUnsponsor(t, users[1], repos[1], 1, 16)
	setupSponsor(t, users[1], repos[1], 1, 23)
	//TomP2P, day 1: 10h from sam
	//YML, day 1: 8h +1h from sam
	//YML, continues day 2
	//
	//repo balance for day1 (sam total repo_hours day1 19)
	//10/8+1 (20986/5789+723)
	return users, repos
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

func schedSQL(t *testing.T, col string, table string, order string) []IDNumberDay {
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
}

func setupSponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day int, hour int) {
	e := SponsorEvent{
		Id:        uuid.New(),
		Uid:       *userId,
		RepoId:    *repoId,
		EventType: Active,
		SponsorAt: time.Time{}.Add((time.Duration(day) * 24 * time.Hour) + (time.Duration(hour) * time.Hour)),
	}
	err1, err2 := insertOrUpdateSponsor(&e)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func setupUnsponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day int, hour int) {
	e := SponsorEvent{
		Id:          uuid.New(),
		Uid:         *userId,
		RepoId:      *repoId,
		EventType:   Inactive,
		UnsponsorAt: time.Time{}.Add((time.Duration(day) * 24 * time.Hour) + (time.Duration(hour) * time.Hour)),
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
