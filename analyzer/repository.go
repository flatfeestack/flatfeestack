package main

import (
	git "github.com/libgit2/git2go/v33"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// CloneOrUpdateRepository clones the repository if it is not already on the disk, else update it
func cloneOrUpdateRepository(location ...string) (*git.Repository, error) {
	// check if the repository can successfully be updated
	repo, err := updateRepository(location)
	// if not try to clone it
	if err != nil {
		log.Debugf("probably new repo, %v", err)
		return cloneRepository(location)
	}
	return repo, nil
}

// CloneRepository clones the repository as a single branch repository with the desired branch
func cloneRepository(location []string) (*git.Repository, error) {
	u, err := url.Parse(location[0])
	if err != nil {
		return nil, err
	}
	folderName := u.Host + strings.ReplaceAll(u.Path, "/", "")
	folderName = strings.ReplaceAll(folderName, ".", "")
	// clone just one branch
	// git clone https://github.com/torvalds/linux.git --shallow-since="3 months ago" -n
	// git clone git://git.kernel.org/pub/scm/linux/kernel/git/next/linux-next.git --shallow-since="3 months ago" -n
	cmd := exec.Command("git", "clone", location[0], `--shallow-since="6 months ago"`, "-n", ".")
	cmd.Dir = opts.GitBasePath + "/" + folderName
	err = os.MkdirAll(opts.GitBasePath+"/"+folderName, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	for i := 1; i < len(location); i++ {
		cmd := exec.Command("git", "remote", "add", "r"+strconv.Itoa(i), location[i])
		cmd.Dir = opts.GitBasePath + "/" + folderName
		err = cmd.Run()
		if err != nil {
			return nil, err
		}
		cmd = exec.Command("git", "fetch", "r"+strconv.Itoa(i), `--shallow-since="3 months ago"`, "-a")
		cmd.Dir = opts.GitBasePath + "/" + folderName
		err = cmd.Run()
		if err != nil {
			return nil, err
		}
	}

	r, err := git.OpenRepository(opts.GitBasePath + "/" + folderName)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(opts.GitBasePath+"/"+folderName+".date", []byte(time.Now().Format(time.RFC3339)), 0666); err != nil {
		return nil, err
	}
	return r, nil

}

// CloneRepository updates the repository and checks out the desired branch
func updateRepository(location []string) (*git.Repository, error) {
	u, err := url.Parse(location[0])
	if err != nil {
		return nil, err
	}
	folderName := u.Host + strings.ReplaceAll(u.Path, "/", "")
	folderName = strings.ReplaceAll(folderName, ".", "")

	r, err := git.OpenRepository(opts.GitBasePath + "/" + folderName)
	if err != nil {
		return nil, err
	}

	content, err := ioutil.ReadFile(opts.GitBasePath + "/" + folderName + ".date")
	if err != nil {
		return nil, err
	}
	lastUpdate, err := time.Parse(time.RFC3339, string(content))
	if err != nil {
		return nil, err
	}
	if lastUpdate.After(time.Now().AddDate(0, 0, -1)) {
		return r, nil
	}
	if err := os.WriteFile(opts.GitBasePath+"/"+folderName+".date", []byte(time.Now().Format(time.RFC3339)), 0666); err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "fetch", `--shallow-since="3 months ago"`, "-a")
	cmd.Dir = opts.GitBasePath + "/" + folderName
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	cmd = exec.Command("git", "update-ref", "HEAD", "refs/remotes/origin/HEAD")
	cmd.Dir = opts.GitBasePath + "/" + folderName
	err = cmd.Run()
	if err != nil {
		return nil, err
	}
	//if default branch is changed after the repo has been cloned, this is not updated automatically, but can easily be fixed locally
	//https://stackoverflow.com/questions/51274430/change-from-master-to-a-new-default-branch-git
	cmd = exec.Command("git", "remote", "set-head", "origin", "-a")
	cmd.Dir = opts.GitBasePath + "/" + folderName
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	r, err = git.OpenRepository(opts.GitBasePath + "/" + folderName)
	if err != nil {
		return nil, err
	}
	return r, err
}
