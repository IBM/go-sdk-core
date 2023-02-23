//go:build all || fast
// +build all fast

package core

/**
 * (C) Copyright IBM Corp. 2020.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalPrimitiveString(t *testing.T) {
	type MyModel struct {
		Prop              *string
		PropSlice         []string
		PropSliceSlice    [][]string
		PropMap           map[string]string
		PropSliceMap      map[string][]string
		PropMapSlice      []map[string]string
		PropSliceMapSlice []map[string][]string
	}

	jsonTemplate := `{
		"prop": "%s1",
		"prop_slice": ["%s1", "%s2"],
		"prop_slice_slice": [["%s1"], ["%s2", "%s3"], ["%s4"]],
		"prop_map": { "key1": "%s1", "key2": "%s2" },
		"prop_slice_map": { "key1": ["%s1", "%s2"], "key2": ["%s3", "%s4"] },
		"prop_map_slice": [{"key1": "%s1"}, {"key2": "%s2"}],
		"prop_slice_map_slice": [{"key1": ["%s1"]}, {"key2": ["%s2", "%s3", "%s4"]} ],
		
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"null_prop": null,
		"slice_with_null": [ null, "%s1", null ],
		"slice_slice_with_null": [ null, [null], ["%s1"], ["%s1", null] ]
	}`

	s1 := "value1"
	s2 := "value2"
	s3 := "value3"
	s4 := "value4"
	var zeroValue string

	jsonString := strings.ReplaceAll(jsonTemplate, "%s1", s1)
	jsonString = strings.ReplaceAll(jsonString, "%s2", s2)
	jsonString = strings.ReplaceAll(jsonString, "%s3", s3)
	jsonString = strings.ReplaceAll(jsonString, "%s4", s4)

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, s1, *(model.Prop))

	var aString string
	err = UnmarshalPrimitive(rawMap, "prop", &aString)
	assert.Nil(t, err)
	assert.Equal(t, s1, aString)

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, s1, model.PropSlice[0])
	assert.Equal(t, s2, model.PropSlice[1])

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, s1, model.PropSliceSlice[0][0])
	assert.Equal(t, s2, model.PropSliceSlice[1][0])
	assert.Equal(t, s3, model.PropSliceSlice[1][1])
	assert.Equal(t, s4, model.PropSliceSlice[2][0])

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, s1, model.PropMap["key1"])
	assert.Equal(t, s2, model.PropMap["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, s1, model.PropSliceMap["key1"][0])
	assert.Equal(t, s2, model.PropSliceMap["key1"][1])
	assert.Equal(t, s3, model.PropSliceMap["key2"][0])
	assert.Equal(t, s4, model.PropSliceMap["key2"][1])

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, s1, model.PropMapSlice[0]["key1"])
	assert.Equal(t, s2, model.PropMapSlice[1]["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, s1, model.PropSliceMapSlice[0]["key1"][0])
	assert.Equal(t, s2, model.PropSliceMapSlice[1]["key2"][0])
	assert.Equal(t, s3, model.PropSliceMapSlice[1]["key2"][1])
	assert.Equal(t, s4, model.PropSliceMapSlice[1]["key2"][2])

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue, model.PropSlice[0])
	assert.Equal(t, s1, model.PropSlice[1])
	assert.Equal(t, zeroValue, model.PropSlice[2])

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[1][0])
	assert.Equal(t, s1, model.PropSliceSlice[2][0])
	assert.Equal(t, s1, model.PropSliceSlice[3][0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[3][1])

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "", &model.Prop)
	assert.NotNil(t, err)
	assert.Equal(t, "the 'propertyName' parameter is required", err.Error())
	t.Logf("Expected error: %s\n", err.Error())
}

func TestUnmarshalPrimitiveBool(t *testing.T) {
	type MyModel struct {
		Prop              *bool
		PropSlice         []bool
		PropSliceSlice    [][]bool
		PropMap           map[string]bool
		PropSliceMap      map[string][]bool
		PropMapSlice      []map[string]bool
		PropSliceMapSlice []map[string][]bool
	}

	jsonTemplate := `{
		"prop": %b1,
		"prop_slice": [%b1, %b2],
		"prop_slice_slice": [[%b1], [%b2, %b1], [%b2]],
		"prop_map": { "key1": %b1, "key2": %b2 },
		"prop_slice_map": { "key1": [%b2, %b1], "key2": [%b1, %b2] },
		"prop_map_slice": [{"key1": %b1}, {"key2": %b2}],
		"prop_slice_map_slice": [{"key1": [%b1]}, {"key2": [%b2, %b2, %b1]} ],
		
		"bad_type":  "string",
		"not_a_slice": 38,
		"bad_slice_type": [38, 26],
		"null_prop": null,
		"slice_with_null": [ null, %b1, null ],
		"slice_slice_with_null": [ null, [null], [%b1], [%b1, null] ]
	}`

	b1 := true
	b2 := false
	var zeroValue bool

	jsonString := strings.ReplaceAll(jsonTemplate, "%b1", "true")
	jsonString = strings.ReplaceAll(jsonString, "%b2", "false")

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, b1, *(model.Prop))

	var aBool bool
	err = UnmarshalPrimitive(rawMap, "prop", &aBool)
	assert.Nil(t, err)
	assert.Equal(t, b1, aBool)

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, b1, model.PropSlice[0])
	assert.Equal(t, b2, model.PropSlice[1])

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, b1, model.PropSliceSlice[0][0])
	assert.Equal(t, b2, model.PropSliceSlice[1][0])
	assert.Equal(t, b1, model.PropSliceSlice[1][1])
	assert.Equal(t, b2, model.PropSliceSlice[2][0])

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, b1, model.PropMap["key1"])
	assert.Equal(t, b2, model.PropMap["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, b2, model.PropSliceMap["key1"][0])
	assert.Equal(t, b1, model.PropSliceMap["key1"][1])
	assert.Equal(t, b1, model.PropSliceMap["key2"][0])
	assert.Equal(t, b2, model.PropSliceMap["key2"][1])

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, b1, model.PropMapSlice[0]["key1"])
	assert.Equal(t, b2, model.PropMapSlice[1]["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, b1, model.PropSliceMapSlice[0]["key1"][0])
	assert.Equal(t, b2, model.PropSliceMapSlice[1]["key2"][0])
	assert.Equal(t, b2, model.PropSliceMapSlice[1]["key2"][1])
	assert.Equal(t, b1, model.PropSliceMapSlice[1]["key2"][2])

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue, model.PropSlice[0])
	assert.Equal(t, b1, model.PropSlice[1])
	assert.Equal(t, zeroValue, model.PropSlice[2])

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[1][0])
	assert.Equal(t, b1, model.PropSliceSlice[2][0])
	assert.Equal(t, b1, model.PropSliceSlice[3][0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[3][1])

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())
}

func TestUnmarshalPrimitiveByteArray(t *testing.T) {
	type MyModel struct {
		Prop              *[]byte
		PropSlice         [][]byte
		PropSliceSlice    [][][]byte
		PropMap           map[string][]byte
		PropSliceMap      map[string][][]byte
		PropMapSlice      []map[string][]byte
		PropSliceMapSlice []map[string][][]byte
	}

	s1 := "You're gonna need a bigger boat."
	s2 := "I'm gonna make him an offer he can't refuse."
	encodedString1 := base64.StdEncoding.EncodeToString([]byte(s1))
	encodedString2 := base64.StdEncoding.EncodeToString([]byte(s2))
	assert.NotNil(t, encodedString1)
	assert.NotNil(t, encodedString2)
	var zeroValue []byte

	jsonStringTemplate := `{
		"prop": "%s1",
		"prop_slice": ["%s1", "%s2"],
		"prop_slice_slice": [["%s1"], ["%s2", "%s1"], ["%s2"]],
		"prop_map": { "key1": "%s2", "key2": "%s1" },
		"prop_slice_map": { "key1": ["%s1", "%s2"], "key2": ["%s2", "%s1"] },
		"prop_map_slice": [{"key1": "%s2"}, {"key2": "%s1"}],
		"prop_slice_map_slice": [{"key1": ["%s1"]}, {"key2": ["%s2", "%s2", "%s1"]} ],
		
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"null_prop": null,
		"slice_with_null": [ null, "%s1", null ],
		"slice_slice_with_null": [ null, [null], ["%s1"], ["%s1", null] ]
	}`

	jsonString := strings.ReplaceAll(jsonStringTemplate, "%s1", encodedString1)
	jsonString = strings.ReplaceAll(jsonString, "%s2", encodedString2)

	// t.Logf("json string: %s\n", jsonString)

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, s1, string(*(model.Prop)))

	var aBA []byte
	err = UnmarshalPrimitive(rawMap, "prop", &aBA)
	assert.Nil(t, err)
	assert.Equal(t, s1, string(aBA))

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, s1, string(model.PropSlice[0]))
	assert.Equal(t, s2, string(model.PropSlice[1]))

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, s1, string(model.PropSliceSlice[0][0]))
	assert.Equal(t, s2, string(model.PropSliceSlice[1][0]))
	assert.Equal(t, s1, string(model.PropSliceSlice[1][1]))
	assert.Equal(t, s2, string(model.PropSliceSlice[2][0]))

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, s2, string(model.PropMap["key1"]))
	assert.Equal(t, s1, string(model.PropMap["key2"]))

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, s1, string(model.PropSliceMap["key1"][0]))
	assert.Equal(t, s2, string(model.PropSliceMap["key1"][1]))
	assert.Equal(t, s2, string(model.PropSliceMap["key2"][0]))
	assert.Equal(t, s1, string(model.PropSliceMap["key2"][1]))

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, s2, string(model.PropMapSlice[0]["key1"]))
	assert.Equal(t, s1, string(model.PropMapSlice[1]["key2"]))

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, s1, string(model.PropSliceMapSlice[0]["key1"][0]))
	assert.Equal(t, s2, string(model.PropSliceMapSlice[1]["key2"][0]))
	assert.Equal(t, s2, string(model.PropSliceMapSlice[1]["key2"][1]))
	assert.Equal(t, s1, string(model.PropSliceMapSlice[1]["key2"][2]))

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue, model.PropSlice[0])
	assert.Equal(t, s1, string(model.PropSlice[1]))
	assert.Equal(t, zeroValue, model.PropSlice[2])

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[1][0])
	assert.Equal(t, s1, string(model.PropSliceSlice[2][0]))
	assert.Equal(t, s1, string(model.PropSliceSlice[3][0]))
	assert.Equal(t, zeroValue, model.PropSliceSlice[3][1])

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())
}

func TestUnmarshalPrimitiveInt64(t *testing.T) {
	type MyModel struct {
		Prop              *int64
		PropSlice         []int64
		PropSliceSlice    [][]int64
		PropMap           map[string]int64
		PropSliceMap      map[string][]int64
		PropMapSlice      []map[string]int64
		PropSliceMapSlice []map[string][]int64
	}

	jsonTemplate := `{
		"prop": %n1,
		"prop_slice": [%n1, %n2],
		"prop_slice_slice": [[%n1], [%n2, %n3], [%n4]],
		"prop_map": { "key1": %n1, "key2": %n2 },
		"prop_slice_map": { "key1": [%n1, %n2], "key2": [%n3, %n4] },
		"prop_map_slice": [{"key1": %n1}, {"key2": %n2}],
		"prop_slice_map_slice": [{"key1": [%n1]}, {"key2": [%n2, %n3, %n4]} ],
		
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [true, false],
		"null_prop": null,
		"slice_with_null": [ null, %n1, null ],
		"slice_slice_with_null": [ null, [null], [%n1], [%n1, null] ]
	}`

	n1 := int64(44)
	n2 := int64(74)
	n3 := int64(27)
	n4 := int64(50)
	var zeroValue int64

	jsonString := strings.ReplaceAll(jsonTemplate, "%n1", fmt.Sprintf("%d", n1))
	jsonString = strings.ReplaceAll(jsonString, "%n2", fmt.Sprintf("%d", n2))
	jsonString = strings.ReplaceAll(jsonString, "%n3", fmt.Sprintf("%d", n3))
	jsonString = strings.ReplaceAll(jsonString, "%n4", fmt.Sprintf("%d", n4))

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, n1, *(model.Prop))

	var anInt int64
	err = UnmarshalPrimitive(rawMap, "prop", &anInt)
	assert.Nil(t, err)
	assert.Equal(t, n1, anInt)

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, n1, model.PropSlice[0])
	assert.Equal(t, n2, model.PropSlice[1])

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, n1, model.PropSliceSlice[0][0])
	assert.Equal(t, n2, model.PropSliceSlice[1][0])
	assert.Equal(t, n3, model.PropSliceSlice[1][1])
	assert.Equal(t, n4, model.PropSliceSlice[2][0])

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, n1, model.PropMap["key1"])
	assert.Equal(t, n2, model.PropMap["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, n1, model.PropSliceMap["key1"][0])
	assert.Equal(t, n2, model.PropSliceMap["key1"][1])
	assert.Equal(t, n3, model.PropSliceMap["key2"][0])
	assert.Equal(t, n4, model.PropSliceMap["key2"][1])

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, n1, model.PropMapSlice[0]["key1"])
	assert.Equal(t, n2, model.PropMapSlice[1]["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, n1, model.PropSliceMapSlice[0]["key1"][0])
	assert.Equal(t, n2, model.PropSliceMapSlice[1]["key2"][0])
	assert.Equal(t, n3, model.PropSliceMapSlice[1]["key2"][1])
	assert.Equal(t, n4, model.PropSliceMapSlice[1]["key2"][2])

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue, model.PropSlice[0])
	assert.Equal(t, n1, model.PropSlice[1])
	assert.Equal(t, zeroValue, model.PropSlice[2])

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[1][0])
	assert.Equal(t, n1, model.PropSliceSlice[2][0])
	assert.Equal(t, n1, model.PropSliceSlice[3][0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[3][1])

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())
}

func TestUnmarshalPrimitiveFloat32(t *testing.T) {
	type MyModel struct {
		Prop              *float32
		PropSlice         []float32
		PropSliceSlice    [][]float32
		PropMap           map[string]float32
		PropSliceMap      map[string][]float32
		PropMapSlice      []map[string]float32
		PropSliceMapSlice []map[string][]float32
	}

	jsonTemplate := `{
		"prop": %n1,
		"prop_slice": [%n1, %n2],
		"prop_slice_slice": [[%n1], [%n2, %n3], [%n4]],
		"prop_map": { "key1": %n1, "key2": %n2 },
		"prop_slice_map": { "key1": [%n1, %n2], "key2": [%n3, %n4] },
		"prop_map_slice": [{"key1": %n1}, {"key2": %n2}],
		"prop_slice_map_slice": [{"key1": [%n1]}, {"key2": [%n2, %n3, %n4]} ],
		
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [true, false],
		"null_prop": null,
		"slice_with_null": [ null, %n1, null ],
		"slice_slice_with_null": [ null, [null], [%n1], [%n1, null] ]
	}`

	n1 := float32(44.5)
	n2 := float32(74.8)
	n3 := float32(27.1)
	n4 := float32(50.9)
	var zeroValue float32

	jsonString := strings.ReplaceAll(jsonTemplate, "%n1", fmt.Sprintf("%f", n1))
	jsonString = strings.ReplaceAll(jsonString, "%n2", fmt.Sprintf("%f", n2))
	jsonString = strings.ReplaceAll(jsonString, "%n3", fmt.Sprintf("%f", n3))
	jsonString = strings.ReplaceAll(jsonString, "%n4", fmt.Sprintf("%f", n4))

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, n1, *(model.Prop))

	var aFloat float32
	err = UnmarshalPrimitive(rawMap, "prop", &aFloat)
	assert.Nil(t, err)
	assert.Equal(t, n1, aFloat)

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, n1, model.PropSlice[0])
	assert.Equal(t, n2, model.PropSlice[1])

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, n1, model.PropSliceSlice[0][0])
	assert.Equal(t, n2, model.PropSliceSlice[1][0])
	assert.Equal(t, n3, model.PropSliceSlice[1][1])
	assert.Equal(t, n4, model.PropSliceSlice[2][0])

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, n1, model.PropMap["key1"])
	assert.Equal(t, n2, model.PropMap["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, n1, model.PropSliceMap["key1"][0])
	assert.Equal(t, n2, model.PropSliceMap["key1"][1])
	assert.Equal(t, n3, model.PropSliceMap["key2"][0])
	assert.Equal(t, n4, model.PropSliceMap["key2"][1])

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, n1, model.PropMapSlice[0]["key1"])
	assert.Equal(t, n2, model.PropMapSlice[1]["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, n1, model.PropSliceMapSlice[0]["key1"][0])
	assert.Equal(t, n2, model.PropSliceMapSlice[1]["key2"][0])
	assert.Equal(t, n3, model.PropSliceMapSlice[1]["key2"][1])
	assert.Equal(t, n4, model.PropSliceMapSlice[1]["key2"][2])

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue, model.PropSlice[0])
	assert.Equal(t, n1, model.PropSlice[1])
	assert.Equal(t, zeroValue, model.PropSlice[2])

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[1][0])
	assert.Equal(t, n1, model.PropSliceSlice[2][0])
	assert.Equal(t, n1, model.PropSliceSlice[3][0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[3][1])

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())
}

func TestUnmarshalPrimitiveFloat64(t *testing.T) {
	type MyModel struct {
		Prop              *float64
		PropSlice         []float64
		PropSliceSlice    [][]float64
		PropMap           map[string]float64
		PropSliceMap      map[string][]float64
		PropMapSlice      []map[string]float64
		PropSliceMapSlice []map[string][]float64
	}

	jsonTemplate := `{
		"prop": %n1,
		"prop_slice": [%n1, %n2],
		"prop_slice_slice": [[%n1], [%n2, %n3], [%n4]],
		"prop_map": { "key1": %n1, "key2": %n2 },
		"prop_slice_map": { "key1": [%n1, %n2], "key2": [%n3, %n4] },
		"prop_map_slice": [{"key1": %n1}, {"key2": %n2}],
		"prop_slice_map_slice": [{"key1": [%n1]}, {"key2": [%n2, %n3, %n4]} ],
		
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [true, false],
		"null_prop": null,
		"slice_with_null": [ null, %n1, null ],
		"slice_slice_with_null": [ null, [null], [%n1], [%n1, null] ]
	}`

	n1 := float64(44.5)
	n2 := float64(74.8)
	n3 := float64(27.1)
	n4 := float64(50.9)
	var zeroValue float64

	jsonString := strings.ReplaceAll(jsonTemplate, "%n1", fmt.Sprintf("%f", n1))
	jsonString = strings.ReplaceAll(jsonString, "%n2", fmt.Sprintf("%f", n2))
	jsonString = strings.ReplaceAll(jsonString, "%n3", fmt.Sprintf("%f", n3))
	jsonString = strings.ReplaceAll(jsonString, "%n4", fmt.Sprintf("%f", n4))

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, n1, *(model.Prop))

	var aFloat float64
	err = UnmarshalPrimitive(rawMap, "prop", &aFloat)
	assert.Nil(t, err)
	assert.Equal(t, n1, aFloat)

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, n1, model.PropSlice[0])
	assert.Equal(t, n2, model.PropSlice[1])

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, n1, model.PropSliceSlice[0][0])
	assert.Equal(t, n2, model.PropSliceSlice[1][0])
	assert.Equal(t, n3, model.PropSliceSlice[1][1])
	assert.Equal(t, n4, model.PropSliceSlice[2][0])

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, n1, model.PropMap["key1"])
	assert.Equal(t, n2, model.PropMap["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, n1, model.PropSliceMap["key1"][0])
	assert.Equal(t, n2, model.PropSliceMap["key1"][1])
	assert.Equal(t, n3, model.PropSliceMap["key2"][0])
	assert.Equal(t, n4, model.PropSliceMap["key2"][1])

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, n1, model.PropMapSlice[0]["key1"])
	assert.Equal(t, n2, model.PropMapSlice[1]["key2"])

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, n1, model.PropSliceMapSlice[0]["key1"][0])
	assert.Equal(t, n2, model.PropSliceMapSlice[1]["key2"][0])
	assert.Equal(t, n3, model.PropSliceMapSlice[1]["key2"][1])
	assert.Equal(t, n4, model.PropSliceMapSlice[1]["key2"][2])

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue, model.PropSlice[0])
	assert.Equal(t, n1, model.PropSlice[1])
	assert.Equal(t, zeroValue, model.PropSlice[2])

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[1][0])
	assert.Equal(t, n1, model.PropSliceSlice[2][0])
	assert.Equal(t, n1, model.PropSliceSlice[3][0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[3][1])

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())
}

func TestUnmarshalPrimitiveDate(t *testing.T) {
	type MyModel struct {
		Prop              *strfmt.Date
		PropSlice         []strfmt.Date
		PropSliceSlice    [][]strfmt.Date
		PropMap           map[string]strfmt.Date
		PropSliceMap      map[string][]strfmt.Date
		PropMapSlice      []map[string]strfmt.Date
		PropSliceMapSlice []map[string][]strfmt.Date
	}

	jsonTemplate := `{
		"prop": "%d1",
		"prop_slice": ["%d1", "%d2"],
		"prop_slice_slice": [["%d1"], ["%d2", "%d3"], ["%d4" ]],
		"prop_map": { "key1": "%d1", "key2": "%d2" },
		"prop_slice_map": { "key1": ["%d1", "%d2"], "key2": ["%d3", "%d4"] },
		"prop_map_slice": [{"key1": "%d1"}, {"key2": "%d2"}],
		"prop_slice_map_slice": [{"key1": ["%d1"]}, {"key2": ["%d2", "%d3", "%d4"]} ],
		
		"bad_type":  true,
		"bad_date1": "",
		"bad_date2": "10-27-2004",
		"bad_date3": "she/he was a psycho",
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"null_prop": null,
		"slice_with_null": [ null, "%d1", null ],
		"slice_slice_with_null": [ null, [null], ["%d1"], ["%d1", null] ]
	}`

	d1 := "2004-10-27"
	d2 := "2007-10-28"
	d3 := "2013-10-30"
	d4 := "2018-10-28"
	var zeroValue strfmt.Date

	jsonString := strings.ReplaceAll(jsonTemplate, "%d1", d1)
	jsonString = strings.ReplaceAll(jsonString, "%d2", d2)
	jsonString = strings.ReplaceAll(jsonString, "%d3", d3)
	jsonString = strings.ReplaceAll(jsonString, "%d4", d4)

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, d1, model.Prop.String())

	var aDate strfmt.Date
	err = UnmarshalPrimitive(rawMap, "prop", &aDate)
	assert.Nil(t, err)
	assert.Equal(t, d1, aDate.String())

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, d1, model.PropSlice[0].String())
	assert.Equal(t, d2, model.PropSlice[1].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, d1, model.PropSliceSlice[0][0].String())
	assert.Equal(t, d2, model.PropSliceSlice[1][0].String())
	assert.Equal(t, d3, model.PropSliceSlice[1][1].String())
	assert.Equal(t, d4, model.PropSliceSlice[2][0].String())

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, d1, model.PropMap["key1"].String())
	assert.Equal(t, d2, model.PropMap["key2"].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, d1, model.PropSliceMap["key1"][0].String())
	assert.Equal(t, d2, model.PropSliceMap["key1"][1].String())
	assert.Equal(t, d3, model.PropSliceMap["key2"][0].String())
	assert.Equal(t, d4, model.PropSliceMap["key2"][1].String())

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, d1, model.PropMapSlice[0]["key1"].String())
	assert.Equal(t, d2, model.PropMapSlice[1]["key2"].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, d1, model.PropSliceMapSlice[0]["key1"][0].String())
	assert.Equal(t, d2, model.PropSliceMapSlice[1]["key2"][0].String())
	assert.Equal(t, d3, model.PropSliceMapSlice[1]["key2"][1].String())
	assert.Equal(t, d4, model.PropSliceMapSlice[1]["key2"][2].String())

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue.String(), model.PropSlice[0].String())
	assert.Equal(t, d1, model.PropSlice[1].String())
	assert.Equal(t, zeroValue.String(), model.PropSlice[2].String())

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue.String(), model.PropSliceSlice[1][0].String())
	assert.Equal(t, d1, model.PropSliceSlice[2][0].String())
	assert.Equal(t, d1, model.PropSliceSlice[3][0].String())
	assert.Equal(t, zeroValue.String(), model.PropSliceSlice[3][1].String())

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_date1", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_date1'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_date2", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_date2'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_date3", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_date3'"))
	t.Logf("Expected error: %s\n", err.Error())
}

func TestUnmarshalPrimitiveDateTime(t *testing.T) {
	type MyModel struct {
		Prop              *strfmt.DateTime
		PropSlice         []strfmt.DateTime
		PropSliceSlice    [][]strfmt.DateTime
		PropMap           map[string]strfmt.DateTime
		PropSliceMap      map[string][]strfmt.DateTime
		PropMapSlice      []map[string]strfmt.DateTime
		PropSliceMapSlice []map[string][]strfmt.DateTime
	}

	jsonTemplate := `{
		"prop": "%d1",
		"prop_slice": ["%d1", "%d2"],
		"prop_slice_slice": [["%d1"], ["%d2", "%d3"], ["%d4" ]],
		"prop_map": { "key1": "%d1", "key2": "%d2" },
		"prop_slice_map": { "key1": ["%d1", "%d2"], "key2": ["%d3", "%d4"] },
		"prop_map_slice": [{"key1": "%d1"}, {"key2": "%d2"}],
		"prop_slice_map_slice": [{"key1": ["%d1"]}, {"key2": ["%d2", "%d3", "%d4"]} ],
		
		"bad_type":  true,
		"bad_date1": "",
		"bad_date2": "10-27-2004T00:00:00Z",
		"bad_date3": "1970-01-01 18:30:00Z",
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"null_prop": null,
		"slice_with_null": [ null, "%d1", null ],
		"slice_slice_with_null": [ null, [null], ["%d1"], ["%d1", null] ]
	}`

	d1 := "1969-07-20T20:17:00"
	d2 := "1963-11-22T18:30:00Z"
	d3 := "2001-09-11T13:46:00.333Z"
	d4 := "2011-05-02T20:00:00.011Z"
	var zeroValue strfmt.DateTime

	jsonString := strings.ReplaceAll(jsonTemplate, "%d1", d1)
	jsonString = strings.ReplaceAll(jsonString, "%d2", d2)
	jsonString = strings.ReplaceAll(jsonString, "%d3", d3)
	jsonString = strings.ReplaceAll(jsonString, "%d4", d4)

	// Expected values need to include ms
	d1 = "1969-07-20T20:17:00.000Z"
	d2 = "1963-11-22T18:30:00.000Z"

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, d1, model.Prop.String())

	var aDateTime strfmt.DateTime
	err = UnmarshalPrimitive(rawMap, "prop", &aDateTime)
	assert.Nil(t, err)
	assert.Equal(t, d1, aDateTime.String())

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, d1, model.PropSlice[0].String())
	assert.Equal(t, d2, model.PropSlice[1].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, d1, model.PropSliceSlice[0][0].String())
	assert.Equal(t, d2, model.PropSliceSlice[1][0].String())
	assert.Equal(t, d3, model.PropSliceSlice[1][1].String())
	assert.Equal(t, d4, model.PropSliceSlice[2][0].String())

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, d1, model.PropMap["key1"].String())
	assert.Equal(t, d2, model.PropMap["key2"].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, d1, model.PropSliceMap["key1"][0].String())
	assert.Equal(t, d2, model.PropSliceMap["key1"][1].String())
	assert.Equal(t, d3, model.PropSliceMap["key2"][0].String())
	assert.Equal(t, d4, model.PropSliceMap["key2"][1].String())

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, d1, model.PropMapSlice[0]["key1"].String())
	assert.Equal(t, d2, model.PropMapSlice[1]["key2"].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, d1, model.PropSliceMapSlice[0]["key1"][0].String())
	assert.Equal(t, d2, model.PropSliceMapSlice[1]["key2"][0].String())
	assert.Equal(t, d3, model.PropSliceMapSlice[1]["key2"][1].String())
	assert.Equal(t, d4, model.PropSliceMapSlice[1]["key2"][2].String())

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue.String(), model.PropSlice[0].String())
	assert.Equal(t, d1, model.PropSlice[1].String())
	assert.Equal(t, zeroValue.String(), model.PropSlice[2].String())

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue.String(), model.PropSliceSlice[1][0].String())
	assert.Equal(t, d1, model.PropSliceSlice[2][0].String())
	assert.Equal(t, d1, model.PropSliceSlice[3][0].String())
	assert.Equal(t, zeroValue.String(), model.PropSliceSlice[3][1].String())

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	// It turns out that the standard strfmt.DateTime unmarshal code will actually successfully unmarshal ""
	// as a DateTime value (epoch time - 1970-01-01T00:00:00.000Z).
	err = UnmarshalPrimitive(rawMap, "bad_date1", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, "1970-01-01T00:00:00.000Z", model.Prop.String())

	err = UnmarshalPrimitive(rawMap, "bad_date2", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_date2'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_date3", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_date3'"))
	t.Logf("Expected error: %s\n", err.Error())
}

func TestUnmarshalPrimitiveUUID(t *testing.T) {
	type MyModel struct {
		Prop              *strfmt.UUID
		PropSlice         []strfmt.UUID
		PropSliceSlice    [][]strfmt.UUID
		PropMap           map[string]strfmt.UUID
		PropSliceMap      map[string][]strfmt.UUID
		PropMapSlice      []map[string]strfmt.UUID
		PropSliceMapSlice []map[string][]strfmt.UUID
	}

	jsonTemplate := `{
		"prop": "%u1",
		"prop_slice": ["%u1", "%u2"],
		"prop_slice_slice": [["%u1"], ["%u2", "%u3"], ["%u4" ]],
		"prop_map": { "key1": "%u1", "key2": "%u2" },
		"prop_slice_map": { "key1": ["%u1", "%u2"], "key2": ["%u3", "%u4"] },
		"prop_map_slice": [{"key1": "%u1"}, {"key2": "%u2"}],
		"prop_slice_map_slice": [{"key1": ["%u1"]}, {"key2": ["%u2", "%u3", "%u4"]} ],
		
		"empty_uuid": "",
		"bad_uuid": "not a real uuid",
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"null_prop": null,
		"slice_with_null": [ null, "%u1", null ],
		"slice_slice_with_null": [ null, [null], ["%u1"], ["%u1", null] ]
	}`

	u1 := "63769e9f-94e6-4ab6-8c68-dd33f69fb535"
	u2 := "e43db1b8-673a-4033-bf18-ded07172700f"
	u3 := "7c5a5c8c-bba1-453b-8e65-c56ffd0aab07"
	u4 := "43bde04f-5581-448e-bd51-50f554c41ac4"
	var zeroValue strfmt.UUID

	jsonString := strings.ReplaceAll(jsonTemplate, "%u1", u1)
	jsonString = strings.ReplaceAll(jsonString, "%u2", u2)
	jsonString = strings.ReplaceAll(jsonString, "%u3", u3)
	jsonString = strings.ReplaceAll(jsonString, "%u4", u4)

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, u1, model.Prop.String())

	var aUuid strfmt.UUID
	err = UnmarshalPrimitive(rawMap, "prop", &aUuid)
	assert.Nil(t, err)
	assert.Equal(t, u1, aUuid.String())

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, u1, model.PropSlice[0].String())
	assert.Equal(t, u2, model.PropSlice[1].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, u1, model.PropSliceSlice[0][0].String())
	assert.Equal(t, u2, model.PropSliceSlice[1][0].String())
	assert.Equal(t, u3, model.PropSliceSlice[1][1].String())
	assert.Equal(t, u4, model.PropSliceSlice[2][0].String())

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, u1, model.PropMap["key1"].String())
	assert.Equal(t, u2, model.PropMap["key2"].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, u1, model.PropSliceMap["key1"][0].String())
	assert.Equal(t, u2, model.PropSliceMap["key1"][1].String())
	assert.Equal(t, u3, model.PropSliceMap["key2"][0].String())
	assert.Equal(t, u4, model.PropSliceMap["key2"][1].String())

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, u1, model.PropMapSlice[0]["key1"].String())
	assert.Equal(t, u2, model.PropMapSlice[1]["key2"].String())

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, u1, model.PropSliceMapSlice[0]["key1"][0].String())
	assert.Equal(t, u2, model.PropSliceMapSlice[1]["key2"][0].String())
	assert.Equal(t, u3, model.PropSliceMapSlice[1]["key2"][1].String())
	assert.Equal(t, u4, model.PropSliceMapSlice[1]["key2"][2].String())

	// Tests involving a JSON null value
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue.String(), model.PropSlice[0].String())
	assert.Equal(t, u1, model.PropSlice[1].String())
	assert.Equal(t, zeroValue.String(), model.PropSlice[2].String())

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue.String(), model.PropSliceSlice[1][0].String())
	assert.Equal(t, u1, model.PropSliceSlice[2][0].String())
	assert.Equal(t, u1, model.PropSliceSlice[3][0].String())
	assert.Equal(t, zeroValue.String(), model.PropSliceSlice[3][1].String())

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "empty_uuid", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "bad_uuid", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
}

func TestUnmarshalPrimitiveAny(t *testing.T) {
	type MyModel struct {
		Prop              interface{}
		PropSlice         []interface{}
		PropSliceSlice    [][]interface{}
		PropMap           map[string]interface{}
		PropSliceMap      map[string][]interface{}
		PropMapSlice      []map[string]interface{}
		PropSliceMapSlice []map[string][]interface{}
	}

	jsonTemplate := `{
		"prop": "%s1",
		"prop_slice": [%n1, %n2],
		"prop_slice_slice": [["%s1"], [%b1, %b2], [%n1]],
		"prop_map": { "key1": %b1, "key2": %b2 },
		"prop_slice_map": { "key1": [%n1, %n2], "key2": [%n2, %n1] },
		"prop_map_slice": [{"key1": %f1}, {"key2": %f1}],
		"prop_slice_map_slice": [{"key1": ["%s1"]}, {"key2": [%n1, %n1, %n2]} ],
		
		"ok_type":  true,
		"not_a_slice": false,
		"ok_slice_type": [38, 26],
		"null_prop": null,
		"slice_with_null": [ null, "%s1", null ],
		"slice_slice_with_null": [ null, [null], ["%s1"], [%f1, null] ]
	}`

	s1 := "value1"
	n1 := int64(74)
	n2 := int64(44)
	b1 := true
	b2 := false
	f1 := float64(39.0001)
	var zeroValue interface{}

	jsonString := strings.ReplaceAll(jsonTemplate, "%s1", s1)
	jsonString = strings.ReplaceAll(jsonString, "%n1", fmt.Sprintf("%d", n1))
	jsonString = strings.ReplaceAll(jsonString, "%n2", fmt.Sprintf("%d", n2))
	jsonString = strings.ReplaceAll(jsonString, "%b1", "true")
	jsonString = strings.ReplaceAll(jsonString, "%b2", "false")
	jsonString = strings.ReplaceAll(jsonString, "%f1", fmt.Sprintf("%f", f1))

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)
	assert.Equal(t, s1, model.Prop.(string))

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, n1, int64(model.PropSlice[0].(float64)))
	assert.Equal(t, n2, int64(model.PropSlice[1].(float64)))

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	assert.Equal(t, s1, model.PropSliceSlice[0][0].(string))
	assert.Equal(t, b1, model.PropSliceSlice[1][0].(bool))
	assert.Equal(t, b2, model.PropSliceSlice[1][1].(bool))
	assert.Equal(t, n1, int64(model.PropSliceSlice[2][0].(float64)))

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)
	assert.Equal(t, b1, model.PropMap["key1"].(bool))
	assert.Equal(t, b2, model.PropMap["key2"].(bool))

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)
	assert.Equal(t, n1, int64(model.PropSliceMap["key1"][0].(float64)))
	assert.Equal(t, n2, int64(model.PropSliceMap["key1"][1].(float64)))
	assert.Equal(t, n2, int64(model.PropSliceMap["key2"][0].(float64)))
	assert.Equal(t, n1, int64(model.PropSliceMap["key2"][1].(float64)))

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)
	assert.Equal(t, f1, model.PropMapSlice[0]["key1"].(float64))
	assert.Equal(t, f1, model.PropMapSlice[1]["key2"].(float64))

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)
	assert.Equal(t, s1, model.PropSliceMapSlice[0]["key1"][0].(string))
	assert.Equal(t, n1, int64(model.PropSliceMapSlice[1]["key2"][0].(float64)))
	assert.Equal(t, n1, int64(model.PropSliceMapSlice[1]["key2"][1].(float64)))
	assert.Equal(t, n2, int64(model.PropSliceMapSlice[1]["key2"][2].(float64)))

	// Tests involving a JSON null value
	model.Prop = "bad value"
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	model.PropSlice = []interface{}{"bad1", "bad2"}
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	model.PropMap = map[string]interface{}{"key1": "value1"}
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue, model.PropSlice[0])
	assert.Equal(t, s1, model.PropSlice[1].(string))
	assert.Equal(t, zeroValue, model.PropSlice[2])

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[1][0])
	assert.Equal(t, s1, model.PropSliceSlice[2][0].(string))
	assert.Equal(t, f1, model.PropSliceSlice[3][0].(float64))
	assert.Equal(t, zeroValue, model.PropSliceSlice[3][1])

	// Negative tests
	model.Prop = nil
	err = UnmarshalPrimitive(rawMap, "ok_type", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "ok_slice_type", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
}

func TestUnmarshalPrimitiveAnyObject(t *testing.T) {
	type MyModel struct {
		Prop              map[string]interface{}
		PropSlice         []map[string]interface{}
		PropSliceSlice    [][]map[string]interface{}
		PropMap           map[string]map[string]interface{}
		PropSliceMap      map[string][]map[string]interface{}
		PropMapSlice      []map[string]map[string]interface{}
		PropSliceMapSlice []map[string][]map[string]interface{}
	}

	jsonTemplate := `{
		"prop": %o1,
		"prop_slice": [%o1, %o2],
		"prop_slice_slice": [[%o1], [%o2, %o1], [%o2]],
		"prop_map": { "key1": %o1, "key2": %o2 },
		"prop_slice_map": { "key1": [%o1, %o2], "key2": [%o2, %o1] },
		"prop_map_slice": [{"key1": %o1}, {"key2": %o2}],
		"prop_slice_map_slice": [{"key1": [%o1]}, {"key2": [%o1, %o1, %o2]} ],
		
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"null_prop": null,
		"slice_with_null": [ null, %o1, null ],
		"slice_slice_with_null": [ null, [null], [%o1], [%o1, null] ]
	}`

	o1 := `{"field1": "value1"}`
	o2 := `{"field2": "value2"}`
	var zeroValue map[string]interface{}
	o1AsAnyObject := map[string]interface{}{"field1": "value1"}

	jsonString := strings.ReplaceAll(jsonTemplate, "%o1", o1)
	jsonString = strings.ReplaceAll(jsonString, "%o2", o2)

	rawMap := unmarshalMap(jsonString)
	assert.NotNil(t, rawMap)

	model := new(MyModel)

	var err error

	// Positive tests
	err = UnmarshalPrimitive(rawMap, "prop", &model.Prop)
	assert.Nil(t, err)
	assert.NotNil(t, model.Prop)

	err = UnmarshalPrimitive(rawMap, "prop_slice", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)

	err = UnmarshalPrimitive(rawMap, "prop_slice_slice", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)

	assert.Equal(t, 3, len(model.PropSliceSlice))
	assert.Equal(t, 1, len(model.PropSliceSlice[0]))
	assert.Equal(t, 2, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))

	err = UnmarshalPrimitive(rawMap, "prop_map", &model.PropMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "prop_slice_map", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "prop_map_slice", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "prop_slice_map_slice", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceMapSlice)

	// Tests involving a JSON null value
	model.Prop = make(map[string]interface{})
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.Prop)
	assert.Nil(t, err)
	assert.Nil(t, model.Prop)

	model.PropSlice = make([]map[string]interface{}, 1)
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSlice)

	model.PropMap = make(map[string]map[string]interface{})
	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMap)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMap)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropMapSlice)

	err = UnmarshalPrimitive(rawMap, "null_prop", &model.PropSliceMapSlice)
	assert.Nil(t, err)
	assert.Nil(t, model.PropSliceMapSlice)

	model.PropSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_with_null", &model.PropSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSlice)
	assert.Equal(t, 3, len(model.PropSlice))
	assert.Equal(t, zeroValue, model.PropSlice[0])
	assert.Equal(t, o1AsAnyObject, model.PropSlice[1])
	assert.Equal(t, zeroValue, model.PropSlice[2])

	model.PropSliceSlice = nil
	err = UnmarshalPrimitive(rawMap, "slice_slice_with_null", &model.PropSliceSlice)
	assert.Nil(t, err)
	assert.NotNil(t, model.PropSliceSlice)
	assert.Equal(t, 4, len(model.PropSliceSlice))
	assert.Equal(t, 0, len(model.PropSliceSlice[0]))
	assert.Equal(t, 1, len(model.PropSliceSlice[1]))
	assert.Equal(t, 1, len(model.PropSliceSlice[2]))
	assert.Equal(t, 2, len(model.PropSliceSlice[3]))

	assert.Nil(t, model.PropSliceSlice[0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[1][0])
	assert.Equal(t, o1AsAnyObject, model.PropSliceSlice[2][0])
	assert.Equal(t, o1AsAnyObject, model.PropSliceSlice[3][0])
	assert.Equal(t, zeroValue, model.PropSliceSlice[3][1])

	// Negative tests
	err = UnmarshalPrimitive(rawMap, "bad_type", &model.Prop)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_type'"))

	err = UnmarshalPrimitive(rawMap, "not_a_slice", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'not_a_slice'"))
	t.Logf("Expected error: %s\n", err.Error())

	err = UnmarshalPrimitive(rawMap, "bad_slice_type", &model.PropSlice)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'bad_slice_type'"))
}

// Utility function that unmarshals a JSON string into a map containing RawMessage's.
func unmarshalMap(jsonString string) (result map[string]json.RawMessage) {
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		err := fmt.Errorf("Error unmarshalling initial json string %s\nerror: %s\n", jsonString, err.Error())
		panic(err)
	}
	return
}
