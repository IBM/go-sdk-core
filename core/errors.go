package core

// (C) Copyright IBM Corp. 2024.
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
	"errors"
	"fmt"
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

	// Error returns the message associated with a given error and guarantees
	// every instance of Problem also implements the native `error` interface.
	Error() string
}

// IBMError holds the base set of fields that all error types
// should include. It is geared towards embedding in other
// structs and it should not be used on its own (so it is not exported).
type IBMError struct {

	// Summary is the informative, user-friendly message that describes
	// the error and what caused it.
	Summary string `json:"summary" validate:"required"` // required

	// System describes the actual component or tool that the error
	// occurred in. For example, an error that occurs in this library
	// will have a system value of "go-sdk-core".
	System string `json:"system" validate:"required"` // required

	// Version provides the version of the component or tool that the
	// error occurred in.
	Version string `json:"version" validate:"required"` // required

	// discriminator is a private property that is not ever meant to be
	// seen by the end user. It's sole purpose is to enforce uniqueness
	// for the computed ID of errors that would otherwise have the same
	// ID. For example, if two SDKError instances are created with the
	// same System and Function values, they would end up with the same
	// ID. This property allows us to "discriminate" between such errors.
	discriminator string `json:"discriminator,omitempty"` // optional

	// causedBy allows for the storage of an error from a previous system,
	// if there is one.
	causedBy Problem `json:"caused_by,omitempty"` // optional
}

// Error returns the error message and implements the native
// `error` interface.
func (e *IBMError) Error() string {
	return e.Summary
}

// GetBaseSignature provides a convenient way of
// retrieving the fields needed to compute the
// error ID that are common to every kind of error.
func (e *IBMError) GetBaseSignature() string {
	return fmt.Sprintf("%s%s%s", e.System, e.discriminator, getPreviousErrorID(e.causedBy))
}

// GetCausedBy returns the underlying `causedBy` error, if it exists.
func (e *IBMError) GetCausedBy() Problem {
	return e.causedBy
}

// Unwrap implements an interface the native Go "errors" package uses to
// check for embedded errors in a given error instance. IBM error types
// are not embedded in the traditional sense, but they chain previous
// error instances together with the "causedBy" field. This allows error
// interface instances to be cast into any of the error types in the chain
// using the native "errors.As" function. This can be useful for, as an
// example, extracting an HTTPError from the chain if it exists.
// Note that this Unwrap method returns only the chain of "caused by" errors;
// it does not include the error instance the method is called on - that is
// looked at separately by the "errors" package in functions like "As".
func (e *IBMError) Unwrap() []error {
	causedBy := e.GetCausedBy() 
	if causedBy == nil {
		return nil
	}

	errs := []error{ causedBy }

	var toUnwrap interface{ Unwrap() []error }
	if errors.As(causedBy, &toUnwrap) {
		causedByChain := toUnwrap.Unwrap()
		if causedByChain != nil {
			errs = append(errs, causedByChain...)
		}
	}

	return errs
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
	Function string `json:"function" validate:"required"` // required

	// A computed stack trace including the relevant
	// function names, files, and line numbers invoked
	// leading up to the origination of the error.
	stack []sdkStackFrame `json:"stack,omitempty"` // optional
}

// GetConsoleMessage returns all public fields of
// the error, formatted in YAML.
func (e *SDKError) GetConsoleMessage() string {
	return ComputeConsoleMessage(e)
}

// GetDebugMessage returns all information
// about the error, formatted in YAML.
func (e *SDKError) GetDebugMessage() string {
	return ComputeDebugMessage(e)
}

// GetID returns the computed identifier, computed from the
// `System`, `discriminator`, and `Function` fields, as well as the
// identifier of the `causedBy` error, if it exists.
func (e *SDKError) GetID() string {
	return CreateIDHash("sdk_error", e.GetBaseSignature(), e.Function)
}

// SDKError provides a type suited to errors that
// occur as the result of an HTTP request. It extends
// the base `IBMError` type with fields to store
// information about the HTTP request/response.
type HTTPError struct {
	*IBMError

	// OperationID identifies the operation of an API
	// that the failed request was made to.
	OperationID string `json:"operation_id,omitempty"`

	// Response contains the full HTTP error response
	// returned as a result of the failed request,
	// including the body and all headers.
	Response *DetailedResponse `json:"response" validate:"required"`

	// ErrorCode is the code returned from the API
	// in the error response, identifying the issue.
	ErrorCode string `json:"error_code,omitempty"`
}

