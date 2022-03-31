package main

import (
	"log"
	"net/http"

	"github.com/thatoddmailbox/fsbrowse"
	"github.com/thatoddmailbox/seafile-browse/config"
	"github.com/thatoddmailbox/seafile-browse/seafile"
)

func main() {
	log.Println("seafile-browse")

	cfg, err := config.Load("config.toml")
	if err != nil {
		panic(err)
	}
	defer cfg.Close()

	storage := seafile.NewStorageWithFS(cfg.FS())
	repos, err := storage.ListRepoIDs()
	if err != nil {
		panic(err)
	}

	log.Println(repos)

	repo, err := storage.OpenRepo(repos[0])
	if err != nil {
		panic(err)
	}
	log.Println(repo)

	commit, err := repo.GetLatestCommit()
	if err != nil {
		panic(err)
	}
	log.Println(commit)

	fsys, err := commit.GetFS()
	if err != nil {
		panic(err)
	}
	log.Println(fsys)

	log.Fatal(http.ListenAndServe(":9253", fsbrowse.FileServer(fsys)))
}
