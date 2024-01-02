package main

import (
	"io/ioutil"
	"log"

	vcs "github.com/unix-world/smartgoext/vcs-repo"
)

func main() {

	remote := "https://github.com/Masterminds/vcs.git"
	local, _ := ioutil.TempDir("", "go-vcs")
	repo, err := vcs.NewRepo(remote, local)
	if err != nil {
		log.Println("ERR: NewRepo", err)
		return
	}

	err = repo.Get()
	if err != nil {
		log.Println("ERR", "Unable to checkout SVN repo", err)
		return
	}

	ci, err := repo.Current()
	if err != nil {
		log.Println("ERR: Current", err)
		return
	}
	log.Println(ci)

}