// GetConsoleMessage returns all public fields of
// the error, formatted in YAML.
func (e *HTTPError) GetConsoleMessage() string {
	return ComputeConsoleMessage(e)
}

// GetDebugMessage returns all information about
// the error, formatted in YAML.
func (e *HTTPError) GetDebugMessage() string {
	return ComputeDebugMessage(e)
}

// GetID returns the computed identifier, computed from the
// `System`, `discriminator`, `OperationID`, `Response`, and
// `ErrorCode` fields, as well as the identifier of the
// `causedBy` error, if it exists.
func (e *HTTPError) GetID() string {
	// TODO: add the ErrorCode to the hash once we have the ability to enumerate error codes in an API.
	return CreateIDHash("http_error", e.GetBaseSignature(), e.OperationID, fmt.Sprint(e.Response.GetStatusCode()))
}

// AuthenticationError describes the error returned when
// authentication over HTTP fails.
type AuthenticationError struct {
	Err      error `json:"err,omitempty"`
	*HTTPError
}

// infoProvider is a function type that must return two strings:
// first, the name of the system (e.g. "go-sdk-core")
// and second, the semantic version number as a string (e.g. "5.1.2")
type infoProvider func() (string, string)

// Error creation functions

// IBMErrorf creates and returns a new instance of an
// IBMError struct. It is private as it is primarily
// meant for embedding IBMError structs in other types.
func IBMErrorf(err error, summary, system, version, discriminator string) *IBMError {
	// Leaving summary blank is a convenient way to
	// use the message from the underlying error.
	if summary == "" {
		summary = err.Error()
	}

	newError := &IBMError{
		Summary: summary,
		System: system,
		Version: version,
		discriminator: discriminator,
	}

	var causedBy Problem
	if errors.As(err, &causedBy) {
		newError.causedBy = causedBy
	}

	return newError
}

// SDKErrorf creates and returns a new instance of `SDKError`.
func SDKErrorf(err error, summary, discriminator string, getInfo infoProvider) *SDKError {
	system, version := getInfo()

	function := computeFunctionName(system)
	stack := getStackInfo(system)

	return &SDKError{
		IBMError: IBMErrorf(err, summary, system, version, discriminator),
		Function: function,
		stack: stack,
	}
}

// RepurposeSDKError provides a convenient way to take an error from
// another function in the same system and contextualize it to the current
// function. Should only be used in public (exported) functions.
func RepurposeSDKError(err error, discriminator string) error {
	if err == nil {
		return err
	}

	// It only makes sense to carry out this logic with SDK Errors.
	var sdkErr *SDKError
	if !errors.As(err, &sdkErr) {
		return err
	}

	// Special behavior to allow errors coming from a method that wraps a
	// "*WithContext" method to maintain the discriminator of the originating
	// error. Otherwise, we would lose all of that data in the wrap.
	if discriminator != "" {
		sdkErr.discriminator = discriminator
	}

	// Recompute the function to reflect this public boundary (but let the stack
	// remain as it is - it is the path to the original error origination point).
	sdkErr.Function = computeFunctionName(sdkErr.System)

	return sdkErr
}

// httpErrorf creates and returns a new instance of `HTTPError`.
func httpErrorf(summary string, response *DetailedResponse) *HTTPError {
	return &HTTPError{
		IBMError: IBMErrorf(nil, summary, "", "", ""),
		Response: response,
	}
}

// NewAuthenticationError is a deprecated function that was previously used for creating new
// AuthenticationError structs. HTTPError types should be used instead of AuthenticationError types.
func NewAuthenticationError(response *DetailedResponse, err error) *AuthenticationError {
	GetLogger().Warn("NewAuthenticationError is deprecated and should not be used.")
	authError := authenticationErrorf(err, response, "unknown", func() (string, string) { return "unknown", "unknown" })
	return authError
}

// authenticationErrorf creates and returns a new instance of `AuthenticationError`.
func authenticationErrorf(err error, response *DetailedResponse, operationID string, getInfo infoProvider) *AuthenticationError {
	// This function should always be called with non-nil
	// error/DetailedResponse instances.
	if err == nil || response == nil {
		return nil
	}

	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		httpErr = httpErrorf(err.Error(), response)
	}

	enrichHTTPError(httpErr, operationID, getInfo)

	return &AuthenticationError{
		HTTPError: httpErr,
		Err: err,
	}
}

