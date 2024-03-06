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
// behavior of custom IBM problem message types.
type Problem interface {

	// GetConsoleMessage returns a message suited to the practitioner
	// or end user. It should tell the user what went wrong, and why,
	// without unnecessary implementation details.
	GetConsoleMessage() string

	// GetDebugMessage returns a message suited to the developer, in
	// order to assist in debugging. It should give enough information
	// for the developer to identify the root cause of the issue.
	GetDebugMessage() string

	// GetID returns an identifier or code for a given problem. It is computed
	// from the attributes of the problem, so that the same problems will always
	// have the same ID, even when encountered by different users.
	GetID() string

	// Error returns the message associated with a given problem and guarantees
	// every instance of Problem also implements the native `error` interface.
	Error() string
}

// IBMProblem holds the base set of fields that all problem types
// should include. It is geared towards embedding in other
// structs and it should not be used on its own (so it is not exported).
type IBMProblem struct {

	// Summary is the informative, user-friendly message that describes
	// the problem and what caused it.
	Summary string

	// Component is a structure providing information about the actual
	// component that the problem occurred in: the name of the component
	// and the version of the component being used with the problem occurred.
	// Examples of components include cloud services, SDK clients, the IBM
	// Terraform Provider, etc. For programming libraries, the Component name
	// should match the module name for the library (i.e. the name a user
	// would use to install it).
	Component *ProblemComponent

	// Severity represents the severity level of the problem, e.g.
	// error, warning, or info.
	Severity ProblemSeverity

	// discriminator is a private property that is not ever meant to be
	// seen by the end user. It's sole purpose is to enforce uniqueness
	// for the computed ID of problems that would otherwise have the same
	// ID. For example, if two SDKProblem instances are created with the
	// same Component and Function values, they would end up with the same
	// ID. This property allows us to "discriminate" between such problems.
	discriminator string // optional

	// causedBy allows for the storage of a problem from a previous component,
	// if there is one.
	causedBy Problem // optional
}

// Error returns the problem's message and implements the native
// `error` interface.
func (e *IBMProblem) Error() string {
	return e.Summary
}

// GetBaseSignature provides a convenient way of
// retrieving the fields needed to compute the
// hash that are common to every kind of problem.
func (e *IBMProblem) GetBaseSignature() string {
	return fmt.Sprintf("%s%s%s%s", e.Component.Name, e.Severity, e.discriminator, getPreviousProblemID(e.causedBy))
}

// GetCausedBy returns the underlying `causedBy` problem, if it exists.
func (e *IBMProblem) GetCausedBy() Problem {
	return e.causedBy
}

// Unwrap implements an interface the native Go "errors" package uses to
// check for embedded problems in a given problem instance. IBM problem types
// are not embedded in the traditional sense, but they chain previous
// problem instances together with the "causedBy" field. This allows error
// interface instances to be cast into any of the problem types in the chain
// using the native "errors.As" function. This can be useful for, as an
// example, extracting an HTTPProblem from the chain if it exists.
// Note that this Unwrap method returns only the chain of "caused by" problems;
// it does not include the error instance the method is called on - that is
// looked at separately by the "errors" package in functions like "As".
func (e *IBMProblem) Unwrap() []error {
	causedBy := e.GetCausedBy()
	if causedBy == nil {
		return nil
	}

	errs := []error{causedBy}

	var toUnwrap interface{ Unwrap() []error }
	if errors.As(causedBy, &toUnwrap) {
		causedByChain := toUnwrap.Unwrap()
		if causedByChain != nil {
			errs = append(errs, causedByChain...)
		}
	}

	return errs
}

// SDKProblem provides a type suited to problems that
// occur in SDK projects. It extends the base
// `IBMProblem` type with a field to store the
// function being called when the problem occurs.
type SDKProblem struct {
	*IBMProblem

	// Function provides the name of the in-code
	// function or method in which the problem
	// occurred.
	Function string `json:"function" validate:"required"` // required

	// A computed stack trace including the relevant
	// function names, files, and line numbers invoked
	// leading up to the origination of the problem.
	stack []sdkStackFrame // optional
}

// GetConsoleMessage returns all public fields of
// the problem, formatted in YAML.
func (e *SDKProblem) GetConsoleMessage() string {
	return ComputeConsoleMessage(e)
}

// GetDebugMessage returns all information
// about the problem, formatted in YAML.
func (e *SDKProblem) GetDebugMessage() string {
	return ComputeDebugMessage(e)
}

