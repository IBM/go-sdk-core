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
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/go-openapi/strfmt"

	"testing"

	assert "github.com/stretchr/testify/assert"
)

const (
	// To enable debug logging during test execution, set this to "LevelDebug"
	vpcauthTestLogLevel              LogLevel = LevelError
	vpcauthMockIAMProfileCRN         string   = "crn:iam-profile:123"
	vpcauthMockIAMProfileID          string   = "iam-id-123"
	vpcauthMockURL                   string   = "http://vpc.metadata.service.com"
	vpcauthTestAccessToken1          string   = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	vpcauthTestAccessToken2          string   = "3yJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
	vpcauthTestInstanceIdentityToken string   = "Xj7Gle500MachEOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJ1c2VybmFtZSI6ImhlbGxvIiwicm9sZSI6InVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhZG1pbmlzdHJhdG9yIiwiZGVwbG95bWVudF9hZG1pbiJdLCJzdWIiOiJoZWxsbyIsImlzcyI6IkpvaG4iLCJhdWQiOiJEU1giLCJ1aWQiOiI5OTkiLCJpYXQiOjE1NjAyNzcwNTEsImV4cCI6MTU2MDI4MTgxOSwianRpIjoiMDRkMjBiMjUtZWUyZC00MDBmLTg2MjMtOGNkODA3MGI1NDY4In0.cIodB4I6CCcX8vfIImz7Cytux3GpWyObt9Gkur5g1QI"
)

//
// Tests involving the construction of authenticators.
//

func TestVpcAuthCtorErrors(t *testing.T) {
	var err error
	var auth *VpcInstanceAuthenticator

	// Error: both IAMProfileCRN and IBMProfileID are specified
	auth, err = NewVpcInstanceAuthenticatorBuilder().
		SetIAMProfileCRN(vpcauthMockIAMProfileCRN).
		SetIAMProfileID(vpcauthMockIAMProfileID).Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}

func TestVpcAuthCtorSuccess(t *testing.T) {
	var err error
	var auth *VpcInstanceAuthenticator

	// 1. No properties (default configuration)
	auth, err = NewVpcInstanceAuthenticatorBuilder().Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, "", auth.IAMProfileCRN)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)

	// 2. only IAMProfileCRN
	auth, err = NewVpcInstanceAuthenticatorBuilder().
		SetIAMProfileCRN(vpcauthMockIAMProfileCRN).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, vpcauthMockIAMProfileCRN, auth.IAMProfileCRN)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)

	// 2. only IAMProfileID
	auth, err = NewVpcInstanceAuthenticatorBuilder().
		SetIAMProfileID(vpcauthMockIAMProfileID).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, "", auth.IAMProfileCRN)
	assert.Equal(t, vpcauthMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)

	// 3. only URL
	auth, err = NewVpcInstanceAuthenticatorBuilder().
		SetURL(vpcauthMockURL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, "", auth.IAMProfileCRN)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, vpcauthMockURL, auth.URL)

	// 4. IAMProfileID and URL
	auth, err = NewVpcInstanceAuthenticatorBuilder().
		SetIAMProfileID(vpcauthMockIAMProfileID).
		SetURL(vpcauthMockURL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, "", auth.IAMProfileCRN)
	assert.Equal(t, vpcauthMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, vpcauthMockURL, auth.URL)
}

func TestVpcAuthCtorFromMapErrors(t *testing.T) {
	var err error
	var auth *VpcInstanceAuthenticator
	var configProps map[string]string

	// Error: nil config map
	auth, err = newVpcInstanceAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: both IAMProfileCRN and IAMProfileID specified
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_CRN: vpcauthMockIAMProfileCRN,
		PROPNAME_IAM_PROFILE_ID:  vpcauthMockIAMProfileID,
	}
	auth, err = newVpcInstanceAuthenticatorFromMap(configProps)
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}

