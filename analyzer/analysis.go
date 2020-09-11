package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"time"
)

func analyzeRepository(src string, since time.Time, until time.Time) (map[Contributor]Contribution, error) {
	repo, err := CloneOrUpdateRepository(src)
	if err != nil {
		return nil, err
	}

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
		return nil, err
	}

	err = commits.ForEach(func(c *object.Commit) error {
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
		for index := range stats {
			changes.Addition += stats[index].Addition
			changes.Deletion += stats[index].Deletion
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
	if err != nil {
		return nil, err
	}
	return authorMap, nil
}

func weightContributions(contributions map[Contributor]Contribution) (map[Contributor]FlatFeeWeight, error) {

	// Parameters
	additionWeight := 0.4
	deletionWeight := 0.3
	commitWeight := 0.2
	mergeWeight := 0.1

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
		additionPercentage := float64(contributions[author].Changes.Addition) / float64(totalChanges.Addition)
		deletionPercentage := float64(contributions[author].Changes.Deletion) / float64(totalChanges.Deletion)
		mergePercentage := float64(contributions[author].Merges) / float64(totalMerge)
		commitPercentage := float64(contributions[author].Commits) / float64(totalCommit)

		authorMap[author] = FlatFeeWeight{
			Contributor: author,
			Weight:      additionPercentage*additionWeight + deletionPercentage*deletionWeight + mergePercentage*mergeWeight + commitPercentage*commitWeight,
		}
	}

	return authorMap, nil

}

func weightContributionsWithPlatformInformation(contributions map[Contributor]ContributionWithPlatformInformation) (map[Contributor]FlatFeeWeight, error) {

	// Parameter

	// Category "Changes" devided into additions and deletions. both must sum up to 1
	additionWeight := 0.7
	deletionWeight := 0.3

	// Category "GitHistory" devided into commits and merges. both must sum up to 1
	commitWeight := 0.7
	mergeWeight := 0.3

	// Intercategory weights between categories Changes and Githistory. all must sum up to 1
	changesWeight := 0.8
	gitHistoryWeight := 0.2

	authorMap := make(map[Contributor]FlatFeeWeight)

	totalChanges := CommitChange{
		Addition: 0,
		Deletion: 0,
	}
	totalMerge := 0
	totalCommit := 0
	var authors []Contributor

	for _, v := range contributions {
		authors = append(authors, v.GitInformation.Contributor)
		totalChanges = CommitChange{
			Addition: totalChanges.Addition + v.GitInformation.Changes.Addition,
			Deletion: totalChanges.Deletion + v.GitInformation.Changes.Deletion,
		}
		totalMerge += v.GitInformation.Merges
		totalCommit += v.GitInformation.Commits
	}

	for _, author := range authors {
		authorChangesWeighted := float64(contributions[author].GitInformation.Changes.Addition)*additionWeight + float64(contributions[author].GitInformation.Changes.Deletion)*deletionWeight
		totalChangesWeighted := float64(totalChanges.Addition)*additionWeight + float64(totalChanges.Deletion)*deletionWeight
		changesPercentage := authorChangesWeighted / totalChangesWeighted

		authorGitHistoryWeighted := float64(contributions[author].GitInformation.Merges)*mergeWeight + float64(contributions[author].GitInformation.Commits)*commitWeight
		totalGitHistoryWeighted := float64(totalMerge)*mergeWeight + float64(totalCommit)*commitWeight
		gitHistoryPercentage := authorGitHistoryWeighted / totalGitHistoryWeighted

		authorMap[author] = FlatFeeWeight{
			Contributor: author,
			Weight:      changesPercentage*changesWeight + gitHistoryPercentage*gitHistoryWeight,
		}
	}

	return authorMap, nil

}
