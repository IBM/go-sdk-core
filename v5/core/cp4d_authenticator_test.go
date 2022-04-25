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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

const (
	// To enable debug logging during test execution, set this to "LevelDebug"
	cp4dAuthTestLogLevel LogLevel = LevelError

	// These access tokens were obtained by running curl to invoke the POST /v1/authorize against a CP4D environment.

	// Username/password
	// curl -k -X POST https://<host>/icp4d-api/v1/authorize -H 'Content-Type: application/json' \
	//      -d '{"username": "testuser", "password": "<password>" }'
	// #nosec
	cp4dUsernamePwd1 = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6IlVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhY2Nlc3NfY2F0YWxvZyIsImNhbl9wcm92aXNpb24iLCJzaWduX2luX29ubHkiXSwiZ3JvdXBzIjpbMTAwMDBdLCJzdWIiOiJ0ZXN0dXNlciIsImlzcyI6IktOT1hTU08iLCJhdWQiOiJEU1giLCJ1aWQiOiIxMDAwMzMxMDAzIiwiYXV0aGVudGljYXRvciI6ImRlZmF1bHQiLCJpYXQiOjE2MTA0ODg2NzAsImV4cCI6MTYxMDUzMTgzNH0.K5Mqsv3E9MMXotuhUWbAcTUe41thzSaiFOolnvNxIVPwApSJr_VappL8GTR6BgwPz5gB4MX9w8mVsh0vX8g5naRHWryKNxloHiWiOzCpI982EACkb7Lvdpo5vq_wOANM4OW5Q7cyWXMrqQMz1wF-4-1EyYHBbAKWWGmSQZ6iW7wgMxoeP027vGTD96IVFhgOrvX1hEBDMZ0S9gfKU0bthUMEDKoWONcFuWlHQChhh7agjP2RS4d3Rcjx2oHtx_zuH5bEXxn9g4Dj2v9Bkn6aOFQivSGFUlaus_6opZ6x5aCPi6SXnO_xOY_f2XKU-DUg-yN5BeX7fXu35JQTGFcgwQ"
	// #nosec
	cp4dUsernamePwd2 = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6IlVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhY2Nlc3NfY2F0YWxvZyIsImNhbl9wcm92aXNpb24iLCJzaWduX2luX29ubHkiXSwiZ3JvdXBzIjpbMTAwMDBdLCJzdWIiOiJ0ZXN0dXNlciIsImlzcyI6IktOT1hTU08iLCJhdWQiOiJEU1giLCJ1aWQiOiIxMDAwMzMxMDAzIiwiYXV0aGVudGljYXRvciI6ImRlZmF1bHQiLCJpYXQiOjE2MTA0ODg2NzcsImV4cCI6MTYxMDUzMTg0MX0.NZiTLN_D5ayzHNonVdba-B_ej7mugBsDgEUa63KUtEOhXOJ_E4ZOlwVKKhd0PUvn0ZPabfaD4-hmekDC6-TgnCn3aAkwwfM7hAaY9XvFwECzRC2yg08dgenc8TiOHhQ_7tYnpJxNbhyN-3guiMl3YgB46rbPSWbpgbd0z8uzZwVpSOY3AfbgLSyiDWN9dqOELeUxI1tUubMVWspbflrXhpS-p61UGsO3uqBSpMPw6fG19kSdaKdRaGx-Wg2uiTAZqAVrGNAGIVX-X_sx2EJcjYYK8_1n9O5bHSPQ3HonOtgYvqD2FbbBdiX9H7TblucYr-mMTjktuWXDYprvqRHg2g"

	// Username/apikey
	// curl -k -X POST https://<host>/icp4d-api/v1/authorize -H 'Content-Type: application/json' \
	//      -d '{"username": "testuser", "api_key": "<apikey>" }'
	// #nosec
	cp4dUsernameApikey1 = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6IlVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhY2Nlc3NfY2F0YWxvZyIsImNhbl9wcm92aXNpb24iLCJzaWduX2luX29ubHkiXSwiZ3JvdXBzIjpbMTAwMDBdLCJzdWIiOiJ0ZXN0dXNlciIsImlzcyI6IktOT1hTU08iLCJhdWQiOiJEU1giLCJ1aWQiOiIxMDAwMzMxMDAzIiwiYXV0aGVudGljYXRvciI6ImRlZmF1bHQiLCJpYXQiOjE2MTA1NDgyNDgsImV4cCI6MTYxMDU5MTQxMn0.I8MgxrapKRt0nOn0F41NtLHQ5HGmInZNaJIWcNwyBgLWI5YY_98kpKLecN5d9Ll9g0_lapAFs_b8xpTya0Lvnp2Q81SloRFpDhAMUVHVWq46g2dvZd1JpoFB8NHwrkz2qE_JUHBIonJmQusy8vMm1m1CPy0pE6fTYH1d5EJG2vLo6f2eFiDizLfGxb0ym9lUOkK6dgNZw2T32N8IoSYNan6BQU25Jai6llWRLwZda7R521EPEw2AtPDsd95AxoTd8f4pptxfkL2uXpT35wRguap_09sRlvDTR18Ghs-GbtCh3Do-8OPGEFYKvJkSHNpiXPw8pvHEe5jCGl3l3F5vXQ"
	// #nosec
	cp4dUsernameApikey2 = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6IlVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhY2Nlc3NfY2F0YWxvZyIsImNhbl9wcm92aXNpb24iLCJzaWduX2luX29ubHkiXSwiZ3JvdXBzIjpbMTAwMDBdLCJzdWIiOiJ0ZXN0dXNlciIsImlzcyI6IktOT1hTU08iLCJhdWQiOiJEU1giLCJ1aWQiOiIxMDAwMzMxMDAzIiwiYXV0aGVudGljYXRvciI6ImRlZmF1bHQiLCJpYXQiOjE2MTA1NDg5ODgsImV4cCI6MTYxMDU5MjE1Mn0.NQcEUveRFm87ZZhh6v74kqHNC-MC_frLg3dQxUk5gcviN32DjSeg6THDIpQoi85I1tkEMuuOYyZejrh4f_AsEteNVRXOdrmprB35VqDBdIlH1jFUl2DIVQXR93CKr_Flh31RPFDd43Ut9ZHraZaUWmnzlJxv8170t4-5f2eJASG2EqDZXqxqu9zEpHBvBefwkgKClWcFF9VcfJJCqRkbBNhNZhRQu5sH62VUiQqS-CStMsYn8NCgvj5WMqgcXMMFSX3B6poPvhhk-uPtUAiK50iPnEbQlTZNajLAAd-whn8TV2LFOrfKCfO-USWy-lbG8F-koM0tfAi0N4WzySqErg"
)

