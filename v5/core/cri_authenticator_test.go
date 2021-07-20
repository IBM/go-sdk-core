// +build all slow auth

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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"testing"

	assert "github.com/stretchr/testify/assert"
)

const (
	testLogLevel          LogLevel = LevelDebug
	criMockCRTokenFile    string   = "../resources/cr-token.txt"
	criMockIAMProfileName string   = "iam-user-123"
	criMockIAMProfileID   string   = "iam-id-123"
	criTestCRToken1       string   = "cr-token-1"
	criTestAccessToken1   string   = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	criTestAccessToken2   string   = "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	criTestRefreshToken   string   = "Xj7Gle500MachEOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
)

// Struct that models the request body for the "create_access_token" operation
type instanceIdentityTokenPrototype struct {
	ExpiresIn int `json:"expires_in"`
}

func TestCriCtorErrors(t *testing.T) {
	var err error
	var auth *CriAuthenticator

	// Error: missing IAMProfileName and IBMProfileID.
	auth, err = NewCriAuthenticator("", "", "", "", "", "", "", false, "", nil)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientID.
	auth, err = NewCriAuthenticator("", "", criMockIAMProfileName, "", "", "", "client-secret", false, "", nil)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientSecret.
	auth, err = NewCriAuthenticator("", "", "", "iam-id-123", "", "client-id", "", false, "", nil)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}

func TestCriCtorSuccess(t *testing.T) {
	var err error
	var auth *CriAuthenticator
	var expectedHeaders = map[string]string{
		"header1": "value1",
	}

	// Success - only required params
	// 1. only IAMProfileName
	auth, err = NewCriAuthenticator("", "", criMockIAMProfileName, "", "", "", "", false, "", nil)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRI, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.InstanceMetadataServiceURL)
	assert.Equal(t, criMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// 2. only IAMProfileID
	auth, err = NewCriAuthenticator("", "", "", criMockIAMProfileID, "", "", "", false, "", nil)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRI, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.InstanceMetadataServiceURL)
	assert.Equal(t, "", auth.IAMProfileName)
	assert.Equal(t, criMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// Success - all parameters
	auth, err = NewCriAuthenticator("cr-token-file", "http://1.1.1.1", criMockIAMProfileName, criMockIAMProfileID,
		defaultIamTokenServerEndpoint, "client-id", "client-secret", true, "scope1", expectedHeaders)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRI, auth.AuthenticationType())
	assert.Equal(t, "cr-token-file", auth.CRTokenFilename)
	assert.Equal(t, "http://1.1.1.1", auth.InstanceMetadataServiceURL)
	assert.Equal(t, criMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, criMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.URL)
	assert.Equal(t, "client-id", auth.ClientID)
	assert.Equal(t, "client-secret", auth.ClientSecret)
	assert.Equal(t, true, auth.DisableSSLVerification)
	assert.Equal(t, expectedHeaders, auth.Headers)
}

func TestCriCtorFromMapErrors(t *testing.T) {
	var err error
	var auth *CriAuthenticator
	var configProps map[string]string

	// Error: nil config map
	auth, err = newCriAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing IAMProfileName and IAMProfileID
	configProps = map[string]string{}
	auth, err = newCriAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientID.
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_NAME: criMockIAMProfileName,
		PROPNAME_CLIENT_SECRET:    "client-secret",
	}
	auth, err = newCriAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientSecret.
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_ID: "iam-id-123",
		PROPNAME_CLIENT_ID:      "client-id",
	}
	auth, err = newCriAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}
