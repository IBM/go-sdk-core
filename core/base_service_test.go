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
	"bytes"
	"encoding/base64"
	"fmt"
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

func TestRequestResponseAsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintf(w, `{"name": "wonder woman"}`)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:      server.URL,
		Username: "xxx",
		Password: "yyy",
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.authenticator.AuthenticationType())
	detailedResponse, _ := service.Request(req, new(Foo))
	assert.Equal(t, "wonder woman", *detailedResponse.Result.(*Foo).Name)
}

func TestRequestFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:      server.URL,
		Username: "xxx",
		Password: "yyy",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	_, err := service.Request(req, new(Foo))
	assert.NotNil(t, err)
}

func TestClient(t *testing.T) {
	mockClient := http.Client{}
	service, _ := NewBaseService(&ServiceOptions{IAMApiKey: "test"}, "watson", "watson")
	service.SetHTTPClient(&mockClient)
	assert.ObjectsAreEqual(mockClient, service.Client)
}

func TestRequestForDefaultUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintf(w, `{"name": "wonder woman"}`)
		assert.Contains(t, r.Header.Get("User-Agent"), "ibm-go-sdk-core")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:      server.URL,
		Username: "xxx",
		Password: "yyy",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	service.Request(req, new(Foo))
}

func TestRequestForProvidedUserAgent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintf(w, `{"name": "wonder woman"}`)
		assert.Contains(t, r.Header.Get("User-Agent"), "provided user agent")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:      server.URL,
		Username: "xxx",
		Password: "yyy",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	headers := http.Header{}
	headers.Add("User-Agent", "provided user agent")
	service.SetDefaultHeaders(headers)
	service.Request(req, new(Foo))
}
func TestIncorrectCreds(t *testing.T) {
	options := &ServiceOptions{
		URL:      "xxx",
		Username: "{yyy}",
		Password: "zzz",
	}
	_, serviceErr := NewBaseService(options, "watson", "watson")
	expectedError := fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "username", "username")
	assert.Equal(t, expectedError.Error(), serviceErr.Error())
}

func TestIncorrectUsernameAndPassword(t *testing.T) {
	options := &ServiceOptions{
		URL: "xxx",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	err := service.SetUsernameAndPassword("{xxx}", "yyy")
	expectedError := fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "username", "username")
	assert.Equal(t, expectedError.Error(), err.Error())

	err = service.SetUsernameAndPassword("xxx", "{yyy}")
	expectedError = fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "password", "password")
	assert.Equal(t, expectedError.Error(), err.Error())

}

func TestIncorrectURL(t *testing.T) {
	options := &ServiceOptions{
		URL:      "{xxx}",
		Username: "yyy",
		Password: "zzz",
	}
	_, serviceErr := NewBaseService(options, "watson", "watson")
	expectedError := fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "URL", "URL")
	assert.Equal(t, expectedError.Error(), serviceErr.Error())
}

