//go:build all || auth
// +build all auth

package core

// (C) Copyright IBM Corp. 2021.
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
	"fmt"
	"net/http"
	"net/http/httptest"

	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

const (
	// To enable debug logging during test execution, set this to "LevelDebug"
	containerAuthTestLogLevel       LogLevel = LevelError
	containerAuthMockCRTokenFile    string   = "../resources/cr-token.txt"
	containerAuthEmptyCRTokenFile   string   = "../resources/empty-cr-token.txt"
	containerAuthMockIAMProfileName string   = "iam-user-123"
	containerAuthMockIAMProfileID   string   = "iam-id-123"
	containerAuthMockClientID       string   = "client-id-1"
	containerAuthMockClientSecret   string   = "client-secret-1"
	containerAuthTestCRToken1       string   = "cr-token-1"
	containerAuthTestAccessToken1   string   = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	containerAuthTestAccessToken2   string   = "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	containerAuthTestRefreshToken   string   = "Xj7Gle500MachEOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
)

func TestContainerAuthCtorErrors(t *testing.T) {
	var err error
	var auth *ContainerAuthenticator

	// Error: missing IAMProfileName and IBMProfileID.
	auth, err = NewContainerAuthenticatorBuilder().Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientID.
	auth, err = NewContainerAuthenticatorBuilder().
		SetIAMProfileName(containerAuthMockIAMProfileName).
		SetClientIDSecret("", containerAuthMockClientSecret).
		SetClient(nil).
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientSecret.
	auth, err = NewContainerAuthenticatorBuilder().
		SetIAMProfileID(containerAuthMockIAMProfileID).
		SetClientIDSecret(containerAuthMockClientID, "").
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}

func TestContainerAuthCtorSuccess(t *testing.T) {
	var err error
	var auth *ContainerAuthenticator
	var expectedHeaders = map[string]string{
		"header1": "value1",
	}

	// Success - only required params
	// 1. only IAMProfileName
	auth, err = NewContainerAuthenticatorBuilder().
		SetIAMProfileName(containerAuthMockIAMProfileName).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CONTAINER, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, containerAuthMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// 2. only IAMProfileID
	auth, err = NewContainerAuthenticatorBuilder().
		SetIAMProfileID(containerAuthMockIAMProfileID).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CONTAINER, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.IAMProfileName)
	assert.Equal(t, containerAuthMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// Success - all parameters
	auth, err = NewContainerAuthenticatorBuilder().
		SetCRTokenFilename("cr-token-file").
		SetIAMProfileName(containerAuthMockIAMProfileName).
		SetIAMProfileID(containerAuthMockIAMProfileID).
		SetURL(defaultIamTokenServerEndpoint).
		SetClientIDSecret(containerAuthMockClientID, containerAuthMockClientSecret).
		SetDisableSSLVerification(true).
		SetScope("scope1").
		SetHeaders(expectedHeaders).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CONTAINER, auth.AuthenticationType())
	assert.Equal(t, "cr-token-file", auth.CRTokenFilename)
	assert.Equal(t, containerAuthMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, containerAuthMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.URL)
	assert.Equal(t, containerAuthMockClientID, auth.ClientID)
	assert.Equal(t, containerAuthMockClientSecret, auth.ClientSecret)
	assert.Equal(t, true, auth.DisableSSLVerification)
	assert.Equal(t, expectedHeaders, auth.Headers)
}

func TestContainerAuthCtorFromMapErrors(t *testing.T) {
	var err error
	var auth *ContainerAuthenticator
	var configProps map[string]string

	// Error: nil config map
	auth, err = newContainerAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing IAMProfileName and IAMProfileID
	configProps = map[string]string{}
	auth, err = newContainerAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientID.
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_NAME: containerAuthMockIAMProfileName,
		PROPNAME_CLIENT_SECRET:    containerAuthMockClientSecret,
	}
	auth, err = newContainerAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: missing ClientSecret.
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_ID: containerAuthMockIAMProfileID,
		PROPNAME_CLIENT_ID:      containerAuthMockClientID,
	}
	auth, err = newContainerAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}
