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
var errColor = color.New(color.FgRed)
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

		fmt.Printf(
			"\n%s %s\n",
			activityColor.Sprint("Sum of activity:"),
			getDuration(allDuration, activityColor),
		)

		return nil
	}
}

func drawModel(model *timeline.Model) string {
	res := activityColor.Sprintf(`%s Start %s %s
`, model.StartTime.Format(time.RFC822), model.Issue.Key, model.Issue.Fields.Summary)
	res += activityColor.Sprintf(`   + %s
`, model.Description)

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
		res += activityColor.Sprintf(`   %s Duration: `, model.FinishTime.Format(time.RFC822))
		res += getDuration(model.Duration(), activityColor) + "\n"
		return res
	}
	res += activityColor.Sprint(`   Activity `)
	res += getDuration(time.Now().Sub(model.StartTime), activityColor) + "\n"
	return res
}

func drawRest(model *timeline.Model, prevModel *timeline.Model) string {
	res := doNothingColor.Sprintf(`   %s Do nothing
`, prevModel.FinishTime.Format(time.RFC822))
	dur := model.StartTime.Sub(prevModel.FinishTime)
	interval := dur / (time.Hour / 2)
	for interval > 0 {
		interval--
		res += doNothingColor.Sprintln("   |")
	}
	res += doNothingColor.Sprintf(`   %s Duration: `, model.StartTime.Format(time.RFC822))
	res += getDuration(dur, doNothingColor) + "\n"
	return res
}

func getDuration(duration time.Duration, defaultColor *color.Color) string {
	durationString := defaultColor.Sprint(duration.String())
	if duration < 0 {
		durationString = errColor.Sprint(duration.String())
	}
	return durationString
}
