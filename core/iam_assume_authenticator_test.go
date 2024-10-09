//go:build all || slow || auth

package core

// (C) Copyright IBM Corp. 2024.
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
	"strings"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

var (
	// To enable debug logging during test execution, set this to "LevelDebug"
	iamAssumeTestLogLevel LogLevel = LevelError

	iamAssumeMockProfileCRN    = "mock-profile-crn"
	iamAssumeMockProfileID     = "mock-profile-id"
	iamAssumeMockProfileName   = "mock-profile-name"
	iamAssumeMockAccountID     = "mock-account-id"
	iamAssumeMockApiKey        = "mock-apikey"
	iamAssumeMockClientID      = "bx"
	iamAssumeMockClientSecret  = "bx"
	iamAssumeMockURL           = "https://mock.iam.com"
	iamAssumeMockScope         = "scope1,scope2"
	iamAssumeMockUserToken1    = "eyJraWQiOiIyMDI0MDkwMjA4NDIiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC01NTAwMDBFRDZKIiwiaWQiOiJJQk1pZC01NTAwMDBFRDZKIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiYmY2YTA0NDQtZDk3YS00OWYxLTkzNTgtZmFkMGRmODZiNmRiIiwiaWRlbnRpZmllciI6IjU1MDAwMEVENkoiLCJnaXZlbl9uYW1lIjoiUGhpbCIsImZhbWlseV9uYW1lIjoiQWRhbXMiLCJuYW1lIjoiUGhpbCBBZGFtcyIsImVtYWlsIjoicGhpbF9hZGFtc0B1cy5pYm0uY29tIiwic3ViIjoicGhpbF9hZGFtc0B1cy5pYm0uY29tIiwiYXV0aG4iOnsic3ViIjoicGhpbF9hZGFtc0B1cy5pYm0uY29tIiwiaWFtX2lkIjoiSUJNaWQtNTUwMDAwRUQ2SiIsIm5hbWUiOiJQaGlsIEFkYW1zIiwiZ2l2ZW5fbmFtZSI6IlBoaWwiLCJmYW1pbHlfbmFtZSI6IkFkYW1zIiwiZW1haWwiOiJwaGlsX2FkYW1zQHVzLmlibS5jb20ifSwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiOGI0ZGEzNzNjNmY4NDk0ODg4OTg2ZmNjNDk5MmVhMmQiLCJmcm96ZW4iOnRydWV9LCJpYXQiOjE3MjczMDE1NjQsImV4cCI6MTcyNzMwNTE2NCwiaXNzIjoiaHR0cHM6Ly9pYW0uY2xvdWQuaWJtLmNvbS9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIG9wZW5pZCIsImNsaWVudF9pZCI6ImRlZmF1bHQiLCJhY3IiOjEsImFtciI6WyJwd2QiXX0.NHyw3JedZdHawuBTbFzfdYu5ESweUGXGOktmqUEB2plRmkleZQlyVZv1oXN2XWfTgXxr4er6LPGiZvglGCIKeABg557wZSg_kkgBCd2QABVJTJTcuQXC8zzgCKoKiIunHaBKzT--lvix-wGrlBb6D8zhcLBND1Xp5vXaGlLA9IIfe_HEEsmcUxqGCtQA5zb18dvQQFvXc_3ZVk5jM8pGNJXBO8R9ZAE_yA5Jc3wszSmclqhXWbmH3zxZfKuXbsPxsRJQUk4rEAvCUfQNBuFVhJkYQubxNKcOVOf67Up7-IxuxH7P9NBgqTYcHXKDx38foNpCX0ssrEgq2b36AQI2gA"                                                                                                // #nosec
	iamAssumeMockUserToken2    = "eyJraWQiOiIyMDI0MDkwMjA4NDIiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJJQk1pZC01NTAwMDBFRDZKIiwiaWQiOiJJQk1pZC01NTAwMDBFRDZKIiwicmVhbG1pZCI6IklCTWlkIiwianRpIjoiYmVmZjIyZjctY2Q2OC00MDViLWEyMzYtYmI0OTJlYmE0ZGRhIiwiaWRlbnRpZmllciI6IjU1MDAwMEVENkoiLCJnaXZlbl9uYW1lIjoiUGhpbCIsImZhbWlseV9uYW1lIjoiQWRhbXMiLCJuYW1lIjoiUGhpbCBBZGFtcyIsImVtYWlsIjoicGhpbF9hZGFtc0B1cy5pYm0uY29tIiwic3ViIjoicGhpbF9hZGFtc0B1cy5pYm0uY29tIiwiYXV0aG4iOnsic3ViIjoicGhpbF9hZGFtc0B1cy5pYm0uY29tIiwiaWFtX2lkIjoiSUJNaWQtNTUwMDAwRUQ2SiIsIm5hbWUiOiJQaGlsIEFkYW1zIiwiZ2l2ZW5fbmFtZSI6IlBoaWwiLCJmYW1pbHlfbmFtZSI6IkFkYW1zIiwiZW1haWwiOiJwaGlsX2FkYW1zQHVzLmlibS5jb20ifSwiYWNjb3VudCI6eyJ2YWxpZCI6dHJ1ZSwiYnNzIjoiOGI0ZGEzNzNjNmY4NDk0ODg4OTg2ZmNjNDk5MmVhMmQiLCJmcm96ZW4iOnRydWV9LCJpYXQiOjE3MjczMDE3MzgsImV4cCI6MTcyNzMwNTMzOCwiaXNzIjoiaHR0cHM6Ly9pYW0uY2xvdWQuaWJtLmNvbS9pZGVudGl0eSIsImdyYW50X3R5cGUiOiJ1cm46aWJtOnBhcmFtczpvYXV0aDpncmFudC10eXBlOmFwaWtleSIsInNjb3BlIjoiaWJtIG9wZW5pZCIsImNsaWVudF9pZCI6ImRlZmF1bHQiLCJhY3IiOjEsImFtciI6WyJwd2QiXX0.Yi2zlgrwwxpt0XhC6jQrZnDNHoFt2cE9vY9W3tRBcNVGAmGN2pYqTcdwlKEVKjc7MtR-SfaiVPc_4iVpEfYNeG-ISXma7x-ZKvpUoo41fGUY7AzEH336FZcPpPoGnFfKPafUUXaEIHcwzIobRBxmIlMbXKwEiQEu1BBDxIUYXDP-wkLEJ95PB8gTAbrx8yrGVTFpp9mOvanePMzwHj7sQXZ3E0InVTBk4HDFSb51ggvor09rLTDtHU8WDdh4GNuRS76MURRpZ3aLWIEtgvUGwgmZxatxwJeLHxZqtfBzbXS4JhhQOl5vUg_4DavSA7luwZbdZYbZbj22KJGm0qo6Rg"                                                                                                // #nosec
	iamAssumeMockProfileToken1 = "eyJraWQiOiIyMDI0MDkwMjA4NDIiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJpYW0tUHJvZmlsZS03YWRmOWMwNi01NmZlLTQwODEtOTE2Yy1hYzg3MmFmYWUzZjQiLCJpZCI6ImlhbS1Qcm9maWxlLTdhZGY5YzA2LTU2ZmUtNDA4MS05MTZjLWFjODcyYWZhZTNmNCIsInJlYWxtaWQiOiJpYW0iLCJqdGkiOiIxY2NlMGU1Zi05YTk5LTQzOTktOGNmYi1mN2U4YmRlMTM4ZjYiLCJpZGVudGlmaWVyIjoiUHJvZmlsZS03YWRmOWMwNi01NmZlLTQwODEtOTE2Yy1hYzg3MmFmYWUzZjQiLCJuYW1lIjoiQXNzdW1lZFByb2ZpbGUxIiwic3ViIjoiUHJvZmlsZS03YWRmOWMwNi01NmZlLTQwODEtOTE2Yy1hYzg3MmFmYWUzZjQiLCJzdWJfdHlwZSI6IlByb2ZpbGUiLCJhdXRobiI6eyJzdWIiOiJwaGlsX2FkYW1zQHVzLmlibS5jb20iLCJpYW1faWQiOiJJQk1pZC01NTAwMDBFRDZKIiwibmFtZSI6IlBoaWwgQWRhbXMiLCJnaXZlbl9uYW1lIjoiUGhpbCIsImZhbWlseV9uYW1lIjoiQWRhbXMiLCJlbWFpbCI6InBoaWxfYWRhbXNAdXMuaWJtLmNvbSJ9LCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI4YjRkYTM3M2M2Zjg0OTQ4ODg5ODZmY2M0OTkyZWEyZCIsImZyb3plbiI6dHJ1ZX0sImlhdCI6MTcyNzMwMTU2NCwiZXhwIjoxNzI3MzA1MTYxLCJpc3MiOiJodHRwczovL2lhbS5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6YXNzdW1lIiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiZGVmYXVsdCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.VtMNv7gHScrnWfuHHRXxp62AYRSDY5_RQZw8Wdj-hgMX7qmgquaKSvfwTooGJyamuUl0WNNW6avrqU0TVebyc-Aci4e71NchJf1nSol0EIxYQum8LBBUfyMcOVLfuPSAdabEUTqLR1nh1oxrRlSAVt5hLSDnQ-2WS8OrAWG8fWEvACrzXhrPUF5Ko702V7Y-Gnksoz3nkDvLeoVx6jwF3izrJ-1NwGuMGNLfu8E3zSl9utbY4FSSvEheHii1h1QfNYl9FCJpMWCfwpJVCktKlOlP_9g-lirWMoJ_lEc2DA-Pl54Ozmos08G7DoOmgmrtxvUcGXSc7_FKhj77LDuo0g" // #nosec
	iamAssumeMockProfileToken2 = "eyJraWQiOiIyMDI0MDkwMjA4NDIiLCJhbGciOiJSUzI1NiJ9.eyJpYW1faWQiOiJpYW0tUHJvZmlsZS03YWRmOWMwNi01NmZlLTQwODEtOTE2Yy1hYzg3MmFmYWUzZjQiLCJpZCI6ImlhbS1Qcm9maWxlLTdhZGY5YzA2LTU2ZmUtNDA4MS05MTZjLWFjODcyYWZhZTNmNCIsInJlYWxtaWQiOiJpYW0iLCJqdGkiOiI4NTRhNjQ0Zi01MmY0LTRmNjMtYmE5Yy0yODBjYjkzYjE5MjkiLCJpZGVudGlmaWVyIjoiUHJvZmlsZS03YWRmOWMwNi01NmZlLTQwODEtOTE2Yy1hYzg3MmFmYWUzZjQiLCJuYW1lIjoiQXNzdW1lZFByb2ZpbGUxIiwic3ViIjoiUHJvZmlsZS03YWRmOWMwNi01NmZlLTQwODEtOTE2Yy1hYzg3MmFmYWUzZjQiLCJzdWJfdHlwZSI6IlByb2ZpbGUiLCJhdXRobiI6eyJzdWIiOiJwaGlsX2FkYW1zQHVzLmlibS5jb20iLCJpYW1faWQiOiJJQk1pZC01NTAwMDBFRDZKIiwibmFtZSI6IlBoaWwgQWRhbXMiLCJnaXZlbl9uYW1lIjoiUGhpbCIsImZhbWlseV9uYW1lIjoiQWRhbXMiLCJlbWFpbCI6InBoaWxfYWRhbXNAdXMuaWJtLmNvbSJ9LCJhY2NvdW50Ijp7InZhbGlkIjp0cnVlLCJic3MiOiI4YjRkYTM3M2M2Zjg0OTQ4ODg5ODZmY2M0OTkyZWEyZCIsImZyb3plbiI6dHJ1ZX0sImlhdCI6MTcyNzMwMTczOSwiZXhwIjoxNzI3MzA1MzM2LCJpc3MiOiJodHRwczovL2lhbS5jbG91ZC5pYm0uY29tL2lkZW50aXR5IiwiZ3JhbnRfdHlwZSI6InVybjppYm06cGFyYW1zOm9hdXRoOmdyYW50LXR5cGU6YXNzdW1lIiwic2NvcGUiOiJpYm0gb3BlbmlkIiwiY2xpZW50X2lkIjoiZGVmYXVsdCIsImFjciI6MSwiYW1yIjpbInB3ZCJdfQ.lr5HsElwOBMyyQ855KCPLeSXKYHImmogVpKzD_eGFI8kPWhd7lFslbC_6nALfehWpyMG4xCILg3eK-lGdkntVZ92mmKZKEzOd4GRm_bgIU_Ul0zyiZurnE8u5MDBSMp-sIQ5yPzBsDFxkAhm7f2Dt3TmjyUHGS8AIs_uQ8ldT0l0rj-rK6YjfF-hDsm414690dIbTaoSLUg679Qto0peiQ7HmE0RX91QkRqbKg7BKkIpsKxepnJoqlZuPBoKH8o8-dkqpW8ktm2e-Sk_eOJTiB07I0x1212gwuQFD_8Y7YYfzMqqtJmLiSwgHOnjHGSkqirfC_zA5rbjuI3a4yOr0g" // #nosec
)

