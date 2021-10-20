package seafile

import (
	"bufio"
	"encoding/json"
	"io/fs"
)

type Commit struct {
	repoID string
	fsys   fs.FS

	CommitID    string `json:"commit_id"`
	RootID      string `json:"root_id"`
	Description string `json:"description"`
	CTime       uint64 `json:"ctime"`
	ParentID    string `json:"parent_id"`
}

// GetFS returns an FS of the Repo's state at the given Commit.
func (c *Commit) GetFS() (*FS, error) {
	return newFS(c)
}

func newCommit(repoID string, fsys fs.FS, f fs.File) (*Commit, error) {
	c := Commit{
		repoID: repoID,
		fsys:   fsys,
	}

	defer f.Close()

	err := json.NewDecoder(bufio.NewReader(f)).Decode(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
