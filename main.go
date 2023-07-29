package main

import (
	"fmt"
	"html"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

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
	repoIDs, err := storage.ListRepoIDs()
	if err != nil {
		panic(err)
	}

	repos := map[string]*seafile.Repo{}
	repoFSs := map[string]fs.FS{}
	for _, repoID := range repoIDs {
		repos[repoID], err = storage.OpenRepo(repoID)
		if err != nil {
			panic(err)
		}

		commit, err := repos[repoID].GetLatestCommit()
		if err != nil {
			panic(err)
		}

		latestFS, err := commit.GetFS()
		if err != nil {
			panic(err)
		}

		repoLatestFSHandler[repoID] = fsbrowse.FileServer(latestFS)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			// repo list
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><head><title>seafile-browse</title><body>Select a library:<ul>")
			for repoID, _ := range repos {
				fmt.Fprintf(w, "<li><a href=\"%s/\">%s</li>", html.EscapeString(repoID), html.EscapeString(repoID))
			}
			fmt.Fprintf(w, "</ul></body></html>")
			return
		}

		// repo contents
		path := r.URL.Path[1:]
		path = strings.TrimSuffix(path, "/")

		pathParts := strings.Split(path, "/")
		repoID := pathParts[0]
		repoPath := pathParts[1:]

		repoFS, repoExists := repoFSs[repoID]
		if !repoExists {
			fmt.Fprintf(w, "Repo ID invalid")
			return
		}

		// kinda janky, we reuse the request but rewrite its path
		r.URL.Path = strings.Join(repoPath, "/")
		fsbrowse.ServeHTTPStateless(w, r, repoFS, "", "")
	})

	port := 9253
	log.Printf("Listening on port %d...", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
