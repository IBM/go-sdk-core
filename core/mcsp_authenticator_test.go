//go:build all || slow || auth

package core

// (C) Copyright IBM Corp. 2023.
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
	"strings"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

var (
	// To enable debug logging during test execution, set this to "LevelDebug"
	mcspAuthTestLogLevel LogLevel = LevelError
	mcspAuthMockApiKey            = "mock-apikey"
	mcspAuthMockURL               = "https://mock.mcsp.com"

	mcspAuthTestAccessToken1 string = "eyJraWQiOiJ0WlZVVnQxSmZYR0ZSM3VuczFQLU12cWJuSGE0c2hPUnRJZEM1ZDR0d2o0IiwiYWxnIjoiUlMyNTYifQ.eyJpc3MiOiJodHRwczovL3NpdXNlcm1ndG1zLW1zcC11c2VyLW1hbmFnZXIuYXBwcy5hcC1kcC0xMDEuMnd4aC5wMS5vcGVuc2hpZnRhcHBzLmNvbS9zaXVzZXJtZ3IvYXBpLzEuMCIsImF1ZCI6ImNybjp2MTphd3Mtc3RhZ2luZzpwdWJsaWM6d3hvOnVzLWVhc3QtMTpzdWIvMjAyMzEwMjUtMDQ1Ni01NzI3LTIwNWMtOGU2YWJhM2FiZmUzOjIwMjMxMDI1LTEwMzUtMzY2Ny01MDA2LWUzNjU1N2IyOGRhODo6IiwiZXhwIjoxNjk5MDMzNzM2LCJqdGkiOiJfRFpHSWJPbHlSWmF5TjlFTUowWXpBIiwiaWF0IjoxNjk5MDI2NTM2LCJuYmYiOjE2OTkwMjY1MDYsInRlbmFudElkIjoiMjAyMzEwMjUtMTAzNS0zNjY3LTUwMDYtZTM2NTU3YjI4ZGE4Iiwic3Vic2NyaXB0aW9uSWQiOiIyMDIzMTAyNS0wNDU2LTU3MjctMjA1Yy04ZTZhYmEzYWJmZTMiLCJzdWIiOiI5MGRjZjU4ZC00NzgzLTNmOGUtOGMxNi05ZGU3NTMwNDE0ODAiLCJlbnRpdHlUeXBlIjoiVVNFUiIsImVtYWlsIjoic3Z0X3N0YWdlX2Vzc2VudGlhbEB3by1jZC50ZXN0aW5hdG9yLmNvbSIsIm5hbWUiOiJzdnRfc3RhZ2VfZXNzZW50aWFsQHdvLWNkLnRlc3RpbmF0b3IuY29tIiwiZGlzcGxheW5hbWUiOiJzdnRfc3RhZ2VfZXNzZW50aWFsQHdvLWNkLnRlc3RpbmF0b3IuY29tIiwiaWRwIjp7InJlYWxtTmFtZSI6ImNsb3VkSWRlbnRpdHlSZWFsbSIsImlzcyI6Imh0dHBzOi8vd28taWJtLXN0Zy52ZXJpZnkuaWJtLmNvbS9vaWRjL2VuZHBvaW50L2RlZmF1bHQifSwiZ3JvdXBzIjpbXSwicm9sZXMiOlsiQWRtaW4iXX0.alYTel_rX1JlN9tciTLl5fXSjs4CYbjq7Ywow8aGVG0ONm_GYNyNfhUQ4SGxvvxpA7inXQg-Hcx_K0pTEVPqrV-OUMNBcXJXcAO-ZszEcDgca_BdSxOAVTXV5Y8LkbBRJjJn3bzcZ5Yq0y0cTP0z-tSnRtmP8USyLrOclE3WLV966t_AFi2i0t1FnHFi7pHBoji4idwDK3uYHhduXsHDjiHD2QmydFXKNHYAIAP8De9aCDLsRfVE56ga9Gx2CQ46R5V5tfy5KkYor6RtBAifn-TZUGX5OOai3V-5DqtUrVtIdIGODJCAhFYiruOu4INOgwPdLQgzF0V3uqYeifyQCw"
	mcspAuthTestAccessToken2 string = "eyJraWQiOiJ0WlZVVnQxSmZYR0ZSM3VuczFQLU12cWJuSGE0c2hPUnRJZEM1ZDR0d2o0IiwiYWxnIjoiUlMyNTYifQ.eyJpc3MiOiJodHRwczovL3NpdXNlcm1ndG1zLW1zcC11c2VyLW1hbmFnZXIuYXBwcy5hcC1kcC0xMDEuMnd4aC5wMS5vcGVuc2hpZnRhcHBzLmNvbS9zaXVzZXJtZ3IvYXBpLzEuMCIsImF1ZCI6ImNybjp2MTphd3Mtc3RhZ2luZzpwdWJsaWM6d3hvOnVzLWVhc3QtMTpzdWIvMjAyMzEwMjUtMDQ1Ni01NzI3LTIwNWMtOGU2YWJhM2FiZmUzOjIwMjMxMDI1LTEwMzUtMzY2Ny01MDA2LWUzNjU1N2IyOGRhODo6IiwiZXhwIjoxNjk5MDQ1MDUyLCJqdGkiOiI1dkpvdk85SXJtRnUwWlZTTFBxTmZnIiwiaWF0IjoxNjk5MDM3ODUyLCJuYmYiOjE2OTkwMzc4MjIsInRlbmFudElkIjoiMjAyMzEwMjUtMTAzNS0zNjY3LTUwMDYtZTM2NTU3YjI4ZGE4Iiwic3Vic2NyaXB0aW9uSWQiOiIyMDIzMTAyNS0wNDU2LTU3MjctMjA1Yy04ZTZhYmEzYWJmZTMiLCJzdWIiOiI5MGRjZjU4ZC00NzgzLTNmOGUtOGMxNi05ZGU3NTMwNDE0ODAiLCJlbnRpdHlUeXBlIjoiVVNFUiIsImVtYWlsIjoic3Z0X3N0YWdlX2Vzc2VudGlhbEB3by1jZC50ZXN0aW5hdG9yLmNvbSIsIm5hbWUiOiJzdnRfc3RhZ2VfZXNzZW50aWFsQHdvLWNkLnRlc3RpbmF0b3IuY29tIiwiZGlzcGxheW5hbWUiOiJzdnRfc3RhZ2VfZXNzZW50aWFsQHdvLWNkLnRlc3RpbmF0b3IuY29tIiwiaWRwIjp7InJlYWxtTmFtZSI6ImNsb3VkSWRlbnRpdHlSZWFsbSIsImlzcyI6Imh0dHBzOi8vd28taWJtLXN0Zy52ZXJpZnkuaWJtLmNvbS9vaWRjL2VuZHBvaW50L2RlZmF1bHQifSwiZ3JvdXBzIjpbXSwicm9sZXMiOlsiQWRtaW4iXX0.eFDY62qebPUehd-Bkz9xNzJjNwoGkLYBFhybo-Py97gc100wp9WItBcC409O86mZxsH79zCDqGOHNrrVirh11yv0iv7D2_wt9hHDpHsG48pNmzvLzkRKy-a7xW_YsYB_Es3h3FeXv-nRWBxWLGdel6kkW-OAl1hnuC53r0n2ADO863ifbUlvzhxECWJSsMMCH_ZSJ_ejzGQcKNtPMRYNAgnsdey5qEvQ_Ae_ntt7iGCsOpYfmky0U3CZhMd9QkIvoQC8ulpkYmusmVQzAosCqQtgNGSBP2ekvYgI79v3ZB3c3oQC1aEJOuUGXhrbP7PRnLAkgnEZDAbrIMlQyP9ddA"
)

