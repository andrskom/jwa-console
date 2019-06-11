package action

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/andrskom/jwa-console/pkg/timeline"
)

func Publish(
	timelineComponent *timeline.Component,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if err := timelineComponent.Publish(); err != nil {
			return err
		}
		fmt.Println(`Worklog sent`)
		return nil
	}
}
