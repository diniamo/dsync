package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"github.com/diniamo/dsync/internal/virtfs"
	log "github.com/diniamo/glog"
	"github.com/urfave/cli/v3"
)

type Empty struct {}

type PaddedPrinter struct {
	max int
	extra int
}

func (p *PaddedPrinter) PrintTransaction(nameFrom, nameTo string, action, path string) {
	left := fmt.Sprintf("%s -> %s (%s)", nameFrom, nameTo, action)
	n := len(left)
	if n > p.max {
		p.max = n
	}

	padding := strings.Repeat(" ", (p.max - n) + p.extra)

	fmt.Println(left + padding + path)
}

func errorf(format string, a ...any) error {
	return errors.New(fmt.Sprintf(format, a...))
}

func errorIndent(err error) error {
	return errors.New("  " + err.Error())
}

func run(ctx context.Context, cmd *cli.Command) error {
	var err error

	specs := cmd.StringArgs("spec")
	port := cmd.String("port")

	vFSs := make([]virtfs.VirtualFS, len(specs))
	for i, spec := range specs {
		vFSs[i], err = virtfs.New(spec, port)
		if err != nil {
			return errorf("Failed to parse or connect to \"%s\": %s", spec, err)
		}

		defer vFSs[i].Close()
	}

	printer := PaddedPrinter{extra: 3}

	pathsDone := map[string]Empty{}
	// TODO: more descriptive errors
	for i, vFSOuter := range vFSs {
		for walker := vFSOuter.Walk(); walker.Step(); {
			if walker.Err() != nil {
				return errorf("%s: Failed to step walker: %s", specs[i], walker.Err())
			}

			relative := vFSOuter.Rel(walker.Path())
			
			if _, ok := pathsDone[relative]; ok {
				continue
			}
			pathsDone[relative] = Empty{}

			statOuter := walker.Stat()
			isDirOuter := statOuter.IsDir()

			if !isDirOuter && !statOuter.Mode().IsRegular() {
				continue
			}

			for j, vFSInner := range vFSs {
				if j == i {
					continue
				}

				statInner, err := vFSInner.Lstat(relative)

				if err != nil {
					if os.IsNotExist(err) {
						if isDirOuter {
							printer.PrintTransaction(specs[i], specs[j], "mkdir", relative)

							err = vFSInner.Mkdir(relative, statOuter.Mode().Perm())
							if err != nil {
								return errorIndent(err)
							}

							err = vFSInner.Chmod(relative, statOuter.Mode().Perm())
							if err != nil {
								return errorIndent(err)
							}
						} else {
							printer.PrintTransaction(specs[i], specs[j], "copy", relative)

							// TODO: lazy init
							fileOuter, err := vFSOuter.Open(relative)
							if err != nil {
								return errorIndent(err)
							}

							fileInner, err := vFSInner.Create(relative)
							if err != nil {
								return errorIndent(err)
							}

							_, err = io.Copy(fileInner, fileOuter)
							if err != nil {
								return errorIndent(err)
							}

							fileOuter.Close()
							fileInner.Close()

							err = vFSInner.Chmtime(relative, statOuter.ModTime())
							if err != nil {
								return errorIndent(err)
							}

							err = vFSInner.Chmod(relative, statOuter.Mode().Perm())
							if err != nil {
								return errorIndent(err)
							}
						}

						continue
					} else {
						return errorf("%s: failed to stat %s", specs[j], relative)
					}
				}

				if isDirOuter {
					continue
				}

				if !statInner.Mode().IsRegular() {
					return errorf("%s is a regular file in %s, but not in %s. This must be resolved manually.", relative, specs[i], specs[j])
				}

				// SFTP doesn't seem to retain mtime precision below seconds,
				// so truncating for comparison is a must.
				mtimeOuter := statOuter.ModTime().Truncate(time.Second)
				mtimeInner := statInner.ModTime().Truncate(time.Second)

				switch mtimeOuter.Compare(mtimeInner) {
				case -1:
					printer.PrintTransaction(specs[j], specs[i], "overwrite", relative)

					fileInner, err := vFSInner.Open(relative)
					if err != nil {
						return errorIndent(err)
					}

					fileOuter, err := vFSOuter.Create(relative)
					if err != nil {
						return errorIndent(err)
					}

					_, err = io.Copy(fileOuter, fileInner)
					if err != nil {
						return errorIndent(err)
					}

					fileInner.Close()
					fileOuter.Close()

					err = vFSOuter.Chmtime(relative, mtimeInner)
					if err != nil {
						return errorIndent(err)
					}
				case 1:
					printer.PrintTransaction(specs[i], specs[j], "overwrite", relative)

					fileOuter, err := vFSOuter.Open(relative)
					if err != nil {
						return errorIndent(err)
					}

					fileInner, err := vFSInner.Create(relative)
					if err != nil {
						return errorIndent(err)
					}

					_, err = io.Copy(fileInner, fileOuter)
					if err != nil {
						return errorIndent(err)
					}

					fileOuter.Close()
					fileInner.Close()

					err = vFSInner.Chmtime(relative, mtimeOuter)
					if err != nil {
						return errorIndent(err)
					}
				}
			}
		}
	}

	return nil
}

func main() {
	cmd := cli.Command{
		Name: "dsync",
		Usage: "P2P file synchronization tool using the SSH protocol",
		UsageText: `dsync [global options] <path spec... (at least 2)>

Local paths must be specified as is.
Remote paths must be specified in the following format: [user@]host:path
The host must not contain a port number. It may be specified using the --port/-p flag instead.
If a user is not specified, the current user will be used.`,
		Arguments: []cli.Argument{
			&cli.StringArgs{
				Name: "spec",
				Min: 2,
				Max: math.MaxInt,
			},
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "port",
				Aliases: []string{"p"},
				Value: "22",
			},
		},
		Action: run,
	}

	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
