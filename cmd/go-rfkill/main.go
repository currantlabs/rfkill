// +build linux

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/currantlabs/rfkill"
)

func filter(c *cli.Context) rfkill.Filter {
	if i := c.Int("idx"); i != -1 {
		return rfkill.WithIndex(i)
	}
	if t := c.String("type"); t != "" {
		return rfkill.WithTypeName(t)
	}
	if n := c.String("name"); n != "" {
		return rfkill.WithName(n)
	}
	return rfkill.Any()
}

func cmdFactory(act func(s *rfkill.Switch)) func(*cli.Context) {
	return func(c *cli.Context) {
		r, err := rfkill.Open()
		if err != nil {
			log.Fatal(err)
		}
		ss, err := r.Switches(filter(c))
		if err != nil {
			log.Fatalf("failed: %s", err)
		}
		for _, s := range ss {
			act(s)
		}
	}
}

func main() {

	app := cli.NewApp()
	app.Name = "go-rfkill"
	app.Usage =
		` Tool for switching wireless diveces.

	 	Implemented in Go for testing github.com/currantlabs/rfkill package`
	app.Author = "Tzu-Jung Lee"
	app.Email = "roylee17@currantlabs.com"
	app.Version = "0.0.1"

	flg := []cli.Flag{
		cli.IntFlag{Name: "idx, i", Value: -1, Usage: "switch index"},
		cli.StringFlag{Name: "type, t", Value: "", Usage: "switch type"},
		cli.StringFlag{Name: "name, n", Value: "", Usage: "switch name"},
	}

	app.Commands = []cli.Command{
		{
			Name:   "list",
			Usage:  "List switches.",
			Flags:  flg,
			Action: cmdFactory(func(s *rfkill.Switch) { fmt.Println(s) }),
		},
		{
			Name:   "block",
			Usage:  "Block the specified switches",
			Flags:  flg,
			Action: cmdFactory(func(s *rfkill.Switch) { s.Block() }),
		},
		{
			Name:   "unblock",
			Usage:  "Unblock the specified switches",
			Flags:  flg,
			Action: cmdFactory(func(s *rfkill.Switch) { s.Unblock() }),
		},
		{
			Name:  "event",
			Usage: "Listen for the switch events",
			Flags: flg,
			Action: func(c *cli.Context) {
				r, err := rfkill.Open()
				if err != nil {
					log.Fatal(err)
				}
				r.Listen(func(e rfkill.Event) {
					log.Printf("event: %v", e)
				}, 100*time.Millisecond)
				select {}
			},
		},
	}

	app.Run(os.Args)
}
