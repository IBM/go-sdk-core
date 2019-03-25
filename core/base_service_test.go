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
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
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
	service, _ := NewBaseService(options, "watson", "watson")
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
	assert.Equal(t, "The username shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your username", serviceErr.Error())
}

func TestIncorrectURL(t *testing.T) {
	options := &ServiceOptions{
		URL:      "{xxx}",
		Username: "yyy",
		Password: "zzz",
	}
	_, serviceErr := NewBaseService(options, "watson", "watson")
	assert.Equal(t, "The URL shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your URL", serviceErr.Error())
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
}

func TestAuthenticationUserNamePassword(t *testing.T) {
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

	service.Request(req, new(Foo))
}

func TestIAMAuthentication(t *testing.T) {
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

	service.Request(req, new(Foo))
}

func TestLoadingFromCredentialFile(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/ibm-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)
	options := &ServiceOptions{}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.Equal(t, service.Options.IAMApiKey, "5678efgh")
	os.Unsetenv("IBM_CREDENTIALS_FILE")

	options2 := &ServiceOptions{IAMApiKey: "xxx"}
	service2, _ := NewBaseService(options2, "watson", "watson")
	assert.Equal(t, service2.Options.IAMApiKey, "xxx")
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
	assert.Equal(t, "icp-xxx", service2.Options.Password)

	options3 := &ServiceOptions{
		Username: "apikey",
		Password: "nobasicauth",
	}
	service3, _ := NewBaseService(options3, "watson", "watson")
	assert.NotNil(t, service3.TokenManager)
	assert.Equal(t, "nobasicauth", service3.Options.IAMApiKey)

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
	assert.Nil(t, service2.TokenManager)
	service2.SetIAMAccessToken("new token")
	assert.Equal(t, "new token", service2.Options.IAMAccessToken)
}

func TestSetIAMAPIKey(t *testing.T) {
	options := &ServiceOptions{
		IAMApiKey: "test",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	err := service.SetIAMAPIKey("{bad}")
	assert.NotNil(t, err)

	service, _ = NewBaseService(options, "watson", "watson")
	service.SetIAMAPIKey("good")
	assert.Equal(t, "good", service.Options.IAMApiKey)
}

func TestSetURL(t *testing.T) {
	options := &ServiceOptions{
		IAMApiKey: "test",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	err := service.SetURL("{bad url}")
	assert.NotNil(t, err)
}

func TestLoadingFromVCAPServices(t *testing.T) {
	vcapServices := `{
		"watson": [{
			"credentials": {
				"url": "https://gateway.watsonplatform.net/compare-comply/api",
				"username": "bogus username",
				"password": "bogus password",
				"apikey": "bogus apikey"
			}
		}]
	}`
	os.Setenv("VCAP_SERVICES", vcapServices)
	service, _ := NewBaseService(&ServiceOptions{}, "watson", "watson")
	assert.Equal(t, "bogus apikey", service.Options.IAMApiKey)
	os.Unsetenv("VCAP_SERVICES")

	vcapServices2 := `{
		"watson": [{
			"credentials": {
				"url": "https://gateway.watsonplatform.net/compare-comply/api",
				"username": "bogus username",
				"password": "bogus password"
			}
		}]
	}`
	os.Setenv("VCAP_SERVICES", vcapServices2)
	service2, _ := NewBaseService(&ServiceOptions{}, "watson", "watson")
	assert.Equal(t, "bogus username", service2.Options.Username)
	os.Unsetenv("VCAP_SERVICES")
}

func TestNoAuth(t *testing.T) {
	_, err := NewBaseService(&ServiceOptions{}, "watson", "watson")
	assert.NotNil(t, err)
}

func TestErrorMessage(t *testing.T) {
	msg1 := []byte(`{"error":"error1"}`)
	response1 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg1)),
	}
	message1 := getErrorMessage(&response1)
	assert.Equal(t, message1, "error1")

	msg2 := []byte(`{"message":"error2"}`)
	response2 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg2)),
	}
	message2 := getErrorMessage(&response2)
	assert.Equal(t, message2, "error2")

	msg3 := []byte(`{"errors":[{"message":"error3"}]}`)
	response3 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg3)),
	}
	message3 := getErrorMessage(&response3)
	assert.Equal(t, message3, "error3")

	msg4 := []byte(`{"msg":"error4"}`)
	response4 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg4)),
	}
	message4 := getErrorMessage(&response4)
	assert.Equal(t, message4, "Unknown Error")
}
