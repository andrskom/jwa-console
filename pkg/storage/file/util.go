package file

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type DB struct {
	dir      string
	initFile string
}

func New(dir string, initFile string) *DB {
	return &DB{dir: dir, initFile: initFile}
}

func (db *DB) Init() error {
	_, err := os.Stat(db.dir)
	if err == nil {
		return errors.New("app already init")
	}
	if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(db.dir, 0755); err != nil {
		return err
	}

	if _, err := os.Create(filepath.Join(db.dir, db.initFile)); err != nil {
		return err
	}
	return nil
}

func (db *DB) validateInit() error {
	if _, err := os.Stat(db.dir); err != nil {
		return errors.Wrap(err, "init validation error")
	}
	return nil
}

func (db *DB) ReadData(file string) ([]byte, error) {
	if err := db.validateInit(); err != nil {
		return nil, errors.Wrap(err, "read data err")
	}
	return ioutil.ReadFile(filepath.Join(db.dir, file))
}

func (db *DB) WriteData(file string, data []byte) error {
	if err := db.validateInit(); err != nil {
		return errors.Wrap(err, "write data err")
	}
	return ioutil.WriteFile(filepath.Join(db.dir, file), data, 0644)
}
