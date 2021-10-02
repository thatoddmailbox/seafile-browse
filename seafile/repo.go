package seafile

import "io/fs"

type Repo struct {
	id   string
	fsys fs.FS
}

func newRepo(id string, fsys fs.FS) *Repo {
	return &Repo{
		id:   id,
		fsys: fsys,
	}
}
