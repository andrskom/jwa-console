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
	"github.com/andrskom/jwa-console/pkg/config"
	"github.com/andrskom/jwa-console/pkg/creds"
	"github.com/andrskom/jwa-console/pkg/jiraf"
	"github.com/andrskom/jwa-console/pkg/storage/file"
	"github.com/andrskom/jwa-console/pkg/tag"
	"github.com/andrskom/jwa-console/pkg/timeline"
)

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Version = "v0.1.0"
	app.Name = "jwac"
	app.Usage = "Jira worklog assistant console"
	dbFilePath, err := getDotRc()
	if err != nil {
		log.Fatalf("Can't get db path: %s", err.Error())
	}
	db := file.New(dbFilePath, "init")
	credsComponent := creds.New(db)
	jiraFactory := jiraf.NewFactory(credsComponent)
	cfg := config.NewComponent(db)
	if err := cfg.Init(); err != nil {
		log.Fatalln(err)
	}
	tagComponent := tag.NewComponent(cfg)

	timelineComponent := timeline.NewComponent(db, jiraFactory, cfg)

	startFlags := []cli.Flag{
		cli.StringFlag{
			Name:  "m",
			Usage: "One line description",
		},
		cli.BoolFlag{
			Name:  "pd",
			Usage: "Use prev descr for this task",
		},
		cli.StringFlag{
			Name:  "t",
			Usage: "Tag for description, use -nt if you don't want use tag now",
		},
		cli.BoolFlag{
			Name:  "nt",
			Usage: "No tags for description",
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
			Action: action.Start(timelineComponent, tagComponent),
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

				if err := action.Start(timelineComponent, tagComponent)(c); err != nil {
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
			Name:    "show",
			Aliases: []string{"log", "ps"},
			Usage:   "Show logged",
			Action:  action.Show(timelineComponent),
		},
		{
			Name:   "status",
			Usage:  "Status of current task",
			Action: action.Status(timelineComponent),
		},
		{
			Name:    "publish",
			Aliases: []string{"push"},
			Usage:   "Status of current task",
			Action:  action.Publish(timelineComponent),
		},
		{
			Name:   "completion",
			Usage:  "Completion for terminal",
			Action: action.Completion(),
		},
		{
			Name:  "edit",
			Usage: "Edit params of work record",
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
				cli.StringFlag{
					Name:  "task",
					Usage: "Task key",
				},
			},
			Action: action.Edit(timelineComponent),
		},
		{
			Name:  "change",
			Usage: "Change to next task, equal to stop and start",
			Flags: startFlags,
			Action: func(c *cli.Context) error {
				if err := action.Stop(timelineComponent)(c); err != nil {
					return err
				}
				if err := action.Start(timelineComponent, tagComponent)(c); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "config",
			Usage: "Configuration",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "l",
					Usage: "List of configs",
				},
				cli.StringFlag{
					Name: "set",
					Usage: `Set config value.
Use ':' as separator for key and value.
Use ',' as separator for slice of strings. `,
				},
			},
			Action: action.Config(cfg),
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
