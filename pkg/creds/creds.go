package creds

import (
	"encoding/json"

	"github.com/andrskom/jwa-console/pkg/storage/file"
)

type Component struct {
	db   *file.DB
	file string
}

func New(db *file.DB) *Component {
	return &Component{db: db, file: "auth.json"}
}

func (s *Component) Save(m *Model) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return s.db.WriteData(s.file, data)
}

func (s *Component) Get() (*Model, error) {
	data, err := s.db.ReadData(s.file)
	if err != nil {
		return nil, err
	}
	var res Model
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
