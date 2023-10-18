package core

// (C) Copyright IBM Corp. 2023.
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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ghodss/yaml"
	"strings"
)

// Type definitions

// Problem is an interface that describes the common
// behavior of custom IBM error message types.
type Problem interface {

	// GetConsoleMessage returns a message suited to the practitioner
	// or end user. It should tell the user what went wrong, and why,
	// without unnecessary implementation details.
	GetConsoleMessage() string

	// GetDebugMessage returns a message suited to the developer, in
	// order to assist in debugging. It should give enough information
	// for the developer to identify the root cause of the issue.
	GetDebugMessage() string

	// GetID returns an identifier or code for a given error. It is computed
	// from the attributes of the error, so that the same errors will always
	// have the same ID, even when encountered by different users.
	GetID() string
}

// IBMError holds the base set of fields that all error types
// should include. It is geared towards embedding in other
// structs, rather than for usage on its own.
type IBMError struct {

	// Summary is the informative, user-friendly message that describes
	// the error and what caused it.
	Summary string // required

	// System describes the actual component or tool that the error
	// occurred in. For example, an error that occurs in this library
	// will have a system value of "go-sdk-core".
	System string // required

	// Version provides the version of the component or tool that the
	// error occurred in.
	Version string // required

	// discriminator is a private property that is not ever meant to be
	// seen by the end user. It's sole purpose is to enforce uniqueness
	// for the computed ID of errors that would otherwise have the same
	// ID. For example, if two SDKError instances are created with the
	// same System and Function values, they would end up with the same
	// ID. This property allows us to "discriminate" between such errors.
	discriminator string // optional

	// causedBy allows for the storage of a "previous error", if there
	// is a relevant one.
	causedBy error // optional
}

// Error returns the error message and implements the native
// `error` interface.
func (e *IBMError) Error() string {
	return e.Summary
}

// GetConsoleMessage returns the error message.
func (e *IBMError) GetConsoleMessage() string {
	return e.Summary
}

// GetDebugMessage returns all public fields of
// the error, formatted in YAML.
func (e *IBMError) GetDebugMessage() string {
	// TODO: should this also include the debug message for any
	// "causedBy" errors, if they exist?
	return getErrorInfoAsYAML(e)
}

// GetID returns the computed identifier, computed from the
// `System` and `discriminator` fields, as well as the
// identifier of the `causedBy` error, if it exists.
func (e *IBMError) GetID() string {
	// TODO: should the prefix be error-type based or system based (i.e. "go-sdk-core") ?
	return createIDHash("ibm_error", e.getBaseSignature())
}

// getBaseSignature provides a convenient way of
// retrieving the fields needed to compute the
// error ID that are common to every kind of error.
func (e *IBMError) getBaseSignature() string {
	return fmt.Sprintf("%s%s%s", e.System, e.discriminator, getPreviousErrorID(e.causedBy))
}

// SDKError provides a type suited to errors that
// occur in SDK projects. It extends the base
// `IBMError` type with a field to store the
// function being called when the error occurs.
type SDKError struct {
	*IBMError

	// Function provides the name of the in-code
	// function or method in which the error
	// occurred.
	Function string // required
}

// GetDebugMessage returns all public fields of
// the error, formatted in YAML.
func (e *SDKError) GetDebugMessage() string {
	return getErrorInfoAsYAML(e)
}

// GetID returns the computed identifier, computed from the
// `System`, `discriminator`, and `Function` fields, as well as the
// identifier of the `causedBy` error, if it exists.
func (e *SDKError) GetID() string {
	return createIDHash("sdk_error", e.getBaseSignature(), e.Function)
}

// SDKError provides a type suited to errors that
// occur as the result of an HTTP request. It extends
// the base `IBMError` type with fields to store
// information about the HTTP request/response.
type HTTPError struct {
	*IBMError

	// OperationID identifies the operation of an API
	// that the failed request was made to.
	OperationID string

	// Response contains the full HTTP error response
	// returned as a result of the failed request,
	// including the body and all headers.
	Response *DetailedResponse

	// ErrorCode is the code returned from the API
	// in the error response, identifying the issue.
	ErrorCode string
}

// GetDebugMessage returns all public fields of
// the error, formatted in YAML.
func (e *HTTPError) GetDebugMessage() string {
	return getErrorInfoAsYAML(e)
}

// GetID returns the computed identifier, computed from the
// `System`, `discriminator`, `OperationID`, `Response`, and
// `ErrorCode` fields, as well as the identifier of the
// `causedBy` error, if it exists.
func (e *HTTPError) GetID() string {
	return createIDHash("http_error", e.getBaseSignature(), e.OperationID, fmt.Sprint(e.Response.GetStatusCode()), e.ErrorCode)
}

