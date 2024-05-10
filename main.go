package main

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
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
			Value: 50,
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
		if fileSize == 0 {
			fileSize, err = devSize(sf)
			if err != nil {
				return err
			}
		}
		fmt.Println("source file size ", fileSize)

		bufSize := c.Int("buf")
		if fileSize < int64(bufSize)*2 {
			return fmt.Errorf("file size too short")
		}

		round := c.Int("round")
		if round < 1 {
			round = 1
		}
		records := make([]int64, round)
		batchChan := make(chan struct{}, c.Int("batch"))
		var wg sync.WaitGroup
		for i := range records {
			wg.Add(1)
			go func(i int, records []int64) {
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
				te := time.Since(now).Microseconds()
				records[i] = te
				fmt.Printf("read %d data, time elapsed: %s\n", len(buf), cv(te))
			}(i, records)
		}
		wg.Wait()
		fmt.Printf("round: %d\n", round)
		fmt.Printf("records for elapsed time: %v\n", records)
		sort.Slice(records, func(i, j int) bool {
			return records[i] < records[j]
		})
		if len(records) > 5 {
			fmt.Printf("Fastest 5: %v\n", records[:5])
			fmt.Printf("Slowest 5: %v\n", records[len(records)-5:])
		}
		var total int64
		for _, item := range records {
			total += item
		}
		fmt.Printf("Average speed: %s\n", cv(total/int64(round)))
		return nil
	},
}

func cv(d int64) string {
	if d < 1000 {
		return fmt.Sprintf("%d us", d)
	}
	d /= 1000
	if d < 1000 {
		return fmt.Sprintf("%d ms", d)
	}

	return fmt.Sprintf("%d s", d/1000)
}

func devSize(f *os.File) (int64, error) {
	n, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}
	return n, nil
}