// Tests involving the Builder
func TestIamAssumeAuthBuilderErrors(t *testing.T) {
	var err error
	var auth *IamAssumeAuthenticator

	// Error: no apikey
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey("").
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: invalid apikey
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey("{invalid-apikey}").
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: apikey and client-id set, but no client-secret
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetClientIDSecret(iamAssumeMockClientID, "").
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: apikey and client-secret set, but no client-id
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetClientIDSecret("", iamAssumeMockClientSecret).
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: no trusted profile specified
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())

	// Error: specify profile name with no account id
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileName(iamAssumeMockProfileName).
		Build()
	assert.NotNil(t, err)
	assert.Nil(t, auth)
	t.Logf("Expected error: %s", err.Error())
}

func TestIamAssumeAuthBuilderSuccess(t *testing.T) {
	var err error
	var auth *IamAssumeAuthenticator
	var expectedHeaders = map[string]string{
		"header1": "value1",
	}

	// Specify apikey and profile id.
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileID(iamAssumeMockProfileID).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, iamAssumeMockApiKey, auth.iamDelegate.ApiKey)
	assert.Empty(t, auth.iamDelegate.RefreshToken)
	assert.Empty(t, auth.iamDelegate.URL)
	assert.Empty(t, auth.iamDelegate.ClientId)
	assert.Empty(t, auth.iamDelegate.ClientSecret)
	assert.False(t, auth.iamDelegate.DisableSSLVerification)
	assert.Empty(t, auth.iamDelegate.Scope)
	assert.Nil(t, auth.iamDelegate.Headers)
	assert.Empty(t, auth.url)
	assert.Empty(t, auth.iamProfileCRN)
	assert.Equal(t, iamAssumeMockProfileID, auth.iamProfileID)
	assert.Empty(t, auth.iamProfileName)
	assert.Empty(t, auth.iamAccountID)
	assert.Equal(t, AUTHTYPE_IAM, auth.iamDelegate.AuthenticationType())
	assert.Equal(t, AUTHTYPE_IAM_ASSUME, auth.AuthenticationType())

	// Specify apikey and profile crn.
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileCRN(iamAssumeMockProfileCRN).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, iamAssumeMockProfileCRN, auth.iamProfileCRN)
	assert.Empty(t, auth.iamProfileID)
	assert.Empty(t, auth.iamProfileName)
	assert.Empty(t, auth.iamAccountID)

	// Specify apikey and profile name.
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileName(iamAssumeMockProfileName).
		SetIAMAccountID(iamAssumeMockAccountID).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Empty(t, auth.iamProfileCRN)
	assert.Empty(t, auth.iamProfileID)
	assert.Equal(t, iamAssumeMockProfileName, auth.iamProfileName)
	assert.Equal(t, iamAssumeMockAccountID, auth.iamAccountID)

	// Specify various IAM-related properties.
	auth, err = NewIamAssumeAuthenticatorBuilder().
		SetURL(iamAssumeMockURL).
		SetIAMProfileCRN(iamAssumeMockProfileCRN).
		SetApiKey(iamAssumeMockApiKey).
		SetClientIDSecret(iamAssumeMockClientID, iamAssumeMockClientSecret).
		SetDisableSSLVerification(true).
		SetScope(iamAssumeMockScope).
		SetHeaders(expectedHeaders).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)
	assert.Equal(t, iamAssumeMockURL, auth.url)
	assert.Equal(t, iamAssumeMockProfileCRN, auth.iamProfileCRN)
	assert.Empty(t, auth.iamProfileID)
	assert.Empty(t, auth.iamProfileName)
	assert.Empty(t, auth.iamAccountID)
	assert.True(t, auth.disableSSLVerification)
	assert.Equal(t, expectedHeaders, auth.headers)
	assert.Equal(t, iamAssumeMockApiKey, auth.iamDelegate.ApiKey)
	assert.Equal(t, iamAssumeMockURL, auth.iamDelegate.URL)
	assert.Equal(t, iamAssumeMockClientID, auth.iamDelegate.ClientId)
	assert.Equal(t, iamAssumeMockClientSecret, auth.iamDelegate.ClientSecret)
	assert.True(t, auth.iamDelegate.DisableSSLVerification)
	assert.Equal(t, iamAssumeMockScope, auth.iamDelegate.Scope)
	assert.Equal(t, expectedHeaders, auth.iamDelegate.Headers)
	assert.Equal(t, AUTHTYPE_IAM, auth.iamDelegate.AuthenticationType())
	assert.Equal(t, AUTHTYPE_IAM_ASSUME, auth.AuthenticationType())

	// Exercise the NewBuilder method and verify that it returns a builder that can
	// be used to construct an authenticator equivalent to "auth".
	builder := auth.NewBuilder()
	auth2, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth2)
	assert.Equal(t, auth, auth2)
}