func TestContainerAuthCtorFromMapSuccess(t *testing.T) {
	var err error
	var auth *ContainerAuthenticator
	var configProps map[string]string

	// Success - only required params
	// 1. only IAMProfileName
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_NAME: containerAuthMockIAMProfileName,
	}
	auth, err = newContainerAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CONTAINER, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, containerAuthMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// 2. only IAMProfileID
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_ID: containerAuthMockIAMProfileID,
	}
	auth, err = newContainerAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CONTAINER, auth.AuthenticationType())
	assert.Equal(t, "", auth.CRTokenFilename)
	assert.Equal(t, "", auth.IAMProfileName)
	assert.Equal(t, containerAuthMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)
	assert.Equal(t, "", auth.ClientID)
	assert.Equal(t, "", auth.ClientSecret)
	assert.Equal(t, false, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)

	// Success - all params
	configProps = map[string]string{
		PROPNAME_CRTOKEN_FILENAME: containerAuthMockCRTokenFile,
		PROPNAME_IAM_PROFILE_NAME: containerAuthMockIAMProfileName,
		PROPNAME_IAM_PROFILE_ID:   containerAuthMockIAMProfileID,
		PROPNAME_AUTH_URL:         defaultIamTokenServerEndpoint,
		PROPNAME_CLIENT_ID:        containerAuthMockClientID,
		PROPNAME_CLIENT_SECRET:    containerAuthMockClientSecret,
		PROPNAME_AUTH_DISABLE_SSL: "true",
		PROPNAME_SCOPE:            "scope1",
	}
	auth, err = newContainerAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_CONTAINER, auth.AuthenticationType())
	assert.Equal(t, containerAuthMockCRTokenFile, auth.CRTokenFilename)
	assert.Equal(t, containerAuthMockIAMProfileName, auth.IAMProfileName)
	assert.Equal(t, containerAuthMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.URL)
	assert.Equal(t, containerAuthMockClientID, auth.ClientID)
	assert.Equal(t, containerAuthMockClientSecret, auth.ClientSecret)
	assert.Equal(t, true, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)
}
func TestContainerAuthDefaultURL(t *testing.T) {
	auth := &ContainerAuthenticator{}
	s := auth.url()
	assert.Equal(t, s, defaultIamTokenServerEndpoint)
	assert.Equal(t, auth.URL, defaultIamTokenServerEndpoint)
}

