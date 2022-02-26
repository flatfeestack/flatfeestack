package main

import (
	"fmt"
	libgit "github.com/libgit2/git2go/v33"
	log "github.com/sirupsen/logrus"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Contribution struct {
	Names    []string
	Addition int
	Deletion int
	Merges   int
	Commits  int
}

const (
	// Represents the factor of the total changed lines
	// with witch the merger gets rewarded for merging the branch.
	// Changed lines in normal commits are considered with factor 1
	// while the changed lines in merges (summary of the size of the merge)
	// are considered with this factor.
	mergedLinesWeight = 0.1
	// Category "Changes" divided into additions and deletions.
	// Both must sum up to 1
	additionWeight = 0.7
	deletionWeight = 0.3
	// Category "GitHistory" divided into commits and merges.
	// Both must sum up to 1
	commitWeight = 0.7
	mergeWeight  = 0.3
	// Intercategory weights between categories Changes and Githistory.
	// All must sum up to 1.
	// Only when platformInformation IS NOT considered
	changesWeight    = 0.7
	gitHistoryWeight = 0.3
)

var (
	defaultTime time.Time
)

//small contributors get more, to be encouraged, also the committer that
//have most lines know the repository the best
func smallCommitter(input int) float64 {
	//https://www.desmos.com/calculator/k0f7hv7hg5
	if input == 0 {
		return 0
	}
	return (2 / math.Pow(1.02, float64(input))) + 1
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
func analyzeRepositoryFromRepository(repo *libgit.Repository, startTime time.Time, stopTime time.Time) (map[string]Contribution, error) {
	gitAnalysisStart := time.Now()
	authorMap := map[string]Contribution{}
	authorLock := &sync.Mutex{}
	seen := map[string]bool{}
	seenLock := &sync.Mutex{}
	commitCounter := int64(0)
	wg := &sync.WaitGroup{}

	revWalk, err := repo.Walk()
	if err != nil {
		return nil, err
	}
	defer revWalk.Free()

	// Start out at the head
	err = revWalk.PushHead()
	if err != nil {
		return nil, err
	}

	err = revWalk.Iterate(func(commit *libgit.Commit) bool {
		if (startTime != defaultTime && commit.Author().When.Before(startTime) && commit.Committer().When.Before(startTime)) &&
			(stopTime != defaultTime && commit.Author().When.After(stopTime) && commit.Committer().When.After(stopTime)) {
			fmt.Printf("time reached: %v", commit.Id().String())
			return false
		}
		wg.Add(1)
		loop(repo, &commitCounter, authorMap, authorLock, commit, seen, seenLock, wg)
		return true
	})

	//since we have not the full history, revision walker throws an error if we don't find an old parent. This is ok, ignore it
	if gErr, ok := err.(*libgit.GitError); ok {
		if gErr.Code != libgit.ErrorCodeNotFound && gErr.Class != libgit.ErrorClassOdb {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	wg.Wait()

	fmt.Printf("---> git analysis in %dms\n", time.Since(gitAnalysisStart).Milliseconds())

	return authorMap, nil
}

func alreadyProcessed(commit string, seen map[string]bool, seenLock *sync.Mutex) bool {
	seenLock.Lock()
	defer seenLock.Unlock()
	if _, found := seen[commit]; found {
		return true
	}
	seen[commit] = true
	return false
}

func loop(repo *libgit.Repository, commitCounter *int64, authorMap map[string]Contribution, authorLock *sync.Mutex, commit *libgit.Commit, seen map[string]bool, seenLock *sync.Mutex, wg *sync.WaitGroup) {
	defer commit.Free()
	defer wg.Done()

	if alreadyProcessed(commit.Id().String(), seen, seenLock) {
		return
	}

	atomic.AddInt64(commitCounter, 1)
	numParents := commit.ParentCount()

	for i := uint(0); i < numParents; i++ {
		parentCommit := commit.Parent(i)
		if parentCommit == nil {
			continue
		}
		if i == 0 { //if it's a merge, the author gets only credit for the parent 0
			collectInfo(commit, parentCommit, authorMap, authorLock, repo, commitCounter)
		}
		wg.Add(1)
		go loop(repo, commitCounter, authorMap, authorLock, parentCommit, seen, seenLock, wg)
	}
}

func collectInfo(commit *libgit.Commit, parentCommit *libgit.Commit, authorMap map[string]Contribution, authorLock *sync.Mutex, repo *libgit.Repository, commitCounter *int64) error {
	start := time.Now()

	parentTree, err := parentCommit.Tree()
	if err != nil {
		return err
	}
	defer parentTree.Free()

	commitTree, err := commit.Tree()
	if err != nil {
		return err
	}
	defer commitTree.Free()

	diff, err := repo.DiffTreeToTree(parentTree, commitTree, nil)
	if err != nil {
		return err
	}
	defer diff.Free()

	stats, err := diff.Stats()
	defer stats.Free()
	if err != nil {
		return err
	}

	log.Infof("commit: %v (%d) [%v/%v] | %v, time: %v\n", commit.Id().String(), *commitCounter, stats.Insertions(), stats.Deletions(), commit.Summary(), time.Since(start).Milliseconds())

	author := commit.Author()
	committer := commit.Committer()
	parentCount := commit.ParentCount()
	fillAuthorMap(author, committer, parentCount, authorLock, authorMap, stats)

	return nil
}

func fillAuthorMap(author *libgit.Signature, committer *libgit.Signature, parentCount uint, authorLock *sync.Mutex, authorMap map[string]Contribution, stats *libgit.DiffStats) {
	authorFactor := 1.0
	merge := 0
	if parentCount > 1 {
		//author is commiter (author merged)
		authorFactor = mergedLinesWeight
		merge = 1
	}

	authorLock.Lock()
	defer authorLock.Unlock()
	if author != nil {
		addToMap(author, authorMap, stats, authorFactor, merge)
		if committer != nil && committer.Email != "noreply@github.com" && author.Email != committer.Email {
			addToMap(committer, authorMap, stats, mergedLinesWeight, merge)
		}
	}

}

func addToMap(author *libgit.Signature, authorMap map[string]Contribution, stats *libgit.DiffStats, authorFactor float64, merge int) {
	c1, _ := authorMap[author.Email]

	names := authorMap[author.Email].Names
	if !contains(authorMap[author.Email].Names, author.Name) {
		names = append(names, author.Name)
		sort.Strings(names)
	}
	authorMap[author.Email] = Contribution{
		Names:    names,
		Addition: int(float64(stats.Insertions()) * authorFactor),
		Deletion: int(float64(stats.Deletions()) * authorFactor),
		Merges:   c1.Merges + merge,
		Commits:  c1.Commits + 1,
	}
}

// weightContributions calculates the scores of the contributors by weighting the collected metrics (repository)
func weightContributions(contributions map[string]Contribution) ([]FlatFeeWeight, error) {
	result := []FlatFeeWeight{}
	var totalAdd, totalDel float64
	var totalMerge, totalCommit int

	for _, v := range contributions {
		totalAdd += smallCommitter(v.Addition)
		totalDel += smallCommitter(v.Deletion)
		totalMerge += v.Merges
		totalCommit += v.Commits
	}

	for email, contribution := range contributions {
		// calculation of changes category
		totalChangesWeighted := totalAdd*additionWeight + totalDel*deletionWeight
		var changesPercentage float64
		if totalChangesWeighted == 0 {
			changesPercentage = 0
		} else {
			authorChangesWeighted := smallCommitter(contribution.Addition)*additionWeight + smallCommitter(contribution.Deletion)*deletionWeight
			changesPercentage = authorChangesWeighted / totalChangesWeighted
		}

		// calculation of git history category
		var gitHistoryPercentage float64
		totalGitHistoryWeighted := float64(totalMerge)*mergeWeight + float64(totalCommit)*commitWeight
		if totalGitHistoryWeighted == 0 {
			gitHistoryPercentage = 0
		} else {
			authorGitHistoryWeighted := float64(contribution.Merges)*mergeWeight + float64(contribution.Commits)*commitWeight
			gitHistoryPercentage = authorGitHistoryWeighted / totalGitHistoryWeighted
		}

		result = append(result, FlatFeeWeight{
			Names:  contribution.Names,
			Email:  email,
			Weight: changesPercentage*changesWeight + gitHistoryPercentage*gitHistoryWeight,
		})
	}

	return result, nil
}

func contains(s []string, e string) bool {
	if s == nil {
		return false
	}
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