// getJSONRequestBody unmarshals the body contained in 'req' into 'result'
func getJSONRequestBody(req *http.Request, result interface{}) error {
	defer req.Body.Close()
	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}

	// Unmarshal the byte array as JSON.
	err = json.Unmarshal(buf, result)
	if err != nil {
		return err
	}

	return nil
}

func TestCp4dConfigErrorsPW1(t *testing.T) {
	var err error

	// Tests using the original ctor.

	// Missing URL.
	_, err = NewCloudPakForDataAuthenticator("", "mookie", "betts", false, nil)
	assert.NotNil(t, err)

	// Missing Username.
	_, err = NewCloudPakForDataAuthenticator("cp4d-url", "", "betts", false, nil)
	assert.NotNil(t, err)

	// Missing Password.
	_, err = NewCloudPakForDataAuthenticator("cp4d-url", "mookie", "", false, nil)
	assert.NotNil(t, err)
}

func TestCp4dConfigErrorsPW2(t *testing.T) {
	var err error

	// Tests using the new "UsingPassword" ctor.

	// Missing URL.
	_, err = NewCloudPakForDataAuthenticatorUsingPassword("", "mookie", "betts", false, nil)
	assert.NotNil(t, err)

	// Missing Username.
	_, err = NewCloudPakForDataAuthenticatorUsingPassword("cp4d-url", "", "betts", false, nil)
	assert.NotNil(t, err)

	// Missing Password.
	_, err = NewCloudPakForDataAuthenticatorUsingPassword("cp4d-url", "mookie", "", false, nil)
	assert.NotNil(t, err)
}

