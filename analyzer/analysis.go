package main

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"os"
	"time"
)

func analyzeRepository(src string, since time.Time, until time.Time) (map[Contributor]CommitChange, error) {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:      src,
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}

	authorMap := make(map[Contributor]CommitChange)

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

		stats, err := c.Stats()
		if err != nil {
			return err
		}
		contribution := CommitChange{
			Addition: 0,
			Deletion: 0,
		}
		for index := range stats {
			contribution.Addition += stats[index].Addition
			contribution.Deletion += stats[index].Deletion
		}

		if _, found := authorMap[author]; !found {
			authorMap[author] = contribution
		} else {
			authorMap[author] = CommitChange{
				Addition: authorMap[author].Addition + contribution.Addition,
				Deletion: authorMap[author].Deletion + contribution.Deletion,
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return authorMap, nil
}
