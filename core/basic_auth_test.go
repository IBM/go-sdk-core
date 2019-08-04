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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicAuthUsername(t *testing.T) {
	config := BasicAuthConfig{
		Username: "{username}",
		Password: "password",
	}

	err := config.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, BASICAUTH_USERNAME_INVALID, err.Error())

	config = BasicAuthConfig{
		Username: "\"username\"",
		Password: "password",
	}

	err = config.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, BASICAUTH_USERNAME_INVALID, err.Error())

	config = BasicAuthConfig{
		Username: "",
		Password: "password",
	}

	err = config.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, BASICAUTH_USERNAME_PASSWORD_EMPTY, err.Error())
}

func TestBasicAuthPassword(t *testing.T) {
	config := BasicAuthConfig{
		Username: "username",
		Password: "{password}",
	}

	err := config.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, BASICAUTH_PASSWORD_INVALID, err.Error())

	config = BasicAuthConfig{
		Username: "username",
		Password: "\"password\"",
	}

	err = config.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, BASICAUTH_PASSWORD_INVALID, err.Error())

	config = BasicAuthConfig{
		Username: "username",
		Password: "",
	}

	err = config.Validate()
	assert.NotNil(t, err)
	assert.Equal(t, BASICAUTH_USERNAME_PASSWORD_EMPTY, err.Error())
}

func TestBasicAuthAuthenticate(t *testing.T) {
	// Create a basic auth config.
	config := &BasicAuthConfig{
		Username: "foo",
		Password: "bar",
	}

	// Create a BasicAuthenticator instance from this config.
	authenticator, err := NewAuthenticator(config)
	assert.Nil(t, err)
	assert.NotNil(t, authenticator)
	assert.Equal(t, authenticator.AuthenticationType(), "basic")

	// Create a new Request object.
	request, err := NewRequestBuilder("GET").
		ConstructHTTPURL("https://localhost/placeholder/url", nil, nil).
		Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)

	// Test the "Authenticate" method to make sure the correct header is added to the Request.
	authenticator.Authenticate(request)
	assert.NotNil(t, request.Header.Get("Authorization"))
	assert.Equal(t, request.Header.Get("Authorization"), "Basic Zm9vOmJhcg==")
}

func TestBasicAuthAuthenticateError(t *testing.T) {
	// Create an incorrect basic auth config.
	config := &BasicAuthConfig{
		Username: "",
		Password: "bar",
	}

	// Make sure that NewAuthenticator returns an error.
	authenticator, err := NewAuthenticator(config)
	assert.NotNil(t, err)
	assert.Nil(t, authenticator)
}
