package db

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
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

func TestMain(m *testing.M) {
	file, err := os.CreateTemp("", "sqlite")
	defer os.Remove(file.Name())

	err = InitDb("sqlite3", file.Name(), "")
	if err != nil {
		log.Fatal(err)
	}

	code := m.Run()

	err = db.Close()
	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	if err != nil {
		log.Warnf("Could not start resource: %s", err)
	}

	os.Exit(code)
}

func runSQL(files ...string) error {
	for _, file := range files {
		if file == "" {
			continue
		}
		//https://stackoverflow.com/questions/12518876/how-to-check-if-a-file-exists-in-go
		if _, err := os.Stat(file); err == nil {
			fileBytes, err := ioutil.ReadFile(file)
			if err != nil {
				return err
			}

			//https://stackoverflow.com/questions/12682405/strip-out-c-style-comments-from-a-byte
			re := regexp.MustCompile("(?s)//.*?\n|/\\*.*?\\*/|(?s)--.*?\n|(?s)#.*?\n")
			newBytes := re.ReplaceAll(fileBytes, nil)

			requests := strings.Split(string(newBytes), ";")
			for _, request := range requests {
				request = strings.TrimSpace(request)
				if len(request) > 0 {
					_, err := db.Exec(request)
					if err != nil {
						return fmt.Errorf("[%v] %v", request, err)
					}
				}
			}
		} else {
			log.Printf("ignoring file %v (%v)", file, err)
		}
	}
	return nil
}

func setup() {
	err := runSQL("init.sql")
	if err != nil {
		log.Fatalf("Could not run init.sql scripts: %s", err)
	}
}
func teardown() {
	err := runSQL("drop_test.sql")
	if err != nil {
		log.Fatalf("Could not run drop_test.sql: %s", err)
	}
}

func TestUser(t *testing.T) {
	setup()
	defer teardown()

	payOutId := uuid.New()
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		PaymentCycleOutId: payOutId,
		Email:             "email",
	}

	err := InsertUser(&u)
	assert.Nil(t, err)

	u2, err := FindUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u2)

	u3, err := FindUserByEmail("email")
	assert.Nil(t, err)
	assert.NotNil(t, u3)

	u.Email = "email2"
	err = UpdateUser(&u)
	assert.Nil(t, err)

	//cannot change Email
	u4, err := FindUserByEmail("email2")
	assert.Nil(t, err)
	assert.Nil(t, u4)

	u5, err := FindUserById(u.Id)
	assert.Nil(t, err)
	assert.NotNil(t, u5)
}

func TestSponsor(t *testing.T) {
	setup()
	defer teardown()

	payOutId := uuid.New()
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		PaymentCycleOutId: payOutId,
		Email:             "email",
	}

	r := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("giturl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := InsertUser(&u)
	assert.Nil(t, err)
	err = InsertOrUpdateRepo(&r)
	assert.Nil(t, err)

	t1 := time.Time{}.Add(time.Duration(1) * time.Second)
	t2 := time.Time{}.Add(time.Duration(2) * time.Second)
	t3 := time.Time{}.Add(time.Duration(3) * time.Second)
	t4 := time.Time{}.Add(time.Duration(4) * time.Second)

	s1 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Active,
		SponsorAt:   t1,
		UnSponsorAt: &t1,
	}

	s2 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		SponsorAt:   t2,
		UnSponsorAt: &t2,
	}

	s3 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Active,
		SponsorAt:   t3,
		UnSponsorAt: &t3,
	}

	err = InsertOrUpdateSponsor(&s1)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s2)
	assert.Nil(t, err)
	err = InsertOrUpdateSponsor(&s3)
	assert.Nil(t, err)

	rs, err := FindSponsoredReposById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rs))

	s4 := SponsorEvent{
		Id:          uuid.New(),
		Uid:         u.Id,
		RepoId:      r.Id,
		EventType:   Inactive,
		SponsorAt:   t4,
		UnSponsorAt: &t4,
	}
	err = InsertOrUpdateSponsor(&s4)
	assert.Nil(t, err)

	rs, err = FindSponsoredReposById(u.Id)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rs))
}

func TestRepo(t *testing.T) {
	setup()
	defer teardown()

	r := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("giturl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := InsertOrUpdateRepo(&r)
	assert.Nil(t, err)

	r2, err := FindRepoById(uuid.New())
	assert.Nil(t, err)
	assert.Nil(t, r2)

	r3, err := FindRepoById(r.Id)
	assert.Nil(t, err)
	assert.NotNil(t, r3)
}

func saveTestUser(t *testing.T, email string) uuid.UUID {
	payOutId := uuid.New()
	u := User{
		Id:                uuid.New(),
		StripeId:          stringPointer("strip-id"),
		PaymentCycleOutId: payOutId,
		Email:             email,
	}

	err := InsertUser(&u)
	assert.Nil(t, err)
	return u.Id
}

func TestGitEmail(t *testing.T) {
	setup()
	defer teardown()

	uid := saveTestUser(t, "email1")

	err := InsertGitEmail(uid, "email1", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	err = InsertGitEmail(uid, "email2", stringPointer("A"), time.Now())
	assert.Nil(t, err)
	emails, err := FindGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(emails))
	err = DeleteGitEmail(uid, "email2")
	assert.Nil(t, err)
	emails, err = FindGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(emails))
	err = DeleteGitEmail(uid, "email1")
	assert.Nil(t, err)
	emails, err = FindGitEmailsByUserId(uid)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(emails))
}

