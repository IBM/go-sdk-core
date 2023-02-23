//go:build all || fast || basesvc
// +build all fast basesvc

package core

// (C) Copyright IBM Corp. 2019, 2022.
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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
)

// getRetryableHTTPClient returns the "retryable" Client hidden inside the specified http.Client instance
// or nil if "client" is not hiding a retryable Client instance.
func getRetryableHTTPClient(client *http.Client) *retryablehttp.Client {
	if client != nil {
		if client.Transport != nil {
			// A retryable client will have its Transport field set to an
			// instance of retryablehttp.RoundTripper.
			if rt, ok := client.Transport.(*retryablehttp.RoundTripper); ok {
				return rt.Client
			}
		}
	}
	return nil
}

func TestClone(t *testing.T) {
	var service *BaseService = nil
	var err error

	// Verify nil.Clone() == nil
	assert.Nil(t, service.Clone())

	// Verify a non-nil service is cloned correctly.
	options := &ServiceOptions{
		URL:           "https://myservice.ibm.com/api/v1",
		Authenticator: &NoAuthAuthenticator{},
	}
	service, err = NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, "https://myservice.ibm.com/api/v1", service.Options.URL)

	clone := service.Clone()
	assert.NotNil(t, clone)
	assert.Equal(t, service.Client, clone.Client)
	assert.Equal(t, service.UserAgent, clone.UserAgent)
	assert.Equal(t, service.DefaultHeaders, clone.DefaultHeaders)
	assert.Equal(t, service.Options.URL, clone.Options.URL)
	assert.Equal(t, service.Options.Authenticator, clone.Options.Authenticator)
	assert.Equal(t, service.Options.EnableGzipCompression, clone.Options.EnableGzipCompression)
}

// Test a normal JSON-based response.
func TestRequestGoodResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"name": "wonder woman"}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("POST")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	// Use a cloned service to verify it works ok.
	service = service.Clone()

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, http.StatusCreated, detailedResponse.StatusCode)
	assert.Equal(t, "application/json", detailedResponse.Headers.Get("Content-Type"))

	result, ok := detailedResponse.Result.(*Foo)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)
	assert.NotNil(t, foo)
	assert.Equal(t, "wonder woman", *(result.Name))
}

// Test a normal JSON-based response using a vendor-specific Content-Type
func TestRequestGoodResponseCustomJSONContentType(t *testing.T) {
	customContentType := "application/vnd.sdksquad.custom.semantics+json;charset=UTF8"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", customContentType)
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"name": "wonder woman"}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("POST")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	// Use a cloned service to verify it works ok.
	service = service.Clone()

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, http.StatusCreated, detailedResponse.StatusCode)
	assert.Equal(t, customContentType, detailedResponse.Headers.Get("Content-Type"))

	result, ok := detailedResponse.Result.(*Foo)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)
	assert.NotNil(t, foo)
	assert.Equal(t, "wonder woman", *(result.Name))
}

// Test a JSON-based response that should be returned as a stream (io.ReadCloser).
func TestRequestGoodResponseJSONStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"name": "wonder woman"}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("POST")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	// Use a cloned service to verify it works ok.
	service = service.Clone()

	detailedResponse, err := service.Request(req, new(io.ReadCloser))
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, http.StatusCreated, detailedResponse.StatusCode)
	assert.Equal(t, "application/json", detailedResponse.Headers.Get("Content-Type"))

	result, ok := detailedResponse.Result.(io.ReadCloser)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)

	// Read the bytes from the response body and decode as JSON to verify.
	responseBytes, err := io.ReadAll(result)
	assert.Nil(t, err)
	assert.NotNil(t, responseBytes)

	// Decode the byte array as JSON.
	var foo *Foo
	err = json.NewDecoder(bytes.NewReader(responseBytes)).Decode(&foo)
	assert.Nil(t, err)
	assert.NotNil(t, foo)
	assert.Equal(t, "wonder woman", *(foo.Name))
}

