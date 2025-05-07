package conn

import (
	"fmt"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func EstablishConnection(user, host, port, password string) (client *ssh.Client, ftp *sftp.Client, err error) {
	config := ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(string(password))},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	
	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%s", host, port), &config)
	if err != nil {
		return
	}

	ftp, err = sftp.NewClient(client)
	if err != nil {
		client.Close()
		// return
	}
	
	return
}
