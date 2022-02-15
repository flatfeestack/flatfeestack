package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	libgit "github.com/libgit2/git2go/v33"
	log "github.com/sirupsen/logrus"
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

	gitAnalysisStart := time.Now()

	//https://github.com/libgit2/git2go/issues/729
	head, _ := repo.Head()
	commit, _ := repo.LookupCommit(head.Target())
	parentCommit := commit.Parent(0)
	lineAdd := 0
	lineDel := 0
	commitCounter := 0
	for commit != nil && parentCommit != nil {
		commitCounter++
		numParents := commit.ParentCount()
		if numParents > 0 {

			parentTree, _ := parentCommit.Tree()
			commitTree, _ := commit.Tree()

			log.Infof("a %v, %v", commit.Author().Name, commit.Author().When)
			log.Infof("c %v", commit.Committer().Name)

			diff, _ := repo.DiffTreeToTree(parentTree, commitTree, nil)
			max, _ := diff.Stats()

			lineAdd += max.Insertions()
			lineDel += max.Deletions()

			author := commit.Author()
			commiter := commit.Committer()

			merge := 0
			commitNr := 1

			if numParents > 1 {
				merge = 1
				commitNr = 0
			}

			if author != nil {
				if _, found := authorMap[author.Email]; !found {
					c1 := Contribution{
						Names:    []string{author.Name},
						Addition: lineAdd,
						Deletion: lineDel,
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
						Addition: authorMap[author.Email].Addition + lineAdd,
						Deletion: authorMap[author.Email].Deletion + lineDel,
						Merges:   authorMap[author.Email].Merges + merge,
						Commits:  authorMap[author.Email].Commits + commitNr,
					}
				}
			}
			if commiter != nil {
				if _, found := authorMap[commiter.Email]; !found {
					c1 := Contribution{
						Names:    []string{commiter.Name},
						Addition: int(float64(lineAdd) * mergedLinesWeight),
						Deletion: int(float64(lineDel) * mergedLinesWeight),
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
						Addition: authorMap[commiter.Email].Addition + int(float64(lineAdd)*mergedLinesWeight),
						Deletion: authorMap[commiter.Email].Deletion + int(float64(lineDel)*mergedLinesWeight),
						Merges:   authorMap[commiter.Email].Merges + merge,
						Commits:  authorMap[commiter.Email].Commits + commitNr,
					}
				}
			}

		}
		commit1 := commit.Parent(0)
		commit.Free()
		commit = commit1
		parentCommit = commit.Parent(0)
	}

	fmt.Printf("---> git analysis in %dms (%d commits)\n", time.Since(gitAnalysisStart).Milliseconds(), commitCounter)
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
