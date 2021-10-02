package seafile

import "io/fs"

type FS struct {
	c    *Commit
	root *File
}

// Open opens the named file.
func (sfsys *FS) Open(name string) (fs.File, error) {
	return sfsys.root.open(name)
}

func newFS(c *Commit) (*FS, error) {
	sfsys := FS{
		c: c,
	}

	root, err := newFile(&sfsys, c.RootID, nil)
	if err != nil {
		return nil, err
	}
	sfsys.root = root

	return &sfsys, nil
}
