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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	consoleKey      = "console"
	consoleValue    = "my-console-message"
	debugKey        = "debug"
	debugValue      = "my-debug-message"
	messageTemplate = "---\n%s: %s\n---\n"
)

func TestComputeConsoleMessage(t *testing.T) {
	message := ComputeConsoleMessage(&MockOrderableProblem{})
	expected := fmt.Sprintf(messageTemplate, consoleKey, consoleValue)
	assert.Equal(t, expected, message)
}

func TestComputeDebugMessage(t *testing.T) {
	message := ComputeDebugMessage(&MockOrderableProblem{})
	expected := fmt.Sprintf(messageTemplate, debugKey, debugValue)
	assert.Equal(t, expected, message)
}

func TestCreateIDHash(t *testing.T) {
	hash := CreateIDHash("my-prefix", "component", "discriminator")
	assert.Equal(t, "my-prefix-9507ef8a", hash)

	hash = CreateIDHash("other-prefix", "component", "discriminator", "function", "caused_by_id")
	assert.Equal(t, "other-prefix-f24346b0", hash)
}

func TestGetProblemInfoAsYAML(t *testing.T) {
	mockOP := &MockOrderableProblem{}
	message := getProblemInfoAsYAML(mockOP.GetConsoleOrderedMaps())
	expected := fmt.Sprintf(messageTemplate, consoleKey, consoleValue)
	assert.Equal(t, expected, message)
}

func TestGetComponentInfo(t *testing.T) {
	component := getComponentInfo()
	assert.NotNil(t, component)
	assert.Equal(t, MODULE_NAME, component.Name)
	assert.Equal(t, __VERSION__, component.Version)
}

type MockOrderableProblem struct{}

func (m *MockOrderableProblem) GetConsoleOrderedMaps() *OrderedMaps {
	orderedMaps := NewOrderedMaps()
	orderedMaps.Add(consoleKey, consoleValue)
	return orderedMaps
}

func (m *MockOrderableProblem) GetDebugOrderedMaps() *OrderedMaps {
	orderedMaps := NewOrderedMaps()
	orderedMaps.Add(debugKey, debugValue)
	return orderedMaps
}