func TestVpcAuthCtorFromMapSuccess(t *testing.T) {
	var err error
	var auth *VpcInstanceAuthenticator
	var configProps map[string]string

	// 1. Default configuration (no properties)
	configProps = map[string]string{}
	auth, err = newVpcInstanceAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, "", auth.IAMProfileCRN)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)

	// 2. only IAMProfileCRN
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_CRN: vpcauthMockIAMProfileCRN,
	}
	auth, err = newVpcInstanceAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, vpcauthMockIAMProfileCRN, auth.IAMProfileCRN)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)

	// 3. only IAMProfileID
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_ID: vpcauthMockIAMProfileID,
	}
	auth, err = newVpcInstanceAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, "", auth.IAMProfileCRN)
	assert.Equal(t, vpcauthMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, "", auth.URL)

	// 4. only URL
	configProps = map[string]string{
		PROPNAME_AUTH_URL: vpcauthMockURL,
	}
	auth, err = newVpcInstanceAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, "", auth.IAMProfileCRN)
	assert.Equal(t, "", auth.IAMProfileID)
	assert.Equal(t, vpcauthMockURL, auth.URL)

	// 5. IAMProfileID and URL
	configProps = map[string]string{
		PROPNAME_IAM_PROFILE_ID: vpcauthMockIAMProfileID,
		PROPNAME_AUTH_URL:       vpcauthMockURL,
	}
	auth, err = newVpcInstanceAuthenticatorFromMap(configProps)
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, AUTHTYPE_VPC, auth.AuthenticationType())
	assert.Equal(t, "", auth.IAMProfileCRN)
	assert.Equal(t, vpcauthMockIAMProfileID, auth.IAMProfileID)
	assert.Equal(t, vpcauthMockURL, auth.URL)
}

func TestVpcAuthDefaultURL(t *testing.T) {
	auth := &VpcInstanceAuthenticator{}
	s := auth.url()
	assert.Equal(t, s, vpcauthDefaultIMSEndpoint)
	assert.Equal(t, auth.URL, vpcauthDefaultIMSEndpoint)
}

