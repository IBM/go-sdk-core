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
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Foo struct {
	Name *string `json:"name,omitempty"`
}

// Test a normal JSON-based response.
func TestGoodResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"name": "wonder woman"}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("POST")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	detailedResponse, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, http.StatusCreated, detailedResponse.StatusCode)
	assert.Equal(t, "application/json", detailedResponse.Headers.Get("Content-Type"))

	result, ok := detailedResponse.Result.(*Foo)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)
	assert.Equal(t, "wonder woman", *(result.Name))
}

// Verify that extra fields in result are silently ignored.
func TestGoodResponseJSONExtraFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"name": "wonder woman", "age": 42}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options, "watson", "watson")
	detailedResponse, _ := service.Request(req, new(Foo))
	result, ok := detailedResponse.Result.(*Foo)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)
	assert.Equal(t, "wonder woman", *result.Name)
}

// Test a non-JSON response.
func TestGoodResponseNonJSON(t *testing.T) {
	expectedResponse := []byte("This is a non-json response.")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/octet-stream")
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())
	detailedResponse, _ := service.Request(req, new(io.ReadCloser))
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, "application/octet-stream", detailedResponse.GetHeaders().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, detailedResponse.GetStatusCode())
	assert.NotNil(t, detailedResponse.Result)
	result, ok := detailedResponse.Result.(io.ReadCloser)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)

	// Read the bytes from the response body and verify.
	actualResponse, err := ioutil.ReadAll(result)
	assert.Nil(t, err)
	assert.NotNil(t, actualResponse)
	assert.Equal(t, expectedResponse, actualResponse)
}

// Test a non-JSON response with no Content-Type set.
func TestGoodResponseNonJSONNoContentType(t *testing.T) {
	expectedResponse := []byte("This is a non-json response.")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "")
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())
	detailedResponse, _ := service.Request(req, new(io.ReadCloser))
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, "", detailedResponse.GetHeaders().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, detailedResponse.GetStatusCode())
	assert.NotNil(t, detailedResponse.Result)
	result, ok := detailedResponse.Result.(io.ReadCloser)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)

	// Read the bytes from the response body and verify.
	actualResponse, err := ioutil.ReadAll(result)
	assert.Nil(t, err)
	assert.NotNil(t, actualResponse)
	assert.Equal(t, expectedResponse, actualResponse)
}

// Test a JSON response that causes a deserialization error.
func TestGoodResponseJSONDeserFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"name": {"unknown_object_id": "abc123"}}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options, "watson", "watson")
	detailedResponse, err := service.Request(req, new(Foo))
	assert.NotNil(t, detailedResponse)
	assert.NotNil(t, err)
	assert.NotNil(t, detailedResponse.RawResult)
	assert.Nil(t, detailedResponse.Result)
	assert.Equal(t,
		true,
		strings.HasPrefix(err.Error(), "An error occurred while unmarshalling the response body:"))
	// t.Log("Decode error:\n", err.Error())
}

// Test a good response with no response body.
func TestGoodResponseNoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	detailedResponse, err := service.Request(req, nil)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, 201, detailedResponse.StatusCode)
	assert.Equal(t, "", detailedResponse.Headers.Get("Content-Type"))
	assert.Nil(t, detailedResponse.Result)
	assert.Nil(t, detailedResponse.RawResult)
}

// Example of a JSON error structure.
var jsonErrorResponse string = `{
    "errors":[
        {
            "code":"error-vpc-1",
            "message":"Invalid value for 'param-1': bad value",
            "more_info":"https://myservice.com/more/info/about/the/error",
            "target":{
                "name":"param-1",
                "type":"parameter",
                "value":"bad value"
            }
        },
        {
            "code":"error-vpc-2",
            "message":"A validation error occurred for field 'field-1'.",
            "more_info":"https://myservice.com/more/info/about/the/error",
            "target":{
                "name":"field-1",
                "type":"field",
                "value":"invalid-field-1-value"
            }
        },
        {
            "code":"error-vpc-3",
            "message":"Unrecognized header found in request: X-CUSTOM-HEADER",
            "more_info":"https://myservice.com/more/info/about/the/error",
            "target":{
                "name":"X-CUSTOM-HEADER",
                "type":"header"
            }
        }
    ],
    "trace":"unique-error-identifier"
}`

