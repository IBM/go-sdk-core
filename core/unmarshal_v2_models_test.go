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

package core

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// This struct simulates a generated model.
type MyModel struct {
	// A string property.
	Foo *string `json:"foo" validate:"required"`

	// An integer property.
	Bar *int64 `json:"bar" validate:"required"`
}

// This simulates a generated unmarshal function.
func UnmarshalMyModel(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(MyModel)
	err = UnmarshalPrimitive(m, "foo", &obj.Foo)
	if err != nil {
		return
	}
	err = UnmarshalPrimitive(m, "bar", &obj.Bar)
	if err != nil {
		return
	}

	objPtrPtr := result.(**MyModel)
	*objPtrPtr = obj
	return
}

// This struct simulates a generated model with properties that involve models.
type ModelStruct struct {
	Model         *MyModel             `json:"model" validate:"required"`
	ModelSlice    []MyModel            `json:"model_slice" validate:"required"`
	ModelMap      map[string]MyModel   `json:"model_map" validate:"required"`
	ModelSliceMap map[string][]MyModel `json:"model_slice_map" validate:"required"`
}

// This simulates a generated unmarshal function.
func UnmarshalModelStruct(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(ModelStruct)
	err = UnmarshalModel(m, "model", &obj.Model, UnmarshalMyModel)
	if err != nil {
		return
	}
	err = UnmarshalModel(m, "model_slice", &obj.ModelSlice, UnmarshalMyModel)
	if err != nil {
		return
	}
	err = UnmarshalModel(m, "model_map", &obj.ModelMap, UnmarshalMyModel)
	if err != nil {
		return
	}
	err = UnmarshalModel(m, "model_slice_map", &obj.ModelSliceMap, UnmarshalMyModel)
	if err != nil {
		return
	}
	objPtrPtr := result.(**ModelStruct)
	*objPtrPtr = obj
	return
}

func TestUnmarshalModelInstanceNil(t *testing.T) {
	jsonString := `{ 
		"null_model": null,
		"empty_model": { }
	}`
	rawMap := unmarshalMap(jsonString)

	var err error
	var myModel *MyModel

	// Unmarshal a missing property.
	myModel = nil
	err = UnmarshalModel(rawMap, "missing_model", &myModel, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Nil(t, myModel)

	// Unmarshal an explicit null value.
	myModel = nil
	err = UnmarshalModel(rawMap, "null_model", &myModel, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Nil(t, myModel)

	// Unmarshal an "empty" model instance.
	myModel = nil
	err = UnmarshalModel(rawMap, "empty_model", &myModel, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, myModel)
	assert.Nil(t, myModel.Foo)
	assert.Nil(t, myModel.Bar)
}

func TestUnmarshalModelInstance(t *testing.T) {
	jsonString := `{ "foo": "string1", "bar": 44 }`

	var err error
	var myModel *MyModel

	// Unmarshal an instance of the model.
	rawMap := unmarshalMap(jsonString)
	err = UnmarshalModel(rawMap, "", &myModel, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, myModel)
	assert.Equal(t, "string1", *(myModel.Foo))
	assert.Equal(t, int64(44), *(myModel.Bar))

}

func TestUnmarshalModelSliceNil(t *testing.T) {
	jsonString := `{ 
		"null_slice": null,
		"empty_slice": []
	}`
	rawMap := unmarshalMap(jsonString)

	var err error
	var mySlice []MyModel

	// Unmarshal a missing property.
	mySlice = nil
	err = UnmarshalModel(rawMap, "missing_slice", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mySlice))

	// Unmarshal an explicit null value.
	mySlice = nil
	err = UnmarshalModel(rawMap, "null_slice", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mySlice))

	// Unmarshal an explicit null value.
	mySlice = nil
	err = UnmarshalModel(rawMap, "empty_slice", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mySlice))
}

