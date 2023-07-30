module github.com/thatoddmailbox/seafile-browse

go 1.16

require (
	github.com/BurntSushi/toml v1.0.0
	github.com/thatoddmailbox/fsbrowse v0.1.0
	github.com/thatoddmailbox/sftpfs v0.1.0
	golang.org/x/crypto v0.11.0
)

replace github.com/thatoddmailbox/fsbrowse => ../fsbrowse

replace github.com/thatoddmailbox/sftpfs => ../sftpfs