func TestAnalysisRequest(t *testing.T) {
	setup()
	defer teardown()

	r := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("gitUrl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := InsertOrUpdateRepo(&r)
	assert.Nil(t, err)

	ar, err := FindLatestAnalysisRequest(r.Id)
	assert.Nil(t, err)
	assert.Nil(t, ar)

	a := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   r.Id,
		DateFrom: day1,
		DateTo:   day2,
		GitUrl:   *r.GitUrl,
	}
	err = InsertAnalysisRequest(a, time.Now())
	assert.Nil(t, err)

	//insertAnalysisResponse

	abd, err := FindLatestAnalysisRequest(r.Id)
	assert.Nil(t, err)
	assert.Equal(t, abd.Id, a.Id)
	assert.Equal(t, abd.RepoId, a.RepoId)
}

func TestAnalysisRequest2(t *testing.T) {
	setup()
	defer teardown()

	r1 := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("gitUrl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := InsertOrUpdateRepo(&r1)
	assert.Nil(t, err)

	r2 := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url2"),
		GitUrl:      stringPointer("gitUrl2"),
		Source:      stringPointer("source2"),
		Name:        stringPointer("name2"),
		Description: stringPointer("desc2"),
	}
	err = InsertOrUpdateRepo(&r2)
	assert.Nil(t, err)

	a1 := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   r1.Id,
		DateFrom: day1,
		DateTo:   day2,
		GitUrl:   *r1.GitUrl,
	}
	err = InsertAnalysisRequest(a1, time.Now())
	assert.Nil(t, err)

	a2 := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   r1.Id,
		DateFrom: day2,
		DateTo:   day3,
		GitUrl:   *r1.GitUrl,
	}
	err = InsertAnalysisRequest(a2, time.Now())
	assert.Nil(t, err)

	//insertAnalysisResponse

	abd, err := FindLatestAnalysisRequest(r1.Id)
	assert.Nil(t, err)
	assert.Equal(t, abd.Id, a2.Id)
	assert.Equal(t, abd.RepoId, a2.RepoId)

	a3 := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   r2.Id,
		DateFrom: day3,
		DateTo:   day4,
		GitUrl:   *r2.GitUrl,
	}
	err = InsertAnalysisRequest(a3, time.Now())
	assert.Nil(t, err)

	//insertAnalysisResponse

	abd, err = FindLatestAnalysisRequest(r1.Id)
	assert.Nil(t, err)
	assert.Equal(t, abd.Id, a2.Id)
	assert.Equal(t, abd.RepoId, a2.RepoId)

	a4 := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   r2.Id,
		DateFrom: day4,
		DateTo:   day5,
		GitUrl:   *r2.GitUrl,
	}
	err = InsertAnalysisRequest(a4, time.Now())
	assert.Nil(t, err)

	alar, err := FindAllLatestAnalysisRequest(day2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(alar))
	assert.Equal(t, alar[0].RepoId, r1.Id)
	assert.Equal(t, alar[1].RepoId, r2.Id)

	alar, err = FindAllLatestAnalysisRequest(day3)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(alar))
	assert.Equal(t, alar[0].RepoId, r1.Id)
	assert.Equal(t, alar[1].RepoId, r2.Id)
}

func TestAnalysisResponse(t *testing.T) {
	setup()
	defer teardown()

	r := Repo{
		Id:          uuid.New(),
		Url:         stringPointer("url"),
		GitUrl:      stringPointer("gitUrl"),
		Source:      stringPointer("source"),
		Name:        stringPointer("name"),
		Description: stringPointer("desc"),
	}
	err := InsertOrUpdateRepo(&r)
	assert.Nil(t, err)

	a := AnalysisRequest{
		Id:       uuid.New(),
		RepoId:   r.Id,
		DateFrom: day1,
		DateTo:   day2,
		GitUrl:   *r.GitUrl,
	}
	err = InsertAnalysisRequest(a, time.Now())
	assert.Nil(t, err)

	err = InsertAnalysisResponse(a.Id, "tom", []string{"tom"}, 0.5, time.Now())
	assert.Nil(t, err)
	err = InsertAnalysisResponse(a.Id, "tom", []string{"tom"}, 0.4, time.Now())
	assert.NotNil(t, err)
	err = InsertAnalysisResponse(a.Id, "tom2", []string{"tom2"}, 0.4, time.Now())
	assert.Nil(t, err)

	ar, err := FindAnalysisResults(a.Id)
	assert.Equal(t, 2, len(ar))
	assert.Equal(t, ar[0].GitNames[0], "tom")
	assert.Equal(t, ar[1].GitNames[0], "tom2")

	fmt.Printf("AAAA %v\n", a.Id)

	err = UpdateAnalysisRequest(a.Id, day2, stringPointer("test"))
	assert.Nil(t, err)

	alr, err := FindLatestAnalysisRequest(r.Id)
	assert.Nil(t, err)
	assert.Equal(t, day3.Nanosecond(), alr.ReceivedAt.Nanosecond())
}
