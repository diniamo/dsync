package virtfs

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kr/fs"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type RemoteFS struct {
	base string
	client *ssh.Client
	ftp *sftp.Client
}

func (r RemoteFS) Rel(path string) (ret string) {
	ret, _ = filepath.Rel(r.base, path)
	return
}

func (r RemoteFS) Walk() *fs.Walker {
	return r.ftp.Walk(r.base)
}

func (r RemoteFS) Lstat(path string) (os.FileInfo, error) {
	return r.ftp.Lstat(filepath.Join(r.base, path))
}

func (r RemoteFS) Mkdir(path string, perm os.FileMode) error {
	path = filepath.Join(r.base, path)
	
	err := r.ftp.Mkdir(path)
	if err != nil {
		return err
	}
	
	return r.Chmod(path, perm)
}

func (r RemoteFS) Chmod(path string, mode os.FileMode) error {
	return r.ftp.Chmod(filepath.Join(r.base, path), mode)
}

func (r RemoteFS) Create(path string) (io.ReadWriteCloser, error) {
	return r.ftp.Create(filepath.Join(r.base, path))
}

func (r RemoteFS) Open(path string) (io.ReadCloser, error) {
	return r.ftp.Open(filepath.Join(r.base, path))
}

func (r RemoteFS) Chmtime(path string, mtime time.Time) error {
	return r.ftp.Chtimes(filepath.Join(r.base, path), time.Time{}, mtime)
}

func (r RemoteFS) Close() error {
	return nil
}