func TestUnmarshalModelSlice(t *testing.T) {
	jsonString := `[ { "foo": "string1", "bar": 44 }, { "foo": "string2", "bar": 74 } ]`
	var err error
	var mySlice []MyModel

	// Unmarshal a model slice.
	rawSlice := unmarshalSlice(jsonString)
	err = UnmarshalModel(rawSlice, "", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(mySlice))
	assert.Equal(t, "string1", *(mySlice[0].Foo))
	assert.Equal(t, int64(44), *(mySlice[0].Bar))
	assert.Equal(t, "string2", *(mySlice[1].Foo))
	assert.Equal(t, int64(74), *(mySlice[1].Bar))
}

func TestUnmarshalModelMapNil(t *testing.T) {
	jsonString := `{ 
		"null_map": null,
		"empty_map": { }
	}`
	rawMap := unmarshalMap(jsonString)

	var err error
	var myMap map[string]MyModel

	// Unmarshal a missing property.
	myMap = nil
	err = UnmarshalModel(rawMap, "missing_map", &myMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(myMap))

	// Unmarshal an explicit null value.
	myMap = nil
	err = UnmarshalModel(rawMap, "null_map", &myMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, myMap)
	assert.Equal(t, 0, len(myMap))

	// Unmarshal an empty map.
	myMap = nil
	err = UnmarshalModel(rawMap, "empty_map", &myMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, myMap)
	assert.Equal(t, 0, len(myMap))
}

func TestUnmarshalModelMap(t *testing.T) {
	jsonString := `{
		"model1": { "foo": "string1", "bar": 44 }, 
		"model2": { "foo": "string2", "bar": 74 }
	}`
	var err error
	var myMap map[string]MyModel

	// Unmarshal a model map.
	rawMap := unmarshalMap(jsonString)
	err = UnmarshalModel(rawMap, "", &myMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(myMap))
	assert.NotNil(t, myMap["model1"])
	assert.NotNil(t, myMap["model2"])
	
	_, foundIt := myMap["bad_key"]
	assert.False(t, foundIt)
	
	assert.Equal(t, "string1", *(myMap["model1"].Foo))
	assert.Equal(t, int64(44), *(myMap["model1"].Bar))
	assert.Equal(t, "string2", *(myMap["model2"].Foo))
	assert.Equal(t, int64(74), *(myMap["model2"].Bar))
}

func TestUnmarshalModelSliceMapNil(t *testing.T) {
	jsonString := `{ 
		"null_prop": null, 
		"empty_slice_map": { 
			"empty_slice": [] 
		},
		"null_slice_map": { 
			"null_slice": null
		}
	}`
	rawMap := unmarshalMap(jsonString)

	var err error
	var mySliceMap map[string][]MyModel

	// Unmarshal a missing property.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "missing_prop", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mySliceMap))

	// Unmarshal an explicit null value.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "null_prop", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(mySliceMap))

	// Unmarshal a map with an empty slice.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "empty_slice_map", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(mySliceMap))
	assert.NotNil(t, mySliceMap["empty_slice"])
	assert.Equal(t, 0, len(mySliceMap["empty_slice"]))

	// Unmarshal a map with an explicit null slice.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "null_slice_map", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(mySliceMap))
	assert.NotNil(t, mySliceMap["null_slice"])
	assert.Equal(t, 0, len(mySliceMap["null_slice"]))
}

