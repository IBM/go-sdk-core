//go:build all || fast || auth
// +build all fast auth

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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBearerToken(t *testing.T) {
	authenticator := &BearerTokenAuthenticator{
		BearerToken: "",
	}
	err := authenticator.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(ERRORMSG_PROP_MISSING, "BearerToken").Error(), err.Error())

	authenticator, err = NewBearerTokenAuthenticator("my-bearer-token")
	assert.NotNil(t, authenticator)
	assert.Nil(t, err)
}

func TestBearerTokenAuthenticate(t *testing.T) {
	authenticator := &BearerTokenAuthenticator{
		BearerToken: "my-bearer-token",
	}
	assert.Equal(t, authenticator.AuthenticationType(), AUTHTYPE_BEARER_TOKEN)

	// Create a new Request object.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://localhost/placeholder/url", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Test the "Authenticate" method to make sure the correct header is added to the Request.
	_ = authenticator.Authenticate(request)
	assert.Equal(t, request.Header.Get("Authorization"), "Bearer my-bearer-token")
}

func TestNewBearerTokenAuthenticatorFromMap(t *testing.T) {
	_, err := newBearerTokenAuthenticatorFromMap(nil)
	assert.NotNil(t, err)

	var props = map[string]string{
		PROPNAME_BEARER_TOKEN: "",
	}
	_, err = newBearerTokenAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_BEARER_TOKEN: "my-token",
	}
	authenticator, err := newBearerTokenAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, "my-token", authenticator.BearerToken)
}