// startMockIAMServer will start a mock server endpoint that supports both the
// Instance Metadata Service and IAM operations that we'll need to call.
func startMockIAMServer(t *testing.T) *httptest.Server {
	// Create the mock server.
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		operationPath := req.URL.EscapedPath()

		if operationPath == "/identity/token" {
			// If this is an invocation of the IAM "get_token" operation,
			// then validate it a bit and then send back a good response.
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Accept"))
			assert.Equal(t, FORM_URL_ENCODED_HEADER, req.Header.Get("Content-Type"))
			assert.Equal(t, containerAuthTestCRToken1, req.FormValue("cr_token"))
			assert.Equal(t, iamGrantTypeCRToken, req.FormValue("grant_type"))

			iamProfileID := req.FormValue("profile_id")
			iamProfileName := req.FormValue("profile_name")
			assert.True(t, iamProfileName != "" || iamProfileID != "")

			// Assume that we'll return a 200 OK status code.
			statusCode := http.StatusOK

			// This is the access token we'll send back in the mock response.
			// We'll default to token 1, then see if the caller asked for token 2
			// via the scope setting below.
			accessToken := containerAuthTestAccessToken1

			// We'll use the scope value to control the behavior of this mock endpoint so that we can force
			// certain things to happen:
			// 1. whether to return the first or second access token.
			// 2. whether we should validate the basic-auth header.
			// 3. whether we should return a bad status code.
			// Yes, this is kinda subversive, but sometimes we need to be creative on these big jobs :)
			scope := req.FormValue("scope")

			if scope == "send-second-token" {
				accessToken = containerAuthTestAccessToken2
			} else if scope == "check-basic-auth" {
				username, password, ok := req.BasicAuth()
				assert.True(t, ok)
				assert.Equal(t, containerAuthMockClientID, username)
				assert.Equal(t, containerAuthMockClientSecret, password)
			} else if scope == "check-user-headers" {
				assert.Equal(t, "Value-1", req.Header.Get("User-Header-1"))
				assert.Equal(t, "iam.cloud.ibm.com", req.Host)
			} else if scope == "status-bad-request" {
				statusCode = http.StatusBadRequest
			} else if scope == "status-unauthorized" {
				statusCode = http.StatusUnauthorized
			} else if scope == "sleep" {
				time.Sleep(3 * time.Second)
			}

			expiration := GetCurrentTime() + 3600
			res.WriteHeader(statusCode)
			switch statusCode {
			case http.StatusOK:
				fmt.Fprintf(res, `{"access_token": "%s", "token_type": "Bearer", "expires_in": 3600, "expiration": %d, "refresh_token": "%s"}`,
					accessToken, expiration, containerAuthTestRefreshToken)
			case http.StatusBadRequest:
				fmt.Fprintf(res, `Sorry, bad request!`)

			case http.StatusUnauthorized:
				fmt.Fprintf(res, `Sorry, you are not authorized!`)
			}
		} else {
			assert.Fail(t, "unknown operation path: "+operationPath)
		}
	}))
	return server
}

func TestContainerAuthRetrieveCRTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	// Set the authenticator to read the CR token from our mock file.
	auth := &ContainerAuthenticator{
		CRTokenFilename: containerAuthMockCRTokenFile,
	}
	crToken, err := auth.retrieveCRToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestCRToken1, crToken)
}

func TestContainerAuthRetrieveCRTokenFail(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	// Use a non-existent cr token file.
	auth := &ContainerAuthenticator{
		CRTokenFilename: "bogus-cr-token-file",
	}
	crToken, err := auth.retrieveCRToken()
	assert.NotNil(t, err)
	assert.Equal(t, "", crToken)
	t.Logf("Expected error: %s", err.Error())
}

func TestContainerAuthGetTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	auth := &ContainerAuthenticator{
		CRTokenFilename: containerAuthMockCRTokenFile,
		IAMProfileName:  containerAuthMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Verify that we initially have no token data cached on the authenticator.
	assert.Nil(t, auth.getTokenData())

	// Force the first fetch and verify we got the first access token.
	var accessToken string
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)

	// Verify that the access token was returned by GetToken() and also
	// stored in the authenticator's tokenData field as well.
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)
	assert.Equal(t, containerAuthTestAccessToken1, auth.getTokenData().AccessToken)

	// We should also get back a nil error from synchronizedRequestToken()
	assert.Nil(t, auth.synchronizedRequestToken())

	// Call GetToken() again and verify that we get the cached value.
	// Note: we'll Set Scope so that if the IAM operation is actually called again,
	// we'll receive the second access token.  We don't want the IAM operation called again yet.
	auth.Scope = "send-second-token"
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)

	// Force expiration and verify that GetToken() fetched the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	auth.IAMProfileName = ""
	auth.IAMProfileID = containerAuthMockIAMProfileID
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, containerAuthTestAccessToken2, accessToken)
	assert.Equal(t, containerAuthTestAccessToken2, auth.getTokenData().AccessToken)
}

func TestContainerAuthRequestTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	auth := &ContainerAuthenticator{
		CRTokenFilename: containerAuthMockCRTokenFile,
		IAMProfileName:  containerAuthMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Verify that RequestToken() returns a response with a valid refresh token.
	tokenResponse, err := auth.RequestToken()
	assert.Nil(t, err)
	assert.NotNil(t, tokenResponse)
	assert.Equal(t, containerAuthTestRefreshToken, tokenResponse.RefreshToken)
}

