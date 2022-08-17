//go:build all || fast
// +build all fast

package core

// (C) Copyright IBM Corp. 2019, 2021.
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
	"encoding/json"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
)

func TestIsJSONMimeType(t *testing.T) {
	assert.True(t, IsJSONMimeType("application/json"))
	assert.True(t, IsJSONMimeType("APPlication/json"))
	assert.True(t, IsJSONMimeType("application/json;blah"))
	assert.True(t, IsJSONMimeType("application/vnd.docker.distribution.manifest.v2+json"))
	assert.True(t, IsJSONMimeType("application/vnd.anothervendor.custom.semantics+json"))
	assert.True(t, IsJSONMimeType("application/vnd.yet.another.vendor.with.custom.semantics.blah.v3+json;charset=UTF8"))

	assert.False(t, IsJSONMimeType("application/json-patch+patch"))
	assert.False(t, IsJSONMimeType("YOapplication/jsonYO"))
	assert.False(t, IsJSONMimeType("YOapplication/vnd.docker.distribution.manifest.v2+jsonYO"))
}

func TestIsJSONPatchMimeType(t *testing.T) {
	assert.True(t, IsJSONPatchMimeType("application/json-patch+json"))
	assert.True(t, IsJSONPatchMimeType("APPlication/json-PATCH+json"))
	assert.True(t, IsJSONPatchMimeType("application/json-patch+json;charset=UTF8"))

	assert.False(t, IsJSONPatchMimeType("application/json"))
	assert.False(t, IsJSONPatchMimeType("YOapplication/json-patch+jsonYO"))
}

func TestStringNilMapper(t *testing.T) {
	var s = "test string"
	assert.Equal(t, "", StringNilMapper(nil))
	assert.Equal(t, "test string", StringNilMapper(&s))
}

func TestValidateNotNil(t *testing.T) {
	var str *string
	assert.Nil(t, str)
	err := ValidateNotNil(str, "str should not be nil!")
	assert.NotNil(t, err, "Should have gotten an error for nil 'str' pointer")
	msg := err.Error()
	assert.Equal(t, "str should not be nil!", msg)

	type MyOperationOptions struct {
		Parameter1 *string
	}

	var options *MyOperationOptions
	assert.Nil(t, options, "options should be nil!")
	err = ValidateNotNil(options, "options param should not be nil")
	assert.NotNil(t, err, "Should have gotten an error for nil 'y' ptr")
	msg = err.Error()
	assert.Equal(t, "options param should not be nil", msg)

	err = ValidateNotNil("str", "")
	assert.Nil(t, err)
}

// This function is used to demonstrate the problem with comparing
// a function argument received as an "interface{}"" value with nil.
func isNilAsIntf(obj interface{}) bool {
	return obj == nil
}

func TestIsNil(t *testing.T) {
	assert.True(t, IsNil(nil))
	assert.False(t, IsNil("test"))

	type MyInnerModel struct {
		Name *string
	}

	type MyModel struct {
		InnerModel *MyInnerModel
		MyMap      map[string]interface{}
		MySlice    []string
	}
	myModel := &MyModel{}
	assert.NotNil(t, myModel)

	assert.True(t, IsNil(myModel.InnerModel))
	assert.True(t, myModel.InnerModel == nil)
	assert.False(t, isNilAsIntf(myModel.InnerModel))

	assert.True(t, IsNil(myModel.MyMap))
	assert.True(t, myModel.MyMap == nil)
	assert.False(t, isNilAsIntf(myModel.MyMap))

	assert.True(t, IsNil(myModel.MySlice))
	assert.True(t, myModel.MySlice == nil)
	assert.False(t, isNilAsIntf(myModel.MySlice))

	// Declare (but don't initialize) local vars of type map and slice.
	var testMap map[string]json.RawMessage
	var testSlice []json.RawMessage

	// Map and slice are nil.
	assert.True(t, testMap == nil)
	assert.True(t, testSlice == nil)

	// As interface{} variables, map and slice are NOT nil.
	assert.False(t, isNilAsIntf(testMap))
	assert.False(t, isNilAsIntf(testSlice))

	// IsNil() looks at the interface's value so will appear as nil.
	assert.True(t, IsNil(testMap))
	assert.True(t, IsNil(testSlice))

	// Declare two interface{} variables to hold the map and slice values.
	var testMapIntf interface{} = nil
	var testSliceIntf interface{} = nil

	// First, verify the interface{} variables are in fact nil.
	assert.True(t, testMapIntf == nil)
	assert.True(t, testSliceIntf == nil)

	// Assign the map and slice to the interface{} variables.
	testMapIntf = testMap
	testSliceIntf = testSlice

	// As interface{} values, map and slice are NOT nil.
	assert.False(t, testMapIntf == nil)   //nolint:staticcheck
	assert.False(t, testSliceIntf == nil) //nolint:staticcheck

	// IsNil() looks at the interface's value so will appear as nil.
	assert.True(t, IsNil(testMapIntf))
	assert.True(t, IsNil(testSliceIntf))

	// Initialize the map and slice.
	testMap = make(map[string]json.RawMessage)
	testSlice = make([]json.RawMessage, 0)

	assert.False(t, testMap == nil)
	assert.False(t, testSlice == nil)

	assert.Equal(t, 0, len(testMap))
	assert.Equal(t, 0, len(testSlice))

	assert.False(t, IsNil(testMap))
	assert.False(t, IsNil(testSlice))

	assert.False(t, isNilAsIntf(testMap))
	assert.False(t, isNilAsIntf(testSlice))
}

