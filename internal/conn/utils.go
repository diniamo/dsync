package conn

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func ParseRemote(remote string) (user, host, port, path string) {
	start := 0
	if strings.HasPrefix(remote, "ssh://") {
		start = len("ssh://")
	}

	for i := start; i < len(remote); i++ {
		c := remote[i]
		if host == "" {
			if user == "" && c == '@' {
				user = remote[start:i]
				start = i + 1
			} else if c == ':' {
				host = remote[start:i]
				start = i + 1
			} else if c == '/' {
				host = remote[start:i]
				path = remote[i+1:]
				return user, host, port, path
			}
		} else { // if port == ""
			if c == '/' {
				port = remote[start:i]
				path = remote[i+1:]
				return user, host, port, path
			}
		}
	}

	if host == "" {
		host = remote[start:]
	} else if port == "" {
		port = remote[start:]
	}
	
	return user, host, port, path
}

func PromptPassword(user, host string) (password string, err error) {
	fmt.Printf("(%s@%s) Password: ", user, host)
	bytes, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	
	if err == nil {
		password = string(bytes)
	}

	return
}