// Verify that extra fields in result are silently ignored.
func TestRequestGoodResponseJSONExtraFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"name": "wonder woman", "age": 42}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options)

	// Use a cloned service to verify it works ok.
	service = service.Clone()

	var foo *Foo
	detailedResponse, _ := service.Request(req, &foo)
	result, ok := detailedResponse.Result.(*Foo)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)
	assert.Equal(t, "wonder woman", *result.Name)
}

// Test a binary response.
func TestRequestGoodResponseStream(t *testing.T) {
	expectedResponse := []byte("This is an octet stream response.")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/octet-stream")
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
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
	actualResponse, err := io.ReadAll(result)
	assert.Nil(t, err)
	assert.NotNil(t, actualResponse)
	assert.Equal(t, expectedResponse, actualResponse)
}

// Test a text response.
func TestRequestGoodResponseText(t *testing.T) {
	expectedResponse := "This is a text response."
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/plain")
		fmt.Fprint(w, expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())
	detailedResponse, err := service.Request(req, new([]byte))
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, "text/plain", detailedResponse.GetHeaders().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, detailedResponse.GetStatusCode())
	assert.NotNil(t, detailedResponse.Result)
	responseBytes, ok := detailedResponse.Result.([]byte)
	assert.Equal(t, true, ok)
	assert.NotNil(t, responseBytes)
	assert.Equal(t, expectedResponse, string(responseBytes))
}

// Test a string response.
func TestRequestGoodResponseString(t *testing.T) {
	expectedBytes := []byte("This is a string response.")
	expectedResponse := string(expectedBytes)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/plain")
		fmt.Fprint(w, expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())

	var responseString *string
	detailedResponse, err := service.Request(req, &responseString)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, "text/plain", detailedResponse.GetHeaders().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, detailedResponse.GetStatusCode())
	assert.NotNil(t, detailedResponse.Result)
	assert.NotNil(t, responseString)
	assert.Equal(t, expectedResponse, *responseString)

	resultField, ok := detailedResponse.Result.(*string)
	assert.Equal(t, true, ok)
	assert.NotNil(t, resultField)
	assert.Equal(t, responseString, resultField)
	assert.Equal(t, *responseString, *resultField)
}

// Test a non-JSON response with no Content-Type set.
func TestRequestGoodResponseNonJSONNoContentType(t *testing.T) {
	expectedResponse := []byte("This is a non-json response.")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "")
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
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
	actualResponse, err := io.ReadAll(result)
	assert.Nil(t, err)
	assert.NotNil(t, actualResponse)
	assert.Equal(t, expectedResponse, actualResponse)
}

// Test a JSON response with no Content-Type set.
func TestRequestGoodResponseByteSliceNoContentType(t *testing.T) {
	expectedResponse := []byte("This is a non-json response.")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "")
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())
	var rawResponse []byte
	detailedResponse, err := service.Request(req, &rawResponse)
	assert.NotNil(t, detailedResponse)
	assert.NotNil(t, rawResponse)
	assert.Nil(t, err)
	assert.Equal(t, "", detailedResponse.GetHeaders().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, detailedResponse.GetStatusCode())
	assert.NotNil(t, detailedResponse.Result)
	assert.Equal(t, expectedResponse, rawResponse)
}

// Test unexpected response content.
func TestRequestUnexpectedResponse(t *testing.T) {
	expectedResponse := []byte("This is an unexpected response.")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "")
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())
	var rawResponse map[string]json.RawMessage
	detailedResponse, err := service.Request(req, &rawResponse)
	assert.NotNil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.NotNil(t, detailedResponse.Result)
	assert.Equal(t, "", detailedResponse.GetHeaders().Get("Content-Type"))
	assert.Equal(t, http.StatusOK, detailedResponse.GetStatusCode())
}

// Test a JSON response that causes a deserialization error.
func TestRequestGoodResponseJSONDeserFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"name": {"unknown_object_id": "abc123"}}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options)

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.NotNil(t, detailedResponse)
	assert.NotNil(t, err)
	assert.NotNil(t, detailedResponse.RawResult)
	assert.Nil(t, detailedResponse.Result)
	assert.Equal(t,
		true,
		strings.HasPrefix(err.Error(), "An error occurred while unmarshalling the response body:"))
	// t.Log("Decode error:\n", err.Error())
}

