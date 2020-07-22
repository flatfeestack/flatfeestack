package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
)

func CloneOrUpdateRepository(src string) (*git.Repository, error) {
	repo, err := UpdateRepository(src)
	if err != nil {
		return CloneRepository(src)
	}
	return repo, nil
}

func CloneRepository(src string) (*git.Repository, error) {
	folderName := src[8 : len(src)-4]
	return git.PlainClone(os.Getenv("GO_GIT_BASE_PATH")+"/"+folderName, false, &git.CloneOptions{
		URL:           src,
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", os.Getenv("GO_GIT_DEFAULT_BRANCH"))),
		SingleBranch:  true,
	})
}

func UpdateRepository(src string) (*git.Repository, error) {
	folderName := src[8 : len(src)-4]
	repo, err := git.PlainOpen(os.Getenv("GO_GIT_BASE_PATH") + "/" + folderName)
	if err != nil {
		return nil, err
	}
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
	})
	if err != nil && err.Error() != "already up-to-date" {
		return repo, err
	}
	return repo, nil
}
