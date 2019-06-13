package core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

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

// constants for ICP4D
const (
	PRE_AUTH_PATH = "/v1/preauth/validateAuth"
)

// ICP4DTokenInfo : Response struct from token request
type ICP4DTokenInfo struct {
	Username    string   `json:"username,omitempty"`
	Role        string   `json:"role,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	Subject     string   `json:"sub,omitempty"`
	Issuer      string   `json:"iss,omitempty"`
	Audience    string   `json:"aud,omitempty"`
	UID         string   `json:"uid,omitempty"`
	MessageCode string   `json:"_messageCode_,omitempty"`
	Message     string   `json:"message,omitempty"`
	AccessToken string   `json:"accessToken,omitempty"`
}

// ICP4DTokenManager : Manager for handling ICP4D authentication
type ICP4DTokenManager struct {
	username        string
	password        string
	url             string
	userAccessToken string
	client          *http.Client
	timeForNewToken int64
	tokenInfo       *ICP4DTokenInfo
}

// NewICP4DTokenManager : New instance of ICP4D token manager
func NewICP4DTokenManager(icp4dURL, username, password, accessToken string) *ICP4DTokenManager {
	url := fmt.Sprintf("%s%s", icp4dURL, PRE_AUTH_PATH)

	icp4dTokenManager := ICP4DTokenManager{
		username:        username,
		password:        password,
		url:             url,
		userAccessToken: accessToken,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
	return &icp4dTokenManager
}

// GetToken : Return token set by user or fresh token
// The source of the token is determined by the following logic:
// 1. If user provides their own managed access token, assume it is valid and send it
// 2.  a) If this class is managing tokens and does not yet have one, make a request for one
//     b) If this class is managing tokens and the token has expired, request a new one
// 3. If this class is managing tokens and has a valid token stored, send it
func (tm *ICP4DTokenManager) GetToken() (string, error) {
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

// SetICP4DAccessToken : sets a self-managed access token.
// The access token should be valid and not yet expired.
func (tm *ICP4DTokenManager) SetICP4DAccessToken(userAccessToken string) {
	tm.userAccessToken = userAccessToken
}

// DisableSSLVerification skips SSL verification
func (tm *ICP4DTokenManager) DisableSSLVerification() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	tm.client.Transport = tr
}

// requestToken : A new access token for ICP4D authentication
func (tm *ICP4DTokenManager) requestToken() (*ICP4DTokenInfo, error) {
	builder := NewRequestBuilder(GET).ConstructHTTPURL(tm.url, nil, nil)

	req, err := builder.Build()
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(tm.username, tm.password)

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

	tokenInfo := ICP4DTokenInfo{}
	json.NewDecoder(resp.Body).Decode(&tokenInfo)
	defer resp.Body.Close()
	return &tokenInfo, nil
}

func (tm *ICP4DTokenManager) isTokenExpired() bool {
	if tm.timeForNewToken == 0 {
		return true
	}

	currentTime := GetCurrentTime()
	return tm.timeForNewToken < currentTime
}

// Decode and saves the access token
func (tm *ICP4DTokenManager) saveToken(tokenInfo *ICP4DTokenInfo) {
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