// Tests that construct an authenticator via map properties.
func TestIamAssumeAuthenticatorFromMap(t *testing.T) {
	_, err := newIamAssumeAuthenticatorFromMap(nil)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s", err.Error())

	var props = map[string]string{
		PROPNAME_AUTH_URL: iamAssumeMockURL,
		PROPNAME_APIKEY:   iamAssumeMockApiKey,
	}
	_, err = newIamAssumeAuthenticatorFromMap(props)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s", err.Error())

	props = map[string]string{
		PROPNAME_APIKEY:         "",
		PROPNAME_IAM_PROFILE_ID: "",
	}
	_, err = newIamAssumeAuthenticatorFromMap(props)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s", err.Error())

	props = map[string]string{
		PROPNAME_APIKEY:          iamAssumeMockApiKey,
		PROPNAME_IAM_PROFILE_CRN: iamAssumeMockProfileCRN,
	}
	authenticator, err := newIamAssumeAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, iamAssumeMockApiKey, authenticator.iamDelegate.ApiKey)
	assert.Equal(t, iamAssumeMockProfileCRN, authenticator.iamProfileCRN)
	assert.Equal(t, AUTHTYPE_IAM_ASSUME, authenticator.AuthenticationType())

	props = map[string]string{
		PROPNAME_APIKEY:           iamAssumeMockApiKey,
		PROPNAME_IAM_PROFILE_NAME: iamAssumeMockProfileName,
		PROPNAME_IAM_ACCOUNT_ID:   iamAssumeMockAccountID,
		PROPNAME_AUTH_DISABLE_SSL: "true",
		PROPNAME_CLIENT_ID:        iamAssumeMockClientID,
		PROPNAME_CLIENT_SECRET:    iamAssumeMockClientSecret,
		PROPNAME_SCOPE:            iamAssumeMockScope,
	}
	authenticator, err = newIamAssumeAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, iamAssumeMockProfileName, authenticator.iamProfileName)
	assert.Equal(t, iamAssumeMockAccountID, authenticator.iamAccountID)
	assert.True(t, authenticator.disableSSLVerification)
	assert.Equal(t, iamAssumeMockApiKey, authenticator.iamDelegate.ApiKey)
	assert.True(t, authenticator.iamDelegate.DisableSSLVerification)
	assert.Equal(t, iamAssumeMockClientID, authenticator.iamDelegate.ClientId)
	assert.Equal(t, iamAssumeMockClientSecret, authenticator.iamDelegate.ClientSecret)
	assert.Equal(t, iamAssumeMockScope, authenticator.iamDelegate.Scope)
	assert.Equal(t, AUTHTYPE_IAM_ASSUME, authenticator.AuthenticationType())
}

