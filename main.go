package main

import (
	"fmt"
	"html"
	"io/fs"
	"log"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/thatoddmailbox/fsbrowse"
	"github.com/thatoddmailbox/seafile-browse/config"
	"github.com/thatoddmailbox/seafile-browse/seafile"
)

func renderRepoList(w http.ResponseWriter, r *http.Request, snapshotName string) {

}

type repoInfo struct {
	repoIDs []string
	repos   map[string]*seafile.Repo
	repoFSs map[string]fs.FS
}

var allRepoInfo map[string]repoInfo = map[string]repoInfo{}

func getRepoInfoForSnapshot(snapshot string, cfg *config.Config) ([]string, map[string]*seafile.Repo, map[string]fs.FS) {
	ri, exists := allRepoInfo[snapshot]
	if exists {
		return ri.repoIDs, ri.repos, ri.repoFSs
	}

	f := cfg.FS()
	if snapshot != "" {
		var err error
		f, err = fs.Sub(cfg.SnapshotFS(), path.Join(snapshot, cfg.Path()))
		if err != nil {
			panic(err)
		}
	}

	storage := seafile.NewStorageWithFS(f)
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

		repoFSs[repoID] = latestFS
	}

	allRepoInfo[snapshot] = repoInfo{repoIDs, repos, repoFSs}

	return repoIDs, repos, repoFSs
}

func main() {
	log.Println("seafile-browse")

	cfg, err := config.Load("config.toml")
	if err != nil {
		panic(err)
	}
	defer cfg.Close()

	snapshots := []string{}
	if cfg.HaveSnapshots() {
		sf := cfg.SnapshotFS()
		entries, err := fs.ReadDir(sf, ".")
		if err != nil {
			panic(err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				snapshots = append(snapshots, entry.Name())
			}
		}

		sort.Strings(snapshots)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]
		path = strings.TrimSuffix(path, "/")

		activeSnapshot := ""

		pathParts := strings.Split(path, "/")
		if len(pathParts) > 0 {
			if strings.HasPrefix(pathParts[0], "snapshot:") {
				activeSnapshot = strings.SplitN(pathParts[0], ":", 2)[1]
				pathParts = pathParts[1:]
			}
		}

		notice := ""
		if cfg.HaveSnapshots() {
			if activeSnapshot == "" {
				notice = "You are viewing the latest data."
			} else {
				notice = "You are viewing snapshot <code>" + html.EscapeString(activeSnapshot) + "</code>."
			}

			notice += " <a href=\"snapshots/\">View snapshots</a><br><br>"
		}

		_, repos, repoFSs := getRepoInfoForSnapshot(activeSnapshot, cfg)

		if len(pathParts) == 0 || path == "" {
			// repo list
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><head><title>seafile-browse</title><body>")
			if cfg.HaveSnapshots() {
				fmt.Fprint(w, notice)
			}
			fmt.Fprintf(w, "Select a library:<ul>")
			for repoID, _ := range repos {
				fmt.Fprintf(w, "<li><a href=\"%s/\">%s</li>", html.EscapeString(repoID), html.EscapeString(repoID))
			}
			fmt.Fprintf(w, "</ul></body></html>")
			return
		}

		// repo contents

		repoID := pathParts[0]
		repoPath := pathParts[1:]

		if repoID == "snapshots" {
			// snapshots list
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, "<html><head><title>seafile-browse</title><body>")
			fmt.Fprintf(w, "<a href=\"/\">View latest data</a><br><br>")
			fmt.Fprintf(w, "Or, select a snapshot:<ul>")
			for _, snapshot := range snapshots {
				fmt.Fprintf(w, "<li><a href=\"/snapshot:%s/\">%s</li>", html.EscapeString(snapshot), html.EscapeString(snapshot))
			}
			fmt.Fprintf(w, "</ul></body></html>")
			return
		}

		repoFS, repoExists := repoFSs[repoID]
		if !repoExists {
			fmt.Fprintf(w, "Repo ID invalid")
			return
		}
		// kinda janky, we reuse the request but rewrite its path
		r.URL.Path = strings.Join(repoPath, "/")
		fsbrowse.ServeHTTPStateless(w, r, repoFS, "", notice)
	})

	port := 9253
	log.Printf("Listening on port %d...", port)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
