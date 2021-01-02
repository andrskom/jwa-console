package filedb

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestSerializer(
	t *testing.T,
	expectedObj interface{},
	resultData []byte,
	resultErr error,
) func(obj interface{}) ([]byte, error) {
	return func(obj interface{}) ([]byte, error) {
		require.Equal(t, expectedObj, obj)

		return resultData, resultErr
	}
}

func getTestDeserializer(
	t *testing.T,
	expectedObj interface{},
	expectedData []byte,
	resultErr error,
) func(data []byte, obj interface{}) error {
	return func(data []byte, obj interface{}) error {
		require.Equal(t, expectedData, data)
		require.Equal(t, expectedObj, obj)

		return resultErr
	}
}

func TestJSON_CreateTableIfNotExists_ValidatorErr_ExpectedErr(t *testing.T) {
	validator := NewTableNameValidator('b')
	tSerializer := getTestSerializer(t, nil, nil, nil)
	tDeserializer := getTestDeserializer(t, nil, nil, nil)

	db := NewJSON(tSerializer, tDeserializer, "", validator)

	err := db.CreateTableIfNotExists("a")
	assert.True(t, errors.Is(err, ErrUnexpectedRuneInTableName))
}

func TestJSON_CreateTableIfNotExists_FileAlreadyExists_ExpectedNil(t *testing.T) {
	validator := NewTableNameValidator('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')
	tSerializer := getTestSerializer(t, nil, nil, nil)
	tDeserializer := getTestDeserializer(t, nil, nil, nil)

	tmpDir := os.TempDir()
	f, err := ioutil.TempFile(tmpDir, "*.json")
	require.NoError(t, err)
	require.NoError(t, f.Close())
	defer os.Remove(f.Name())

	db := NewJSON(tSerializer, tDeserializer, tmpDir, validator)

	err = db.CreateTableIfNotExists(strings.TrimRight(filepath.Base(f.Name()), ".json"))
	assert.NoError(t, err)
}

func TestJSON_CreateTableIfNotExists_CreatingFile_ExpectedNil(t *testing.T) {
	validator := NewTableNameValidator('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')
	tSerializer := getTestSerializer(t, nil, nil, nil)
	tDeserializer := getTestDeserializer(t, nil, nil, nil)

	tmpDir := os.TempDir()
	uniqName := strconv.FormatInt(time.Now().UnixNano(), 10)
	defer os.Remove(filepath.Join(tmpDir, uniqName+".json"))

	db := NewJSON(tSerializer, tDeserializer, tmpDir, validator)

	err := db.CreateTableIfNotExists(uniqName)
	assert.NoError(t, err)
}

func TestJSON_Set_TableCreatingErr_ExpctedErr(t *testing.T) {
	validator := NewTableNameValidator()
	tSerializer := getTestSerializer(t, nil, nil, nil)
	tDeserializer := getTestDeserializer(t, nil, nil, nil)

	tmpDir := os.TempDir()
	uniqName := strconv.FormatInt(time.Now().UnixNano(), 10)
	defer os.Remove(filepath.Join(tmpDir, uniqName+".json"))

	db := NewJSON(tSerializer, tDeserializer, tmpDir, validator)

	data := "a"
	err := db.Set(uniqName, data)
	assert.True(t, errors.Is(err, ErrUnexpectedRuneInTableName))
}

func TestJSON_Set_SerializationErr_ExpectedErr(t *testing.T) {
	eErr := errors.New("b")

	validator := NewTableNameValidator('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')
	tSerializer := getTestSerializer(t, "a", nil, eErr)
	tDeserializer := getTestDeserializer(t, nil, nil, nil)

	tmpDir := os.TempDir()
	uniqName := strconv.FormatInt(time.Now().UnixNano(), 10)
	defer os.Remove(filepath.Join(tmpDir, uniqName+".json"))

	db := NewJSON(tSerializer, tDeserializer, tmpDir, validator)

	data := "a"
	err := db.Set(uniqName, data)
	assert.True(t, errors.Is(err, eErr))
}

func TestJSON_Get_DeserializationErr_ExpectedErr(t *testing.T) {
	eErr := errors.New("b")

	validator := NewTableNameValidator('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')
	tDeserializer := getTestDeserializer(t, "", []byte(`"a"`), eErr)

	tmpDir := os.TempDir()
	uniqName := strconv.FormatInt(time.Now().UnixNano(), 10)
	defer os.Remove(filepath.Join(tmpDir, uniqName+".json"))

	db := NewJSON(json.Marshal, tDeserializer, tmpDir, validator)

	require.NoError(t, db.Set(uniqName, "a"))

	var data string

	err := db.Get(uniqName, data)
	assert.True(t, errors.Is(err, eErr))
}

func TestJSON_Get_PositiveCase_NoErr(t *testing.T) {
	validator := NewTableNameValidator('0', '1', '2', '3', '4', '5', '6', '7', '8', '9')

	tmpDir := os.TempDir()
	uniqName := strconv.FormatInt(time.Now().UnixNano(), 10)
	defer os.Remove(filepath.Join(tmpDir, uniqName+".json"))

	db := NewJSON(json.Marshal, json.Unmarshal, tmpDir, validator)

	require.NoError(t, db.Set(uniqName, "a"))

	var data string

	err := db.Get(uniqName, &data)
	assert.NoError(t, err)
	assert.Equal(t, "a", data)
}