func TestIamAssumeAuthDefaultURL(t *testing.T) {
	auth, err := NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileID(iamAssumeMockProfileID).
		Build()
	assert.Nil(t, err)

	assert.Equal(t, defaultIamTokenServerEndpoint, auth.getURL())
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.url)
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.iamDelegate.url())
	assert.Equal(t, defaultIamTokenServerEndpoint, auth.iamDelegate.URL)
}

// startMockIAMAssumeServer will start a mock server endpoint that supports the
// "apikey" and "assume" flavors of the IAM getToken operation.
func startMockIAMAssumeServer(t *testing.T) *httptest.Server {
	var numUserTokenRequests = 0
	var numProfileTokenRequests = 0

	// Create the mock server.
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		operationPath := req.URL.EscapedPath()
		assert.Equal(t, "/identity/token", operationPath)

		// Assume that we'll return a 200 OK status code.
		statusCode := http.StatusOK
		responseBody := ""

		// Validate some common parts of the input request.
		assert.Equal(t, APPLICATION_JSON, req.Header.Get("Accept"))
		assert.Equal(t, FORM_URL_ENCODED_HEADER, req.Header.Get("Content-Type"))
		userAgent := req.Header.Get(headerNameUserAgent)
		grantType := req.FormValue("grant_type")

		var responseToken string

		// Validate and reply to the request based on the grant type.
		if grantType == iamAuthGrantTypeApiKey {
			numUserTokenRequests++
			assert.True(t, strings.HasPrefix(userAgent,
				fmt.Sprintf("%s/%s", sdkName, "iam-authenticator")))
			if numUserTokenRequests == 1 {
				responseToken = iamAssumeMockUserToken1
			} else {
				responseToken = iamAssumeMockUserToken2
			}
			apikey := req.FormValue("apikey")
			if apikey != iamAssumeMockApiKey {
				statusCode = http.StatusBadRequest
				responseBody = "Bad Request: invalid apikey"
			} else {
				responseBody = fmt.Sprintf(`{"access_token": "%s", "refresh_token": "not_available", "token_type": "Bearer", "expires_in": 3600, "expiration": %d}`,
					responseToken, GetCurrentTime()+3600)
			}
		} else if grantType == iamGrantTypeAssume {
			numProfileTokenRequests++
			assert.True(t, strings.HasPrefix(userAgent,
				fmt.Sprintf("%s/%s", sdkName, "iam-assume-authenticator")))
			profileCRN := req.FormValue("profile_crn")
			profileID := req.FormValue("profile_id")
			profileName := req.FormValue("profile_name")
			accountID := req.FormValue("account")
			if numProfileTokenRequests == 1 {
				responseToken = iamAssumeMockProfileToken1
			} else {
				responseToken = iamAssumeMockProfileToken2
			}

			if !validateTrustedProfile(profileCRN, profileID, profileName, accountID) {
				statusCode = http.StatusBadRequest
				responseBody = "Bad Request: invalid trusted profile"
			} else {
				responseBody = fmt.Sprintf(`{"access_token": "%s", "token_type": "Bearer", "expires_in": 3600, "expiration": %d}`,
					responseToken, GetCurrentTime()+3600)
			}
		} else {
			// error - incorrect grant type.
			statusCode = http.StatusBadRequest
			responseBody = "Bad Request: invalid grant type"
		}

		res.WriteHeader(statusCode)
		fmt.Fprint(res, responseBody)
	}))
	return server
}

