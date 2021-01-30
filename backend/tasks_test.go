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

func TestDailyRunner1(t *testing.T) {
	setup()
	defer teardown()

	setupContributionScenario1(t)

	err := dailyRunner(time.Time{}.Add(time.Duration(48) * time.Hour))
	assert.Nil(t, err)

	rh := schedSQL(t, "user_id, repo_hours, day", "daily_repo_hours", "repo_hours")
	assert.Equal(t, 1, len(rh))
	assert.Equal(t, 47, rh[0].Number)

	rb := schedSQL(t, "repo_id, balance, day", "daily_repo_balance", "balance")
	assert.Equal(t, 2, len(rb))
	assert.Equal(t, 13457, rb[0].Number) // 27500 * (23/47)
	assert.Equal(t, 14042, rb[1].Number) // 27500 * (24/47)
}

func TestDailyRunner2(t *testing.T) {
	setup()
	defer teardown()

	setupContributionScenario2(t)

	err := dailyRunner(time.Time{}.Add(time.Duration(48) * time.Hour))
	assert.Nil(t, err)

	rh := schedSQL(t, "user_id, repo_hours, day", "daily_repo_hours", "repo_hours")
	assert.Equal(t, 2, len(rh))
	assert.Equal(t, 71, rh[1].Number) //24+23+24
	assert.Equal(t, 41, rh[0].Number)

	rb := schedSQL(t, "repo_id, balance, day", "daily_repo_balance", "balance")
	assert.Equal(t, 4, len(rb))
	assert.Equal(t, 8908, rb[0].Number)  // 27500 * (23/71) -repo1
	assert.Equal(t, 9295, rb[1].Number)  // 27500 * (24/71) -repo4
	assert.Equal(t, 12743, rb[2].Number) // 27500 * (19/41) -repo3
	assert.Equal(t, 24051, rb[3].Number) // 27500 * (24/71) + 27500 * (22/41) -repo2
}

func TestWeeklyRunner(t *testing.T) {
	setup()
	defer teardown()

	repos, _ := setupContributionScenario2(t)
	nextWeek := time.Time{}.Add(time.Duration(24*7) * time.Hour)
	setupContributor(t, *repos[0], time.Time{}, nextWeek, []string{"tom@tom.tom", "jon@jon.jon", "me@me.me"}, []float64{0.3, 0.5, 0.2})

	err := dailyRunner(time.Time{}.Add(time.Duration(2*24) * time.Hour))
	assert.Nil(t, err)

	err = dailyRunner(time.Time{}.Add(time.Duration(3*24) * time.Hour))
	assert.Nil(t, err)

	err = weeklyRunner(time.Time{}.Add(time.Duration(24*7) * time.Hour))
	assert.Nil(t, err)

	rb1 := schedSQL(t, "repo_id, balance, day", "weekly_repo_balance", "balance")
	assert.Equal(t, 5, len(rb1))

	//repo1 has a balance of 43283, check if we split it correctly 0.3 (12984), 0.5 (21641), 0.2 (8656)
	rb2 := schedSQLEmail(t)
	assert.Equal(t, 3, len(rb2))
	assert.Equal(t, 8656, rb2[0].Number)
	assert.Equal(t, 12984, rb2[1].Number)
	assert.Equal(t, 21641, rb2[2].Number)
}

func setupContributor(t *testing.T, repo uuid.UUID, from time.Time, to time.Time, email []string, weight []float64) {
	aid := uuid.New()
	//id uuid.UUID, repo_id uuid.UUID, date_from time.Time, date_to time.Time, branch string
	err := saveAnalysisRequest(aid, repo, from, to, "test", timeNow())
	assert.Nil(t, err)
	for k, v := range email {
		w1 := FlatFeeWeight{
			Email:  v,
			Weight: weight[k],
		}
		err = saveAnalysisResponse(aid, &w1, timeNow())
		assert.Nil(t, err)
	}
}

func setupContributionScenario1(t *testing.T) ([]*uuid.UUID, []*uuid.UUID) {
	users := setupUsers(t, "tom@tom.tom", "sam@sam.sam", "unreferenced-user")
	repos := setupRepos(t, "tomp2p", "yaml", "sql", "unreferenced-repo")
	setupSponsor(t, users[0], repos[0], 25)
	setupSponsor(t, users[0], repos[1], 24)
	setupSponsor(t, users[1], repos[0], 0)    //should not be counted
	setupUnsponsor(t, users[1], repos[0], 23) //should not be counted
	setupSponsor(t, users[0], repos[2], 48)   //should not be counted
	return repos, users
}

// repo 0 has 2 sponsors from hours 25, 48 -> total per month 29*24 + 23, 28 * 24 = 719 (user0), 672 (user1) -- repo0
// repo 1 has 2 sponsors from hours 24, 26 -> total per month 30*24, 29*24 + 22 = 720 (user0), 718 (user1) -- repo1
func setupContributionScenario2(t *testing.T) ([]*uuid.UUID, []*uuid.UUID) {
	users := setupUsers(t, "tom@tom.tom", "sam@sam.sam", "unreferenced-user", "jon@jon.jon")
	repos := setupRepos(t, "tomp2p", "yaml", "sql", "xml", "json", "unreferenced-repo")
	setupSponsor(t, users[0], repos[0], 25) //u0, 25, u3, 48
	setupSponsor(t, users[0], repos[1], 24)
	setupSponsor(t, users[1], repos[1], 26)
	setupSponsor(t, users[1], repos[2], 29)
	setupSponsor(t, users[0], repos[3], 1)  //counted 24h in the next day
	setupSponsor(t, users[0], repos[4], 48) //not counted in daily
	setupSponsor(t, users[3], repos[0], 48) //not counted in daily
	return repos, users
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
	sql := "SELECT email, balance, day FROM weekly_email_payout ORDER by balance"
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

func setupSponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day int) {
	e := SponsorEvent{
		Id:          uuid.New(),
		Uid:         *userId,
		RepoId:      *repoId,
		EventType:   SPONSOR,
		SponsorAt:   time.Time{}.Add(time.Duration(day) * time.Hour),
		UnsponsorAt: time.Time{},
	}
	err1, err2 := sponsor(&e)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func setupUnsponsor(t *testing.T, userId *uuid.UUID, repoId *uuid.UUID, day int64) {
	e := SponsorEvent{
		Id:          uuid.New(),
		Uid:         *userId,
		RepoId:      *repoId,
		EventType:   UNSPONSOR,
		SponsorAt:   time.Time{},
		UnsponsorAt: time.Time{}.Add(time.Duration(day) * time.Hour),
	}
	err1, err2 := sponsor(&e)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
}

func setupUser(email string) (*uuid.UUID, error) {
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		Email:             stringPointer(email),
		Subscription:      stringPointer("sub"),
		SubscriptionState: stringPointer("ACTIVE"),
		PayoutETH:         stringPointer("0x123"),
	}

	err := saveUser(&u)
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
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	id, err := saveRepo(&r)
	if err != nil {
		return nil, err
	}
	return id, nil
}
