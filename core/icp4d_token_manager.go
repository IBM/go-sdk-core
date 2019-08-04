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

	ERRORMSG_USERPW_MISSING = "The Username and Password properties are required if a user-managed access token is not specified, but one or both were not specified."
	ERRORMSG_URL_MISSING    = "The URL property is required, but was not specified."
)

// This struct contains the configuration associated with the ICP4D Authenticator.
type ICP4DConfig struct {
	URL                    string
	Username               string
	Password               string
	AccessToken            string
	DisableSSLVerification bool
}

// Validates the specified ICP4DConfig instance.
func (config ICP4DConfig) Validate() error {
	// If the AccessToken is specified, then we'll use that directly and no other validation is necessary.
	if config.AccessToken != "" {
		return nil
	}

	// If AccessToken is not specified, then Username and Password are required.
	if config.Username == "" || config.Password == "" {
		return fmt.Errorf(ERRORMSG_USERPW_MISSING)
	}

	// Finally, make sure the URL field was specified.
	if config.URL == "" {
		return fmt.Errorf(ERRORMSG_URL_MISSING)
	}

	return nil
}

func (ICP4DConfig) AuthenticationType() string {
	return AUTHTYPE_ICP4D
}

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
type ICP4DAuthenticator struct {
	username        string
	password        string
	url             string
	userAccessToken string
	client          *http.Client
	timeForNewToken int64
	tokenInfo       *ICP4DTokenInfo
}

// NewICP4DAuthenticator : Instantiate a new ICP4DAuthenticator
func NewICP4DAuthenticator(config *ICP4DConfig) (*ICP4DAuthenticator, error) {
	// Make sure the config is valid.
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s%s", config.URL, PRE_AUTH_PATH)

	authenticator := &ICP4DAuthenticator{
		username:        config.Username,
		password:        config.Password,
		url:             url,
		userAccessToken: config.AccessToken,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	if config.DisableSSLVerification {
		authenticator.DisableSSLVerification()
	}

	return authenticator, nil
}

func (ICP4DAuthenticator) AuthenticationType() string {
	return AUTHTYPE_ICP4D
}

// Perform the authentication by constructing the Authorization header
// value from the ICP4D access token (either user-supplied or obtained from the token service).
func (authenticator ICP4DAuthenticator) Authenticate(request *http.Request) error {
	token, err := authenticator.GetToken()
	if err != nil {
		return err
	}

	authHeader := fmt.Sprintf(`%s %s`, TOKENTYPE_BEARER, token)
	request.Header.Set(HEADER_NAME_AUTHORIZATION, authHeader)
	return nil
}

// GetToken : Return token set by user or fresh token
// The source of the token is determined by the following logic:
// 1. If user provides their own managed access token, assume it is valid and send it
// 2.  a) If this class is managing tokens and does not yet have one, make a request for one
//     b) If this class is managing tokens and the token has expired, request a new one
// 3. If this class is managing tokens and has a valid token stored, send it
func (tm *ICP4DAuthenticator) GetToken() (string, error) {
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
func (tm *ICP4DAuthenticator) SetICP4DAccessToken(userAccessToken string) {
	tm.userAccessToken = userAccessToken
}

func (tm *ICP4DAuthenticator) DisableSSLVerification() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	tm.client.Transport = tr
}

// requestToken : A new access token for ICP4D authentication
func (tm *ICP4DAuthenticator) requestToken() (*ICP4DTokenInfo, error) {
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

func (tm *ICP4DAuthenticator) isTokenExpired() bool {
	if tm.timeForNewToken == 0 {
		return true
	}

	currentTime := GetCurrentTime()
	return tm.timeForNewToken < currentTime
}

// Decode and saves the access token
func (tm *ICP4DAuthenticator) saveToken(tokenInfo *ICP4DTokenInfo) {
	tm.tokenInfo = tokenInfo

	claims := jwt.StandardClaims{}
	if token, _ := jwt.ParseWithClaims(tm.tokenInfo.AccessToken, &claims, nil); token != nil {
		timeToLive := claims.ExpiresAt - claims.IssuedAt
		expireTime := claims.ExpiresAt
		timeForNewToken := expireTime - int64(float64(timeToLive)*0.2)
		tm.timeForNewToken = timeForNewToken
	}
}