// GetID returns the computed identifier, computed from the
// `Component`, `discriminator`, and `Function` fields, as well as the
// identifier of the `causedBy` problem, if it exists.
func (e *SDKProblem) GetID() string {
	return CreateIDHash("sdk", e.GetBaseSignature(), e.Function)
}

// SDKProblem provides a type suited to problems that
// occur as the result of an HTTP request. It extends
// the base `IBMProblem` type with fields to store
// information about the HTTP request/response.
type HTTPProblem struct {
	*IBMProblem

	// OperationID identifies the operation of an API
	// that the failed request was made to.
	OperationID string

	// ErrorCode is the code returned from the API
	// in the error response, identifying the issue.
	ErrorCode string

	// Response contains the full HTTP error response
	// returned as a result of the failed request,
	// including the body and all headers.
	Response *DetailedResponse
}

// GetConsoleMessage returns all public fields of
// the problem, formatted in YAML.
func (e *HTTPProblem) GetConsoleMessage() string {
	return ComputeConsoleMessage(e)
}

// GetDebugMessage returns all information about
// the problem, formatted in YAML.
func (e *HTTPProblem) GetDebugMessage() string {
	return ComputeDebugMessage(e)
}

// GetID returns the computed identifier, computed from the
// `Component`, `discriminator`, `OperationID`, `Response`, and
// `ErrorCode` fields, as well as the identifier of the
// `causedBy` problem, if it exists.
func (e *HTTPProblem) GetID() string {
	// TODO: add the ErrorCode to the hash once we have the ability to enumerate error codes in an API.
	return CreateIDHash("http", e.GetBaseSignature(), e.OperationID, fmt.Sprint(e.Response.GetStatusCode()))
}

func (e *HTTPProblem) getHeader(key string) (string, bool) {
	value := e.Response.Headers.Get(key)
	return value, value != ""
}

// AuthenticationError describes the problem returned when
// authentication over HTTP fails.
type AuthenticationError struct {
	Err error `json:"err,omitempty"`
	*HTTPProblem
}

// ProblemSeverity simulates an enum by defining a string type that should
// be one of a few given values. For now, ErrorSeverity is the only supported
// value.
type ProblemSeverity string

// ProblemComponent is a structure that holds information about a given component.
type ProblemComponent struct {
	Name string
	Version string
}

// Note: this doesn't actually provide type safety like a real enum would but
// it serves as helpful documentation for understanding expected values.
const (
	ErrorSeverity ProblemSeverity = "error"
	WarningSeverity ProblemSeverity = "warning"
)

// Error creation functions

func ibmProblemf(err error, severity ProblemSeverity, component *ProblemComponent, summary, discriminator string) *IBMProblem {
	// Leaving summary blank is a convenient way to
	// use the message from the underlying problem.
	if summary == "" {
		summary = err.Error()
	}

	newError := &IBMProblem{
		Summary:       summary,
		Component:     component,
		discriminator: discriminator,
		Severity:      severity,
	}

	var causedBy Problem
	if errors.As(err, &causedBy) {
		newError.causedBy = causedBy
	}

	return newError
}

// IBMErrorf creates and returns a new instance of an IBMProblem struct with "error"
// level severity. It is primarily meant for embedding IBMProblem structs in other types.
func IBMErrorf(err error, component *ProblemComponent, summary, discriminator string) *IBMProblem {
	return ibmProblemf(err, ErrorSeverity, component, summary, discriminator)
}

// SDKErrorf creates and returns a new instance of `SDKProblem` with "error" level severity.
func SDKErrorf(err error, summary, discriminator string, component *ProblemComponent) *SDKProblem {
	function := computeFunctionName(component.Name)
	stack := getStackInfo(component.Name)

	return &SDKProblem{
		IBMProblem: IBMErrorf(err, component, summary, discriminator),
		Function:   function,
		stack:      stack,
	}
}

// RepurposeSDKProblem provides a convenient way to take a problem from
// another function in the same component and contextualize it to the current
// function. Should only be used in public (exported) functions.
func RepurposeSDKProblem(err error, discriminator string) error {
	if err == nil {
		return err
	}

	// It only makes sense to carry out this logic with SDK Errors.
	var sdkErr *SDKProblem
	if !errors.As(err, &sdkErr) {
		return err
	}

	// Special behavior to allow SDK problems coming from a method that wraps a
	// "*WithContext" method to maintain the discriminator of the originating
	// problem. Otherwise, we would lose all of that data in the wrap.
	if discriminator != "" {
		sdkErr.discriminator = discriminator
	}

	// Recompute the function to reflect this public boundary (but let the stack
	// remain as it is - it is the path to the original problem origination point).
	sdkErr.Function = computeFunctionName(sdkErr.Component.Name)

	return sdkErr
}

