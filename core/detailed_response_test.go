package core

/**
 * Copyright 2019 IBM All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStructure struct {
	Name string `json:"name"`
}

func TestDetailedResponseJson(t *testing.T) {
	testStructure := TestStructure{
		Name: "wonder woman",
	}

	headers := http.Header{}
	headers.Add("accept", "application/json")

	response := &DetailedResponse{
		StatusCode: 200,
		Result:     testStructure,
		Headers:    headers,
	}
	assert.Equal(t, response.GetResult(), testStructure)
	assert.Equal(t, response.GetStatusCode(), 200)
	assert.Equal(t, response.GetHeaders().Get("accept"), "application/json")
	response.String()
}

func TestDetailedResponseNonJson(t *testing.T) {
	response := &DetailedResponse{
		StatusCode: 200,
		Result:     make(chan int),
	}
	assert.Equal(t, response.GetStatusCode(), 200)
	fmt.Println(response.String())
}
