package action

import (
	"errors"
	"fmt"

	"github.com/urfave/cli"

	"github.com/andrskom/jwa-console/pkg/timeline"
)

func Start(
	timelineComponent *timeline.Component,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		taskID := c.Args().Get(0)
		if taskID == "" {
			return errors.New("u must set number of task as last args")
		}

		var opts *timeline.StartOpts
		if len(c.String("m")) > 0 {
			if opts == nil {
				opts = new(timeline.StartOpts)
			}
			opts.Description = c.String("m")
		}
		if c.Bool("pd") {
			if opts == nil {
				opts = new(timeline.StartOpts)
			}
			opts.UsePrevDescription = true
		}

		model, err := timelineComponent.Start(taskID, opts)
		if err != nil {
			return err
		}
		fmt.Printf(`Start task %s %s
`, model.Issue.Key, model.Issue.Fields.Summary)
		return nil
	}
}
