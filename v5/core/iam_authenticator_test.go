//go:build all || slow || auth
// +build all slow auth

package core

// (C) Copyright IBM Corp. 2019, 2021.
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
	"os"
	"strings"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

var (
	// To enable debug logging during test execution, set this to "LevelDebug"
	iamAuthTestLogLevel     LogLevel = LevelError
	iamAuthMockApiKey                = "mock-apikey"
	iamAuthMockRefreshToken          = "mock-refresh-token"
	iamAuthMockClientID              = "bx"
	iamAuthMockClientSecret          = "bx"
	iamAuthMockURL                   = "https://mock.iam.com"
	iamAuthMockScope                 = "scope1,scope2"

	iamAuthTestAccessToken1 string = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	iamAuthTestAccessToken2 string = "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	iamAuthTestRefreshToken string = "Xj7Gle500MachEOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
)

// Tests involving the Builder
func TestIamAuthBuilderErrors(t *testing.T) {
	var err error
	var auth *IamAuthenticator

	// Error: no apikey or refresh token
	auth, err = NewIamAuthenticatorBuilder().
		SetApiKey("").
		SetRefreshToken("").
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: invalid apikey
	auth, err = NewIamAuthenticatorBuilder().
		SetApiKey("{invalid-apikey}").
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: apikey and client-id set, but no client-secret
	auth, err = NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetClientIDSecret(iamAuthMockClientID, "").
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: apikey and client-secret set, but no client-id
	auth, err = NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetClientIDSecret("", iamAuthMockClientSecret).
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

}

func TestIamAuthBuilderSuccess(t *testing.T) {
	var err error
	var auth *IamAuthenticator
	var expectedHeaders = map[string]string{
		"header1": "value1",
	}

	// Specify apikey.
	auth, err = NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetClient(nil).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, iamAuthMockApiKey, auth.ApiKey)
	assert.Empty(t, auth.RefreshToken)
	assert.Empty(t, auth.URL)
	assert.Empty(t, auth.ClientId)
	assert.Empty(t, auth.ClientSecret)
	assert.False(t, auth.DisableSSLVerification)
	assert.Empty(t, auth.Scope)
	assert.Nil(t, auth.Headers)
	assert.Equal(t, AUTHTYPE_IAM, auth.AuthenticationType())

	// Specify refresh token along with client id/secret.
	auth, err = NewIamAuthenticatorBuilder().
		SetRefreshToken(iamAuthMockRefreshToken).
		SetClientIDSecret(iamAuthMockClientID, iamAuthMockClientSecret).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Empty(t, auth.ApiKey)
	assert.Empty(t, auth.URL)
	assert.Equal(t, iamAuthMockRefreshToken, auth.RefreshToken)
	assert.Equal(t, iamAuthMockClientID, auth.ClientId)
	assert.Equal(t, iamAuthMockClientSecret, auth.ClientSecret)
	assert.False(t, auth.DisableSSLVerification)
	assert.Empty(t, auth.Scope)
	assert.Nil(t, auth.Headers)
	assert.Equal(t, AUTHTYPE_IAM, auth.AuthenticationType())

	// Success: specify both apikey and refresh token
	auth, err = NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetRefreshToken(iamAuthMockRefreshToken).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Specify apikey with other properties.
	auth, err = NewIamAuthenticatorBuilder().
		SetURL(iamAuthMockURL).
		SetApiKey(iamAuthMockApiKey).
		SetClientIDSecret(iamAuthMockClientID, iamAuthMockClientSecret).
		SetDisableSSLVerification(true).
		SetScope(iamAuthMockScope).
		SetHeaders(expectedHeaders).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, iamAuthMockApiKey, auth.ApiKey)
	assert.Empty(t, auth.RefreshToken)
	assert.Equal(t, iamAuthMockURL, auth.URL)
	assert.Equal(t, iamAuthMockClientID, auth.ClientId)
	assert.Equal(t, iamAuthMockClientSecret, auth.ClientSecret)
	assert.True(t, auth.DisableSSLVerification)
	assert.Equal(t, iamAuthMockScope, auth.Scope)
	assert.Equal(t, expectedHeaders, auth.Headers)
	assert.Equal(t, AUTHTYPE_IAM, auth.AuthenticationType())

	// Specify refresh token with other properties.
	auth, err = NewIamAuthenticatorBuilder().
		SetURL(iamAuthMockURL).
		SetRefreshToken(iamAuthMockRefreshToken).
		SetClientIDSecret(iamAuthMockClientID, iamAuthMockClientSecret).
		SetDisableSSLVerification(true).
		SetScope(iamAuthMockScope).
		SetHeaders(expectedHeaders).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Empty(t, auth.ApiKey)
	assert.Equal(t, iamAuthMockRefreshToken, auth.RefreshToken)
	assert.Equal(t, iamAuthMockURL, auth.URL)
	assert.Equal(t, iamAuthMockClientID, auth.ClientId)
	assert.Equal(t, iamAuthMockClientSecret, auth.ClientSecret)
	assert.True(t, auth.DisableSSLVerification)
	assert.Equal(t, iamAuthMockScope, auth.Scope)
	assert.Equal(t, expectedHeaders, auth.Headers)
	assert.Equal(t, AUTHTYPE_IAM, auth.AuthenticationType())
}

