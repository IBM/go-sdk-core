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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPProblemEmbedsIBMProblem(t *testing.T) {
	httpProb := &HTTPProblem{}

	// Check that the methods defined by IBMProblem are supported here.
	// The implementations are tested elsewhere.
	assert.NotNil(t, httpProb.Error)
	assert.NotNil(t, httpProb.GetBaseSignature)
	assert.NotNil(t, httpProb.GetCausedBy)
	assert.NotNil(t, httpProb.Unwrap)
}

func TestHTTPProblemGetConsoleMessage(t *testing.T) {
	httpProb := getPopulatedHTTPProblem()
	message := httpProb.GetConsoleMessage()
	expected := `---
id: http-1850a8fa
summary: Bad request
severity: error
operation_id: create_resource
status_code: 400
error_code: invalid-input
component:
  name: my-service
  version: v1
request_id: abc123
correlation_id: xyz789
---
`
	assert.Equal(t, expected, message)
}

func TestHTTPProblemGetDebugMessage(t *testing.T) {
	httpProb := getPopulatedHTTPProblem()
	message := httpProb.GetDebugMessage()
	expected := `---
id: http-1850a8fa
summary: Bad request
severity: error
operation_id: create_resource
status_code: 400
error_code: invalid-input
component:
  name: my-service
  version: v1
request_id: abc123
correlation_id: xyz789
response:
  status_code: 400
  headers:
    Content-Type:
    - application/json
    X-Correlation-Id:
    - xyz789
    X-Request-Id:
    - abc123
  result:
    errorCode: invalid-input
---
`
	assert.Equal(t, expected, message)
}

func TestHTTPProblemGetID(t *testing.T) {
	httpProb := getPopulatedHTTPProblem()
	assert.Equal(t, "http-1850a8fa", httpProb.GetID())
}

func TestHTTPProblemGetConsoleOrderedMaps(t *testing.T) {
	httpProb := getPopulatedHTTPProblem()
	orderedMaps := httpProb.GetConsoleOrderedMaps()
	assert.NotNil(t, orderedMaps)

	maps := orderedMaps.GetMaps()
	assert.NotNil(t, maps)
	assert.Len(t, maps, 9)

	assert.Equal(t, "id", maps[0].Key)
	assert.Equal(t, "http-1850a8fa", maps[0].Value)

	assert.Equal(t, "summary", maps[1].Key)
	assert.Equal(t, "Bad request", maps[1].Value)

	assert.Equal(t, "severity", maps[2].Key)
	assert.Equal(t, ErrorSeverity, maps[2].Value)

	assert.Equal(t, "operation_id", maps[3].Key)
	assert.Equal(t, "create_resource", maps[3].Value)

	assert.Equal(t, "status_code", maps[4].Key)
	assert.Equal(t, 400, maps[4].Value)

	assert.Equal(t, "error_code", maps[5].Key)
	assert.Equal(t, "invalid-input", maps[5].Value)

	assert.Equal(t, "component", maps[6].Key)
	assert.Equal(t, "my-service", maps[6].Value.(*ProblemComponent).Name)
	assert.Equal(t, "v1", maps[6].Value.(*ProblemComponent).Version)

	assert.Equal(t, "request_id", maps[7].Key)
	assert.Equal(t, "abc123", maps[7].Value)

	assert.Equal(t, "correlation_id", maps[8].Key)
	assert.Equal(t, "xyz789", maps[8].Value)
}

