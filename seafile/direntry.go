package seafile

import "io/fs"

type DirEntry struct {
	d *direntInternal
}

func (d *DirEntry) Name() string {
	return d.d.Name
}

func (d *DirEntry) IsDir() bool {
	return d.Type().IsDir()
}

func (d *DirEntry) Type() fs.FileMode {
	i, err := d.Info()
	if err != nil {
		panic(err)
	}

	return i.Mode().Type()
}

func (d *DirEntry) Info() (fs.FileInfo, error) {
	return &FileInfo{
		d: d.d,
	}, nil
}
