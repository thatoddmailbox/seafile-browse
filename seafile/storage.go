package seafile

import "io/fs"

type Storage struct {
	fsys fs.FS
}

// ListRepoIDs returns a list of all repo IDs.
func (s *Storage) ListRepoIDs() ([]string, error) {
	entries, err := fs.ReadDir(s.fsys, "storage/commits")
	if err != nil {
		return nil, err
	}

	result := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			result = append(result, entry.Name())
		}
	}
	return result, nil
}

// OpenRepo opens the Repo with the given ID.
func (s *Storage) OpenRepo(repoID string) (*Repo, error) {
	return newRepo(repoID, s.fsys), nil
}

// NewStorageWithFS creates a new Storage with the given fs.FS.
func NewStorageWithFS(fsys fs.FS) *Storage {
	return &Storage{
		fsys: fsys,
	}
}
