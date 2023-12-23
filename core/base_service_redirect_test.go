//go:build all || slow || auth
// +build all slow auth

package core

// (C) Copyright IBM Corp. 2023.
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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: this unit test depends on some bogus hostnames being defined in /etc/hosts.
// Append this to your /etc/hosts file:
//    # for testing
//    127.0.0.1 region1.cloud.ibm.com region2.cloud.ibm.com region1.notcloud.ibm.com region2.notcloud.ibm.com

var (
	operationPath string = "/api/redirector"

	// To enable debug mode while running these tests, set this to LevelDebug.
	redirectTestLogLevel LogLevel = LevelError
)

// Start a mock server that will redirect requests to the second mock server
// located at "redirectServerURL"
func startMockServer1(t *testing.T, redirectServerURL string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf(`server1 received request: %s %s`, req.Method, req.URL.String())

		// Make sure the Authorization header was sent.
		assert.NotEmpty(t, req.Header.Get("Authorization"))

		path := req.URL.Path
		location := redirectServerURL + path

		// Create the response (a 302 redirect).
		w.Header().Add("Location", location)
		w.WriteHeader(http.StatusFound)
		t.Logf(`Sent redirect request to: %s`, location)
	}))
	return server
}

// Start a second mock server to which redirected requests will be sent.
func startMockServer2(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf(`server2 received request: %s %s`, req.Method, req.URL.String())

		// Create the response.
		if req.Header.Get("Authorization") != "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"name":"Jason Bourne"}`)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	return server
}

func startServers(t *testing.T, host1 string, host2 string) (server1 *httptest.Server, server1URL string,
	server2 *httptest.Server, server2URL string) {
	server2 = startMockServer2(t)
	server2URL = strings.ReplaceAll(server2.URL, "127.0.0.1", host2)
	t.Logf(`Server 2 listening on endpoint: %s (%s)`, server2URL, server2.URL)

	server1 = startMockServer1(t, server2URL)
	server1URL = strings.ReplaceAll(server1.URL, "127.0.0.1", host1)
	t.Logf(`Server 1 listening on endpoint: %s (%s)`, server1URL, server1.URL)

	return
}

func testRedirection(t *testing.T, host1 string, host2 string, expectedStatusCode int) {
	GetLogger().SetLogLevel(redirectTestLogLevel)

	// Both servers within trusted domain.
	server1, server1URL, server2, _ := startServers(t, host1, host2)
	defer server1.Close()
	defer server2.Close()

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL(server1URL, operationPath, nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, err := NewBearerTokenAuthenticator("this is not a secret")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server1.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, expectedStatusCode, detailedResponse.StatusCode)
	if expectedStatusCode >= 200 && expectedStatusCode <= 299 {
		assert.Nil(t, err)

		result, ok := detailedResponse.Result.(*Foo)
		assert.Equal(t, true, ok)
		assert.NotNil(t, result)
		assert.NotNil(t, foo)
		assert.Equal(t, "Jason Bourne", *result.Name)
	} else {
		assert.NotNil(t, err)
	}
}

func TestRedirectAuthSuccess1(t *testing.T) {
	testRedirection(t, "region1.cloud.ibm.com", "region2.cloud.ibm.com", http.StatusOK)
}

func TestRedirectAuthSuccess2(t *testing.T) {
	testRedirection(t, "region1.cloud.ibm.com", "region1.cloud.ibm.com", http.StatusOK)
}

func TestRedirectAuthSuccess3(t *testing.T) {
	testRedirection(t, "region1.notcloud.ibm.com", "region1.notcloud.ibm.com", http.StatusOK)
}

func TestRedirectAuthFail1(t *testing.T) {
	testRedirection(t, "region1.notcloud.ibm.com", "region2.cloud.ibm.com", http.StatusUnauthorized)
}

func TestRedirectAuthFail2(t *testing.T) {
	testRedirection(t, "region1.cloud.ibm.com", "region2.notcloud.ibm.com", http.StatusUnauthorized)
}

func TestRedirectAuthFail3(t *testing.T) {
	testRedirection(t, "region1.notcloud.ibm.com", "region2.notcloud.ibm.com", http.StatusUnauthorized)
}
