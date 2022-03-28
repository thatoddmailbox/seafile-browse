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
			Path string
		}
		SFTP *struct {
			Host     string
			User     string
			Password string
			Path     string
		}
	}

	f      fs.FS
	sftpFS *sftpfs.Client
}

func (c *Config) initFS() error {
	var err error

	if c.Location.Local != nil {
		c.f = os.DirFS(c.Location.Local.Path)
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

		if c.Location.SFTP.Path != "" {
			subFS, err := fs.Sub(c.sftpFS, c.Location.SFTP.Path)
			if err != nil {
				return err
			}

			c.f = subFS
		}

		return nil
	}

	return errors.New("config: could not determine location type")
}

func (c *Config) Close() {
	if c.Location.SFTP != nil {
		c.sftpFS.Close()
	}
}

func (c *Config) FS() fs.FS {
	return c.f
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
