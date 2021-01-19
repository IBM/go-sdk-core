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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// This simulates a generated model with its unmarshal function.
type MyModel struct {
	Foo *string `json:"foo" validate:"required"`
	Bar *int64  `json:"bar" validate:"required"`
}

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
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// This simulates a generated model with properties that involve models, along with its unmarshal function.
type ModelStruct struct {
	Model         *MyModel             `json:"model" validate:"required"`
	ModelSlice    []MyModel            `json:"model_slice" validate:"required"`
	ModelMap      map[string]MyModel   `json:"model_map" validate:"required"`
	ModelSliceMap map[string][]MyModel `json:"model_slice_map" validate:"required"`
}

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
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// This simulates a "parent" struct (interface) with a discriminator along with one "substruct" and their unmarshal functions.
// The Vehicle struct is not defined here because the linter informed me that it wasn't being used :)
type VehicleIntf interface {
	isaVehicle() bool
}

func UnmarshalVehicle(m map[string]json.RawMessage, result interface{}) (err error) {
	var discValue string
	err = UnmarshalPrimitive(m, "vehicle_type", &discValue)
	if err != nil {
		return
	}
	if discValue == "" {
		err = fmt.Errorf("discriminator property 'vehicle_type' not found in JSON object")
		return
	}
	if discValue == "Car" {
		err = UnmarshalCar(m, result)
	} else {
		err = fmt.Errorf("unrecognized value for discriminator property 'vehicle_type': %s", discValue)
	}
	return
}

type Car struct {
	VehicleType *string `json:"vehicle_type" validate:"required"`
	Make        *string `json:"make,omitempty"`
	BodyStyle   *string `json:"body_style,omitempty"`
}