func TestCriCtorFromMapSuccess(t *testing.T) {
	var err error
	var auth *CriAuthenticator
	var configProps map[string]string

	// Success - only required params
	// 1. only IAMProfileName
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_NAME: criMockIAMProfileName,
	}
	auth, err = newCriAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRI, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.InstanceMetadataServiceURL)
	assert.Equal(t, criMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// 2. only IAMProfileID
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_ID: criMockIAMProfileID,
	}
	auth, err = newCriAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRI, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.InstanceMetadataServiceURL)
	assert.Equal(t, "", auth.IAMProfileName)
	assert.Equal(t, criMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// Success - all params
	configProps = map[string]string{
		PROPNAME_CRTOKEN_FILENAME:              "cr-token-file",
		PROPNAME_INSTANCE_METADATA_SERVICE_URL: "http://1.1.1.1",
		PROPNAME_IAM_PROFILE_NAME:              criMockIAMProfileName,
		PROPNAME_IAM_PROFILE_ID:                "iam-id-123",
		PROPNAME_AUTH_URL:                      defaultIamTokenServerEndpoint,
		PROPNAME_CLIENT_ID:                     "client-id",
		PROPNAME_CLIENT_SECRET:                 "client-secret",
		PROPNAME_AUTH_DISABLE_SSL:              "true",
		PROPNAME_SCOPE:                         "scope1",
	}
	auth, err = newCriAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CRI, auth.AuthenticationType())
	assert.Equal(t, "cr-token-file", auth.CRTokenFilename)
	assert.Equal(t, "http://1.1.1.1", auth.InstanceMetadataServiceURL)
	assert.Equal(t, criMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, "iam-id-123", auth.IAMProfileID)
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.URL)
	assert.Equal(t, "client-id", auth.ClientID)
	assert.Equal(t, "client-secret", auth.ClientSecret)
	assert.Equal(t, true, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)
}

func TestCriReadCRTokenFromFileSuccess(t *testing.T) {
	var auth *CriAuthenticator
	var err error
	var crToken string

	// Success
	auth = &CriAuthenticator{
		CRTokenFilename: criMockCRTokenFile,
	}
	crToken, err = auth.readCRTokenFromFile()
	assert.Nil(t, err)
	assert.Equal(t, criTestCRToken1, crToken)
}

func TestCriReadCRTokenFromFileFail(t *testing.T) {
	var auth *CriAuthenticator
	var err error
	var crToken string

	// Use a non-existent cr token filename.
	auth = &CriAuthenticator{
		CRTokenFilename: "bogus-cr-token-file",
	}
	crToken, err = auth.readCRTokenFromFile()
	assert.NotNil(t, err)
	assert.Equal(t, "", crToken)
	t.Logf("Expected error: %s", err.Error())
}

