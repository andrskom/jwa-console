package timeline

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"

	"github.com/andrskom/jwa-console/pkg/jiraf"
	"github.com/andrskom/jwa-console/pkg/storage/file"
)

const (
	IssueStatuNameInProgress = "In Progress"
)

type Component struct {
	db          *file.DB
	file        string
	jiraFactory *jiraf.Factory
}

func NewComponent(db *file.DB, jiraFactory *jiraf.Factory) *Component {
	return &Component{db: db, jiraFactory: jiraFactory, file: "timeline.json"}
}

func (c *Component) Init() error {
	data, err := json.Marshal(Timeline{List: make([]*Model, 0)})
	if err != nil {
		return err
	}
	return c.db.WriteData(c.file, data)
}

func (c *Component) GetJiraFactory() *jiraf.Factory {
	return c.jiraFactory
}

type StartOpts struct {
	UsePrevDescription bool
	Description        string
}

func (o *StartOpts) Validate() error {
	if o == nil {
		return nil
	}
	if o.UsePrevDescription && len(o.Description) > 0 {
		return errors.New("u must use either -m or -pd ")
	}

	return nil
}

func (c *Component) Start(taskID string, opts *StartOpts) (*Model, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}
	timeline, err := c.getTimeline()
	if err != nil {
		return nil, err
	}

	model, err := timeline.GetCurrent()
	if err != nil && err != ErrTimelineEmpty {
		return nil, err
	}
	if err != ErrTimelineEmpty && !model.IsFinished() {
		return nil, errors.New("last task is not finished")
	}

	client, err := c.jiraFactory.GetClient()
	if err != nil {
		return nil, err
	}

	issue, resp, err := client.Issue.Get(taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("uexpected jira response, while try to get issue: %s", resp.Status)
	}

	// if issue.Fields.Status.Name != IssueStatuNameInProgress {
	// 	return nil, fmt.Errorf(
	// 		"status of task must be '%s' for start, actual is '%s'",
	// 		IssueStatuNameInProgress,
	// 		issue.Fields.Status.Name,
	// 	)
	// }

	newModel := NewModel(issue)
	if opts != nil {
		if opts.UsePrevDescription {
			set := false
			for i := len(timeline.List) - 1; i >= 0; i-- {
				if timeline.List[i].Issue.Key == newModel.Issue.Key {
					newModel.Description = timeline.List[i].Description
					set = true
					break
				}
			}
			if !set {
				return nil, errors.New(
					"can't use description as prev the same task, because it does not found",
				)
			}
		}
		if len(opts.Description) > 0 {
			newModel.Description = opts.Description
		}
	}
	timeline.Add(newModel)

	if err := c.saveTimeline(timeline); err != nil {
		return nil, err
	}
	return newModel, nil
}

func (c *Component) Stop() (*Model, error) {
	timeline, err := c.getTimeline()
	if err != nil {
		return nil, err
	}

	model, err := timeline.GetCurrent()
	if err != nil && err != ErrTimelineEmpty {
		return nil, err
	}
	if err != ErrTimelineEmpty && model.IsFinished() {
		return nil, errors.New("last task already finished")
	}
	model.Finish()
	if err := c.saveTimeline(timeline); err != nil {
		return nil, err
	}

	return model, nil
}

func (c *Component) Get() (*Timeline, error) {
	return c.getTimeline()
}

func (c *Component) GetCurrent() (*Model, error) {
	model, err := c.getTimeline()
	if err != nil {
		return nil, err
	}
	return model.GetCurrent()
}

func (c *Component) Publish() error {
	jiraClient, err := c.jiraFactory.GetClient()
	if err != nil {
		return err
	}

	models, err := c.getTimeline()
	if err != nil {
		return err
	}
	user, resp, err := jiraClient.User.GetSelf()
	if err != nil {
		return fmt.Errorf("unexpected response code while try to get user: %d", resp.StatusCode)
	}

	now := jira.Time(time.Now())
	for _, model := range models.List {
		_, resp, err := jiraClient.Issue.AddWorklogRecord(model.Issue.Key, &jira.WorklogRecord{
			Author:           user,
			Created:          &now,
			Updated:          &now,
			Started:          (*jira.Time)(&model.StartTime),
			TimeSpentSeconds: int(model.Duration().Seconds()),
			Comment:          model.Description,
		})
		if err != nil {
			return fmt.Errorf("unexpected response code while try to send worklog: %d", resp.StatusCode)
		}
	}

	return c.saveTimeline(&Timeline{List: make([]*Model, 0)})
}

type EditOpts struct {
	Description *string
	StartTime   *time.Time
	FinishTime  *time.Time
}

func (c *Component) Edit(num int, opts EditOpts) error {
	tl, err := c.getTimeline()
	if err != nil {
		return err
	}
	if len(tl.List) <= num {
		return errors.New("bad number of record")
	}
	if opts.Description != nil {
		tl.List[num].Description = *opts.Description
	}
	if opts.StartTime != nil {
		if num > 0 && opts.StartTime.Sub(tl.List[num-1].FinishTime) < 0 {
			return errors.New("can't set start time before finish time previously record")
		}
		tl.List[num].StartTime = *opts.StartTime
	}
	if opts.FinishTime != nil {
		if !tl.List[num].IsFinished() {
			return errors.New("u can't edit finish time while task not stopped")
		}
		if num+1 != len(tl.List) && tl.List[num+1].StartTime.Sub(*opts.FinishTime) < 0 {
			return errors.New("can't set finish time after start time next record")
		}
		tl.List[num].FinishTime = *opts.FinishTime
	}
	return c.saveTimeline(tl)
}

func (c *Component) getTimeline() (*Timeline, error) {
	data, err := c.db.ReadData(c.file)
	if err != nil {
		return nil, err
	}

	var res Timeline
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *Component) saveTimeline(t *Timeline) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	return c.db.WriteData(c.file, data)
}