func (*Car) isaVehicle() bool {
	return true
}
func UnmarshalCar(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(Car)
	err = UnmarshalPrimitive(m, "vehicle_type", &obj.VehicleType)
	if err != nil {
		return
	}
	err = UnmarshalPrimitive(m, "make", &obj.Make)
	if err != nil {
		return
	}
	err = UnmarshalPrimitive(m, "body_style", &obj.BodyStyle)
	if err != nil {
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
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
	assert.Nil(t, mySlice)
	assert.Equal(t, 0, len(mySlice))

	// Unmarshal an explicit null value.
	mySlice = nil
	err = UnmarshalModel(rawMap, "null_slice", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Nil(t, mySlice)
	assert.Equal(t, 0, len(mySlice))

	// Unmarshal an empty slice.
	mySlice = nil
	err = UnmarshalModel(rawMap, "empty_slice", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, mySlice)
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
	assert.Nil(t, myMap)
	assert.Equal(t, 0, len(myMap))

	// Unmarshal an explicit null value.
	myMap = nil
	err = UnmarshalModel(rawMap, "null_map", &myMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Nil(t, myMap)
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
	assert.Nil(t, mySliceMap)
	assert.Equal(t, 0, len(mySliceMap))

	// Unmarshal an explicit null value.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "null_prop", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Nil(t, mySliceMap)
	assert.Equal(t, 0, len(mySliceMap))

	// Unmarshal a map with an empty slice.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "empty_slice_map", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(mySliceMap))
	assert.NotNil(t, mySliceMap["empty_slice"])
	assert.Equal(t, 0, len(mySliceMap["empty_slice"]))

	// Unmarshal a map with an explicit null slice. Result should be an empty map.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "null_slice_map", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, mySliceMap)
	assert.Equal(t, 0, len(mySliceMap))
	assert.Nil(t, mySliceMap["null_slice"])
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

func TestUnmarshalModelAbstractInstance(t *testing.T) {
	jsonString := `{ "vehicle_type": "Car", "make": "Ford", "body_style": "coupe"}`

	var err error
	var myVehicle VehicleIntf

	// Unmarshal an instance of the parent (should end up with a Car).
	rawMap := unmarshalMap(jsonString)
	err = UnmarshalModel(rawMap, "", &myVehicle, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.NotNil(t, myVehicle)
	myCar, ok := myVehicle.(*Car)
	assert.True(t, ok)
	assert.NotNil(t, myCar)
	assert.Equal(t, "Car", *(myCar.VehicleType))
	assert.Equal(t, "Ford", *(myCar.Make))
	assert.Equal(t, "coupe", *(myCar.BodyStyle))

	// Unmarshal an instance of the substruct directly.
	myCar = nil
	err = UnmarshalModel(rawMap, "", &myCar, UnmarshalCar)
	assert.Nil(t, err)
	assert.NotNil(t, myCar)
	assert.Equal(t, "Car", *(myCar.VehicleType))
	assert.Equal(t, "Ford", *(myCar.Make))
	assert.Equal(t, "coupe", *(myCar.BodyStyle))
}

func TestUnmarshalModelAbstractSlice(t *testing.T) {
	jsonString := `[{ "vehicle_type": "Car", "make": "Ford", "body_style": "coupe"}]`

	var err error
	var myVehicleSlice []VehicleIntf

	// Unmarshal a slice of Vehicles.
	rawSlice := unmarshalSlice(jsonString)
	err = UnmarshalModel(rawSlice, "", &myVehicleSlice, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myVehicleSlice))
	myCar, ok := myVehicleSlice[0].(*Car)
	assert.True(t, ok)
	assert.NotNil(t, myCar)
	assert.Equal(t, "Car", *(myCar.VehicleType))
	assert.Equal(t, "Ford", *(myCar.Make))
	assert.Equal(t, "coupe", *(myCar.BodyStyle))

	// Unmarshal a slice of Cars directly.
	var myCarSlice []Car
	err = UnmarshalModel(rawSlice, "", &myCarSlice, UnmarshalCar)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myCarSlice))
	assert.Equal(t, "Car", *(myCarSlice[0].VehicleType))
	assert.Equal(t, "Ford", *(myCarSlice[0].Make))
	assert.Equal(t, "coupe", *(myCarSlice[0].BodyStyle))

}

func TestUnmarshalModelAbstractMap(t *testing.T) {
	jsonString := `{ "car1": { "vehicle_type": "Car", "make": "Ford", "body_style": "coupe"} }`

	var err error
	var myVehicleMap map[string]VehicleIntf

	rawMap := unmarshalMap(jsonString)
	err = UnmarshalModel(rawMap, "", &myVehicleMap, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myVehicleMap))
	myCar, ok := myVehicleMap["car1"].(*Car)
	assert.True(t, ok)
	assert.NotNil(t, myCar)
	assert.Equal(t, "Car", *(myCar.VehicleType))
	assert.Equal(t, "Ford", *(myCar.Make))
	assert.Equal(t, "coupe", *(myCar.BodyStyle))

	var myCarMap map[string]Car
	err = UnmarshalModel(rawMap, "", &myCarMap, UnmarshalCar)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myCarMap))
	assert.Equal(t, "Car", *(myCarMap["car1"].VehicleType))
	assert.Equal(t, "Ford", *(myCarMap["car1"].Make))
	assert.Equal(t, "coupe", *(myCarMap["car1"].BodyStyle))
}

func TestUnmarshalModelAbstractSliceMap(t *testing.T) {
	jsonString := `{ "carSlice1": [ { "vehicle_type": "Car", "make": "Ford", "body_style": "coupe"} ] }`

	var err error
	var myVehicleSliceMap map[string][]VehicleIntf

	rawMap := unmarshalMap(jsonString)
	err = UnmarshalModel(rawMap, "", &myVehicleSliceMap, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myVehicleSliceMap))
	myCar, ok := myVehicleSliceMap["carSlice1"][0].(*Car)
	assert.True(t, ok)
	assert.NotNil(t, myCar)
	assert.Equal(t, "Car", *(myCar.VehicleType))
	assert.Equal(t, "Ford", *(myCar.Make))
	assert.Equal(t, "coupe", *(myCar.BodyStyle))

	var myCarSliceMap map[string][]Car
	err = UnmarshalModel(rawMap, "", &myCarSliceMap, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myCarSliceMap))
	assert.Equal(t, 1, len(myCarSliceMap["carSlice1"]))
	assert.Equal(t, "Car", *(myCarSliceMap["carSlice1"][0].VehicleType))
	assert.Equal(t, "Ford", *(myCarSliceMap["carSlice1"][0].Make))
	assert.Equal(t, "coupe", *(myCarSliceMap["carSlice1"][0].BodyStyle))
}

