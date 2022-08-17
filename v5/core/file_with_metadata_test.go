package core

// (C) Copyright IBM Corp. 2021.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileWithMetadataFields(t *testing.T) {
	data := io.NopCloser(bytes.NewReader([]byte("test")))
	filename := "test.txt"
	contentType := "application/octet-stream"

	model := FileWithMetadata{
		Data:        data,
		Filename:    &filename,
		ContentType: &contentType,
	}

	assert.NotNil(t, model.Data)
	assert.NotNil(t, model.Filename)
	assert.NotNil(t, model.ContentType)
}

func TestNewFileWithMetadata(t *testing.T) {
	data := io.NopCloser(bytes.NewReader([]byte("test")))
	model, err := NewFileWithMetadata(data)

	assert.Nil(t, err)
	myData := model.Data
	assert.NotNil(t, myData)

	assert.Nil(t, model.Filename)
	assert.Nil(t, model.ContentType)
}

func TestUnmarshalFileWithMetadata(t *testing.T) {
	var err error

	// setup the test by creating a temp directory and file for the unmarshaler to read
	err = os.Mkdir("tempdir", 0755)
	assert.Nil(t, err)

	message := []byte("test")
	err = os.WriteFile("tempdir/test-file.txt", message, 0644)
	assert.Nil(t, err)

	// mock what user input would look like - a map converted from a JSON string
	exampleJsonString := `{"data": "tempdir/test-file.txt", "filename": "test-file.txt", "content_type": "text/plain"}`

	var mapifiedString map[string]json.RawMessage
	err = json.Unmarshal([]byte(exampleJsonString), &mapifiedString)
	assert.Nil(t, err)

	var model *FileWithMetadata

	err = UnmarshalFileWithMetadata(mapifiedString, &model)
	assert.Nil(t, err)

	data := model.Data
	assert.NotNil(t, data)

	assert.NotNil(t, model.Filename)
	assert.Equal(t, "test-file.txt", *model.Filename)

	assert.NotNil(t, model.ContentType)
	assert.Equal(t, "text/plain", *model.ContentType)

	err = os.RemoveAll("tempdir")
	assert.Nil(t, err)
}