// Tests involving the Builder
func TestMCSPAuthBuilderErrors(t *testing.T) {
	var err error
	var auth *MCSPAuthenticator

	// Error: no apikey
	auth, err = NewMCSPAuthenticatorBuilder().
		SetApiKey("").
		SetURL(mcspAuthMockURL).
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: no url
	auth, err = NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL("").
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}

func TestMCSPAuthBuilderSuccess(t *testing.T) {
	var err error
	var auth *MCSPAuthenticator
	var expectedHeaders = map[string]string{
		"header1": "value1",
	}

	// Specify apikey.
	auth, err = NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(mcspAuthMockURL).
		SetClient(nil).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, mcspAuthMockApiKey, auth.ApiKey)
	assert.Equal(t, mcspAuthMockURL, auth.URL)
	assert.False(t, auth.DisableSSLVerification)
	assert.Nil(t, auth.Headers)
	assert.Equal(t, AUTHTYPE_MCSP, auth.AuthenticationType())

	// Specify apikey with other properties.
	auth, err = NewMCSPAuthenticatorBuilder().
		SetURL(mcspAuthMockURL).
		SetApiKey(mcspAuthMockApiKey).
		SetDisableSSLVerification(true).
		SetHeaders(expectedHeaders).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, mcspAuthMockApiKey, auth.ApiKey)
	assert.Equal(t, mcspAuthMockURL, auth.URL)
	assert.True(t, auth.DisableSSLVerification)
	assert.Equal(t, expectedHeaders, auth.Headers)
	assert.Equal(t, AUTHTYPE_MCSP, auth.AuthenticationType())
}

