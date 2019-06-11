package action

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli"

	"github.com/andrskom/jwa-console/pkg/timeline"
)

var doNothingColor = color.New(color.FgBlack, 1)
var activityColor = color.New(color.FgGreen)
var warnColor = color.New(color.FgYellow)

func Show(
	timelineComponent *timeline.Component,
) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		model, err := timelineComponent.Get()
		if err != nil {
			return err
		}
		if len(model.List) == 0 {
			warnColor.Println(`Nothing`)
		}
		var prevTask *timeline.Model
		allDuration := time.Duration(0)
		for i, task := range model.List {
			if task.IsFinished() {
				allDuration += task.Duration()
			} else {
				allDuration += time.Now().Sub(task.StartTime)
			}
			if prevTask != nil {
				fmt.Print(drawRest(task, prevTask))
			}
			fmt.Printf("%2d %s", i, drawModel(task))
			prevTask = task

		}
		activityColor.Printf(`
Sum of activity: %s
`, allDuration.String())

		return nil
	}
}

func drawModel(model *timeline.Model) string {
	res := activityColor.Sprintf(`%s Start %s %s
`, model.StartTime.Format(time.RFC822), model.Issue.Key, model.Issue.Fields.Summary)
	res += activityColor.Sprintln("   |")

	var interval time.Duration
	if model.IsFinished() {
		interval = model.Duration() / (time.Hour / 2)
	} else {
		interval = time.Now().Sub(model.StartTime) / (time.Hour / 2)
	}
	for interval > 0 {
		interval--
		res += activityColor.Sprintln("   |")
	}

	if model.IsFinished() {
		res += activityColor.Sprintf(`   %s Duration: %s
`, model.FinishTime.Format(time.RFC822), model.Duration().String())
		return res
	}
	res += activityColor.Sprintf(`   Activity %s
`, time.Now().Sub(model.StartTime).String())
	return res
}

func drawRest(model *timeline.Model, prevModel *timeline.Model) string {
	res := doNothingColor.Sprintf(`%s Do nothing
`, prevModel.FinishTime.Format(time.RFC822))
	dur := model.StartTime.Sub(prevModel.FinishTime)
	interval := dur / (time.Hour / 2)
	for interval > 0 {
		interval--
		res += doNothingColor.Sprintln("   |")
	}
	res += doNothingColor.Sprintf(`%s Duration: %s
`, model.StartTime.Format(time.RFC822), dur.String())
	return res
}