func TestContainerAuthRequestTokenError1(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	// Force an error while resolving the service URL.
	auth := &ContainerAuthenticator{
		CRTokenFilename: containerAuthMockCRTokenFile,
		IAMProfileName:  containerAuthMockIAMProfileName,
		URL:             "123:badpath",
	}

	iamToken, err := auth.RequestToken()
	assert.NotNil(t, err)
	assert.Nil(t, iamToken)
	t.Logf("Expected error: %s\n", err.Error())
}

func TestContainerAuthRequestTokenError2(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	// Force an error due to an empty CR token.
	auth := &ContainerAuthenticator{
		CRTokenFilename: containerAuthEmptyCRTokenFile,
		IAMProfileName:  containerAuthMockIAMProfileName,
	}

	iamToken, err := auth.RequestToken()
	assert.NotNil(t, err)
	assert.Nil(t, iamToken)
	t.Logf("Expected error: %s\n", err.Error())
}

func TestContainerAuthAuthenticateSuccess(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	// Set up the authenticator to use the cr token file.
	auth, err := NewContainerAuthenticatorBuilder().
		SetCRTokenFilename(containerAuthMockCRTokenFile).
		SetIAMProfileID(containerAuthMockIAMProfileID).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Create a new Request object to simulate an API request that needs authentication.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://myservice.localhost/api/v1", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Try to authenticate the request.
	err = auth.Authenticate(request)

	// Verify that it succeeded.
	assert.Nil(t, err)
	authHeader := request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+containerAuthTestAccessToken1, authHeader)

	// Call Authenticate again to make sure we used the cached access token.
	// We'll do this by setting scope to request the second token,
	// we'll expect the first token to be returned which verifies that we didn't
	// call the IAM "get token" operation again.
	auth.Scope = "send-second-token"
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader = request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+containerAuthTestAccessToken1, authHeader)

	// Force expiration and verify that Authenticate() fetched the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader = request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+containerAuthTestAccessToken2, authHeader)
}

func TestContainerAuthAuthenticateFailNoCRToken(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	// Set up the authenticator with a bogus cr token filename
	// so that we can't successfully retrieve a CR Token value.
	auth, err := NewContainerAuthenticatorBuilder().
		SetCRTokenFilename("bogus-cr-token-file").
		SetIAMProfileName(containerAuthMockIAMProfileName).
		SetURL("https://bogus.iam.endpoint").
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Create a new Request object to simulate an API request that needs authentication.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://myservice.localhost/api/v1", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Try to authenticate the request (should fail)
	err = auth.Authenticate(request)

	// Validate the resulting error is a valid AuthenticationError.
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	// The auth error should match the original error message
	assert.Equal(t, err.Error(), authErr.Error())
}

func TestContainerAuthAuthenticateFailIAM(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	// Set up the authenticator to use our mock cr token file,
	// and set scope to cause the mock IAM server to send a bad status code for the IAM call.
	auth := &ContainerAuthenticator{
		CRTokenFilename: containerAuthMockCRTokenFile,
		IAMProfileName:  containerAuthMockIAMProfileName,
		URL:             server.URL,
		Scope:           "status-bad-request",
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Create a new Request object to simulate an API request that needs authentication.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://myservice.localhost/api/v1", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Try to authenticate the request (should fail)
	err = auth.Authenticate(request)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	// The casted error should match the original error message
	assert.Contains(t, authErr.Error(), "Sorry, bad request!")
	assert.Equal(t, http.StatusBadRequest, authErr.Response.StatusCode)
}

func TestContainerAuthBackgroundTokenRefreshSuccess(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	auth := &ContainerAuthenticator{
		CRTokenFilename: containerAuthMockCRTokenFile,
		IAMProfileName:  containerAuthMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)

	// Now simulate being in the refresh window where the token is not expired but still needs to be refreshed.
	auth.getTokenData().RefreshTime = GetCurrentTime() - 1

	// Authenticator should detect the need to get a new access token in the background but use the current
	// cached access token for this next GetToken() call.
	// Set "scope" to cause the mock server to return the second access token the next time
	// we call the IAM "get token" operation.
	auth.Scope = "send-second-token"
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)

	// Wait for the background thread to finish.
	time.Sleep(2 * time.Second)
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken2, accessToken)
}