func TestHTTPProblemGetDebugOrderedMaps(t *testing.T) {
	httpProb := getPopulatedHTTPProblem()
	orderedMaps := httpProb.GetDebugOrderedMaps()
	assert.NotNil(t, orderedMaps)

	maps := orderedMaps.GetMaps()
	assert.NotNil(t, maps)
	assert.Len(t, maps, 10)

	assert.Equal(t, "id", maps[0].Key)
	assert.Equal(t, "http-1850a8fa", maps[0].Value)

	assert.Equal(t, "summary", maps[1].Key)
	assert.Equal(t, "Bad request", maps[1].Value)

	assert.Equal(t, "severity", maps[2].Key)
	assert.Equal(t, ErrorSeverity, maps[2].Value)

	assert.Equal(t, "operation_id", maps[3].Key)
	assert.Equal(t, "create_resource", maps[3].Value)

	assert.Equal(t, "status_code", maps[4].Key)
	assert.Equal(t, 400, maps[4].Value)

	assert.Equal(t, "error_code", maps[5].Key)
	assert.Equal(t, "invalid-input", maps[5].Value)

	assert.Equal(t, "component", maps[6].Key)
	assert.Equal(t, "my-service", maps[6].Value.(*ProblemComponent).Name)
	assert.Equal(t, "v1", maps[6].Value.(*ProblemComponent).Version)

	assert.Equal(t, "request_id", maps[7].Key)
	assert.Equal(t, "abc123", maps[7].Value)

	assert.Equal(t, "correlation_id", maps[8].Key)
	assert.Equal(t, "xyz789", maps[8].Value)

	assert.Equal(t, "response", maps[9].Key)
	assert.Equal(t, *getPopulatedDetailedResponse(), maps[9].Value)
}

func TestHTTPProblemGetDebugOrderedMapsWithoutOptionals(t *testing.T) {
	httpProb := getPopulatedHTTPProblemWithoutOptionals()
	orderedMaps := httpProb.GetDebugOrderedMaps()
	assert.NotNil(t, orderedMaps)

	maps := orderedMaps.GetMaps()
	assert.NotNil(t, maps)
	assert.Len(t, maps, 7)

	assert.Equal(t, "id", maps[0].Key)
	assert.Equal(t, "http-1850a8fa", maps[0].Value)

	assert.Equal(t, "summary", maps[1].Key)
	assert.Equal(t, "Bad request", maps[1].Value)

	assert.Equal(t, "severity", maps[2].Key)
	assert.Equal(t, ErrorSeverity, maps[2].Value)

	assert.Equal(t, "operation_id", maps[3].Key)
	assert.Equal(t, "create_resource", maps[3].Value)

	assert.Equal(t, "status_code", maps[4].Key)
	assert.Equal(t, 400, maps[4].Value)

	assert.Equal(t, "component", maps[5].Key)
	assert.Equal(t, "my-service", maps[5].Value.(*ProblemComponent).Name)
	assert.Equal(t, "v1", maps[5].Value.(*ProblemComponent).Version)

	assert.Equal(t, "response", maps[6].Key)
	assert.Equal(t, DetailedResponse{StatusCode: 400, Result: ""}, maps[6].Value)
}

func TestHTTPProblemGetHeader(t *testing.T) {
	httpProb := httpErrorf("Bad request", getPopulatedDetailedResponse())
	val, ok := httpProb.getHeader("doesnt-exist")
	assert.Empty(t, val)
	assert.False(t, ok)

	val, ok = httpProb.getHeader("content-type")
	assert.Equal(t, "application/json", val)
	assert.True(t, ok)
}

func TestHTTPProblemGetErrorCodeEmpty(t *testing.T) {
	httpProb := httpErrorf("Bad request", &DetailedResponse{})
	assert.Empty(t, httpProb.getErrorCode())
}

func TestHTTPProblemGetErrorCode(t *testing.T) {
	httpProb := httpErrorf("Bad request", getPopulatedDetailedResponse())
	assert.Equal(t, "invalid-input", httpProb.getErrorCode())
}

func TestHTTPProblemIsWithProblem(t *testing.T) {
	firstProb := httpErrorf("Bad request", getPopulatedDetailedResponse())
	EnrichHTTPProblem(firstProb, "create_resource", NewProblemComponent("service", "1.0.0"))

	secondProb := httpErrorf("Invalid input", getPopulatedDetailedResponse())
	EnrichHTTPProblem(secondProb, "create_resource", NewProblemComponent("service", "1.2.3"))

	assert.NotEqual(t, firstProb, secondProb)
	assert.True(t, errors.Is(firstProb, secondProb))
}

