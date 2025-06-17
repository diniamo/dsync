package virtfs

import (
	"errors"
	"io"
	"os"
	osUser "os/user"
	"path/filepath"
	"time"

	log "github.com/diniamo/glog"
	"github.com/kr/fs"
)

type VirtualFS interface {
	Rel(string) string

	Walk() *fs.Walker
	Lstat(string) (os.FileInfo, error)
	Mkdir(string, os.FileMode) error
	Chmod(string, os.FileMode) error
	Create(string) (io.ReadWriteCloser, error)
	Open(string) (io.ReadCloser, error)
	Chmtime(string, time.Time) error

	Close() error
}

func New(spec, port string) (VirtualFS, error) {
	user, host, path := parseSpec(spec)

	log.Notef("pars: %s -> %s, %s, %s", spec, user, host, path)
	if path == "" {
		return nil, errors.New("missing path")
	}
	
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	if host == "" {
		if _, err = os.Lstat(path); err != nil {
			return nil, err
		}

		return LocalFS{path}, nil
	} else {
		if user == "" {
			currentUser, err := osUser.Current()
			if err != nil {
				return nil, err
			}
			
			user = currentUser.Username
		}

		password, err := promptPassword(user, host)
		if err != nil {
			return nil, err
		}

		client, ftp, err := connect(user, host, port, password)
		if err != nil {
			return nil, err
		}

		if _, err = ftp.Lstat(path); os.IsNotExist(err) {
			return nil, err
		}

		return RemoteFS{path, client, ftp}, nil
	}
}