func TestContainerAuthBackgroundTokenRefreshFail(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	auth := &ContainerAuthenticator{
		CRTokenFilename: containerAuthMockCRTokenFile,
		IAMProfileName:  containerAuthMockIAMProfileName,
		URL:             server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)

	// Now simulate being in the refresh window where the token is not expired but still needs to be refreshed.
	auth.getTokenData().RefreshTime = GetCurrentTime() - 1

	// Authenticator should detect the need to get a new access token in the background but use the current
	// cached access token for this next GetToken() call.
	// Set "scope" to cause the mock server to return an error the next time the IAM "get token" operation is called.
	auth.Scope = "status-unauthorized"
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)

	// Wait for the background thread to finish.
	time.Sleep(2 * time.Second)

	// The background token refresh triggered by the previous GetToken() call above failed,
	// but the authenticator is still holding a valid, unexpired access token,
	// so this next GetToken() call should succeed and return the first access token
	// that we had previously cached.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)

	// Next, simulate the expiration of the token, then we should expect
	// an error from GetToken().
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	accessToken, err = auth.GetToken()
	assert.NotNil(t, err)
	assert.Empty(t, accessToken)
	t.Logf("Expected error: %s\n", err.Error())
}

func TestContainerAuthClientIdAndSecret(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	auth, err := NewContainerAuthenticatorBuilder().
		SetCRTokenFilename(containerAuthMockCRTokenFile).
		SetIAMProfileName(containerAuthMockIAMProfileName).
		SetClientIDSecret(containerAuthMockClientID, containerAuthMockClientSecret).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Force the first fetch and verify we got the first access token.
	auth.Scope = "check-basic-auth"
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)
}

func TestContainerAuthDisableSSL(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	auth, err := NewContainerAuthenticatorBuilder().
		SetCRTokenFilename(containerAuthMockCRTokenFile).
		SetIAMProfileName(containerAuthMockIAMProfileName).
		SetURL(server.URL).
		SetDisableSSLVerification(true).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)

	// Next, verify that the authenticator's Client is configured correctly.
	assert.NotNil(t, auth.Client)
	assert.NotNil(t, auth.Client.Transport)
	transport, ok := auth.Client.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestContainerAuthUserHeaders(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	headers := make(map[string]string)
	headers["User-Header-1"] = "Value-1"
	headers["Host"] = "iam.cloud.ibm.com"

	auth, err := NewContainerAuthenticatorBuilder().
		SetCRTokenFilename(containerAuthMockCRTokenFile).
		SetIAMProfileName(containerAuthMockIAMProfileName).
		SetURL(server.URL).
		SetHeaders(headers).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Force the first fetch and verify we got the first access token.
	auth.Scope = "check-user-headers"
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)
}

func TestContainerAuthGetTokenTimeout(t *testing.T) {
	GetLogger().SetLogLevel(containerAuthTestLogLevel)

	server := startMockIAMServer(t)
	defer server.Close()

	auth, err := NewContainerAuthenticatorBuilder().
		SetCRTokenFilename(containerAuthMockCRTokenFile).
		SetIAMProfileName(containerAuthMockIAMProfileName).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, containerAuthTestAccessToken1, accessToken)

	// Next, tell the mock server to sleep for a bit, force the expiration of the token,
	// and configure the client with a short timeout.
	auth.Scope = "sleep"
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	auth.Client.Timeout = time.Second * 2
	accessToken, err = auth.GetToken()
	assert.Empty(t, accessToken)
	assert.NotNil(t, err)
	assert.NotNil(t, err.Error())
	t.Logf("Expected error: %s\n", err.Error())
	_, ok := err.(*AuthenticationError)
	assert.True(t, ok)
}
