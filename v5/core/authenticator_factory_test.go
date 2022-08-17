//go:build all || fast
// +build all fast

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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: the following functions are used from the config_utils_test.go file:
// setTestEnvironment()
// clearTestEnvironment()
// setTestVCAP()
// clearTestVCAP()

func TestGetAuthenticatorFromEnvironment1(t *testing.T) {
	os.Setenv("IBM_CREDENTIALS_FILE", "../resources/my-credentials.env")

	authenticator, err := GetAuthenticatorFromEnvironment("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())

	authenticator, err = GetAuthenticatorFromEnvironment("service2")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, authenticator.AuthenticationType())

	authenticator, err = GetAuthenticatorFromEnvironment("service3")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_CP4D, authenticator.AuthenticationType())

	authenticator, err = GetAuthenticatorFromEnvironment("service6")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())
	iamAuthenticator, ok := authenticator.(*IamAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, iamAuthenticator)
	assert.Equal(t, "scope1 scope2 scope3", iamAuthenticator.Scope)

	authenticator, err = GetAuthenticatorFromEnvironment("service7")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_CONTAINER, authenticator.AuthenticationType())
	containerAuth, ok := authenticator.(*ContainerAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, containerAuth)
	assert.Equal(t, "crtoken.txt", containerAuth.CRTokenFilename)
	assert.Equal(t, "iam-user1", containerAuth.IAMProfileName)
	assert.Equal(t, "iam-id1", containerAuth.IAMProfileID)
	assert.Equal(t, "https://iamhost/iam/api", containerAuth.URL)
	assert.Equal(t, "iam-client1", containerAuth.ClientID)
	assert.Equal(t, "iam-secret1", containerAuth.ClientSecret)
	assert.True(t, containerAuth.DisableSSLVerification)
	assert.Equal(t, "scope1", containerAuth.Scope)

	// VPC Authenticator with default config.
	authenticator, err = GetAuthenticatorFromEnvironment("service8a")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_VPC, authenticator.AuthenticationType())
	vpcAuth, ok := authenticator.(*VpcInstanceAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, vpcAuth)
	assert.Empty(t, vpcAuth.IAMProfileCRN)
	assert.Empty(t, vpcAuth.IAMProfileID)
	assert.Empty(t, vpcAuth.URL)

	// VPC Authenticator with profile crn and url configured.
	authenticator, err = GetAuthenticatorFromEnvironment("service8b")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_VPC, authenticator.AuthenticationType())
	vpcAuth, ok = authenticator.(*VpcInstanceAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, vpcAuth)
	assert.Equal(t, "crn:iam-profile1", vpcAuth.IAMProfileCRN)
	assert.Empty(t, vpcAuth.IAMProfileID)
	assert.Equal(t, "http://vpc.imds.com/api", vpcAuth.URL)

	// VPC Authenticator with profile id configured.
	authenticator, err = GetAuthenticatorFromEnvironment("service8c")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_VPC, authenticator.AuthenticationType())
	vpcAuth, ok = authenticator.(*VpcInstanceAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, vpcAuth)
	assert.Empty(t, vpcAuth.IAMProfileCRN)
	assert.Equal(t, "iam-profile1-id", vpcAuth.IAMProfileID)
	assert.Empty(t, vpcAuth.URL)

	// IAM Authenticator using refresh token.
	authenticator, err = GetAuthenticatorFromEnvironment("service9")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())
	iamAuth, ok := authenticator.(*IamAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, iamAuth)
	assert.Empty(t, iamAuth.ApiKey)
	assert.Equal(t, "refresh-token", iamAuth.RefreshToken)
	assert.Equal(t, "user1", iamAuth.ClientId)
	assert.Equal(t, "secret1", iamAuth.ClientSecret)
	assert.Equal(t, "https://iam.refresh-token.com", iamAuth.URL)

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestGetAuthenticatorFromEnvironment2(t *testing.T) {
	setTestEnvironment()

	authenticator, err := GetAuthenticatorFromEnvironment("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())

	authenticator, err = GetAuthenticatorFromEnvironment("service2")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, authenticator.AuthenticationType())

	authenticator, err = GetAuthenticatorFromEnvironment("service3")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_CP4D, authenticator.AuthenticationType())

	authenticator, err = GetAuthenticatorFromEnvironment("not_a_service")
	assert.Nil(t, err)
	assert.Nil(t, authenticator)

	authenticator, err = GetAuthenticatorFromEnvironment("service7")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_CONTAINER, authenticator.AuthenticationType())
	containerAuth, ok := authenticator.(*ContainerAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, containerAuth)
	assert.Equal(t, "crtoken.txt", containerAuth.CRTokenFilename)
	assert.Equal(t, "iam-user2", containerAuth.IAMProfileName)
	assert.Equal(t, "iam-id2", containerAuth.IAMProfileID)
	assert.Equal(t, "https://iamhost/iam/api", containerAuth.URL)
	assert.Equal(t, "iam-client2", containerAuth.ClientID)
	assert.Equal(t, "iam-secret2", containerAuth.ClientSecret)
	assert.False(t, containerAuth.DisableSSLVerification)
	assert.Equal(t, "scope2 scope3", containerAuth.Scope)

	authenticator, err = GetAuthenticatorFromEnvironment("service8")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_VPC, authenticator.AuthenticationType())
	vpcAuth, ok := authenticator.(*VpcInstanceAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, vpcAuth)
	assert.Equal(t, "crn:iam-profile1", vpcAuth.IAMProfileCRN)
	assert.Equal(t, "http://vpc.imds.com/api", vpcAuth.URL)

	authenticator, err = GetAuthenticatorFromEnvironment("service9")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_BEARER_TOKEN, authenticator.AuthenticationType())
	btAuth, ok := authenticator.(*BearerTokenAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, btAuth)
	assert.Equal(t, "my-token", btAuth.BearerToken)

	authenticator, err = GetAuthenticatorFromEnvironment("service10")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_NOAUTH, authenticator.AuthenticationType())
	noAuth, ok := authenticator.(*NoAuthAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, noAuth)

	authenticator, err = GetAuthenticatorFromEnvironment("service11")
	assert.NotNil(t, err)
	assert.Nil(t, authenticator)

	authenticator, err = GetAuthenticatorFromEnvironment("service12")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())
	iamAuth, ok := authenticator.(*IamAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, iamAuth)
	assert.Equal(t, "my-apikey", iamAuth.ApiKey)

	authenticator, err = GetAuthenticatorFromEnvironment("service13")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_CONTAINER, authenticator.AuthenticationType())
	containerAuth, ok = authenticator.(*ContainerAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, containerAuth)
	assert.Equal(t, "iam-user2", containerAuth.IAMProfileName)

	clearTestEnvironment()
}

func TestGetAuthenticatorFromEnvironment3(t *testing.T) {
	setTestVCAP(t)

	authenticator, err := GetAuthenticatorFromEnvironment("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())

	authenticator, err = GetAuthenticatorFromEnvironment("service2")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_BASIC, authenticator.AuthenticationType())

	authenticator, err = GetAuthenticatorFromEnvironment("service3")
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, AUTHTYPE_IAM, authenticator.AuthenticationType())

	clearTestVCAP()
}
