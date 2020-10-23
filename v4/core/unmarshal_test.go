// +build all fast

package core

/**
 * (C) Copyright IBM Corp. 2019.
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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyMap(t *testing.T) {
	var i int64 = 38
	var s string = "string value"
	innerMap := map[string]interface{}{
		"key": "value",
	}

	originalMap := map[string]interface{}{
		"key1": &s,
		"key2": &i,
		"key3": innerMap,
	}

	newMap := CopyMap(originalMap)
	assert.NotNil(t, newMap)
	assert.Equal(t, originalMap, newMap)
	assert.Equal(t, &s, newMap["key1"])
	assert.Equal(t, &i, newMap["key2"])
	assert.Equal(t, innerMap, newMap["key3"])
}

func TestUnmarshalString(t *testing.T) {
	jsonString := `{
		"good_prop": "string value",
		"good_slice": ["string1", "string2"],
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalString(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, "string value", *value)

	value, err = UnmarshalString(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalString(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalString(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalStringSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []string{"string1", "string2"}, slice)

	slice, err = UnmarshalStringSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalStringSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalStringSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalStringSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalStringMap(t *testing.T) {
	jsonString := `{
		"good_map": {"key1": "value1"},
		"not_a_map":  true,
		"bad_value_type" : {"key1": false},
		"good_slice": [{"key1": "value1"}, {"key2": "value2"}],
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"bad_slice_value_type": [{"key1": "value1"}, {"key2": false}],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalStringMap(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, map[string]string{"key1": "value1"}, value)

	value, err = UnmarshalStringMap(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a map[string]string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalStringMap(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalStringMap(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalStringMap(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalStringMapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []map[string]string{{"key1": "value1"}, {"key2": "value2"}}, slice)

	slice, err = UnmarshalStringMapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalStringMapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalStringMapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalStringMapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalStringMapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalByteArray(t *testing.T) {
	encodedString := base64.StdEncoding.EncodeToString([]byte("deadbeef"))
	assert.NotNil(t, encodedString)

	jsonTemplate := `{
		"good_prop": "%s",
		"good_slice": ["%s","%s"],
		"bad_type": true,
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"invalid_byte_array": "this is not an encoded string!",
		"invalid_byte_array_slice": ["this is not an encoded string!"],
		"null_prop": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, encodedString, encodedString, encodedString)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalByteArray(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, []byte("deadbeef"), *value)

	value, err = UnmarshalByteArray(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a base64-encoded string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalByteArray(testMap, "invalid_byte_array")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalByteArray(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalByteArray(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalByteArraySlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, [][]byte{[]byte("deadbeef"), []byte("deadbeef")}, slice)

	slice, err = UnmarshalByteArraySlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArraySlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be a base64-encoded string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArraySlice(testMap, "invalid_byte_array_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArraySlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalByteArraySlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalByteArrayMap(t *testing.T) {
	deadbeef := []byte("deadbeef")
	encodedString := base64.StdEncoding.EncodeToString(deadbeef)
	assert.NotNil(t, encodedString)

	jsonTemplate := `{
		"good_map": {"key1": "%s"},
		"not_a_map":  true,
		"bad_value_type" : {"key1": false},
		"good_slice": [{"key1": "%s"}, {"key2": "%s"}],
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"bad_slice_value_type": [{"key1": "%s"}, {"key2": false}],
		"invalid_byte_array": {"key1": "this is not an encoded string!"},
		"invalid_byte_array_slice": [{"key1": "this is not an encoded string!"}],
		"null_prop": null
	}`

	jsonString := fmt.Sprintf(jsonTemplate, encodedString, encodedString, encodedString, encodedString)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalByteArrayMap(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	expectedMap := map[string][]byte{"key1": deadbeef}
	assert.Equal(t, expectedMap, value)

	value, err = UnmarshalByteArrayMap(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalByteArrayMap(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a base64-encoded string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalByteArrayMap(testMap, "invalid_byte_array")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalByteArrayMap(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalByteArrayMap(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalByteArrayMapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)

	expectedMap2 := map[string][]byte{"key2": deadbeef}
	var expectedSlice []map[string][]byte
	expectedSlice = append(expectedSlice, expectedMap)
	expectedSlice = append(expectedSlice, expectedMap2)
	assert.Equal(t, expectedSlice, slice)

	slice, err = UnmarshalByteArrayMapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArrayMapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be a map[string]string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArrayMapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be a base64-encoded string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArrayMapSlice(testMap, "invalid_byte_array_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArrayMapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalByteArrayMapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalBool(t *testing.T) {
	jsonString := `{
		"good_prop": true,
		"good_slice": [false, true],
		"bad_type": "true",
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalBool(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, true, *value)

	value, err = UnmarshalBool(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a boolean but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalBool(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalBool(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalBoolSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []bool{false, true}, slice)

	slice, err = UnmarshalBoolSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalBoolSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a boolean but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalBoolSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalBoolSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalBoolMap(t *testing.T) {
	jsonString := `{
		"good_map": {"key1": true},
		"not_a_map":  "not a map",
		"bad_value_type" : {"key1": "false"},
		"good_slice": [{"key1": false}, {"key2": true}],
		"not_a_slice": false,
		"bad_slice_type": [38, 26],
		"bad_slice_value_type": [{"key1": false}, {"key2": 38}],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalBoolMap(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, map[string]bool{"key1": true}, value)

	value, err = UnmarshalBoolMap(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a map[string]bool but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalBoolMap(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a bool but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalBoolMap(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalBoolMap(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalBoolMapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []map[string]bool{{"key1": false}, {"key2": true}}, slice)

	slice, err = UnmarshalBoolMapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalBoolMapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]bool but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalBoolMapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a bool but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalBoolMapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalBoolMapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalInt64(t *testing.T) {
	jsonString := `{
		"good_prop": 32,
		"good_slice": [74, 44],
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": ["blah"],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalInt64(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, int64(32), *value)

	value, err = UnmarshalInt64(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a integer but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalInt64(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalInt64(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalInt64Slice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []int64{74, 44}, slice)

	slice, err = UnmarshalInt64Slice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalInt64Slice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a integer but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalInt64Slice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalInt64Slice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalInt64Map(t *testing.T) {
	jsonString := `{
		"good_map": {"key1": 38},
		"not_a_map":  "not a map",
		"bad_value_type" : {"key1": "bad_value"},
		"good_slice": [{"key1": 38}, {"key2": 26}],
		"not_a_slice": "false",
		"bad_slice_type": [38, 26],
		"bad_slice_value_type": [{"key1": 38}, {"key2": true}],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalInt64Map(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, map[string]int64{"key1": 38}, value)

	value, err = UnmarshalInt64Map(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a map[string]int64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalInt64Map(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a int64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalInt64Map(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalInt64Map(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalInt64MapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []map[string]int64{{"key1": 38}, {"key2": 26}}, slice)

	slice, err = UnmarshalInt64MapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalInt64MapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]int64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalInt64MapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a int64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalInt64MapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalInt64MapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalFloat32(t *testing.T) {
	jsonString := `{
		"good_prop": 32.3,
		"good_slice": [74.5, 44.8],
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": ["blah"],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalFloat32(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, float32(32.3), *value)

	value, err = UnmarshalFloat32(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a float32 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalFloat32(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalFloat32(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalFloat32Slice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []float32{74.5, 44.8}, slice)

	slice, err = UnmarshalFloat32Slice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat32Slice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a float32 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat32Slice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalFloat32Slice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalFloat32Map(t *testing.T) {
	jsonString := `{
		"good_map": {"key1": 38.5},
		"not_a_map":  "not a map",
		"bad_value_type" : {"key1": "bad_value"},
		"good_slice": [{"key1": 38.5}, {"key2": 26.2}],
		"not_a_slice": "false",
		"bad_slice_type": [38.5, 26.2],
		"bad_slice_value_type": [{"key1": 38.5}, {"key2": true}],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalFloat32Map(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, map[string]float32{"key1": 38.5}, value)

	value, err = UnmarshalFloat32Map(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]float32 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalFloat32Map(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a float32 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalFloat32Map(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalFloat32Map(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalFloat32MapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []map[string]float32{{"key1": 38.5}, {"key2": 26.2}}, slice)

	slice, err = UnmarshalFloat32MapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat32MapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]float32 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat32MapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a float32 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat32MapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalFloat32MapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalFloat64(t *testing.T) {
	jsonString := `{
		"good_prop": 32.3,
		"good_slice": [74.5, 44.8],
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": ["blah"],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalFloat64(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, float64(32.3), *value)

	value, err = UnmarshalFloat64(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a float64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalFloat64(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalFloat64(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalFloat64Slice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []float64{74.5, 44.8}, slice)

	slice, err = UnmarshalFloat64Slice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat64Slice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a float64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat64Slice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalFloat64Slice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalFloat64Map(t *testing.T) {
	jsonString := `{
		"good_map": {"key1": 38.5},
		"not_a_map":  "not a map",
		"bad_value_type" : {"key1": "bad_value"},
		"good_slice": [{"key1": 38.5}, {"key2": 26.2}],
		"not_a_slice": "false",
		"bad_slice_type": [38.5, 26.2],
		"bad_slice_value_type": [{"key1": 38.5}, {"key2": true}],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalFloat64Map(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, map[string]float64{"key1": 38.5}, value)

	value, err = UnmarshalFloat64Map(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]float64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalFloat64Map(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a float64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalFloat64Map(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalFloat64Map(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalFloat64MapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []map[string]float64{{"key1": 38.5}, {"key2": 26.2}}, slice)

	slice, err = UnmarshalFloat64MapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat64MapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]float64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat64MapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a float64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat64MapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalFloat64MapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalUUID(t *testing.T) {
	uuid1 := "9fab83da-98cb-4f18-a7ba-b6f0435c9673"
	uuid2 := "12ab83da-98cb-4f18-a7ba-b6f0435c0000"

	jsonTemplate := `{
		"good_prop": "%s",
		"good_slice": ["%s","%s"],
		"bad_type": true,
		"not_a_slice": false,
		"bad_slice_type": [true, false],
		"null_prop": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, uuid1, uuid1, uuid2)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalUUID(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, uuid1, value.String())

	value, err = UnmarshalUUID(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a UUID but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalUUID(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalUUID(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalUUIDSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	actual := []string{slice[0].String(), slice[1].String()}
	assert.Equal(t, []string{uuid1, uuid2}, actual)

	slice, err = UnmarshalUUIDSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalUUIDSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a UUID but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalUUIDSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalUUIDSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalUUIDMap(t *testing.T) {
	uuid1 := "9fab83da-98cb-4f18-a7ba-b6f0435c9673"
	uuid2 := "12ab83da-98cb-4f18-a7ba-b6f0435c0000"

	jsonTemplate := `{
		"good_map": {"key1": "%s"},
		"not_a_map":  "not a map",
		"bad_value_type" : {"key1": 38},
		"good_slice": [{"key1": "%s"}, {"key2": "%s"}],
		"not_a_slice": "false",
		"bad_slice_type": ["%s", "%s"],
		"bad_slice_value_type": [{"key1": "%s"}, {"key2": true}],
		"null_prop": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, uuid1, uuid1, uuid2, uuid1, uuid2, uuid1)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalUUIDMap(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, uuid1, value["key1"].String())

	value, err = UnmarshalUUIDMap(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]UUID but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalUUIDMap(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a UUID but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalUUIDMap(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalUUIDMap(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalUUIDMapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, uuid1, slice[0]["key1"].String())
	assert.Equal(t, uuid2, slice[1]["key2"].String())

	slice, err = UnmarshalUUIDMapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalUUIDMapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]UUID but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalUUIDMapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a UUID but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalUUIDMapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalUUIDMapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalDate(t *testing.T) {
	date1 := "1970-01-01"
	date2 := "2019-12-23"

	jsonTemplate := `{
		"good_prop": "%s",
		"good_slice": ["%s","%s"],
		"bad_type": true,
		"not_a_slice": false,
		"bad_slice_type": [true, false],
		"invalid_date": "this is not a valid date",
		"invalid_date_slice": ["another invalid date value"],
		"null_prop": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, date1, date1, date2)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalDate(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, date1, value.String())

	value, err = UnmarshalDate(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a Date but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalDate(testMap, "invalid_date")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalDate(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalDate(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalDateSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	actual := []string{slice[0].String(), slice[1].String()}
	assert.Equal(t, []string{date1, date2}, actual)

	slice, err = UnmarshalDateSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a Date but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateSlice(testMap, "invalid_date_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalDateSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalDateMap(t *testing.T) {
	date1 := "1970-01-01"
	date2 := "2019-12-23"

	jsonTemplate := `{
		"good_map": {"key1": "%s"},
		"not_a_map":  "not a map",
		"bad_value_type" : {"key1": 38},
		"good_slice": [{"key1": "%s"}, {"key2": "%s"}],
		"not_a_slice": "false",
		"bad_slice_type": ["%s", "%s"],
		"bad_slice_value_type": [{"key1": "%s"}, {"key2": true}],
		"null_prop": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, date1, date1, date2, date1, date2, date1)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalDateMap(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, date1, value["key1"].String())

	value, err = UnmarshalDateMap(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]Date but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalDateMap(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a Date but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalDateMap(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalDateMap(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalDateMapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, date1, slice[0]["key1"].String())
	assert.Equal(t, date2, slice[1]["key2"].String())

	slice, err = UnmarshalDateMapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateMapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]Date but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateMapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a Date but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateMapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalDateMapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalDateTime(t *testing.T) {
	datetime1 := "1970-01-01T01:02:03"
	datetime2 := "2019-12-23T23:59:59Z"
	datetime3 := "2019-12-31T23:59:59.333Z"

	jsonTemplate := `{
		"good_prop": "%s",
		"good_slice": ["%s","%s"],
		"bad_type": true,
		"not_a_slice": false,
		"bad_slice_type": [true, false],
		"invalid_datetime": "this is an invalid datetime value",
		"invalid_datetime_slice": ["another invalid datetime value"],
		"null_prop": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, datetime1, datetime2, datetime3)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalDateTime(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, datetime1+".000Z", value.String())

	value, err = UnmarshalDateTime(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a DateTime but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalDateTime(testMap, "invalid_datetime")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalDateTime(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalDateTime(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalDateTimeSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	expectedSlice := []string{"2019-12-23T23:59:59.000Z", "2019-12-31T23:59:59.333Z"}
	actualSlice := []string{slice[0].String(), slice[1].String()}
	assert.Equal(t, expectedSlice, actualSlice)

	slice, err = UnmarshalDateTimeSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a DateTime but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeSlice(testMap, "invalid_datetime_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalDateTimeSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalDateTimeMap(t *testing.T) {
	datetime1 := "1970-01-01T01:02:03"
	datetime2 := "2019-12-23T23:59:59Z"

	jsonTemplate := `{
		"good_map": {"key1": "%s"},
		"not_a_map":  "not a map",
		"bad_value_type" : {"key1": 38},
		"good_slice": [{"key1": "%s"}, {"key2": "%s"}],
		"not_a_slice": "false",
		"bad_slice_type": ["%s", "%s"],
		"bad_slice_value_type": [{"key1": "%s"}, {"key2": true}],
		"null_prop": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, datetime1, datetime1, datetime2, datetime1, datetime2, datetime1)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalDateTimeMap(testMap, "good_map")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, datetime1+".000Z", value["key1"].String())

	value, err = UnmarshalDateTimeMap(testMap, "not_a_map")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]DateTime but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalDateTimeMap(testMap, "bad_value_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "value should be a DateTime but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalDateTimeMap(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalDateTimeMap(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalDateTimeMapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, datetime1+".000Z", slice[0]["key1"].String())
	assert.Equal(t, "2019-12-23T23:59:59.000Z", slice[1]["key2"].String())

	slice, err = UnmarshalDateTimeMapSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeMapSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a map[string]DateTime but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeMapSlice(testMap, "bad_slice_value_type")
	assert.NotNil(t, err)
	assert.NotNil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "value should be a DateTime but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeMapSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalDateTimeMapSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalObject(t *testing.T) {
	jsonString := `{
		"good_prop": {"foo": "bar"},
		"good_slice": [
			{"name": "object1"},
			{"name": "object2"},
			{"name": "object3"}
		],
		"bad_type":  true,
		"not_a_slice": false,
		"bad_slice_type": [false],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalObject(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, "bar", value["foo"])

	value, err = UnmarshalAnyMap(testMap, "good_prop")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, "bar", value["foo"])

	value, err = UnmarshalObject(testMap, "bad_type")
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.True(t, strings.Contains(err.Error(), "should be a JSON object but was"))
	t.Logf("Expected error: %s\n", err.Error())

	value, err = UnmarshalObject(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalObject(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalObjectSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, "object1", slice[0]["name"])
	assert.Equal(t, "object2", slice[1]["name"])
	assert.Equal(t, "object3", slice[2]["name"])

	slice, err = UnmarshalAnyMapSlice(testMap, "good_slice")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, "object1", slice[0]["name"])
	assert.Equal(t, "object2", slice[1]["name"])
	assert.Equal(t, "object3", slice[2]["name"])

	slice, err = UnmarshalObjectSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalObjectSlice(testMap, "bad_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a JSON object but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalObjectSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalObjectSlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalAny(t *testing.T) {
	jsonString := `{
		"prop1": {"foo": "bar"},
		"prop2": true,
		"prop3": 33,
		"prop4": "a string",
		"slice1": [
			{"name": "object1"},
			{"name": "object2"},
			{"name": "object3"}
		],
		"slice2": [
		    true,
		    false
		],
		"slice3": [
		    74,
		    44
		],
		"slice4": [
		    "football",
		    "baseball"
		],
		"null_prop": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalAny(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)

	value, err = UnmarshalAny(testMap, "prop2")
	assert.Nil(t, err)
	assert.NotNil(t, value)

	value, err = UnmarshalAny(testMap, "prop3")
	assert.Nil(t, err)
	assert.NotNil(t, value)

	value, err = UnmarshalAny(testMap, "prop4")
	assert.Nil(t, err)
	assert.NotNil(t, value)

	value, err = UnmarshalAny(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, value)

	value, err = UnmarshalAny(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, value)

	slice, err := UnmarshalAnySlice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)

	slice, err = UnmarshalAnySlice(testMap, "slice2")
	assert.Nil(t, err)
	assert.NotNil(t, slice)

	slice, err = UnmarshalAnySlice(testMap, "slice3")
	assert.Nil(t, err)
	assert.NotNil(t, slice)

	slice, err = UnmarshalAnySlice(testMap, "slice4")
	assert.Nil(t, err)
	assert.NotNil(t, slice)

	slice, err = UnmarshalAnySlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalAnySlice(testMap, "null_prop")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

// unmarshalJson is a utility function that unmarshals a JSON string into the specified target object in
// much the same way as the BaseService.Request() method does so that we can simulate that behavior here for testing.
func unmarshalJson(t *testing.T, jsonString string, target interface{}) (result interface{}, err error) {
	buffer := []byte(jsonString)
	err = json.NewDecoder(bytes.NewReader(buffer)).Decode(&target)
	result = target
	return
}

// unmarshalJsonToMap is a convenience function used by the various test methods to unmarshal
// a JSON string into a generic map.
func unmarshalJsonToMap(t *testing.T, jsonString string) (result map[string]interface{}, err error) {
	v, err := unmarshalJson(t, jsonString, make(map[string]interface{}))
	if err != nil {
		return
	}

	var ok bool
	result, ok = v.(map[string]interface{})
	if !ok {
		err = fmt.Errorf("couldn't cast result to a map!")
	}
	return
}
