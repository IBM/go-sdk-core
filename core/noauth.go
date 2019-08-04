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
	"net/http"
)

// This struct contains the configuration associated with the "NoAuth" Authenticator,
// which is a placeholder-type Authenticator which performs no authentication.
type NoauthConfig struct {
}

func (this NoauthConfig) Validate() error {
	return nil
}

func (NoauthConfig) AuthenticationType() string {
	return AUTHTYPE_NOAUTH
}

type NoauthAuthenticator struct {
}

func NewNoauthAuthenticator(configObj *NoauthConfig) *NoauthAuthenticator {
	return &NoauthAuthenticator{}
}

func (NoauthAuthenticator) AuthenticationType() string {
	return AUTHTYPE_NOAUTH
}

func (this NoauthAuthenticator) Authenticate(request *http.Request) error {
	// Nothing to do since we're not providing any authentication.
	return nil
}