// Example of a non-JSON error response body.
var nonJsonErrorResponse string = `This is a non-JSON error response body.`

// Test an error response with a JSON response body.
func TestErrorResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, jsonErrorResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options, "watson", "watson")
	response, err := service.Request(req, new(Foo))
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Result)
	assert.Nil(t, response.RawResult)
	errorMap, ok := response.GetResultAsMap()
	assert.Equal(t, true, ok)
	assert.NotNil(t, errorMap)
	assert.Equal(t, "Invalid value for 'param-1': bad value", err.Error())
	// t.Log("Error map contents:\n", errorMap)
}

// Test an error response with an invalid JSON response body.
func TestErrorResponseJSONDeserError(t *testing.T) {
	var expectedResponse = []byte(`"{"this is a malformed": "json object".......`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options, "watson", "watson")
	response, err := service.Request(req, new(Foo))
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.RawResult)
	assert.Equal(t, expectedResponse, response.RawResult)
	assert.Equal(t, http.StatusText(http.StatusForbidden), err.Error())
}

// Test error response with a non-JSON response body.
func TestErrorResponseNotJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, nonJsonErrorResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options, "watson", "watson")
	response, err := service.Request(req, new(Foo))
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.RawResult)
	s := string(response.RawResult)
	assert.Equal(t, nonJsonErrorResponse, s)
	assert.Equal(t, http.StatusText(http.StatusBadRequest), err.Error())
}

// Test an error response with no response body.
func TestErrorResponseNoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	detailedResponse, err := service.Request(req, new(Foo))
	assert.NotNil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, http.StatusInternalServerError, detailedResponse.StatusCode)
	assert.Equal(t, "", detailedResponse.Headers.Get("Content-Type"))
	assert.Nil(t, detailedResponse.Result)
	assert.Nil(t, detailedResponse.RawResult)
	assert.Equal(t, http.StatusText(http.StatusInternalServerError), err.Error())
}

func TestClient(t *testing.T) {
	mockClient := http.Client{}
	authenticator, _ := NewBasicAuthenticator("username", "password")
	service, _ := NewBaseService(&ServiceOptions{Authenticator: authenticator}, "watson", "watson")
	service.SetHTTPClient(&mockClient)
	assert.ObjectsAreEqual(mockClient, service.Client)
}

func TestRequestForDefaultUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"name": "wonder woman"}`)
		assert.Contains(t, r.Header.Get("User-Agent"), "ibm-go-sdk-core")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator, _ := NewBasicAuthenticator("username", "password")
	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	_, _ = service.Request(req, new(Foo))
}

func TestRequestForProvidedUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"name": "wonder woman"}`)
		assert.Contains(t, r.Header.Get("User-Agent"), "provided user agent")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}
	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	headers := http.Header{}
	headers.Add("User-Agent", "provided user agent")
	service.SetDefaultHeaders(headers)
	_, _ = service.Request(req, new(Foo))
}

func TestIncorrectURL(t *testing.T) {
	authenticator, _ := NewNoAuthAuthenticator()
	options := &ServiceOptions{
		URL:           "{xxx}",
		Authenticator: authenticator,
	}
	_, serviceErr := NewBaseService(options, "watson", "watson")
	expectedError := fmt.Errorf(ERRORMSG_PROP_INVALID, "URL")
	assert.Equal(t, expectedError.Error(), serviceErr.Error())
}

func TestDisableSSLVerification(t *testing.T) {
	options := &ServiceOptions{
		URL:           "test.com",
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.Nil(t, service.Client.Transport)
	service.DisableSSLVerification()
	assert.NotNil(t, service.Client.Transport)
}

func TestBasicAuth1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		username, password, ok := r.BasicAuth()
		assert.Equal(t, ok, true)
		assert.Equal(t, "mookie", username)
		assert.Equal(t, "betts", password)
	}))
	defer server.Close()

	options := &ServiceOptions{
		URL: server.URL,
		Authenticator: &BasicAuthenticator{
			Username: "mookie",
			Password: "betts",
		},
	}

	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestBasicAuth2(t *testing.T) {
	firstTime := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		username, password, ok := r.BasicAuth()
		assert.Equal(t, ok, true)
		if firstTime {
			assert.Equal(t, "foo", username)
			assert.Equal(t, "bar", password)
			firstTime = false
		} else {
			assert.Equal(t, "mookie", username)
			assert.Equal(t, "betts", password)
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL: server.URL,
		Authenticator: &BasicAuthenticator{
			Username: "foo",
			Password: "bar",
		},
	}

	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)

	service.Options.Authenticator = &BasicAuthenticator{
		Username: "mookie",
		Password: "betts",
	}

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestBasicAuthConfigError(t *testing.T) {
	options := &ServiceOptions{
		URL: "https://myservice",
		Authenticator: &BasicAuthenticator{
			Username: "mookie",
			Password: "",
		},
	}

	service, err := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, err)
	assert.Nil(t, service)
}

