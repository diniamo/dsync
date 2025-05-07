package main

import (
	"fmt"
	"os"
	osUser "os/user"
	"path/filepath"

	"github.com/diniamo/dsync/internal/conn"
	"github.com/diniamo/dsync/internal/log"
	"github.com/diniamo/dsync/internal/walk"
)

const usage = "Usage: dsync <local path> <remote (path)>"
const remoteFormat = "Remote format: [scp://][user@]host[:port][/path]"

func currentUser() (user string, err error) {
	currentUser, err := osUser.Current()
	if err == nil {
		user = currentUser.Username
	}
	
	return
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "%s\n%s\n", usage, remoteFormat)
		os.Exit(1)
	}

	local := os.Args[1]
	remote := os.Args[2]

	local, err := filepath.Abs(local)
	if err != nil {
		log.Fatalf("Failed to convert local path to absolute: %s", err)
	}

	localInfo, err := os.Lstat(local)
	if err != nil {
		log.Fatalf("Failed to stat local target: %s", err)
	}
	

	user, host, port, remote := conn.ParseRemote(remote)
	if host == "" {
		log.FatalColor.Fprintf(os.Stderr, "%s is not a valid remote\n", remote)
		fmt.Fprintln(os.Stderr, remoteFormat)
		os.Exit(1)
	}
	if user == "" {
		user, err = currentUser()
		if err != nil {
			return
		}
	}
	if port == "" {
		port = "22"
	}
	if remote == "" {
		// The local path is expanded to absolute above
		remote = local
	}
	
	password, err := conn.PromptPassword(user, host)
	if err != nil {
		log.Fatalf("Failed to prompt for password: %s", err)
	}

	client, ftp, err := conn.EstablishConnection(user, host, port, password)
	if err != nil {
		log.Fatalf("Failed to establish ssh/sftp connection: %s", err)
	}
	defer client.Close()
	defer ftp.Close()


	var answer string
	fmt.Print("This operation may overwrite files on either side, since only their last modified times are considered. It's recommended to make backups.\nAre you sure you want to continue? [y/N] ")
	fmt.Scanln(&answer)
	if answer != "y" && answer != "Y" {
		return
	}

	
	walk.WalkLocal(local, ftp, remote)

	if !localInfo.IsDir() {
		return
	}

	// Another walker is necessary, because the previous one cannot handle cases
	// where a file exists on the remote, but not locally.
	walk.WalkRemote(ftp, remote, local)
}
