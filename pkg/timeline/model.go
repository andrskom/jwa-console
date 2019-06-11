package timeline

import (
	"errors"
	"time"

	jira "github.com/andygrunwald/go-jira"
)

var ErrTimelineEmpty = errors.New("timeline is empty")

type Model struct {
	Finished    bool
	StartTime   time.Time
	FinishTime  time.Time
	Description string
	Issue       *jira.Issue
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
	return m.FinishTime.Sub(m.StartTime)
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
