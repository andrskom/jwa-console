package timeline

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/pkg/errors"

	"github.com/andrskom/jwa-console/pkg/config"
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
	cfg         *config.Component
}

func NewComponent(db *file.DB, jiraFactory *jiraf.Factory, cfg *config.Component) *Component {
	return &Component{db: db, jiraFactory: jiraFactory, file: "timeline.json", cfg: cfg}
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

func (c *Component) BuildModel(taskID string, opts *StartOpts) (*Model, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	timeline, err := c.getTimeline()
	if err != nil {
		return nil, err
	}

	client, err := c.jiraFactory.GetClient()
	if err != nil {
		return nil, err
	}

	issue, resp, err := client.Issue.Get(taskID, nil)
	if err != nil {
		return nil, fmt.Errorf("uexpected jira response, while try to get issue: %s", resp.Status)
	}

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

	return newModel, nil
}

func (c *Component) Start(newModel *Model) (*Model, error) {

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

	cfg, err := c.cfg.GetCfg()
	if err != nil {
		return nil, err
	}

	if len(cfg.StatusesForStart) != 0 {
		hasStatus := false
		for _, st := range cfg.StatusesForStart {
			if newModel.Issue.Fields.Status.Name == st {
				hasStatus = true
			}
		}
		if !hasStatus {
			return nil, fmt.Errorf(
				"status of task must be '%s' for start, actual is '%s'",
				strings.Join(cfg.StatusesForStart, ","),
				newModel.Issue.Fields.Status.Name,
			)
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

	now := jira.Time(time.Now().Round(time.Second).Add(time.Millisecond))
	lastSentIndex := 0
	for i, model := range models.List {
		if model.Duration() <= time.Minute {
			log.Printf("%d [%s] Not sent, because duration less than minute", i, model.Issue.Key)
			continue
		}
		startTime := model.StartTime.Round(time.Second).Add(time.Millisecond)
		comment := model.Description
		if len(model.Tag) > 0 {
			comment = "#" + model.Tag + " " + comment
		}
		_, resp, err := jiraClient.Issue.AddWorklogRecord(model.Issue.Key, &jira.WorklogRecord{
			Author:           user,
			UpdateAuthor:     user,
			Created:          &now,
			Updated:          &now,
			Started:          (*jira.Time)(&startTime),
			TimeSpentSeconds: int(model.Duration().Seconds()),
			IssueID:          model.Issue.ID,
			Comment:          comment,
		})
		if err != nil {
			if saveErr := c.saveTimeline(&Timeline{List: models.List[lastSentIndex+1:]}); saveErr != nil {
				log.Printf("Can't save not sent tasks to file, last sent %d", lastSentIndex)
			}
			log.Println(err.Error())
			return fmt.Errorf(
				"unexpected response code while try to send worklog #%d: %d, for issue: %s",
				i,
				resp.StatusCode,
				model.Issue.Key,
			)
		}
		lastSentIndex = i
	}

	return c.saveTimeline(&Timeline{List: make([]*Model, 0)})
}

type EditOpts struct {
	Description *string
	StartTime   *time.Time
	FinishTime  *time.Time
	Task        *string
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
	if opts.Task != nil {
		client, err := c.jiraFactory.GetClient()
		if err != nil {
			return err
		}
		issue, resp, err := client.Issue.Get(*opts.Task, nil)
		if err != nil {
			return fmt.Errorf(
				"unexpected response code while try to get jira issue: %s, respo code: %d",
				*opts.Task,
				resp.StatusCode,
			)
		}

		tl.List[num].Issue = issue
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
