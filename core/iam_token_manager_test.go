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

	tokenManager, err := NewIAMTokenManager("", server.URL, "", "", "")
	assert.Equal(t, err, nil)

	tokenInfo, err := tokenManager.requestToken()
	assert.Equal(t, tokenInfo.AccessToken, "oAeisG8yqPY7sFR_x66Z15")
	assert.Equal(t, err, nil)
}

func TestIAMRequestTokenFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	tokenManager, err := NewIAMTokenManager("", server.URL, "", "", "")
	assert.Equal(t, err, nil)

	_, err = tokenManager.requestToken()
	assert.Equal(t, err.Error(), "Sorry you are forbidden")
}

func TestIsIAMTokenExpired(t *testing.T) {
	tokenManager, err := NewIAMTokenManager("iamApiKey", "", "", "", "")
	assert.Equal(t, err, nil)

	isExpired := tokenManager.isTokenExpired()
	assert.Equal(t, isExpired, true)

	tokenManager.timeForNewToken = GetCurrentTime() + 3000
	isExpired = tokenManager.isTokenExpired()
	assert.Equal(t, isExpired, false)
}

func TestGetToken(t *testing.T) {
	// # Case 1:
	tokenManager, err := NewIAMTokenManager("iamApiKey", "", "", "", "")
	assert.Equal(t, err, nil)

	tokenManager.SetIAMAccessToken("user access token")
	token, err := tokenManager.GetToken()
	assert.Equal(t, token, "user access token")
	assert.Equal(t, err, nil)

	// Case 2 a:
	firstCall := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if !firstCall {
			fmt.Fprintf(w, `{
				"access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": 1524167011,
				"refresh_token": "jy4gl91BQ"
			}`)
			firstCall = true
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
	tokenManager, err = NewIAMTokenManager("iamApiKey", server.URL, "", "", "")
	assert.Equal(t, err, nil)

	token, err = tokenManager.GetToken()
	assert.Equal(t, token, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI")
	assert.Nil(t, err)

	// Case 2 b:
	tokenManager.timeForNewToken = time.Now().Unix() - 3600
	tokenManager.GetToken()
	assert.Equal(t, tokenManager.tokenInfo.AccessToken, "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI")

	// case 3
	tokenManager.tokenInfo = &IAMTokenInfo{
		AccessToken:  "test",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Expiration:   time.Now().Unix(),
		RefreshToken: "jy4gl91BQ",
	}
	token, err = tokenManager.GetToken()
	assert.Equal(t, token, tokenManager.tokenInfo.AccessToken)
}

func TestIamClientIdOnly(t *testing.T) {
	_, err := NewIAMTokenManager("iamApiKey", "", "", "foo", "")
	assert.NotEqual(t, err, nil)
}

func TestIamClientSecretOnly(t *testing.T) {
	_, err := NewIAMTokenManager("iamApiKey", "", "", "", "bar")
	assert.NotEqual(t, err, nil)
}

func TestCalcTimeForNewToken(t *testing.T) {
	const timeToLive int64 = 3600
	const expireTime int64 = 1563911183
	const expected int64 = expireTime - 720 // 720 is 20% of 3600

	tokenManager, err := NewIAMTokenManager("iamApiKey", "", "", "", "")
	assert.Equal(t, err, nil)

	actual := tokenManager.calcTimeForNewToken(expireTime, timeToLive)
	assert.Equal(t, expected, actual)
}