func TestMCSPAuthReuseAuthenticator(t *testing.T) {
	auth, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(mcspAuthMockURL).
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

	// Now re-use the authenticator with a new service.
	service, err = NewBaseService(&ServiceOptions{
		URL:           "don't care",
		Authenticator: auth,
	})
	assert.Nil(t, err)
	assert.NotNil(t, service)
}

// Tests that construct an authenticator via map properties.
func TestNewMCSPAuthenticatorFromMap(t *testing.T) {
	_, err := newMCSPAuthenticatorFromMap(nil)
	assert.NotNil(t, err)

	// Missing ApiKey
	var props = map[string]string{
		PROPNAME_AUTH_URL: mcspAuthMockURL,
	}
	_, err = newMCSPAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	// Missing URL
	props = map[string]string{
		PROPNAME_APIKEY: mcspAuthMockApiKey,
	}
	_, err = newMCSPAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	// Valid configuration.
	props = map[string]string{
		PROPNAME_APIKEY:   mcspAuthMockApiKey,
		PROPNAME_AUTH_URL: mcspAuthMockURL,
	}
	authenticator, err := newMCSPAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, mcspAuthMockApiKey, authenticator.ApiKey)
	assert.Equal(t, mcspAuthMockURL, authenticator.URL)
	assert.Equal(t, AUTHTYPE_MCSP, authenticator.AuthenticationType())

	props = map[string]string{
		PROPNAME_AUTH_URL:         mcspAuthMockURL,
		PROPNAME_APIKEY:           mcspAuthMockApiKey,
		PROPNAME_AUTH_DISABLE_SSL: "false",
	}
	authenticator, err = newMCSPAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, mcspAuthMockApiKey, authenticator.ApiKey)
	assert.Equal(t, mcspAuthMockURL, authenticator.URL)
	assert.False(t, authenticator.DisableSSLVerification)
	assert.Equal(t, AUTHTYPE_MCSP, authenticator.AuthenticationType())
}

func TestMCSPAuthenticateFail(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Sorry you are not authorized"))
	}))
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
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

// Struct that describes the requestBody for the "get token" operation.
type mcspRequestBody struct {
	ApiKey *string `json:"apikey,omitempty"`
}

func startMockServer(t *testing.T) *httptest.Server {
	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Unmarshal the request body and verify.
		requestBody := &mcspRequestBody{}
		_ = json.NewDecoder(req.Body).Decode(requestBody)
		defer req.Body.Close()
		assert.NotNil(t, requestBody.ApiKey)
		assert.Equal(t, mcspAuthMockApiKey, *requestBody.ApiKey)

		// Create the response.
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{"token":"%s","token_type":"jwt","expires_in":7200}`, mcspAuthTestAccessToken1)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{"token":"%s","token_type":"jwt","expires_in":7200}`, mcspAuthTestAccessToken2)
		}
	}))
	return server
}

func TestMCSPGetTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Force expiration and verify that we got the second access token.
	authenticator.getTokenData().Expiration = GetCurrentTime() - 7200
	_, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, mcspAuthTestAccessToken2, authenticator.getTokenData().AccessToken)
}

func TestMCSPGetCachedToken(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Set the expiration time to "force" the use of the cached token.
	tokenData := authenticator.getTokenData()
	tokenData.Expiration = GetCurrentTime() + 1800
	tokenData.RefreshTime = GetCurrentTime() + 1500

	// Subsequent fetch should still return first access token.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())
}

func TestMCSPBackgroundTokenRefresh(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	tokenData := authenticator.getTokenData()
	tokenData.Expiration = GetCurrentTime() + 1800
	tokenData.RefreshTime = GetCurrentTime() - 720

	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Wait for the background thread to finish
	time.Sleep(5 * time.Second)
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken2, token)
	assert.NotNil(t, authenticator.getTokenData())
}