// startMockVPCServer will start a mock server endpoint that supports both of the
// VPC Instance Metadata Service operations that the authenticator will need to invoke
// (create_access_token and create_iam_token).
// The "scenario" input parameter is simply a string passed in by individual testcases to
// indicate the specific behavior that is needed by that testcase.
func startMockVPCServer(t *testing.T, scenario string) *httptest.Server {
	// In our handler function below, we keep a count of the number of invocations of
	// the "create_iam_token" operation so we can simulate the use of different
	// IAM access tokens.
	var iamTokenCount int = 0

	// For calls to the 'create_iam_token' operation we expect the Authorization header
	// to contain the instance identity token.
	expectedAuthorizationHeader := "Bearer " + vpcauthTestInstanceIdentityToken

	// Create the mock server.
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		operationPath := req.URL.EscapedPath()

		// Process the request according to the operation being invoked.
		switch operationPath {
		case vpcauthOperationPathCreateAccessToken:
			// Process the 'create_access_token' operation invocation.

			// Verify some parts of the request.
			assert.Equal(t, PUT, req.Method)
			assert.NotEmpty(t, req.URL.Query().Get("version"))
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Accept"))
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Content-Type"))
			assert.Equal(t, vpcauthMetadataFlavor, req.Header.Get("Metadata-Flavor"))

			// Simulate a timeout situation by sleeping for a few seconds
			// while the client will use a short timeout value.
			if scenario == "vpc-token-timeout" {
				time.Sleep(2 * time.Second)
			}

			if scenario == "vpc-token-fail" {
				// Force a BadRequest.
				res.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(res, `Sorry, operation 'create_access_token' failed!`)
			} else {
				// Success scenario.

				// This struct models the request body for the 'create_access_token' operation.
				type createAccessTokenRequestBody struct {
					// Time in seconds before the access token expires.
					ExpiresIn *int64 `json:"expires_in,omitempty"`
				}

				// Unmarshal the request body.
				requestBody := &createAccessTokenRequestBody{}
				_ = json.NewDecoder(req.Body).Decode(requestBody)
				defer req.Body.Close()
				assert.NotNil(t, requestBody.ExpiresIn)

				createdAt := time.Now()
				expiresAt := createdAt.Add(300 * time.Second)
				dtCreatedAt := strfmt.DateTime(createdAt)
				dtExpiresAt := strfmt.DateTime(expiresAt)
				response := &vpcTokenResponse{
					AccessToken: StringPtr(vpcauthTestInstanceIdentityToken),
					CreatedAt:   &dtCreatedAt,
					ExpiresAt:   &dtExpiresAt,
					ExpiresIn:   Int64Ptr(300),
				}

				res.WriteHeader(http.StatusOK)

				buf, err := json.Marshal(response)
				assert.Nil(t, err)
				fmt.Fprintf(res, "%s", (string(buf)))
			}

		case vpcauthOperationPathCreateIamToken:
			// Process the 'create_iam_token' operation invocation.

			// Verify some parts of the request.
			assert.Equal(t, POST, req.Method)
			assert.NotEmpty(t, req.URL.Query().Get("version"))
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Accept"))
			assert.Equal(t, APPLICATION_JSON, req.Header.Get("Content-Type"))
			assert.Equal(t, expectedAuthorizationHeader, req.Header.Get("Authorization"))

			// Models a trusted profile (includes both CRN and ID fields).
			type trustedProfileIdentity struct {
				// The unique identifier for this trusted profile.
				ID *string `json:"id,omitempty"`

				// The CRN for this trusted profile.
				CRN *string `json:"crn,omitempty"`
			}

			// Models the request body for the 'create_iam_token' operation.
			type createIamTokenRequestBody struct {
				TrustedProfile *trustedProfileIdentity `json:"trusted_profile,omitempty"`
			}

			// These specific scenarios can be used to perform specific validation of the request body.
			requestBody := &createIamTokenRequestBody{}
			_ = json.NewDecoder(req.Body).Decode(requestBody)
			defer req.Body.Close()

			switch scenario {
			case "profile-none":
				assert.NotNil(t, requestBody)
				assert.Nil(t, requestBody.TrustedProfile)

			case "profile-crn":
				assert.NotNil(t, requestBody)
				assert.NotNil(t, requestBody.TrustedProfile)
				assert.NotNil(t, requestBody.TrustedProfile.CRN)
				assert.Nil(t, requestBody.TrustedProfile.ID)
				assert.Equal(t, vpcauthMockIAMProfileCRN, *requestBody.TrustedProfile.CRN)

			case "profile-id":
				assert.NotNil(t, requestBody)
				assert.NotNil(t, requestBody.TrustedProfile)
				assert.Nil(t, requestBody.TrustedProfile.CRN)
				assert.NotNil(t, requestBody.TrustedProfile.ID)
				assert.Equal(t, vpcauthMockIAMProfileID, *requestBody.TrustedProfile.ID)

			default:
			}

			// Determine which IAM access token should be returned.
			// We'll return the first access token value the first time the operation is called,
			// then the second access token for subequent invocations.
			var accessToken *string
			iamTokenCount++
			if iamTokenCount == 1 {
				accessToken = StringPtr(vpcauthTestAccessToken1)
			} else {
				accessToken = StringPtr(vpcauthTestAccessToken2)
			}

			// Simulate timeout situations if requested by sleeping for a few seconds
			// while the client will use a short timeout value.
			if scenario == "iam-token1-timeout" && iamTokenCount == 1 {
				time.Sleep(2 * time.Second)
			} else if scenario == "iam-token2-timeout" && iamTokenCount > 1 {
				time.Sleep(2 * time.Second)
			}

			// Determine what to send back in the response.
			if scenario == "iam-token1-fail" && iamTokenCount == 1 {
				// Simulate a failure when the 1st token is requested.
				res.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(res, `Sorry, operation 'create_iam_token' failed!`)
			} else if scenario == "iam-token2-fail" && iamTokenCount > 1 {
				// Simulate a failure when the 2nd token is requested.
				res.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(res, "Error, operation 'create_iam_token' failed!")
			} else {
				// Success scenario.
				res.WriteHeader(http.StatusOK)
				createdAt := time.Now()
				expiresAt := createdAt.Add(3600 * time.Second)
				dtCreatedAt := strfmt.DateTime(createdAt)
				dtExpiresAt := strfmt.DateTime(expiresAt)
				response := &vpcTokenResponse{
					AccessToken: accessToken,
					CreatedAt:   &dtCreatedAt,
					ExpiresAt:   &dtExpiresAt,
					ExpiresIn:   Int64Ptr(3600),
				}

				buf, err := json.Marshal(response)
				assert.Nil(t, err)
				fmt.Fprintf(res, "%s", (string(buf)))
			}

		default:
			// Internal testcase error - should never get here :)
			res.WriteHeader(http.StatusNotFound)
			msg := "Unknown operation path: " + operationPath
			fmt.Fprintf(res, "%s", msg)
			assert.Fail(t, msg)
		}
	}))
	return server
}

