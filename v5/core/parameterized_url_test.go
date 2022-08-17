//go:build all || fast
// +build all fast

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
	"testing"

	"github.com/stretchr/testify/assert"
)

const parameterizedUrl = "{scheme}://{domain}:{port}"

var defaultUrlVariables = map[string]string{
	"scheme": "http",
	"domain": "ibm.com",
	"port":   "9300",
}

func TestConstructServiceURLWithNil(t *testing.T) {
	url, err := ConstructServiceURL(parameterizedUrl, defaultUrlVariables, nil)

	assert.Equal(t, url, "http://ibm.com:9300")
	assert.Nil(t, err)
}

func TestConstructServiceURLWithSomeProvidedVariables(t *testing.T) {
	providedUrlVariables := map[string]string{
		"scheme": "https",
		"port":   "22",
	}

	url, err := ConstructServiceURL(parameterizedUrl, defaultUrlVariables, providedUrlVariables)

	assert.Equal(t, url, "https://ibm.com:22")
	assert.Nil(t, err)
}

func TestConstructServiceURLWithAllProvidedVariables(t *testing.T) {
	var providedUrlVariables = map[string]string{
		"scheme": "https",
		"domain": "google.com",
		"port":   "22",
	}

	url, err := ConstructServiceURL(parameterizedUrl, defaultUrlVariables, providedUrlVariables)

	assert.Equal(t, url, "https://google.com:22")
	assert.Nil(t, err)
}

func TestConstructServiceURLWithInvalidVariable(t *testing.T) {
	var providedUrlVariables = map[string]string{
		"server": "value",
	}

	url, err := ConstructServiceURL(parameterizedUrl, defaultUrlVariables, providedUrlVariables)

	assert.Equal(t, url, "")
	assert.EqualError(
		t,
		err,
		"'server' is an invalid variable name.\nValid variable names: [domain port scheme].",
	)
}