func validateTrustedProfile(profileCRN, profileID, profileName, accountID string) bool {
	numParams := 0
	if profileCRN != "" {
		numParams++
	}
	if profileID != "" {
		numParams++
	}
	if profileName != "" {
		numParams++
	}

	if numParams != 1 {
		return false
	}

	if (profileName == "") != (accountID == "") {
		return false
	}

	if profileCRN != "" && profileCRN != iamAssumeMockProfileCRN {
		return false
	}

	if profileID != "" && profileID != iamAssumeMockProfileID {
		return false
	}

	if profileName != "" && profileName != iamAssumeMockProfileName {
		return false
	}

	if accountID != "" && accountID != iamAssumeMockAccountID {
		return false
	}

	return true
}

func TestIamAssumeAuthGetTokenSuccess(t *testing.T) {
	GetLogger().SetLogLevel(iamAssumeTestLogLevel)

	server := startMockIAMAssumeServer(t)
	defer server.Close()

	auth, err := NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileID(iamAssumeMockProfileID).
		SetURL(server.URL).
		Build()
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
	assert.Equal(t, iamAssumeMockProfileToken1, accessToken)
	assert.Equal(t, iamAssumeMockProfileToken1, auth.getTokenData().AccessToken)

	// We should also get back a nil error from synchronizedRequestToken()
	// because calling it should NOT result in a new token request.
	assert.Nil(t, auth.synchronizedRequestToken())

	// Call GetToken() again and verify that we get the cached value.
	// Note: we'll Set Scope so that if the IAM operation is actually called again,
	// we'll receive the second access token. We don't want the IAM operation called again yet.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAssumeMockProfileToken1, accessToken)

	// Force expiration and verify that GetToken() fetched the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, iamAssumeMockProfileToken2, accessToken)
	assert.Equal(t, iamAssumeMockProfileToken2, auth.getTokenData().AccessToken)
}

