package main

import (
	"fmt"
	git "github.com/libgit2/git2go/v34"
	log "github.com/sirupsen/logrus"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"math"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
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
	changesWeight    = 0.5
	gitHistoryWeight = 0.5
)

var (
	defaultTime      time.Time
	excludeEmails    = []string{"noreply@github.com"}
	includedTrailers = []string{"Signed-off-by", "Reviewed-by"}
	//https://github.com/mcnijman/go-emailaddress/blob/master/emailaddress.go
	findAmpersandRegexp = regexp.MustCompile("(?i)([&][A-Z0-9%]+[;])")
	findCommonRegexp    = regexp.MustCompile("(?i)([A-Z0-9._%+-]+@[A-Z0-9.-]+\\.[A-Z]{2,24})")
	rfc5322             = "(?i)(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9]))\\.){3}(?:(2(5[0-5]|[0-4][0-9])|1[0-9][0-9]|[1-9]?[0-9])|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])"
	validRfc5322Regexp  = regexp.MustCompile(fmt.Sprintf("^%s*$", rfc5322))
	caser               = cases.Title(language.Und)
)

// small contributors get more, to be encouraged, also the committer that
// have most lines know the repository the best
func smallCommitter(input int) float64 {
	//https://www.desmos.com/calculator/k0f7hv7hg5
	if input == 0 {
		return 0
	}
	return float64(input) * ((2 / math.Pow(1.02, float64(input))) + 1)
}

// analyzeRepository manages the whole analysis process (opens the repository and initialized the analysis)
func analyzeRepository(startTime time.Time, stopTime time.Time, location string) (map[string]Contribution, error) {
	cloneUpdateStart := time.Now()
	repo, err := cloneOrUpdate(location)
	if err != nil {
		return nil, err
	}
	defer repo.Free()

	log.Infof("---> cloned/updated repository in %dms\n", time.Since(cloneUpdateStart).Milliseconds())

	authorMap := map[string]Contribution{}
	gitAnalysisStart := time.Now()
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

	err = revWalk.PushHead()
	if err != nil {
		return nil, err
	}

	err = revWalk.Iterate(func(commit *git.Commit) bool {
		wg.Add(1)
		loop(repo, &commitCounter, authorMap, authorLock, commit, seen, seenLock, wg, startTime, stopTime)
		return true
	})

	//since we have not the full history, revision walker throws an error if we don't find an old parent. This is ok, ignore it
	if gErr, ok := err.(*git.GitError); ok {
		if gErr.Code != git.ErrorCodeNotFound && gErr.Class != git.ErrorClassOdb {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	wg.Wait()

	log.Infof("---> #%v git analysis in %dms\n", commitCounter, time.Since(gitAnalysisStart).Milliseconds())
	return authorMap, nil
}

func loop(repo *git.Repository, commitCounter *int64, authorMap map[string]Contribution, authorLock *sync.Mutex, commit *git.Commit, seen map[string]bool, seenLock *sync.Mutex, wg *sync.WaitGroup, startTime time.Time, stopTime time.Time) {
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
			collectInfo(commit, parentCommit, authorMap, authorLock, repo, startTime, stopTime)
		}
		wg.Add(1)
		go loop(repo, commitCounter, authorMap, authorLock, parentCommit, seen, seenLock, wg, startTime, stopTime)
	}
}

func collectInfo(commit *git.Commit, parentCommit *git.Commit, authorMap map[string]Contribution, authorLock *sync.Mutex, repo *git.Repository, startTime time.Time, stopTime time.Time) error {
	start := time.Now()

	if expired(commit, startTime, stopTime) {
		return nil
	}

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

	author := commit.Author()
	committer := commit.Committer()
	ts, err := git.MessageTrailers(commit.Message())
	if err != nil {
		return err
	}

	log.Debugf("commit: %v (%v/%v) [%v/%v] | %v, time: %v",
		commit.Id().String(), author.Email, committer.Email, stats.Insertions(),
		stats.Deletions(), commit.Summary(), time.Since(start).Milliseconds())

	parentCount := commit.ParentCount()
	fillAuthorMap(author, committer, ts, parentCount, authorLock, authorMap, stats)

	return nil
}

