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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/go-yaml/yaml"
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

// getErrorInfoAsYAML formats the ordered error data as
// YAML for human/machine readable printing.
func getErrorInfoAsYAML(orderedMaps *OrderedMaps) string {
	asYaml, err := yaml.Marshal(orderedMaps.GetMaps())

	if err != nil {
		return fmt.Sprintf("Error serializing the error information: %s", err.Error())
	}
	return fmt.Sprintf("---\n%s---\n", asYaml)
}

func ComputeConsoleMessage(o OrderableProblem) string {
	return getErrorInfoAsYAML(o.GetConsoleOrderedMaps())
}

func ComputeDebugMessage(o OrderableProblem) string {
	return getErrorInfoAsYAML(o.GetDebugOrderedMaps())
}

/* TODO: things we might need to add to errors in general:
		- A flag to determine if the chain originated from HTTP or not
*/

// EnrichHTTPError takes an error that should be an SDKError and, if it originated
// as an HTTPError, populates the fields of the underlying HTTP error with the
// given service/operation information.
func EnrichUnderlyingHTTPError(err error, operationID string, getInfo infoProvider) {
	// Expect an SDK to call this function, passing in an SDKError instance
	// that originated here in the core.
	sdkErr, ok := err.(*SDKError)
	if !ok {
		return
	}

	// If the error originated from an HTTP error response, populate the
	// HTTPError instance with details from the SDK that weren't available
	// in the core at error creation time.
	httpErr := sdkErr.GetHTTPError()
	if httpErr != nil {
		enrichHTTPError(httpErr, operationID, getInfo)
	}
}

// enrichHTTPError takes an HTTPError instance alongside information about the request
// and adds the extra info to the instance. It also loosely deserializes the response
// in order to set additional information, like the error code.
func enrichHTTPError(httpErr *HTTPError, operationID string, getInfo infoProvider) {
	system, version := getInfo()

	httpErr.System = system
	httpErr.Version = version
	httpErr.OperationID = operationID

	// TODO: think about how we might pull a discriminator from an API, if needed

	if httpErr.Response.Result != nil {
		// If the error response was a standard JSON body, the result will be a map
		// and we can do a decent job of guessing the code.
		if resultMap, ok := httpErr.Response.Result.(map[string]interface{}); ok {
			httpErr.ErrorCode = getErrorCode(resultMap)
		}
	}
}
