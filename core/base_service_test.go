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
	"bytes"
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
		URL:           server.URL,
		Authenticator: &NoAuthAuthenticator{},
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

	authenticator, err := NewBasicAuthenticator("xxx", "yyy")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	_, err = service.Request(req, new(Foo))
	assert.NotNil(t, err)
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
		fmt.Fprintf(w, `{"name": "wonder woman"}`)
		assert.Contains(t, r.Header.Get("User-Agent"), "ibm-go-sdk-core")
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	authenticator, _ := NewBasicAuthenticator("username", "password")
	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
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

	authenticator := &NoAuthAuthenticator{}
	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, _ := NewBaseService(options, "watson", "watson")
	headers := http.Header{}
	headers.Add("User-Agent", "provided user agent")
	service.SetDefaultHeaders(headers)
	service.Request(req, new(Foo))
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

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
	req, _ := builder.Build()

	_, err := service.Request(req, new(Foo))
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

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
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

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
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

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
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
			fmt.Fprintf(w, `{
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

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
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
		w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
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
			fmt.Fprintf(w, `{
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

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
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
			assert.Equal(t, "Bearer eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI", r.Header.Get("Authorization"))
		}
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
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
		w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	builder := NewRequestBuilder("GET").
		ConstructHTTPURL(server.URL, nil, nil).
		AddQuery("Version", "2018-22-09")
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

func TestAuthNotConfigured(t *testing.T) {
	service, err := NewBaseService(&ServiceOptions{}, "noauth_service", "noauth_service")
	assert.NotNil(t, err)
	assert.Nil(t, service)
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