func TestIamAuthReuseAuthenticator(t *testing.T) {
	auth, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Use the authenticator to construct a service.
	service, err := NewBaseService(&ServiceOptions{
		URL:           "don't care",
		Authenticator: auth,
	})
	assert.Nil(t, err)
	assert.NotNil(t, service)

	// Simulate use of the authenticator by setting the RefreshToken
	// field (this will be set when processing an IAM get-token response).
	auth.RefreshToken = iamAuthMockRefreshToken

	// Now re-use the authenticator with a new service.
	service, err = NewBaseService(&ServiceOptions{
		URL:           "don't care",
		Authenticator: auth,
	})
	assert.Nil(t, err)
	assert.NotNil(t, service)
}

// Tests that construct an authenticator via map properties.
func TestNewIamAuthenticatorFromMap(t *testing.T) {
	_, err := newIamAuthenticatorFromMap(nil)
	assert.NotNil(t, err)

	var props = map[string]string{
		PROPNAME_AUTH_URL: iamAuthMockURL,
	}
	_, err = newIamAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_APIKEY: "",
	}
	_, err = newIamAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_APIKEY: iamAuthMockApiKey,
	}
	authenticator, err := newIamAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, iamAuthMockApiKey, authenticator.ApiKey)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())

	props = map[string]string{
		PROPNAME_APIKEY:           iamAuthMockApiKey,
		PROPNAME_AUTH_DISABLE_SSL: "false",
		PROPNAME_CLIENT_ID:        iamAuthMockClientID,
		PROPNAME_CLIENT_SECRET:    iamAuthMockClientSecret,
		PROPNAME_SCOPE:            iamAuthMockScope,
	}
	authenticator, err = newIamAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, iamAuthMockApiKey, authenticator.ApiKey)
	assert.Empty(t, authenticator.RefreshToken)
	assert.False(t, authenticator.DisableSSLVerification)
	assert.Equal(t, iamAuthMockClientID, authenticator.ClientId)
	assert.Equal(t, iamAuthMockClientSecret, authenticator.ClientSecret)
	assert.Equal(t, iamAuthMockScope, authenticator.Scope)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())

	props = map[string]string{
		PROPNAME_REFRESH_TOKEN:    iamAuthMockRefreshToken,
		PROPNAME_AUTH_DISABLE_SSL: "false",
		PROPNAME_CLIENT_ID:        iamAuthMockClientID,
		PROPNAME_CLIENT_SECRET:    iamAuthMockClientSecret,
		PROPNAME_SCOPE:            iamAuthMockScope,
	}
	authenticator, err = newIamAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Empty(t, authenticator.ApiKey)
	assert.Equal(t, iamAuthMockRefreshToken, authenticator.RefreshToken)
	assert.False(t, authenticator.DisableSSLVerification)
	assert.Equal(t, iamAuthMockClientID, authenticator.ClientId)
	assert.Equal(t, iamAuthMockClientSecret, authenticator.ClientSecret)
	assert.Equal(t, iamAuthMockScope, authenticator.Scope)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())
}

func TestIamAuthDefaultURL(t *testing.T) {
	auth := &IamAuthenticator{}
	s := auth.url()
	assert.Equal(t, s, defaultIamTokenServerEndpoint)
	assert.Equal(t, auth.URL, defaultIamTokenServerEndpoint)
}

