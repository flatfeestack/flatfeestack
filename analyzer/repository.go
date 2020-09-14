package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
)

func CloneOrUpdateRepository(src string) (*git.Repository, error) {
	repo, err := UpdateRepository(src)
	if err != nil {
		fmt.Println(err.Error())
		return CloneRepository(src)
	}
	return repo, nil
}

func CloneRepository(src string) (*git.Repository, error) {
	folderName := src[8 : len(src)-4]
	fmt.Println("the branch is", os.Getenv("GO_GIT_DEFAULT_BRANCH"))
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
	fmt.Println("this is really executed", os.Getenv("GO_GIT_DEFAULT_BRANCH"))
	firstRefSpecArgument := fmt.Sprintf("*/%s:*/%s", os.Getenv("GO_GIT_DEFAULT_BRANCH"), os.Getenv("GO_GIT_DEFAULT_BRANCH"))
	//secondRefSpecArgument := fmt.Sprintf("HEAD:refs/heads/%s", os.Getenv("GO_GIT_DEFAULT_BRANCH"))
	err = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{config.RefSpec(firstRefSpecArgument)},
	})
	//RefSpecs: []config.RefSpec{"refs/heads/master-3.x:refs/remotes/origin/master-3.x", "HEAD:refs/heads/master-3.x"},
	if err != nil && err.Error() != "already up-to-date" {
		return repo, err
	}
	//w, err := repo.Worktree()
	//
	//err = w.Checkout(&git.CheckoutOptions{
	//	Branch: plumbing.ReferenceName(fmt.Sprintf("refs/*/%s", os.Getenv("GO_GIT_DEFAULT_BRANCH"))),
	//	Force: true,
	//})

	return repo, err
}