func TestDisableSSLverification(t *testing.T) {
	options := &ServiceOptions{
		URL:      "test.com",
		Username: "xxx",
		Password: "yyy",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.Nil(t, service.Client.Transport)
	service.DisableSSLVerification()
	assert.NotNil(t, service.Client.Transport)

	options2 := &ServiceOptions{
		ICP4DAccessToken: "icp4d token",
	}
	service2, _ := NewBaseService(options2, "watson", "watson")
	assert.Nil(t, service2.Client.Transport)
	service2.DisableSSLVerification()
	assert.NotNil(t, service2.Client.Transport)
}

func TestBasicAuthWithAuthConfigField(t *testing.T) {
	encodedBasicAuth := base64.StdEncoding.EncodeToString([]byte("xxx:yyy"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "Basic "+encodedBasicAuth, r.Header["Authorization"][0])
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &BasicAuthConfig{
		Username: "xxx",
		Password: "yyy",
	}

	options := &ServiceOptions{
		URL:        server.URL,
		AuthConfig: authConfig,
	}

	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestBasicAuthWithSetAuthenticator(t *testing.T) {
	encodedBasicAuth := base64.StdEncoding.EncodeToString([]byte("xxx:yyy"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "Basic "+encodedBasicAuth, r.Header["Authorization"][0])
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &BasicAuthConfig{
		Username: "xxx",
		Password: "yyy",
	}

	options := &ServiceOptions{
		URL: server.URL,
	}

	service, _ := NewBaseService(options, "watson", "watson")
	assert.Nil(t, service.authenticator)

	service.SetAuthenticator(authConfig)
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestBasicAuthWithDeprecatedProps(t *testing.T) {
	encodedBasicAuth := base64.StdEncoding.EncodeToString([]byte("xxx:yyy"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "Basic "+encodedBasicAuth, r.Header["Authorization"][0])
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:      server.URL,
		Username: "xxx",
		Password: "yyy",
	}
	service, _ := NewBaseService(options, "watson", "watson")

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestNoauthWithAuthConfigField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "", r.Header.Get("Authorization"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &NoauthConfig{}

	options := &ServiceOptions{
		URL:        server.URL,
		AuthConfig: authConfig,
	}

	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.authenticator.AuthenticationType())

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestNoauthWithSetAuthenticator(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "", r.Header.Get("Authorization"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &NoauthConfig{}

	options := &ServiceOptions{
		URL: server.URL,
	}

	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Nil(t, service.authenticator)

	err = service.SetAuthenticator(authConfig)
	assert.Nil(t, err)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.authenticator.AuthenticationType())

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestNoauthWithDeprecatedProps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, "", r.Header.Get("Authorization"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:                server.URL,
		AuthenticationType: AUTHTYPE_NOAUTH,
	}

	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, service.authenticator.AuthenticationType())

	_, err = service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestIAMWithAuthConfigField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "grant_type") {
			fmt.Fprintf(w, `{
				"access_token": "captain marvel",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": 1524167011,
				"refresh_token": "jy4gl91BQ"
			}`)
		} else {
			assert.Equal(t, "Bearer captain marvel", r.Header["Authorization"][0])
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &IAMConfig{
		URL:    server.URL,
		ApiKey: "xxxxx",
	}

	options := &ServiceOptions{
		URL:        server.URL,
		AuthConfig: authConfig,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestIAMWithSetAuthenticator(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "grant_type") {
			fmt.Fprintf(w, `{
				"access_token": "captain marvel",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": 1524167011,
				"refresh_token": "jy4gl91BQ"
			}`)
		} else {
			assert.Equal(t, "Bearer captain marvel", r.Header["Authorization"][0])
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &IAMConfig{
		URL:    server.URL,
		ApiKey: "xxxxx",
	}

	options := &ServiceOptions{
		URL: server.URL,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.Nil(t, service.authenticator)
	service.SetAuthenticator(authConfig)
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestIAMWithDeprecatedProps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "grant_type") {
			fmt.Fprintf(w, `{
				"access_token": "captain marvel",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": 1524167011,
				"refresh_token": "jy4gl91BQ"
			}`)
		} else {
			assert.Equal(t, "Bearer captain marvel", r.Header["Authorization"][0])
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:       server.URL,
		IAMApiKey: "xxxxx",
		IAMURL:    server.URL,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestIAMFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		URL:       server.URL,
		IAMApiKey: "xxxxx",
		IAMURL:    server.URL,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.NotNil(t, err)
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

func TestICP4DWithAuthConfigField(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.String(), "preauth") {
			fmt.Fprintf(w, `{
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
			assert.Equal(t, "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI", r.Header["Authorization"][0])
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &ICP4DConfig{
		Username: "bogus",
		Password: "bogus",
		URL:      server.URL,
	}

	options := &ServiceOptions{
		URL:        server.URL,
		AuthConfig: authConfig,
	}

	service, _ := NewBaseService(options, "watson", "watson")
	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

// Test ICP4D with SetAuthenticator() call.
func TestICP4DWithSetAuthenticator(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.String(), "preauth") {
			fmt.Fprintf(w, `{
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
			assert.Equal(t, "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI", r.Header["Authorization"][0])
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &ICP4DConfig{
		Username: "bogus",
		Password: "bogus",
		URL:      server.URL,
	}
	options := &ServiceOptions{}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.Nil(t, service.authenticator)

	service.SetAuthenticator(authConfig)
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestICP4DWithDeprecatedProps(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if strings.Contains(r.URL.String(), "preauth") {
			fmt.Fprintf(w, `{
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
			assert.Equal(t, "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI", r.Header["Authorization"][0])
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	options := &ServiceOptions{
		Username:           "bogus",
		Password:           "bogus",
		ICP4DURL:           server.URL,
		AuthenticationType: AUTHTYPE_ICP4D,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestICP4DFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &ICP4DConfig{
		Username: "bogus",
		Password: "bogus",
		URL:      server.URL,
	}
	options := &ServiceOptions{
		URL: server.URL,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.Nil(t, service.authenticator)

	service.SetAuthenticator(authConfig)
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

func TestIAMDefaultIdSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "grant_type") {
			fmt.Fprintf(w, `{
                "access_token": "captain marvel",
                "token_type": "Bearer",
                "expires_in": 3600,
                "expiration": 1524167011,
                "refresh_token": "jy4gl91BQ"
            }`)
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "bx", username)
			assert.Equal(t, "bx", password)
		} else {
			assert.Equal(t, "Bearer captain marvel", r.Header["Authorization"][0])
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &IAMConfig{
		ApiKey: "xxxxx",
		URL:    server.URL,
	}

	options := &ServiceOptions{
		URL:        server.URL,
		AuthConfig: authConfig,
	}

	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestIAMWithIdSecret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		body, _ := ioutil.ReadAll(r.Body)
		if strings.Contains(string(body), "grant_type") {
			fmt.Fprintf(w, `{
                "access_token": "captain marvel",
                "token_type": "Bearer",
                "expires_in": 3600,
                "expiration": 1524167011,
                "refresh_token": "jy4gl91BQ"
            }`)
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "foo", username)
			assert.Equal(t, "bar", password)
		} else {
			assert.Equal(t, "Bearer captain marvel", r.Header["Authorization"][0])
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authConfig := &IAMConfig{
		ApiKey:       "xxxxx",
		URL:          server.URL,
		ClientId:     "foo",
		ClientSecret: "bar",
	}

	options := &ServiceOptions{
		URL:        server.URL,
		AuthConfig: authConfig,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)

	_, err := service.Request(req, new(Foo))
	assert.Nil(t, err)
}

func TestIAMWithClientIdOnly(t *testing.T) {
	authConfig := &IAMConfig{
		ApiKey:   "xxxxx",
		ClientId: "foo",
	}
	options := &ServiceOptions{
		URL:        "don't care",
		AuthConfig: authConfig,
	}
	_, err := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, err)
}

func TestIAMWithClientSecretOnly(t *testing.T) {
	authConfig := &IAMConfig{
		ApiKey:       "xxxxx",
		ClientSecret: "bar",
	}
	options := &ServiceOptions{
		URL:        "don't care",
		AuthConfig: authConfig,
	}
	_, err := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, err)
}

func TestIAMNoApiKey(t *testing.T) {
	authConfig := &IAMConfig{
		URL:          "don't care",
		ClientId:     "foo",
		ClientSecret: "bar",
	}
	options := &ServiceOptions{
		URL:        "don't care",
		AuthConfig: authConfig,
	}
	_, err := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, err)
}

func TestCredFileIAM(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/ibm-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)
	options := &ServiceOptions{}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service.authenticator.AuthenticationType())
}

func TestCredFileICP4D(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/ibm-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)
	options := &ServiceOptions{}
	service, err := NewBaseService(options, "myservice1", "myservice1")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_ICP4D, service.authenticator.AuthenticationType())
	assert.Equal(t, "https://icp4durl", service.Options.URL)
}

func TestCredFileBasicAuth(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/ibm-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)
	options := &ServiceOptions{}
	service, err := NewBaseService(options, "myservice2", "myservice2")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.authenticator.AuthenticationType())
	assert.Equal(t, "https://basicurl", service.Options.URL)
}

func TestICPAuthentication(t *testing.T) {
	options := &ServiceOptions{
		IAMApiKey: "xxx",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.Equal(t, "xxx", service.Options.IAMApiKey)

	options2 := &ServiceOptions{
		IAMApiKey: "icp-xxx",
	}
	service2, _ := NewBaseService(options2, "watson", "watson")
	assert.Equal(t, AUTHTYPE_BASIC, service2.authenticator.AuthenticationType())

	options3 := &ServiceOptions{
		Username: "apikey",
		Password: "nobasicauth",
	}
	service3, _ := NewBaseService(options3, "watson", "watson")
	assert.NotNil(t, service3.authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service3.authenticator.AuthenticationType())

	options4 := &ServiceOptions{
		Username: "apikey",
		Password: "{nobasicauth}",
	}
	_, err4 := NewBaseService(options4, "watson", "watson")
	assert.NotNil(t, err4)

	options5 := &ServiceOptions{
		IAMApiKey: "icp-test}",
	}
	_, err5 := NewBaseService(options5, "watson", "watson")
	assert.NotNil(t, err5)

	options6 := &ServiceOptions{
		IAMApiKey: "{test}",
	}
	_, err6 := NewBaseService(options6, "watson", "watson")
	assert.NotNil(t, err6)
}

func TestSetIAMAccessToken(t *testing.T) {
	options1 := &ServiceOptions{
		IAMApiKey: "apikey",
	}
	service1, _ := NewBaseService(options1, "watson", "watson")
	assert.Equal(t, "apikey", service1.Options.IAMApiKey)
	service1.SetIAMAccessToken("newIAMAccessToken")

	options2 := &ServiceOptions{
		Username: "hello",
		Password: "pwd",
	}
	service2, _ := NewBaseService(options2, "watson", "watson")
	assert.NotNil(t, service2.authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service2.authenticator.AuthenticationType())

	service2.SetIAMAccessToken("new token")
	assert.NotNil(t, service2.authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service2.authenticator.AuthenticationType())
}

func TestSetIAMAPIKey(t *testing.T) {
	options := &ServiceOptions{
		IAMApiKey: "test",
	}
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service.authenticator.AuthenticationType())

	err = service.SetIAMAPIKey("{bad}")
	assert.NotNil(t, err)

	service, err = NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service.authenticator.AuthenticationType())
	service.SetIAMAPIKey("good")
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service.authenticator.AuthenticationType())
	assert.Equal(t, "good", service.Options.IAMApiKey)
}

func TestSetAuthenticator(t *testing.T) {
	// Create a basic auth config.
	basicConfig := &BasicAuthConfig{
		Username: "foo",
		Password: "bar",
	}

	options := &ServiceOptions{
		URL:        "http://localhost/fake/url",
		AuthConfig: basicConfig,

		// Include these temporarily until NewBaseService no longer requires them.
		Username: "blah",
		Password: "blah",
	}
	service, err := NewBaseService(options, "foo", "bar")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.authenticator.AuthenticationType())

	iamConfig := &IAMConfig{
		ApiKey: "my api key",
	}
	err = service.SetAuthenticator(iamConfig)
	assert.Nil(t, err)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service.authenticator.AuthenticationType())
}

func TestSetURL(t *testing.T) {
	options := &ServiceOptions{
		IAMApiKey: "test",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	err := service.SetURL("{bad url}")
	assert.NotNil(t, err)
}

func TestSetICP4DAccessToken(t *testing.T) {
	options := &ServiceOptions{
		Username: "bogus",
		Password: "bogus",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service.authenticator.AuthenticationType())

	service.SetICP4DAccessToken("some icp4d token")
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_ICP4D, service.authenticator.AuthenticationType())

	service.SetICP4DAccessToken("resetting some icp4d token")
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_ICP4D, service.authenticator.AuthenticationType())
}

func TestICP4DAuthentication(t *testing.T) {
	options := &ServiceOptions{
		AuthenticationType: "ICP4D",
		Username:           "bogus",
		Password:           "bogus",
	}
	_, err := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, err)
	assert.Equal(t, ERRORMSG_URL_MISSING, err.Error())

	options.ICP4DURL = "test.com"
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
}

func TestLoadingFromVCAPServices(t *testing.T) {
	vcapServices := `{
		"service1": [{
			"credentials": {
				"url": "https://gateway.watsonplatform.net/compare-comply/api",
				"username": "bogus username",
				"password": "bogus password",
				"apikey": "bogus apikey"
			}
		}]
	}`
	os.Setenv("VCAP_SERVICES", vcapServices)
	service, err := NewBaseService(&ServiceOptions{}, "service1", "service1")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.NotNil(t, service.authenticator)
	assert.Equal(t, AUTHTYPE_IAM, service.authenticator.AuthenticationType())
	os.Unsetenv("VCAP_SERVICES")

	vcapServices2 := `{
		"service2": [{
			"credentials": {
				"url": "https://gateway.watsonplatform.net/compare-comply/api",
				"username": "bogus username",
				"password": "bogus password"
			}
		}]
	}`
	os.Setenv("VCAP_SERVICES", vcapServices2)
	service2, err := NewBaseService(&ServiceOptions{}, "service2", "service2")
	assert.Nil(t, err)
	assert.NotNil(t, service2)
	assert.NotNil(t, service2.authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, service2.authenticator.AuthenticationType())
	os.Unsetenv("VCAP_SERVICES")
}

func TestAuthNotConfigured(t *testing.T) {
	builder := NewRequestBuilder("GET").
		ConstructHTTPURL("foo", nil, nil)
	req, _ := builder.Build()

	service, err := NewBaseService(&ServiceOptions{}, "noauth_service", "noauth_service")
	assert.Nil(t, err)
	assert.NotNil(t, service)
	assert.Nil(t, service.authenticator)

	_, err = service.Request(req, new(Foo))
	assert.NotNil(t, err)
}

func TestErrorMessage(t *testing.T) {
	msg1 := []byte(`{"error":"error1"}`)
	response1 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg1)),
	}
	message1 := getErrorMessage(&response1)
	assert.Equal(t, "error1", message1)

	msg2 := []byte(`{"message":"error2"}`)
	response2 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg2)),
	}
	message2 := getErrorMessage(&response2)
	assert.Equal(t, "error2", message2)

	msg3 := []byte(`{"errors":[{"message":"error3"}]}`)
	response3 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg3)),
	}
	message3 := getErrorMessage(&response3)
	assert.Equal(t, "error3", message3)

	msg4 := []byte(`{"msg":"error4"}`)
	response4 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg4)),
	}
	message4 := getErrorMessage(&response4)
	assert.Equal(t, "Unknown Error", message4)

	msg5 := []byte(`{"errorMessage":"error5"}`)
	response5 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg5)),
	}
	message5 := getErrorMessage(&response5)
	assert.Equal(t, "error5", message5)
}
