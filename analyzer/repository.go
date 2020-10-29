package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
)

// CloneOrUpdateRepository clones the repository if it is not already on the disk, else update it
func CloneOrUpdateRepository(src string, branch string) (*git.Repository, error) {
	// check if the repository can successfully be updated
	repo, err := UpdateRepository(src, branch)
	// if not try to clone it
	if err != nil {
		fmt.Println(err.Error())
		return CloneRepository(src, branch)
	}
	return repo, nil
}

// CloneRepository clones the repository as a single branch repository with the desired branch
func CloneRepository(src string, branch string) (*git.Repository, error) {
	folderName := src[8 : len(src)-4]
	// clone just one branch
	return git.PlainClone(os.Getenv("GO_GIT_BASE_PATH")+"/"+folderName, false, &git.CloneOptions{
		URL:           src,
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		SingleBranch:  true,
	})
}

// CloneRepository updates the repository and checks out the desired branch
func UpdateRepository(src string, branch string) (*git.Repository, error) {
	folderName := src[8 : len(src)-4]
	repo, err := git.PlainOpen(os.Getenv("GO_GIT_BASE_PATH") + "/" + folderName)
	if err != nil {
		return nil, err
	}
	// take just one remote branch and assign it to a local branch with the same name
	firstRefSpecArgument := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{config.RefSpec(firstRefSpecArgument)},
	})
	if err != nil && err.Error() != "already up-to-date" {
		return repo, err
	}
	w, err := repo.Worktree()

	if err != nil {
		return repo, err
	}

	// checkout the correct branch
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		Force:  true,
	})

	return repo, nil
}
