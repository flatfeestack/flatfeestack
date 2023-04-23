package db

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

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
