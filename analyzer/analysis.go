package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"time"
)

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
			return err
		}
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

		if _, found := authorMap[author]; !found {
			authorMap[author] = Contribution{
				Contributor: author,
				Changes:     changes,
				Merges:      merge,
				Commits:     commit,
			}
		} else {
			authorMap[author] = Contribution{
				Contributor: author,
				Changes: CommitChange{
					Addition: authorMap[author].Changes.Addition + changes.Addition,
					Deletion: authorMap[author].Changes.Deletion + changes.Deletion,
				},
				Merges:  authorMap[author].Merges + merge,
				Commits: authorMap[author].Commits + commit,
			}
		}
		return nil
	})
	fmt.Println()
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

	for _, author := range authors {
		authorChangesWeighted := float64(contributions[author].Changes.Addition)*additionWeight + float64(contributions[author].Changes.Deletion)*deletionWeight
		totalChangesWeighted := float64(totalChanges.Addition)*additionWeight + float64(totalChanges.Deletion)*deletionWeight
		changesPercentage := authorChangesWeighted / totalChangesWeighted

		authorGitHistoryWeighted := float64(contributions[author].Merges)*mergeWeight + float64(contributions[author].Commits)*commitWeight
		totalGitHistoryWeighted := float64(totalMerge)*mergeWeight + float64(totalCommit)*commitWeight
		gitHistoryPercentage := authorGitHistoryWeighted / totalGitHistoryWeighted

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

	for _, author := range authors {
		authorChangesWeighted := float64(contributions[author].GitInformation.Changes.Addition)*additionWeight + float64(contributions[author].GitInformation.Changes.Deletion)*deletionWeight
		totalChangesWeighted := float64(totalChanges.Addition)*additionWeight + float64(totalChanges.Deletion)*deletionWeight
		changesPercentage := authorChangesWeighted / totalChangesWeighted

		authorGitHistoryWeighted := float64(contributions[author].GitInformation.Merges)*mergeWeight + float64(contributions[author].GitInformation.Commits)*commitWeight
		totalGitHistoryWeighted := float64(totalMerge)*mergeWeight + float64(totalCommit)*commitWeight
		gitHistoryPercentage := authorGitHistoryWeighted / totalGitHistoryWeighted

		authorIssuesWeighted := float64(len(contributions[author].PlatformInformation.IssueInformation.Author))*issueWeight + float64(getSumOfCommentsOfIssues(contributions[author].PlatformInformation.IssueInformation.Author))*issueCommentsWeight + float64(contributions[author].PlatformInformation.IssueInformation.Commenter)*issueCommenterWeight
		totalIssuesWeighted := float64(totalIssues)*issueWeight + float64(totalComments)*issueCommentsWeight + float64(totalCommenter)*issueCommenterWeight
		issuesPercentage := authorIssuesWeighted / totalIssuesWeighted

		authorPullRequestsWeighted := getAuthorPullRequestValue(contributions[author].PlatformInformation.PullRequestInformation.Author)*pullRequestAuthorWeight + float64(contributions[author].PlatformInformation.PullRequestInformation.Reviewer)*pullRequestReviewerWeight
		totalPullRequestsWeighted := totalPullRequestsValue*pullRequestAuthorWeight + float64(totalPullRequestsReviews)*pullRequestReviewerWeight
		pullRequestPercentage := authorPullRequestsWeighted / totalPullRequestsWeighted

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
