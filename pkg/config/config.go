package config

import (
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/andrskom/jwa-console/pkg/storage/file"
)

type Model struct {
	Tags               []string `json:"tags"`
	StatusesForStart   []string `json:"statusesForStart"`
	AutoChangeStatusTo string   `json:"autoChangeStatusTo"`
}

func (m *Model) Set(key string, val string) error {
	switch key {
	case "tags":
		m.Tags = strings.Split(val, ",")
	case "statusesForStart":
		m.StatusesForStart = strings.Split(val, ",")
	case "autoChangeStatusTo":
		m.AutoChangeStatusTo = val
	default:
		return errors.New("unexpected key of config field")
	}
	return nil
}

func (m *Model) AsMap() map[string]string {
	return map[string]string{
		"tags":               strings.Join(m.Tags, ","),
		"statusesForStart":   strings.Join(m.StatusesForStart, ","),
		"autoChangeStatusTo": m.AutoChangeStatusTo,
	}
}

type Component struct {
	db   *file.DB
	file string
}

func NewComponent(db *file.DB) *Component {
	return &Component{db: db, file: "config.json"}
}

func (c *Component) Init() error {
	_, err := c.db.ReadData(c.file)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err == nil {
		return nil
	}

	bytes, err := json.Marshal(Model{Tags: make([]string, 0), StatusesForStart: make([]string, 0)})
	if err != nil {
		return err
	}

	return c.db.WriteData(c.file, bytes)
}

func (c *Component) GetCfg() (*Model, error) {
	bytes, err := c.db.ReadData(c.file)
	if err != nil {
		return nil, err
	}

	var cfg Model
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (c *Component) Save(m *Model) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return c.db.WriteData(c.file, bytes)
}
