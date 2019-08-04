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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func TestGetTokenSuccess(t *testing.T) {
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
		// fmt.Println("mock server Authorization: ", r.Header.Get("Authorization"))

		// Note: the header value below reflects "john"/"snow" for username and password.
		assert.Equal(t, "Basic am9objpzbm93", r.Header.Get("Authorization"))
	}))
	defer server.Close()

	tokenManager, err := NewICP4DAuthenticator(&ICP4DConfig{
		URL:      server.URL,
		Username: "john",
		Password: "snow",
	})
	assert.Nil(t, err)
	assert.NotNil(t, tokenManager)
	tokenManager.DisableSSLVerification()

	// case 2a
	accessToken, _ := tokenManager.GetToken()
	assert.Equal(t, accessToken, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI")

	// case 1
	tokenManager, err = NewICP4DAuthenticator(&ICP4DConfig{
		URL:         server.URL,
		AccessToken: "user access token",
		Username:    "john",
		Password:    "snow",
	})
	assert.Nil(t, err)
	assert.NotNil(t, tokenManager)
	assert.Nil(t, tokenManager.tokenInfo)

	accessToken, err = tokenManager.GetToken()
	assert.Equal(t, accessToken, "user access token")
	assert.Equal(t, err, nil)

	// case 2b, token expired
	tokenManager.timeForNewToken = GetCurrentTime() - 3000
	tokenManager.SetICP4DAccessToken("")
	accessToken, _ = tokenManager.GetToken()
	assert.Equal(t, accessToken, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI")
}

func TestGetTokenFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	tokenManager, err := NewICP4DAuthenticator(&ICP4DConfig{
		URL:      server.URL,
		Username: "john",
		Password: "snow",
	})
	assert.Nil(t, err)
	assert.NotNil(t, tokenManager)

	_, err = tokenManager.GetToken()
	assert.Equal(t, err.Error(), "Sorry you are forbidden")
}

func TestIsTokenExpired(t *testing.T) {
	tokenManager, err := NewICP4DAuthenticator(&ICP4DConfig{
		URL:      "http://myhost/my/url",
		Username: "john",
		Password: "snow",
	})
	assert.Nil(t, err)
	assert.NotNil(t, tokenManager)

	isExpired := tokenManager.isTokenExpired()
	assert.True(t, isExpired)

	tokenManager.timeForNewToken = GetCurrentTime() + 3000
	isExpired = tokenManager.isTokenExpired()
	assert.False(t, isExpired)
}
