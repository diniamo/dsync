package walk

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/diniamo/dsync/internal/log"

	"github.com/kr/fs"
	"github.com/pkg/sftp"
)

func WalkLocal(local string, ftp *sftp.Client, remote string) {
	for walker := fs.Walk(local); walker.Step(); {
		if walker.Err() != nil {
			log.Errorf("Failed to step local walker: %s", walker.Err())
			continue
		}

		localStat := walker.Stat()

		if !localStat.Mode().IsRegular() && !localStat.IsDir() {
			continue
		}

		localPath := walker.Path()
		relative, _ := filepath.Rel(local, localPath) // Won't fail

		remotePath := sftp.Join(remote, relative)
		remoteStat, err := ftp.Lstat(remotePath)

		if os.IsNotExist(err) {
			if localStat.IsDir() {
				paddedPrint("local -> remote (mkdir)", relative)

				err = ftp.Mkdir(remotePath)
				if err != nil {
					log.Errorf("  failed to create directory, skipping: %s", err)
					walker.SkipDir()
					continue
				}

				err = ftp.Chmod(remotePath, localStat.Mode().Perm())
				if err != nil {
					log.Warnf("  failed to chmod directory: %s", err)
				}
			} else {
				paddedPrint("local -> remote (new)", relative)

				localFile, err := os.Open(localPath)
				if err != nil {
					log.Errorf("  failed to open local file: %s", err)
					continue;
				}

				remoteFile, err := ftp.OpenFile(remotePath, os.O_CREATE | os.O_TRUNC | os.O_WRONLY)
				if err != nil {
					log.Errorf("  failed to open remote file: %s", err)
					goto a
				}

				_, err = io.Copy(remoteFile, localFile)

				remoteFile.Close()
				if err == nil {
					err = ftp.Chtimes(remotePath, time.Time{}, localStat.ModTime())
					if err != nil {
						log.Errorf("  failed to change mtime of remote file: %s", err)
					}

					err = ftp.Chmod(remotePath, localStat.Mode().Perm())
					if err != nil {
						log.Warnf("  failed to chmod remote file: %s", err)
					}
				} else {
					log.Errorf("  copy failed: %s", err)
				}

			a:
				localFile.Close()
				continue;
			}
		} else if err != nil {
			log.Errorf("Failed to stat remote path %s: %s", relative, err)
			continue
		}

		if localStat.IsDir() {
			continue
		}

		if !remoteStat.Mode().IsRegular() {
			log.Errorf("%s is a regular file locally, but not remotely. This must be resolved manually.", relative)
			continue
		}

		cmp := localStat.ModTime().Compare(remoteStat.ModTime())
		if cmp == -1 {
			paddedPrint("remote -> local", relative)

			remoteFile, err := ftp.Open(remotePath)
			if err != nil {
				log.Errorf("  failed to open remote file: %s", err)
				continue;
			}

			localFile, err := os.OpenFile(localPath, os.O_TRUNC | os.O_WRONLY, localStat.Mode().Perm())
			if err != nil {
				log.Errorf("  failed to open local file: %s", err)
				goto b
			}

			_, err = io.Copy(localFile, remoteFile)

			localFile.Close()
			if err == nil {
				err = os.Chtimes(localPath, time.Time{}, remoteStat.ModTime())
				if err != nil {
					log.Errorf("  failed to change mtime of local file: %s", err)
				}
			} else {
				log.Errorf("  copy failed: %s", err)
			}

		b:
			remoteFile.Close()
		} else if cmp == 1 {
			paddedPrint("local -> remote", relative)

			localFile, err := os.Open(localPath)
			if err != nil {
				log.Errorf("  failed to open local file: %s", err)
				continue
			}

			remoteFile, err := ftp.OpenFile(remotePath, os.O_TRUNC | os.O_WRONLY)
			if err != nil {
				log.Errorf("  failed to open remote file: %s", err)
				goto c
			}

			_, err = io.Copy(remoteFile, localFile)

			remoteFile.Close()
			if err == nil {
				err = ftp.Chtimes(remotePath, time.Time{}, localStat.ModTime())
				if err != nil {
					log.Errorf("  failed to change mtime of remotefile: %s", err)
				}
			} else {
				log.Errorf("  copy failed: %s", err)
			}

		c:
			localFile.Close()
		}
	}
}
