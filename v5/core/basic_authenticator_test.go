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

func TestBasicAuthUsername(t *testing.T) {
	authenticator := &BasicAuthenticator{
		Username: "{username}",
		Password: "password",
	}
	err := authenticator.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(ERRORMSG_PROP_INVALID, "Username").Error(), err.Error())

	_, err = NewBasicAuthenticator("\"username\"", "password")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(ERRORMSG_PROP_INVALID, "Username").Error(), err.Error())

	_, err = NewBasicAuthenticator("", "password")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(ERRORMSG_PROP_MISSING, "Username").Error(), err.Error())

	authenticator, err = NewBasicAuthenticator("username", "password")
	assert.NotNil(t, authenticator)
	assert.Nil(t, err)
}

func TestBasicAuthPassword(t *testing.T) {
	_, err := NewBasicAuthenticator("username", "{password}")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(ERRORMSG_PROP_INVALID, "Password").Error(), err.Error())

	_, err = NewBasicAuthenticator("username", "\"password\"")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(ERRORMSG_PROP_INVALID, "Password").Error(), err.Error())

	_, err = NewBasicAuthenticator("username", "")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Errorf(ERRORMSG_PROP_MISSING, "Password").Error(), err.Error())

	authenticator, err := NewBasicAuthenticator("username", "password")
	assert.NotNil(t, authenticator)
	assert.Nil(t, err)
}

func TestBasicAuthAuthenticate(t *testing.T) {
	authenticator := &BasicAuthenticator{
		Username: "foo",
		Password: "bar",
	}

	assert.Equal(t, authenticator.AuthenticationType(), AUTHTYPE_BASIC)

	// Create a new Request object.
	builder, err := NewRequestBuilder("GET").ConstructHTTPURL("https://localhost/placeholder/url", nil, nil)
	assert.Nil(t, err)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Test the "Authenticate" method to make sure the correct header is added to the Request.
	_ = authenticator.Authenticate(request)
	assert.Equal(t, request.Header.Get("Authorization"), "Basic Zm9vOmJhcg==")
}

func TestNewBasicAuthenticatorFromMap(t *testing.T) {
	_, err := newBasicAuthenticatorFromMap(nil)
	assert.NotNil(t, err)

	var props = map[string]string{
		PROPNAME_USERNAME: "my-user",
		PROPNAME_PASSWORD: "",
	}
	_, err = newBasicAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_USERNAME: "",
		PROPNAME_PASSWORD: "my-password",
	}
	_, err = newBasicAuthenticatorFromMap(props)
	assert.NotNil(t, err)

	props = map[string]string{
		PROPNAME_USERNAME: "mookie",
		PROPNAME_PASSWORD: "betts",
	}
	authenticator, err := newBasicAuthenticatorFromMap(props)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, "mookie", authenticator.Username)
	assert.Equal(t, "betts", authenticator.Password)
}