// Error creation functions

// ibmErrorf creates and returns a new instance of an
// IBMError struct. It is private as it is primarily
// meant for embedding IBMError structs in other types.
func ibmErrorf(err error, summary, system, version, discriminator string) *IBMError {
	if summary == "" {
		summary = err.Error()
	}
	return &IBMError{
		Summary: summary,
		System: system,
		Version: version,
		discriminator: discriminator,
		causedBy: err,
	}
}

// coreSDKErrorf is a convenience function to create local instances of `SDKError`
// with a consistent System and Version value.
func coreSDKErrorf(err error, summary, discriminator, function string) *SDKError {
	system, version := getSystemInfo()
	return &SDKError{
		IBMError: ibmErrorf(err, summary, system, version, discriminator),
		Function: function,
	}
}

// SDKErrorf creates and returns a new instance of `SDKError`.
func SDKErrorf(err error, summary, system, version, discriminator, function string) *SDKError {
	return &SDKError{
		IBMError: ibmErrorf(err, summary, system, version, discriminator),
		Function: function,
	}
}

// coreHTTPErrorf is a convenience function to create local instances of `HTTPError`
// with a consistent System and Version value.
func coreHTTPErrorf(err error, summary, discriminator, operationID, code string, response *DetailedResponse) *HTTPError {
	system, version := getSystemInfo()
	return &HTTPError{
		IBMError: ibmErrorf(err, summary, system, version, discriminator),
		OperationID: operationID,
		Response: response,
		ErrorCode: code,
	}
}

// HTTPErrorf creates and returns a new instance of `HTTPError`.
func HTTPErrorf(err error, summary, system, version, discriminator, operationID, code string, response *DetailedResponse) *HTTPError {
	return &HTTPError{
		IBMError: ibmErrorf(err, summary, system, version, discriminator),
		OperationID: operationID,
		Response: response,
		ErrorCode: code,
	}
}

// rewrapSDKError provides a convenient way to modify the Function
// field of an SDKError when the error is being returned through
// a public boundary, in order to keep the field scoped to the
// function the user is actually calling. It maintains the rest
// of the error details.
func rewrapSDKError(err error, function string) error {
	if err != nil {
		if sdkError, ok := err.(*SDKError); ok {
			sdkError.Function = function
			err = sdkError
		}
	}

	return err
}

// TODO: consider accepting an options model for future-oriented flexibility
// (for all functions in this file).

// RewrapClientError provides a convenient way for depending libraries (e.g. generated
// SDKs) to take errors coming from this library and modify them based on the
// public context of that tool.
func RewrapClientError(err error, system, version, function, operationID string) error {
		if err == nil {
			return err
		}

		// Keep error details but modify based on new context.
		switch _err := err.(type) {
		case SDKError:
			_err.System = system
			_err.Version = version
			_err.Function = function
			err = _err
		case HTTPError:
			_err.System = system
			_err.Version = version

			// This is especially useful because the core library doesn't
			// have access to operation IDs when creating errors. The
			// invoking SDK can use this function to add that information.
			_err.OperationID = operationID
			err = _err
		}

		return err
}


// Utility functions

// createIDHash computes a unique ID based on a given prefix
// and error attribute fields.
func createIDHash(prefix string, fields ...string) string {
	signature := strings.Join(fields, "")
	hash := sha256.Sum256([]byte(signature))
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(hash[:4]))
}

// getPreviousErrorID looks at the "causedBy" error and if it
// is an instance of a "Problem", returns the ID.
func getPreviousErrorID(err error) string {
	if (err != nil) {
		// It only makes sense to look for an ID if it is an
		// instance of Problem and not just a basic Go error
		if problem, ok := err.(Problem); ok {
			return problem.GetID()
		}
	}
	return ""
}

// getErrorInfoAsYAML formats all of the public fields of the struct
// implementing Problem as YAML and returns the data as a string.
func getErrorInfoAsYAML(problem Problem) string {
	// This gets called from each "Problem"'s `GetDebugMessage` method,
	// so don't call that here!
	yamlifiedStruct, err := yaml.Marshal(problem)
	if err != nil {
		return fmt.Sprintf("%s\n\nError serializing the error information: %s", problem.GetConsoleMessage(), err.Error())
	}
	return fmt.Sprintf("---\n%s---\n", yamlifiedStruct)
}

// getSystemInfo is a convenient way to access the name of the
// system alongside the current semantic version of the library.
func getSystemInfo() (string, string) {
	return "go-sdk-core", __VERSION__
}