// Tests involving the legacy ctor
func TestIamConfigErrors(t *testing.T) {
	var err error

	// Missing ApiKey.
	_, err = NewIamAuthenticator("", "", "foo", "bar", false, nil)
	assert.NotNil(t, err)

	// Invalid ApiKey.
	_, err = NewIamAuthenticator("{invalid-apikey}", "", "foo", "bar", false, nil)
	assert.NotNil(t, err)

	// Missing ClientId.
	_, err = NewIamAuthenticator(iamAuthMockApiKey, "", "", "bar", false, nil)
	assert.NotNil(t, err)

	// Missing ClientSecret.
	_, err = NewIamAuthenticator(iamAuthMockApiKey, "", "foo", "", false, nil)
	assert.NotNil(t, err)
}

func TestIamAuthenticateFail(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Sorry you are not authorized"))
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	// Create a new Request object.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://localhost/placeholder/url", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	err = authenticator.Authenticate(request)
	assert.NotNil(t, err)
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	// The casted error should match the original error message
	assert.Equal(t, err.Error(), authErr.Error())
}

func TestIamGetTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		assert.Nil(t, err)
		assert.Len(t, r.Form["apikey"], 1)
		assert.Len(t, r.Form["grant_type"], 1)
		assert.Len(t, r.Form["response_type"], 1)
		assert.Empty(t, r.Form["scope"])

		w.WriteHeader(http.StatusOK)
		expiration := GetCurrentTime() + 3600
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "%s"
			}`, iamAuthTestAccessToken1, expiration, iamAuthTestRefreshToken)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		} else {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "%s"
			}`, iamAuthTestAccessToken2, expiration, iamAuthTestRefreshToken)
			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, iamAuthMockClientID, username)
			assert.Equal(t, iamAuthMockClientSecret, password)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Also make sure we get back a nil error from synchronizedRequestToken().
	assert.Nil(t, authenticator.synchronizedRequestToken())

	// Force expiration and verify that we got the second access token.
	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
	authenticator.ClientId = iamAuthMockClientID
	authenticator.ClientSecret = iamAuthMockClientSecret
	_, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, iamAuthTestAccessToken2, authenticator.getTokenData().AccessToken)

	// Test the RequestToken() method to make sure we can get a refresh token.
	tokenResponse, err := authenticator.RequestToken()
	assert.Nil(t, err)
	assert.NotNil(t, tokenResponse)
	assert.Equal(t, iamAuthTestRefreshToken, tokenResponse.RefreshToken)
}

func TestIamGetTokenSuccessRT(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	var newRefreshToken string = "new-refresh-token"

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		assert.Nil(t, err)
		assert.Len(t, r.Form["refresh_token"], 1)
		assert.Len(t, r.Form["grant_type"], 1)
		assert.Equal(t, "refresh_token", r.Form["grant_type"][0])
		assert.Len(t, r.Form["response_type"], 1)
		assert.Len(t, r.Form["scope"], 1)

		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, iamAuthMockClientID, username)
		assert.Equal(t, iamAuthMockClientSecret, password)

		w.WriteHeader(http.StatusOK)
		expiration := GetCurrentTime() + 3600
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "%s"
			}`, iamAuthTestAccessToken1, expiration, iamAuthTestRefreshToken)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "%s"
			}`, iamAuthTestAccessToken2, expiration, newRefreshToken)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetRefreshToken(iamAuthMockRefreshToken).
		SetClientIDSecret(iamAuthMockClientID, iamAuthMockClientSecret).
		SetURL(server.URL).
		SetScope(iamAuthMockScope).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Nil(t, authenticator.getTokenData())
	assert.Equal(t, iamAuthMockRefreshToken, authenticator.RefreshToken)

	// Force the first fetch and verify we got the first access token.
	// From this first fetch, we should also the first refresh token saved to the authenticator.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, iamAuthTestRefreshToken, authenticator.RefreshToken)

	// Force expiration and verify that we got the second access token.
	// At this point, we should also have a second refresh token saved in the authenticator.
	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
	_, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, iamAuthTestAccessToken2, authenticator.getTokenData().AccessToken)
	assert.Equal(t, newRefreshToken, authenticator.RefreshToken)
}

