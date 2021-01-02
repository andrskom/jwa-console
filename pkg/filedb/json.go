package filedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sync"
)

const (
	defaultDBDir     = ".jwac"
	defaultExtension = ".json"
)

// ErrUnexpectedRuneInTableName is err for checking validating status outside.
var ErrUnexpectedRuneInTableName = errors.New("unexpected rune in table name")

// TableNameValidator is a simple validator by runes whitelist.
type TableNameValidator struct {
	expectedChars map[rune]struct{}
}

// NewTableNameValidator init validators with list of runes.
func NewTableNameValidator(runes ...rune) *TableNameValidator {
	runeMap := make(map[rune]struct{})
	for _, r := range runes {
		runeMap[r] = struct{}{}
	}

	return &TableNameValidator{
		expectedChars: runeMap,
	}
}

// Validate table name, returns a wrapped error with a bad rune info.
func (p TableNameValidator) Validate(tableName string) error {
	for _, rune := range tableName {
		if _, ok := p.expectedChars[rune]; !ok {
			return fmt.Errorf("%w, rune: %v", ErrUnexpectedRuneInTableName, rune)
		}
	}

	return nil
}

type tableNameValidator interface {
	Validate(tableName string) error
}

// JSON is a simple file db.
// Store data in json format.
type JSON struct {
	mu                 sync.Mutex
	serializer         func(obj interface{}) ([]byte, error)
	deserializer       func(data []byte, obj interface{}) error
	dirPath            string
	tableNameValidator tableNameValidator
}

// NewJSON is the func for configuring db.
func NewJSON(
	serializer func(obj interface{}) ([]byte, error),
	deserializer func(data []byte, obj interface{}) error,
	dirPath string,
	tableNameValidator tableNameValidator,
) *JSON {
	return &JSON{
		serializer:         serializer,
		deserializer:       deserializer,
		dirPath:            dirPath,
		tableNameValidator: tableNameValidator,
	}
}

// InitJSON init all db staff with default settings and components.
func InitJSON() (*JSON, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("json db initing try to read homedir for current user, %w", err)
	}

	dirPath := filepath.Join(usr.HomeDir, defaultDBDir)

	_, err = os.Stat(dirPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("json db creating dir for db, %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("json db getting status of dir db, %w", err)
	}

	return NewJSON(
		json.Marshal,
		json.Unmarshal,
		dirPath,
		NewTableNameValidator(
			'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
			'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'_',
		),
	), nil
}

// Get fills objects if it's possible.
func (j *JSON) Get(tableName string, object interface{}) error {
	data, err := ioutil.ReadFile(filepath.Join(j.dirPath, tableName+defaultExtension))
	if err != nil {
		return fmt.Errorf("json db reading data from file, %w", err)
	}

	if err := j.deserializer(data, object); err != nil {
		return fmt.Errorf("json db deserializing, %w", err)
	}

	return nil
}

// Set data to table if it's possible.
func (j *JSON) Set(tableName string, object interface{}) error {
	if err := j.CreateTableIfNotExists(tableName); err != nil {
		return fmt.Errorf("creating file for table if not exists, %w", err)
	}

	data, err := j.serializer(object)
	if err != nil {
		return fmt.Errorf("json db serializeng, %w", err)
	}

	if err := ioutil.WriteFile(filepath.Join(j.dirPath, tableName+defaultExtension), data, 0644); err != nil {
		return fmt.Errorf("json db writing data, %w", err)
	}

	return nil
}

// CreateTableIfNotExists helps to make an easy idempotent set operation.
// If table is already exists, returns no error.
// It table isn't exists, tries to create it.
func (j *JSON) CreateTableIfNotExists(tableName string) error {
	if err := j.tableNameValidator.Validate(tableName); err != nil {
		return fmt.Errorf("validatting table name %s error, %w", tableName, err)
	}

	tablePath := filepath.Join(j.dirPath, tableName+defaultExtension)

	_, err := os.Stat(tablePath)
	if os.IsNotExist(err) {
		f, err := os.Create(tablePath)
		if err != nil {
			return fmt.Errorf("creating table %s, %w", tableName, err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("closing creatred file, %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("getting status of file, %w", err)
	}

	return nil
}
