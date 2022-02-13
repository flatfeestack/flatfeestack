package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	libgit "github.com/libgit2/git2go/v33"
	"sort"
	"time"
)

type Contribution struct {
	Names    []string
	Addition int
	Deletion int
	Merges   int
	Commits  int
}

// analyzeRepositoryFromString manages the whole analysis process (opens the repository and initialized the analysis)
func analyzeRepositoryFromString(src string, since time.Time, until time.Time, branch string) (map[string]Contribution, error) {
	cloneUpdateStart := time.Now()
	repo, err := cloneOrUpdateRepository(src, branch)
	if err != nil {
		return nil, err
	}

	fmt.Printf("---> cloned/updated repository in %dms\n", time.Since(cloneUpdateStart).Milliseconds())
	return analyzeRepositoryFromRepository(repo, since, until)
}

// analyzeRepositoryFromRepository uses go-git to extract the metrics from the opened repository
func analyzeRepositoryFromRepository(repo *libgit.Repository, since time.Time, until time.Time) (map[string]Contribution, error) {
	authorMap := make(map[string]Contribution)

	var timeZeroValue time.Time
	var options git.LogOptions

	if since != timeZeroValue {
		options.Since = &since
	}

	if until != timeZeroValue {
		options.Until = &until
	}

	wlk, err := repo.Walk()

	wlk.Iterate(func(commit *libgit.Commit) bool {

		return true
	})

	commits, err := repo.Log(&options)
	if err != nil {
		return authorMap, err
	}

	commitCounter := 0

	gitAnalysisStart := time.Now()
	err = commits.ForEach(func(c *object.Commit) error {
		commitCounter++
		fmt.Printf("\033[2K\r%d commits", commitCounter)

		merge := 0
		commit := 1

		if len(c.ParentHashes) > 1 {
			merge = 1
			commit = 0
		}

		stats, err := c.Stats()
		if err != nil {
			if err != nil {
				return err
			}
		}

		lineAdd := 0
		lineDel := 0
		// count the lines if it's not a merge, otherwise use a factor
		if merge == 0 {
			for index := range stats {
				lineAdd += stats[index].Addition
				lineDel += stats[index].Deletion
			}
		} else {
			for index := range stats {
				lineAdd += int(float64(stats[index].Addition) * mergedLinesWeight)
				lineDel += int(float64(stats[index].Deletion) * mergedLinesWeight)
			}
		}

		if _, found := authorMap[c.Author.Email]; !found {
			c1 := Contribution{
				Names:    []string{c.Author.Name},
				Addition: lineAdd,
				Deletion: lineDel,
				Merges:   merge,
				Commits:  commit,
			}

			authorMap[c.Author.Email] = c1
		} else {
			names := authorMap[c.Author.Email].Names
			if !contains(authorMap[c.Author.Email].Names, c.Author.Name) {
				names = append(names, c.Author.Name)
				sort.Strings(names)
			}
			authorMap[c.Author.Email] = Contribution{
				Names:    names,
				Addition: authorMap[c.Author.Email].Addition + lineAdd,
				Deletion: authorMap[c.Author.Email].Deletion + lineDel,
				Merges:   authorMap[c.Author.Email].Merges + merge,
				Commits:  authorMap[c.Author.Email].Commits + commit,
			}
		}

		return nil
	})

	fmt.Printf("---> git analysis in %dms (%d commits)\n", time.Since(gitAnalysisStart).Milliseconds(), commitCounter)

	if err != nil {
		return authorMap, err
	}
	return authorMap, nil
}

// weightContributions calculates the scores of the contributors by weighting the collected metrics (repository)
func weightContributions(contributions map[string]Contribution) ([]FlatFeeWeight, error) {
	weightContributionsStart := time.Now()

	authors := []FlatFeeWeight{}

	totalAdd := 0
	totalDel := 0
	totalMerge := 0
	totalCommit := 0

	for _, v := range contributions {
		totalAdd += v.Addition
		totalDel += v.Deletion
		totalMerge += v.Merges
		totalCommit += v.Commits
	}

	totalAmountOfAuthors := len(contributions)

	for email, contribution := range contributions {
		// calculation of changes category
		authorChangesWeighted := float64(contribution.Addition)*additionWeight + float64(contribution.Deletion)*deletionWeight
		totalChangesWeighted := float64(totalAdd)*additionWeight + float64(totalDel)*deletionWeight
		var changesPercentage float64
		if totalChangesWeighted == 0 {
			changesPercentage = 1.0 / float64(totalAmountOfAuthors)
		} else {
			changesPercentage = authorChangesWeighted / totalChangesWeighted
		}

		// calculation of git history category
		authorGitHistoryWeighted := float64(contribution.Merges)*mergeWeight + float64(contribution.Commits)*commitWeight
		totalGitHistoryWeighted := float64(totalMerge)*mergeWeight + float64(totalCommit)*commitWeight
		var gitHistoryPercentage float64
		if totalGitHistoryWeighted == 0 {
			gitHistoryPercentage = 1.0 / float64(totalAmountOfAuthors)
		} else {
			gitHistoryPercentage = authorGitHistoryWeighted / totalGitHistoryWeighted
		}

		authors = append(authors, FlatFeeWeight{
			Names:  contribution.Names,
			Email:  email,
			Weight: changesPercentage*changesWeight + gitHistoryPercentage*gitHistoryWeight,
		})
	}

	weightContributionsEnd := time.Now()
	fmt.Printf("---> weight contributions in %dms\n", weightContributionsEnd.Sub(weightContributionsStart).Milliseconds())
	return authors, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
