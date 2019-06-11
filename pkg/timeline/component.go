package timeline

import (
	"encoding/json"
	"fmt"
	"time"

	jira "github.com/andygrunwald/go-jira"
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

func (c *Component) Start(taskID string) (*Model, error) {
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
		})
		if err != nil {
			return fmt.Errorf("unexpected response code while try to send worklog: %d", resp.StatusCode)
		}
	}

	return c.saveTimeline(&Timeline{List: make([]*Model, 0)})
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
