//go:build all || fast || basesvc
// +build all fast basesvc

package core

// (C) Copyright IBM Corp. 2019.
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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStructure struct {
	Name string `json:"name"`
}

func TestDetailedResponseJsonSuccess(t *testing.T) {
	testStructure := TestStructure{
		Name: "wonder woman",
	}

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	response := &DetailedResponse{
		StatusCode: 200,
		Result:     testStructure,
		Headers:    headers,
	}
	assert.Equal(t, 200, response.GetStatusCode())
	assert.Equal(t, "application/json", response.GetHeaders().Get("Content-Type"))
	assert.Equal(t, testStructure, response.GetResult())
	assert.Nil(t, response.GetRawResult())
	m, ok := response.GetResultAsMap()
	assert.Equal(t, false, ok)
	assert.Nil(t, m)

	s := response.String()
	assert.NotEmpty(t, s)
	t.Logf("detailed response:\n%s", s)
}

func TestDetailedResponseNonJson(t *testing.T) {
	responseBody := []byte(`This is a non-json response body.`)

	headers := http.Header{}
	headers.Add("Content-Type", "application/octet-stream")

	response := &DetailedResponse{
		StatusCode: 200,
		RawResult:  responseBody,
		Headers:    headers,
	}
	assert.Equal(t, 200, response.GetStatusCode())
	assert.Equal(t, "application/octet-stream", response.GetHeaders().Get("Content-Type"))
	assert.Equal(t, responseBody, response.GetRawResult())
	assert.Nil(t, response.GetResult())
	m, ok := response.GetResultAsMap()
	assert.Equal(t, false, ok)
	assert.Nil(t, m)
}

func TestDetailedResponseJsonMap(t *testing.T) {
	errorMap := make(map[string]interface{})
	errorMap["message"] = "An error message."

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")

	response := &DetailedResponse{
		StatusCode: 400,
		Result:     errorMap,
		Headers:    headers,
	}
	assert.Equal(t, 400, response.GetStatusCode())
	assert.Equal(t, "application/json", response.GetHeaders().Get("Content-Type"))
	m, ok := response.GetResultAsMap()
	assert.Equal(t, true, ok)
	assert.Equal(t, errorMap, m)
	assert.Nil(t, response.GetRawResult())
}
