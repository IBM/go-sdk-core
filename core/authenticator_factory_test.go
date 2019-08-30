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
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: the following functions are used from the config_utils_test.go file:
// setTestEnvironment()
// clearTestEnvironment()
// setTestVCAP()
// clearTestVCAP()

func TestGetAuthenticatorFromEnvironment1(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

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

	clearTestEnvironment()
}

func TestGetAuthenticatorFromEnvironment3(t *testing.T) {
	setTestVCAP()

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
