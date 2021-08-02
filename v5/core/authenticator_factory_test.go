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
	assert.Equal(t, AUTHTYPE_CRAUTH, authenticator.AuthenticationType())
	crAuthenticator, ok := authenticator.(*ComputeResourceAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, crAuthenticator)
	assert.Equal(t, "crtoken.txt", crAuthenticator.CRTokenFilename)
	assert.Equal(t, "iam-user1", crAuthenticator.IAMProfileName)
	assert.Equal(t, "iam-id1", crAuthenticator.IAMProfileID)
	assert.Equal(t, "https://iamhost/iam/api", crAuthenticator.URL)
	assert.Equal(t, "iam-client1", crAuthenticator.ClientID)
	assert.Equal(t, "iam-secret1", crAuthenticator.ClientSecret)
	assert.True(t, crAuthenticator.DisableSSLVerification)
	assert.Equal(t, "scope1", crAuthenticator.Scope)

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
	assert.Equal(t, AUTHTYPE_CRAUTH, authenticator.AuthenticationType())
	crAuthenticator, ok := authenticator.(*ComputeResourceAuthenticator)
	assert.True(t, ok)
	assert.NotNil(t, crAuthenticator)
	assert.Equal(t, "crtoken.txt", crAuthenticator.CRTokenFilename)
	assert.Equal(t, "iam-user2", crAuthenticator.IAMProfileName)
	assert.Equal(t, "iam-id2", crAuthenticator.IAMProfileID)
	assert.Equal(t, "https://iamhost/iam/api", crAuthenticator.URL)
	assert.Equal(t, "iam-client2", crAuthenticator.ClientID)
	assert.Equal(t, "iam-secret2", crAuthenticator.ClientSecret)
	assert.False(t, crAuthenticator.DisableSSLVerification)
	assert.Equal(t, "scope2 scope3", crAuthenticator.Scope)
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
