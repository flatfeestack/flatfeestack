package main

type Contributor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CommitChange struct {
	Addition int `json:"addition"`
	Deletion int `json:"deletion"`
}

type Contribution struct {
	Contributor Contributor  `json:"contributor"`
	Changes     CommitChange `json:"changes"`
}