func TestMCSPBackgroundTokenRefreshFailure(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{"token":"%s","token_type":"jwt","expires_in": 7200}`, mcspAuthTestAccessToken1)
			firstCall = false
		} else {
			_, _ = w.Write([]byte("Sorry you are forbidden"))
		}
	}))
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Successfully fetch the first token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	tokenData := authenticator.getTokenData()
	tokenData.Expiration = GetCurrentTime() + 1800
	tokenData.RefreshTime = GetCurrentTime() - 720

	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Wait for the background thread to finish.
	time.Sleep(5 * time.Second)
	_, err = authenticator.GetToken()
	assert.NotNil(t, err)
	assert.Equal(t, "Error while trying to get access token", err.Error())
	// We don't expect an AuthenticateError to be returned, so casting should fail
	_, ok := err.(*AuthenticationError)
	assert.False(t, ok)
}

func TestMCSPBackgroundTokenRefreshIdle(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now simulate the client being idle for 10 minutes into the refresh time
	tenMinutesBeforeNow := GetCurrentTime() - 600
	tokenData := authenticator.getTokenData()
	tokenData.Expiration = GetCurrentTime() + 1800
	tokenData.RefreshTime = tenMinutesBeforeNow

	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// RefreshTime should have advanced by 1 minute from the current time
	newRefreshTime := GetCurrentTime() + 60
	assert.Equal(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

	// In the next request, the RefreshTime should be unchanged and another thread
	// shouldn't be spawned to request another token once more since the first thread already spawned
	// a goroutine & refreshed the token.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)

	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

	// Wait for the background thread to finish and verify both the RefreshTime & tokenData were updated
	time.Sleep(5 * time.Second)
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken2, token)
	assert.NotNil(t, authenticator.getTokenData())
	assert.NotEqual(t, newRefreshTime, authenticator.getTokenData().RefreshTime)
}

func TestMCSPDisableSSL(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	server := startMockServer(t)
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		SetDisableSSLVerification(true).
		Build()
	assert.Nil(t, err)

	token, err := authenticator.GetToken()
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.Client)
	assert.NotNil(t, authenticator.Client.Transport)
	transport, ok := authenticator.Client.Transport.(*http.Transport)
	assert.True(t, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestMCSPUserHeaders(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"token":"%s","token_type":"jwt","expires_in":7200}`, mcspAuthTestAccessToken1)
		assert.Equal(t, "Value1", r.Header.Get("Header1"))
		assert.Equal(t, "Value2", r.Header.Get("Header2"))
		assert.Equal(t, "mcsp.cloud.ibm.com", r.Host)
	}))
	defer server.Close()

	var headers = map[string]string{
		"Header1": "Value1",
		"Header2": "Value2",
		"Host":    "mcsp.cloud.ibm.com",
	}

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		SetHeaders(headers).
		Build()
	assert.Nil(t, err)

	token, err := authenticator.GetToken()
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.Nil(t, err)
}

func TestMCSPGetTokenFailure(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	var expectedResponse = []byte("Sorry you are forbidden")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write(expectedResponse)
	}))
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)

	_, err = authenticator.GetToken()
	assert.NotNil(t, err)
	assert.Equal(t, string(expectedResponse), err.Error())

	// We expect an AuthenticationError to be returned, so cast the returned error.
	authError, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authError)
	assert.NotNil(t, authError.Error())
	assert.NotNil(t, authError.Response)
	rawResult := authError.Response.GetRawResult()
	assert.NotNil(t, rawResult)
	assert.Equal(t, expectedResponse, rawResult)
	statusCode := authError.Response.GetStatusCode()
	assert.Equal(t, string(expectedResponse), authError.Error())
	assert.Equal(t, http.StatusForbidden, statusCode)
}

func TestMCSPGetTokenTimeoutError(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if firstCall {
			fmt.Fprintf(w, `{"token":"%s","token_type":"jwt","expires_in":7200}`, mcspAuthTestAccessToken1)
			firstCall = false
		} else {
			time.Sleep(3 * time.Second)
			fmt.Fprintf(w, `{"token":"%s","token_type":"jwt","expires_in":7200}`, mcspAuthTestAccessToken2)
		}
	}))
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Force expiration and verify that we got a timeout error
	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600

	// Set the client timeout to something very low
	authenticator.Client.Timeout = time.Second * 2
	token, err = authenticator.GetToken()
	assert.Empty(t, token)
	assert.NotNil(t, err)
	assert.NotNil(t, err.Error())

	// We don't expect a AuthenticateError to be returned, so casting should fail.
	_, ok := err.(*AuthenticationError)
	assert.False(t, ok)
}