func TestValidateStruct(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
	}

	type User struct {
		FirstName *string   `json:"fname" validate:"required"`
		LastName  *string   `json:"lname" validate:"required"`
		Addresses []Address `json:"address" validate:"dive"`
	}

	type NoRequiredFields struct {
		FirstName *string `json:"fname"`
		LastName  *string `json:"lname"`
	}

	type StringPtrs struct {
		Field *string `validate:"required,ne="`
	}

	address := &Address{
		Street: "Eavesdown Docks",
		City:   "",
	}

	firstName := "Blossom"
	lastName := "Powerpuff"
	user := &User{
		FirstName: &firstName,
		LastName:  &lastName,
		Addresses: []Address{*address},
	}

	goodStruct := &Address{
		Street: "Beltorre Drive",
		City:   "Georgetown, TX",
	}

	badStruct := &Address{
		Street: "Beltorre Drive",
	}

	noReqFields := &NoRequiredFields{}

	stringPtrs := &StringPtrs{}

	var err error

	err = ValidateStruct(goodStruct, "goodStruct")
	assert.Nil(t, err)

	err = ValidateStruct(noReqFields, "noReqFields")
	assert.Nil(t, err)

	err = ValidateStruct(user, "userPtr")
	assert.NotNil(t, err)
	t.Logf("[01] Expected error: %s\n", err.Error())

	err = ValidateStruct(nil, "nilPtr")
	assert.NotNil(t, err)
	t.Logf("[02] Expected error: %s\n", err.Error())

	err = ValidateStruct(badStruct, "badStruct")
	assert.NotNil(t, err)
	t.Logf("[03] Expected error: %s\n", err.Error())

	err = ValidateStruct(address, "emptyRequiredFeild")
	assert.NotNil(t, err)
	t.Logf("[04] Expected error: %s\n", err.Error())

	err = ValidateStruct(stringPtrs, "stringPtrs")
	assert.NotNil(t, err)
	t.Logf("[05] Expected error: %s\n", err.Error())

	var addressPtr *Address = nil
	err = ValidateStruct(addressPtr, "addressPtr")
	assert.NotNil(t, err)

	stringPtrStruct := &StringPtrs{
		Field: StringPtr("XYZ"),
	}
	err = ValidateStruct(stringPtrStruct, "stringPtrStruct")
	assert.Nil(t, err)

	stringPtrStruct.Field = StringPtr("")
	err = ValidateStruct(stringPtrStruct, "stringPtrStruct")
	assert.NotNil(t, err)
	t.Logf("[06] Expected error: %s\n", err.Error())

	stringPtrStruct.Field = nil
	err = ValidateStruct(stringPtrStruct, "stringPtrStruct")
	assert.NotNil(t, err)
	t.Logf("[07] Expected error: %s\n", err.Error())
}