func TestRequestNoAuthenticatorFailure(t *testing.T) {
	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL("https://myservice.ibm.com/api", "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           "https://myservice.ibm.com/api",
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options)

	// Now force the authenticator to be nil.
	service.Options.Authenticator = nil

	_, err = service.Request(req, nil)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
}

// Test a good response with no response body.
func TestRequestGoodResponseNoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
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

// Test a good response with unexpected response body.
func TestRequestGoodResponseUnexpectedBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "{}")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	detailedResponse, err := service.Request(req, nil)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, 201, detailedResponse.StatusCode)
	assert.Nil(t, detailedResponse.Result)
	assert.Nil(t, detailedResponse.RawResult)
}

// Test request with result that receives a good response with no response body.
func TestRequestWithResultGoodResponseNoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	var rawResult map[string]json.RawMessage
	detailedResponse, err := service.Request(req, &rawResult)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, 201, detailedResponse.StatusCode)
	assert.Equal(t, "", detailedResponse.Headers.Get("Content-Type"))
	assert.Nil(t, rawResult)
	assert.Nil(t, detailedResponse.Result)
	assert.Nil(t, detailedResponse.RawResult)
}

// Test request with result that receives a good response with no response body and JSON content-type.
func TestRequestWithResultGoodResponseNoBodyJSONObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	var rawResult map[string]json.RawMessage
	detailedResponse, err := service.Request(req, &rawResult)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, 201, detailedResponse.StatusCode)
	assert.Nil(t, rawResult)
	assert.Nil(t, detailedResponse.Result)
	assert.Nil(t, detailedResponse.RawResult)
}

// Test request with result that receives a good response with no response body and JSON content-type.
func TestRequestWithResultGoodResponseNoBodyJSONArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	var rawSlice []json.RawMessage
	detailedResponse, err := service.Request(req, &rawSlice)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, 201, detailedResponse.StatusCode)
	assert.Nil(t, rawSlice)
	assert.Nil(t, detailedResponse.Result)
	assert.Nil(t, detailedResponse.RawResult)
}

// Test request with result that receives a good response with no response body and JSON content-type.
func TestRequestWithResultGoodResponseNoBodyString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/plain")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	var result *string
	detailedResponse, err := service.Request(req, &result)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, 201, detailedResponse.StatusCode)
	assert.Nil(t, result)
	assert.Nil(t, detailedResponse.Result)
	assert.Nil(t, detailedResponse.RawResult)
}

// Test request with result that receives a good response with an empty object body.
func TestRequestWithResultGoodResponseEmptyObjectBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "{}")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	var rawResult map[string]json.RawMessage
	detailedResponse, err := service.Request(req, &rawResult)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, 201, detailedResponse.StatusCode)
	assert.NotNil(t, rawResult)
	assert.NotNil(t, detailedResponse.Result)
	assert.Nil(t, detailedResponse.RawResult)
}

// Test request with result that receives a good response with an empty array body and JSON content-type.
func TestRequestGoodResponseEmptyArrayBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "[]")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	var rawSlice []json.RawMessage
	detailedResponse, err := service.Request(req, &rawSlice)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, 201, detailedResponse.StatusCode)
	assert.NotNil(t, rawSlice)
	assert.NotNil(t, detailedResponse.Result)
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
var nonJSONErrorResponse string = `This is a non-JSON error response body.`

// Test an error response with a JSON response body.
func TestRequestErrorResponseJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, jsonErrorResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options)

	var foo *Foo
	response, err := service.Request(req, &foo)
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
func TestRequestErrorResponseJSONDeserError(t *testing.T) {
	var expectedResponse = []byte(`"{"this is a malformed": "json object".......`)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options)

	var foo *Foo
	response, err := service.Request(req, &foo)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.RawResult)
	assert.Equal(t, expectedResponse, response.RawResult)
	assert.Equal(t, http.StatusText(http.StatusForbidden), err.Error())
}