func TestIamAssumeAuthGetTokenSuccess10SecWindow(t *testing.T) {
	GetLogger().SetLogLevel(iamAssumeTestLogLevel)

	server := startMockIAMAssumeServer(t)
	defer server.Close()

	auth, err := NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileCRN(iamAssumeMockProfileCRN).
		SetURL(server.URL).
		Build()
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
	assert.Equal(t, iamAssumeMockProfileToken1, accessToken)
	assert.Equal(t, iamAssumeMockProfileToken1, auth.getTokenData().AccessToken)

	// Call GetToken() again and verify that we get the cached value.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAssumeMockProfileToken1, accessToken)

	// Force expiration and verify that GetToken() fetched the second access token.
	// We'll set expiration to be current-time + <iamExpirationWindow> (10 secs),
	// to test the scenario where we should refresh the token when we are within 10 secs
	// of expiration.
	auth.getTokenData().Expiration = GetCurrentTime() + iamExpirationWindow
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.NotNil(t, auth.getTokenData())
	assert.Equal(t, iamAssumeMockProfileToken2, accessToken)
	assert.Equal(t, iamAssumeMockProfileToken2, auth.getTokenData().AccessToken)
}

func TestIamAssumeAuthRequestTokenError1(t *testing.T) {
	GetLogger().SetLogLevel(iamAssumeTestLogLevel)

	// Force an error while resolving the service URL.
	auth, err := NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileCRN(iamAssumeMockProfileCRN).
		SetURL("https://badhost").
		Build()
	assert.Nil(t, err)

	iamToken, err := auth.RequestToken()
	assert.NotNil(t, err)
	assert.Nil(t, iamToken)
	t.Logf("Expected error: %s\n", err.Error())
}

