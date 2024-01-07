package handler

import "github.com/google/uuid"

type PublicUser struct {
	Id    uuid.UUID `json:"id"`
	Name  *string   `json:"name,omitempty"`
	Image *string   `json:"image"`
}
