package seafile

import (
	"compress/zlib"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"path"
	"strings"
)

const typeFile = 1
const typeDir = 3

type direntInternal struct {
	ID       string `json:"id"`
	Mode     uint32 `json:"mode"`
	Modifier string `json:"modifier"`
	MTime    uint64 `json:"mtime"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
}

type fileInternal struct {
	// only for files
	BlockIDs []string `json:"block_ids"`

	// only for dirs
	Dirents []direntInternal `json:"dirents"`

	Type    int `json:"type"`
	Version int `json:"version"`
}

type File struct {
	seafileFsys *FS
	fileID      string

	i fileInternal
	d *direntInternal

	direntIdx int

	closed bool

	totalByteOffset int64
	blockRemaining  int64
	blockIdx        uint
	blockFile       fs.File
}

func (f *File) openSub(sub string) (*File, error) {
	for _, dirent := range f.i.Dirents {
		if dirent.Name == sub {
			return newFile(f.seafileFsys, dirent.ID, &dirent)
		}
	}

	return nil, fs.ErrNotExist
}

func (f *File) open(name string) (*File, error) {
	parts := strings.Split(name, "/")

	if len(name) == 0 {
		parts = []string{}
	}

	currentLevel := f
	var err error
	for _, part := range parts {
		if part == "." {
			continue
		}

		currentLevel, err = currentLevel.openSub(part)
		if err != nil {
			return nil, err
		}
	}

	return currentLevel, nil
}

func (f *File) Stat() (fs.FileInfo, error) {
	return &FileInfo{
		d: f.d,
	}, nil
}

func (f *File) openBlockIdx(i uint) (fs.File, error) {
	blockID := f.i.BlockIDs[f.blockIdx]
	blockPath := path.Join("storage", "blocks", f.seafileFsys.c.repoID, blockID[:2], blockID[2:])
	return f.seafileFsys.c.fsys.Open(blockPath)
}

func (f *File) Read(b []byte) (int, error) {
	if f.closed {
		return 0, fs.ErrClosed
	}

	if f.i.Type != typeFile {
		return 0, fs.ErrInvalid
	}

	totalRead := 0
	totalRequested := int64(len(b))
	totalRemaining := f.d.Size - int64(f.totalByteOffset)

	for totalRequested > 0 && totalRemaining > 0 {
		var err error

		if f.blockFile == nil {
			f.blockFile, err = f.openBlockIdx(f.blockIdx)
			if err != nil {
				return totalRead, err
			}

			blockFileInfo, err := f.blockFile.Stat()
			if err != nil {
				return totalRead, err
			}
			f.blockRemaining = blockFileInfo.Size()
		}

		n, err := f.blockFile.Read(b[totalRead:])
		totalRead += n
		if err != nil && err != io.EOF {
			return totalRead, err
		}

		// update counters
		f.totalByteOffset += int64(n)
		totalRequested -= int64(n)
		totalRemaining -= int64(n)
		f.blockRemaining -= int64(n)

		if f.blockRemaining == 0 || err == io.EOF {
			// onto the next block
			f.blockIdx++

			err := f.blockFile.Close()
			if err != nil {
				return totalRead, err
			}
			f.blockFile = nil
		}
	}

	var err error
	if totalRemaining == 0 {
		err = io.EOF
	}

	return totalRead, err
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	// TODO: this is not really efficient, ESPECIALLY for SeekCurrent/SeekEnd
	// TODO: this can be heavily optimized

	absoluteOffset := offset
	if whence == io.SeekCurrent {
		absoluteOffset = f.totalByteOffset + offset
	} else if whence == io.SeekEnd {
		absoluteOffset = f.d.Size + offset
	}

	if absoluteOffset < -1 {
		return f.totalByteOffset, errors.New("seafile: tried to seek before start")
	}
	if absoluteOffset > f.d.Size {
		absoluteOffset = f.d.Size
	}

	if f.blockFile != nil {
		err := f.blockFile.Close()
		if err != nil {
			return f.totalByteOffset, err
		}

		f.blockFile = nil
	}

	f.totalByteOffset = 0
	f.blockRemaining = 0
	f.blockIdx = 0

	var err error

	for f.totalByteOffset < absoluteOffset {
		f.blockFile, err = f.openBlockIdx(f.blockIdx)
		if err != nil {
			return f.totalByteOffset, err
		}

		blockFileInfo, err := f.blockFile.Stat()
		if err != nil {
			return f.totalByteOffset, err
		}
		f.blockRemaining = blockFileInfo.Size()

		distanceRemaining := absoluteOffset - f.totalByteOffset

		if f.blockRemaining < distanceRemaining {
			// we need to seek to the right place in this block
			// then get out of here

			blockFileSeek, ok := f.blockFile.(io.Seeker)
			if !ok {
				return f.totalByteOffset, errors.New("seafile: underlying fs.File does not implement io.Seeker")
			}

			blockOffset, err := blockFileSeek.Seek(distanceRemaining, io.SeekStart)
			if err != nil {
				return f.totalByteOffset, err
			}

			f.blockRemaining -= blockOffset
			break
		}

		// skip this block
		f.totalByteOffset += f.blockRemaining
		f.blockRemaining = 0
		f.blockIdx++

		err = f.blockFile.Close()
		if err != nil {
			return f.totalByteOffset, err
		}
		f.blockFile = nil
	}

	return f.totalByteOffset, nil
}

func (f *File) ReadDir(n int) ([]fs.DirEntry, error) {
	if f.i.Type != typeDir {
		return []fs.DirEntry{}, fs.ErrInvalid
	}

	result := []fs.DirEntry{}

	for i := range f.i.Dirents[f.direntIdx:] {
		if n > 0 && i == n {
			break
		}

		result = append(result, &DirEntry{
			d: &f.i.Dirents[i],
		})

		if n > 0 {
			f.direntIdx++
		}
	}

	return result, nil
}

func (f *File) Close() error {
	f.closed = true

	if f.blockFile != nil {
		err := f.blockFile.Close()
		if err != nil {
			return err
		}

		f.blockFile = nil
	}

	return nil
}

func newFile(seafileFsys *FS, fileID string, d *direntInternal) (*File, error) {
	ret := File{
		seafileFsys: seafileFsys,
		fileID:      fileID,

		d: d,

		closed: false,

		totalByteOffset: 0,
		blockIdx:        0,
	}

	if d != nil && d.ID == "0000000000000000000000000000000000000000" && ((d.Mode & modeIsDir) != 0) {
		// it's an empty directory, special case
		// TODO: is version right?
		ret.i.Dirents = []direntInternal{}
		ret.i.Type = typeDir
		ret.i.Version = 3
		return &ret, nil
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
