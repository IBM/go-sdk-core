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
	"time"

	assert "github.com/stretchr/testify/assert"
)

func TestIAMRequestTokenSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"access_token": "oAeisG8yqPY7sFR_x66Z15",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": 1524167011,
			"refresh_token": "jy4gl91BQ"
		}`)
		username, password, ok := r.BasicAuth()
		assert.Equal(t, ok, true)
		assert.Equal(t, username, "bx")
		assert.Equal(t, password, "bx")
	}))
	defer server.Close()

	tokenManager, err := NewIAMAuthenticator(&IAMConfig{
		ApiKey: "bogus-apikey",
		URL:    server.URL,
	})
	assert.Nil(t, err)

	tokenInfo, err := tokenManager.requestToken()
	assert.Equal(t, tokenInfo.AccessToken, "oAeisG8yqPY7sFR_x66Z15")
	assert.Nil(t, err)
}

func TestIAMRequestTokenFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	tokenManager, err := NewIAMAuthenticator(&IAMConfig{
		ApiKey: "bogus-apikey",
		URL:    server.URL,
	})
	assert.Nil(t, err)

	_, err = tokenManager.requestToken()
	assert.Equal(t, "Sorry you are forbidden", err.Error())
}

func TestIsIAMTokenExpired(t *testing.T) {
	tokenManager, err := NewIAMAuthenticator(&IAMConfig{
		ApiKey: "bogus-apikey",
	})
	assert.Equal(t, err, nil)

	isExpired := tokenManager.isTokenExpired()
	assert.Equal(t, isExpired, true)

	tokenManager.timeForNewToken = GetCurrentTime() + 3000
	isExpired = tokenManager.isTokenExpired()
	assert.Equal(t, isExpired, false)
}

func TestGetTokenWithUserAccessToken(t *testing.T) {
	tokenManager, err := NewIAMAuthenticator(&IAMConfig{
		ApiKey: "bogus-apikey",
	})
	assert.Nil(t, err)
	assert.NotNil(t, tokenManager)
	assert.Nil(t, tokenManager.tokenInfo)

	tokenManager.SetIAMAccessToken("user access token")
	token, err := tokenManager.GetToken()
	assert.Nil(t, err)
	assert.Nil(t, tokenManager.tokenInfo)
	assert.Equal(t, "user access token", token)

	tokenManager, err = NewIAMAuthenticator(&IAMConfig{
		AccessToken: "user access token #2",
	})
	assert.Nil(t, err)
	assert.NotNil(t, tokenManager)
	assert.Nil(t, tokenManager.tokenInfo)

	token, err = tokenManager.GetToken()
	assert.Nil(t, err)
	assert.Nil(t, tokenManager.tokenInfo)
	assert.Equal(t, "user access token #2", token)
}

func TestGetToken(t *testing.T) {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": 1524167011,
				"refresh_token": "jy4gl91BQ"
			}`)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{
				"access_token": "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": 1524167011,
				"refresh_token": "jy4gl91BQ"
			}`)
		}
	}))
	defer server.Close()

	tokenManager, err := NewIAMAuthenticator(&IAMConfig{
		ApiKey: "bogus-apikey",
		URL:    server.URL,
	})
	assert.Nil(t, err)
	assert.Nil(t, tokenManager.tokenInfo)

	token, err := tokenManager.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		token)
	assert.NotNil(t, tokenManager.tokenInfo)

	// Case 2 b: force expiration
	tokenManager.timeForNewToken = time.Now().Unix() - 3600
	_, err = tokenManager.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, tokenManager.tokenInfo)
	assert.Equal(t, "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
		tokenManager.tokenInfo.AccessToken)

	// case 3
	tokenManager.tokenInfo = &IAMTokenInfo{
		AccessToken:  "test",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Expiration:   time.Now().Unix(),
		RefreshToken: "jy4gl91BQ",
	}
	token, err = tokenManager.GetToken()
	assert.Equal(t, tokenManager.tokenInfo.AccessToken, token)
}

func TestIamClientIdOnly(t *testing.T) {
	_, err := NewIAMAuthenticator(&IAMConfig{
		ApiKey:       "bogus-apikey",
		URL:          "",
		AccessToken:  "",
		ClientId:     "foo",
		ClientSecret: "",
	})
	assert.NotNil(t, err)
}

func TestIamClientSecretOnly(t *testing.T) {
	_, err := NewIAMAuthenticator(&IAMConfig{
		ApiKey:       "bogus-apikey",
		URL:          "",
		AccessToken:  "",
		ClientId:     "",
		ClientSecret: "bar",
	})
	assert.NotNil(t, err)
}

func TestCalcTimeForNewToken(t *testing.T) {
	const timeToLive int64 = 3600
	const expireTime int64 = 1563911183
	const expected int64 = expireTime - 720 // 720 is 20% of 3600

	tokenManager, err := NewIAMAuthenticator(&IAMConfig{
		ApiKey: "bogus-apikey",
	})
	assert.Nil(t, err)

	actual := tokenManager.calcTimeForNewToken(expireTime, timeToLive)
	assert.Equal(t, expected, actual)
}
