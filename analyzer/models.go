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
	Merges		int			`json:"merges"`
	Commits		int 		`json:"commits"`
}

type FlatFeeWeight struct {
	Contributor Contributor	`json:"contributor"`
	Weight		float64		`json:"weight"`
}

type RequestGQLRepositoryInformation struct {
	Data	GQLData		`json:"data"`
}

type GQLData struct {
	Repository 	GQLRepository `json:"repository"`
}

type GQLRepository struct {
	Issues			GQLIssueConnection		`json:"issues"`
}

type GQLIssueConnection struct {
	Edges	[]GQLIssueEdge	`json:"edges"`
}

type GQLIssueEdge struct {
	Cursor	string		`json:"cursor"`
	Node	GQLIssue	`json:"node"`
}

type GQLIssue struct {
	Title		string						`json:"title"`
	Number		int							`json:"number"`
	Author		GQLActor					`json:"author"`
	Comments	GQLIssueCommentConnection	`json:"comments"`
}

type GQLActor struct {
	Login	string	`json:"login"`
}

type GQLIssueCommentConnection struct {
	Edges		[]GQLIssueCommentEdge		`json:"edges"`
	TotalCount	int						`json:"totalCount"`
}

type GQLIssueCommentEdge struct {
	Cursor	string			`json:"cursor"`
	Node	GQLIssueComment	`json:"node"`
}

type GQLIssueComment struct {
	Author	GQLActor	`json:"author"`
}
