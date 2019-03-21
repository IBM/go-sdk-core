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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

func TestRequestToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"access_token": "oAeisG8yqPY7sFR_x66Z15",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": 1524167011,
			"refresh_token": "jy4gl91BQ"
		}`)
	}))
	defer server.Close()

	tokenManager := NewTokenManager("", server.URL, "")
	tokenInfo, err := tokenManager.requestToken()
	assert.Equal(t, tokenInfo.AccessToken, "oAeisG8yqPY7sFR_x66Z15")
	assert.Equal(t, err, nil)
}

func TestRequestTokenFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	tokenManager := NewTokenManager("", server.URL, "")
	_, err := tokenManager.requestToken()
	assert.Equal(t, err.Error(), "Sorry you are forbidden")
}

func TestRefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"access_token": "oAeisG8yqPY7sFR_x66Z15",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": 1524167011,
			"refresh_token": "jy4gl91BQ"
		}`)
	}))
	defer server.Close()

	tokenManager := NewTokenManager("", server.URL, "")
	tokenInfo, err := tokenManager.refreshToken()
	assert.Equal(t, tokenInfo.AccessToken, "oAeisG8yqPY7sFR_x66Z15")
	assert.Equal(t, err, nil)
}

func TestIsTokenExpired(t *testing.T) {
	tokenManager := NewTokenManager("iamApiKey", "", "")
	tokenManager.tokenInfo = &TokenInfo{
		AccessToken:  "oAeisG8yqPY7sFR_x66Z15",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Expiration:   time.Now().Unix() + 6000,
		RefreshToken: "jy4gl91BQ",
	}

	assert.Equal(t, tokenManager.isTokenExpired(), false)
	tokenManager.tokenInfo.Expiration = time.Now().Unix() - 3600
	assert.Equal(t, tokenManager.isTokenExpired(), true)
}

func TestIsRefreshTokenExpired(t *testing.T) {
	tokenManager := NewTokenManager("iamApiKey", "", "")
	tokenManager.tokenInfo = &TokenInfo{
		AccessToken:  "oAeisG8yqPY7sFR_x66Z15",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Expiration:   time.Now().Unix(),
		RefreshToken: "jy4gl91BQ",
	}

	assert.Equal(t, tokenManager.isRefreshTokenExpired(), false)
	tokenManager.tokenInfo.Expiration = time.Now().Unix() - (8 * 24 * 3600)
	assert.Equal(t, tokenManager.isRefreshTokenExpired(), true)
}

func TestGetToken(t *testing.T) {
	// # Case 1:
	tokenManager := NewTokenManager("iamApiKey", "", "")
	tokenManager.SetAccessToken("user access token")
	token, err := tokenManager.GetToken()
	assert.Equal(t, token, "user access token")
	assert.Equal(t, err, nil)

	// Case 2:
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"access_token": "hellohello",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": 1524167011,
			"refresh_token": "jy4gl91BQ"
		}`)
	}))
	defer server.Close()
	tokenManager = NewTokenManager("", server.URL, "")
	tokenManager.SetIAMAPIKey("iamApiKey")
	token, err = tokenManager.GetToken()
	assert.Equal(t, token, "hellohello")
	assert.Nil(t, err)

	// Case 3: Refresh token not expired
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"access_token": "captain marvel",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": 1524167011,
			"refresh_token": "jy4gl91BQ"
		}`)
		body, _ := ioutil.ReadAll(r.Body)
		assert.Contains(t, string(body), "grant_type=refresh_token")
	}))
	defer server2.Close()

	tokenManager = NewTokenManager("iamApiKey", server2.URL, "")
	tokenManager.tokenInfo = &TokenInfo{
		AccessToken:  "oAeisG8yqPY7sFR_x66Z15",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Expiration:   time.Now().Unix(),
		RefreshToken: "jy4gl91BQ",
	}
	tokenManager.tokenInfo.Expiration = time.Now().Unix() - 3600
	// tokenManager.tokenInfo.Expiration = time.Now().Unix() - (20 * 24 * 3600)
	tokenManager.GetToken()
	assert.Equal(t, tokenManager.tokenInfo.AccessToken, "captain marvel")

	// Case 3: Refresh token expired
	server3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
				"access_token": "captain america",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": 1524167011,
				"refresh_token": "jy4gl91BQ"
			}`)
		body, _ := ioutil.ReadAll(r.Body)
		assert.Contains(t, string(body), "grant_type=urn")
	}))
	defer server2.Close()

	tokenManager = NewTokenManager("iamApiKey", server3.URL, "")
	tokenManager.tokenInfo = &TokenInfo{
		AccessToken:  "oAeisG8yqPY7sFR_x66Z15",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Expiration:   time.Now().Unix(),
		RefreshToken: "jy4gl91BQ",
	}
	tokenManager.tokenInfo.Expiration = time.Now().Unix() - (20 * 24 * 3600)
	tokenManager.GetToken()
	assert.Equal(t, tokenManager.tokenInfo.AccessToken, "captain america")

	tokenManager.tokenInfo = &TokenInfo{
		AccessToken:  "test",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		Expiration:   time.Now().Unix(),
		RefreshToken: "jy4gl91BQ",
	}
	token, err = tokenManager.GetToken()
	assert.Equal(t, token, tokenManager.tokenInfo.AccessToken)
}
