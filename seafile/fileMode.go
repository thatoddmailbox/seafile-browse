package seafile

import (
	"io/fs"
	"time"
)

type FileInfo struct {
	f *File
}

func (i *FileInfo) Name() string {
	return i.f.d.Name
}

func (i *FileInfo) Size() int64 {
	return i.f.d.Size
}

func (i *FileInfo) Mode() fs.FileMode {
	// TODO: implement
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