// startMockServer will start a mock server endpoint that supports both the
// Instance Metadata Service and IAM operations that we'll need to call.
func startMockServer(t *testing.T) *httptest.Server {
	// Create the mock server.
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		operationPath := req.URL.EscapedPath()
		method := req.Method

		if operationPath == "/instance_identity/v1/token" {
			// If this is an invocation of the IMDS "create_access_token" operation,
			// then validate it a bit and then send back a good response.
			assert.Equal(t, "PUT", method)
			assert.Equal(t, imdsVersionDate, req.URL.Query()["version"][0])
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Accept"))
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Content-Type"))
			assert.Equal(t, imdsMetadataFlavor, req.Header.Get("Metadata-Flavor"))

			// Read and unmarshal the request body.
			requestBody := &instanceIdentityTokenPrototype{}
			_ = json.NewDecoder(req.Body).Decode(requestBody)
			defer req.Body.Close()

			assert.NotNil(t, requestBody)
			assert.Equal(t, crtokenLifetime, requestBody.ExpiresIn)

			res.WriteHeader(http.StatusOK)
			fmt.Fprintf(res, `{"access_token":"%s"}`, criTestCRToken1)
		} else if operationPath == "/identity/token" {
			// If this is an invocation of the IAM "get_token" operation,
			// then validate it a bit and then send back a good response.
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Accept"))
			assert.Equal(t, FORM_URL_ENCODED_HEADER, req.Header.Get("Content-Type"))
			assert.Equal(t, criTestCRToken1, req.FormValue("cr_token"))
			assert.Equal(t, iamGrantTypeCRToken, req.FormValue("grant_type"))

			iamProfileID := req.FormValue("profile_id")
			iamProfileName := req.FormValue("profile_name")
			assert.True(t, iamProfileName != "" || iamProfileID != "")

			// Assume that we'll return a 200 OK status code.
			statusCode := http.StatusOK

			// This is the access token we'll send back in the mock response.
			// We'll default to token 1, then see if the caller asked for token 2.
			accessToken := criTestAccessToken1

			// We'll use the scope value to control the behavior of this mock endpoint so that we can force
			// certain things to happen:
			// 1. whether to return the first or second access token.
			// 2. whether we should validate the basic-auth header.
			// 3. whether we should return a bad status code.
			// Yes, this is kinda subversive but sometimes we need to be creative on these big jobs :)
			scope := req.FormValue("scope")

			if scope == "send-second-token" {
				accessToken = criTestAccessToken2
			} else if scope == "check-basic-auth" {
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, "user1", username)
				assert.Equal(t, "password1", password)
			} else if scope == "status-bad-request" {
				statusCode = http.StatusBadRequest
			} else if scope == "status-forbidden" {
				statusCode = http.StatusForbidden
			} else if scope == "status-unauthorized" {
				statusCode = http.StatusUnauthorized
			}

			expiration := GetCurrentTime() + 3600
			res.WriteHeader(statusCode)
			switch statusCode {
			case http.StatusOK:
				fmt.Fprintf(res, `{"access_token": "%s", "token_type": "Bearer", "expires_in": 3600, "expiration": %d, "refresh_token": "%s"}`,
					accessToken, expiration, criTestRefreshToken)
			case http.StatusBadRequest:
				fmt.Fprintf(res, `Sorry, bad request!`)

			case http.StatusForbidden:
				fmt.Fprintf(res, `Sorry, you are forbidden!`)

			case http.StatusUnauthorized:
				fmt.Fprintf(res, `Sorry, you are not authorized!`)
			}
		} else {
			assert.Fail(t, "unknown operation path: "+operationPath)
		}
	}))
	return server
}

func TestCriRetrieveCRTokenFromIMDSSuccess(t *testing.T) {
	GetLogger().SetLogLevel(testLogLevel)

	server := startMockServer(t)
	defer server.Close()

	var auth *CriAuthenticator
	var err error
	var crToken string

	// Success
	auth = &CriAuthenticator{
		InstanceMetadataServiceURL: server.URL,
	}
	crToken, err = auth.retrieveCRTokenFromIMDS()
	assert.Nil(t, err)
	assert.Equal(t, criTestCRToken1, crToken)
}

func TestCriRetrieveCRTokenFromIMDSFail(t *testing.T) {
	var auth *CriAuthenticator
	var err error
	var crToken string

	auth = &CriAuthenticator{
		InstanceMetadataServiceURL: "http://bogus.imds.endpoint",
	}
	crToken, err = auth.retrieveCRTokenFromIMDS()
	assert.NotNil(t, err)
	assert.Equal(t, "", crToken)
	t.Logf("Expected error: %s", err.Error())
}

func TestCriAuthenticateFailNoCRToken(t *testing.T) {
	GetLogger().SetLogLevel(testLogLevel)

	// Set up the authenticator with both a bogus cr token filename and imds endpoint
	// so that we can't successfully retrieve a CR Token value.
	auth := &CriAuthenticator{
		CRTokenFilename:            "bogus-cr-token-file",
		InstanceMetadataServiceURL: "http://bogus.imds.endpoint",
		IAMProfileName:             criMockIAMProfileName,
		URL:                        "https://bogus.iam.endpoint",
	}
	assert.NotNil(t, auth)
	err := auth.Validate()
	assert.Nil(t, err)
	assert.Equal(t, AUTHTYPE_CRI, auth.AuthenticationType())

	// Create a new Request object to simulate an API request that needs authentication.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://myservice.localhost/api/v1", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Try to authenticate the request (should fail)
	err = auth.Authenticate(request)

	// Validate the resulting error is a valid
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	// The casted error should match the original error message
	assert.Equal(t, err.Error(), authErr.Error())
}

