package core

/**
 * Copyright 2019 IBM All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

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
	credential1 := LoadFromVCAPServices("watson")
	assert.Equal(t, "bogus apikey", credential1.APIKey)
	os.Unsetenv("VCAP_SERVICES")

	credential2 := LoadFromVCAPServices("watson")
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
	credential3 := LoadFromVCAPServices("watson")
	assert.Nil(t, credential3)
	os.Unsetenv("VCAP_SERVICES")
}
