package db

import (
	"backend/utils"
	"github.com/google/uuid"
	"time"
)

func SetupRepo(url string) (*uuid.UUID, error) {
	r := Repo{
		Id:          uuid.New(),
		Url:         utils.StringPointer(url),
		GitUrl:      utils.StringPointer(url),
		Source:      utils.StringPointer("github"),
		Name:        utils.StringPointer("name"),
		Description: utils.StringPointer("desc"),
		CreatedAt:   time.Time{},
	}
	err := InsertOrUpdateRepo(&r)
	if err != nil {
		return nil, err
	}
	return &r.Id, nil
}
