// This code comes from https://github.com/golang/exp/blob/master/shiny/driver/x11driver/shm_linux_amd64.go
package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/ghetzel/cli"
	"github.com/ghetzel/shmtool/shm"
	"io"
	"os"
	"strconv"
)

const DefaultLogLevel = `info`

func main() {
	app := cli.NewApp()
	app.Name = `shmtool`
	app.Usage = `a command line utility for interacting with SysV-style shared memory segments`
	app.Version = shm.Version
	app.EnableBashCompletion = false
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of logging verbosity`,
			Value:  DefaultLogLevel,
			EnvVar: `LOGLEVEL`,
		},
	}

	app.Before = func(c *cli.Context) error {
		// set log verbosity
		if lvl := c.String(`log-level`); lvl != `` {
			if l, err := log.ParseLevel(lvl); err == nil {
				log.SetLevel(l)
			} else {
				log.Fatalf("Invalid log level '%s'", lvl)
				return fmt.Errorf("%v", err)
			}
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:      `open`,
			Usage:     `Create or open a shared memory buffer and write the contents of standard input to it`,
			ArgsUsage: `[ID]`,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  `offset, o`,
					Usage: `The number of bytes to skip before beginning the write operation`,
				},
				cli.IntFlag{
					Name:  `size, s`,
					Usage: `The size (in bytes) of the shared memory segment (if creating)`,
				},
			},
			Action: func(c *cli.Context) {
				var size int

				if c.NArg() == 0 {
					size = c.Int(`size`)

					if size == 0 {
						log.Fatalf("Must specify a segment size")
					}
				}

				var segment *shm.Segment
				var err error

				if size > 0 {
					segment, err = shm.Create(size)
				} else {
					if segmentId, err := strconv.ParseUint(c.Args().First(), 10, 64); err == nil {
						segment, err = shm.Open(int(segmentId))
					} else {
						log.Fatalf("Failed to parse segment ID: %v", err)
						return
					}
				}

				if err == nil {
					if offset := c.Int(`offset`); offset > 0 {
						segment.Offset = offset
					}

					log.Debugf("Opened shared memory segment %d: size is %d, offset is %d", segment.Id, segment.Size, segment.Offset)
					fmt.Printf("%d\n", segment.Id)

					if n, err := io.Copy(segment, os.Stdin); err == nil || err == io.EOF {
						log.Infof("Wrote %d bytes to shared memory", n)
					} else {
						log.Errorf("Failed to copy input: %v", err)
					}
				} else {
					log.Fatalf("Failed to open shared memory segment: %v", err)
				}
			},
		}, {
			Name:      `read`,
			Usage:     `Read the contents of a shared memory buffer to standard output`,
			ArgsUsage: `ID`,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  `offset, o`,
					Usage: `The number of bytes to skip before beginning the read operation`,
				},
				cli.IntFlag{
					Name:  `size, s`,
					Usage: `The number of bytes to read from the shared memory segment`,
				},
			},
			Action: func(c *cli.Context) {
				if id, err := strconv.ParseUint(c.Args().First(), 10, 64); err == nil {
					segmentId := int(id)

					if segment, err := shm.Open(segmentId); err == nil {
						readSize := c.Int(`size`)

						if readSize > segment.Size || readSize == 0 {
							readSize = segment.Size
						}

						if offset := c.Int(`offset`); offset > 0 {
							segment.Offset = offset
						}

						log.Debugf("Opened shared memory segment %d: size is %d, offset is %d", segment.Id, segment.Size, segment.Offset)
						log.Debugf("Reading %d bytes...", readSize)

						if n, err := io.CopyN(os.Stdout, segment, int64(readSize)); err == nil {
							log.Infof("Read %d bytes from shared memory", n)
						} else {
							log.Fatalf("Failed to read from shared memory segment: %v", err)
						}
					} else {
						log.Fatalf("Failed to open shared memory segment %d: %v", segmentId, err)
					}
				} else {
					log.Fatalf("Must specify a valid segment ID: %v", err)
				}
			},
		}, {
			Name:      `rm`,
			Usage:     `Remove a shared memory segment`,
			ArgsUsage: `ID`,
			Action: func(c *cli.Context) {
				if id, err := strconv.ParseUint(c.Args().First(), 10, 64); err == nil {
					if err := shm.DestroySegment(int(id)); err == nil {
						log.Infof("Destroyed segment %d", id)
					} else {
						log.Fatalf("Failed to destroy segment %d: %v", id, err)
					}
				} else {
					log.Fatalf("Must specify a segment ID: %v", err)
				}
			},
		},
	}

	app.Run(os.Args)
}
