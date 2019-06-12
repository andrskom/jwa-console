package main

import (
	"log"
	"os/user"
	"path/filepath"

	"github.com/getlantern/systray"
	"github.com/rjeczalik/notify"

	"github.com/andrskom/jwa-console/pkg/creds"
	"github.com/andrskom/jwa-console/pkg/jiraf"
	"github.com/andrskom/jwa-console/pkg/storage/file"
	"github.com/andrskom/jwa-console/pkg/timeline"
	"github.com/andrskom/jwa-console/pkg/tray"
)

func main() {
	dbFilePath, err := getDotRc()
	if err != nil {
		log.Fatalf("Can't get db path: %s", err.Error())
	}

	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch(filepath.Join(dbFilePath, "timeline.json"), c, notify.Write); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)


	db := file.New(dbFilePath, "init")
	credsComponent := creds.New(db)
	jiraFactory := jiraf.NewFactory(credsComponent)

	timelineComponent := timeline.NewComponent(db, jiraFactory)

	greyAsset, err := tray.Asset("assets/grey.png")
	if err != nil {
		log.Fatalf("Can't read grey asset: %s", err.Error())
	}
	yellow, err := tray.Asset("assets/yellow.png")
	if err != nil {
		log.Fatalf("Can't read green asset: %s", err.Error())
	}

	systray.Run(func() {
		cur, err := timelineComponent.GetCurrent()
		if err != nil {
			if err == timeline.ErrTimelineEmpty {
				systray.SetIcon(greyAsset)
			} else {
				log.Fatalf("can't read db: %s", err.Error())
			}
		} else {
			if cur.IsFinished() {
				systray.SetIcon(greyAsset)
			} else {
				systray.SetIcon(yellow)
			}
		}

		for {
			<-c
			cur, err := timelineComponent.GetCurrent()
			if err != nil {
				if err == timeline.ErrTimelineEmpty {
					systray.SetIcon(greyAsset)
					continue
				}
				log.Fatalf("can't read db: %s", err.Error())
			}
			if cur.IsFinished() {
				systray.SetIcon(greyAsset)
			} else {
				systray.SetIcon(yellow)
			}
		}
	}, func() {

	})
}


func getDotRc() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, ".jwarc"), nil
}
