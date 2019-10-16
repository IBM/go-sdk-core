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

// Map containing environment variables used in testing.
var testEnvironment = map[string]string{
	"SERVICE_1_URL":              "https://service1/api",
	"SERVICE_1_DISABLE_SSL":      "true",
	"SERVICE_1_AUTH_TYPE":        "IaM",
	"SERVICE_1_APIKEY":           "my-api-key",
	"SERVICE_1_CLIENT_ID":        "my-client-id",
	"SERVICE_1_CLIENT_SECRET":    "my-client-secret",
	"SERVICE_1_AUTH_URL":         "https://iamhost/iam/api",
	"SERVICE_1_AUTH_DISABLE_SSL": "true",
	"SERVICE2_URL":               "https://service2/api",
	"SERVICE2_DISABLE_SSL":       "false",
	"SERVICE2_AUTH_TYPE":         "bAsIC",
	"SERVICE2_USERNAME":          "my-user",
	"SERVICE2_PASSWORD":          "my-password",
	"SERVICE3_URL":               "https://service3/api",
	"SERVICE3_DISABLE_SSL":       "false",
	"SERVICE3_AUTH_TYPE":         "Cp4D",
	"SERVICE3_AUTH_URL":          "https://cp4dhost/cp4d/api",
	"SERVICE3_USERNAME":          "my-cp4d-user",
	"SERVICE3_PASSWORD":          "my-cp4d-password",
	"SERVICE3_AUTH_DISABLE_SSL":  "false",
}

// Set the environment variables described in our map.
func setTestEnvironment() {
	for key, value := range testEnvironment {
		os.Setenv(key, value)
	}
}

// Clear the test-related environment variables.
func clearTestEnvironment() {
	for key := range testEnvironment {
		os.Unsetenv(key)
	}
}

// Sets a test VCAP_SERVICES value in the environment for testing.
func setTestVCAP() {
	vcapServices := `
	{
		"service-1":[{
			"credentials":{
				"url":"https://service1/api",
				"username":"my-vcap-user",
				"password":"my-vcap-password",
				"apikey":"my-vcap-apikey1"
	        }
		}],
		"service2":[{
			"credentials":{
				"url":"https://service2/api",
				"username":"my-vcap-user",
				"password":"my-vcap-password"
			}
		}],
		"service3":[{
			"credentials":{
				"url":"https://service3/api",
				"apikey":"my-vcap-apikey3"
			}
		}]
	}`

	os.Setenv("VCAP_SERVICES", vcapServices)
}

func clearTestVCAP() {
	os.Unsetenv("VCAP_SERVICES")
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

	props, err := getServiceProperties("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service1/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "true", props[PROPNAME_SVC_DISABLE_SSL])
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
	assert.Equal(t, strings.ToUpper(AUTHTYPE_BASIC), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))
	assert.Equal(t, "my-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-password", props[PROPNAME_PASSWORD])

	props, err = getServiceProperties("service3")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service3/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "false", props[PROPNAME_SVC_DISABLE_SSL])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_CP4D), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))
	assert.Equal(t, "my-cp4d-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-cp4d-password", props[PROPNAME_PASSWORD])
	assert.Equal(t, "https://cp4dhost/cp4d/api", props[PROPNAME_AUTH_URL])
	assert.Equal(t, "false", props[PROPNAME_AUTH_DISABLE_SSL])

	props, err = getServiceProperties("not_a_service")
	assert.Nil(t, err)
	assert.Nil(t, props)

	os.Unsetenv("IBM_CREDENTIALS_FILE")
}

func TestGetServicePropertiesFromEnvironment(t *testing.T) {
	setTestEnvironment()

	props, err := getServiceProperties("service-1")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service1/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "true", props[PROPNAME_SVC_DISABLE_SSL])
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
	assert.Equal(t, strings.ToUpper(AUTHTYPE_BASIC), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))
	assert.Equal(t, "my-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-password", props[PROPNAME_PASSWORD])

	props, err = getServiceProperties("service3")
	assert.Nil(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, "https://service3/api", props[PROPNAME_SVC_URL])
	assert.Equal(t, "false", props[PROPNAME_SVC_DISABLE_SSL])
	assert.Equal(t, strings.ToUpper(AUTHTYPE_CP4D), strings.ToUpper(props[PROPNAME_AUTH_TYPE]))
	assert.Equal(t, "my-cp4d-user", props[PROPNAME_USERNAME])
	assert.Equal(t, "my-cp4d-password", props[PROPNAME_PASSWORD])
	assert.Equal(t, "https://cp4dhost/cp4d/api", props[PROPNAME_AUTH_URL])
	assert.Equal(t, "false", props[PROPNAME_AUTH_DISABLE_SSL])

	props, err = getServiceProperties("not_a_service")
	assert.Nil(t, err)
	assert.Nil(t, props)

	clearTestEnvironment()
	assert.Equal(t, "", os.Getenv("SERVICE_1_URL"))
}

func TestGetServicePropertiesFromVCAP(t *testing.T) {
	setTestVCAP()

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

func TestLoadFromVCAPServices(t *testing.T) {
	vcapServices := `{
		"watson": [{
			"credentials": {
				"url": "https://gateway.watsonplatform.net/compare-comply/api",
				"username": "bogus username",
				"password": "bogus password",
				"apikey": "bogus apikey"
			}
		}]
	}`
	os.Setenv("VCAP_SERVICES", vcapServices)
	credential1 := loadFromVCAPServices("watson")
	assert.Equal(t, "bogus apikey", credential1.APIKey)
	os.Unsetenv("VCAP_SERVICES")

	credential2 := loadFromVCAPServices("watson")
	assert.Nil(t, credential2)

	vcapServicesFail := `{
		"watson": [
			"credentials": {
				"url": "https://gateway.watsonplatform.net/compare-comply/api",
				"username": "bogus username",
				"password": "bogus password",
				"apikey": "bogus apikey"
			}
		}]
	}`
	os.Setenv("VCAP_SERVICES", vcapServicesFail)
	credential3 := loadFromVCAPServices("watson")
	assert.Nil(t, credential3)
	os.Unsetenv("VCAP_SERVICES")
}
