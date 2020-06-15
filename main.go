package main

import (
	gc "github.com/rthornton128/goncurses"
	"github.com/urfave/cli"

	"io/ioutil"
	"log"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "glt"
	app.Usage = "Git Local Transform"
	app.Version = "1.0.0"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "d, debug",
			Usage: "Write debug log (glt.log)",
		},
	}
	app.Action = func(c *cli.Context) {
		repo, err := OpenCurrentRepository()
		if err != nil {
			log.Fatalf("error opening repository: %v", err)
		}

		dirty := repo.IsDirty()
		if dirty == true {
			log.Fatal("git directory has uncommited changes, please stash and try agian.")
		}

		commits, err := repo.GetLog(10)
		if err != nil {
			log.Fatalf("error getting commit log: %v", err)
		}

		if c.IsSet("debug") {
			// Initialize file logging just before curses
			f, err := os.OpenFile("glt.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				log.Fatalf("error opening log file: %v", err)
			}
			defer f.Close()
			log.SetOutput(f)
		} else {
			log.SetFlags(0)
			log.SetOutput(ioutil.Discard)
		}

		stdscr, err := gc.Init()
		if err != nil {
			log.Fatal("goncurses init:", err)
		}
		defer gc.End()
		gc.Raw(true)
		gc.CBreak(true)
		gc.Echo(false)
		gc.StartColor()
		gc.Cursor(1)

		gc.InitPair(1, gc.C_WHITE, gc.C_BLUE)
		gc.InitPair(2, gc.C_YELLOW, gc.C_BLUE)
		gc.InitPair(3, gc.C_RED, gc.C_BLACK)

		commit := selectCommit(stdscr, commits)
		if commit == nil {
			return
		}

		log.Println("Entering Edit")
		logCommit(commit)

		commit = editCommit(stdscr, commit)
		if commit != nil {
			log.Println("After Edit")
			logCommit(commit)

			refChange, err := repo.SaveCommitIfModified(commit)
			if err != nil {
				log.Fatalf("Error saving commit: %s", err)
			}
			if refChange != "" {
				log.Printf("Successfully saved: %s", refChange)
			}
			showResult(stdscr, refChange)
		}
	}
	app.Run(os.Args)
}
