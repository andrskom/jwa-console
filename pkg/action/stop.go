package action

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/andrskom/jwa-console/pkg/timeline"
)

func Stop(
	timelineComponent *timeline.Component,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		model, err := timelineComponent.Stop()
		if err != nil {
			return err
		}
		fmt.Printf(`Stop task %s %s
`, model.Issue.Key, model.Issue.Fields.Summary)
		return nil
	}
}
