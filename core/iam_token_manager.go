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

	FRACTION_OF_TIME_TO_LIVE = 0.8

	ERRORMSG_CLIENT_ID_SECRET = "You specified only one of 'ClientId' and 'ClientSecret', but those values must be supplied together."
	ERRORMSG_APIKEY_MISSING   = "The ApiKey property is required if the AccessToken property is not specified."
)

// This struct contains the configuration associated with the IAM Authenticator.
type IAMConfig struct {
	AccessToken  string
	ApiKey       string
	ClientId     string
	ClientSecret string
	URL          string
}

// Validates the specified IAMConfig instance.
func (config IAMConfig) Validate() error {
	// If the AccessToken is specified, then we'll use that directly and no other validation is necessary.
	if config.AccessToken != "" {
		return nil
	}

	// If AccessToken is not specified, then ApiKey is required.
	if config.ApiKey == "" {
		return fmt.Errorf(ERRORMSG_APIKEY_MISSING)
	}

	if HasBadFirstOrLastChar(config.ApiKey) {
		return fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "ApiKey", "ApiKey")
	}

	// Validate ClientId and ClientSecret.  They must both be specified togther or neither should be specified.
	if (config.ClientId == "" && config.ClientSecret == "") || (config.ClientId != "" && config.ClientSecret != "") {
		// Do nothing as this is the valid scenario
	} else {
		// Only one of client id/secret was specified... error.
		return fmt.Errorf(ERRORMSG_CLIENT_ID_SECRET)
	}

	return nil
}

func (IAMConfig) AuthenticationType() string {
	return AUTHTYPE_IAM
}

// IAMTokenInfo : Response struct from token request
type IAMTokenInfo struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	Expiration   int64  `json:"expiration"`
}

// IAMAuthenticator : The IAM token manager
type IAMAuthenticator struct {
	userAccessToken string
	apiKey          string
	url             string
	clientId        string
	clientSecret    string
	client          *http.Client
	timeForNewToken int64
	tokenInfo       *IAMTokenInfo
}

// NewIAMAuthenticator : Instantiate a new IAMAuthenticator.
func NewIAMAuthenticator(config *IAMConfig) (*IAMAuthenticator, error) {
	// Make sure the config is valid.
	err := config.Validate()
	if err != nil {
		return nil, err
	}

	// Use a default URL if not specified in the config.
	url := config.URL
	if url == "" {
		url = DEFAULT_IAM_URL
	}

	// Use default values for clientId and clientSecret if not specified in the config.
	clientId := config.ClientId
	clientSecret := config.ClientSecret
	if clientId == "" {
		clientId = DEFAULT_IAM_CLIENT_ID
	}
	if clientSecret == "" {
		clientSecret = DEFAULT_IAM_CLIENT_SECRET
	}

	authenticator := &IAMAuthenticator{
		userAccessToken: config.AccessToken,
		apiKey:          config.ApiKey,
		url:             url,
		clientId:        clientId,
		clientSecret:    clientSecret,
		client: &http.Client{
			Timeout: time.Second * 30,
		},
	}
	return authenticator, nil
}

func (IAMAuthenticator) AuthenticationType() string {
	return AUTHTYPE_IAM
}

// Perform the authentication by constructing the Authorization header
// value from the IAM access token (either user-supplied or obtained from the token service).
func (authenticator IAMAuthenticator) Authenticate(request *http.Request) error {
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
func (tm *IAMAuthenticator) GetToken() (string, error) {
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
//
// Deprecated: Use the IAMConfig struct to reconfigure the IAM Authenticator.
func (tm *IAMAuthenticator) SetIAMAccessToken(userAccessToken string) {
	tm.userAccessToken = userAccessToken
}

// SetIAMAPIKey : Set API key so that SDK manages token
//
// Deprecated: Use the IAMConfig struct to reconfigure the IAM Authenticator.
func (tm *IAMAuthenticator) SetIAMAPIKey(key string) {
	tm.apiKey = key
}

// Request an IAM token using an API key
func (tm *IAMAuthenticator) requestToken() (*IAMTokenInfo, error) {
	builder := NewRequestBuilder(POST).
		ConstructHTTPURL(tm.url, nil, nil)

	builder.AddHeader(CONTENT_TYPE, DEFAULT_CONTENT_TYPE).
		AddHeader(Accept, APPLICATION_JSON)

	builder.AddFormData("grant_type", "", "", REQUEST_TOKEN_GRANT_TYPE).
		AddFormData("apikey", "", "", tm.apiKey).
		AddFormData("response_type", "", "", REQUEST_TOKEN_RESPONSE_TYPE)

	req, err := builder.Build()
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(tm.clientId, tm.clientSecret)

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

	tokenInfo := &IAMTokenInfo{}
	json.NewDecoder(resp.Body).Decode(tokenInfo)
	defer resp.Body.Close()
	return tokenInfo, nil
}

func (tm *IAMAuthenticator) isTokenExpired() bool {
	if tm.timeForNewToken == 0 {
		return true
	}

	currentTime := GetCurrentTime()
	return tm.timeForNewToken < currentTime
}

// Decode and saves the access token
func (tm *IAMAuthenticator) saveToken(tokenInfo *IAMTokenInfo) {
	tm.tokenInfo = tokenInfo

	claims := jwt.StandardClaims{}
	if token, _ := jwt.ParseWithClaims(tm.tokenInfo.AccessToken, &claims, nil); token != nil {
		tm.timeForNewToken = tm.calcTimeForNewToken(claims.ExpiresAt, claims.ExpiresAt-claims.IssuedAt)
	}
}

func (tm *IAMAuthenticator) calcTimeForNewToken(expireTime int64, timeToLive int64) int64 {
	return int64(float64(expireTime) - (float64(timeToLive) * (1.0 - FRACTION_OF_TIME_TO_LIVE)))
}
