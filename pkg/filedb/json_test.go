package filedb

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// marshalIndent is a test func for beautify file data.
// It's a copy of https://pkg.go.dev/encoding/json#MarshalIndent from stdlib with predefined prefix and indent.
func marshalIndent(v interface{}) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, "", " ")
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestTableNameValidator_Validate(t *testing.T) {
	validator := NewTableNameValidator('a', 'b')

	t.Run("name contains expected chars", func(t *testing.T) {
		a := assert.New(t)

		a.NoError(validator.Validate("abba"))
	})

	t.Run("name contains not expected chars", func(t *testing.T) {
		a := assert.New(t)

		err := validator.Validate("abba_")
		a.True(errors.Is(err, ErrUnexpectedRuneInTableName), "Unexpected err received")
	})
}