func TestCriGetTokenSuccess(t *testing.T) {
	var auth *CriAuthenticator
	var err error

	GetLogger().SetLogLevel(testLogLevel)

	// Start up our mock server
	server := startMockServer(t)
	defer server.Close()

	// Setup the authenticator to read the CR token from our mock file.
	auth = &CriAuthenticator{
		CRTokenFilename: criMockCRTokenFile,
		IAMProfileName:  criMockIAMProfileName,
		URL:             server.URL,
	}
	err = auth.Validate()
	assert.Nil(t, err)

	// Verify that we initially have no token data cached on the authenticator.
	assert.Nil(t, auth.getTokenData())

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, criTestAccessToken1, accessToken)
	assert.Equal(t, criTestAccessToken1, auth.getTokenData().AccessToken)

	// Force expiration and verify that we got the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 3600
	auth.Scope = "send-second-token"
	auth.IAMProfileName = ""
	auth.IAMProfileID = criMockIAMProfileID
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, criTestAccessToken2, accessToken)
	assert.Equal(t, criTestAccessToken2, auth.getTokenData().AccessToken)

	// Test the RequestToken() method to make sure we can get a RefreshToken.
	tokenResponse, err := auth.RequestToken()
	assert.Nil(t, err)
	assert.NotNil(t, tokenResponse)
	assert.Equal(t, criTestRefreshToken, tokenResponse.RefreshToken)
}

// func TestCriGetTokenSuccess(t *testing.T) {
// 	firstCall := true
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		err := r.ParseForm()
// 		assert.Nil(t, err)
// 		assert.Len(t, r.Form["apikey"], 1)
// 		assert.Len(t, r.Form["grant_type"], 1)
// 		assert.Len(t, r.Form["response_type"], 1)
// 		assert.Empty(t, r.Form["scope"])

// 		w.WriteHeader(http.StatusOK)
// 		expiration := GetCurrentTime() + 3600
// 		if firstCall {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "%s"
// 			}`, AccessToken1, expiration, RefreshToken)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "%s"
// 			}`, AccessToken2, expiration, RefreshToken)
// 			username, password, ok := r.BasicAuth()
// 			assert.True(t, ok)
// 			assert.Equal(t, "mookie", username)
// 			assert.Equal(t, "betts", password)
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())

// 	// Force the first fetch and verify we got the first access token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Force expiration and verify that we got the second access token.
// 	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
// 	authenticator.ClientId = "mookie"
// 	authenticator.ClientSecret = "betts"
// 	_, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, authenticator.getTokenData())
// 	assert.Equal(t, AccessToken2, authenticator.getTokenData().AccessToken)

// 	// Test the RequestToken() method to make sure we can get a RefreshToken.
// 	tokenResponse, err := authenticator.RequestToken()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, tokenResponse)
// 	assert.Equal(t, RefreshToken, tokenResponse.RefreshToken)
// }

// func TestCriAuthenticateFail(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		_, _ = w.Write([]byte("Sorry you are not authorized"))
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, authenticator)
// 	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())

// 	// Create a new Request object.
// 	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://localhost/placeholder/url", nil, nil)
// 	assert.Nil(t, err)

// 	request, err := builder.Build()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, request)

// 	err = authenticator.Authenticate(request)
// 	// Validate the resulting error is a valid
// 	assert.NotNil(t, err)
// 	authErr, ok := err.(*AuthenticationError)
// 	assert.True(t, ok)
// 	assert.NotNil(t, authErr)
// 	assert.EqualValues(t, authErr, err)
// 	// The casted error should match the original error message
// 	assert.Equal(t, err.Error(), authErr.Error())
// }

// func TestCriGetTokenSuccess(t *testing.T) {
// 	firstCall := true
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		err := r.ParseForm()
// 		assert.Nil(t, err)
// 		assert.Len(t, r.Form["apikey"], 1)
// 		assert.Len(t, r.Form["grant_type"], 1)
// 		assert.Len(t, r.Form["response_type"], 1)
// 		assert.Empty(t, r.Form["scope"])

