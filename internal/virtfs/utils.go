package virtfs

import (
	"fmt"
	"net"
	"os"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func parseSpec(spec string) (user, host, path string) {
	if spec[0] == '/' {
		path = spec
		return
	}

	start := 0
	for i, c := range spec {
		if user == "" && c == '@' {
			user = spec[start:i]
			start = i + 1
		} else if host == "" && c == ':' {
			host = spec[start:i]
			start = i + 1
		}
	}

	if host == "" {
		host = spec[start:]
	} else {
		path = spec[start:]
	}
	
	return
}

func promptPassword(user, host string) (password string, err error) {
	fmt.Printf("(%s@%s) Password: ", user, host)
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	
	if err == nil {
		password = string(raw)
	}

	return
}

func connect(user, host, port, password string) (client *ssh.Client, ftp *sftp.Client, err error) {
	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(string(password))},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	
	client, err = ssh.Dial("tcp", net.JoinHostPort(host, port), &config)
	if err != nil {
		return
	}

	ftp, err = sftp.NewClient(client)
	if err != nil {
		client.Close()
	}
	
	return
}