func fillAuthorMap(author *git.Signature, committer *git.Signature, ts []git.Trailer, parentCount uint, authorLock *sync.Mutex, authorMap map[string]Contribution, stats *git.DiffStats) {
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
		addToMap(author.Email, author.Name, authorMap, stats, authorFactor, merge)
		if committer != nil && !contains(excludeEmails, committer.Email) && author.Email != committer.Email {
			addToMap(committer.Email, committer.Name, authorMap, stats, mergedLinesWeight, merge)
		}
		for _, v := range ts {
			if contains(includedTrailers, v.Key) {
				n := strings.ReplaceAll(v.Value, "@@", "@")
				e := findEmail(n)
				i := strings.IndexByte(v.Value, '<')
				if e == "" && i < 0 { //no email, and no <>
					e = "no-email-found@flatfeestack.com"
					log.Warnf("no email found in [%v]", v.Value)
				} else if e != "" && i < 0 {
					var err error
					n, err = emailToName(e)
					if err != nil {
						log.Warnf("no name found, using email as name [%v]", v.Value)
					}
				} else {
					n = strings.TrimSpace(v.Value[:i])
				}

				n = findAmpersandRegexp.ReplaceAllString(n, "")
				addToMap(e, n, authorMap, stats, mergedLinesWeight, merge)
			}
		}
	}
}

func addToMap(authorEmail string, authorName string, authorMap map[string]Contribution, stats *git.DiffStats, authorFactor float64, merge int) {
	if strings.Contains(authorEmail, "users.noreply.github.com") {
		return
	}

	c1, _ := authorMap[authorEmail]

	names := authorMap[authorEmail].Names
	if !contains(authorMap[authorEmail].Names, authorName) {
		names = append(names, authorName)
		sort.Strings(names)
	}

	authorMap[authorEmail] = Contribution{
		Names:    names,
		Addition: c1.Addition + int(float64(stats.Insertions())*authorFactor),
		Deletion: c1.Deletion + int(float64(stats.Deletions())*authorFactor),
		Merges:   c1.Merges + merge,
		Commits:  c1.Commits + 1,
	}
}

// weightContributions calculates the scores of the contributors by weighting the collected metrics (repository)
func weightContributions(contributions map[string]Contribution) ([]FlatFeeWeight, error) {
	var result []FlatFeeWeight
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
			authorChangesWeighted := (smallCommitter(contribution.Addition) * additionWeight) + (smallCommitter(contribution.Deletion) * deletionWeight)
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

		log.Infof("authors: %v=+%v/-%v, c:%v,m:%v", contribution.Names, contribution.Addition, contribution.Deletion, contribution.Commits, contribution.Merges)
	}

	return result, nil
}

// clones the repository if it is not already on the disk, else update it
func cloneOrUpdate(gitUrl string) (*git.Repository, error) {

	p, err := pathName(gitUrl)
	if err != nil {
		return nil, err
	}

	o := git.CheckoutOptions{Strategy: git.CheckoutForce}
	repo, err := git.Clone(gitUrl, p, &git.CloneOptions{CheckoutOptions: o})

	//directory already existing,
	if err != nil {
		// still log the error, could be the issue is something else than "directory does already exist"
		log.Errorln(err)

		repo, err = git.OpenRepository(p)
		if err != nil {
			return nil, err
		}

		remote, err := repo.Remotes.Lookup("origin")
		if err != nil {
			return nil, err
		}
		defer remote.Free()

		err = remote.Fetch([]string{}, nil, "")
		if err != nil {
			return nil, err
		}
		rh, err := remote.Ls()
		if err != nil {
			return nil, err
		}

		remoteCommit, err := repo.LookupCommit(rh[0].Id)
		if err != nil {
			return nil, err
		}
		defer remoteCommit.Free()

		err = repo.ResetToCommit(remoteCommit, git.ResetHard, &o)
		if err != nil {
			return nil, err
		}
	}

	return repo, nil
}

func pathName(gitUrl string) (string, error) {
	alreadyExists, err := exists(gitUrl)
	if err != nil {
		return "", err
	}

	if alreadyExists {
		return gitUrl, nil
	}

	u, err := url.Parse(gitUrl)
	if err != nil {
		return "", err
	}
	folderName := u.Host + strings.ReplaceAll(u.Path, "/", "")
	folderName = strings.ReplaceAll(folderName, ".", "")
	return opts.GitBasePath + "/" + folderName, nil
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

func expired(commit *git.Commit, startTime time.Time, stopTime time.Time) bool {
	if (startTime != defaultTime && commit.Author().When.Before(startTime) && commit.Committer().When.Before(startTime)) ||
		(stopTime != defaultTime && commit.Author().When.After(stopTime) && commit.Committer().When.After(stopTime)) {
		log.Debugf("time reached: %v -  %v/%v", commit.Id().String(), commit.Author().When, commit.Committer().When)
		return true
	}
	return false
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

func emailToName(email string) (string, error) {
	i := strings.IndexByte(email, '@')
	if i < 0 {
		return "", fmt.Errorf("not an email %v", email)
	}
	name := email[:i]
	name = strings.ReplaceAll(name, ".", " ")
	name = caser.String(name)
	return name, nil
}

func findEmail(haystack string) string {
	result := findCommonRegexp.FindString(haystack)
	if result != "" {
		if validRfc5322Regexp.MatchString(result) {
			return result
		}
	}
	return ""
}

// https://stackoverflow.com/a/10510783
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