func TestCp4dConfigErrorsAPIKey(t *testing.T) {
	var err error

	// Tests using the new "UsingAPIKey" ctor.

	// Missing URL.
	_, err = NewCloudPakForDataAuthenticatorUsingAPIKey("", "mookie", "betts", false, nil)
	assert.NotNil(t, err)

	// Missing Username.
	_, err = NewCloudPakForDataAuthenticatorUsingAPIKey("cp4d-url", "", "betts", false, nil)
	assert.NotNil(t, err)

	// Missing APIKey.
	_, err = NewCloudPakForDataAuthenticatorUsingAPIKey("cp4d-url", "mookie", "", false, nil)
	assert.NotNil(t, err)
}

func TestCp4dAuthenticateFailure(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Sorry you are not authorized"))
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingPassword(server.URL, "mookie", "betts", false, nil)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, authenticator.AuthenticationType(), AUTHTYPE_CP4D)

	// Create a new Request object.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://localhost/placeholder/url", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	err = authenticator.Authenticate(request)
	// Validate the resulting error is a valid
	assert.NotNil(t, err)
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	// The casted error should match the original error message
	assert.Equal(t, err.Error(), authErr.Error())
}

func verifyAuthRequest(t *testing.T, r *http.Request,
	expectedUsername string, expectedPassword string, expectedApikey string) {

	assert.True(t, strings.HasSuffix(r.URL.String(), "/v1/authorize"))
	assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
	var requestBody *cp4dRequestBody
	err := getJSONRequestBody(r, &requestBody)
	assert.Nil(t, err)
	assert.NotNil(t, requestBody)
	assert.Equal(t, expectedUsername, requestBody.Username)
	assert.Equal(t, expectedPassword, requestBody.Password)
	assert.Equal(t, expectedApikey, requestBody.APIKey)
}

func TestCp4dGetTokenSuccessPW(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "mookie", "betts", "")

		w.WriteHeader(http.StatusOK)

		if firstCall {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernamePwd1)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernamePwd2)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticator(server.URL, "mookie", "betts", false, nil)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	// Force the first fetch and verify we got the correct access token back
	accessToken, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernamePwd1, accessToken)

	// Also make sure we get back a nil error from synchronizedRequestToken().
	assert.Nil(t, authenticator.synchronizedRequestToken())

	// Force an expiration and verify we get back the second access token.
	authenticator.setTokenData(nil)
	accessToken, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, cp4dUsernamePwd2, accessToken)
}

func TestCp4dGetTokenSuccessAPIKey(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "mookie", "", "my_apikey")

		w.WriteHeader(http.StatusOK)

		if firstCall {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey1)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey2)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingAPIKey(server.URL, "mookie", "my_apikey", false, nil)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)

	// Force the first fetch and verify we got the correct access token back
	accessToken, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, accessToken)

	// Force an expiration and verify we get back the second access token.
	authenticator.setTokenData(nil)
	accessToken, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, cp4dUsernameApikey2, accessToken)
}

func TestCp4dGetCachedTokenSuccessPW(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "john", "snow", "")

		w.WriteHeader(http.StatusOK)

		if firstCall {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernamePwd1)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernamePwd2)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingPassword(server.URL, "john", "snow", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernamePwd1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// To mock the cache, set the expiration on the existing token to be somewhere in the valid timeframe
	authenticator.getTokenData().Expiration = GetCurrentTime() + 9999

	// Subsequent fetch should still return first access token.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernamePwd1, token)
	assert.NotNil(t, authenticator.getTokenData())
}

func TestCp4dGetCachedTokenAPIKey(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "john", "", "King of the North")

		w.WriteHeader(http.StatusOK)

		if firstCall {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey1)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey2)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingAPIKey(server.URL, "john", "King of the North", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// To mock the cache, set the expiration on the existing token to be somewhere in the valid timeframe
	authenticator.getTokenData().Expiration = GetCurrentTime() + 9999

	// Subsequent fetch should still return first access token.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.NotNil(t, authenticator.getTokenData())
}

