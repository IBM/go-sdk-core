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
	"net/http"
	"net/http/httptest"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestCp4dConfigErrors(t *testing.T) {
	var err error

	// Missing URL.
	_, err = NewCloudPakForDataAuthenticator("", "mookie", "betts", false, nil)
	assert.NotNil(t, err)

	// Missing Username.
	_, err = NewCloudPakForDataAuthenticator("cp4d-url", "", "betts", false, nil)
	assert.NotNil(t, err)

	// Missing Password.
	_, err = NewCloudPakForDataAuthenticator("cp4d-url", "mookie", "", false, nil)
	assert.NotNil(t, err)
}

func TestCp4dGetTokenSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
		username, password, ok := r.BasicAuth()
		assert.Equal(t, true, ok)
		assert.Equal(t, "mookie", username)
		assert.Equal(t, "betts", password)
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "mookie", "betts", false, nil)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	// Force the first fetch and verify we got the correct access token back
	accessToken, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		accessToken)

	// Force an expiration and verify we get back the second access token.
	authenticator.tokenData = nil
	accessToken, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.tokenData)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		accessToken)
}

func TestCp4dDisableSSL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
	}))
	defer server.Close()

	authenticator := &CloudPakForDataAuthenticator{
		URL:                    server.URL,
		Username:               "mookie",
		Password:               "betts",
		DisableSSLVerification: true,
	}

	token, err := authenticator.getToken()
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.Client)
	assert.NotNil(t, authenticator.Client.Transport)
	transport, ok := authenticator.Client.Transport.(*http.Transport)
	assert.Equal(t, true, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.Equal(t, true, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestCp4dUserHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
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
		username, password, ok := r.BasicAuth()
		assert.Equal(t, true, ok)
		assert.Equal(t, "mookie", username)
		assert.Equal(t, "betts", password)
		assert.Equal(t, "Value1", r.Header.Get("Header1"))
		assert.Equal(t, "Value2", r.Header.Get("Header2"))
	}))
	defer server.Close()

	headers := make(map[string]string)
	headers["Header1"] = "Value1"
	headers["Header2"] = "Value2"

	authenticator := &CloudPakForDataAuthenticator{
		URL:      server.URL,
		Username: "mookie",
		Password: "betts",
		Headers:  headers,
	}

	token, err := authenticator.getToken()
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.Nil(t, err)
}

func TestGetTokenFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	authenticator := &CloudPakForDataAuthenticator{
		URL:      server.URL,
		Username: "john",
		Password: "snow",
	}

	_, err := authenticator.getToken()
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

func TestNewCloudPakForDataAuthenticatorFromMap(t *testing.T) {
	_, err := newCloudPakForDataAuthenticatorFromMap(nil)
	assert.NotNil(t, err)

	var props = map[string]string{
		PROPNAME_AUTH_URL: "cp4d-url",
	}
	_, err = newCloudPakForDataAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_AUTH_URL: "cp4d-url",
		PROPNAME_USERNAME: "mookie",
	}
	_, err = newCloudPakForDataAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_AUTH_URL: "cp4d-url",
		PROPNAME_PASSWORD: "betts",
	}
	_, err = newCloudPakForDataAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_AUTH_URL:         "cp4d-url",
		PROPNAME_USERNAME:         "mookie",
		PROPNAME_PASSWORD:         "betts",
		PROPNAME_AUTH_DISABLE_SSL: "true",
	}
	authenticator, err := newCloudPakForDataAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, "cp4d-url", authenticator.URL)
	assert.Equal(t, "mookie", authenticator.Username)
	assert.Equal(t, "betts", authenticator.Password)
	assert.Equal(t, true, authenticator.DisableSSLVerification)
}
