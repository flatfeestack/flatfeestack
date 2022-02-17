package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	libgit "github.com/libgit2/git2go/v33"
	"sort"
	"sync"
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

	gitAnalysisStart := time.Now()

	//https://github.com/libgit2/git2go/issues/729
	head, _ := repo.Head()
	commit, _ := repo.LookupCommit(head.Target())
	seen := map[string]bool{}
	loop(repo, authorMap, commit, seen)
	time.Sleep(time.Minute * 2)
	fmt.Printf("%v", authorMap)
	fmt.Printf("---> git analysis in %dms\n", time.Since(gitAnalysisStart).Milliseconds())

	return authorMap, nil
}

var commitCounter = 0
var lock2 = sync.Mutex{}

func loop(repo *libgit.Repository, authorMap map[string]Contribution, commit *libgit.Commit, seen map[string]bool) {

	parentCommit := commit.Parent(0)

	for commit != nil && parentCommit != nil {

		lock2.Lock()
		if _, found := seen[commit.Id().String()]; found {
			lock2.Unlock()
			return
		}
		seen[commit.Id().String()] = true
		lock2.Unlock()
		fmt.Printf("search %v\n", commit.Id().String())

		commitCounter++
		numParents := commit.ParentCount()
		if numParents > 0 {
			collectInfo(commit, parentCommit, authorMap, repo)
			if numParents > 1 {
				//This is a merge, go for the commits
				for i := uint(1); i < numParents; i++ {
					parentI := commit.Parent(i)
					if parentI != nil {
						go loop(repo, authorMap, parentI, seen)
					}
				}
			}
		}
		commit = parentCommit
		parentCommit = commit.Parent(0)
		fmt.Printf("counter: %v\n", commitCounter)

	}
}

var lock = sync.Mutex{}

func collectInfo(commit *libgit.Commit, parentCommit *libgit.Commit, authorMap map[string]Contribution, repo *libgit.Repository) error {
	author := commit.Author()
	commiter := commit.Committer()

	merge := 0
	commitNr := 1

	parentTree, err := parentCommit.Tree()
	if err != nil {
		return err
	}
	commitTree, err := commit.Tree()
	if err != nil {
		return err
	}

	start := time.Now()
	diff, err := repo.DiffTreeToTree(parentTree, commitTree, nil)
	fmt.Printf("time: %v\n", time.Since(start).Milliseconds())
	if err != nil {
		return err
	}
	stats, err := diff.Stats()
	if err != nil {
		return err
	}

	fmt.Printf("commit: %v/%v | %v\n", stats.Insertions(), stats.Deletions(), commit.Summary())

	if commit.ParentCount() > 1 {
		merge = 1
		commitNr = 0
	}

	lock.Lock()
	defer lock.Unlock()
	if author != nil {
		if _, found := authorMap[author.Email]; !found {
			c1 := Contribution{
				Names:    []string{author.Name},
				Addition: stats.Insertions(),
				Deletion: stats.Deletions(),
				Merges:   merge,
				Commits:  commitNr,
			}
			authorMap[author.Email] = c1
		} else {
			names := authorMap[author.Email].Names
			if !contains(authorMap[author.Email].Names, author.Name) {
				names = append(names, author.Name)
				sort.Strings(names)
			}
			authorMap[author.Email] = Contribution{
				Names:    names,
				Addition: authorMap[author.Email].Addition + stats.Insertions(),
				Deletion: authorMap[author.Email].Deletion + stats.Deletions(),
				Merges:   authorMap[author.Email].Merges + merge,
				Commits:  authorMap[author.Email].Commits + commitNr,
			}
		}
	}
	if commiter != nil {
		if _, found := authorMap[commiter.Email]; !found {
			c1 := Contribution{
				Names:    []string{commiter.Name},
				Addition: int(float64(stats.Insertions()) * mergedLinesWeight),
				Deletion: int(float64(stats.Deletions()) * mergedLinesWeight),
				Merges:   merge,
				Commits:  commitNr,
			}
			authorMap[commiter.Email] = c1
		} else {
			names := authorMap[commiter.Email].Names
			if !contains(authorMap[commiter.Email].Names, commiter.Name) {
				names = append(names, commiter.Name)
				sort.Strings(names)
			}
			authorMap[commiter.Email] = Contribution{
				Names:    names,
				Addition: authorMap[commiter.Email].Addition + int(float64(stats.Insertions())*mergedLinesWeight),
				Deletion: authorMap[commiter.Email].Deletion + int(float64(stats.Deletions())*mergedLinesWeight),
				Merges:   authorMap[commiter.Email].Merges + merge,
				Commits:  authorMap[commiter.Email].Commits + commitNr,
			}
		}
	}

	return nil
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