func TestCp4dGetTokenAuthFailure(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("Forbidden!"))
	}))
	defer server.Close()

	authenticator := &CloudPakForDataAuthenticator{
		URL:      server.URL,
		Username: "john",
		Password: "snow",
	}

	_, err := authenticator.GetToken()
	assert.NotNil(t, err)
	assert.Equal(t, "Forbidden!", err.Error())

	// We expect an AuthenticationError to be returned, so cast the returned error
	authError, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authError)
	assert.NotNil(t, authError.Error())
	assert.NotNil(t, authError.Response)
	rawResult := authError.Response.GetRawResult()
	assert.NotNil(t, rawResult)
	assert.Equal(t, []byte("Forbidden!"), rawResult)

	statusCode := authError.Response.GetStatusCode()
	assert.Equal(t, "Forbidden!", authError.Error())
	assert.Equal(t, http.StatusForbidden, statusCode)
}

func TestCp4dGetTokenDeserFailure(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Bad access token response!"))
	}))
	defer server.Close()

	authenticator := &CloudPakForDataAuthenticator{
		URL:      server.URL,
		Username: "john",
		Password: "snow",
	}

	_, err := authenticator.GetToken()
	assert.NotNil(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "error unmarshalling authentication response"))

	// We expect something other than an AuthenticationError to be returned.
	authError, ok := err.(*AuthenticationError)
	assert.False(t, ok)
	assert.Nil(t, authError)
}

func TestCp4dBackgroundTokenRefreshSuccess(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "john", "", "King of the North")

		w.WriteHeader(http.StatusOK)

		if firstCall {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey1)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey2)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingAPIKey(server.URL, "john", "King of the North", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	authenticator.getTokenData().Expiration = GetCurrentTime() + 3600
	authenticator.getTokenData().RefreshTime = GetCurrentTime() - 720

	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Wait for the background thread to finish.
	time.Sleep(5 * time.Second)
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey2, token)
	assert.NotNil(t, authenticator.getTokenData())
}

func TestCp4dBackgroundTokenRefreshAuthFailure(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "john", "snow", "")

		if firstCall {
			t.Logf("Sending back 200!")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernamePwd1)
			firstCall = false
		} else {
			t.Logf("Sending back 403!")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprintf(w, "Forbidden!")
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingPassword(server.URL, "john", "snow", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernamePwd1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now put the test in the "refresh window" where the token is not expired but still needs to be refreshed.
	authenticator.getTokenData().Expiration = GetCurrentTime() + 3600
	authenticator.getTokenData().RefreshTime = GetCurrentTime() - 720

	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet
	// expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernamePwd1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Wait for the background thread to finish
	time.Sleep(5 * time.Second)
	_, err = authenticator.GetToken()
	assert.NotNil(t, err)
	assert.Equal(t, "Forbidden!", err.Error())

	// We expect an AuthenticationError to be returned, so casting should work.
	_, ok := err.(*AuthenticationError)
	assert.True(t, ok)
}

func TestCp4dBackgroundTokenRefreshIdle(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "john", "", "King of the North")

		w.WriteHeader(http.StatusOK)

		if firstCall {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey1)
			firstCall = false
		} else {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey2)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingAPIKey(server.URL, "john", "King of the North", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// // Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Now simulate the client being idle for 10 minutes into the refresh time.
	authenticator.getTokenData().Expiration = GetCurrentTime() + 3600
	tenMinutesBeforeNow := GetCurrentTime() - 600
	authenticator.getTokenData().RefreshTime = tenMinutesBeforeNow

	// Authenticator should detect the need to refresh and request a new access token IN THE BACKGROUND when we call
	// GetToken() again. The immediate response should be the token which was already stored, since it's not yet expired.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// RefreshTime should have advanced by 1 minute from the current time.
	newRefreshTime := GetCurrentTime() + 60
	assert.Equal(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

	// In the next request, the RefreshTime should be unchanged and another thread
	// shouldn't be spawned to request another token once more since the first thread already spawned
	// a goroutine & refreshed the token.
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.NotNil(t, authenticator.getTokenData())
	assert.Equal(t, newRefreshTime, authenticator.getTokenData().RefreshTime)

	// Wait for the background thread to finish and verify both the RefreshTime & tokenData were updated
	time.Sleep(5 * time.Second)
	token, err = authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey2, token)
	assert.NotNil(t, authenticator.getTokenData())
	assert.NotEqual(t, newRefreshTime, authenticator.getTokenData().RefreshTime)
}