// 		w.WriteHeader(http.StatusOK)
// 		expiration := GetCurrentTime() + 3600
// 		if firstCall {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "%s"
// 			}`, AccessToken1, expiration, RefreshToken)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "%s"
// 			}`, AccessToken2, expiration, RefreshToken)
// 			username, password, ok := r.BasicAuth()
// 			assert.True(t, ok)
// 			assert.Equal(t, "mookie", username)
// 			assert.Equal(t, "betts", password)
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())

// 	// Force the first fetch and verify we got the first access token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Force expiration and verify that we got the second access token.
// 	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
// 	authenticator.ClientId = "mookie"
// 	authenticator.ClientSecret = "betts"
// 	_, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, authenticator.getTokenData())
// 	assert.Equal(t, AccessToken2, authenticator.getTokenData().AccessToken)

// 	// Test the RequestToken() method to make sure we can get a RefreshToken.
// 	tokenResponse, err := authenticator.RequestToken()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, tokenResponse)
// 	assert.Equal(t, RefreshToken, tokenResponse.RefreshToken)
// }

// func TestCriGetTokenSuccessWithScope(t *testing.T) {
// 	firstCall := true
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		err := r.ParseForm()
// 		assert.Nil(t, err)
// 		assert.Len(t, r.Form["apikey"], 1)
// 		assert.Len(t, r.Form["grant_type"], 1)
// 		assert.Len(t, r.Form["response_type"], 1)
// 		assert.Len(t, r.Form["scope"], 1)
// 		assert.Equal(t, "scope1 scope2", r.Form["scope"][0])

// 		w.WriteHeader(http.StatusOK)
// 		expiration := GetCurrentTime() + 3600
// 		if firstCall {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken1, expiration)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken2, expiration)
// 			username, password, ok := r.BasicAuth()
// 			assert.True(t, ok)
// 			assert.Equal(t, "mookie", username)
// 			assert.Equal(t, "betts", password)
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())
// 	authenticator.Scope = "scope1 scope2"

// 	// Force the first fetch and verify we got the first access token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Force expiration and verify that we got the second access token.
// 	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
// 	authenticator.ClientId = "mookie"
// 	authenticator.ClientSecret = "betts"
// 	_, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.NotNil(t, authenticator.getTokenData())
// 	assert.Equal(t, AccessToken2, authenticator.getTokenData().AccessToken)
// }
// func TestCriGetCachedToken(t *testing.T) {
// 	firstCall := true
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		expiration := GetCurrentTime() + 3600
// 		if firstCall {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken1, expiration)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken2, expiration)
// 			_, _, ok := r.BasicAuth()
// 			assert.True(t, ok)
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())

// 	// Force the first fetch and verify we got the first access token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Subsequent fetch should still return first access token.
// 	token, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())
// }

// func TestCriBackgroundTokenRefresh(t *testing.T) {
// 	firstCall := true
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		expiration := GetCurrentTime() + 3600
// 		w.WriteHeader(http.StatusOK)
// 		if firstCall {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken1, expiration)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken2, expiration)
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())

// 	// Force the first fetch and verify we got the first access token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
// 	authenticator.getTokenData().RefreshTime = GetCurrentTime() - 720

// 	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
// 	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
// 	// expired.
// 	token, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Wait for the background thread to finish
// 	time.Sleep(5 * time.Second)
// 	token, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken2, token)
// 	assert.NotNil(t, authenticator.getTokenData())
// }

// func TestCriBackgroundTokenRefreshFailure(t *testing.T) {
// 	firstCall := true
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		expiration := GetCurrentTime() + 3600
// 		w.WriteHeader(http.StatusOK)
// 		if firstCall {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken1, expiration)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			_, _ = w.Write([]byte("Sorry you are forbidden"))
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())