// Test error response with a non-JSON response body.
func TestRequestErrorResponseNotJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, nonJSONErrorResponse)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options)

	var foo *Foo
	response, err := service.Request(req, &foo)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	assert.Nil(t, response.Result)
	assert.NotNil(t, response.RawResult)
	s := string(response.RawResult)
	assert.Equal(t, nonJSONErrorResponse, s)
	assert.Equal(t, http.StatusText(http.StatusBadRequest), err.Error())
}

// Test an error response with no response body.
func TestRequestErrorResponseNoBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
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
	service, _ := NewBaseService(&ServiceOptions{Authenticator: authenticator})
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
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator, _ := NewBasicAuthenticator("username", "password")
	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, _ := NewBaseService(options)

	var foo *Foo
	_, _ = service.Request(req, &foo)
}

func TestRequestForProvidedUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"name": "wonder woman"}`)
		assert.Contains(t, r.Header.Get("User-Agent"), "provided user agent")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator := &NoAuthAuthenticator{}
	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, _ := NewBaseService(options)
	headers := http.Header{}
	headers.Add("User-Agent", "provided user agent")
	service.SetDefaultHeaders(headers)

	var foo *Foo
	_, _ = service.Request(req, &foo)
}

func TestRequestHostHeaderDefault(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprint(w, `{"name": "wonder woman"}`)
		assert.Equal(t, "server1.cloud.ibm.com", r.Host)
	}))
	defer server.Close()

	authenticator := &NoAuthAuthenticator{}
	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, _ := NewBaseService(options)
	headers := http.Header{}
	headers.Add("Host", "server1.cloud.ibm.com")
	service.SetDefaultHeaders(headers)

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	var foo *Foo
	_, _ = service.Request(req, &foo)
}

func TestIncorrectURL(t *testing.T) {
	authenticator, _ := NewNoAuthAuthenticator()
	options := &ServiceOptions{
		URL:           "{xxx}",
		Authenticator: authenticator,
	}
	_, serviceErr := NewBaseService(options)
	expectedError := fmt.Errorf(ERRORMSG_PROP_INVALID, "URL")
	assert.Equal(t, expectedError.Error(), serviceErr.Error())
}

func TestDisableSSLVerification(t *testing.T) {
	options := &ServiceOptions{
		URL:           "test.com",
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options)
	assert.False(t, service.IsSSLDisabled())
	service.DisableSSLVerification()
	assert.True(t, service.IsSSLDisabled())

	// Try another test while setting the service client to nil
	service, _ = NewBaseService(options)
	service.Client = nil
	assert.False(t, service.IsSSLDisabled())
	service.DisableSSLVerification()
	assert.NotNil(t, service.Client)
	assert.True(t, service.IsSSLDisabled())
}

func TestDisableSSLVerificationWithRetries(t *testing.T) {
	options := &ServiceOptions{
		URL:           "test.com",
		Authenticator: &NoAuthAuthenticator{},
	}
	service, _ := NewBaseService(options)

	// Verify that we can first enable retries, then disable SSL
	assert.False(t, service.IsSSLDisabled())
	service.EnableRetries(3, 30*time.Second)
	assert.False(t, service.IsSSLDisabled())
	service.DisableSSLVerification()
	assert.True(t, service.IsSSLDisabled())
	service.DisableRetries()
	assert.True(t, service.IsSSLDisabled())

	// Verify that we can first disable SSL, then enable retries, etc.
	service, _ = NewBaseService(options)
	assert.False(t, service.IsSSLDisabled())
	service.DisableSSLVerification()
	assert.True(t, service.IsSSLDisabled())
	service.EnableRetries(0, 0)
	assert.True(t, service.IsSSLDisabled())
	service.DisableRetries()
	assert.True(t, service.IsSSLDisabled())

	// Verify that we can first enable retries, then disable SSL, etc.
	service, _ = NewBaseService(options)
	assert.False(t, service.IsSSLDisabled())
	assert.False(t, isRetryableClient(service.Client))

	service.EnableRetries(0, 0)
	assert.False(t, service.IsSSLDisabled())
	assert.True(t, isRetryableClient(service.Client))

	service.DisableSSLVerification()
	assert.True(t, service.IsSSLDisabled())
	assert.True(t, service.IsSSLDisabled())

	service.DisableRetries()
	assert.True(t, service.IsSSLDisabled())
	assert.False(t, isRetryableClient(service.Client))
}

func TestRequestBasicAuth1(t *testing.T) {
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

	service, _ := NewBaseService(options)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	_, err = service.Request(req, nil)
	assert.Nil(t, err)
}

func TestRequestBasicAuth2(t *testing.T) {
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
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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

	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	_, err = service.Request(req, nil)
	assert.Nil(t, err)

	service.Options.Authenticator = &BasicAuthenticator{
		Username: "mookie",
		Password: "betts",
	}

	_, err = service.Request(req, nil)
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

	service, err := NewBaseService(options)
	assert.NotNil(t, err)
	assert.Nil(t, service)
}

func TestRequestNoAuth1(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "", r.Header.Get("Authorization"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
	assert.Nil(t, err)
	builder.AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
	}

	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())

	_, err = service.Request(req, nil)
	assert.Nil(t, err)
}

func TestRequestNoAuth2(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "", r.Header.Get("Authorization"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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

	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.Options.Authenticator.AuthenticationType())

	service.Options.Authenticator = &NoAuthAuthenticator{}
	assert.Nil(t, err)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.Options.Authenticator.AuthenticationType())

	_, err = service.Request(req, nil)
	assert.Nil(t, err)
}

func TestRequestIAMAuth(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "grant_type") {
			assert.Equal(t, true, firstCall)
			firstCall = false
			expiration := GetCurrentTime() + 3600
			fmt.Fprintf(w, `{
				"access_token": "captain marvel",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			assert.Equal(t, "", r.Header.Get("Authorization"))
		} else {
			assert.Equal(t, "Bearer captain marvel", r.Header.Get("Authorization"))
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service.Options.Authenticator.AuthenticationType())

	_, err = service.Request(req, nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	assert.Nil(t, err)

	// Subsequent request should not request new access token
	_, err = service.Request(req, nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}
	assert.Nil(t, err)
}

func TestRequestIAMFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.NotNil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.NotNil(t, detailedResponse.GetHeaders())
	assert.NotNil(t, detailedResponse.GetRawResult())
	statusCode := detailedResponse.GetStatusCode()
	assert.Equal(t, http.StatusForbidden, statusCode)
	assert.Contains(t, err.Error(), "Sorry you are forbidden")
}

func TestRequestIAMFailureRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "20")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("Sorry rate limit has been exceeded"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.NotNil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.NotNil(t, detailedResponse.GetRawResult())
	statusCode := detailedResponse.GetStatusCode()
	headers := detailedResponse.GetHeaders()
	assert.NotNil(t, headers)
	assert.Equal(t, http.StatusTooManyRequests, statusCode)
	assert.Contains(t, headers, "Retry-After")
	assert.Contains(t, err.Error(), "Sorry rate limit has been exceeded")
}

func TestRequestIAMWithIdSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := io.ReadAll(r.Body)
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
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	_, err = service.Request(req, nil)
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
		})
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
		})
	assert.NotNil(t, err)
}

func TestRequestIAMNoApiKey(t *testing.T) {
	_, err := NewBaseService(
		&ServiceOptions{
			URL: "don't care",
			Authenticator: &IamAuthenticator{
				URL:          "don't care",
				ClientId:     "foo",
				ClientSecret: "bar",
			},
		})
	assert.NotNil(t, err)
}

