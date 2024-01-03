package db

import (
	"backend/pkg/util"
	"github.com/google/uuid"
	"time"
)

func SetupRepo(url string) (*uuid.UUID, error) {
	r := Repo{
		Id:          uuid.New(),
		Url:         util.StringPointer(url),
		GitUrl:      util.StringPointer(url),
		Source:      util.StringPointer("github"),
		Name:        util.StringPointer("name"),
		Description: util.StringPointer("desc"),
		CreatedAt:   time.Time{},
	}
	err := InsertOrUpdateRepo(&r)
	if err != nil {
		return nil, err
	}
	return &r.Id, nil
}
