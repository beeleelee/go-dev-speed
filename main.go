package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	local := []*cli.Command{
		randReadCmd,
	}

	app := &cli.App{
		Name:     "dev-speed",
		Flags:    []cli.Flag{},
		Commands: local,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

var randReadCmd = &cli.Command{
	Name:  "rand-read",
	Usage: "",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "batch",
			Usage: "",
			Value: 10,
		},
		&cli.IntFlag{
			Name:  "buf",
			Usage: "",
			Value: 131072,
		},
		&cli.IntFlag{
			Name:  "round",
			Usage: "",
			Value: 30,
		},
	},
	Action: func(c *cli.Context) error {
		args := c.Args().Slice()
		if len(args) < 1 {
			return fmt.Errorf("[usage] xlfs random-read-os [filename]")
		}

		sf, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer sf.Close()
		sfinfo, err := sf.Stat()
		if err != nil {
			return err
		}
		fileSize := sfinfo.Size()
		fmt.Println("source file size ", fileSize)

		bufSize := c.Int("buf")
		if fileSize < int64(bufSize)*2 {
			return fmt.Errorf("file size too short")
		}

		round := c.Int("round")
		if round < 1 {
			round = 1
		}
		batchChan := make(chan struct{}, c.Int("batch"))
		var wg sync.WaitGroup
		for i := 0; i < round; i++ {
			wg.Add(1)
			go func(i int) {
				defer func() {
					<-batchChan
					wg.Done()
				}()
				batchChan <- struct{}{}
				buf := make([]byte, bufSize)

				off := rand.Intn(int(fileSize) - bufSize)
				now := time.Now()
				_, err = sf.ReadAt(buf, int64(off))
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Printf("read %d data, time elapsed: %d us\n", len(buf), time.Since(now).Microseconds())
			}(i)
		}
		wg.Wait()
		return nil
	},
}
