package config

import (
	"errors"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/thatoddmailbox/sftpfs"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	Location struct {
		Local *struct {
			Path         string
			SnapshotPath string
			SQLFilePath  string
		}
		SFTP *struct {
			Host         string
			User         string
			Password     string
			Path         string
			SnapshotPath string
			SQLFilePath  string
		}
	}

	path string

	f       fs.FS
	sf      fs.FS
	sqlPath string
	sftpFS  *sftpfs.Client
}

func (c *Config) initFS() error {
	var err error

	if c.Location.Local != nil {
		c.f = os.DirFS(c.Location.Local.Path)
		if c.Location.Local.SnapshotPath != "" {
			c.sf = os.DirFS(c.Location.Local.SnapshotPath)
		}
		c.path = c.Location.Local.Path
		c.sqlPath = c.Location.Local.SQLFilePath
		return nil
	}

	if c.Location.SFTP != nil {
		c.sftpFS, err = sftpfs.Dial("tcp", c.Location.SFTP.Host, &ssh.ClientConfig{
			User: c.Location.SFTP.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(c.Location.SFTP.Password),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		})
		if err != nil {
			return err
		}

		c.f = c.sftpFS
		c.path = c.Location.SFTP.Path

		if c.Location.SFTP.SnapshotPath != "" {
			subFS, err := fs.Sub(c.sftpFS, c.Location.SFTP.SnapshotPath)
			if err != nil {
				return err
			}

			c.sf = subFS
		}

		c.sqlPath = c.Location.SFTP.SQLFilePath

		return nil
	}

	return errors.New("config: could not determine location type")
}

func (c *Config) Close() {
	if c.Location.SFTP != nil {
		c.sftpFS.Close()
	}
}

func (c *Config) Path() string {
	return c.path
}

func (c *Config) FS() fs.FS {
	return c.f
}

func (c *Config) SnapshotFS() fs.FS {
	return c.sf
}

func (c *Config) SQLFilePath() string {
	return c.sqlPath
}

func (c *Config) HaveSnapshots() bool {
	return c.sf != nil
}

func Load(path string) (*Config, error) {
	c := Config{}
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return nil, err
	}

	locationTypeCount := 0

	if c.Location.Local != nil {
		locationTypeCount += 1
	}
	if c.Location.SFTP != nil {
		locationTypeCount += 1
	}

	if locationTypeCount == 0 {
		return nil, errors.New("config: no location defined")
	}
	if locationTypeCount > 1 {
		return nil, errors.New("config: mutiple locations defined")
	}

	err = c.initFS()
	if err != nil {
		return nil, err
	}

	return &c, nil
}