func TestIamGetTokenSuccessWithScope(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		assert.Nil(t, err)
		assert.Len(t, r.Form["apikey"], 1)
		assert.Len(t, r.Form["grant_type"], 1)
		assert.Len(t, r.Form["response_type"], 1)
		assert.Len(t, r.Form["scope"], 1)
		assert.Equal(t, iamAuthMockScope, r.Form["scope"][0])

		w.WriteHeader(http.StatusOK)
		expiration := GetCurrentTime() + 3600
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken1, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		} else {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken2, expiration)
			username, password, ok := r.BasicAuth()
			assert.True(t, ok)
			assert.Equal(t, iamAuthMockClientID, username)
			assert.Equal(t, iamAuthMockClientSecret, password)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		SetScope(iamAuthMockScope).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Force expiration and verify that we got the second access token.
	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
	authenticator.ClientId = iamAuthMockClientID
	authenticator.ClientSecret = iamAuthMockClientSecret
	_, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, iamAuthTestAccessToken2, authenticator.getTokenData().AccessToken)
}
func TestIamGetCachedToken(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		expiration := GetCurrentTime() + 3600
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken1, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		} else {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken2, expiration)
			_, _, ok := r.BasicAuth()
			assert.True(t, ok)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Subsequent fetch should still return first access token.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())
}

func TestIamBackgroundTokenRefresh(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiration := GetCurrentTime() + 3600
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken1, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		} else {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken2, expiration)
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	authenticator.getTokenData().RefreshTime = GetCurrentTime() - 720

	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Wait for the background thread to finish
	time.Sleep(5 * time.Second)
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken2, token)
	assert.NotNil(t, authenticator.getTokenData())
}

func TestIamBackgroundTokenRefreshFailure(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expiration := GetCurrentTime() + 3600
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken1, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		} else {
			_, _ = w.Write([]byte("Sorry you are forbidden"))
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Successfully fetch the first token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	authenticator.getTokenData().RefreshTime = GetCurrentTime() - 720
	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())
	// Wait for the background thread to finish
	time.Sleep(5 * time.Second)
	_, err = authenticator.GetToken()
	assert.NotNil(t, err)
	assert.Equal(t, "Error while trying to get access token", err.Error())
	// We don't expect a AuthenticateError to be returned, so casting should fail
	_, ok := err.(*AuthenticationError)
	assert.False(t, ok)

}

func TestIamBackgroundTokenRefreshIdle(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CurrentTime + 1 hour
		expiration := GetCurrentTime() + 3600
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken1, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		} else {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken2, expiration)
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now simulate the client being idle for 10 minutes into the refresh time
	tenMinutesBeforeNow := GetCurrentTime() - 600
	authenticator.getTokenData().RefreshTime = tenMinutesBeforeNow

	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// RefreshTime should have advanced by 1 minute from the current time
	newRefreshTime := GetCurrentTime() + 60
	assert.Equal(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

	// In the next request, the RefreshTime should be unchanged and another thread
	// shouldn't be spawned to request another token once more since the first thread already spawned
	// a goroutine & refreshed the token.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)

	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

	// Wait for the background thread to finish and verify both the RefreshTime & tokenData were updated
	time.Sleep(5 * time.Second)
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken2, token)
	assert.NotNil(t, authenticator.getTokenData())
	assert.NotEqual(t, newRefreshTime, authenticator.getTokenData().RefreshTime)
}

