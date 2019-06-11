package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"sort"
	"syscall"

	"github.com/urfave/cli"

	"github.com/andrskom/jwa-console/pkg/action"
	"github.com/andrskom/jwa-console/pkg/action/login"
	"github.com/andrskom/jwa-console/pkg/creds"
	"github.com/andrskom/jwa-console/pkg/jiraf"
	"github.com/andrskom/jwa-console/pkg/storage/file"
	"github.com/andrskom/jwa-console/pkg/timeline"
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = "v0.0.1-alpha"
	app.Name = "jwac"
	app.Usage = "Jira worklog assistant console"
	dbFilePath, err := getDotRc()
	if err != nil {
		log.Fatalf("Can't get db path: %s", err.Error())
	}
	db := file.New(dbFilePath, "init")
	credsComponent := creds.New(db)
	jiraFactory := jiraf.NewFactory(credsComponent)

	timelineComponent := timeline.NewComponent(db, jiraFactory)

	startFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "m",
			Usage: "One line description",
		},
		cli.BoolFlag{
			Name:  "pd",
			Usage: "Use prev descr for this task",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Init application",
			Action: func(c *cli.Context) error {
				if err := db.Init(); err != nil {
					return err
				}
				if err := timelineComponent.Init(); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:   "login",
			Usage:  "Login to jira",
			Action: login.Login(credsComponent),
		},
		{
			Name:   "start",
			Usage:  "Start track task",
			Flags:  startFlags,
			Action: action.Start(timelineComponent),
		},
		{
			Name:   "stop",
			Usage:  "Stop track task",
			Action: action.Stop(timelineComponent),
		},
		{
			Name:  "start-and-wait",
			Flags: startFlags,
			Usage: "Start task and stop tracking when u send SIGTERM",
			Action: func(c *cli.Context) (err error) {
				started := false
				signalCh := make(chan os.Signal)
				signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

				if err := action.Start(timelineComponent)(c); err != nil {
					return err
				}
				started = true
				fmt.Print("Use SIGTERM for stop\n")
				<-signalCh
				if !started {
					return nil
				}
				return action.Stop(timelineComponent)(c)
			},
		},
		{
			Name:   "show",
			Usage:  "Show logged",
			Action: action.Show(timelineComponent),
		},
		{
			Name:   "status",
			Usage:  "Status of current task",
			Action: action.Status(timelineComponent),
		},
		{
			Name:   "publish",
			Usage:  "Status of current task",
			Action: action.Publish(timelineComponent),
		},
		{
			Name:   "completion",
			Usage:  "Completion for terminal",
			Action: action.Completion(),
		},
		{
			Name:   "edit",
			Usage:  "Edit params of work record",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "mremove",
					Usage: "Remove description",
				},
				cli.StringFlag{
					Name:  "m",
					Usage: "One line description",
				},
				cli.StringFlag{
					Name:  "start-time",
					Usage: "Start time in format '2006-01-02T15:04'",
				},
				cli.StringFlag{
					Name:  "finish-time",
					Usage: "Finish time in format '2006-01-02T15:04'",
				},
			},
			Action: action.Edit(timelineComponent),
		},
		// {
		// 	Name:   "test",
		// 	Flags: []cli.Flag{
		// 		cli.StringFlag{
		// 			Name:  "m",
		// 			Usage: "One line description",
		// 		},
		// 		cli.BoolFlag{
		// 			Name:  "pd",
		// 			Usage: "Use prev descr for this task",
		// 		},
		// 	},
		// 	Action: action.Test(),
		// },
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func getDotRc() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".jwarc"), nil
}