// 	// Successfully fetch the first token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
// 	authenticator.getTokenData().RefreshTime = GetCurrentTime() - 720
// 	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
// 	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
// 	// expired.
// 	token, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())
// 	// Wait for the background thread to finish
// 	time.Sleep(5 * time.Second)
// 	_, err = authenticator.GetToken()
// 	assert.NotNil(t, err)
// 	assert.Equal(t, "Error while trying to get access token", err.Error())
// 	// We don't expect a AuthenticateError to be returned, so casting should fail
// 	_, ok := err.(*AuthenticationError)
// 	assert.False(t, ok)

// }

// func TestCriBackgroundTokenRefreshIdle(t *testing.T) {
// 	firstCall := true
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// CurrentTime + 1 hour
// 		expiration := GetCurrentTime() + 3600
// 		w.WriteHeader(http.StatusOK)
// 		if firstCall {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken1, expiration)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken2, expiration)
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())

// 	// Force the first fetch and verify we got the first access token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Now simulate the client being idle for 10 minutes into the refresh time
// 	tenMinutesBeforeNow := GetCurrentTime() - 600
// 	authenticator.getTokenData().RefreshTime = tenMinutesBeforeNow

// 	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
// 	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
// 	// expired.
// 	token, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// RefreshTime should have advanced by 1 minute from the current time
// 	newRefreshTime := GetCurrentTime() + 60
// 	assert.Equal(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

// 	// In the next request, the RefreshTime should be unchanged and another thread
// 	// shouldn't be spawned to request another token once more since the first thread already spawned
// 	// a goroutine & refreshed the token.
// 	token, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)

// 	assert.NotNil(t, authenticator.getTokenData())
// 	assert.Equal(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

// 	// Wait for the background thread to finish and verify both the RefreshTime & tokenData were updated
// 	time.Sleep(5 * time.Second)
// 	token, err = authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken2, token)
// 	assert.NotNil(t, authenticator.getTokenData())
// 	assert.NotEqual(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

// }

// func TestCriClientIdAndSecret(t *testing.T) {
// 	expiration := GetCurrentTime() + 3600
// 	accessToken := "oAeisG8yqPY7sFR_x66Z15"
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, `{
// 			"access_token": "%s",
// 			"token_type": "Bearer",
// 			"expires_in": 3600,
// 			"expiration": %d,
// 			"refresh_token": "jy4gl91BQ"
// 		}`, accessToken, expiration)
// 		username, password, ok := r.BasicAuth()
// 		assert.True(t, ok)
// 		assert.Equal(t, "mookie", username)
// 		assert.Equal(t, "betts", password)
// 	}))
// 	defer server.Close()

// 	authenticator := &IamAuthenticator{
// 		ApiKey:       "bogus-apikey",
// 		URL:          server.URL,
// 		ClientId:     "mookie",
// 		ClientSecret: "betts",
// 	}

// 	token, err := authenticator.GetToken()
// 	assert.Equal(t, accessToken, token)
// 	assert.Nil(t, err)
// }

// func TestCriRefreshTimeCalculation(t *testing.T) {
// 	const timeToLive int64 = 3600
// 	const expireTime int64 = 1563911183
// 	const expected int64 = expireTime - 720 // 720 is 20% of 3600

// 	// Simulate a token server response.
// 	tokenResponse := &IamTokenServerResponse{
// 		ExpiresIn:  timeToLive,
// 		Expiration: expireTime,
// 	}

// 	// Create a new token data and verify the refresh time.
// 	tokenData, err := newIamTokenData(tokenResponse)
// 	assert.Nil(t, err)
// 	assert.Equal(t, expected, tokenData.RefreshTime)
// }

