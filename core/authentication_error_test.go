package core

// (C) Copyright IBM Corp. 2024.
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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmbedsHTTPProblem(t *testing.T) {
	authErr := &AuthenticationError{
		Err: errors.New(""),
		HTTPProblem: &HTTPProblem{
			OperationID: "",
			Response:    &DetailedResponse{},
		},
	}

	assert.NotNil(t, authErr.GetConsoleMessage)
	assert.NotNil(t, authErr.GetDebugMessage)
	assert.NotNil(t, authErr.GetID)
	assert.NotNil(t, authErr.getErrorCode)
	assert.NotNil(t, authErr.getHeader)
	assert.NotNil(t, authErr.GetConsoleOrderedMaps)
	assert.NotNil(t, authErr.GetDebugOrderedMaps)
}

func TestNewAuthenticationError(t *testing.T) {
	unknown := "unknown"
	err := errors.New("test")
	resp := getMockAuthResponse()

	authErr := NewAuthenticationError(resp, err)

	assert.NotNil(t, authErr)
	assert.Equal(t, err, authErr.Err)
	assert.Equal(t, resp, authErr.Response)
	assert.Equal(t, "test", authErr.Summary)
	assert.Equal(t, unknown, authErr.OperationID)
	assert.Equal(t, unknown, authErr.Component.Name)
	assert.Equal(t, unknown, authErr.Component.Version)
}

func TestAuthenticationErrorfHTTPProblem(t *testing.T) {
	resp := getMockAuthResponse()
	httpProb := httpErrorf("Unauthorized", resp)
	assert.Empty(t, httpProb.OperationID)
	assert.Empty(t, httpProb.Component)

	authErr := authenticationErrorf(httpProb, nil, "get_token", NewProblemComponent("iam", "v1"))
	assert.Equal(t, httpProb, authErr.Err)
	assert.Equal(t, resp, authErr.Response)
	assert.Equal(t, "Unauthorized", authErr.Summary)
	assert.Equal(t, "get_token", authErr.OperationID)
	assert.Equal(t, "iam", authErr.Component.Name)
	assert.Equal(t, "v1", authErr.Component.Version)
}

func TestAuthenticationErrorfOtherError(t *testing.T) {
	err := errors.New("test")
	resp := getMockAuthResponse()

	authErr := authenticationErrorf(err, resp, "get_token", NewProblemComponent("iam", "v1"))
	assert.NotNil(t, authErr)
	assert.Equal(t, err, authErr.Err)
	assert.Equal(t, resp, authErr.Response)
	assert.Equal(t, "test", authErr.Summary)
	assert.Equal(t, "get_token", authErr.OperationID)
	assert.Equal(t, "iam", authErr.Component.Name)
	assert.Equal(t, "v1", authErr.Component.Version)
}

func TestAuthenticationErrorfNoErr(t *testing.T) {
	authErr := authenticationErrorf(nil, getMockAuthResponse(), "get_token", NewProblemComponent("iam", "v1"))
	assert.Nil(t, authErr)
}

func TestAuthenticationErrorfNoResponse(t *testing.T) {
	authErr := authenticationErrorf(errors.New("not http, needs response"), nil, "get_token", NewProblemComponent("iam", "v1"))
	assert.Nil(t, authErr)
}

func getMockAuthResponse() *DetailedResponse {
	return &DetailedResponse{
		StatusCode: 401,
		Result:     "Unauthorized",
	}
}
