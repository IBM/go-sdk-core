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
	"net/http"
)

// Authenticator describes the set of methods implemented by each authenticator.
type Authenticator interface {
	AuthenticationType() string
	Authenticate(*http.Request) error
	Validate() error
}

// AuthenticationError describes the error returned when authentication fails
type AuthenticationError struct {
	Response *DetailedResponse
	Err      error
	*IBMError
}

func (e *AuthenticationError) Error() string {
	if e.Err == nil {
		return e.Summary
	}
	return e.Err.Error()
}
func (e *AuthenticationError) GetDebugMessage() string {
  return getErrorInfoAsYAML(e)
}
func (e *AuthenticationError) GetID() string {
  return createIDHash("auth_error", e.getBaseSignature(), fmt.Sprint(e.Response.GetStatusCode()))
}

func NewAuthenticationError(response *DetailedResponse, err error) *AuthenticationError {
	// TODO: Log a deprecation notice
	sys, ver := getSystemInfo()
	return AuthenticationErrorf(err, err.Error(), sys, ver, "deprecated", response)
}

func coreAuthenticationErrorf(err error, summary, discriminator string, response *DetailedResponse) *AuthenticationError {
	sys, ver := getSystemInfo()
	return AuthenticationErrorf(err, summary, sys, ver, discriminator, response)
}

func AuthenticationErrorf(err error, summary, system, version, discriminator string, response *DetailedResponse) *AuthenticationError {
  return &AuthenticationError{
    IBMError: ibmErrorf(err, summary, system, version, discriminator),
    Response: response,
    Err:      err,
  }
}
