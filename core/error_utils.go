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
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"runtime"
	"strings"
)

// Private utility functions for our custom error system

// CreateIDHash computes a unique ID based on a given prefix
// and error attribute fields.
func CreateIDHash(prefix string, fields ...string) string {
	signature := strings.Join(fields, "")
	hash := sha256.Sum256([]byte(signature))
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(hash[:4]))
}

// getPreviousErrorID returns the ID of the "causedBy" error, if it exists.
func getPreviousErrorID(problem Problem) string {
	if problem != nil {
		return problem.GetID()
	}
	return ""
}

// getSystemInfo is a convenient way to access the name of the
// system alongside the current semantic version of the library.
func getSystemInfo() (string, string) {
	return "github.com/IBM/go-sdk-core/v5", __VERSION__
}

// computeFunctionName investigates the program counter at a fixed
// skip number (aka point in the stack) of 2, which gives us the
// information about the function the error was created in, and
// returns the name of the function.
func computeFunctionName() string {
	if pc, _, _, ok := runtime.Caller(2); ok {
		return runtime.FuncForPC(pc).Name()
	}

	return ""
}

// getStackInfo invokes helper methods to curate a limited, formatted
// version of the stack trace with only the system-scoped function
// invocations that lead to the creation of the error.
func getStackInfo(system string) []sdkStackFrame {
	if frames, ok := makeFrames(); ok {
		return formatFrames(frames, system)
	}

	// TODO: log that we not compute the stack
	return nil
}

// makeFrames populates a program counter list with data at a
// fixed skip number (4), which gives us the stack information
// at the point in the program that the error was created. This
// function adjusts the list as needed, since the necessary
// list size is not known at first.
func makeFrames() ([]uintptr, bool) {
	pcs := make([]uintptr, 10)
	for {
		n := runtime.Callers(4, pcs)
		if n == 0 {
			return pcs, false
		}
		if n < len(pcs) {
			return pcs[:n], true
		}
		pcs = make([]uintptr, 2*len(pcs))
	}
}

// formatFrames takes a program counter list and formats them
// into a readable format for including in debug messages.
func formatFrames(pcs []uintptr, system string) []sdkStackFrame {
	result := make([]sdkStackFrame, 0)

	if len(pcs) == 0 {
		return result
	}

	// Loop to get frames.
	// A fixed number of PCs can expand to an indefinite number of Frames.
	frames := runtime.CallersFrames(pcs)
	for {
		frame, more := frames.Next()

		// Only the frames in the same system as the error are relevant.
		if strings.HasPrefix(frame.Function, system) {
			stackFrame := sdkStackFrame{
				Function: frame.Function,
				File: frame.File,
				Line: frame.Line,
			}

			result = append(result, stackFrame)
		}

		// Check whether there are more frames to process after this one.
		if !more {
			break
		}
	}

	return result
}

// getErrorInfoAsYAML formats the mapified error data as
// YAML for human/machine readable printing.
func getErrorInfoAsYAML(obj map[string]interface{}) string {
	yamlifiedStruct, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("Error serializing the error information: %s", err.Error())
	}
	return fmt.Sprintf("---\n%s---\n", yamlifiedStruct)
}

func ComputeConsoleMessage(p Problem) string {
	return getErrorInfoAsYAML(getMapWithID(p))
}

func ComputeDebugMessage(p, causedBy Problem, additionalInfo map[string]interface{}) string {
	errorAsMap := getMapWithID(p)

	// Copy any additional fields supplied by a specific error type into the map.
	if additionalInfo != nil {
		for k, v := range additionalInfo {
			errorAsMap[k] = v
		}
	}

	// Compute the current error map's YAML string value.
	errorAsYAML := getErrorInfoAsYAML(errorAsMap)

	// "Recursively" append the chain of causedBy errors to the message.
	if causedBy != nil {
		errorAsYAML = fmt.Sprintf("%sCaused by:\n%s", errorAsYAML, causedBy.GetDebugMessage())
	}

	return errorAsYAML
}

// getMapWithID converts a Problem type to a generic map and adds the computed ID
// to it. This is used for printing out error data as YAML, especially when we want
// to add addtional, unexported fields to the debug message.
func getMapWithID(problem Problem) map[string]interface{} {
	var errorAsMap map[string]interface{}
	jsonBytes, err := json.Marshal(problem)
	if err != nil {
		GetLogger().Debug("Failed to parse Problem as JSON data")
	}
	err = json.Unmarshal(jsonBytes, &errorAsMap)
	if err != nil {
		// TODO: rethink this message
		GetLogger().Debug("Failed to create map from Problem data")
	}

	// Add the ID as a field to the map - it is always relevant.
	errorAsMap["id"] = problem.GetID()

	return errorAsMap
}

/* TODO: things we might need to add to errors in general:
		- A flag to determine if the chain originated from HTTP or not
		- A "deep getter" for retrieving the underlying HTTP error from any level
*/

// EnrichHTTPError takes an error that should be an SDKError from the core,
// checks to see if it was caused by an HTTPError, and if so - populates the
// fields of the HTTP error with the given service/operation information.
func EnrichHTTPError(err error, operationID, system, version string) {
	// Expect an SDK to call this function, passing in an SDKError instance
	// that originated here in the core.
	sdkErr, ok := err.(*SDKError)
	if !ok {
		return
	}

	// If the error has no causedBy error, it didn't originate from an HTTP
	// error response, so there's nothing to do here.
	causedBy := sdkErr.GetCausedBy()
	if causedBy == nil {
		return
	}

	// If the error did originate from an HTTP error response, populate the
	// HTTPError instance with details from the SDK that weren't available
	// in the core at error creation time.
	if httpErr, ok := causedBy.(*HTTPError); ok {
		httpErr.OperationID = operationID
		httpErr.System = system
		httpErr.Version = version

		// TODO: think about how we might pull a discriminator from an API, if needed

		if httpErr.Response.Result != nil {
			// If the error response was a standard JSON body, the result will be a map
			// and we can do a decent job of guessing the code.

			// TODO: enable this once we know we can enumerate codes from an API.
			/*if resultMap, ok := httpErr.Response.Result.(map[string]interface{}); ok {
				httpErr.ErrorCode = getErrorCode(resultMap)
			}*/
		}
	}
}

func getHTTPFromAuthenticatorError(err error) (*HTTPError, bool) {
	if authErr, ok := err.(*AuthenticationError); ok {
		return authErr.ConvertToHTTPError()
	}

	return nil, false
}