func TestMCSPGetTokenServerError(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	var expectedResponse = []byte("Gateway Timeout")

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if firstCall {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"token":"%s","token_type":"jwt","expires_in":7200}`, mcspAuthTestAccessToken1)
			firstCall = false
		} else {
			w.WriteHeader(http.StatusGatewayTimeout)
			_, _ = w.Write(expectedResponse)
		}
	}))
	defer server.Close()

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, mcspAuthTestAccessToken1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Force expiration and verify that we got a server error
	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600
	token, err = authenticator.GetToken()
	assert.NotNil(t, err)

	// We expect an AuthenticationError to be returned, so cast the returned error.
	authError, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authError)
	assert.NotNil(t, authError.Response)
	assert.NotNil(t, authError.Error())

	rawResult := authError.Response.GetRawResult()
	statusCode := authError.Response.GetStatusCode()
	assert.Equal(t, string(expectedResponse), authError.Error())
	assert.Equal(t, expectedResponse, rawResult)
	assert.NotNil(t, rawResult)
	assert.Equal(t, http.StatusGatewayTimeout, statusCode)
	assert.Empty(t, token)
}

func TestMCSPRequestTokenError1(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	authenticator, err := NewMCSPAuthenticatorBuilder().
		SetApiKey(mcspAuthMockApiKey).
		SetURL(mcspAuthMockURL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	// Now forcibly clear the ApiKey field so we can test an error condition.
	authenticator.ApiKey = ""

	_, err = authenticator.RequestToken()
	assert.NotNil(t, err)
	t.Logf("Expected error: %s", err.Error())
}

func TestMCSPRequestTokenError2(t *testing.T) {
	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	// Force an error while resolving the service URL.
	auth := &MCSPAuthenticator{
		ApiKey: mcspAuthMockApiKey,
		URL:    "123:badpath",
	}

	mcspToken, err := auth.RequestToken()
	assert.NotNil(t, err)
	assert.Nil(t, mcspToken)
	t.Logf("Expected error: %s\n", err.Error())
}

func TestMCSPNewTokenDataError1(t *testing.T) {
	tokenData, err := newMCSPTokenData(nil)
	assert.NotNil(t, err)
	assert.Nil(t, tokenData)
	t.Logf("Expected error: %s\n", err.Error())
}

// In order to test with a live token server, create file "mcsptest.env" in the project root.
// It should look like this:
//
//	MCSPTEST1_AUTH_URL=<url>   e.g. https://iam.platform.test.saas.ibm.com
//	MCSPTEST1_AUTH_TYPE=mcsp
//	MCSPTEST1_APIKEY=<apikey>
//
// Then comment out the "t.Skip()" line below, then run these commands:
//
//	cd core
//	go test -v -tags=auth -run=TestMCSPLiveTokenServer
//
// To trace request/response messages, change "mcspAuthTestLogLevel" above to be "LevelDebug".
func TestMCSPLiveTokenServer(t *testing.T) {
	t.Skip("Skipping MCSP integration test...")

	GetLogger().SetLogLevel(mcspAuthTestLogLevel)

	var request *http.Request
	var err error
	var authHeader1 string
	var authHeader2 string

	// Get an mcsp authenticator from the environment.
	t.Setenv("IBM_CREDENTIALS_FILE", "../mcsptest.env")

	auth, err := GetAuthenticatorFromEnvironment("mcsptest1")
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Verify that it is in fact an MCSPAuthenticator instance.
	_, ok := auth.(*MCSPAuthenticator)
	assert.Equal(t, true, ok)

	// Create a new Request object.
	builder, err := NewRequestBuilder("GET").ResolveRequestURL("https://localhost/placeholder/url", "", nil)
	assert.Nil(t, err)
	assert.NotNil(t, builder)
	request, _ = builder.Build()
	assert.NotNil(t, request)

	// Authenticate the request and verify that the Authorization header was added.
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader1 = request.Header.Get("Authorization")
	assert.NotEmpty(t, authHeader1)
	assert.True(t, strings.HasPrefix(authHeader1, "Bearer "))
	t.Logf("Authorization: %s\n", authHeader1)

	// Build a new request and then authenticate that and verify.
	request, _ = builder.Build()
	assert.NotNil(t, request)
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader2 = request.Header.Get("Authorization")
	assert.NotEmpty(t, authHeader2)
	assert.True(t, strings.HasPrefix(authHeader2, "Bearer "))

	// Make sure the auth header values from the two requests are the same.
	// We should have just used the cached access token in the second request.
	assert.Equal(t, authHeader1, authHeader2)

}
