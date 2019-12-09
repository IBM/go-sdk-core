package core

// (C) Copyright IBM Corp. 2019.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsJSONMimeType(t *testing.T) {
	assert.True(t, IsJSONMimeType("application/json"))
	assert.True(t, IsJSONMimeType("APPlication/json"))
	assert.True(t, IsJSONMimeType("application/json;blah"))

	assert.False(t, IsJSONMimeType("application/json-patch+patch"))
	assert.False(t, IsJSONMimeType("YOapplication/jsonYO"))
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

func TestIsNil(t *testing.T) {
	assert.Equal(t, true, isNil(nil))
	assert.Equal(t, false, isNil("test"))
}

func TestValidateStruct(t *testing.T) {
	type Address struct {
		Street string `validate:"required"`
		City   string `validate:"required"`
	}

	type User struct {
		FirstName *string    `json:"fname" validate:"required"`
		LastName  *string    `json:"lname" validate:"required"`
		Addresses []*Address `json:"address" validate:"dive"`
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
		Addresses: []*Address{address},
	}

	goodStruct := &Address{
		Street: "Beltorre Drive",
		City:   "Georgetown, TX",
	}

	badStruct := &Address{
		Street: "Beltorre Drive",
	}

	assert.NotNil(t, ValidateStruct(user, "userPtr"), "Should have a validation error!")
	assert.Nil(t, ValidateStruct(nil, "nil ptr"), "nil pointer should validate cleanly!")
	assert.Nil(t, ValidateStruct(goodStruct, "goodStruct"), "Should not cause a validation error!")
	err := ValidateStruct(badStruct, "badStruct")
	assert.NotNil(t, err, "Should have a validation error!")
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
}