func TestHTTPErrorf(t *testing.T) {
	message := "Bad request"
	httpProb := httpErrorf(message, getPopulatedDetailedResponse())

	// We don't have a lot of information about the request when we
	// create new HTTPProblem objects here in the core.
	assert.NotNil(t, httpProb)
	assert.Equal(t, message, httpProb.Summary)
	assert.Equal(t, getPopulatedDetailedResponse(), httpProb.Response)
	assert.Empty(t, httpProb.discriminator)
	assert.Empty(t, httpProb.Component)
	assert.Empty(t, httpProb.OperationID)
	assert.Nil(t, httpProb.causedBy)
}

func TestPublicEnrichHTTPProblem(t *testing.T) {
	err := httpErrorf("Bad request", &DetailedResponse{})
	assert.Empty(t, err.Component)
	assert.Empty(t, err.OperationID)

	EnrichHTTPProblem(err, "delete_resource", NewProblemComponent("test", "v2"))

	assert.NotEmpty(t, err.Component)
	assert.Equal(t, "test", err.Component.Name)
	assert.Equal(t, "v2", err.Component.Version)
	assert.Equal(t, "delete_resource", err.OperationID)
}

func TestPublicEnrichHTTPProblemWithinSDKProblem(t *testing.T) {
	httpProb := httpErrorf("Bad request", &DetailedResponse{})
	assert.Empty(t, httpProb.Component)
	assert.Empty(t, httpProb.OperationID)

	sdkProb := SDKErrorf(httpProb, "Wrong!", "", NewProblemComponent("sdk", "1.0.0"))
	EnrichHTTPProblem(sdkProb, "delete_resource", NewProblemComponent("test", "v2"))

	assert.NotEmpty(t, httpProb.Component)
	assert.Equal(t, "test", httpProb.Component.Name)
	assert.Equal(t, "v2", httpProb.Component.Version)
	assert.Equal(t, "delete_resource", httpProb.OperationID)
}

func TestPrivateEnrichHTTPProblem(t *testing.T) {
	httpProb := httpErrorf("Bad request", &DetailedResponse{})
	assert.Empty(t, httpProb.Component)
	assert.Empty(t, httpProb.OperationID)

	enrichHTTPProblem(httpProb, "delete_resource", NewProblemComponent("test", "v2"))
	assert.NotEmpty(t, httpProb.Component)
	assert.Equal(t, "test", httpProb.Component.Name)
	assert.Equal(t, "v2", httpProb.Component.Version)
	assert.Equal(t, "delete_resource", httpProb.OperationID)
}

func getPopulatedHTTPProblem() *HTTPProblem {
	return &HTTPProblem{
		IBMProblem: &IBMProblem{
			Summary:       "Bad request",
			Component:     NewProblemComponent("my-service", "v1"),
			Severity:      ErrorSeverity,
			discriminator: "some-issue",
		},
		OperationID: "create_resource",
		Response:    getPopulatedDetailedResponse(),
	}
}

func getPopulatedHTTPProblemWithoutOptionals() *HTTPProblem {
	return &HTTPProblem{
		IBMProblem: &IBMProblem{
			Summary:       "Bad request",
			Component:     NewProblemComponent("my-service", "v1"),
			Severity:      ErrorSeverity,
			discriminator: "some-issue",
		},
		OperationID: "create_resource",
		Response: &DetailedResponse{
			StatusCode: 400,
		},
	}
}

func getPopulatedDetailedResponse() *DetailedResponse {
	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	headers.Add("X-Request-ID", "abc123")
	headers.Add("X-Correlation-ID", "xyz789")

	return &DetailedResponse{
		StatusCode: 400,
		Headers:    headers,
		Result: map[string]interface{}{
			"errorCode": "invalid-input",
		},
	}
}
