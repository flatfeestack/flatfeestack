package main

import (
	"time"
)

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
	Merges      int          `json:"merges"`
	Commits     int          `json:"commits"`
}

type ContributionWithPlatformInformation struct {
	GitInformation      Contribution         `json:"gitInformation"`
	PlatformInformation IssueUserInformation `json:"platformInformation"`
}

type FlatFeeWeight struct {
	Contributor Contributor `json:"contributor"`
	Weight      float64     `json:"weight"`
}

type RequestGQLRepositoryInformation struct {
	Data GQLData `json:"data"`
}

type GQLData struct {
	Repository GQLRepository `json:"repository"`
}

type GQLRepository struct {
	Issues GQLIssueConnection `json:"issues"`
	Issue  GQLIssue           `json:"issue"`
	Ref    GQLRef             `json:"ref"`
}

type GQLIssueConnection struct {
	Nodes    []GQLIssue  `json:"nodes"`
	PageInfo GQLPageInfo `json:"pageInfo"`
}

type GQLIssue struct {
	Title     string                    `json:"title"`
	Number    int                       `json:"number"`
	Author    GQLActor                  `json:"author"`
	Comments  GQLIssueCommentConnection `json:"comments"`
	UpdatedAt time.Time                 `json:"updatedAt"`
}

type GQLActor struct {
	Login string `json:"login"`
}

type GQLIssueCommentConnection struct {
	Nodes    []GQLIssueComment `json:"nodes"`
	PageInfo GQLPageInfo       `json:"pageInfo"`
}

type GQLIssueComment struct {
	Author    GQLActor  `json:"author"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type GQLPageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type GQLRef struct {
	Target GQLGitObject `json:"target"`
}

type GQLGitObject struct {
	History GQLCommitHistoryConnection `json:"history"`
}

type GQLCommitHistoryConnection struct {
	Nodes []GQLCommit `json:"nodes"`
}

type GQLCommit struct {
	Author GQLGitActor
}

type GQLGitActor struct {
	User GQLActor `json:"user"`
}

type IssueUserInformation struct {
	UserName  string `json:"userName"`
	Author    []int  `json:"author"`
	Commenter int    `json:"commenter"`
}