//
// Tests involving the authenticator's internal "retrieveInstanceIdentityToken" method.
//

func TestVpcAuthRetrieveVpcTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "success")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		URL: server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	vpcToken, err := auth.retrieveInstanceIdentityToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestInstanceIdentityToken, vpcToken)
}

func assertAuthError(t *testing.T, err error) {
	authErr, ok := err.(*AuthenticationError)
	assert.True(t, ok)
	assert.NotNil(t, authErr)
	assert.EqualValues(t, authErr, err)
	assert.Equal(t, err.Error(), authErr.Error())
}

func TestVpcAuthRetrieveVpcTokenFail1(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "vpc-token-fail")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		URL: server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	vpcToken, err := auth.retrieveInstanceIdentityToken()
	assert.Empty(t, vpcToken)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthRetrieveVpcTokenFail2(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	// Force an error while resolving the service URL.
	auth := &VpcInstanceAuthenticator{
		URL: "123:badpath",
	}

	vpcToken, err := auth.retrieveInstanceIdentityToken()
	assert.Empty(t, vpcToken)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthRetrieveVpcTokenTimeout(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "vpc-token-timeout")
	defer server.Close()

	shortTimeoutClient := &http.Client{
		Timeout: 1 * time.Second,
	}
	auth, err := NewVpcInstanceAuthenticatorBuilder().
		SetURL(server.URL).
		SetClient(shortTimeoutClient).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	vpcToken, err := auth.retrieveInstanceIdentityToken()
	assert.Empty(t, vpcToken)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

//
// Tests involving the authenticator's internal "retrieveIamAccessToken" method.
//

func TestVpcAuthRetrieveIamTokenSuccessProfileNone(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "profile-none")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		URL: server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	iamTokenServerResponse, err := auth.retrieveIamAccessToken(vpcauthTestInstanceIdentityToken)
	assert.Nil(t, err)
	assert.NotNil(t, iamTokenServerResponse)
	assert.Equal(t, vpcauthTestAccessToken1, iamTokenServerResponse.AccessToken)
}

func TestVpcAuthRetrieveIamTokenSuccessProfileCRN(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "profile-crn")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		URL:           server.URL,
		IAMProfileCRN: vpcauthMockIAMProfileCRN,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	iamTokenServerResponse, err := auth.retrieveIamAccessToken(vpcauthTestInstanceIdentityToken)
	assert.Nil(t, err)
	assert.NotNil(t, iamTokenServerResponse)
	assert.Equal(t, vpcauthTestAccessToken1, iamTokenServerResponse.AccessToken)
}

func TestVpcAuthRetrieveIamTokenSuccessProfileID(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "profile-id")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		URL:          server.URL,
		IAMProfileID: vpcauthMockIAMProfileID,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	iamTokenServerResponse, err := auth.retrieveIamAccessToken(vpcauthTestInstanceIdentityToken)
	assert.Nil(t, err)
	assert.NotNil(t, iamTokenServerResponse)
	assert.Equal(t, vpcauthTestAccessToken1, iamTokenServerResponse.AccessToken)

	iamTokenServerResponse, err = auth.retrieveIamAccessToken(vpcauthTestInstanceIdentityToken)
	assert.Nil(t, err)
	assert.NotNil(t, iamTokenServerResponse)
	assert.Equal(t, vpcauthTestAccessToken2, iamTokenServerResponse.AccessToken)
}

