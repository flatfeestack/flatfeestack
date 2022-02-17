package main

import (
	libgit "github.com/libgit2/git2go/v33"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

// CloneOrUpdateRepository clones the repository if it is not already on the disk, else update it
func cloneOrUpdateRepository(src string, branch string) (*libgit.Repository, error) {
	// check if the repository can successfully be updated
	repo, err := updateRepository(src, branch)
	// if not try to clone it
	if err != nil {
		log.Debugf("probably new repo, %v", err)
		return cloneRepository(src, branch)
	}
	return repo, nil
}

// CloneRepository clones the repository as a single branch repository with the desired branch
func cloneRepository(src string, branch string) (*libgit.Repository, error) {
	folderName := src[8 : len(src)-4]
	// clone just one branch
	// git clone https://github.com/torvalds/linux.git --single-branch --shallow-since="3 months ago" -n

	cmd := exec.Command("git", "clone", src, "--branch", branch, "--single-branch", `--shallow-since="3 months ago"`, "-n", ".")

	cmd.Dir = opts.GitBasePath + "/" + folderName
	err := os.MkdirAll(opts.GitBasePath+"/"+folderName, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	r, err := libgit.OpenRepository(opts.GitBasePath + "/" + folderName)
	if err != nil {
		return nil, err
	}

	/*r, err := git.PlainClone(opts.GitBasePath+"/"+folderName, false, &git.CloneOptions{
		URL:           src,
		Progress:      os.Stdout,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
		SingleBranch:  true,
	})
	if err != nil {
		return nil, err
	}*/
	if err := os.WriteFile(opts.GitBasePath+"/"+folderName+".date", []byte(time.Now().Format(time.RFC3339)), 0666); err != nil {
		return nil, err
	}
	return r, nil

}

// CloneRepository updates the repository and checks out the desired branch
func updateRepository(src string, branch string) (*libgit.Repository, error) {
	folderName := src[8 : len(src)-4]

	r, err := libgit.OpenRepository(opts.GitBasePath + "/" + folderName)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(opts.GitBasePath + "/" + folderName + ".date")
	if err != nil {
		return nil, err
	}
	_, err = time.Parse(time.RFC3339, string(content))
	if err != nil {
		return nil, err
	}
	/*if lastUpdate.After(time.Now().AddDate(0, 0, -1)) {
		return r, nil
	}*/
	if err := os.WriteFile(opts.GitBasePath+"/"+folderName+".date", []byte(time.Now().Format(time.RFC3339)), 0666); err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "fetch", `--shallow-since="3 months ago"`)
	cmd.Dir = opts.GitBasePath + "/" + folderName
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	r, err = libgit.OpenRepository(opts.GitBasePath + "/" + folderName)
	if err != nil {
		return nil, err
	}
	return r, err
}
