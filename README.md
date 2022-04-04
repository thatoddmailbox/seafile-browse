# seafile-browse [![Build](https://github.com/thatoddmailbox/seafile-browse/actions/workflows/build.yml/badge.svg)](https://github.com/thatoddmailbox/seafile-browse/actions/workflows/build.yml)

A command-line client to browse a [Seafile](https://seafile.com) data repository. Uses [fsbrowse](https://github.com/thatoddmailbox/fsbrowse) to provide a web UI to look through the files, and [sftpfs](https://github.com/thatoddmailbox/sftpfs) to allow accessing a Seafile repository stored on a remote machine. Also provides an [fs.FS](https://pkg.go.dev/io/fs#FS)-implementing client in the `seafile` module, which could hypothetically be used directly by something else.

Requires Go 1.16 or newer.

You must have the path of your repository. It can be either over SSH or local.

To configure, create a `config.toml` in the same directory as `seafile-browse`.
For local file access:
```
[location.local]
Path = "path/to/seafile-data"
```

For SFTP access:
```
[location.sftp]
Host = "some.remote.host:22"
User = "username"
Password = "password123"
Path = "path/to/seafile-data"
```