func TestNoAuth1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "", r.Header.Get("Authorization"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}

	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestNoAuth2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "", r.Header.Get("Authorization"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL: server.URL,
		Authenticator: &BasicAuthenticator{
			Username: "foo",
			Password: "bar",
		},
	}

	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	service.Options.Authenticator = &NoAuthAuthenticator{}
	assert.Nil(t, err)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestIAMAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "grant_type") {
			fmt.Fprint(w, `{
				"access_token": "captain marvel",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": 1524167011,
				"refresh_token": "jy4gl91BQ"
			}`)
			assert.Equal(t, "", r.Header.Get("Authorization"))
		} else {
			assert.Equal(t, "Bearer captain marvel", r.Header.Get("Authorization"))
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL: server.URL,
		Authenticator: &IamAuthenticator{
			URL:    server.URL,
			ApiKey: "xxxxx",
		},
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service.Options.Authenticator.AuthenticationType())

	_, err = service.Request(req, new(Foo))
	if err != nil {
		fmt.Println("Error: ", err)
	}
	assert.Nil(t, err)
}

func TestIAMFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL: server.URL,
		Authenticator: &IamAuthenticator{
			URL:    server.URL,
			ApiKey: "xxxxx",
		},
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	_, err = service.Request(req, new(Foo))
	assert.NotNil(t, err)
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

func TestIAMWithIdSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "grant_type") {
			fmt.Fprint(w, `{
                "access_token": "captain marvel",
                "token_type": "Bearer",
                "expires_in": 3600,
                "expiration": 1524167011,
                "refresh_token": "jy4gl91BQ"
            }`)
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "mookie", username)
			assert.Equal(t, "betts", password)
		} else {
			assert.Equal(t, "Bearer captain marvel", r.Header.Get("Authorization"))
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL: server.URL,
		Authenticator: &IamAuthenticator{
			URL:          server.URL,
			ApiKey:       "xxxxx",
			ClientId:     "mookie",
			ClientSecret: "betts",
		},
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestIAMErrorClientIdOnly(t *testing.T) {
	_, err := NewBaseService(
		&ServiceOptions{
			URL: "don't care",
			Authenticator: &IamAuthenticator{
				ApiKey:   "xxxxx",
				ClientId: "foo",
			},
		}, "watson", "watson")
	assert.NotNil(t, err)
}

func TestIAMErrorClientSecretOnly(t *testing.T) {
	_, err := NewBaseService(
		&ServiceOptions{
			URL: "don't care",
			Authenticator: &IamAuthenticator{
				ApiKey:       "xxxxx",
				ClientSecret: "bar",
			},
		}, "watson", "watson")
	assert.NotNil(t, err)
}

func TestIAMNoApiKey(t *testing.T) {
	_, err := NewBaseService(
		&ServiceOptions{
			URL: "don't care",
			Authenticator: &IamAuthenticator{
				URL:          "don't care",
				ClientId:     "foo",
				ClientSecret: "bar",
			},
		}, "watson", "watson")
	assert.NotNil(t, err)
}

func TestCP4DAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.String(), "preauth") {
			fmt.Fprint(w, `{
			"username":"hello",
			"role":"user",
			"permissions":[
				"administrator",
				"deployment_admin"
			],
			"sub":"hello",
			"iss":"John",
			"aud":"DSX",
			"uid":"999",
			"accessToken":"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
			"_messageCode_":"success",
			"message":"success"
		}`)
		} else {
			assert.Equal(t, "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI", r.Header.Get("Authorization"))
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL: server.URL,
		Authenticator: &CloudPakForDataAuthenticator{
			URL:      server.URL,
			Username: "bogus",
			Password: "bogus",
		},
	}

	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestCP4DFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL: server.URL,
		Authenticator: &CloudPakForDataAuthenticator{
			URL:      server.URL,
			Username: "bogus",
			Password: "bogus",
		},
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	_, err = service.Request(req, new(Foo))
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