func TestIamClientIdAndSecret(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	expiration := GetCurrentTime() + 3600
	accessToken := "oAeisG8yqPY7sFR_x66Z15"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"access_token": "%s",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": %d,
			"refresh_token": "jy4gl91BQ"
		}`, accessToken, expiration)
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, iamAuthMockClientID, username)
		assert.Equal(t, iamAuthMockClientSecret, password)
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		SetClientIDSecret(iamAuthMockClientID, iamAuthMockClientSecret).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	token, err := authenticator.GetToken()
	assert.Equal(t, accessToken, token)
	assert.Nil(t, err)
}

func TestIamRefreshTimeCalculation(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	const timeToLive int64 = 3600
	const expireTime int64 = 1563911183
	const expected int64 = expireTime - 720 // 720 is 20% of 3600

	// Simulate a token server response.
	tokenResponse := &IamTokenServerResponse{
		ExpiresIn:  timeToLive,
		Expiration: expireTime,
	}

	// Create a new token data and verify the refresh time.
	tokenData, err := newIamTokenData(tokenResponse)
	assert.Nil(t, err)
	assert.Equal(t, expected, tokenData.RefreshTime)
}

func TestIamDisableSSL(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	expiration := GetCurrentTime() + 3600
	accessToken := "oAeisG8yqPY7sFR_x66Z15"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"access_token": "%s",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": %d,
			"refresh_token": "jy4gl91BQ"
		}`, accessToken, expiration)
		_, _, ok := r.BasicAuth()
		assert.False(t, ok)
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		SetDisableSSLVerification(true).
		Build()
	assert.Nil(t, err)

	token, err := authenticator.GetToken()
	assert.Equal(t, accessToken, token)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.Client)
	assert.NotNil(t, authenticator.Client.Transport)
	transport, ok := authenticator.Client.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestIamUserHeaders(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	expiration := GetCurrentTime() + 3600
	accessToken := "oAeisG8yqPY7sFR_x66Z15"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{
			"access_token": "%s",
			"token_type": "Bearer",
			"expires_in": 3600,
			"expiration": %d,
			"refresh_token": "jy4gl91BQ"
		}`, accessToken, expiration)
		_, _, ok := r.BasicAuth()
		assert.False(t, ok)
		assert.Equal(t, "Value1", r.Header.Get("Header1"))
		assert.Equal(t, "Value2", r.Header.Get("Header2"))
		assert.Equal(t, "iam.cloud.ibm.com", r.Host)
	}))
	defer server.Close()

	var headers = map[string]string{
		"Header1": "Value1",
		"Header2": "Value2",
		"Host":    "iam.cloud.ibm.com",
	}

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		SetHeaders(headers).
		Build()
	assert.Nil(t, err)

	token, err := authenticator.GetToken()
	assert.Equal(t, accessToken, token)
	assert.Nil(t, err)
}

func TestIamGetTokenFailure(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Sorry you are forbidden"))
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)

	var expectedResponse = []byte("Sorry you are forbidden")

	_, err = authenticator.GetToken()
	assert.NotNil(t, err)
	assert.Equal(t, "Sorry you are forbidden", err.Error())
	// We expect an AuthenticationError to be returned, so cast the returned error
	authError, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authError)
	assert.NotNil(t, authError.Error())
	assert.NotNil(t, authError.Response)
	rawResult := authError.Response.GetRawResult()
	assert.NotNil(t, rawResult)
	assert.Equal(t, expectedResponse, rawResult)
	statusCode := authError.Response.GetStatusCode()
	assert.Equal(t, "Sorry you are forbidden", authError.Error())
	assert.Equal(t, http.StatusForbidden, statusCode)
}

func TestIamGetTokenTimeoutError(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		expiration := GetCurrentTime() + 3600
		if firstCall {
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken1, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		} else {
			time.Sleep(3 * time.Second)
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken2, expiration)
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Force expiration and verify that we got a timeout error
	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600

	// Set the client timeout to something very low
	authenticator.Client.Timeout = time.Second * 2
	token, err = authenticator.GetToken()
	assert.Empty(t, token)
	assert.NotNil(t, err)
	assert.NotNil(t, err.Error())
	// We don't expect a AuthenticateError to be returned, so casting should fail
	_, ok := err.(*AuthenticationError)
	assert.False(t, ok)
}

func TestIamGetTokenServerError(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	firstCall := true
	expiration := GetCurrentTime() + 3600
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if firstCall {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{
				"access_token": "%s",
				"token_type": "Bearer",
				"expires_in": 3600,
				"expiration": %d,
				"refresh_token": "jy4gl91BQ"
			}`, iamAuthTestAccessToken1, expiration)
			firstCall = false
			_, _, ok := r.BasicAuth()
			assert.False(t, ok)
		} else {
			w.WriteHeader(http.StatusGatewayTimeout)
			_, _ = w.Write([]byte("Gateway Timeout"))
		}
	}))
	defer server.Close()

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	var expectedResponse = []byte("Gateway Timeout")

	// Force expiration and verify that we got a server error
	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
	token, err = authenticator.GetToken()
	assert.NotNil(t, err)
	// We expect an AuthenticationError to be returned, so cast the returned error
	authError, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authError)
	assert.NotNil(t, authError.Response)
	assert.NotNil(t, authError.Error())
	rawResult := authError.Response.GetRawResult()
	statusCode := authError.Response.GetStatusCode()
	assert.Equal(t, "Gateway Timeout", authError.Error())
	assert.Equal(t, expectedResponse, rawResult)
	assert.NotNil(t, rawResult)
	assert.Equal(t, http.StatusGatewayTimeout, statusCode)
	assert.Empty(t, token)
}

func TestIamRequestTokenError1(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	authenticator, err := NewIamAuthenticatorBuilder().
		SetApiKey(iamAuthMockApiKey).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	// Now forcibly clear the ApiKey field so we can test an error condition.
	authenticator.ApiKey = ""

	_, err = authenticator.RequestToken()
	assert.NotNil(t, err)
	t.Logf("Expected error: %s", err.Error())
}

