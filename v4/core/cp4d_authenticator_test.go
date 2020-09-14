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

func TestCp4dAuthenticateFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Sorry you are not authorized"))
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "mookie", "betts", false, nil)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, authenticator.AuthenticationType(), AUTHTYPE_CP4D)

	// Create a new Request object.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://localhost/placeholder/url", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	err = authenticator.Authenticate(request)
	// Validate the resulting error is a valid
	assert.NotNil(t, err)
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	// The casted error should match the original error message
	assert.Equal(t, err.Error(), authErr.Error())
}

func TestCp4dGetTokenSuccess(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
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
			firstCall = false
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "mookie", username)
			assert.Equal(t, "betts", password)
		} else {
			fmt.Fprintf(w, `{
				"username": "admin",
  				"role": "Admin",
  				"permissions": [
    				"administrator",
    				"manage_catalog",
    				"access_catalog",
    				"manage_policies",
    				"access_policies",
    				"virtualize_transform",
    				"can_provision",
    				"deployment_admin"
  				],
  				"sub": "admin",
  				"iss": "test",
  				"aud": "DSX",
  				"uid": "999",
  				"accessToken": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
  				"_messageCode_": "success",
  				"message": "success"
			}`)
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "mookie", username)
			assert.Equal(t, "betts", password)
		}
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
	assert.Equal(t, "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		accessToken)
}

func TestCp4dGetCachedToken(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
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
			firstCall = false
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "john", username)
			assert.Equal(t, "snow", password)
		} else {
			fmt.Fprintf(w, `{
				"username": "admin",
  				"role": "Admin",
  				"permissions": [
    				"administrator",
    				"manage_catalog",
    				"access_catalog",
    				"manage_policies",
    				"access_policies",
    				"virtualize_transform",
    				"can_provision",
    				"deployment_admin"
  				],
  				"sub": "admin",
  				"iss": "test",
  				"aud": "DSX",
  				"uid": "999",
  				"accessToken": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
  				"_messageCode_": "success",
  				"message": "success"
			}`)
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "john", username)
			assert.Equal(t, "snow", password)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "john", "snow", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// To mock the cache, set the expiration on the existing token to be somewhere in the valid timeframe
	authenticator.tokenData.Expiration = GetCurrentTime() + 9999

	// Subsequent fetch should still return first access token.
	token, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)
}

func TestCp4dBackgroundTokenRefresh(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
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
			firstCall = false
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "john", username)
			assert.Equal(t, "snow", password)
		} else {
			fmt.Fprintf(w, `{
				"username": "admin",
  				"role": "Admin",
  				"permissions": [
    				"administrator",
    				"manage_catalog",
    				"access_catalog",
    				"manage_policies",
    				"access_policies",
    				"virtualize_transform",
    				"can_provision",
    				"deployment_admin"
  				],
  				"sub": "admin",
  				"iss": "test",
  				"aud": "DSX",
  				"uid": "999",
  				"accessToken": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
  				"_messageCode_": "success",
  				"message": "success"
			}`)
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "john", username)
			assert.Equal(t, "snow", password)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "john", "snow", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	authenticator.tokenData.Expiration = GetCurrentTime() + 3600
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

func TestCp4dBackgroundTokenRefreshFailure(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
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
			firstCall = false
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "john", username)
			assert.Equal(t, "snow", password)
		} else {
			_, _ = w.Write([]byte("Sorry you are forbidden"))
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "john", "snow", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	authenticator.tokenData.Expiration = GetCurrentTime() + 3600
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
	assert.Equal(t, "Error while trying to parse access token!", err.Error())
	// We don't expect a AuthenticateError to be returned, so casting should fail
	_, ok := err.(*AuthenticationError)
	assert.False(t, ok)
}

