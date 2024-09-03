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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewProblemComponent(t *testing.T) {
	name := "my-sdk"
	version := "1.2.3"
	component := NewProblemComponent(name, version)

	assert.NotNil(t, component)
	assert.Equal(t, name, component.Name)
	assert.Equal(t, version, component.Version)
}

func TestIBMProblemError(t *testing.T) {
	message := "Wrong!"
	ibmProblem := &IBMProblem{
		Summary: message,
	}

	assert.Equal(t, message, ibmProblem.Error())
}

func TestIBMProblemGetBaseSignature(t *testing.T) {
	ibmProblem := &IBMProblem{
		Summary:       "Wrong!",
		Component:     NewProblemComponent("my-sdk", "1.2.3"),
		Severity:      ErrorSeverity,
		discriminator: "some-issue",
		causedBy:      mockProblem{},
	}

	assert.Equal(t, "my-sdkerrorsome-issuemock-abc123", ibmProblem.GetBaseSignature())
}

func TestIBMProblemGetBaseSignatureNoCausedBy(t *testing.T) {
	ibmProblem := &IBMProblem{
		Summary:       "Wrong!",
		Component:     NewProblemComponent("my-sdk", "1.2.3"),
		Severity:      ErrorSeverity,
		discriminator: "some-issue",
	}

	assert.Equal(t, "my-sdkerrorsome-issue", ibmProblem.GetBaseSignature())
}

func TestIBMProblemGetBaseSignatureNoDiscriminator(t *testing.T) {
	ibmProblem := &IBMProblem{
		Summary:   "Wrong!",
		Component: NewProblemComponent("my-sdk", "1.2.3"),
		Severity:  ErrorSeverity,
		causedBy:  mockProblem{},
	}

	assert.Equal(t, "my-sdkerrormock-abc123", ibmProblem.GetBaseSignature())
}

func TestIBMProblemGetCausedBy(t *testing.T) {
	ibmProblem := &IBMProblem{
		Summary: "Wrong!",
	}

	assert.Nil(t, ibmProblem.GetCausedBy())

	data := "test"
	ibmProblem = &IBMProblem{
		causedBy: mockProblem{
			Data: data,
		},
	}

	cb := ibmProblem.GetCausedBy()
	assert.NotNil(t, cb)

	mock, ok := cb.(mockProblem)
	assert.True(t, ok)
	assert.Equal(t, data, mock.Data)
}

// Note: the "Unwrap" method isn't intended to be invoked
// directly, but to enable "errors.As" to populate errors
// with "caused by" problems. So, that's what we test.
func TestIBMProblemUnwrap(t *testing.T) {
	data := "test"

	err := &IBMProblem{
		Summary: data,
		causedBy: mockProblem{
			Data: data,
		},
	}

	assert.Equal(t, data, err.Error())

	var ibmProb *IBMProblem
	isIBMProb := errors.As(err, &ibmProb)
	assert.True(t, isIBMProb)
	assert.Equal(t, data, ibmProb.Summary)

	var mock mockProblem
	ismockProblem := errors.As(err, &mock)
	assert.True(t, ismockProblem)
	assert.Equal(t, data, mock.Data)
}

func TestIBMProblemf(t *testing.T) {
	data := "data"
	causedBy := mockProblem{Data: data}
	severity := WarningSeverity
	componentName := "my-sdk"
	componentVersion := "1.2.3"
	component := NewProblemComponent(componentName, componentVersion)
	summary := "Wrong!"
	discriminator := "some-issue"

	ibmProblem := ibmProblemf(causedBy, severity, component, summary, discriminator)
	assert.NotNil(t, ibmProblem)
	assert.Equal(t, causedBy, ibmProblem.causedBy)
	assert.Equal(t, severity, ibmProblem.Severity)
	assert.Equal(t, component, ibmProblem.Component)
	assert.Equal(t, summary, ibmProblem.Summary)
	assert.Equal(t, discriminator, ibmProblem.discriminator)
}

func TestIBMProblemfNoCausedBy(t *testing.T) {
	severity := ErrorSeverity
	componentName := "my-sdk"
	componentVersion := "1.2.3"
	component := NewProblemComponent(componentName, componentVersion)
	summary := "Wrong!"
	discriminator := "some-issue"

	ibmProblem := ibmProblemf(nil, severity, component, summary, discriminator)
	assert.NotNil(t, ibmProblem)
	assert.Nil(t, ibmProblem.causedBy)
	assert.Equal(t, severity, ibmProblem.Severity)
	assert.Equal(t, component, ibmProblem.Component)
	assert.Equal(t, summary, ibmProblem.Summary)
	assert.Equal(t, discriminator, ibmProblem.discriminator)
}

func TestIBMProblemfCausedByNotProblem(t *testing.T) {
	severity := WarningSeverity
	componentName := "my-sdk"
	componentVersion := "1.2.3"
	component := NewProblemComponent(componentName, componentVersion)
	summary := "Wrong!"
	discriminator := "some-issue"

	ibmProblem := ibmProblemf(errors.New("unused"), severity, component, summary, discriminator)
	assert.NotNil(t, ibmProblem)
	assert.Nil(t, ibmProblem.causedBy)
	assert.Equal(t, severity, ibmProblem.Severity)
	assert.Equal(t, component, ibmProblem.Component)
	assert.Equal(t, summary, ibmProblem.Summary)
	assert.Equal(t, discriminator, ibmProblem.discriminator)
}