func TestVpcAuthRetrieveIamTokenFail1(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "iam-token1-fail")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		URL: server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	iamTokenServerResponse, err := auth.retrieveIamAccessToken(vpcauthTestInstanceIdentityToken)
	assert.Nil(t, iamTokenServerResponse)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthRetrieveIamTokenFail2(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	// Force an error while resolving the service URL.
	auth := &VpcInstanceAuthenticator{
		URL: "123:badpath",
	}

	iamTokenServerResponse, err := auth.retrieveIamAccessToken(vpcauthTestInstanceIdentityToken)
	assert.Nil(t, iamTokenServerResponse)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthRetrieveIamTokenTimeout(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "iam-token1-timeout")
	defer server.Close()

	// Construct an authenticator with a short timeout.
	shortTimeoutClient := &http.Client{
		Timeout: 1 * time.Second,
	}
	auth, err := NewVpcInstanceAuthenticatorBuilder().
		SetURL(server.URL).
		SetClient(shortTimeoutClient).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	iamTokenServerResponse, err := auth.retrieveIamAccessToken(vpcauthTestInstanceIdentityToken)
	assert.Nil(t, iamTokenServerResponse)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

//
// Tests involving the authenticator's "GetToken" method.
//

func TestVpcAuthGetTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "profile-crn")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		IAMProfileCRN: vpcauthMockIAMProfileCRN,
		URL:           server.URL,
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
	assert.Equal(t, vpcauthTestAccessToken1, accessToken)
	assert.Equal(t, vpcauthTestAccessToken1, auth.getTokenData().AccessToken)

	// Call synchronizedRequestToken() to make sure we get back a nil error response.
	assert.True(t, auth.getTokenData().isTokenValid())
	err = auth.synchronizedRequestToken()
	assert.Nil(t, err)

	// Call GetToken() again and verify that we get the cached value.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestAccessToken1, accessToken)

	// Force expiration and verify that GetToken() fetched the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, vpcauthTestAccessToken2, accessToken)
	assert.Equal(t, vpcauthTestAccessToken2, auth.getTokenData().AccessToken)
}

func TestVpcAuthGetTokenFail(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "vpc-token-fail")
	defer server.Close()

	auth, err := NewVpcInstanceAuthenticatorBuilder().
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Force the first fetch and verify we got back an error.
	accessToken, err := auth.GetToken()
	assert.NotNil(t, err)
	assert.Empty(t, accessToken)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthGetTokenTimeout(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "iam-token1-timeout")
	defer server.Close()

	// Construct an authenticator with a short timeout.
	shortTimeoutClient := &http.Client{
		Timeout: 1 * time.Second,
	}
	auth, err := NewVpcInstanceAuthenticatorBuilder().
		SetURL(server.URL).
		SetClient(shortTimeoutClient).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Force the first fetch and verify we got back an error.
	accessToken, err := auth.GetToken()
	assert.NotNil(t, err)
	assert.Empty(t, accessToken)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthGetTokenRefreshSuccess(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "profile-id")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		IAMProfileID: vpcauthMockIAMProfileID,
		URL:          server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestAccessToken1, accessToken)

	// Now simulate being in the refresh window where the token is not expired but still needs to be refreshed.
	auth.getTokenData().RefreshTime = GetCurrentTime() - 1

	// Authenticator should detect the need to get a new access token in the background but use the current
	// cached access token for this next GetToken() call.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestAccessToken1, accessToken)

	// Wait for the background thread to finish.
	// Then call GetToken() again and we should now have the second access token.
	time.Sleep(1 * time.Second)
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestAccessToken2, accessToken)
}