// Test for the deprecated SetURL method.
func TestSetURL(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &IamAuthenticator{
				ApiKey: "xxxxx",
			},
		}, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)

	err = service.SetURL("{bad url}")
	assert.NotNil(t, err)
}

func TestSetServiceURL(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
		}, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)

	err = service.SetServiceURL("{bad url}")
	assert.NotNil(t, err)

	err = service.SetServiceURL("")
	assert.Nil(t, err)
	assert.Equal(t, "", service.Options.URL)
	assert.Equal(t, "", service.GetServiceURL())

	err = service.SetServiceURL("https://myserver.com/api/baseurl")
	assert.Nil(t, err)
	assert.Equal(t, "https://myserver.com/api/baseurl", service.Options.URL)
	assert.Equal(t, "https://myserver.com/api/baseurl", service.GetServiceURL())
}

func TestExtConfigFromCredentialFile(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		}, "service-1", "service-1")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service1/api", service.Options.URL)
	assert.NotNil(t, service.Client.Transport)

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestExtConfigError(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		}, "error4", "error4")
	assert.NotNil(t, err)
	assert.Nil(t, service)

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestExtConfigFromEnvironment(t *testing.T) {
	setTestEnvironment()

	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		}, "service3", "service3")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service3/api", service.Options.URL)
	assert.Nil(t, service.Client.Transport)

	clearTestEnvironment()
}

func TestExtConfigFromVCAP(t *testing.T) {
	setTestVCAP()

	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		}, "service2", "service2")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service2/api", service.Options.URL)
	assert.Nil(t, service.Client.Transport)

	clearTestVCAP()
}

func TestConfigureServiceFromCredFile(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		}, "service_1", "service_1")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "bad url", service.Options.URL)
	assert.Nil(t, service.Client.Transport)

	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

	err = service.ConfigureService("service5")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service5/api", service.Options.URL)
	assert.NotNil(t, service.Client.Transport)

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestConfigureServiceFromVCAP(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		}, "service2", "service2")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "bad url", service.Options.URL)

	setTestVCAP()
	err = service.ConfigureService("service3")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service3/api", service.Options.URL)
	assert.Nil(t, service.Client.Transport)

	clearTestVCAP()
}

func TestConfigureServiceFromEnv(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		}, "service_1", "service_1")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "bad url", service.Options.URL)
	assert.Nil(t, service.Client.Transport)

	setTestEnvironment()
	err = service.ConfigureService("service_1")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service1/api", service.Options.URL)
	assert.NotNil(t, service.Client.Transport)

	clearTestEnvironment()
}

func TestConfigureServiceError(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		}, "service-1", "service-1")
	assert.Nil(t, err)
	err = service.ConfigureService("")
	assert.NotNil(t, err)
	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestAuthNotConfigured(t *testing.T) {
	service, err := NewBaseService(&ServiceOptions{}, "noauth_service", "noauth_service")
	assert.NotNil(t, err)
	assert.Nil(t, service)
}

func testGetErrorMessage(t *testing.T, statusCode int, jsonString string, expectedErrorMsg string) {
	body := []byte(jsonString)
	responseMap, err := decodeAsMap(body)
	assert.Nil(t, err)

	actualErrorMsg := getErrorMessage(responseMap, statusCode)
	assert.Equal(t, expectedErrorMsg, actualErrorMsg)
}

func TestErrorMessage(t *testing.T) {
	testGetErrorMessage(t, http.StatusBadRequest, `{"error":"error1"}`, "error1")

	testGetErrorMessage(t, http.StatusBadRequest, `{"message":"error2"}`, "error2")

	testGetErrorMessage(t, http.StatusBadRequest, `{"errors":[{"message":"error3"}]}`, "error3")

	testGetErrorMessage(t, http.StatusForbidden, `{"msg":"error4"}`, http.StatusText(http.StatusForbidden))

	testGetErrorMessage(t, http.StatusBadRequest, `{"errorMessage":"error5"}`, "error5")
}
