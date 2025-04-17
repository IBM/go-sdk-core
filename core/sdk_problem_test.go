//go:build all || fast || problem

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
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "sigs.k8s.io/yaml/goyaml.v2"
)

func TestSDKProblemEmbedsIBMProblem(t *testing.T) {
	sdkProb := &SDKProblem{}

	// Check that the methods defined by IBMProblem are supported here.
	// The implementations are tested elsewhere.
	assert.NotNil(t, sdkProb.Error)
	assert.NotNil(t, sdkProb.GetBaseSignature)
	assert.NotNil(t, sdkProb.GetCausedBy)
	assert.NotNil(t, sdkProb.Unwrap)
}

func TestSDKProblemGetConsoleMessage(t *testing.T) {
	sdkProb := getPopulatedSDKProblem()
	message := sdkProb.GetConsoleMessage()
	expected := `---
id: sdk-32d4ac5e
summary: Wrong!
severity: warning
function: mysdk.(*MySdkV1).GetResource
component:
  name: my-sdk
  version: 1.2.3
---
`
	assert.Equal(t, expected, message)
}

func TestSDKProblemGetDebugMessage(t *testing.T) {
	sdkProb := getPopulatedSDKProblem()
	message := sdkProb.GetDebugMessage()
	expected := `---
id: sdk-32d4ac5e
summary: Wrong!
severity: warning
function: mysdk.(*MySdkV1).GetResource
component:
  name: my-sdk
  version: 1.2.3
stack:
- function: my-sdk/mysdk.(*MySdkV1).GetResource
  file: /path/my-sdk-project/mysdkv1/my_sdk_v1.go
  line: 237
caused_by:
  id: mock-abc123
  data: some_data
---
`
	assert.Equal(t, expected, message)
}

func TestSDKProblemGetDebugMessageWithCoreProblem(t *testing.T) {
	coreProb := SDKErrorf(nil, "", "", getComponentInfo())
	sdkProb := SDKErrorf(coreProb, "Wrong!", "disc", NewProblemComponent("a", "b"))
	message := sdkProb.GetDebugMessage()
	expected := `---
id: sdk-1518356c
summary: Wrong!
severity: error
function: github.com/IBM/go-sdk-core/v5/core.TestSDKProblemGetDebugMessageWithCoreProblem
component:
  name: a
  version: b
stack: []
core_problem:
  id: sdk-fd790a1d
  function: core.TestSDKProblemGetDebugMessageWithCoreProblem
---
`
	assert.Equal(t, expected, message)
}

func TestSDKProblemGetID(t *testing.T) {
	sdkProb := getPopulatedSDKProblem()
	assert.Equal(t, "sdk-32d4ac5e", sdkProb.GetID())
}

func TestSDKProblemGetConsoleOrderedMaps(t *testing.T) {
	sdkProb := getPopulatedSDKProblem()
	orderedMaps := sdkProb.GetConsoleOrderedMaps()
	assert.NotNil(t, orderedMaps)

	maps := orderedMaps.GetMaps()
	assert.NotNil(t, maps)
	assert.Len(t, maps, 5)

	assert.Equal(t, "id", maps[0].Key)
	assert.Equal(t, "sdk-32d4ac5e", maps[0].Value)

	assert.Equal(t, "summary", maps[1].Key)
	assert.Equal(t, "Wrong!", maps[1].Value)

	assert.Equal(t, "severity", maps[2].Key)
	assert.Equal(t, WarningSeverity, maps[2].Value)

	assert.Equal(t, "function", maps[3].Key)
	assert.Equal(t, "mysdk.(*MySdkV1).GetResource", maps[3].Value)

	assert.Equal(t, "component", maps[4].Key)
	assert.Equal(t, "my-sdk", maps[4].Value.(*ProblemComponent).Name)
	assert.Equal(t, "1.2.3", maps[4].Value.(*ProblemComponent).Version)
}

