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
	"time"

	assert "github.com/stretchr/testify/assert"
)

func TestIamConfigErrors(t *testing.T) {
	var err error

	// Missing ApiKey.
	_, err = NewIamAuthenticator("", "", "foo", "bar", false, nil)
	assert.NotNil(t, err)

	// Invalid ApiKey.
	_, err = NewIamAuthenticator("{invalid-apikey}", "", "foo", "bar", false, nil)
	assert.NotNil(t, err)

	// Missing ClientId.
	_, err = NewIamAuthenticator("my-apikey", "", "", "bar", false, nil)
	assert.NotNil(t, err)

	// Missing ClientSecret.
	_, err = NewIamAuthenticator("my-apikey", "", "foo", "", false, nil)
	assert.NotNil(t, err)
}

func TestIamGetTokenSuccess(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		expiration := GetCurrentTime() + 3600
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.Equal(t, ok, false)
		} else {
			fmt.Fprintf(w, `{
				"access_token": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			username, password, ok := r.BasicAuth()
			assert.Equal(t, ok, true)
			assert.Equal(t, "mookie", username)
			assert.Equal(t, "betts", password)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Force expiration and verify that we got the second access token.
	authenticator.tokenData.Expiration = GetCurrentTime() - 3600
	authenticator.ClientId = "mookie"
	authenticator.ClientSecret = "betts"
	_, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.tokenData)
	assert.Equal(t, "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		authenticator.tokenData.AccessToken)
}

func TestIamGetCachedToken(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		expiration := GetCurrentTime() + 3600
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.Equal(t, ok, false)
		} else {
			fmt.Fprintf(w, `{
				"access_token": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			_, _, ok := r.BasicAuth()
			assert.Equal(t, ok, true)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Subsequent fetch should still return first access token.
	token, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)
}

func TestIamBackgroundTokenRefresh(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiration := GetCurrentTime() + 3600
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.Equal(t, ok, false)
		} else {
			fmt.Fprintf(w, `{
				"access_token": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			_, _, ok := r.BasicAuth()
			assert.Equal(t, ok, false)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	authenticator.tokenData.RefreshTime = GetCurrentTime() - 720
	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// getToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)
	// Wait for the background thread to finish
	time.Sleep(5 * time.Second)
	token, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

}

func TestIamBackgroundTokenRefreshFailure(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiration := GetCurrentTime() + 3600
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.Equal(t, ok, false)
		} else {
			_, _ = w.Write([]byte("Sorry you are forbidden"))
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Successfully fetch the first token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	authenticator.tokenData.RefreshTime = GetCurrentTime() - 720
	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// getToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)
	// Wait for the background thread to finish
	time.Sleep(5 * time.Second)
	_, err = authenticator.getToken()
	assert.NotNil(t, err)
	assert.Equal(t, "Error while trying to get access token", err.Error())

}

func TestIamClientIdAndSecret(t *testing.T) {
	expiration := GetCurrentTime() + 3600
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"access_token": "oAeisG8yqPY7sFR_x66Z15",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": %d,
			"refresh_token": "jy4gl91BQ"
		}`, expiration)
		username, password, ok := r.BasicAuth()
		assert.Equal(t, ok, true)
		assert.Equal(t, "mookie", username)
		assert.Equal(t, "betts", password)
	}))
	defer server.Close()

	authenticator := &IamAuthenticator{
		ApiKey:       "bogus-apikey",
		URL:          server.URL,
		ClientId:     "mookie",
		ClientSecret: "betts",
	}

	token, err := authenticator.getToken()
	assert.Equal(t, "oAeisG8yqPY7sFR_x66Z15", token)
	assert.Nil(t, err)
}

func TestIamRefreshTimeCalculation(t *testing.T) {
	const timeToLive int64 = 3600
	const expireTime int64 = 1563911183
	const expected int64 = expireTime - 720 // 720 is 20% of 3600

	// Simulate a token server response.
	tokenResponse := &iamTokenServerResponse{
		ExpiresIn:  timeToLive,
		Expiration: expireTime,
	}

	// Create a new token data and verify the refresh time.
	tokenData, err := newIamTokenData(tokenResponse)
	assert.Nil(t, err)
	assert.Equal(t, expected, tokenData.RefreshTime)
}

