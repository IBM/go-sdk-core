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
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// for handling token management
const (
	DefaultIAMURL            = "https://iam.cloud.ibm.com/identity/token"
    DefaultIAMClientId       = "bx"
    DefaultIAMClientSecret   = "bx"
	DefaultContentType       = "application/x-www-form-urlencoded"
	RequestTokenGrantType    = "urn:ibm:params:oauth:grant-type:apikey"
	RequestTokenResponseType = "cloud_iam"
	RefreshTokenGrantType    = "refresh_token"
)

// TokenInfo : Response struct from token request
type TokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Expiration   int64  `json:"expiration"`
}

// TokenManager : IAM token information
type TokenManager struct {
	userAccessToken string
	iamAPIkey       string
	iamURL          string
	iamClientId     string
	iamClientSecret string
	tokenInfo       *TokenInfo
	client          *http.Client
}

// NewTokenManager : Instantiate TokenManager
func NewTokenManager(iamAPIkey string, iamURL string, userAccessToken string,
    iamClientId string, iamClientSecret string) (*TokenManager, error) {
	if iamURL == "" {
		iamURL = DefaultIAMURL
	}
	
	if iamClientId == "" && iamClientSecret == "" {
	    iamClientId = DefaultIAMClientId
	    iamClientSecret = DefaultIAMClientSecret
	} else if iamClientId != "" && iamClientSecret != "" {
	    // Do nothing as this is the valid scenario
	} else {
        // Only one of client id/secret was specified... error.
        return nil, fmt.Errorf("You specified only one of 'iamClientId' and 'iamClientSecret', but you must supply both values together or supply neither of them.")       
	}

	tokenManager := TokenManager{
		iamAPIkey:       iamAPIkey,
		iamURL:          iamURL,
		iamClientId:     iamClientId,
		iamClientSecret: iamClientSecret,
		userAccessToken: userAccessToken,
		tokenInfo:       &TokenInfo{},

		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
	return &tokenManager, nil
}

// GetToken : Return token set by user or fresh token
// The source of the token is determined by the following logic:
// 1. If user provides their own managed access token, assume it is valid and send it
// 2. If this class is managing tokens and does not yet have one, make a request for one
// 3. If this class is managing tokens and the token has expired refresh it. In case the refresh token is expired, get a new one
// If this class is managing tokens and has a valid token stored, send it
func (tm *TokenManager) GetToken() (string, error) {
	if tm.userAccessToken != "" {
		return tm.userAccessToken, nil
	} else if tm.tokenInfo.AccessToken == "" {
		tokenInfo, err := tm.requestToken()
		tm.saveTokenInfo(tokenInfo)
		return tm.tokenInfo.AccessToken, err
	} else if tm.isTokenExpired() {
		var tokenInfo *TokenInfo
		var err error
		if tm.isRefreshTokenExpired() {
			tokenInfo, err = tm.requestToken()
		} else {
			tokenInfo, err = tm.refreshToken()
		}
		tm.saveTokenInfo(tokenInfo)
		return tm.tokenInfo.AccessToken, err
	}
	return tm.tokenInfo.AccessToken, nil
}

// SetAccessToken : sets a self-managed IAM access token.
// The access token should be valid and not yet expired.
func (tm *TokenManager) SetAccessToken(userAccessToken string) {
	tm.userAccessToken = userAccessToken
}

// SetIAMAPIKey : Set API key so that SDK manages token
func (tm *TokenManager) SetIAMAPIKey(key string) {
	tm.iamAPIkey = key
}

// makes an HTTP request
func (tm *TokenManager) request(req *http.Request) (*TokenInfo, error) {
    req.SetBasicAuth(tm.iamClientId, tm.iamClientSecret)
	resp, err := tm.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp != nil {
			buff := new(bytes.Buffer)
			buff.ReadFrom(resp.Body)
			return nil, fmt.Errorf(buff.String())
		}
	}

	tokenInfo := TokenInfo{}
	json.NewDecoder(resp.Body).Decode(&tokenInfo)
	defer resp.Body.Close()
	return &tokenInfo, nil
}

// Request an IAM token using an API key
func (tm *TokenManager) requestToken() (*TokenInfo, error) {
	builder := NewRequestBuilder(POST).
		ConstructHTTPURL(tm.iamURL, nil, nil)

	builder.AddHeader(CONTENT_TYPE, DefaultContentType).
		AddHeader(Accept, APPLICATION_JSON)

	// Add form data
	builder.AddFormData("grant_type", "", "", RequestTokenGrantType).
		AddFormData("apikey", "", "", tm.iamAPIkey).
		AddFormData("response_type", "", "", RequestTokenResponseType)

	req, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return tm.request(req)
}

// Refresh an IAM token using a refresh token
func (tm *TokenManager) refreshToken() (*TokenInfo, error) {
	builder := NewRequestBuilder(POST).
		ConstructHTTPURL(tm.iamURL, nil, nil)

	builder.AddHeader(CONTENT_TYPE, DefaultContentType).
		AddHeader(Accept, APPLICATION_JSON)

	builder.AddFormData("grant_type", "", "", RefreshTokenGrantType).
		AddFormData("refresh_token", "", "", tm.tokenInfo.RefreshToken)

	req, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return tm.request(req)
}

// Check if currently stored token is expired.
// Using a buffer to prevent the edge case of the
// oken expiring before the request could be made.
// The buffer will be a fraction of the total TTL. Using 80%.
func (tm *TokenManager) isTokenExpired() bool {
	buffer := 0.8
	expiresIn := tm.tokenInfo.ExpiresIn
	expireTime := tm.tokenInfo.Expiration
	refreshTime := expireTime - (expiresIn * int64(1.0-buffer))
	currTime := time.Now().Unix()
	return refreshTime < currTime
}

// Used as a fail-safe to prevent the condition of a refresh token expiring,
// which could happen after around 30 days. This function will return true
// if it has been at least 7 days and 1 hour since the last token was set
func (tm *TokenManager) isRefreshTokenExpired() bool {
	if tm.tokenInfo.Expiration == 0 {
		return true
	}

	sevenDays := int64(7 * 24 * 3600)
	currTime := time.Now().Unix()
	newTokenTime := tm.tokenInfo.Expiration + sevenDays
	return newTokenTime < currTime
}

// Save the response from the IAM service request to the object's state.
func (tm *TokenManager) saveTokenInfo(tokenInfo *TokenInfo) {
	tm.tokenInfo = tokenInfo
}
