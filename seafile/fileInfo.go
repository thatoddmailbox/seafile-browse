package seafile

import (
	"io/fs"
	"time"
)

const modeIsDir = (1 << 14)

type FileInfo struct {
	d *direntInternal
}

func (i *FileInfo) Name() string {
	if i.d == nil {
		return ""
	}

	return i.d.Name
}

func (i *FileInfo) Size() int64 {
	if i.d == nil {
		return 0
	}

	return i.d.Size
}

func (i *FileInfo) Mode() fs.FileMode {
	if i.d == nil {
		return fs.ModeDir
	}

	if (i.d.Mode & modeIsDir) != 0 {
		return fs.ModeDir
	}

	return 0
}

func (i *FileInfo) ModTime() time.Time {
	// TODO: implement
	return time.Now()
}

func (i *FileInfo) IsDir() bool {
	return i.Mode().IsDir()
}

func (i *FileInfo) Sys() interface{} {
	return nil
}