func TestSDKProblemGetDebugOrderedMaps(t *testing.T) {
	sdkProb := getPopulatedSDKProblem()
	orderedMaps := sdkProb.GetDebugOrderedMaps()
	assert.NotNil(t, orderedMaps)

	maps := orderedMaps.GetMaps()
	assert.NotNil(t, maps)
	assert.Len(t, maps, 7)

	assert.Equal(t, "id", maps[0].Key)
	assert.Equal(t, "sdk-32d4ac5e", maps[0].Value)

	assert.Equal(t, "summary", maps[1].Key)
	assert.Equal(t, "Wrong!", maps[1].Value)

	assert.Equal(t, "severity", maps[2].Key)
	assert.Equal(t, WarningSeverity, maps[2].Value)

	assert.Equal(t, "function", maps[3].Key)
	assert.Equal(t, "mysdk.(*MySdkV1).GetResource", maps[3].Value)

	assert.Equal(t, "component", maps[4].Key)
	assert.Equal(t, "my-sdk", maps[4].Value.(*ProblemComponent).Name)
	assert.Equal(t, "1.2.3", maps[4].Value.(*ProblemComponent).Version)

	assert.Equal(t, "stack", maps[5].Key)
	assert.Equal(t, "my-sdk/mysdk.(*MySdkV1).GetResource", maps[5].Value.([]sdkStackFrame)[0].Function)
	assert.Equal(t, "/path/my-sdk-project/mysdkv1/my_sdk_v1.go", maps[5].Value.([]sdkStackFrame)[0].File)
	assert.Equal(t, 237, maps[5].Value.([]sdkStackFrame)[0].Line)

	assert.Equal(t, "caused_by", maps[6].Key)

	causedByMaps := maps[6].Value.([]yaml.MapItem)
	assert.Len(t, causedByMaps, 2)
	assert.Equal(t, "id", causedByMaps[0].Key)
	assert.Equal(t, "mock-abc123", causedByMaps[0].Value)

	assert.Equal(t, "data", causedByMaps[1].Key)
	assert.Equal(t, "some_data", causedByMaps[1].Value)
}

func TestSDKErrorf(t *testing.T) {
	causedBy := mockProblem{Data: "some_data"}
	summary := "Wrong!"
	discriminator := "some-issue"

	// The "name" value needs to match the component name of the Go SDK Core
	// project in order to test that the component name gets removed from the
	// function name when an error is created.
	component := NewProblemComponent("github.com/IBM/go-sdk-core/v5", "1.2.3")

	sdkProb := SDKErrorf(causedBy, summary, discriminator, component)
	assert.NotNil(t, sdkProb)
	assert.Equal(t, causedBy, sdkProb.causedBy)
	assert.Equal(t, summary, sdkProb.Summary)
	assert.Equal(t, discriminator, sdkProb.discriminator)
	assert.Equal(t, component, sdkProb.Component)
	assert.Equal(t, "core.TestSDKErrorf", sdkProb.Function)
	assert.Equal(t, ErrorSeverity, sdkProb.Severity)

	stack := sdkProb.stack
	assert.NotNil(t, stack)
	assert.Len(t, stack, 1)
	assert.Equal(t, "github.com/IBM/go-sdk-core/v5/core.TestSDKErrorf", stack[0].Function)
	assert.Contains(t, stack[0].File, "core/sdk_problem_test.go")
	// This might be too fragile. If it becomes an issue, we can remove it.
	assert.Equal(t, 181, stack[0].Line)
}

