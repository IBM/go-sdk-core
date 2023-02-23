//go:build all || fast || basesvc
// +build all fast basesvc

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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const vcapServicesKey = "VCAP_SERVICES"

// Sets a test VCAP_SERVICES value in the environment for testing.
func setTestVCAP(t *testing.T) {
	data, err := os.ReadFile("../resources/vcap_services.json")
	if assert.Nil(t, err) {
		os.Setenv(vcapServicesKey, string(data))
	}
}

func clearTestVCAP() {
	os.Unsetenv(vcapServicesKey)
}

func TestGetServicePropertiesError(t *testing.T) {
	_, err := getServiceProperties("")
	assert.NotNil(t, err)
}

func TestGetServicePropertiesNoConfig(t *testing.T) {
	props, err := getServiceProperties("not_a_service")
	assert.Nil(t, err)
	assert.Nil(t, props)
}

func TestGetServicePropertiesFromCredentialFile(t *testing.T) {
	pwd, _ := os.Getwd()
	credentialFilePath := path.Join(pwd, "/../resources/my-credentials.env")
	os.Setenv("IBM_CREDENTIALS_FILE", credentialFilePath)

	props, err := GetServiceProperties("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service1/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "true", props[PROPNAME_SVC_DISABLE_SSL])
	assert.Equal(t, "true", props[PROPNAME_SVC_ENABLE_GZIP])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_IAM), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))
	assert.Equal(t, "my-api-key", props[PROPNAME_APIKEY])
	assert.Equal(t, "my-client-id", props[PROPNAME_CLIENT_ID])
	assert.Equal(t, "my-client-secret", props[PROPNAME_CLIENT_SECRET])
	assert.Equal(t, "https://iamhost/iam/api", props[PROPNAME_AUTH_URL])
	assert.Equal(t, "true", props[PROPNAME_AUTH_DISABLE_SSL])

	props, err = getServiceProperties("service2")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service2/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "false", props[PROPNAME_SVC_DISABLE_SSL])
	assert.Equal(t, "false", props[PROPNAME_SVC_ENABLE_GZIP])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_BASIC), strings.ToUpper(props["AUTHTYPE"]))
	assert.Equal(t, "my-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-password", props[PROPNAME_PASSWORD])

	props, err = getServiceProperties("service3")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service3/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "false", props[PROPNAME_SVC_DISABLE_SSL])
	assert.Equal(t, "notabool", props[PROPNAME_SVC_ENABLE_GZIP])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_CP4D), strings.ToUpper(props["AUTHTYPE"]))
	assert.Equal(t, "my-cp4d-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-cp4d-password", props[PROPNAME_PASSWORD])
	assert.Equal(t, "https://cp4dhost/cp4d/api", props[PROPNAME_AUTH_URL])
	assert.Equal(t, "false", props[PROPNAME_AUTH_DISABLE_SSL])

	props, err = GetServiceProperties("equal_service")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "=https:/my=host.com/my=service/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "=my=api=key=", props[PROPNAME_APIKEY])

	props, err = getServiceProperties("not_a_service")
	assert.Nil(t, err)
	assert.Nil(t, props)

	props, err = getServiceProperties("service6")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "IAM", props[PROPNAME_AUTH_TYPE])
	assert.Equal(t, "my-api-key", props[PROPNAME_APIKEY])
	assert.Equal(t, "https://iamhost/iam/api", props[PROPNAME_AUTH_URL])
	assert.Equal(t, "scope1 scope2 scope3", props[PROPNAME_SCOPE])

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestGetServicePropertiesFromEnvironment(t *testing.T) {
	setTestEnvironment()

	props, err := GetServiceProperties("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service1/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "true", props[PROPNAME_SVC_DISABLE_SSL])
	assert.Equal(t, "true", props[PROPNAME_SVC_ENABLE_GZIP])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_IAM), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))
	assert.Equal(t, "my-api-key", props[PROPNAME_APIKEY])
	assert.Equal(t, "my-client-id", props[PROPNAME_CLIENT_ID])
	assert.Equal(t, "my-client-secret", props[PROPNAME_CLIENT_SECRET])
	assert.Equal(t, "https://iamhost/iam/api", props[PROPNAME_AUTH_URL])
	assert.Equal(t, "true", props[PROPNAME_AUTH_DISABLE_SSL])

	props, err = getServiceProperties("service2")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service2/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "false", props[PROPNAME_SVC_DISABLE_SSL])
	assert.Equal(t, "false", props[PROPNAME_SVC_ENABLE_GZIP])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_BASIC), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))
	assert.Equal(t, "my-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-password", props[PROPNAME_PASSWORD])

	props, err = getServiceProperties("service3")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service3/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "false", props[PROPNAME_SVC_DISABLE_SSL])
	assert.Equal(t, "notabool", props[PROPNAME_SVC_ENABLE_GZIP])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_CP4D), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))
	assert.Equal(t, "my-cp4d-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-cp4d-password", props[PROPNAME_PASSWORD])
	assert.Equal(t, "https://cp4dhost/cp4d/api", props[PROPNAME_AUTH_URL])
	assert.Equal(t, "false", props[PROPNAME_AUTH_DISABLE_SSL])

	props, err = GetServiceProperties("equal_service")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://my=host.com/my=service/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "===my=iam=apikey===", props[PROPNAME_APIKEY])

	props, err = getServiceProperties("not_a_service")
	assert.Nil(t, err)
	assert.Nil(t, props)

	props, err = getServiceProperties("service6")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "iam", props[PROPNAME_AUTH_TYPE])
	assert.Equal(t, "my-api-key", props[PROPNAME_APIKEY])
	assert.Equal(t, "A B C D", props[PROPNAME_SCOPE])

	clearTestEnvironment()
	assert.Equal(t, "", os.Getenv("SERVICE_1_URL"))
}

