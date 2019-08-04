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
	// Supported authentication types.
	AUTHTYPE_BASIC  = "basic"
	AUTHTYPE_NOAUTH = "noauth"
	AUTHTYPE_IAM    = "iam"
	AUTHTYPE_ICP4D  = "icp4d"

	// Internal constants
	TOKENTYPE_BEARER          = "Bearer"
	HEADER_NAME_AUTHORIZATION = "Authorization"

	// Common error messages
	ERRORMSG_UNKNOWN_AUTHTYPE = "Unknown authentication type: %s"
	ERRORMSG_NIL_AUTHCONFIG   = "The 'config' parameter must not be nil."
)

type AuthenticatorConfig interface {
	AuthenticationType() string
	Validate() error
}

type Authenticator interface {
	AuthenticationType() string
	Authenticate(*http.Request) error
}

// This is a "factory" method which will create a new Authenticator instance
// from the specified config.
func NewAuthenticator(config AuthenticatorConfig) (Authenticator, error) {
	if config == nil {
		return nil, fmt.Errorf(ERRORMSG_NIL_AUTHCONFIG)
	}

	err := config.Validate()
	if err != nil {
		return nil, err
	}

	if config.AuthenticationType() == AUTHTYPE_BASIC {
		basicAuthConfig := config.(*BasicAuthConfig)
		return NewBasicAuthenticator(basicAuthConfig), nil
	} else if config.AuthenticationType() == AUTHTYPE_NOAUTH {
		noauthConfig := config.(*NoauthConfig)
		return NewNoauthAuthenticator(noauthConfig), nil
	} else if config.AuthenticationType() == AUTHTYPE_IAM {
		iamConfig := config.(*IAMConfig)
		return NewIAMAuthenticator(iamConfig)
	} else if config.AuthenticationType() == AUTHTYPE_ICP4D {
		icp4dConfig := config.(*ICP4DConfig)
		return NewICP4DAuthenticator(icp4dConfig)
	} else {
		return nil, fmt.Errorf(ERRORMSG_UNKNOWN_AUTHTYPE, config.AuthenticationType())
	}
}