// httpErrorf creates and returns a new instance of `HTTPProblem` with "error" level severity.
func httpErrorf(summary string, response *DetailedResponse) *HTTPProblem {
	return &HTTPProblem{
		IBMProblem: IBMErrorf(nil, nil, summary, ""),
		Response:   response,
	}
}

// NewAuthenticationError is a deprecated function that was previously used for creating new
// AuthenticationError structs. HTTPProblem types should be used instead of AuthenticationError types.
func NewAuthenticationError(response *DetailedResponse, err error) *AuthenticationError {
	GetLogger().Warn("NewAuthenticationError is deprecated and should not be used.")
	authError := authenticationErrorf(err, response, "unknown", NewProblemComponent("unknown", "unknown"))
	return authError
}

// authenticationErrorf creates and returns a new instance of `AuthenticationError`.
func authenticationErrorf(err error, response *DetailedResponse, operationID string, component *ProblemComponent) *AuthenticationError {
	// This function should always be called with non-nil
	// error/DetailedResponse instances.
	if err == nil || response == nil {
		return nil
	}

	var httpErr *HTTPProblem
	if !errors.As(err, &httpErr) {
		httpErr = httpErrorf(err.Error(), response)
	}

	enrichHTTPProblem(httpErr, operationID, component)

	return &AuthenticationError{
		HTTPProblem: httpErr,
		Err:         err,
	}
}

// OrderableProblem provides an interface for retrieving ordered
// representations of problems in order to print YAML messages
// with a controlled ordering of the fields.
type OrderableProblem interface {
	GetConsoleOrderedMaps() *OrderedMaps
	GetDebugOrderedMaps() *OrderedMaps
}

// GetConsoleOrderedMaps returns an ordered-map representation
// of an SDKProblem instance suited for a console message.
func (e *SDKProblem) GetConsoleOrderedMaps() *OrderedMaps {
	orderedMaps := NewOrderedMaps()

	orderedMaps.Add("id", e.GetID())
	orderedMaps.Add("summary", e.Summary)
	orderedMaps.Add("severity", e.Severity)
	orderedMaps.Add("function", e.Function)
	orderedMaps.Add("component", e.Component)

	return orderedMaps
}

// GetDebugOrderedMaps returns an ordered-map representation
// of an SDKProblem instance, with additional information
// suited for a debug message.
func (e *SDKProblem) GetDebugOrderedMaps() *OrderedMaps {
	orderedMaps := e.GetConsoleOrderedMaps()

	orderedMaps.Add("stack", e.stack)

	var orderableCausedBy OrderableProblem
	if errors.As(e.GetCausedBy(), &orderableCausedBy) {
		orderedMaps.Add("caused_by", orderableCausedBy.GetDebugOrderedMaps().GetMaps())
	}

	return orderedMaps
}

// GetConsoleOrderedMaps returns an ordered-map representation
// of an HTTPProblem instance suited for a console message.
func (e *HTTPProblem) GetConsoleOrderedMaps() *OrderedMaps {
	orderedMaps := NewOrderedMaps()

	orderedMaps.Add("id", e.GetID())
	orderedMaps.Add("summary", e.Summary)
	orderedMaps.Add("severity", e.Severity)
	orderedMaps.Add("operation_id", e.OperationID)
	orderedMaps.Add("status_code", fmt.Sprint(e.Response.GetStatusCode()))
	if e.ErrorCode != "" {
		orderedMaps.Add("error_code", e.ErrorCode)
	}
	orderedMaps.Add("component", e.Component)

	// Conditionally add the request ID and correlation ID header values.

	if header, ok := e.getHeader("x-request-id"); ok {
		orderedMaps.Add("request_id", header)
	}

	if header, ok := e.getHeader("x-correlation-id"); ok {
		orderedMaps.Add("correlation_id", header)
	}

	return orderedMaps
}

// GetDebugOrderedMaps returns an ordered-map representation
// of an HTTPProblem instance, with additional information
// suited for a debug message.
func (e *HTTPProblem) GetDebugOrderedMaps() *OrderedMaps {
	orderedMaps := e.GetConsoleOrderedMaps()

	// The RawResult is never helpful in the printed message. Create a hard copy
	// (de-referenced pointer) to remove the raw result from so we don't alter
	// the response stored in the problem object.
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
	orderedMaps.Add("severity", e.Severity)
	orderedMaps.Add("component", e.Component)

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
