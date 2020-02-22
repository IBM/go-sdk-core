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
package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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
		"prop1": "string value",
		"slice1": ["string1", "string2"],
		"incorrect_type":  true,
		"not_a_slice": false,
		"incorrect_slice_type": [38, 26],
		"null_prop": null,
		"null_slice": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalString(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, "string value", *value)

	value, err = UnmarshalString(testMap, "incorrect_type")
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

	slice, err := UnmarshalStringSlice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []string{"string1", "string2"}, slice)

	slice, err = UnmarshalStringSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalStringSlice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalStringSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalStringSlice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalByteArray(t *testing.T) {
	encodedString := base64.StdEncoding.EncodeToString([]byte("deadbeef"))
	assert.NotNil(t, encodedString)

	jsonTemplate := `{
		"prop1": "%s",
		"slice1": ["%s","%s"],
		"incorrect_type": true,
		"not_a_slice": false,
		"incorrect_slice_type": [38, 26],
		"invalid_byte_array": "this is not an encoded string!",
		"invalid_byte_array_slice": ["this is not an encoded string!"],
		"null_prop": null,
		"null_slice": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, encodedString, encodedString, encodedString)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalByteArray(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, []byte("deadbeef"), *value)

	value, err = UnmarshalByteArray(testMap, "incorrect_type")
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

	slice, err := UnmarshalByteArraySlice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, [][]byte{[]byte("deadbeef"), []byte("deadbeef")}, slice)

	slice, err = UnmarshalByteArraySlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArraySlice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a base64-encoded string but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArraySlice(testMap, "invalid_byte_array_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalByteArraySlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalByteArraySlice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalBool(t *testing.T) {
	jsonString := `{
		"prop1": true,
		"slice1": [false, true],
		"incorrect_type": "true",
		"not_a_slice": false,
		"incorrect_slice_type": [38, 26],
		"null_prop": null,
		"null_slice": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalBool(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, true, *value)

	value, err = UnmarshalBool(testMap, "incorrect_type")
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

	slice, err := UnmarshalBoolSlice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []bool{false, true}, slice)

	slice, err = UnmarshalBoolSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalBoolSlice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a boolean but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalBoolSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalBoolSlice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalInt64(t *testing.T) {
	jsonString := `{
		"prop1": 32,
		"slice1": [74, 44],
		"incorrect_type":  true,
		"not_a_slice": false,
		"incorrect_slice_type": ["blah"],
		"null_prop": null,
		"null_slice": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalInt64(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, int64(32), *value)

	value, err = UnmarshalInt64(testMap, "incorrect_type")
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

	slice, err := UnmarshalInt64Slice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []int64{74, 44}, slice)

	slice, err = UnmarshalInt64Slice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalInt64Slice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a integer but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalInt64Slice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalInt64Slice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalFloat64(t *testing.T) {
	jsonString := `{
		"prop1": 32.3,
		"slice1": [74.5, 44.8],
		"incorrect_type":  true,
		"not_a_slice": false,
		"incorrect_slice_type": ["blah"],
		"null_prop": null,
		"null_slice": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalFloat64(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, float64(32.3), *value)

	value, err = UnmarshalFloat64(testMap, "incorrect_type")
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

	slice, err := UnmarshalFloat64Slice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []float64{74.5, 44.8}, slice)

	slice, err = UnmarshalFloat64Slice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat64Slice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a float64 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat64Slice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalFloat64Slice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalFloat32(t *testing.T) {
	jsonString := `{
		"prop1": 32.3,
		"slice1": [74.5, 44.8],
		"incorrect_type":  true,
		"not_a_slice": false,
		"incorrect_slice_type": ["blah"],
		"null_prop": null,
		"null_slice": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalFloat32(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, float32(32.3), *value)

	value, err = UnmarshalFloat32(testMap, "incorrect_type")
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

	slice, err := UnmarshalFloat32Slice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	assert.Equal(t, []float32{74.5, 44.8}, slice)

	slice, err = UnmarshalFloat32Slice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat32Slice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a float32 but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalFloat32Slice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalFloat32Slice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalUUID(t *testing.T) {
	uuid1 := "9fab83da-98cb-4f18-a7ba-b6f0435c9673"
	uuid2 := "12ab83da-98cb-4f18-a7ba-b6f0435c0000"

	jsonTemplate := `{
		"prop1": "%s",
		"slice1": ["%s","%s"],
		"incorrect_type": true,
		"not_a_slice": false,
		"incorrect_slice_type": [true, false],
		"null_prop": null,
		"null_slice": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, uuid1, uuid1, uuid2)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalUUID(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, uuid1, value.String())

	value, err = UnmarshalUUID(testMap, "incorrect_type")
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

	slice, err := UnmarshalUUIDSlice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	actual := []string{slice[0].String(), slice[1].String()}
	assert.Equal(t, []string{uuid1, uuid2}, actual)

	slice, err = UnmarshalUUIDSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalUUIDSlice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a UUID but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalUUIDSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalUUIDSlice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalDate(t *testing.T) {
	date1 := "1970-01-01"
	date2 := "2019-12-23"

	jsonTemplate := `{
		"prop1": "%s",
		"slice1": ["%s","%s"],
		"incorrect_type": true,
		"not_a_slice": false,
		"incorrect_slice_type": [true, false],
		"invalid_date": "this is not a valid date",
		"invalid_date_slice": ["another invalid date value"],
		"null_prop": null,
		"null_slice": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, date1, date1, date2)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalDate(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, date1, value.String())

	value, err = UnmarshalDate(testMap, "incorrect_type")
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

	slice, err := UnmarshalDateSlice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	actual := []string{slice[0].String(), slice[1].String()}
	assert.Equal(t, []string{date1, date2}, actual)

	slice, err = UnmarshalDateSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateSlice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a Date but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateSlice(testMap, "invalid_date_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalDateSlice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalDateTime(t *testing.T) {
	datetime1 := "1970-01-01T01:02:03"
	datetime2 := "2019-12-23T23:59:59Z"
	datetime3 := "2019-12-31T23:59:59.333Z"

	jsonTemplate := `{
		"prop1": "%s",
		"slice1": ["%s","%s"],
		"incorrect_type": true,
		"not_a_slice": false,
		"incorrect_slice_type": [true, false],
		"invalid_datetime": "this is an invalid datetime value",
		"invalid_datetime_slice": ["another invalid datetime value"],
		"null_prop": null,
		"null_slice": null
	}`
	jsonString := fmt.Sprintf(jsonTemplate, datetime1, datetime2, datetime3)

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalDateTime(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, datetime1 + ".000Z", value.String())

	value, err = UnmarshalDateTime(testMap, "incorrect_type")
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

	slice, err := UnmarshalDateTimeSlice(testMap, "slice1")
	assert.Nil(t, err)
	assert.NotNil(t, slice)
	expectedSlice := []string{"2019-12-23T23:59:59.000Z","2019-12-31T23:59:59.333Z"} 
	actualSlice := []string{slice[0].String(), slice[1].String()}
	assert.Equal(t, expectedSlice, actualSlice)

	slice, err = UnmarshalDateTimeSlice(testMap, "not_a_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "should be an array but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeSlice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a DateTime but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeSlice(testMap, "invalid_datetime_slice")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "error decoding"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalDateTimeSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalDateTimeSlice(testMap, "null_slice")
	assert.Nil(t, err)
	assert.Nil(t, slice)
}

func TestUnmarshalObject(t *testing.T) {
	jsonString := `{
		"prop1": {"foo": "bar"},
		"slice1": [
			{"name": "object1"},
			{"name": "object2"},
			{"name": "object3"}
		],
		"incorrect_type":  true,
		"not_a_slice": false,
		"incorrect_slice_type": [false],
		"null_prop": null,
		"null_slice": null
	}`

	testMap, err := unmarshalJsonToMap(t, jsonString)
	assert.Nil(t, err)
	assert.NotNil(t, testMap)

	value, err := UnmarshalObject(testMap, "prop1")
	assert.Nil(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, "bar", value["foo"])

	value, err = UnmarshalObject(testMap, "incorrect_type")
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

	slice, err := UnmarshalObjectSlice(testMap, "slice1")
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

	slice, err = UnmarshalObjectSlice(testMap, "incorrect_slice_type")
	assert.NotNil(t, err)
	assert.Nil(t, slice)
	assert.True(t, strings.Contains(err.Error(), "array element should be a JSON object but was"))
	t.Logf("Expected error: %s\n", err.Error())

	slice, err = UnmarshalObjectSlice(testMap, "XXX")
	assert.Nil(t, err)
	assert.Nil(t, slice)

	slice, err = UnmarshalObjectSlice(testMap, "null_slice")
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
		"null_prop": null,
		"null_slice": null
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

	slice, err = UnmarshalAnySlice(testMap, "null_slice")
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