func TestIamAssumeAuthAuthenticateSuccess(t *testing.T) {
	GetLogger().SetLogLevel(iamAssumeTestLogLevel)

	server := startMockIAMAssumeServer(t)
	defer server.Close()

	auth, err := NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileName(iamAssumeMockProfileName).
		SetIAMAccountID(iamAssumeMockAccountID).
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
	assert.Equal(t, "Bearer "+iamAssumeMockProfileToken1, authHeader)

	// Call Authenticate again to make sure we used the cached access token.
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader = request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+iamAssumeMockProfileToken1, authHeader)

	// Force expiration (in both the Iam and IamAssume authenticators) and
	// verify that Authenticate() fetched the second access token.
	auth.getTokenData().Expiration = GetCurrentTime() - 1
	auth.iamDelegate.getTokenData().Expiration = GetCurrentTime() - 1
	err = auth.Authenticate(request)
	assert.Nil(t, err)
	authHeader = request.Header.Get("Authorization")
	assert.Equal(t, "Bearer "+iamAssumeMockProfileToken2, authHeader)
	assert.Equal(t, iamAssumeMockUserToken2, auth.iamDelegate.getTokenData().AccessToken)
}

func TestIamAssumeAuthAuthenticateFailBadApiKey(t *testing.T) {
	GetLogger().SetLogLevel(iamAssumeTestLogLevel)

	server := startMockIAMAssumeServer(t)
	defer server.Close()

	// Set up the authenticator with a bogus apikey
	// so that we can't successfully retrieve an access token.
	auth, err := NewIamAssumeAuthenticatorBuilder().
		SetApiKey("BAD_APIKEY").
		SetIAMProfileName(iamAssumeMockProfileName).
		SetIAMAccountID(iamAssumeMockAccountID).
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

func TestIamAssumeAuthAuthenticateFailBadProfileCRN(t *testing.T) {
	GetLogger().SetLogLevel(iamAssumeTestLogLevel)

	server := startMockIAMAssumeServer(t)
	defer server.Close()

	// Set up the authenticator with a bogus profile crn value
	// so that we can't successfully retrieve an access token.
	auth, err := NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileCRN("BAD_PROFILE_CRN").
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

func TestIamAssumeAuthBackgroundTokenRefreshSuccess(t *testing.T) {
	GetLogger().SetLogLevel(iamAssumeTestLogLevel)

	server := startMockIAMAssumeServer(t)
	defer server.Close()

	auth, err := NewIamAssumeAuthenticatorBuilder().
		SetApiKey(iamAssumeMockApiKey).
		SetIAMProfileCRN(iamAssumeMockProfileCRN).
		SetURL(server.URL).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	// Force the first fetch and verify we got the first access token.
	accessToken, err := auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAssumeMockProfileToken1, accessToken)

	// Now simulate being in the refresh window where the token is not expired but still needs to be refreshed.
	auth.getTokenData().RefreshTime = GetCurrentTime() - 1

	// Authenticator should detect the need to get a new access token in the background but use the current
	// cached access token for this next GetToken() call.
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAssumeMockProfileToken1, accessToken)

	// Wait for the background thread to finish.
	time.Sleep(2 * time.Second)
	accessToken, err = auth.GetToken()
	assert.Nil(t, err)
	assert.Equal(t, iamAssumeMockProfileToken2, accessToken)
}

