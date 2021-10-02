package seafile

import (
	"io/fs"
	"path"
)

type Repo struct {
	id   string
	fsys fs.FS
}

// GetLatestCommit returns the most recent Commit to the Repo.
func (r *Repo) GetLatestCommit() (*Commit, error) {
	commitPath := path.Join("storage", "commits", r.id)

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

func newRepo(id string, fsys fs.FS) *Repo {
	return &Repo{
		id:   id,
		fsys: fsys,
	}
}