const (
	// #nosec
	cp4dString = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6IlVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhY2Nlc3NfY2F0YWxvZyIsImNhbl9wcm92aXNpb24iLCJzaWduX2luX29ubHkiXSwiZ3JvdXBzIjpbMTAwMDBdLCJzdWIiOiJ0ZXN0dXNlciIsImlzcyI6IktOT1hTU08iLCJhdWQiOiJEU1giLCJ1aWQiOiIxMDAwMzMxMDAzIiwiYXV0aGVudGljYXRvciI6ImRlZmF1bHQiLCJpYXQiOjE2MTA0ODg2NzAsImV4cCI6MTYxMDUzMTgzNH0.K5Mqsv3E9MMXotuhUWbAcTUe41thzSaiFOolnvNxIVPwApSJr_VappL8GTR6BgwPz5gB4MX9w8mVsh0vX8g5naRHWryKNxloHiWiOzCpI982EACkb7Lvdpo5vq_wOANM4OW5Q7cyWXMrqQMz1wF-4-1EyYHBbAKWWGmSQZ6iW7wgMxoeP027vGTD96IVFhgOrvX1hEBDMZ0S9gfKU0bthUMEDKoWONcFuWlHQChhh7agjP2RS4d3Rcjx2oHtx_zuH5bEXxn9g4Dj2v9Bkn6aOFQivSGFUlaus_6opZ6x5aCPi6SXnO_xOY_f2XKU-DUg-yN5BeX7fXu35JQTGFcgwQ"
)

func TestRequestCP4DAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.HasSuffix(r.URL.String(), "/v1/authorize") {
			fmt.Fprintf(w, `{"token":"%s","_messageCode_":"200","message":"success"}`, cp4dString)
		} else {
			expectedAuthHeader := fmt.Sprintf("Bearer %s", cp4dString)
			assert.Equal(t, expectedAuthHeader, r.Header.Get("Authorization"))
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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

	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)

	_, err = service.Request(req, nil)
	assert.Nil(t, err)
}

func TestRequestCP4DFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.NotNil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.NotNil(t, detailedResponse.GetHeaders())
	assert.NotNil(t, detailedResponse.GetRawResult())
	statusCode := detailedResponse.GetStatusCode()
	assert.Equal(t, http.StatusForbidden, statusCode)
	assert.Contains(t, err.Error(), "Sorry you are forbidden")
}

func TestRequestCp4dFailureRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "20")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("Sorry rate limit has been exceeded"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server.URL, "", nil)
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
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Options.Authenticator)

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.NotNil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.NotNil(t, detailedResponse.GetRawResult())
	statusCode := detailedResponse.GetStatusCode()
	headers := detailedResponse.GetHeaders()
	assert.NotNil(t, headers)
	assert.Equal(t, http.StatusTooManyRequests, statusCode)
	assert.Contains(t, headers, "Retry-After")
	assert.Contains(t, err.Error(), "Sorry rate limit has been exceeded")
}

// Test for the deprecated SetURL method.
func TestSetURL(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &IamAuthenticator{
				ApiKey: "xxxxx",
			},
		})
	assert.Nil(t, err)
	assert.NotNil(t, service)

	err = service.SetURL("{bad url}")
	assert.NotNil(t, err)
}

func TestSetServiceURL(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
		})
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

func TestSetUserAgent(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
		})
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotEmpty(t, service.UserAgent)

	service.SetUserAgent("")
	assert.NotEmpty(t, service.UserAgent)

	service.SetUserAgent("my-user-agent")
	assert.Equal(t, "my-user-agent", service.UserAgent)
}

func getRetriesConfig(service *BaseService) (int, time.Duration) {
	if isRetryableClient(service.Client) {
		tr := service.Client.Transport.(*retryablehttp.RoundTripper)
		return tr.Client.RetryMax, tr.Client.RetryWaitMax
	}

	return -1, -1 * time.Second
}

func TestEnableRetries(t *testing.T) {
	options := &ServiceOptions{
		Authenticator: &NoAuthAuthenticator{},
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Client)

	// Save the existing client for comparisons below.
	client := service.Client

	// Enable retries, then verify the config
	// and make sure our client survived.
	service.EnableRetries(5, 30*time.Second)
	assert.True(t, isRetryableClient(service.Client))
	maxRetries, maxInterval := getRetriesConfig(service)
	assert.Equal(t, 5, maxRetries)
	assert.Equal(t, 30*time.Second, maxInterval)
	assert.Equal(t, client, service.GetHTTPClient())

	// Enable retries with a different config and verify.
	service.EnableRetries(6, 60*time.Second)
	assert.True(t, isRetryableClient(service.Client))
	maxRetries, maxInterval = getRetriesConfig(service)
	assert.Equal(t, 6, maxRetries)
	assert.Equal(t, 60*time.Second, maxInterval)
	assert.Equal(t, client, service.GetHTTPClient())

	// Disable retries and make sure the original client is still there.
	service.DisableRetries()
	assert.False(t, isRetryableClient(service.Client))
	assert.Equal(t, client, service.GetHTTPClient())
}