// In order to test with a live IAM server, create file "iamassume.env" in the project root.
// It should look like this:
//
// IAMASSUME1_AUTH_TYPE=iamAssume
// IAMASSUME1_APIKEY=<apikey>
// IAMASSUME1_IAM_PROFILE_ID=<profile id>
//
// Then comment out the "t.Skip()" line below, then run these commands:
//
//	cd core
//	go test -v -tags=auth -run=TestIamAssumeLiveTokenServer
//
// To trace request/response messages, change "iamAssumeTestLogLevel" above to be "LevelDebug".
func TestIamAssumeLiveTokenServer(t *testing.T) {
	t.Skip("Skipping IamAssumeAuthenticator integration test...")

	GetLogger().SetLogLevel(iamAssumeTestLogLevel)

	var request *http.Request
	var err error
	var authHeader string

	// Get an iam authenticator from the environment.
	t.Setenv("IBM_CREDENTIALS_FILE", "../iamassume.env")
	auth, err := GetAuthenticatorFromEnvironment("iamassume1")
	assert.Nil(t, err)
	assert.NotNil(t, auth)

	_, ok := auth.(*IamAssumeAuthenticator)
	assert.Equal(t, true, ok)

	// Create a new Request object.
	builder, err := NewRequestBuilder("GET").ResolveRequestURL("https://localhost/placeholder/url", "", nil)
	assert.Nil(t, err)
	assert.NotNil(t, builder)

	request, _ = builder.Build()
	assert.NotNil(t, request)
	err = auth.Authenticate(request)
	if err != nil {
		authError := err.(*AuthenticationError)
		iamError := authError.Err
		iamResponse := authError.Response
		t.Logf("Unexpected authentication error: %s\n", iamError.Error())
		t.Logf("Authentication response: %v+\n", iamResponse)

	}
	assert.Nil(t, err)

	authHeader = request.Header.Get("Authorization")
	assert.NotEmpty(t, authHeader)
	assert.True(t, strings.HasPrefix(authHeader, "Bearer "))
	t.Logf("Authorization: %s\n", authHeader)
}
