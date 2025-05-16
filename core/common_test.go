//go:build all || fast || basesvc || retries || auth

package core

import (
	"bytes"
	"encoding/json"
	"testing"
)

// (C) Copyright IBM Corp. 2020, 2024.
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

//
// This file contains definitions of various types that are shared among multiple testcase files.
//

type Foo struct {
	Name *string `json:"name,omitempty"`
}

func toJSON(obj interface{}) string {
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(obj)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// Map containing environment variables used in testing.
var testEnvironment = map[string]string{
	"SERVICE_1_URL":                     "https://service1/api",
	"SERVICE_1_DISABLE_SSL":             "true",
	"SERVICE_1_ENABLE_GZIP":             "true",
	"SERVICE_1_AUTH_TYPE":               "IaM",
	"SERVICE_1_APIKEY":                  "my-api-key",
	"SERVICE_1_CLIENT_ID":               "my-client-id",
	"SERVICE_1_CLIENT_SECRET":           "my-client-secret",
	"SERVICE_1_AUTH_URL":                "https://iamhost/iam/api",
	"SERVICE_1_AUTH_DISABLE_SSL":        "true",
	"SERVICE2_URL":                      "https://service2/api",
	"SERVICE2_DISABLE_SSL":              "false",
	"SERVICE2_ENABLE_GZIP":              "false",
	"SERVICE2_AUTH_TYPE":                "bAsIC",
	"SERVICE2_USERNAME":                 "my-user",
	"SERVICE2_PASSWORD":                 "my-password",
	"SERVICE3_URL":                      "https://service3/api",
	"SERVICE3_DISABLE_SSL":              "false",
	"SERVICE3_ENABLE_GZIP":              "notabool",
	"SERVICE3_AUTH_TYPE":                "Cp4D",
	"SERVICE3_AUTH_URL":                 "https://cp4dhost/cp4d/api",
	"SERVICE3_USERNAME":                 "my-cp4d-user",
	"SERVICE3_PASSWORD":                 "my-cp4d-password",
	"SERVICE3_AUTH_DISABLE_SSL":         "false",
	"EQUAL_SERVICE_URL":                 "https://my=host.com/my=service/api",
	"EQUAL_SERVICE_APIKEY":              "===my=iam=apikey===",
	"SERVICE6_AUTH_TYPE":                "iam",
	"SERVICE6_APIKEY":                   "my-api-key",
	"SERVICE6_SCOPE":                    "A B C D",
	"SERVICE7_AUTH_TYPE":                "container",
	"SERVICE7_CR_TOKEN_FILENAME":        "crtoken.txt",
	"SERVICE7_IAM_PROFILE_NAME":         "iam-user2",
	"SERVICE7_IAM_PROFILE_ID":           "iam-id2",
	"SERVICE7_AUTH_URL":                 "https://iamhost/iam/api",
	"SERVICE7_CLIENT_ID":                "iam-client2",
	"SERVICE7_CLIENT_SECRET":            "iam-secret2",
	"SERVICE7_SCOPE":                    "scope2 scope3",
	"SERVICE8_AUTH_TYPE":                "VPC",
	"SERVICE8_IAM_PROFILE_CRN":          "crn:iam-profile1",
	"SERVICE8_AUTH_URL":                 "http://vpc.imds.com/api",
	"SERVICE9_AUTH_TYPE":                "bearerToken",
	"SERVICE9_BEARER_TOKEN":             "my-token",
	"SERVICE10_AUTH_TYPE":               "noauth",
	"SERVICE11_AUTH_TYPE":               "bad_auth_type",
	"SERVICE12_APIKEY":                  "my-apikey",
	"SERVICE13_IAM_PROFILE_NAME":        "iam-user2",
	"SERVICE14_AUTH_TYPE":               "mcsp",
	"SERVICE14_AUTH_URL":                "https://mcsp.ibm.com",
	"SERVICE14_APIKEY":                  "my-api-key",
	"SERVICE14_AUTH_DISABLE_SSL":        "true",
	"SERVICE15_AUTH_URL":                "https://iam.assume.ibm.com",
	"SERVICE15_AUTH_TYPE":               "IAMAssUME",
	"SERVICE15_APIKEY":                  "my-apikey",
	"SERVICE15_IAM_PROFILE_NAME":        "profile-1",
	"SERVICE15_IAM_ACCOUNT_ID":          "account-1",
	"SERVICE16_AUTH_TYPE":               "mcspv2",
	"SERVICE16_APIKEY":                  "my-api-key",
	"SERVICE16_AUTH_URL":                "https://mcspv2.ibm.com",
	"SERVICE16_SCOPE_COLLECTION_TYPE":   "accounts",
	"SERVICE16_SCOPE_ID":                "global_accounts",
	"SERVICE16_INCLUDE_BUILTIN_ACTIONS": "true",
	"SERVICE16_INCLUDE_CUSTOM_ACTIONS":  "true",
	"SERVICE16_INCLUDE_ROLES":           "false",
	"SERVICE16_PREFIX_ROLES":            "true",
	"SERVICE16_CALLER_EXT_CLAIM":        `{"productID":"prod456"}`,
	"SERVICE16_AUTH_DISABLE_SSL":        "true",
}

// setTestEnvironment sets the environment variables described in our map.
// The environment variables are restored to its original value after the test.
func setTestEnvironment(t *testing.T) {
	for key, value := range testEnvironment {
		t.Setenv(key, value)
	}
}