func TestClientWithRetries(t *testing.T) {
	options := &ServiceOptions{
		Authenticator: &NoAuthAuthenticator{},
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Client)

	// Create a customized client and set it on the service
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true,
			},
		},
	}

	service.SetHTTPClient(client)
	actualClient := service.GetHTTPClient()
	assert.Equal(t, client, actualClient)
	assert.Equal(t, *client, *actualClient)
	assert.Equal(t, client, service.Client)

	// Next, enable retries and make sure the client survived.
	service.EnableRetries(4, 90*time.Second)
	assert.True(t, isRetryableClient(service.Client))
	actualClient = service.GetHTTPClient()
	assert.Equal(t, client, actualClient)
	assert.Equal(t, *client, *actualClient)

	// Finally, disable retries and make sure
	// we're left with the same client instance.
	service.DisableRetries()
	assert.False(t, isRetryableClient(service.Client))
	actualClient = service.GetHTTPClient()
	assert.Equal(t, client, actualClient)
	assert.Equal(t, *client, *actualClient)
	assert.Equal(t, client, service.Client)

	// Create a new service and perform the steps in a different order.
	service, err = NewBaseService(options)
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Client)
	assert.NotEqual(t, client, service.Client)

	// Enable retries.
	service.EnableRetries(0, 0)
	assert.True(t, isRetryableClient(service.Client))

	// Next, set our customized client on the service
	service.SetHTTPClient(client)
	assert.True(t, isRetryableClient(service.Client))
	actualClient = service.GetHTTPClient()
	assert.Equal(t, client, actualClient)
	assert.Equal(t, *client, *actualClient)
}

func TestSetEnableGzipCompression(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
		})
	assert.Nil(t, err)
	assert.NotNil(t, service)

	assert.False(t, service.GetEnableGzipCompression())
	assert.False(t, service.Options.EnableGzipCompression)

	service.SetEnableGzipCompression(true)
	assert.True(t, service.GetEnableGzipCompression())
}

