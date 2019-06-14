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

	jwt "github.com/dgrijalva/jwt-go"
)

// constants for IAM token authentication
const (
	DEFAULT_IAM_URL             = "https://iam.cloud.ibm.com/identity/token"
	DEFAULT_IAM_CLIENT_ID       = "bx"
	DEFAULT_IAM_CLIENT_SECRET   = "bx"
	DEFAULT_CONTENT_TYPE        = "application/x-www-form-urlencoded"
	REQUEST_TOKEN_GRANT_TYPE    = "urn:ibm:params:oauth:grant-type:apikey"
	REQUEST_TOKEN_RESPONSE_TYPE = "cloud_iam"
)

// IAMTokenInfo : Response struct from token request
type IAMTokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Expiration   int64  `json:"expiration"`
}

// IAMTokenManager : IAM token manager
type IAMTokenManager struct {
	userAccessToken string
	iamAPIkey       string
	iamURL          string
	iamClientId     string
	iamClientSecret string
	client          *http.Client
	timeForNewToken int64
	tokenInfo       *IAMTokenInfo
}

// NewIAMTokenManager : Instantiate IAMTokenManager
func NewIAMTokenManager(iamAPIkey string, iamURL string, userAccessToken string,
	iamClientId string, iamClientSecret string) (*IAMTokenManager, error) {
	if iamURL == "" {
		iamURL = DEFAULT_IAM_URL
	}

	if iamClientId == "" && iamClientSecret == "" {
		iamClientId = DEFAULT_IAM_CLIENT_ID
		iamClientSecret = DEFAULT_IAM_CLIENT_SECRET
	} else if iamClientId != "" && iamClientSecret != "" {
		// Do nothing as this is the valid scenario
	} else {
		// Only one of client id/secret was specified... error.
		return nil, fmt.Errorf("You specified only one of 'iamClientId' and 'iamClientSecret', but you must supply both values together or supply neither of them.")
	}

	tokenManager := IAMTokenManager{
		userAccessToken: userAccessToken,
		iamAPIkey:       iamAPIkey,
		iamURL:          iamURL,
		iamClientId:     iamClientId,
		iamClientSecret: iamClientSecret,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
	return &tokenManager, nil
}

// GetToken : Return token set by user or fresh token
// The source of the token is determined by the following logic:
// 1. If user provides their own managed access token, assume it is valid and send it
// 2.  a) If this class is managing tokens and does not yet have one, make a request for one
//     b) If this class is managing tokens and the token has expired, request a new one
// 3. If this class is managing tokens and has a valid token stored, send it
func (tm *IAMTokenManager) GetToken() (string, error) {
	if tm.userAccessToken != "" {
		return tm.userAccessToken, nil
	} else if tm.tokenInfo == nil || tm.isTokenExpired() {
		tokenInfo, err := tm.requestToken()
		if err != nil {
			return "", err
		}
		tm.saveToken(tokenInfo)
	}
	return tm.tokenInfo.AccessToken, nil
}

// SetIAMAccessToken : sets a self-managed access token.
// The access token should be valid and not yet expired.
func (tm *IAMTokenManager) SetIAMAccessToken(userAccessToken string) {
	tm.userAccessToken = userAccessToken
}

// SetIAMAPIKey : Set API key so that SDK manages token
func (tm *IAMTokenManager) SetIAMAPIKey(key string) {
	tm.iamAPIkey = key
}

// Request an IAM token using an API key
func (tm *IAMTokenManager) requestToken() (*IAMTokenInfo, error) {
	builder := NewRequestBuilder(POST).
		ConstructHTTPURL(tm.iamURL, nil, nil)

	builder.AddHeader(CONTENT_TYPE, DEFAULT_CONTENT_TYPE).
		AddHeader(Accept, APPLICATION_JSON)

	builder.AddFormData("grant_type", "", "", REQUEST_TOKEN_GRANT_TYPE).
		AddFormData("apikey", "", "", tm.iamAPIkey).
		AddFormData("response_type", "", "", REQUEST_TOKEN_RESPONSE_TYPE)

	req, err := builder.Build()
	if err != nil {
		return nil, err
	}
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

	tokenInfo := IAMTokenInfo{}
	json.NewDecoder(resp.Body).Decode(&tokenInfo)
	defer resp.Body.Close()
	return &tokenInfo, nil
}

func (tm *IAMTokenManager) isTokenExpired() bool {
	if tm.timeForNewToken == 0 {
		return true
	}

	currentTime := GetCurrentTime()
	return tm.timeForNewToken < currentTime
}

// Decode and saves the access token
func (tm *IAMTokenManager) saveToken(tokenInfo *IAMTokenInfo) {
	accessToken := tokenInfo.AccessToken

	claims := jwt.StandardClaims{}
	if token, _ := jwt.ParseWithClaims(accessToken, &claims, nil); token != nil {
		timeToLive := claims.ExpiresAt - claims.IssuedAt
		expireTime := claims.ExpiresAt
		fractionOfTimeToLive := 0.8
		timeForNewToken := expireTime - (timeToLive * int64(1.0-fractionOfTimeToLive))
		tm.timeForNewToken = timeForNewToken
	}

	tm.tokenInfo = tokenInfo
}