func TestHasBadFirstOrLastChar(t *testing.T) {
	assert.Equal(t, true, HasBadFirstOrLastChar("{hello}"))
	assert.Equal(t, true, HasBadFirstOrLastChar("hello}"))
	assert.Equal(t, true, HasBadFirstOrLastChar("\"hello"))
	assert.Equal(t, true, HasBadFirstOrLastChar("hello\""))
	assert.Equal(t, false, HasBadFirstOrLastChar("hello"))
}

func TestPointers(t *testing.T) {
	var str = "test"
	assert.Equal(t, &str, StringPtr(str))

	var boolVar = true
	assert.Equal(t, &boolVar, BoolPtr(boolVar))

	var intVar = int64(23)
	assert.Equal(t, &intVar, Int64Ptr(intVar))

	var float32Var = float32(23)
	assert.Equal(t, &float32Var, Float32Ptr(float32Var))

	var float64Var = float64(23)
	assert.Equal(t, &float64Var, Float64Ptr(float64Var))

	var uuidVar = strfmt.UUID("12345678-1234-1234-1234-123456123456")
	assert.Equal(t, &uuidVar, UUIDPtr(uuidVar))
}

func TestConvertSliceFloat64(t *testing.T) {
	float64Slice := []float64{float64(9.56), float64(4.56), float64(2.4)}
	expected := []string{"9.56", "4.56", "2.4"}
	convertedSlice, err := ConvertSlice(float64Slice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	float64Slice = []float64{}
	convertedSlice, err = ConvertSlice(float64Slice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceFloat32(t *testing.T) {
	float32Slice := []float32{float32(9.56), float32(4.56), float32(2.4)}
	expected := []string{"9.56", "4.56", "2.4"}
	convertedSlice, err := ConvertSlice(float32Slice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	float32Slice = []float32{}
	convertedSlice, err = ConvertSlice(float32Slice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceInt64(t *testing.T) {
	int64Slice := []int64{int64(38), int64(26), int64(22)}
	expected := []string{"38", "26", "22"}
	convertedSlice, err := ConvertSlice(int64Slice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	int64Slice = []int64{}
	convertedSlice, err = ConvertSlice(int64Slice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceInt(t *testing.T) {
	intSlice := []int{3, 2, 1}
	expected := []string{"3", "2", "1"}
	convertedSlice, err := ConvertSlice(intSlice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	intSlice = []int{}
	convertedSlice, err = ConvertSlice(intSlice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceByteArray(t *testing.T) {
	testString := "test string 1..."
	testString2 := "test string 2..."
	byteArray := []byte(testString)
	byteArray2 := []byte(testString2)
	byteArraySlice := [][]byte{byteArray, byteArray2}

	// base64 encoded value
	expected := []string{"dGVzdCBzdHJpbmcgMS4uLg==", "dGVzdCBzdHJpbmcgMi4uLg=="}
	convertedSlice, err := ConvertSlice(byteArraySlice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	byteArraySlice = [][]byte{}
	convertedSlice, err = ConvertSlice(byteArraySlice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceDate(t *testing.T) {
	date1 := strfmt.Date(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))
	date2 := strfmt.Date(time.Date(2020, time.November, 10, 23, 0, 0, 0, time.UTC))

	dateSlice := []strfmt.Date{date1, date2}
	expected := []string{"2009-11-10", "2020-11-10"}
	convertedSlice, err := ConvertSlice(dateSlice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	dateSlice = []strfmt.Date{}
	convertedSlice, err = ConvertSlice(dateSlice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceDateTime(t *testing.T) {
	date1 := strfmt.DateTime(time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC))
	date2 := strfmt.DateTime(time.Date(2020, time.November, 10, 23, 0, 0, 0, time.UTC))

	dateTimeSlice := []strfmt.DateTime{date1, date2}
	expected := []string{"2009-11-10T23:00:00.000Z", "2020-11-10T23:00:00.000Z"}
	convertedSlice, err := ConvertSlice(dateTimeSlice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	dateTimeSlice = []strfmt.DateTime{}
	convertedSlice, err = ConvertSlice(dateTimeSlice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceString(t *testing.T) {
	stringSlice := []string{"testString1", "testString2"}
	expected := []string{"testString1", "testString2"}
	convertedSlice, err := ConvertSlice(stringSlice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	stringSlice = []string{"\"testString1\"", "\"testString2\"", "C:\\Program_Files"}
	expected = []string{"\"testString1\"", "\"testString2\"", "C:\\Program_Files"}
	convertedSlice, err = ConvertSlice(stringSlice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	stringSlice = []string{}
	convertedSlice, err = ConvertSlice(stringSlice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceUUID(t *testing.T) {
	uuidSlice := []strfmt.UUID{
		"9fab83da-98cb-4f18-a7ba-b6f0435c9673",
		"aaffca34-de6d-11ea-87d0-0242ac130003",
	}
	expected := []string{
		"9fab83da-98cb-4f18-a7ba-b6f0435c9673",
		"aaffca34-de6d-11ea-87d0-0242ac130003",
	}
	convertedSlice, err := ConvertSlice(uuidSlice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	uuidSlice = []strfmt.UUID{}
	convertedSlice, err = ConvertSlice(uuidSlice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceBool(t *testing.T) {
	boolSlice := []bool{true, false, true}
	expected := []string{"true", "false", "true"}
	convertedSlice, err := ConvertSlice(boolSlice)

	assert.Nil(t, err)
	assert.NotNil(t, convertedSlice)
	assert.NotEmpty(t, convertedSlice)
	assert.Equal(t, expected, convertedSlice)

	boolSlice = []bool{}
	convertedSlice, err = ConvertSlice(boolSlice)

	assert.Nil(t, err)
	assert.Empty(t, convertedSlice)
}

func TestConvertSliceBadInput(t *testing.T) {
	// map[string]string
	convertedSlice, err := ConvertSlice(map[string]string{"foo": "bar"})
	assert.NotNil(t, err)
	assert.Nil(t, convertedSlice)

	// map[string]byte
	myByteMap := map[string][]byte{"myByteArray": {01, 02, 03, 04}}
	convertedSlice, err = ConvertSlice(myByteMap)
	assert.NotNil(t, err)
	assert.Nil(t, convertedSlice)

	//map[string]interface{}
	myGenericMap := make(map[string]interface{})
	convertedSlice, err = ConvertSlice(myGenericMap)
	assert.NotNil(t, err)
	assert.Nil(t, convertedSlice)

	// empty string
	convertedSlice, err = ConvertSlice("")
	assert.NotNil(t, err)
	assert.Nil(t, convertedSlice)

	// simple string
	convertedSlice, err = ConvertSlice("testString")
	assert.NotNil(t, err)
	assert.Nil(t, convertedSlice)

	// nil input
	var input string
	convertedSlice, err = ConvertSlice(input)
	assert.NotNil(t, err)
	assert.Nil(t, convertedSlice)

	// generic interface
	var i interface{}
	convertedSlice, err = ConvertSlice(i)
	assert.NotNil(t, err)
	assert.Nil(t, convertedSlice)

}

func TestSliceContains(t *testing.T) {
	theSlice := []string{"foo", "bar"}
	assert.True(t, SliceContains(theSlice, "foo"))
	assert.True(t, SliceContains(theSlice, "bar"))
	assert.False(t, SliceContains(theSlice, "gzip"))

	emptySlice := make([]string, 0)
	assert.False(t, SliceContains(emptySlice, "foo"))

	assert.False(t, SliceContains(nil, "foo"))
}

func TestGetQueryParam(t *testing.T) {
	nextURL := "/api/v1/offerings?start=foo&limit=10"
	next, err := GetQueryParam(&nextURL, "start")
	assert.Nil(t, err)
	assert.Equal(t, "foo", *next)

	fqNextURL := "https://acme.com/api/v1/offerings?start=bar&limit=10"
	next, err = GetQueryParam(&fqNextURL, "start")
	assert.Nil(t, err)
	assert.Equal(t, "bar", *next)

	// No query parameter
	next, err = GetQueryParam(&nextURL, "token")
	assert.Nil(t, err)
	assert.Nil(t, next)

	// nil URL
	next, err = GetQueryParam(nil, "start")
	assert.Nil(t, err)
	assert.Nil(t, next)

	// empty URL
	var emptyURL string
	next, err = GetQueryParam(&emptyURL, "start")
	assert.Nil(t, err)
	assert.Nil(t, next)

	// Not a URL (parse fails)
	badURL := "https://foo.bar:baz/api/v1/offerings?start=foo"
	next, err = GetQueryParam(&badURL, "start")
	assert.NotNil(t, err)
	assert.Nil(t, next)

	// No query string
	noQueryStringURL := "/api/v1/offerings"
	next, err = GetQueryParam(&noQueryStringURL, "start")
	assert.Nil(t, err)
	assert.Nil(t, next)

	// Bad query string
	badQueryURL := "/api/v1/offerings?start%XXfoo"
	next, err = GetQueryParam(&badQueryURL, "start")
	assert.NotNil(t, err)
	assert.Nil(t, next)

	// Duplicate param
	dupParamURL := "/api/v1/offerings?start=foo&start=bar&limit=10"
	next, err = GetQueryParam(&dupParamURL, "start")
	assert.Nil(t, err)
	assert.Equal(t, "foo", *next)
}

func TestGetQueryParamAsInt(t *testing.T) {
	nextURL := "/api/v1/offerings?offset=3&limit=10"
	next, err := GetQueryParamAsInt(&nextURL, "offset")
	assert.Nil(t, err)
	assert.Equal(t, int64(3), *next)

	fqNextURL := "https://acme.com/api/v1/offerings?next_offset=38&limit=10"
	next, err = GetQueryParamAsInt(&fqNextURL, "next_offset")
	assert.Nil(t, err)
	assert.Equal(t, int64(38), *next)

	// No query parameter
	next, err = GetQueryParamAsInt(&nextURL, "token")
	assert.Nil(t, err)
	assert.Nil(t, next)

	// nil URL
	next, err = GetQueryParamAsInt(nil, "offset")
	assert.Nil(t, err)
	assert.Nil(t, next)

	// empty URL
	var emptyURL string
	next, err = GetQueryParamAsInt(&emptyURL, "offset")
	assert.Nil(t, err)
	assert.Nil(t, next)

	// Not a URL (parse fails)
	badURL := "https://foo.bar:baz/api/v1/offerings?start=foo"
	next, err = GetQueryParamAsInt(&badURL, "offset")
	assert.NotNil(t, err)
	assert.Nil(t, next)

	// No query string
	noQueryStringURL := "/api/v1/offerings"
	next, err = GetQueryParamAsInt(&noQueryStringURL, "offset")
	assert.Nil(t, err)
	assert.Nil(t, next)

	// Bad query string
	badQueryURL := "/api/v1/offerings?offset%XXfoo"
	next, err = GetQueryParamAsInt(&badQueryURL, "offset")
	assert.NotNil(t, err)
	assert.Nil(t, next)

	// Bad query string
	badIntURL := "/api/v1/offerings?offset=foo"
	next, err = GetQueryParamAsInt(&badIntURL, "offset")
	assert.NotNil(t, err)
	assert.Nil(t, next)

	// Duplicate param
	dupParamURL := "/api/v1/offerings?offset=38&offset=26&limit=10"
	next, err = GetQueryParamAsInt(&dupParamURL, "offset")
	assert.Nil(t, err)
	assert.Equal(t, int64(38), *next)
}

func TestRedactSecrets(t *testing.T) {
	assert.NotContains(t, RedactSecrets("Authorization: Bearer secret"), "secret")
	assert.NotContains(t, RedactSecrets("Authorization: Basic secret"), "secret")
	assert.NotContains(t, RedactSecrets("X-Authorization: secret"), "secret")

	assert.NotContains(t, RedactSecrets("PASSword=secret"), "secret")
	assert.NotContains(t, RedactSecrets("ApIKey=secret"), "secret")
	assert.NotContains(t, RedactSecrets("toKen=secret"), "secret")
	assert.NotContains(t, RedactSecrets("passCode=secret"), "secret")

	assert.NotContains(t, RedactSecrets(`"token": "secret",`), "secret")
	assert.NotContains(t, RedactSecrets(`xxx "apIKEy":    "secret",xxx`), "secret")
}
