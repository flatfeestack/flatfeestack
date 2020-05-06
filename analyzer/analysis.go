package main

import (
	"git-contribution/models"
	"github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing/object"
	"github.com/go-git/go-git/storage/memory"
	"os"
	"time"
)

func analyzeRepository(src string, since time.Time, until time.Time) (map[models.Contributor]models.CommitChange, error) {
	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:      src,
		Progress: os.Stdout,
	})
	if err != nil {
		return nil, err
	}

	authorMap := make(map[models.Contributor]models.CommitChange)

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
		author := models.Contributor{
			Name:  c.Author.Name,
			Email: c.Author.Email,
		}

		stats, err := c.Stats()
		if err != nil {
			return err
		}
		contribution := models.CommitChange{
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
			authorMap[author] = models.CommitChange{
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