func TestVpcAuthGetTokenRefreshFail(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "iam-token2-fail")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		URL: server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestAccessToken1, accessToken)

	// Now simulate being in the refresh window where the token is not expired but still needs to be refreshed.
	auth.getTokenData().RefreshTime = GetCurrentTime() - 1

	// Authenticator should detect the need to get a new access token in the background but use the current
	// cached access token for this next GetToken() call.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestAccessToken1, accessToken)

	// Wait for the background thread to finish.
	time.Sleep(1 * time.Second)

	// The background token refresh triggered by the previous GetToken() call above failed,
	// but the authenticator is still holding a valid, unexpired access token,
	// so this next GetToken() call should succeed and return the first access token
	// that we had previously cached.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestAccessToken1, accessToken)

	// Next, simulate the expiration of the token, then we should expect
	// an error from GetToken().
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	accessToken, err = auth.GetToken()
	assert.NotNil(t, err)
	assert.Empty(t, accessToken)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthGetTokenRefreshTimeout(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "iam-token2-timeout")
	defer server.Close()

	auth, err := NewVpcInstanceAuthenticatorBuilder().
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, vpcauthTestAccessToken1, accessToken)

	// Next, force the expiration of the current cached token and configure the authenticator's
	// client with a short timeout.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	auth.Client.Timeout = 1 * time.Second
	accessToken, err = auth.GetToken()
	assert.NotNil(t, err)
	assert.Empty(t, accessToken)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

//
// Tests involving the authenticator's "RequestToken" method.
//

func TestVpcAuthRequestTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "profile-id")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		IAMProfileID: vpcauthMockIAMProfileID,
		URL:          server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Verify that RequestToken() returns a response with a valid access token.
	tokenResponse, err := auth.RequestToken()
	assert.Nil(t, err)
	assert.NotNil(t, tokenResponse)
	assert.Equal(t, vpcauthTestAccessToken1, tokenResponse.AccessToken)
}

func TestVpcAuthRequestTokenFail(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "iam-token1-fail")
	defer server.Close()

	auth := &VpcInstanceAuthenticator{
		URL: server.URL,
	}
	err := auth.Validate()
	assert.Nil(t, err)

	// Verify that RequestToken() returned an error.
	tokenResponse, err := auth.RequestToken()
	assert.NotNil(t, err)
	assert.Nil(t, tokenResponse)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthRequestTokenTimeout(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "iam-token1-timeout")
	defer server.Close()

	// Construct an authenticator with a short timeout.
	shortTimeoutClient := &http.Client{
		Timeout: 1 * time.Second,
	}
	auth, err := NewVpcInstanceAuthenticatorBuilder().
		SetURL(server.URL).
		SetClient(shortTimeoutClient).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Verify that RequestToken() returned an error.
	accessToken, err := auth.RequestToken()
	assert.NotNil(t, err)
	assert.Empty(t, accessToken)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

//
// Tests involving the authenticator's "Authenticate" method.
//

func TestVpcAuthAuthenticateSuccess(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "profile-none")
	defer server.Close()

	auth, err := NewVpcInstanceAuthenticatorBuilder().
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
	assert.Equal(t, "Bearer "+vpcauthTestAccessToken1, authHeader)

	// Call Authenticate again to make sure we used the cached access token.
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader = request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+vpcauthTestAccessToken1, authHeader)

	// Force expiration and verify that Authenticate() fetched the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader = request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+vpcauthTestAccessToken2, authHeader)
}

func TestVpcAuthAuthenticateFailVpcToken(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "vpc-token-fail")
	defer server.Close()

	auth, err := NewVpcInstanceAuthenticatorBuilder().
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

	// Verify that Authenticate() returned an error.
	err = auth.Authenticate(request)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}

func TestVpcAuthAuthenticateFailIamToken(t *testing.T) {
	GetLogger().SetLogLevel(vpcauthTestLogLevel)

	server := startMockVPCServer(t, "iam-token1-fail")
	defer server.Close()

	auth, err := NewVpcInstanceAuthenticatorBuilder().
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

	// Verify that Authenticate() returned an error.
	err = auth.Authenticate(request)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assertAuthError(t, err)
}
