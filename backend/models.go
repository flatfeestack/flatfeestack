package main

// User schema of the user table
type User struct {
	ID       string  `json:"id"`
	StripeId     string `json:"stripe_id"`
	Email string `json:"email"`
}

type UserRepository interface {
	FindByID(ID string) (*User, error)
	Save(user *User) error
}

type Repo struct {
	ID string `json:"id"`
	Url string `json:"url"`
	Name string `json:"name"`
}

type RepoRepository interface {
	FindByID(ID string) (*Repo, error)
	Save(repo *Repo) error
}

type SponsorEvent struct {
	ID string `json:"id"`
	Uid string `json:"uid"`
	RepoId string `json:"repo_id"`
	Type string `json:"type"`
	Timestamp string `json:"timestamp"`
}

type SponsorEventRepository interface {
	Sponsor(repoID string, uid string) (*SponsorEvent, error)
	Unsponsor(repoId string, uid string) error
}