func TestCp4dBackgroundTokenRefreshIdle(t *testing.T) {
	firstCall := true
	accessToken1 := "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	accessToken2 := "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
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
				"accessToken": %q,
				"_messageCode_":"success",
				"message":"success"
			}`, accessToken1)
			firstCall = false
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "john", username)
			assert.Equal(t, "snow", password)
		} else {
			fmt.Fprintf(w, `{
				"username": "admin",
  				"role": "Admin",
  				"permissions": [
    				"administrator",
    				"manage_catalog",
    				"access_catalog",
    				"manage_policies",
    				"access_policies",
    				"virtualize_transform",
    				"can_provision",
    				"deployment_admin"
  				],
  				"sub": "admin",
  				"iss": "test",
  				"aud": "DSX",
  				"uid": "999",
  				"accessToken": %q,
  				"_messageCode_": "success",
  				"message": "success"
			}`, accessToken2)
			username, password, ok := r.BasicAuth()
			assert.Equal(t, true, ok)
			assert.Equal(t, "john", username)
			assert.Equal(t, "snow", password)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "john", "snow", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// // Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, accessToken1,
		token)
	assert.NotNil(t, authenticator.tokenData)

	// Now simulate the client being idle for 10 minutes into the refresh time
	authenticator.tokenData.Expiration = GetCurrentTime() + 3600
	tenMinutesBeforeNow := GetCurrentTime() - 600
	authenticator.tokenData.RefreshTime = tenMinutesBeforeNow
	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// getToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, accessToken1,
		token)
	assert.NotNil(t, authenticator.tokenData)
	// RefreshTime should have advanced by 1 minute from the current time
	newRefreshTime := GetCurrentTime() + 60
	assert.Equal(t, newRefreshTime, authenticator.tokenData.RefreshTime)

	// In the next request, the RefreshTime should be unchanged and another thread
	// shouldn't be spawned to request another token once more since the first thread already spawned
	// a goroutine & refreshed the token.
	token, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, accessToken1,
		token)
	assert.NotNil(t, authenticator.tokenData)
	assert.Equal(t, newRefreshTime, authenticator.tokenData.RefreshTime)
	// // Wait for the background thread to finish and verify both the RefreshTime & tokenData were updated
	time.Sleep(5 * time.Second)
	token, err = authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, accessToken2,
		token)
	assert.NotNil(t, authenticator.tokenData)
	assert.NotEqual(t, newRefreshTime, authenticator.tokenData.RefreshTime)

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

	var expectedResponse = []byte("Sorry you are forbidden")

	_, err := authenticator.getToken()
	assert.NotNil(t, err)
	assert.Equal(t, "Sorry you are forbidden", err.Error())
	// We expect an AuthenticationError to be returned, so cast the returned error
	authError, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authError)
	assert.NotNil(t, authError.Error())
	assert.NotNil(t, authError.Response)
	rawResult := authError.Response.GetRawResult()
	assert.NotNil(t, rawResult)
	assert.Equal(t, expectedResponse, rawResult)
	statusCode := authError.Response.GetStatusCode()
	assert.Equal(t, "Sorry you are forbidden", authError.Error())
	assert.Equal(t, http.StatusForbidden, statusCode)
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

func TestCp4dGetTokenTimeoutError(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
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
			firstCall = false
		} else {
			time.Sleep(3 * time.Second)
			fmt.Fprintf(w, `{
				"username": "admin",
  				"role": "Admin",
  				"permissions": [
    				"administrator",
    				"manage_catalog",
    				"access_catalog",
    				"manage_policies",
    				"access_policies",
    				"virtualize_transform",
    				"can_provision",
    				"deployment_admin"
  				],
  				"sub": "admin",
  				"iss": "test",
  				"aud": "DSX",
  				"uid": "999",
  				"accessToken": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
  				"_messageCode_": "success",
  				"message": "success"
			}`)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "john", "snow", false, nil)
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
	assert.Equal(t, "", token)
	assert.NotNil(t, err)
	assert.NotNil(t, err.Error())
	// We don't expect a AuthenticateError to be returned, so casting should fail
	_, ok := err.(*AuthenticationError)
	assert.False(t, ok)
}

func TestCp4dGetTokenServerError(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if firstCall {
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
			firstCall = false
		} else {
			w.WriteHeader(http.StatusGatewayTimeout)
			_, _ = w.Write([]byte("Gateway Timeout"))
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "john", "snow", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.tokenData)

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.getToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, authenticator.tokenData)

	var expectedResponse = []byte("Gateway Timeout")

	// Force expiration and verify that we got a server error
	authenticator.tokenData.Expiration = GetCurrentTime() - 3600
	token, err = authenticator.getToken()
	assert.NotNil(t, err)
	// We expect an AuthenticationError to be returned, so cast the returned error
	authError, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authError)
	assert.NotNil(t, authError.Response)
	assert.NotNil(t, authError.Error())
	rawResult := authError.Response.GetRawResult()
	statusCode := authError.Response.GetStatusCode()
	assert.Equal(t, "Gateway Timeout", authError.Error())
	assert.Equal(t, expectedResponse, rawResult)
	assert.NotNil(t, rawResult)
	assert.Equal(t, http.StatusGatewayTimeout, statusCode)
	assert.Equal(t, "", token)
}