func TestIamRequestTokenError2(t *testing.T) {
	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	// Force an error while resolving the service URL.
	auth := &IamAuthenticator{
		URL: "123:badpath",
	}

	iamToken, err := auth.RequestToken()
	assert.NotNil(t, err)
	assert.Nil(t, iamToken)
	t.Logf("Expected error: %s\n", err.Error())
}

func TestIamNewTokenDataError1(t *testing.T) {
	tokenData, err := newIamTokenData(nil)
	assert.NotNil(t, err)
	assert.Nil(t, tokenData)
	t.Logf("Expected error: %s\n", err.Error())
}

// In order to test with a live IAM server, create file "iamtest.env" in the project root.
// It should look like this:
//
//	IAMTEST1_AUTH_URL=<url>   e.g. https://iam.cloud.ibm.com
//	IAMTEST1_AUTH_TYPE=iam
//	IAMTEST1_APIKEY=<apikey>
//
// Then comment out the "t.Skip()" line below, then run these commands:
//
//	cd v<major-version>/core
//	go test -v -tags=auth -run=TestIamLiveTokenServer -v
//
// To trace request/response messages, change "iamAuthTestLogLevel" above to be "LevelDebug".
func TestIamLiveTokenServer(t *testing.T) {
	t.Skip("Skipping IAM integration test...")

	GetLogger().SetLogLevel(iamAuthTestLogLevel)

	var request *http.Request
	var err error
	var authHeader string
	var tokenServerResponse *IamTokenServerResponse

	// Get an iam authenticator from the environment.
	os.Setenv("IBM_CREDENTIALS_FILE", "../../iamtest.env")

	auth, err := GetAuthenticatorFromEnvironment("iamtest1")
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	iamAuth, ok := auth.(*IamAuthenticator)
	assert.Equal(t, true, ok)

	tokenServerResponse, err = iamAuth.RequestToken()
	if err != nil {
		authError := err.(*AuthenticationError)
		iamError := authError.Err
		iamResponse := authError.Response
		t.Logf("Unexpected authentication error: %s\n", iamError.Error())
		t.Logf("Authentication response: %v+\n", iamResponse)

	}
	assert.Nil(t, err)
	assert.NotNil(t, tokenServerResponse)

	accessToken := tokenServerResponse.AccessToken
	assert.NotEmpty(t, accessToken)
	t.Logf("Access token: %s\n", accessToken)

	refreshToken := tokenServerResponse.RefreshToken
	assert.NotEmpty(t, refreshToken)
	t.Logf("Refresh token: %s\n", refreshToken)

	// Create a new Request object.
	builder, err := NewRequestBuilder("GET").ResolveRequestURL("https://localhost/placeholder/url", "", nil)
	assert.Nil(t, err)
	assert.NotNil(t, builder)

	request, _ = builder.Build()
	assert.NotNil(t, request)
	err = auth.Authenticate(request)
	assert.Nil(t, err)

	authHeader = request.Header.Get("Authorization")
	assert.NotEmpty(t, authHeader)
	assert.True(t, strings.HasPrefix(authHeader, "Bearer "))
	t.Logf("Authorization: %s\n", authHeader)

	// Now create a new IamAuthenticator using bx:bx so that we can retrieve
	// the refresh token value and then do some testing with that.
	// We'll use the URL and ApiKey from the original authenticator above.
	newAuth, err := NewIamAuthenticatorBuilder().
		SetURL(iamAuth.URL).
		SetApiKey(iamAuth.ApiKey).
		SetClientIDSecret("bx", "bx").
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, newAuth)

	tokenServerResponse, err = newAuth.RequestToken()
	assert.Nil(t, err)
	assert.NotNil(t, tokenServerResponse)

	refreshToken = tokenServerResponse.RefreshToken
	assert.NotEmpty(t, refreshToken)

	// Create a new IamAuthenticator configured with the refresh token.
	refreshAuth, err := NewIamAuthenticatorBuilder().
		SetURL(newAuth.URL).
		SetRefreshToken(refreshToken).
		SetClientIDSecret("bx", "bx").
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, refreshAuth)
	assert.Equal(t, refreshToken, refreshAuth.RefreshToken)

	// Trigger the authenticator to invoke the "get token" operation.
	// and make sure that we got back an access token and that we
	// saved a different refresh token in the authenticator.
	accessToken, err = refreshAuth.GetToken()
	assert.Nil(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEqual(t, refreshToken, refreshAuth.RefreshToken)
}
