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

// getErrorInfoAsYAML formats the error data as YAML
// for human/machine readable printing.
func getErrorInfoAsYAML(obj interface{}) string {
	yamlifiedStruct, err := yaml.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("Error serializing the error information: %s", err.Error())
	}
	return fmt.Sprintf("---\n%s---\n", yamlifiedStruct)
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
	errorAsMap["ID"] = problem.GetID()

	return errorAsMap
}

// getMapWithCausedBy converts a Problem to a map and, if relevant, adds the
// "causedBy" error, along with its ID field (if it is also a Problem).
func getMapWithCausedBy(problem Problem, causedBy error) map[string]interface{} {
	errorAsMap := getMapWithID(problem)

	if causedBy != nil {
		// Set causedBy in the map. If it is a native golang error,
		// this will stay. Otherwise, it will be overwritten.

		// TODO: consider mapifying the causedBy error in order to inlcude the .Error()
		// message in the case of native errors.
		errorAsMap["CausedBy"] = causedBy

		// If causedBy is a Problem type, we'll update the field
		// so that it includes its hash ID.
		if causedByProblem, ok := causedBy.(Problem); ok {
			errorAsMap["CausedBy"] = getMapWithID(causedByProblem)
		}
	}

	return errorAsMap
}