func TestUnmarshalModelErrors(t *testing.T) {
	modelStruct := new(ModelStruct)
	var rawSlice []json.RawMessage
	var rawMap map[string]json.RawMessage
	var err error

	// Supply a slice when a map is expected.
	rawSlice = unmarshalSlice(`[ { "prop": "value" } ]`)
	err = UnmarshalModel(rawSlice, "", &modelStruct.Model, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.Model)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling core.MyModel"))
	t.Logf("[01] Expected error: %s\n", err.Error())

	// Supply an incorrect map.
	rawMap = unmarshalMap(`{ "prop": "value"}`)
	err = UnmarshalModel(rawMap, "prop", &modelStruct.Model, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.ModelSlice)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'prop' as core.MyModel"))
	t.Logf("[02] Expected error: %s\n", err.Error())

	// Supply a map with an incorrect MyModel instance.
	rawMap = unmarshalMap(`{ "foo": "string", "bar": "string" }`)
	err = UnmarshalModel(rawMap, "", &modelStruct.Model, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.ModelSlice)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling core.MyModel"))
	t.Logf("[03] Expected error: %s\n", err.Error())

	// Supply a map when a slice is expected.
	rawMap = unmarshalMap(`{ "prop": "value"}`)
	err = UnmarshalModel(rawMap, "", &modelStruct.ModelSlice, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.ModelSlice)
	assert.True(t, strings.Contains(err.Error(), "expected 'rawInput' to be a []json.RawMessage"))
	t.Logf("[04] Expected error: %s\n", err.Error())

	// Supply an incorrect map.
	rawMap = unmarshalMap(`{ "prop": "value"}`)
	err = UnmarshalModel(rawMap, "prop", &modelStruct.ModelSlice, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.ModelSlice)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'prop' as []core.MyModel"))
	t.Logf("[05] Expected error: %s\n", err.Error())

	// Supply a map with an incorrect MyModel instance.
	rawMap = unmarshalMap(`{ "prop": [ {"foo": 38, "bar": 38} ] }`)
	err = UnmarshalModel(rawMap, "prop", &modelStruct.ModelSlice, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(modelStruct.ModelSlice))
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'prop' as []core.MyModel"))
	t.Logf("[06] Expected error: %s\n", err.Error())

	// Supply a slice when a map is expected.
	rawSlice = unmarshalSlice(`[ { "prop": "value" } ]`)
	err = UnmarshalModel(rawSlice, "prop", &modelStruct.ModelSlice, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(modelStruct.ModelSlice))
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'prop' as []core.MyModel"))
	t.Logf("[07] Expected error: %s\n", err.Error())

	// Supply a slice when a map is expected.
	rawSlice = unmarshalSlice(`[ { "prop": "value" } ]`)
	err = UnmarshalModel(rawSlice, "", &modelStruct.ModelMap, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.ModelMap)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling map[string]core.MyModel"))
	t.Logf("[08] Expected error: %s\n", err.Error())

	// Supply an incorrect map.
	rawMap = unmarshalMap(`{ "prop": "value"}`)
	err = UnmarshalModel(rawMap, "prop", &modelStruct.ModelMap, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.ModelMap)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'prop' as map[string]core.MyModel"))
	t.Logf("[09] Expected error: %s\n", err.Error())

	// Supply a map with an incorrect MyModel instance.
	rawMap = unmarshalMap(`{ "foo1": {"foo": 38, "bar": 38} }`)
	err = UnmarshalModel(rawMap, "", &modelStruct.ModelMap, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(modelStruct.ModelMap))
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling map[string]core.MyModel"))
	t.Logf("[10] Expected error: %s\n", err.Error())

	// Supply a slice when a map is expected.
	rawSlice = unmarshalSlice(`[ { "prop": "value" } ]`)
	err = UnmarshalModel(rawSlice, "", &modelStruct.ModelSliceMap, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.ModelSliceMap)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling map[string][]core.MyModel"))
	t.Logf("[11] Expected error: %s\n", err.Error())

	// Supply an incorrect map.
	rawMap = unmarshalMap(`{ "prop": "value"}`)
	err = UnmarshalModel(rawMap, "prop", &modelStruct.ModelSliceMap, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, modelStruct.ModelSliceMap)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'prop' as map[string][]core.MyModel"))
	t.Logf("[12] Expected error: %s\n", err.Error())
}

