package action

import (
	"errors"
	"strconv"
	"time"

	"github.com/urfave/cli"

	"github.com/andrskom/jwa-console/pkg/timeline"
)

func Edit(
	timelineComponent *timeline.Component,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if len(c.Args().Get(0)) == 0 {
			return errors.New(
				"u should set number of record as last args(run jwac show for see numbers of records)",
			)
		}
		num, err := strconv.Atoi(c.Args().Get(0))
		if err != nil {
			return err
		}
		if len(c.String("m")) > 0 && c.Bool("mremove") {
			return errors.New("can't edit description and remove description in one cmd")
		}
		opts := timeline.EditOpts{}
		if len(c.String("m")) > 0 {
			opts.Description = new(string)
			*opts.Description = c.String("m")
		}
		if c.Bool("mremove") {
			opts.Description = new(string)
			*opts.Description = ""
		}
		if len(c.String("start-time")) > 0 {
			st, err := time.ParseInLocation("2006-01-02T15:04", c.String("start-time"), time.Local)
			if err != nil {
				return err
			}
			opts.StartTime = &st
		}
		if len(c.String("finish-time")) > 0 {
			ft, err := time.ParseInLocation("2006-01-02T15:04", c.String("finish-time"), time.Local)
			if err != nil {
				return err
			}
			opts.FinishTime = &ft
		}
		return timelineComponent.Edit(num, opts)
	}
}
