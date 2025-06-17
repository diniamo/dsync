package virtfs

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/kr/fs"
)

type LocalFS struct {
	base string
}

func (l LocalFS) Rel(path string) (ret string) {
	ret, _ = filepath.Rel(l.base, path)
	return
}

func (l LocalFS) Walk() *fs.Walker {
	return fs.Walk(l.base)
}

func (l LocalFS) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(filepath.Join(l.base, path))
}

func(l LocalFS) Mkdir(path string, perm os.FileMode) error {
	return os.Mkdir(filepath.Join(l.base, path), perm)
}

func (l LocalFS) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(filepath.Join(l.base, path), mode)
}

func (l LocalFS) Create(path string) (io.ReadWriteCloser, error) {
	return os.Create(filepath.Join(l.base, path))
}

func (l LocalFS) Open(path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(l.base, path))
}

func (l LocalFS) Chmtime(path string, mtime time.Time) error {
	return os.Chtimes(filepath.Join(l.base, path), time.Time{}, mtime)
}

func (l LocalFS) Close() error {
	return nil
}
