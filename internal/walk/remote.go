package walk

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/diniamo/dsync/internal/log"

	"github.com/pkg/sftp"
)

func WalkRemote(ftp *sftp.Client, remote, local string) {
	for walker := ftp.Walk(remote); walker.Step(); {
		if walker.Err() != nil {
			log.Errorf("Failed to step remote walker: %s", walker.Err())
			continue
		}

		remotePath := walker.Path()
		remoteStat := walker.Stat()

		if !remoteStat.Mode().IsRegular() && !remoteStat.IsDir() {
			continue
		}

		relative, _ := filepath.Rel(remote, walker.Path()) // Won't fail

		localPath := sftp.Join(local, relative)
		_, err := os.Lstat(localPath)
		
		if os.IsNotExist(err) {
			if remoteStat.IsDir() {
				paddedPrint("remote -> local (mkdir)", relative)
				
				err = os.Mkdir(localPath, remoteStat.Mode().Perm())
				if err != nil {
					log.Errorf("  failed to create directory: %s", err)
				}
			} else {
				paddedPrint("remote -> local (new)", relative)

				remoteFile, err := ftp.Open(remotePath)
				if err != nil {
					log.Errorf("  failed to open remote file: %s", err)
					continue
				}
			
				localFile, err := os.OpenFile(localPath, os.O_CREATE | os.O_WRONLY, remoteStat.Mode().Perm())
				if err != nil {
					log.Errorf("  failed to open local file: %s", err)
					goto a
				}

				_, err = io.Copy(localFile, remoteFile)

				localFile.Close()
				if err == nil {
					err = os.Chtimes(localPath, time.Time{}, remoteStat.ModTime())
					if err != nil {
						log.Errorf("  failed to change mtime of remote file: %s", err)
					}
				} else {
					log.Errorf("  copy failed: %s", err)
				}

			a:
				remoteFile.Close()
			}
		} else if err != nil {
			log.Errorf("Failed to stat local path %s: %s", relative, err)
		}
	}
}