func TestUnmarshalModelAbstractErrors(t *testing.T) {
	var err error
	var myVehicle VehicleIntf
	var mySlice []VehicleIntf
	var myMap map[string]VehicleIntf
	var mySliceMap map[string][]VehicleIntf
	var rawMap map[string]json.RawMessage
	var rawSlice []json.RawMessage

	// Not a valid discriminator value.
	rawMap = unmarshalMap(`{ "vehicle_type": "EV", "make": "Ford", "body_style": "Mach-E"}`)
	err = UnmarshalModel(rawMap, "", &myVehicle, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.Nil(t, myVehicle)
	t.Logf("[01] Expected error: %s\n", err.Error())

	// Not a valid discriminator value type.
	rawMap = unmarshalMap(`{ "vehicle_type": 44, "make": "Ford", "body_style": "Mach-E"}`)
	err = UnmarshalModel(rawMap, "", &myVehicle, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.Nil(t, myVehicle)
	t.Logf("[02] Expected error: %s\n", err.Error())

	// Not a valid model instance
	rawMap = unmarshalMap(`{ "vehicle_type": "Car", "make": "Ford", "body_style": 44 }`)
	err = UnmarshalModel(rawMap, "", &myVehicle, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.Nil(t, myVehicle)
	t.Logf("[03] Expected error: %s\n", err.Error())

	// Not a valid model slice.
	rawSlice = unmarshalSlice(`[{ "vehicle_type": "Car", "make": "Ford", "body_style": 44 }]`)
	err = UnmarshalModel(rawSlice, "", &mySlice, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(mySlice))
	t.Logf("[04] Expected error: %s\n", err.Error())

	// Not a valid model slice.
	rawSlice = unmarshalSlice(`[{ "vehicle_type": "EV", "make": "Ford", "body_style": 44 }]`)
	err = UnmarshalModel(rawSlice, "", &mySlice, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(mySlice))
	t.Logf("[05] Expected error: %s\n", err.Error())

	// Not a valid model map.
	rawMap = unmarshalMap(`{ "vehicle1": { "vehicle_type": "EV", "make": "Ford", "body_style": 44 } }`)
	err = UnmarshalModel(rawMap, "", &myMap, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(myMap))
	t.Logf("[06] Expected error: %s\n", err.Error())

	// Not a valid model slice map.
	rawMap = unmarshalMap(`{ "vehicleSlice1": [ { "vehicle_type": "EV", "make": "Ford", "body_style": 44 } ] }`)
	err = UnmarshalModel(rawMap, "", &mySliceMap, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.Equal(t, 0, len(myMap))
	t.Logf("[07] Expected error: %s\n", err.Error())
}

// Utility function that unmarshals a JSON string into
func unmarshalSlice(jsonString string) (result []json.RawMessage) {
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		panic(err)
	}
	return
}