func TestUnmarshalModelStruct(t *testing.T) {
	m1 := `{ "foo": "string1", "bar": 44 }`
	m2 := `{ "foo": "string2", "bar": 74 }`
	m3 := `{ "foo": "string3", "bar": 33 }`
	m4 := `{ "foo": "string4", "bar": 21 }`
	jsonTemplate := `{
		"model": %m1,
		"model_slice": [ %m1, %m2 ],
		"model_map": {
			"model1": %m1,
			"model2": %m2,
			"model3": %m3,
			"model4": %m4
		},
		"model_slice_map": {
			"slice1": [ %m1, %m2, %m3 ],
			"slice2": [ %m4 ]
		}
	}`
	jsonString := strings.ReplaceAll(jsonTemplate, "%m1", m1)
	jsonString = strings.ReplaceAll(jsonString, "%m2", m2)
	jsonString = strings.ReplaceAll(jsonString, "%m3", m3)
	jsonString = strings.ReplaceAll(jsonString, "%m4", m4)

	var err error
	var modelStruct *ModelStruct

	rawMap := unmarshalMap(jsonString)
	err = UnmarshalModel(rawMap, "", &modelStruct, UnmarshalModelStruct)
	assert.Nil(t, err)
	assert.NotNil(t, modelStruct)
	assert.NotNil(t, modelStruct.Model)
	assert.Equal(t, 2, len(modelStruct.ModelSlice))
	assert.Equal(t, 4, len(modelStruct.ModelMap))
	assert.Equal(t, 2, len(modelStruct.ModelSliceMap))

	// Verify model instance.
	assert.Equal(t, "string1", *(modelStruct.Model.Foo))
	assert.Equal(t, int64(44), *(modelStruct.Model.Bar))
	
	// Verify model slice.
	assert.Equal(t, "string1", *(modelStruct.ModelSlice[0].Foo))
	assert.Equal(t, int64(44), *(modelStruct.ModelSlice[0].Bar))
	assert.Equal(t, "string2", *(modelStruct.ModelSlice[1].Foo))
	assert.Equal(t, int64(74), *(modelStruct.ModelSlice[1].Bar))
	
	// Verify model map.
	assert.NotNil(t, modelStruct.ModelMap["model1"])
	assert.NotNil(t, modelStruct.ModelMap["model2"])
	assert.NotNil(t, modelStruct.ModelMap["model3"])
	assert.NotNil(t, modelStruct.ModelMap["model4"])
	_, foundIt := modelStruct.ModelMap["bad_key"]
	assert.False(t, foundIt)
	
	assert.Equal(t, "string1", *(modelStruct.ModelMap["model1"].Foo))
	assert.Equal(t, int64(44), *(modelStruct.ModelMap["model1"].Bar))
	assert.Equal(t, "string2", *(modelStruct.ModelMap["model2"].Foo))
	assert.Equal(t, int64(74), *(modelStruct.ModelMap["model2"].Bar))
	assert.Equal(t, "string3", *(modelStruct.ModelMap["model3"].Foo))
	assert.Equal(t, int64(33), *(modelStruct.ModelMap["model3"].Bar))
	assert.Equal(t, "string4", *(modelStruct.ModelMap["model4"].Foo))
	assert.Equal(t, int64(21), *(modelStruct.ModelMap["model4"].Bar))
	
	// Verify model slice map.
	assert.NotNil(t, modelStruct.ModelSliceMap["slice1"])
	assert.Equal(t, 3, len(modelStruct.ModelSliceMap["slice1"]))
	assert.NotNil(t, modelStruct.ModelSliceMap["slice2"])
	assert.Equal(t, 1, len(modelStruct.ModelSliceMap["slice2"]))
	
	assert.Equal(t, "string1", *(modelStruct.ModelSliceMap["slice1"][0].Foo))
	assert.Equal(t, int64(44), *(modelStruct.ModelSliceMap["slice1"][0].Bar))
	assert.Equal(t, "string2", *(modelStruct.ModelSliceMap["slice1"][1].Foo))
	assert.Equal(t, int64(74), *(modelStruct.ModelSliceMap["slice1"][1].Bar))
	assert.Equal(t, "string3", *(modelStruct.ModelSliceMap["slice1"][2].Foo))
	assert.Equal(t, int64(33), *(modelStruct.ModelSliceMap["slice1"][2].Bar))
	
	assert.Equal(t, "string4", *(modelStruct.ModelSliceMap["slice2"][0].Foo))
	assert.Equal(t, int64(21), *(modelStruct.ModelSliceMap["slice2"][0].Bar))
}

// Utility function that unmarshals a JSON string into
func unmarshalSlice(jsonString string) (result []json.RawMessage) {
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		panic(err)
	}
	return
}
