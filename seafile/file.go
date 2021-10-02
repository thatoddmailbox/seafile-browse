package seafile

import (
	"compress/zlib"
	"encoding/json"
	"errors"
	"io/fs"
	"path"
)

const typeFile = 1
const typeDir = 3

type dirent struct {
	ID       string `json:"id"`
	Mode     uint32 `json:"mode"`
	Modifier string `json:"modifier"`
	MTime    uint64 `json:"mtime"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
}

type fileInternal struct {
	// only for files
	BlockIDs []string `json:"block_ids`

	// only for dirs
	Dirents []dirent `json:"dirents"`

	Type    int `json:"type"`
	Version int `json:"version"`
}

type File struct {
	seafileFsys *FS
	fileID      string

	i fileInternal
	d *dirent
}

func (f *File) open(name string) (*File, error) {
	// TODO: fix
	for _, dirent := range f.i.Dirents {
		if dirent.Name == name {
			return newFile(f.seafileFsys, dirent.ID, &dirent)
		}
	}

	return nil, fs.ErrNotExist
}

func (f *File) Stat() (fs.FileInfo, error) {
	return &FileInfo{
		f: f,
	}, errors.New("not implemented")
}

func (f *File) Read(b []byte) (int, error) {
	return 0, errors.New("not implemented")
}

func (f *File) Close() error {
	// not really anything to do
	return nil
}

func newFile(seafileFsys *FS, fileID string, d *dirent) (*File, error) {
	ret := File{
		seafileFsys: seafileFsys,
		fileID:      fileID,

		d: d,
	}

	fsPath := path.Join("storage", "fs", seafileFsys.c.repoID, fileID[:2], fileID[2:])
	f, err := seafileFsys.c.fsys.Open(fsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := zlib.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	err = json.NewDecoder(r).Decode(&ret.i)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}