// func TestCriDisableSSL(t *testing.T) {
// 	expiration := GetCurrentTime() + 3600
// 	accessToken := "oAeisG8yqPY7sFR_x66Z15"
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, `{
// 			"access_token": "%s",
// 			"token_type": "Bearer",
// 			"expires_in": 3600,
// 			"expiration": %d,
// 			"refresh_token": "jy4gl91BQ"
// 		}`, accessToken, expiration)
// 		_, _, ok := r.BasicAuth()
// 		assert.False(t, ok)
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", true, nil)
// 	assert.Nil(t, err)

// 	token, err := authenticator.GetToken()
// 	assert.Equal(t, accessToken, token)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, authenticator.Client)
// 	assert.NotNil(t, authenticator.Client.Transport)
// 	transport, ok := authenticator.Client.Transport.(*http.Transport)
// 	assert.True(t, ok)
// 	assert.NotNil(t, transport.TLSClientConfig)
// 	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
// }

// func TestCriUserHeaders(t *testing.T) {
// 	expiration := GetCurrentTime() + 3600
// 	accessToken := "oAeisG8yqPY7sFR_x66Z15"
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, `{
// 			"access_token": "%s",
// 			"token_type": "Bearer",
// 			"expires_in": 3600,
// 			"expiration": %d,
// 			"refresh_token": "jy4gl91BQ"
// 		}`, accessToken, expiration)
// 		_, _, ok := r.BasicAuth()
// 		assert.False(t, ok)
// 		assert.Equal(t, "Value1", r.Header.Get("Header1"))
// 		assert.Equal(t, "Value2", r.Header.Get("Header2"))
// 		assert.Equal(t, "iam.cloud.ibm.com", r.Host)
// 	}))
// 	defer server.Close()

// 	headers := make(map[string]string)
// 	headers["Header1"] = "Value1"
// 	headers["Header2"] = "Value2"
// 	headers["Host"] = "iam.cloud.ibm.com"

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, headers)
// 	assert.Nil(t, err)

// 	token, err := authenticator.GetToken()
// 	assert.Equal(t, accessToken, token)
// 	assert.Nil(t, err)
// }

// func TestCriGetTokenFailure(t *testing.T) {
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusForbidden)
// 		_, _ = w.Write([]byte("Sorry you are forbidden"))
// 	}))
// 	defer server.Close()

// 	authenticator := &IamAuthenticator{
// 		ApiKey: "bogus-apikey",
// 		URL:    server.URL,
// 	}

// 	var expectedResponse = []byte("Sorry you are forbidden")

// 	_, err := authenticator.GetToken()
// 	assert.NotNil(t, err)
// 	assert.Equal(t, "Sorry you are forbidden", err.Error())
// 	// We expect an AuthenticationError to be returned, so cast the returned error
// 	authError, ok := err.(*AuthenticationError)
// 	assert.True(t, ok)
// 	assert.NotNil(t, authError)
// 	assert.NotNil(t, authError.Error())
// 	assert.NotNil(t, authError.Response)
// 	rawResult := authError.Response.GetRawResult()
// 	assert.NotNil(t, rawResult)
// 	assert.Equal(t, expectedResponse, rawResult)
// 	statusCode := authError.Response.GetStatusCode()
// 	assert.Equal(t, "Sorry you are forbidden", authError.Error())
// 	assert.Equal(t, http.StatusForbidden, statusCode)
// }

// func TestCriGetTokenTimeoutError(t *testing.T) {
// 	firstCall := true
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.WriteHeader(http.StatusOK)
// 		expiration := GetCurrentTime() + 3600
// 		if firstCall {
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken1, expiration)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			time.Sleep(3 * time.Second)
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken2, expiration)
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())

// 	// Force the first fetch and verify we got the first access token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	// Force expiration and verify that we got a timeout error
// 	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600

// 	// Set the client timeout to something very low
// 	authenticator.Client.Timeout = time.Second * 2
// 	token, err = authenticator.GetToken()
// 	assert.Empty(t, token)
// 	assert.NotNil(t, err)
// 	assert.NotNil(t, err.Error())
// 	// We don't expect a AuthenticateError to be returned, so casting should fail
// 	_, ok := err.(*AuthenticationError)
// 	assert.False(t, ok)
// }