func TestCp4dDisableSSL(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifyAuthRequest(t, r, "mookie", "betts", "")

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernamePwd1)
	}))
	defer server.Close()

	authenticator := &CloudPakForDataAuthenticator{
		URL:                    server.URL,
		Username:               "mookie",
		Password:               "betts",
		DisableSSLVerification: true,
	}

	token, err := authenticator.GetToken()
	assert.Equal(t, cp4dUsernamePwd1, token)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator.Client)
	assert.NotNil(t, authenticator.Client.Transport)
	transport, ok := authenticator.Client.Transport.(*http.Transport)
	assert.Equal(t, true, ok)
	assert.NotNil(t, transport.TLSClientConfig)
	assert.Equal(t, true, transport.TLSClientConfig.InsecureSkipVerify)
}

func TestCp4dUserHeaders(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifyAuthRequest(t, r, "mookie", "", "King of the North")
		assert.Equal(t, "Value1", r.Header.Get("Header1"))
		assert.Equal(t, "Value2", r.Header.Get("Header2"))
		assert.Equal(t, "cp4d.cloud.ibm.com", r.Host)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey1)
	}))
	defer server.Close()

	headers := make(map[string]string)
	headers["Header1"] = "Value1"
	headers["Header2"] = "Value2"
	headers["Host"] = "cp4d.cloud.ibm.com"

	authenticator := &CloudPakForDataAuthenticator{
		URL:      server.URL,
		Username: "mookie",
		APIKey:   "King of the North",
		Headers:  headers,
	}

	token, err := authenticator.GetToken()
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.Nil(t, err)
}

func TestNewCloudPakForDataAuthenticatorFromMap(t *testing.T) {
	_, err := newCloudPakForDataAuthenticatorFromMap(nil)
	assert.NotNil(t, err)

	var props = map[string]string{
		PROPNAME_AUTH_URL: "cp4d-url",
	}
	_, err = newCloudPakForDataAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_AUTH_URL: "cp4d-url",
		PROPNAME_USERNAME: "mookie",
	}
	_, err = newCloudPakForDataAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_AUTH_URL: "cp4d-url",
		PROPNAME_PASSWORD: "betts",
	}
	_, err = newCloudPakForDataAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_AUTH_URL: "cp4d-url",
		PROPNAME_APIKEY:   "my_apikey",
	}
	_, err = newCloudPakForDataAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_AUTH_URL:         "cp4d-url",
		PROPNAME_USERNAME:         "mookie",
		PROPNAME_PASSWORD:         "betts",
		PROPNAME_AUTH_DISABLE_SSL: "true",
	}
	authenticator, err := newCloudPakForDataAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, "cp4d-url", authenticator.URL)
	assert.Equal(t, "mookie", authenticator.Username)
	assert.Equal(t, "betts", authenticator.Password)
	assert.Empty(t, authenticator.APIKey)
	assert.Equal(t, true, authenticator.DisableSSLVerification)

	props = map[string]string{
		PROPNAME_AUTH_URL:         "cp4d-url",
		PROPNAME_USERNAME:         "mookie",
		PROPNAME_APIKEY:           "my_apikey",
		PROPNAME_AUTH_DISABLE_SSL: "true",
	}
	authenticator, err = newCloudPakForDataAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, "cp4d-url", authenticator.URL)
	assert.Equal(t, "mookie", authenticator.Username)
	assert.Empty(t, authenticator.Password)
	assert.Equal(t, "my_apikey", authenticator.APIKey)
	assert.Equal(t, true, authenticator.DisableSSLVerification)
}

func TestCp4dGetTokenTimeoutError(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "john", "", "King of the North")

		w.WriteHeader(http.StatusOK)

		if firstCall {
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey1)
			firstCall = false
		} else {
			time.Sleep(3 * time.Second)
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernameApikey2)
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingAPIKey(server.URL, "john", "King of the North", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernameApikey1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Force expiration and verify that we got a timeout error
	authenticator.getTokenData().Expiration = GetCurrentTime() - 3600

	// Set the client timeout to something very low
	authenticator.Client.Timeout = time.Second * 2
	token, err = authenticator.GetToken()
	assert.Equal(t, "", token)
	assert.NotNil(t, err)
	assert.NotNil(t, err.Error())
	// We don't expect a AuthenticateError to be returned, so casting should fail
	_, ok := err.(*AuthenticationError)
	assert.False(t, ok)
}