func TestGetServicePropertiesFromVCAP(t *testing.T) {
	setTestVCAP(t)

	props, err := getServiceProperties("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service1/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "my-vcap-apikey1", props[PROPNAME_APIKEY])
	assert.Equal(t, "my-vcap-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-vcap-password", props[PROPNAME_PASSWORD])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_IAM), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))

	props, err = getServiceProperties("service2")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service2/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "", props[PROPNAME_APIKEY])
	assert.Equal(t, "my-vcap-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-vcap-password", props[PROPNAME_PASSWORD])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_BASIC), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))

	props, err = getServiceProperties("service3")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service3/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "my-vcap-apikey3", props[PROPNAME_APIKEY])
	assert.Equal(t, "", props[PROPNAME_USERNAME])
	assert.Equal(t, "", props[PROPNAME_PASSWORD])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_IAM), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))

	clearTestVCAP()
}

func TestLoadFromVCAPServicesWithServiceEntries(t *testing.T) {
	setTestVCAP(t)
	// Verify we checked service entry names first
	credential1 := loadFromVCAPServices("service_entry_key_and_key_to_service_entries")
	isNotNil := assert.NotNil(t, credential1, "Credentials1 should not be nil")
	if !isNotNil {
		return
	}
	assert.Equal(t, "not-a-username", credential1.Username)
	assert.Equal(t, "not-a-password", credential1.Password)
	assert.Equal(t, "https://on.the.toolchainplatform.net/devops-insights/api", credential1.URL)
	// Verify we checked keys that map to lists of service entries
	credential2 := loadFromVCAPServices("key_to_service_entry_1")
	isNotNil = assert.NotNil(t, credential2, "Credentials2 should not be nil")
	if !isNotNil {
		return
	}
	assert.Equal(t, "my-vcap-apikey3", credential2.APIKey)
	assert.Equal(t, "https://service3/api", credential2.URL)
	credential3 := loadFromVCAPServices("key_to_service_entry_2")
	isNotNil = assert.NotNil(t, credential3, "Credentials3 should not be nil")
	if !isNotNil {
		return
	}
	assert.Equal(t, "not-a-username-3", credential3.Username)
	assert.Equal(t, "not-a-password-3", credential3.Password)
	assert.Equal(t, "https://on.the.toolchainplatform.net/devops-insights-3/api", credential3.URL)
	clearTestVCAP()
}

func TestLoadFromVCAPServicesEmptyService(t *testing.T) {
	setTestVCAP(t)
	// Verify we checked service entry names first
	credential := loadFromVCAPServices("empty_service")
	assert.Nil(t, credential, "Credentials should not be nil")
	clearTestVCAP()
}

func TestLoadFromVCAPServicesNoCredentials(t *testing.T) {
	setTestVCAP(t)
	// Verify we checked service entry names first
	credential := loadFromVCAPServices("no-creds-service")
	assert.Nil(t, credential)
	clearTestVCAP()
}

func TestLoadFromVCAPServicesWithEmptyString(t *testing.T) {
	clearTestVCAP()
	credential := loadFromVCAPServices("watson")
	assert.Nil(t, credential, "Credentials should nil")
}

func TestLoadFromVCAPServicesWithInvalidJSON(t *testing.T) {
	vcapServicesFail := `{
		"watson": [
			"credentials": {
				"url": "https://api.us-south.compare-comply.watson.cloud.ibm.com",
				"username": "bogus username",
				"password": "bogus password",
				"apikey": "bogus apikey"
			}
		}]
	}`
	os.Setenv("VCAP_SERVICES", vcapServicesFail)
	credential := loadFromVCAPServices("watson")
	assert.Nil(t, credential, "Credentials should be nil")
	os.Unsetenv("VCAP_SERVICES")
}