func TestIBMProblemfNoSummary(t *testing.T) {
	data := "data"
	causedBy := mockProblem{Data: data}
	severity := WarningSeverity
	componentName := "my-sdk"
	componentVersion := "1.2.3"
	component := NewProblemComponent(componentName, componentVersion)
	discriminator := "some-issue"

	ibmProblem := ibmProblemf(causedBy, severity, component, "", discriminator)
	assert.NotNil(t, ibmProblem)
	assert.Equal(t, causedBy, ibmProblem.causedBy)
	assert.Equal(t, severity, ibmProblem.Severity)
	assert.Equal(t, component, ibmProblem.Component)
	assert.Equal(t, data, ibmProblem.Summary)
	assert.Equal(t, discriminator, ibmProblem.discriminator)
}

func TestIBMErrorf(t *testing.T) {
	data := "data"
	causedBy := mockProblem{Data: data}
	componentName := "my-sdk"
	componentVersion := "1.2.3"
	component := NewProblemComponent(componentName, componentVersion)
	summary := "Wrong!"
	discriminator := "some-issue"

	ibmProblem := IBMErrorf(causedBy, component, summary, discriminator)
	assert.NotNil(t, ibmProblem)
	assert.Equal(t, causedBy, ibmProblem.causedBy)
	assert.Equal(t, ErrorSeverity, ibmProblem.Severity)
	assert.Equal(t, component, ibmProblem.Component)
	assert.Equal(t, summary, ibmProblem.Summary)
	assert.Equal(t, discriminator, ibmProblem.discriminator)
}

func TestIBMErrorfNoCausedBy(t *testing.T) {
	componentName := "my-sdk"
	componentVersion := "1.2.3"
	component := NewProblemComponent(componentName, componentVersion)
	summary := "Wrong!"
	discriminator := "some-issue"

	ibmProblem := IBMErrorf(nil, component, summary, discriminator)
	assert.NotNil(t, ibmProblem)
	assert.Nil(t, ibmProblem.causedBy)
	assert.Equal(t, ErrorSeverity, ibmProblem.Severity)
	assert.Equal(t, component, ibmProblem.Component)
	assert.Equal(t, summary, ibmProblem.Summary)
	assert.Equal(t, discriminator, ibmProblem.discriminator)
}

func TestIBMErrorfCausedByNotProblem(t *testing.T) {
	componentName := "my-sdk"
	componentVersion := "1.2.3"
	component := NewProblemComponent(componentName, componentVersion)
	summary := "Wrong!"
	discriminator := "some-issue"

	ibmProblem := IBMErrorf(errors.New("unused"), component, summary, discriminator)
	assert.NotNil(t, ibmProblem)
	assert.Nil(t, ibmProblem.causedBy)
	assert.Equal(t, ErrorSeverity, ibmProblem.Severity)
	assert.Equal(t, component, ibmProblem.Component)
	assert.Equal(t, summary, ibmProblem.Summary)
	assert.Equal(t, discriminator, ibmProblem.discriminator)
}

func TestIBMErrorfNoSummary(t *testing.T) {
	data := "data"
	causedBy := mockProblem{Data: data}
	componentName := "my-sdk"
	componentVersion := "1.2.3"
	component := NewProblemComponent(componentName, componentVersion)
	discriminator := "some-issue"

	ibmProblem := IBMErrorf(causedBy, component, "", discriminator)
	assert.NotNil(t, ibmProblem)
	assert.Equal(t, causedBy, ibmProblem.causedBy)
	assert.Equal(t, ErrorSeverity, ibmProblem.Severity)
	assert.Equal(t, component, ibmProblem.Component)
	assert.Equal(t, data, ibmProblem.Summary)
	assert.Equal(t, discriminator, ibmProblem.discriminator)
}

func TestProblemSeverityConstants(t *testing.T) {
	// The values should be equal but the types should not be.
	assert.NotEqual(t, "error", ErrorSeverity)
	assert.EqualValues(t, "error", ErrorSeverity)

	assert.NotEqual(t, "warning", WarningSeverity)
	assert.EqualValues(t, "warning", WarningSeverity)
}

type mockProblem struct {
	Data string
}

func (m mockProblem) GetConsoleMessage() string {
	return ""
}
func (m mockProblem) GetDebugMessage() string {
	return ""
}
func (m mockProblem) GetID() string {
	return "mock-abc123"
}
func (m mockProblem) Error() string {
	return m.Data
}
func (m mockProblem) GetConsoleOrderedMaps() *OrderedMaps {
	orderedMaps := NewOrderedMaps()
	orderedMaps.Add("id", m.GetID())
	return orderedMaps
}
func (m mockProblem) GetDebugOrderedMaps() *OrderedMaps {
	orderedMaps := m.GetConsoleOrderedMaps()
	orderedMaps.Add("data", m.Data)
	return orderedMaps
}
