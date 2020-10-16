package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"time"
)

type ContributionChannel struct {
	Result Contribution
	Reason error
}

func analyzeRepositoryFromString(src string, since time.Time, until time.Time, branch string) (map[Contributor]Contribution, error) {
	cloneUpdateStart := time.Now()
	repo, err := CloneOrUpdateRepository(src, branch)
	cloneUpdateEnd := time.Now()
	fmt.Printf("---> cloned/updated repository in %dms\n", cloneUpdateEnd.Sub(cloneUpdateStart).Milliseconds())
	if err != nil {
		return nil, err
	}

	return analyzeRepositoryFromRepository(repo, since, until)
}

func analyzeRepositoryFromRepository(repo *git.Repository, since time.Time, until time.Time) (map[Contributor]Contribution, error) {
	authorMap := make(map[Contributor]Contribution)
	contributionChannel := make(chan ContributionChannel)

	var timeZeroValue time.Time
	var options git.LogOptions

	if since != timeZeroValue {
		options.Since = &since
	}

	if until != timeZeroValue {
		options.Until = &until
	}

	commits, err := repo.Log(&options)
	if err != nil {
		return authorMap, err
	}

	commitCounter := 0

	gitAnalysisStart := time.Now()
	err = commits.ForEach(func(c *object.Commit) error {
		commitCounter++
		fmt.Printf("\033[2K\r%d commits", commitCounter)
		go func() {
			author := Contributor{
				Name:  c.Author.Name,
				Email: c.Author.Email,
			}

			merge := 0
			commit := 1

			if len(c.ParentHashes) > 1 {
				merge = 1
				commit = 0
			}

			stats, err := c.Stats()
			if err != nil {
				contributionChannel <- ContributionChannel{
					Result: Contribution{
						Contributor: Contributor{
							Name:  "",
							Email: "",
						},
						Changes:     CommitChange{
							Addition: 0,
							Deletion: 0,
						},
						Merges:      0,
						Commits:     0,
					},
					Reason: err,
				}
			} else {
				changes := CommitChange{
					Addition: 0,
					Deletion: 0,
				}

				// only count the lines if its not a merge
				if merge == 0 {
					for index := range stats {
						changes.Addition += stats[index].Addition
						changes.Deletion += stats[index].Deletion
					}
				} else {
					for index := range stats {
						changes.Addition += int(float64(stats[index].Addition) * mergedLinesWeight)
						changes.Deletion += int(float64(stats[index].Deletion) * mergedLinesWeight)
					}
				}

				contributionChannel <- ContributionChannel{
					Result: Contribution{
						Contributor: author,
						Changes:     changes,
						Merges:      merge,
						Commits:     commit,
					},
					Reason: nil,
				}
			}
		}()
		return nil
	})
	fmt.Println()
	answersReceived := 0
	for res := range contributionChannel {
		if res.Reason != nil {
			return nil, err
		} else {
			author := res.Result.Contributor
			if _, found := authorMap[author]; !found {
				authorMap[author] = res.Result
			} else {
				authorMap[author] = Contribution{
					Contributor: author,
					Changes: CommitChange{
						Addition: authorMap[author].Changes.Addition + res.Result.Changes.Addition,
						Deletion: authorMap[author].Changes.Deletion + res.Result.Changes.Deletion,
					},
					Merges:  authorMap[author].Merges + res.Result.Merges,
					Commits: authorMap[author].Commits + res.Result.Commits,
				}
			}
		}

		answersReceived++
		if commitCounter == answersReceived {
			close(contributionChannel)
		}
	}
	gitAnalysisEnd := time.Now()
	fmt.Printf("---> git analysis in %dms (%d commits)\n", gitAnalysisEnd.Sub(gitAnalysisStart).Milliseconds(), commitCounter)

	if err != nil {
		return authorMap, err
	}
	return authorMap, nil
}