func TestIamDisableSSL(t *testing.T) {
	expiration := GetCurrentTime() + 3600
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"access_token": "oAeisG8yqPY7sFR_x66Z15",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": %d,
			"refresh_token": "jy4gl91BQ"
		}`, expiration)
		_, _, ok := r.BasicAuth()
		assert.Equal(t, false, ok)
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", true, nil)
	assert.Nil(t, err)

	token, err := authenticator.getToken()
	assert.Equal(t, token, "oAeisG8yqPY7sFR_x66Z15")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.Client)
	assert.NotNil(t, authenticator.Client.Transport)
	transport, ok := authenticator.Client.Transport.(*http.Transport)
	assert.Equal(t, true, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.Equal(t, true, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestIamUserHeaders(t *testing.T) {
	expiration := GetCurrentTime() + 3600
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"access_token": "oAeisG8yqPY7sFR_x66Z15",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": %d,
			"refresh_token": "jy4gl91BQ"
		}`, expiration)
		_, _, ok := r.BasicAuth()
		assert.Equal(t, ok, false)
		assert.Equal(t, "Value1", r.Header.Get("Header1"))
		assert.Equal(t, "Value2", r.Header.Get("Header2"))
	}))
	defer server.Close()

	headers := make(map[string]string)
	headers["Header1"] = "Value1"
	headers["Header2"] = "Value2"

	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, headers)
	assert.Nil(t, err)

	token, err := authenticator.getToken()
	assert.Equal(t, "oAeisG8yqPY7sFR_x66Z15", token)
	assert.Nil(t, err)
}

func TestIamGetTokenFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	authenticator := &IamAuthenticator{
		ApiKey: "bogus-apikey",
		URL:    server.URL,
	}

	_, err := authenticator.getToken()
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

func TestIamGetTokenTimeoutError(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		expiration := GetCurrentTime() + 3600
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.Equal(t, ok, false)
		} else {
			time.Sleep(3 * time.Second)
			fmt.Fprintf(w, `{
				"access_token": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Force expiration and verify that we got a timeout error
	authenticator.tokenData.Expiration = GetCurrentTime() - 3600
	// Set the client timeout to something very low
	authenticator.Client.Timeout = time.Second * 2
	token, err = authenticator.getToken()
	assert.NotNil(t, err)
	assert.Equal(t, "", token)
}

func TestIamGetTokenServerError(t *testing.T) {
	firstCall := true
	expiration := GetCurrentTime() + 3600
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if firstCall {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.Equal(t, ok, false)
		} else {
			w.WriteHeader(http.StatusGatewayTimeout)
			_, _ = w.Write([]byte("Gateway Timeout"))
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Force expiration and verify that we got a server error
	authenticator.tokenData.Expiration = GetCurrentTime() - 3600
	token, err = authenticator.getToken()
	assert.NotNil(t, err)
	assert.Equal(t, "", token)
}

func TestNewIamAuthenticatorFromMap(t *testing.T) {
	_, err := newIamAuthenticatorFromMap(nil)
	assert.NotNil(t, err)

	var props = map[string]string{
		PROPNAME_AUTH_URL: "iam-url",
	}
	_, err = newIamAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_APIKEY: "",
	}
	_, err = newIamAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_APIKEY: "my-apikey",
	}
	authenticator, err := newIamAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, "my-apikey", authenticator.ApiKey)

	props = map[string]string{
		PROPNAME_APIKEY:           "my-apikey",
		PROPNAME_AUTH_DISABLE_SSL: "huh???",
		PROPNAME_CLIENT_ID:        "mookie",
		PROPNAME_CLIENT_SECRET:    "betts",
	}
	authenticator, err = newIamAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, "my-apikey", authenticator.ApiKey)
	assert.Equal(t, false, authenticator.DisableSSLVerification)
	assert.Equal(t, "mookie", authenticator.ClientId)
	assert.Equal(t, "betts", authenticator.ClientSecret)
}
