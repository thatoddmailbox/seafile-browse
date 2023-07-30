package seafile

import (
	"bufio"
	"errors"
	"io/fs"
	"strings"
)

var ErrGarbageRepo = errors.New("seafile: garbage repo not supported")
var ErrVirtualRepo = errors.New("seafile: virtual repo not supported")

type Storage struct {
	rootFsys fs.FS
	fsys     fs.FS

	haveOptimization bool
	latestCommits    map[string]string
	garbageRepos     map[string]bool
	virtualRepos     map[string]bool
	repoNames        map[string]string
	repoOwners       map[string]string
}

type RepoInfo struct {
	ID      string
	Name    string
	Owner   string
	Virtual bool
	Garbage bool
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
	_, garbage := s.garbageRepos[repoID]
	if garbage {
		return nil, ErrVirtualRepo
	}
	_, virtual := s.virtualRepos[repoID]
	if virtual {
		return nil, ErrVirtualRepo
	}

	return newRepo(repoID, s.fsys, s), nil
}

// GetRepoInfo gets a RepoInfo struct describing the Repo with the given ID.
func (s *Storage) GetRepoInfo(repoID string) (RepoInfo, error) {
	return RepoInfo{
		ID:      repoID,
		Name:    s.repoNames[repoID],
		Owner:   s.repoOwners[repoID],
		Garbage: s.garbageRepos[repoID],
		Virtual: s.virtualRepos[repoID],
	}, nil
}

// ParseSQLFile reads the SQL file at the given path and uses that for optimization.
func (s *Storage) ParseSQLFile(sqlPath string) error {
	sqlFile, err := s.rootFsys.Open(sqlPath)
	if err != nil {
		return err
	}
	defer sqlFile.Close()

	// TODO: this is a terrible way to parse this and is super brittle :)

	s.latestCommits = map[string]string{}
	s.garbageRepos = map[string]bool{}
	s.virtualRepos = map[string]bool{}
	s.repoNames = map[string]string{}
	s.repoOwners = map[string]string{}

	scan := bufio.NewScanner(sqlFile)
	scan.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scan.Scan() {
		line := scan.Text()

		if !strings.HasPrefix(line, "INSERT INTO ") {
			continue
		}

		// TODO: this code is bad :)
		if strings.HasPrefix(line, "INSERT INTO `Branch` VALUES ") {
			valuesString := strings.TrimPrefix(line, "INSERT INTO `Branch` VALUES ")
			valuesString = strings.TrimSuffix(valuesString, ";")

			valuesStrings := strings.Split(valuesString, "),(")

			for _, valueString := range valuesStrings {
				valueParts := strings.Split(valueString, ",")

				branch := valueParts[1]
				if branch != "'master'" {
					// something weird, ignore it
					continue
				}

				repoID := strings.Trim(valueParts[2], "'")
				lastCommitID := strings.Trim(valueParts[3], "')")

				s.latestCommits[repoID] = lastCommitID
			}
		} else if strings.HasPrefix(line, "INSERT INTO `RepoInfo` VALUES ") {
			valuesString := strings.TrimPrefix(line, "INSERT INTO `RepoInfo` VALUES ")
			valuesString = strings.TrimSuffix(valuesString, ";")

			valuesStrings := strings.Split(valuesString, "),(")

			for _, valueString := range valuesStrings {
				valueParts := strings.Split(valueString, ",")

				repoID := strings.Trim(valueParts[1], "'")
				repoName := strings.Trim(valueParts[2], "')")

				s.repoNames[repoID] = repoName
			}
		} else if strings.HasPrefix(line, "INSERT INTO `RepoOwner` VALUES ") {
			valuesString := strings.TrimPrefix(line, "INSERT INTO `RepoOwner` VALUES ")
			valuesString = strings.TrimSuffix(valuesString, ";")

			valuesStrings := strings.Split(valuesString, "),(")

			for _, valueString := range valuesStrings {
				valueParts := strings.Split(valueString, ",")

				repoID := strings.Trim(valueParts[1], "'")
				repoOwner := strings.Trim(valueParts[2], "')")

				s.repoOwners[repoID] = repoOwner
			}
		} else if strings.HasPrefix(line, "INSERT INTO `VirtualRepo` VALUES ") {
			valuesString := strings.TrimPrefix(line, "INSERT INTO `VirtualRepo` VALUES ")
			valuesString = strings.TrimSuffix(valuesString, ";")

			valuesStrings := strings.Split(valuesString, "),(")

			for _, valueString := range valuesStrings {
				valueParts := strings.Split(valueString, ",")

				repoID := strings.Trim(valueParts[1], "'")

				s.virtualRepos[repoID] = true
			}
		} else if strings.HasPrefix(line, "INSERT INTO `GarbageRepos` VALUES ") {
			valuesString := strings.TrimPrefix(line, "INSERT INTO `GarbageRepos` VALUES ")
			valuesString = strings.TrimSuffix(valuesString, ";")

			valuesStrings := strings.Split(valuesString, "),(")

			for _, valueString := range valuesStrings {
				valueParts := strings.Split(valueString, ",")

				repoID := strings.Trim(valueParts[1], "'")

				s.garbageRepos[repoID] = true
			}
		}
	}

	s.haveOptimization = true

	return scan.Err()
}

// NewStorageWithFS creates a new Storage with the given fs.FS.
func NewStorageWithFS(fsys fs.FS) *Storage {
	return NewStorageWithFSSubpath(fsys, ".")
}

// NewStorageWithFSSubpath creates a new Storage with the given fs.FS.
func NewStorageWithFSSubpath(fsys fs.FS, subpath string) *Storage {
	sub, err := fs.Sub(fsys, subpath)
	if err != nil {
		panic(err)
	}

	return &Storage{
		fsys:     sub,
		rootFsys: fsys,
	}
}
