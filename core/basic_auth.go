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
	"fmt"
	"net/http"
)

const (
	// Basic Auth-related error error messages.
	BASICAUTH_USERNAME_INVALID        = "The Username shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding the Username field."
	BASICAUTH_PASSWORD_INVALID        = "The Password shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding the Password field."
	BASICAUTH_USERNAME_PASSWORD_EMPTY = "The Username and Password fields must be non-empty string values."
)

// This struct contains the configuration associated with the Basic Authenticator.
// Both the Username and Password fields must be set to non-empty string values.
type BasicAuthConfig struct {
	Username string
	Password string
}

func (this BasicAuthConfig) Validate() error {
	if HasBadFirstOrLastChar(this.Username) {
		return fmt.Errorf(BASICAUTH_USERNAME_INVALID)
	}
	if HasBadFirstOrLastChar(this.Password) {
		return fmt.Errorf(BASICAUTH_PASSWORD_INVALID)
	}

	if this.Username == "" || this.Password == "" {
		return fmt.Errorf(BASICAUTH_USERNAME_PASSWORD_EMPTY)
	}

	return nil
}

func (BasicAuthConfig) AuthenticationType() string {
	return AUTHTYPE_BASIC
}

type BasicAuthenticator struct {
	config *BasicAuthConfig
}

func NewBasicAuthenticator(configObj *BasicAuthConfig) *BasicAuthenticator {
	return &BasicAuthenticator{
		config: configObj,
	}
}

func (BasicAuthenticator) AuthenticationType() string {
	return AUTHTYPE_BASIC
}

func (this BasicAuthenticator) Authenticate(request *http.Request) error {
	request.SetBasicAuth(this.config.Username, this.config.Password)
	return nil
}