// OrderableProblem provides an interface for retrieving ordered
// representations of errors in order to print YAML messages
// with a controlled ordering of the fields.
type OrderableProblem interface {
	GetConsoleOrderedMaps() *OrderedMaps
	GetDebugOrderedMaps() *OrderedMaps
}

// GetConsoleOrderedMaps returns an ordered-map representation
// of an SDKError instance suited for a console message.
func (e *SDKError) GetConsoleOrderedMaps() *OrderedMaps {
	orderedMaps := NewOrderedMaps()

	orderedMaps.Add("id", e.GetID())
	orderedMaps.Add("summary", e.Summary)
	orderedMaps.Add("function", e.Function)
	orderedMaps.Add("system", e.System)
	orderedMaps.Add("version", e.Version)

	return orderedMaps
}

// GetDebugOrderedMaps returns an ordered-map representation
// of an SDKError instance, with additional information
// suited for a debug message.
func (e *SDKError) GetDebugOrderedMaps() *OrderedMaps {
	orderedMaps := e.GetConsoleOrderedMaps()

	orderedMaps.Add("stack", e.stack)

	var orderableCausedBy OrderableProblem
	if errors.As(e.GetCausedBy(), &orderableCausedBy) {
		orderedMaps.Add("caused_by", orderableCausedBy.GetDebugOrderedMaps().GetMaps())
	}

	return orderedMaps
}

// GetConsoleOrderedMaps returns an ordered-map representation
// of an HTTPError instance suited for a console message.
func (e *HTTPError) GetConsoleOrderedMaps() *OrderedMaps {
	orderedMaps := NewOrderedMaps()

	orderedMaps.Add("id", e.GetID())
	orderedMaps.Add("summary", e.Summary)
	orderedMaps.Add("operation_id", e.OperationID)
	orderedMaps.Add("error_code", e.ErrorCode)
	orderedMaps.Add("system", e.System)
	orderedMaps.Add("version", e.Version)

	return orderedMaps
}

// GetDebugOrderedMaps returns an ordered-map representation
// of an HTTPError instance, with additional information
// suited for a debug message.
func (e *HTTPError) GetDebugOrderedMaps() *OrderedMaps {
	orderedMaps := e.GetConsoleOrderedMaps()

	// The RawResult is never helpful in the printed message. Create a hard copy
	// (de-referenced pointer) to remove the raw result from so we don't alter
	// the response stored in the error object.
	printableResponse := *e.Response
	if printableResponse.Result == nil {
		printableResponse.Result = string(printableResponse.RawResult)
	}
	printableResponse.RawResult = nil
	orderedMaps.Add("response", printableResponse)

	var orderableCausedBy OrderableProblem
	if errors.As(e.GetCausedBy(), &orderableCausedBy) {
		orderedMaps.Add("caused_by", orderableCausedBy.GetDebugOrderedMaps().GetMaps())
	}

	return orderedMaps
}

// GetConsoleOrderedMaps returns an ordered-map representation
// of an AuthenticationError instance suited for a console message.
// Note: Added for compatibility - this is not intended to be used.
func (e *AuthenticationError) GetConsoleOrderedMaps() *OrderedMaps {
	orderedMaps := NewOrderedMaps()

	orderedMaps.Add("id", e.GetID())
	orderedMaps.Add("summary", e.Summary)
	orderedMaps.Add("system", e.System)
	orderedMaps.Add("version", e.Version)

	return orderedMaps
}

// GetDebugOrderedMaps returns an ordered-map representation
// of an AuthenticationError instance, with additional information
// suited for a debug message.
// Note: Added for compatibility - this is not intended to be used.
func (e *AuthenticationError) GetDebugOrderedMaps() *OrderedMaps {
	orderedMaps := e.GetConsoleOrderedMaps()

	orderedMaps.Add("response", e.Response)

	var orderableCausedBy OrderableProblem
	if errors.As(e.GetCausedBy(), &orderableCausedBy) {
		orderedMaps.Add("caused_by", orderableCausedBy.GetDebugOrderedMaps().GetMaps())
	}

	return orderedMaps
}
