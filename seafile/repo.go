package seafile

import (
	"io/fs"
	"log"
	"path"
)

type Repo struct {
	id   string
	fsys fs.FS
	s    *Storage
}

// GetLatestCommit returns the most recent Commit to the Repo.
func (r *Repo) GetLatestCommit() (*Commit, error) {
	commitPath := path.Join("storage", "commits", r.id)

	if r.s.haveOptimization {
		lastCommit := r.s.latestCommits[r.id]
		log.Println("ok", r.id, lastCommit)

		if lastCommit != "" {
			path := lastCommit[:2] + "/" + lastCommit[2:]

			f, err := r.fsys.Open(commitPath + "/" + path)
			if err != nil {
				return nil, err
			}

			commit, err := newCommit(r.id, r.fsys, f)
			if err == nil {
				return commit, nil
			}
		}
	}

	var latestCommit *Commit
	err := fs.WalkDir(r.fsys, commitPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		f, err := r.fsys.Open(path)
		if err != nil {
			return err
		}

		commit, err := newCommit(r.id, r.fsys, f)
		if err != nil {
			return err
		}

		if latestCommit == nil || commit.CTime > latestCommit.CTime {
			latestCommit = commit
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return latestCommit, nil
}

func newRepo(id string, fsys fs.FS, s *Storage) *Repo {
	return &Repo{
		id:   id,
		fsys: fsys,
		s:    s,
	}
}