func weightContributions(contributions map[Contributor]Contribution) (map[Contributor]FlatFeeWeight, error) {
	weightContributionsStart := time.Now()

	authorMap := make(map[Contributor]FlatFeeWeight)

	totalChanges := CommitChange{
		Addition: 0,
		Deletion: 0,
	}
	totalMerge := 0
	totalCommit := 0

	var authors []Contributor

	for _, v := range contributions {
		authors = append(authors, v.Contributor)
		totalChanges = CommitChange{
			Addition: totalChanges.Addition + v.Changes.Addition,
			Deletion: totalChanges.Deletion + v.Changes.Deletion,
		}
		totalMerge += v.Merges
		totalCommit += v.Commits
	}

	totalAmountOfAuthors := len(authors)

	for _, author := range authors {
		authorChangesWeighted := float64(contributions[author].Changes.Addition)*additionWeight + float64(contributions[author].Changes.Deletion)*deletionWeight
		totalChangesWeighted := float64(totalChanges.Addition)*additionWeight + float64(totalChanges.Deletion)*deletionWeight
		var changesPercentage float64
		if totalChangesWeighted == 0 {
			changesPercentage = 1.0 / float64(totalAmountOfAuthors)
		} else {
			changesPercentage = authorChangesWeighted / totalChangesWeighted
		}

		authorGitHistoryWeighted := float64(contributions[author].Merges)*mergeWeight + float64(contributions[author].Commits)*commitWeight
		totalGitHistoryWeighted := float64(totalMerge)*mergeWeight + float64(totalCommit)*commitWeight
		var gitHistoryPercentage float64
		if totalGitHistoryWeighted == 0 {
			gitHistoryPercentage = 1.0 / float64(totalAmountOfAuthors)
		} else {
			gitHistoryPercentage = authorGitHistoryWeighted / totalGitHistoryWeighted
		}

		authorMap[author] = FlatFeeWeight{
			Contributor: author,
			Weight:      changesPercentage*changesWeight + gitHistoryPercentage*gitHistoryWeight,
		}
	}

	weightContributionsEnd := time.Now()
	fmt.Printf("---> weight contributions in %dms\n", weightContributionsEnd.Sub(weightContributionsStart).Milliseconds())
	return authorMap, nil
}