func TestSDKErrorfNoCausedBy(t *testing.T) {
	summary := "Wrong!"
	discriminator := "some-issue"

	// Testing behavior of a component name that doesn't actually match
	// the component, which ideally would never happen but still seems
	// good to capture. This will also captures the fact that the stack
	// will not be computed when using the wrong component name.
	component := NewProblemComponent("my-sdk", "1.2.3")

	sdkProb := SDKErrorf(nil, summary, discriminator, component)
	assert.NotNil(t, sdkProb)
	assert.Nil(t, sdkProb.causedBy)
	assert.Equal(t, summary, sdkProb.Summary)
	assert.Equal(t, discriminator, sdkProb.discriminator)
	assert.Equal(t, component, sdkProb.Component)
	assert.Equal(t, "github.com/IBM/go-sdk-core/v5/core.TestSDKErrorfNoCausedBy", sdkProb.Function)
	assert.Equal(t, ErrorSeverity, sdkProb.Severity)
	assert.Equal(t, []sdkStackFrame{}, sdkProb.stack)
}

func TestSDKErrorfNoSummary(t *testing.T) {
	message := "some_data"
	causedBy := mockProblem{Data: message}
	discriminator := "some-issue"

	component := NewProblemComponent("github.com/IBM/go-sdk-core/v5", "1.2.3")

	sdkProb := SDKErrorf(causedBy, "", discriminator, component)
	assert.NotNil(t, sdkProb)
	assert.Equal(t, causedBy, sdkProb.causedBy)
	assert.Equal(t, message, sdkProb.Summary)
	assert.Equal(t, discriminator, sdkProb.discriminator)
	assert.Equal(t, component, sdkProb.Component)
	assert.Equal(t, "core.TestSDKErrorfNoSummary", sdkProb.Function)
	assert.Equal(t, ErrorSeverity, sdkProb.Severity)

	stack := sdkProb.stack
	assert.NotNil(t, stack)
	assert.Len(t, stack, 1)
	assert.Equal(t, "github.com/IBM/go-sdk-core/v5/core.TestSDKErrorfNoSummary", stack[0].Function)
	assert.Contains(t, stack[0].File, "core/sdk_problem_test.go")
}

func TestSDKErrorfDoesntUseSDKCausedBy(t *testing.T) {
	sdkProb := getPopulatedSDKProblem()
	newSDKProb := SDKErrorf(sdkProb, "", "", NewProblemComponent("a", "b"))
	assert.Nil(t, newSDKProb.causedBy)
	assert.NotNil(t, newSDKProb.nativeCausedBy)
	assert.Equal(t, sdkProb, newSDKProb.nativeCausedBy)
	assert.Nil(t, newSDKProb.coreProblem)
}

func TestSDKErrorfStoreCoreProblem(t *testing.T) {
	coreProb := SDKErrorf(nil, "", "", getComponentInfo())
	sdkProb := SDKErrorf(coreProb, "", "", NewProblemComponent("a", "b"))
	assert.Nil(t, sdkProb.causedBy)
	assert.NotNil(t, sdkProb.nativeCausedBy)
	assert.Equal(t, coreProb, sdkProb.nativeCausedBy)
	assert.NotNil(t, sdkProb.coreProblem)
	assert.Equal(t, coreProb.GetID(), sdkProb.coreProblem.ID)
	assert.Equal(t, coreProb.Function, sdkProb.coreProblem.Function)
}

func TestSDKErrorfHTTPCausedByNotSetForCoreProblem(t *testing.T) {
	httpProb := getPopulatedHTTPProblem()
	coreProb := SDKErrorf(httpProb, "", "", getComponentInfo())
	assert.Nil(t, coreProb.causedBy)
	assert.Nil(t, coreProb.nativeCausedBy)
	assert.NotNil(t, coreProb.httpProblem)
	assert.Equal(t, httpProb, coreProb.httpProblem)
}

func TestSDKErrorfGetHTTPCausedByFromCoreProblem(t *testing.T) {
	httpProb := getPopulatedHTTPProblem()

	coreProb := SDKErrorf(httpProb, "", "", getComponentInfo())
	assert.Nil(t, coreProb.causedBy)
	assert.NotNil(t, coreProb.httpProblem)
	assert.Equal(t, httpProb, coreProb.httpProblem)

	sdkProb := SDKErrorf(coreProb, "", "", NewProblemComponent("a", "b"))
	assert.NotNil(t, sdkProb.causedBy)
	assert.Equal(t, httpProb, sdkProb.causedBy)

	assert.NotNil(t, sdkProb.nativeCausedBy)
	assert.Equal(t, coreProb, sdkProb.nativeCausedBy)
}