func TestCp4dGetTokenServerError(t *testing.T) {
	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)

	firstCall := true
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		verifyAuthRequest(t, r, "john", "snow", "")

		if firstCall {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{ "_messageCode_":"200", "message":"success", "token":"%s"}`, cp4dUsernamePwd1)
			firstCall = false
		} else {
			w.WriteHeader(http.StatusGatewayTimeout)
			fmt.Fprintf(w, "Gateway Timeout")
		}
	}))
	defer server.Close()

	authenticator, err := NewCloudPakForDataAuthenticatorUsingPassword(server.URL, "john", "snow", false, nil)
	assert.Nil(t, err)
	assert.Nil(t, authenticator.getTokenData())

	// Force the first fetch and verify we got the first access token.
	token, err := authenticator.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, cp4dUsernamePwd1, token)
	assert.NotNil(t, authenticator.getTokenData())

	// Force expiration and verify that we got a server error.
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
	assert.NotNil(t, rawResult)
	assert.Equal(t, "Gateway Timeout", authError.Error())
	assert.Equal(t, []byte("Gateway Timeout"), rawResult)
	assert.Equal(t, http.StatusGatewayTimeout, statusCode)
	assert.Empty(t, token)
}

//
// In order to test with a live CP4D server, create file "cp4dtest.env" in the project root.
// It should look like this:
//
// CP4DTEST1_AUTH_URL=<url>   e.g. https://cpd350-cpd-cpd350.apps.wml-kf-cluster.os.fyre.ibm.com/icp4d-api
// CP4DTEST1_AUTH_TYPE=cp4d
// CP4DTEST1_USERNAME=<username>
// CP4DTEST1_PASSWORD=<password>
// CP4DTEST1_AUTH_DISABLE_SSL=true
//
// CP4DTEST2_AUTH_URL=<url>   e.g. https://cpd350-cpd-cpd350.apps.wml-kf-cluster.os.fyre.ibm.com/icp4d-api
// CP4DTEST2_AUTH_TYPE=cp4d
// CP4DTEST2_USERNAME=<username>
// CP4DTEST2_APIKEY=<apikey>
// CP4DTEST2_AUTH_DISABLE_SSL=true
//
// Then uncomment the function below, then run these commands:
// cd v<major-version>/core
// go test -v -tags=auth -run=TestCp4dLiveTokenServer
//

// func TestCp4dLiveTokenServer(t *testing.T) {
//	GetLogger().SetLogLevel(cp4dAuthTestLogLevel)
//
// 	var request *http.Request
// 	var err error
// 	var authHeader string

// 	// Get two cp4d authenticators from the environment.
// 	// "cp4dtest1" uses username/password
// 	// "cp4dtest2" uses username/apikey
// 	os.Setenv("IBM_CREDENTIALS_FILE", "../../cp4dtest.env")

// 	auth1, err := GetAuthenticatorFromEnvironment("cp4dtest1")
// 	assert.Nil(t, err)
// 	assert.NotNil(t, auth1)

// 	auth2, err := GetAuthenticatorFromEnvironment("cp4dtest2")
// 	assert.Nil(t, err)
// 	assert.NotNil(t, auth2)

// 	// Create a new Request object.
// 	builder, err := NewRequestBuilder("GET").ResolveRequestURL("https://localhost/placeholder/url", "", nil)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, builder)

// 	request, _ = builder.Build()
// 	assert.NotNil(t, request)
// 	err = auth1.Authenticate(request)
// 	assert.Nil(t, err)

// 	authHeader = request.Header.Get("Authorization")
// 	assert.NotEmpty(t, authHeader)
// 	assert.True(t, strings.HasPrefix(authHeader, "Bearer "))
// 	t.Logf("Authorization: %s\n", authHeader)

// 	request, _ = builder.Build()
// 	assert.NotNil(t, request)
// 	err = auth2.Authenticate(request)
// 	assert.Nil(t, err)

// 	authHeader = request.Header.Get("Authorization")
// 	assert.NotEmpty(t, authHeader)
// 	assert.True(t, strings.HasPrefix(authHeader, "Bearer "))
// 	t.Logf("Authorization: %s\n", authHeader)
// }