func weightContributionsWithPlatformInformation(contributions map[Contributor]ContributionWithPlatformInformation) (map[Contributor]FlatFeeWeight, error) {
	weightContributionsStart := time.Now()


	authorMap := make(map[Contributor]FlatFeeWeight)

	totalChanges := CommitChange{
		Addition: 0,
		Deletion: 0,
	}
	totalMerge := 0
	totalCommit := 0

	totalIssues := 0
	totalComments := 0
	totalCommenter := 0

	totalPullRequestsValue := 0.0
	totalPullRequestsReviews := 0

	var authors []Contributor

	for _, v := range contributions {
		authors = append(authors, v.GitInformation.Contributor)
		totalChanges = CommitChange{
			Addition: totalChanges.Addition + v.GitInformation.Changes.Addition,
			Deletion: totalChanges.Deletion + v.GitInformation.Changes.Deletion,
		}
		totalMerge += v.GitInformation.Merges
		totalCommit += v.GitInformation.Commits
		totalIssues += len(v.PlatformInformation.IssueInformation.Author)
		totalComments += getSumOfCommentsOfIssues(v.PlatformInformation.IssueInformation.Author)
		totalCommenter += v.PlatformInformation.IssueInformation.Commenter
		totalPullRequestsValue += getAuthorPullRequestValue(v.PlatformInformation.PullRequestInformation.Author)
		totalPullRequestsReviews += v.PlatformInformation.PullRequestInformation.Reviewer
	}

	totalAmountOfAuthors := len(authors)

	for _, author := range authors {
		authorChangesWeighted := float64(contributions[author].GitInformation.Changes.Addition)*additionWeight + float64(contributions[author].GitInformation.Changes.Deletion)*deletionWeight
		totalChangesWeighted := float64(totalChanges.Addition)*additionWeight + float64(totalChanges.Deletion)*deletionWeight
		var changesPercentage float64
		if totalChangesWeighted == 0 {
			changesPercentage = 1.0 / float64(totalAmountOfAuthors)
		} else {
			changesPercentage = authorChangesWeighted / totalChangesWeighted
		}

		authorGitHistoryWeighted := float64(contributions[author].GitInformation.Merges)*mergeWeight + float64(contributions[author].GitInformation.Commits)*commitWeight
		totalGitHistoryWeighted := float64(totalMerge)*mergeWeight + float64(totalCommit)*commitWeight
		var gitHistoryPercentage float64
		if totalGitHistoryWeighted == 0 {
			gitHistoryPercentage = 1.0 / float64(totalAmountOfAuthors)
		} else {
			gitHistoryPercentage = authorGitHistoryWeighted / totalGitHistoryWeighted
		}

		authorIssuesWeighted := float64(len(contributions[author].PlatformInformation.IssueInformation.Author))*issueWeight + float64(getSumOfCommentsOfIssues(contributions[author].PlatformInformation.IssueInformation.Author))*issueCommentsWeight + float64(contributions[author].PlatformInformation.IssueInformation.Commenter)*issueCommenterWeight
		totalIssuesWeighted := float64(totalIssues)*issueWeight + float64(totalComments)*issueCommentsWeight + float64(totalCommenter)*issueCommenterWeight
		var issuesPercentage float64
		if totalIssuesWeighted == 0 {
			issuesPercentage = 1.0 / float64(totalAmountOfAuthors)
		} else {
			issuesPercentage = authorIssuesWeighted / totalIssuesWeighted
		}

		authorPullRequestsWeighted := getAuthorPullRequestValue(contributions[author].PlatformInformation.PullRequestInformation.Author)*pullRequestAuthorWeight + float64(contributions[author].PlatformInformation.PullRequestInformation.Reviewer)*pullRequestReviewerWeight
		totalPullRequestsWeighted := totalPullRequestsValue*pullRequestAuthorWeight + float64(totalPullRequestsReviews)*pullRequestReviewerWeight
		var pullRequestPercentage float64
		if totalPullRequestsWeighted == 0 {
			pullRequestPercentage = 1.0 / float64(totalAmountOfAuthors)
		} else {
			pullRequestPercentage = authorPullRequestsWeighted / totalPullRequestsWeighted
		}

		authorMap[author] = FlatFeeWeight{
			Contributor: author,
			Weight:      changesPercentage*changesWeightPlatformInfo + gitHistoryPercentage*gitHistoryWeightPlatformInfo + issuesPercentage*issueCategoryWeightPlatformInfo + pullRequestPercentage*pullRequestCategoryWeightPlatformInfo,
		}
	}

	weightContributionsEnd := time.Now()
	fmt.Printf("---> weight contributions in %dms\n", weightContributionsEnd.Sub(weightContributionsStart).Milliseconds())
	return authorMap, nil
}

func getSumOfCommentsOfIssues(issues []int) int {
	totalComments := 0
	for _, i := range issues {
		totalComments += i
	}
	return totalComments
}

func getAuthorPullRequestValue(pullRequests []PullRequestInformation) float64 {
	totalScore := 0.0

	for _, request := range pullRequests {
		isApproved := false
		for _, review := range request.Reviews {
			if review == "APPROVED" {
				isApproved = true
				break
			}
		}
		currentValue := 0.0
		switch request.State {
		case "CLOSED":
			currentValue = pullRequestClosedValue
		case "OPEN":
			currentValue = pullRequestOpenValue
		case "MERGED":
			currentValue = pullRequestMergedValue
		default:
			currentValue = pullRequestOpenValue
		}

		if isApproved {
			totalScore += currentValue * approvedMultiplier
		} else {
			totalScore += currentValue
		}
	}
	return totalScore
}