func TestRepurposeSDKProblem(t *testing.T) {
	sdkProb := getPopulatedSDKProblem()
	assert.Equal(t, "some-issue", sdkProb.discriminator)

	err := RepurposeSDKProblem(sdkProb, "new-disc")
	newSDKProb, ok := err.(*SDKProblem)
	assert.True(t, ok)
	assert.Equal(t, "new-disc", newSDKProb.discriminator)
	assert.Equal(t, "github.com/IBM/go-sdk-core/v5/core.TestRepurposeSDKProblem", newSDKProb.Function)
	assert.Equal(t, sdkProb.Severity, newSDKProb.Severity)
	assert.Equal(t, sdkProb.Summary, newSDKProb.Summary)
}

func TestRepurposeSDKProblemNilProblem(t *testing.T) {
	err := RepurposeSDKProblem(nil, "new-disc")
	assert.Nil(t, err)
}

func TestRepurposeSDKProblemNonSDKProblem(t *testing.T) {
	mockProb := mockProblem{}
	err := RepurposeSDKProblem(mockProb, "new-disc")
	assert.Equal(t, mockProb, err)
}

func TestSDKProblemIsWithProblem(t *testing.T) {
	firstProb := SDKErrorf(nil, "Some error", "disc", getComponentInfo())
	secondProb := SDKErrorf(nil, "Same error, different message", "disc", getComponentInfo())
	assert.NotEqual(t, firstProb, secondProb)
	assert.True(t, errors.Is(firstProb, secondProb))
}

func TestSDKProblemIsWithNative(t *testing.T) {
	firstProb := SDKErrorf(context.Canceled, "Some error", "disc", getComponentInfo())
	secondProb := SDKErrorf(firstProb, "Wrapping error", "disc", getComponentInfo())
	assert.True(t, errors.Is(secondProb, context.Canceled))
}

func TestSDKProblemIsWithCoreProblem(t *testing.T) {
	firstProb := SDKErrorf(nil, "msg", "disc", getComponentInfo())
	secondProb := SDKErrorf(firstProb, "", "other-disc", NewProblemComponent("a", "b"))
	assert.False(t, errors.Is(firstProb, secondProb))
	assert.True(t, errors.Is(secondProb, firstProb))
}

func TestNewSparseSDKProblem(t *testing.T) {
	sparse := newSparseSDKProblem(getPopulatedSDKProblem())
	assert.NotNil(t, sparse)
	assert.Equal(t, "sdk-32d4ac5e", sparse.ID)
	assert.Equal(t, "mysdk.(*MySdkV1).GetResource", sparse.Function)
}

func TestIsCoreProblem(t *testing.T) {
	assert.False(t, isCoreProblem(getPopulatedSDKProblem()))
	assert.True(t, isCoreProblem(SDKErrorf(nil, "", "", getComponentInfo())))
}

func getPopulatedSDKProblem() *SDKProblem {
	return &SDKProblem{
		IBMProblem: &IBMProblem{
			Summary:       "Wrong!",
			Component:     NewProblemComponent("my-sdk", "1.2.3"),
			Severity:      WarningSeverity,
			discriminator: "some-issue",
			causedBy: mockProblem{
				Data: "some_data",
			},
		},
		Function: "mysdk.(*MySdkV1).GetResource",
		stack: []sdkStackFrame{
			{
				Function: "my-sdk/mysdk.(*MySdkV1).GetResource",
				File:     "/path/my-sdk-project/mysdkv1/my_sdk_v1.go",
				Line:     237,
			},
		},
	}
}