func TestExtConfigFromCredentialFile(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

	service, _ := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	err := service.ConfigureService("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service1/api", service.Options.URL)
	assert.True(t, service.IsSSLDisabled())
	assert.True(t, service.GetEnableGzipCompression())
	assert.Nil(t, getRetryableHTTPClient(service.Client))

	service, _ = NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	err = service.ConfigureService("service2")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service2/api", service.Options.URL)
	assert.False(t, service.IsSSLDisabled())
	assert.False(t, service.GetEnableGzipCompression())

	// Verify retryable client enabled with default config.
	actualClient := getRetryableHTTPClient(service.Client)
	assert.NotNil(t, actualClient)
	expectedClient := NewRetryableHTTPClient()
	assert.Equal(t, expectedClient.RetryMax, actualClient.RetryMax)
	assert.Equal(t, expectedClient.RetryWaitMax, actualClient.RetryWaitMax)

	service, _ = NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	err = service.ConfigureService("service3")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service3/api", service.Options.URL)
	assert.False(t, service.IsSSLDisabled())
	assert.False(t, service.GetEnableGzipCompression())
	assert.Nil(t, getRetryableHTTPClient(service.Client))

	service, _ = NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	err = service.ConfigureService("service4")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service4/api", service.Options.URL)
	assert.False(t, service.IsSSLDisabled())
	assert.False(t, service.GetEnableGzipCompression())

	// Verify retryable client with specified config
	actualClient = getRetryableHTTPClient(service.Client)
	assert.NotNil(t, actualClient)
	assert.Equal(t, int(5), actualClient.RetryMax)
	assert.Equal(t, time.Duration(10)*time.Second, actualClient.RetryWaitMax)

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestExtConfigError(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

	service, _ := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	err := service.ConfigureService("error4")
	assert.NotNil(t, err)

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestExtConfigFromEnvironment(t *testing.T) {
	setTestEnvironment()

	service, _ := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	err := service.ConfigureService("service3")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service3/api", service.Options.URL)
	assert.False(t, service.IsSSLDisabled())
	assert.False(t, service.GetEnableGzipCompression())

	clearTestEnvironment()
}

func TestExtConfigFromVCAP(t *testing.T) {
	setTestVCAP(t)

	service, _ := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	err := service.ConfigureService("service2")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service2/api", service.Options.URL)
	assert.False(t, service.IsSSLDisabled())

	clearTestVCAP()
}

func TestConfigureServiceFromCredFile(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "bad url", service.Options.URL)
	assert.False(t, service.IsSSLDisabled())

	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

	err = service.ConfigureService("service5")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service5/api", service.Options.URL)
	assert.True(t, service.IsSSLDisabled())
	assert.False(t, service.GetEnableGzipCompression())

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestConfigureServiceFromVCAP(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "bad url", service.Options.URL)

	setTestVCAP(t)
	err = service.ConfigureService("service3")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service3/api", service.Options.URL)
	assert.False(t, service.IsSSLDisabled())

	clearTestVCAP()
}

func TestConfigureServiceFromEnv(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
			URL:           "bad url",
		})
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "bad url", service.Options.URL)
	assert.False(t, service.IsSSLDisabled())

	setTestEnvironment()
	err = service.ConfigureService("service_1")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Equal(t, "https://service1/api", service.Options.URL)
	assert.True(t, service.IsSSLDisabled())

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
		})
	assert.Nil(t, err)
	err = service.ConfigureService("")
	assert.NotNil(t, err)
	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestAuthNotConfigured(t *testing.T) {
	service, err := NewBaseService(&ServiceOptions{})
	assert.NotNil(t, err)
	assert.Nil(t, service)
}

func TestAuthInterfaceNilValue(t *testing.T) {
	var authenticator *NoAuthAuthenticator = nil

	options := &ServiceOptions{
		Authenticator: authenticator,
	}

	service, err := NewBaseService(options)
	assert.Nil(t, service)
	assert.NotNil(t, err)
	assert.Equal(t, ERRORMSG_NO_AUTHENTICATOR, err.Error())
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

	testGetErrorMessage(t, http.StatusInternalServerError,
		`{"error":{"statusCode":500,"message":"Internal Server Error"}}`,
		"Internal Server Error")

	testGetErrorMessage(t, http.StatusInternalServerError,
		`{"message":{"statusCode":500,"message":"Internal Server Error"}}`,
		"Internal Server Error")

	testGetErrorMessage(t, http.StatusInternalServerError,
		`{"errorMessage":{"statusCode":500,"message":"Internal Server Error"}}`,
		"Internal Server Error")
}

func getTLSVersion(service *BaseService) int {
	var tlsVersion int = -1
	client := service.GetHTTPClient()
	if client != nil {
		tr := client.Transport.(*http.Transport)
		tlsVersion = int(tr.TLSClientConfig.MinVersion)
	}
	return tlsVersion
}

func TestMinSSLVersion(t *testing.T) {
	service, err := NewBaseService(
		&ServiceOptions{
			Authenticator: &NoAuthAuthenticator{},
		})
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.Client)

	// Check the default config.
	assert.Equal(t, getTLSVersion(service), tls.VersionTLS12)

	// Set a insecureClient with different value.
	insecureClient := &http.Client{}
	insecureClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS10,
		},
	}
	service.SetHTTPClient(insecureClient)
	assert.Equal(t, getTLSVersion(service), tls.VersionTLS12)

	// Check retryable client config.
	service.EnableRetries(3, 30*time.Second)
	assert.Equal(t, getTLSVersion(service), tls.VersionTLS12)
}