// func TestCriGetTokenServerError(t *testing.T) {
// 	firstCall := true
// 	expiration := GetCurrentTime() + 3600
// 	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		if firstCall {
// 			w.WriteHeader(http.StatusOK)
// 			fmt.Fprintf(w, `{
// 				"access_token": "%s",
// 				"token_type": "Bearer",
// 				"expires_in": 3600,
// 				"expiration": %d,
// 				"refresh_token": "jy4gl91BQ"
// 			}`, AccessToken1, expiration)
// 			firstCall = false
// 			_, _, ok := r.BasicAuth()
// 			assert.False(t, ok)
// 		} else {
// 			w.WriteHeader(http.StatusGatewayTimeout)
// 			_, _ = w.Write([]byte("Gateway Timeout"))
// 		}
// 	}))
// 	defer server.Close()

// 	authenticator, err := NewIamAuthenticator("bogus-apikey", server.URL, "", "", false, nil)
// 	assert.Nil(t, err)
// 	assert.Nil(t, authenticator.getTokenData())

// 	// Force the first fetch and verify we got the first access token.
// 	token, err := authenticator.GetToken()
// 	assert.Nil(t, err)
// 	assert.Equal(t, AccessToken1, token)
// 	assert.NotNil(t, authenticator.getTokenData())

// 	var expectedResponse = []byte("Gateway Timeout")

// 	// Force expiration and verify that we got a server error
// 	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
// 	token, err = authenticator.GetToken()
// 	assert.NotNil(t, err)
// 	// We expect an AuthenticationError to be returned, so cast the returned error
// 	authError, ok := err.(*AuthenticationError)
// 	assert.True(t, ok)
// 	assert.NotNil(t, authError)
// 	assert.NotNil(t, authError.Response)
// 	assert.NotNil(t, authError.Error())
// 	rawResult := authError.Response.GetRawResult()
// 	statusCode := authError.Response.GetStatusCode()
// 	assert.Equal(t, "Gateway Timeout", authError.Error())
// 	assert.Equal(t, expectedResponse, rawResult)
// 	assert.NotNil(t, rawResult)
// 	assert.Equal(t, http.StatusGatewayTimeout, statusCode)
// 	assert.Empty(t, token)
// }

// func TestNewIamAuthenticatorFromMap(t *testing.T) {
// 	_, err := newIamAuthenticatorFromMap(nil)
// 	assert.NotNil(t, err)

// 	var props = map[string]string{
// 		PROPNAME_AUTH_URL: "iam-url",
// 	}
// 	_, err = newIamAuthenticatorFromMap(props)
// 	assert.NotNil(t, err)

// 	props = map[string]string{
// 		PROPNAME_APIKEY: "",
// 	}
// 	_, err = newIamAuthenticatorFromMap(props)
// 	assert.NotNil(t, err)

// 	props = map[string]string{
// 		PROPNAME_APIKEY: "my-apikey",
// 	}
// 	authenticator, err := newIamAuthenticatorFromMap(props)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, authenticator)
// 	assert.Equal(t, "my-apikey", authenticator.ApiKey)

// 	props = map[string]string{
// 		PROPNAME_APIKEY:           "my-apikey",
// 		PROPNAME_AUTH_DISABLE_SSL: "huh???",
// 		PROPNAME_CLIENT_ID:        "mookie",
// 		PROPNAME_CLIENT_SECRET:    "betts",
// 		PROPNAME_SCOPE:            "scope1 scope2",
// 	}
// 	authenticator, err = newIamAuthenticatorFromMap(props)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, authenticator)
// 	assert.Equal(t, "my-apikey", authenticator.ApiKey)
// 	assert.False(t, authenticator.DisableSSLVerification)
// 	assert.Equal(t, "mookie", authenticator.ClientId)
// 	assert.Equal(t, "betts", authenticator.ClientSecret)
// 	assert.Equal(t, "scope1 scope2", authenticator.Scope)
// }
