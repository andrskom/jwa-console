package timeline

import (
	"errors"
	"time"

	"github.com/andygrunwald/go-jira"
)

var ErrTimelineEmpty = errors.New("timeline is empty")

type Model struct {
	Finished    bool
	StartTime   time.Time
	FinishTime  time.Time
	Description string
	Issue       *jira.Issue
	Tag         string
}

func NewModel(issue *jira.Issue) *Model {
	return &Model{
		StartTime: time.Now(),
		Issue:     issue,
	}
}

func (m *Model) IsFinished() bool {
	return m.Finished
}

func (m *Model) Finish() {
	m.Finished = true
	m.FinishTime = time.Now()
}

func (m *Model) Duration() time.Duration {
	return m.FinishTime.Sub(m.StartTime).Round(time.Second)
}

func (m *Model) ActivityDuration() time.Duration {
	return time.Now().Sub(m.StartTime).Round(time.Second)
}

type Timeline struct {
	List []*Model
}

func (t *Timeline) GetCurrent() (*Model, error) {
	if len(t.List) == 0 {
		return nil, ErrTimelineEmpty
	}
	return t.List[len(t.List)-1], nil
}

func (t *Timeline) Add(m *Model) {
	t.List = append(t.List, m)
}

type DurationDescription struct {
	Duration time.Duration
	Summary  string
}

func (t *Timeline) GetDurationsByTasks() map[string]DurationDescription {
	res := make(map[string]DurationDescription)
	for _, task := range t.List {
		if _, ok := res[task.Issue.Key]; !ok {
			res[task.Issue.Key] = DurationDescription{
				Duration: time.Duration(0),
				Summary:  task.Issue.Fields.Summary,
			}
		}
		m := res[task.Issue.Key]
		if task.IsFinished() {
			m.Duration += task.Duration()
		} else {
			m.Duration += task.ActivityDuration()
		}
		res[task.Issue.Key] = m

	}
	return res
}
