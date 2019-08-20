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
	service, _ := NewBaseService(options, "watson", "watson")
	detailedResponse, _ := service.Request(req, new(Foo))
	assert.Equal(t, "wonder woman", *detailedResponse.Result.(*Foo).Name)
}

// Verify that extra fields in result are silently ignored
func TestRequestResponseJSONWithExtraFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintf(w, `{"name": "wonder woman", "age": 42}`)
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
	result, ok := detailedResponse.Result.(*Foo)
	assert.Equal(t, ok, true)
	assert.NotNil(t, result)
	assert.Equal(t, "wonder woman", *result.Name)
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

func TestIncorrectUsernameAndPassword(t *testing.T) {
	options := &ServiceOptions{
		URL:      "xxx",
		Username: "yyy",
		Password: "zzz",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	err := service.SetUsernameAndPassword("{xxx}", "yyy")
	assert.Equal(t, "The username shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your username", err.Error())

	err = service.SetUsernameAndPassword("xxx", "{yyy}")
	assert.Equal(t, "The password shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your password", err.Error())

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

	options2 := &ServiceOptions{
		ICP4DAccessToken: "icp4d token",
	}
	service2, _ := NewBaseService(options2, "watson", "watson")
	assert.Nil(t, service2.Client.Transport)
	service2.DisableSSLVerification()
	assert.NotNil(t, service2.Client.Transport)
	assert.NotNil(t, service2.ICP4DTokenManager.client.Transport)
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

func TestIAMAuthenticationFail(t *testing.T) {
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

	_, err := service.Request(req, new(Foo))
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

func TestICP4DAuthenticationSuccess(t *testing.T) {
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
		AuthenticationType: "icp4d",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	service.Request(req, new(Foo))
}

func TestICP4DAuthenticationFail(t *testing.T) {
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
		Username:           "bogus",
		Password:           "bogus",
		ICP4DURL:           server.URL,
		AuthenticationType: "icp4d",
	}
	service, _ := NewBaseService(options, "watson", "watson")

	_, err := service.Request(req, new(Foo))
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

func TestIAMBasicAuthDefault(t *testing.T) {
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
			assert.Equal(t, ok, true)
			assert.Equal(t, username, "bx")
			assert.Equal(t, password, "bx")
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

func TestIAMBasicAuthNonDefault(t *testing.T) {
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
			assert.Equal(t, ok, true)
			assert.Equal(t, username, "foo")
			assert.Equal(t, password, "bar")
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
		URL:             server.URL,
		IAMApiKey:       "xxxxx",
		IAMURL:          server.URL,
		IAMClientId:     "foo",
		IAMClientSecret: "bar",
	}
	service, _ := NewBaseService(options, "watson", "watson")

	service.Request(req, new(Foo))
}

func TestIAMBasicAuthClientIdOnly(t *testing.T) {
	options := &ServiceOptions{
		URL:         "don't care",
		IAMApiKey:   "xxxxx",
		IAMClientId: "foo",
	}
	_, err := NewBaseService(options, "watson", "watson")
	assert.NotEqual(t, err, nil)
}

func TestIAMBasicAuthClientSecretOnly(t *testing.T) {
	options := &ServiceOptions{
		URL:             "don't care",
		IAMApiKey:       "xxxxx",
		IAMClientSecret: "foo",
	}
	_, err := NewBaseService(options, "watson", "watson")
	assert.NotEqual(t, err, nil)
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
	assert.NotNil(t, service3.IAMTokenManager)
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
	assert.Nil(t, service2.IAMTokenManager)
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

func TestSetICP4DAccessToken(t *testing.T) {
	options := &ServiceOptions{
		Username: "bogus",
		Password: "bogus",
	}
	service, _ := NewBaseService(options, "watson", "watson")
	assert.Nil(t, service.ICP4DTokenManager)
	service.SetICP4DAccessToken("some icp4d token")
	assert.NotNil(t, service.ICP4DTokenManager)
	assert.Equal(t, "some icp4d token", service.ICP4DTokenManager.userAccessToken)
	service.SetICP4DAccessToken("resetting some icp4d token")
	assert.Equal(t, "resetting some icp4d token", service.ICP4DTokenManager.userAccessToken)
}

func TestICP4DAuthentication(t *testing.T) {
	options := &ServiceOptions{
		AuthenticationType: "ICP4D",
		Username:           "bogus",
		Password:           "bogus",
	}
	_, err := NewBaseService(options, "watson", "watson")
	assert.NotNil(t, err)
	assert.Equal(t, "The ICP4DURL is mandatory for ICP4D", err.Error())

	options.ICP4DURL = "test.com"
	service, err := NewBaseService(options, "watson", "watson")
	assert.Nil(t, err)
	assert.NotNil(t, service)
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

	msg5 := []byte(`{"errorMessage":"error5"}`)
	response5 := http.Response{
		Body: ioutil.NopCloser(bytes.NewBuffer(msg5)),
	}
	message5 := getErrorMessage(&response5)
	assert.Equal(t, message5, "error5")
}
