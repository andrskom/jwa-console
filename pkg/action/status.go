package action

import (
	"fmt"
	"time"

	"github.com/urfave/cli"

	"github.com/andrskom/jwa-console/pkg/timeline"
)

func Status(
	timelineComponent *timeline.Component,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		model, err := timelineComponent.GetCurrent()
		if err != nil {
			if err == timeline.ErrTimelineEmpty {
				warnColor.Println("Timeline is empty")
			}
			return err
		}
		if model.IsFinished() {
			doNothinDuration := time.Now().Sub(model.FinishTime)
			fmt.Printf(`Last task: %s %s
Do nothing: %s
`, model.Issue.Key, model.Issue.Fields.Summary, doNothinDuration.String())
			return nil
		}

		taskDuration := time.Now().Sub(model.StartTime)
		fmt.Printf(`Current task: %s %s
Activity: %s
`, model.Issue.Key, model.Issue.Fields.Summary, taskDuration.String())

		return nil
	}
}
