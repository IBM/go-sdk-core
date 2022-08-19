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
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MyModel simulates a generated model with its unmarshal function.
type MyModel struct {
	Foo *string `json:"foo" validate:"required"`
	Bar *int64  `json:"bar" validate:"required"`
}

// AssertEqual asserts that 'm' is equivalent to 'this'.
func (this MyModel) AssertEqual(t *testing.T, m MyModel) {
	assert.True(t, (this.Foo == nil && m.Foo == nil) || (this.Foo != nil && m.Foo != nil && *this.Foo == *m.Foo))
	assert.True(t, (this.Bar == nil && m.Bar == nil) || (this.Bar != nil && m.Bar != nil && *this.Bar == *m.Bar))
}

// Instances of MyModel used in the tests below.
var myModel1 MyModel = MyModel{Foo: StringPtr("string1"), Bar: Int64Ptr(44)}
var myModel2 MyModel = MyModel{Foo: StringPtr("string2"), Bar: Int64Ptr(74)}
var myModel3 MyModel = MyModel{Foo: StringPtr("string3"), Bar: Int64Ptr(33)}
var myModel4 MyModel = MyModel{Foo: StringPtr("string4"), Bar: Int64Ptr(21)}
var myModelZeroValue MyModel

// Simulated "generated" unmarshal function for MyModel struct.
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

// ModelStruct simulates a generated model with properties that involve models, along with its unmarshal function.
type ModelStruct struct {
	Model           *MyModel             `json:"model" validate:"required"`
	ModelSlice      []MyModel            `json:"model_slice" validate:"required"`
	ModelMap        map[string]MyModel   `json:"model_map" validate:"required"`
	ModelSliceMap   map[string][]MyModel `json:"model_slice_map" validate:"required"`
	ModelSliceSlice [][]MyModel          `json:"model_slice_slice" validate:"required"`
}

// Simulated "generated" unmarshal function for ModelStruct struct.
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
	err = UnmarshalModel(m, "model_slice_slice", &obj.ModelSliceSlice, UnmarshalMyModel)
	if err != nil {
		return
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

// VehicleIntf simulates a "parent" struct (interface) with a discriminator along with one "substruct" and their unmarshal functions.
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

// Car serves as a discriminated subclass (substruct?) of Vehicle.
type Car struct {
	VehicleType *string `json:"vehicle_type" validate:"required"`
	Make        *string `json:"make,omitempty"`
	BodyStyle   *string `json:"body_style,omitempty"`
}

// AssertEqual asserts that 'c' is equivalent to 'this'.
func (this Car) AssertEqual(t *testing.T, c Car) {
	assert.True(t, (this.VehicleType == nil && c.VehicleType == nil) || (this.VehicleType != nil && c.VehicleType != nil && *this.VehicleType == *c.VehicleType))
	assert.True(t, (this.Make == nil && c.Make == nil) || (this.Make != nil && c.Make != nil && *this.Make == *c.Make))
	assert.True(t, (this.BodyStyle == nil && c.BodyStyle == nil) || (this.BodyStyle != nil && c.BodyStyle != nil && *this.BodyStyle == *c.BodyStyle))
}

// Instance of Car used in tests below.
var car1 Car = Car{VehicleType: StringPtr("Car"), Make: StringPtr("Ford"), BodyStyle: StringPtr("Coupe")}

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

//
// Test methods.
//

func TestUnmarshalModelInstanceNil(t *testing.T) {
	var err error
	var myModel *MyModel

	jsonString := `{ 
		"null_model": null,
		"empty_model": {}
	}`
	rawMap := unmarshalMap(jsonString)

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
	myModelZeroValue.AssertEqual(t, *myModel)
}

func TestUnmarshalModelInstance(t *testing.T) {
	var err error
	var actualModel *MyModel

	jsonString := toJSON(myModel1)
	rawMap := unmarshalMap(jsonString)

	// Unmarshal an instance of the model.
	err = UnmarshalModel(rawMap, "", &actualModel, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, actualModel)
	myModel1.AssertEqual(t, *actualModel)

	// Unmarshal with "zero" input should return an error
	var zeroRawMap map[string]json.RawMessage
	actualModel = nil
	err = UnmarshalModel(zeroRawMap, "", &actualModel, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, actualModel)
}

func TestUnmarshalModelSliceNil(t *testing.T) {
	var err error
	var mySlice []MyModel

	jsonTemplate := `{ 
		"null_slice": null,
		"empty_slice": [],
		"slice_with_null": [ null, %m1, null ]
	}`
	jsonString := strings.ReplaceAll(jsonTemplate, "%m1", toJSON(myModel1))
	rawMap := unmarshalMap(jsonString)

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

	// Unmarshal a slice with a null value.
	mySlice = nil
	err = UnmarshalModel(rawMap, "slice_with_null", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, mySlice)
	assert.Equal(t, 3, len(mySlice))
	myModelZeroValue.AssertEqual(t, mySlice[0])
	myModel1.AssertEqual(t, mySlice[1])
	myModelZeroValue.AssertEqual(t, mySlice[2])
}

func TestUnmarshalModelSlice(t *testing.T) {
	var err error
	var mySlice []MyModel

	jsonTemplate := `[ %m1, %m2 ]`
	jsonString := strings.ReplaceAll(jsonTemplate, "%m1", toJSON(myModel1))
	jsonString = strings.ReplaceAll(jsonString, "%m2", toJSON(myModel2))
	rawSlice := unmarshalSlice(jsonString)

	// Unmarshal a model slice.
	err = UnmarshalModel(rawSlice, "", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(mySlice))
	myModel1.AssertEqual(t, mySlice[0])
	myModel2.AssertEqual(t, mySlice[1])

	// Unmarshal with "zero" input should return an error
	var zeroSlice []json.RawMessage
	mySlice = nil
	err = UnmarshalModel(zeroSlice, "", &mySlice, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.Nil(t, mySlice)
}

func TestUnmarshalModelSliceSliceNil(t *testing.T) {
	var err error
	var mySlice [][]MyModel

	jsonTemplate := `{ 
		"null_slice": null,
		"empty_slice": [],
		"slice_with_null1": [ null ],
		"slice_with_null2": [ null, [ null, %m1, null ], null ]
	}`
	jsonString := strings.ReplaceAll(jsonTemplate, "%m1", toJSON(myModel1))
	rawMap := unmarshalMap(jsonString)

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

	// Unmarshal slice with "null" inner slice.
	mySlice = nil
	err = UnmarshalModel(rawMap, "slice_with_null1", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, mySlice)
	assert.Equal(t, 1, len(mySlice))
	assert.Equal(t, 0, len(mySlice[0]))

	// Unmarshal slice with nulls within the inner slices.
	mySlice = nil
	err = UnmarshalModel(rawMap, "slice_with_null2", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, mySlice)

	// Outer slice should have 3 elements.
	assert.Equal(t, 3, len(mySlice))

	// The first inner slice should be empty.
	assert.Equal(t, 0, len(mySlice[0]))

	// The second inner slice should have 3 elements.
	assert.Equal(t, 3, len(mySlice[1]))
	myModelZeroValue.AssertEqual(t, mySlice[1][0])
	myModel1.AssertEqual(t, mySlice[1][1])
	myModelZeroValue.AssertEqual(t, mySlice[1][2])

	// The third inner slice should be empty.
	assert.Equal(t, 0, len(mySlice[2]))
}

func TestUnmarshalModelSliceSlice(t *testing.T) {
	var err error
	var mySlice [][]MyModel

	jsonTemplate := `[ [ %m1, %m2 ], [ %m3, %m4, %m2, %m1], [ %m4, {} ] ]`
	jsonString := strings.ReplaceAll(jsonTemplate, "%m1", toJSON(myModel1))
	jsonString = strings.ReplaceAll(jsonString, "%m2", toJSON(myModel2))
	jsonString = strings.ReplaceAll(jsonString, "%m3", toJSON(myModel3))
	jsonString = strings.ReplaceAll(jsonString, "%m4", toJSON(myModel4))
	rawSlice := unmarshalSlice(jsonString)

	// Unmarshal the slice of model slices.
	err = UnmarshalModel(rawSlice, "", &mySlice, UnmarshalMyModel)
	assert.Nil(t, err)

	// Outer slice should have 3 elements.
	assert.Equal(t, 3, len(mySlice))

	// The first inner slice should have 2 elements.
	assert.Equal(t, 2, len(mySlice[0]))
	myModel1.AssertEqual(t, mySlice[0][0])
	myModel2.AssertEqual(t, mySlice[0][1])

	// The second inner slice should have 4 elements.
	assert.Equal(t, 4, len(mySlice[1]))
	myModel3.AssertEqual(t, mySlice[1][0])
	myModel4.AssertEqual(t, mySlice[1][1])
	myModel2.AssertEqual(t, mySlice[1][2])
	myModel1.AssertEqual(t, mySlice[1][3])

	// The third inner slice should ahve 2 elements.
	assert.Equal(t, 2, len(mySlice[2]))
	myModel4.AssertEqual(t, mySlice[2][0])
	myModelZeroValue.AssertEqual(t, mySlice[2][1])
}

func TestUnmarshalModelMapNil(t *testing.T) {
	var err error
	var myMap map[string]MyModel

	jsonString := `{ 
		"null_map": null,
		"empty_map": {},
		"map_with_null": { 
			"model1": null 
		}
	}`
	rawMap := unmarshalMap(jsonString)

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

	// Unmarshal a map with a null entry.
	myMap = nil
	err = UnmarshalModel(rawMap, "map_with_null", &myMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, myMap)
	assert.Equal(t, 1, len(myMap))
	assert.NotNil(t, myMap["model1"])
	myModelZeroValue.AssertEqual(t, myMap["model1"])
}

func TestUnmarshalModelMap(t *testing.T) {
	var err error
	var myMap map[string]MyModel

	jsonTemplate := `{
		"model1": %m1, 
		"model2": %m2
	}`
	jsonString := strings.ReplaceAll(jsonTemplate, "%m1", toJSON(myModel1))
	jsonString = strings.ReplaceAll(jsonString, "%m2", toJSON(myModel2))
	rawMap := unmarshalMap(jsonString)

	// Unmarshal a map of model instances.
	err = UnmarshalModel(rawMap, "", &myMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, myMap)
	assert.Equal(t, 2, len(myMap))

	assert.NotNil(t, myMap["model1"])
	myModel1.AssertEqual(t, myMap["model1"])

	assert.NotNil(t, myMap["model2"])
	myModel2.AssertEqual(t, myMap["model2"])

	_, foundIt := myMap["bad_key"]
	assert.False(t, foundIt)
}

func TestUnmarshalModelSliceMapNil(t *testing.T) {
	var err error
	var mySliceMap map[string][]MyModel

	jsonString := `{ 
		"null_prop": null, 
		"empty_slice_map": { 
			"empty_slice": [] 
		},
		"map_with_null1": { 
			"null_slice": null
		},
		"map_with_null2": {
			"slice_with_null": [ null ]
		}

	}`
	rawMap := unmarshalMap(jsonString)

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

	// Unmarshal a map with a "null" slice.
	// Result should be an empty map.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "map_with_null1", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, mySliceMap)
	assert.Equal(t, 0, len(mySliceMap))

	// Unmarshal a map with a slice that contains a "null" element.
	// Result should be a map with a slice containing the model's zero value.
	mySliceMap = nil
	err = UnmarshalModel(rawMap, "map_with_null2", &mySliceMap, UnmarshalMyModel)
	assert.Nil(t, err)
	assert.NotNil(t, mySliceMap)
	assert.Equal(t, 1, len(mySliceMap))

	assert.NotNil(t, mySliceMap["slice_with_null"])
	assert.Equal(t, 1, len(mySliceMap["slice_with_null"]))
	myModelZeroValue.AssertEqual(t, mySliceMap["slice_with_null"][0])
}

func TestUnmarshalModelStruct(t *testing.T) {
	var err error
	var modelStruct *ModelStruct

	jsonTemplate := `{
		"model": %m1,
		"model_slice": [ %m1, %m2 ],
		"model_map": {
			"model1": %m1,
			"model2": %m2
		},
		"model_slice_map": {
			"slice1": [ %m1, %m2, %m3 ],
			"slice2": [ %m4 ]
		},
		"model_slice_slice": [ [ %m1, %m2 ], [ %m3 ], [ %m4, %m1 ]]
	}`
	jsonString := strings.ReplaceAll(jsonTemplate, "%m1", toJSON(myModel1))
	jsonString = strings.ReplaceAll(jsonString, "%m2", toJSON(myModel2))
	jsonString = strings.ReplaceAll(jsonString, "%m3", toJSON(myModel3))
	jsonString = strings.ReplaceAll(jsonString, "%m4", toJSON(myModel4))
	rawMap := unmarshalMap(jsonString)

	// Unmarshal the entire ModelStruct instance.
	err = UnmarshalModel(rawMap, "", &modelStruct, UnmarshalModelStruct)
	assert.Nil(t, err)
	assert.NotNil(t, modelStruct)

	// Verify model instance.
	assert.NotNil(t, modelStruct.Model)
	myModel1.AssertEqual(t, *modelStruct.Model)

	// Verify model slice.
	assert.Equal(t, 2, len(modelStruct.ModelSlice))
	myModel1.AssertEqual(t, modelStruct.ModelSlice[0])
	myModel2.AssertEqual(t, modelStruct.ModelSlice[1])

	// Verify model map.
	assert.Equal(t, 2, len(modelStruct.ModelMap))
	assert.NotNil(t, modelStruct.ModelMap["model1"])
	myModel1.AssertEqual(t, modelStruct.ModelMap["model1"])
	assert.NotNil(t, modelStruct.ModelMap["model2"])
	myModel2.AssertEqual(t, modelStruct.ModelMap["model2"])
	_, foundIt := modelStruct.ModelMap["bad_key"]
	assert.False(t, foundIt)

	// Verify model slice map (map of model slices).
	assert.Equal(t, 2, len(modelStruct.ModelSliceMap))
	assert.NotNil(t, modelStruct.ModelSliceMap["slice1"])
	assert.Equal(t, 3, len(modelStruct.ModelSliceMap["slice1"]))
	myModel1.AssertEqual(t, modelStruct.ModelSliceMap["slice1"][0])
	myModel2.AssertEqual(t, modelStruct.ModelSliceMap["slice1"][1])
	myModel3.AssertEqual(t, modelStruct.ModelSliceMap["slice1"][2])

	assert.NotNil(t, modelStruct.ModelSliceMap["slice2"])
	assert.Equal(t, 1, len(modelStruct.ModelSliceMap["slice2"]))
	myModel4.AssertEqual(t, modelStruct.ModelSliceMap["slice2"][0])

	// Verify slice of model slices (two-dimensional array of model instances).
	assert.Equal(t, 3, len(modelStruct.ModelSliceSlice))

	assert.Equal(t, 2, len(modelStruct.ModelSliceSlice[0]))
	myModel1.AssertEqual(t, modelStruct.ModelSliceSlice[0][0])
	myModel2.AssertEqual(t, modelStruct.ModelSliceSlice[0][1])

	assert.Equal(t, 1, len(modelStruct.ModelSliceSlice[1]))
	myModel3.AssertEqual(t, modelStruct.ModelSliceSlice[1][0])

	assert.Equal(t, 2, len(modelStruct.ModelSliceSlice[2]))
	myModel4.AssertEqual(t, modelStruct.ModelSliceSlice[2][0])
	myModel1.AssertEqual(t, modelStruct.ModelSliceSlice[2][1])
}

func TestUnmarshalModelAbstractInstance(t *testing.T) {
	var err error
	var myVehicle VehicleIntf

	jsonString := toJSON(car1)
	rawMap := unmarshalMap(jsonString)

	// Unmarshal an instance of the parent (should end up with a Car).
	err = UnmarshalModel(rawMap, "", &myVehicle, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.NotNil(t, myVehicle)

	myCar, ok := myVehicle.(*Car)
	assert.True(t, ok)
	assert.NotNil(t, myCar)
	car1.AssertEqual(t, *myCar)

	// Unmarshal an instance of the substruct directly.
	myCar = nil
	err = UnmarshalModel(rawMap, "", &myCar, UnmarshalCar)
	assert.Nil(t, err)
	assert.NotNil(t, myCar)
	car1.AssertEqual(t, *myCar)
}

func TestUnmarshalModelAbstractSlice(t *testing.T) {
	var err error

	jsonTemplate := `[ %c1 ]`
	jsonString := strings.ReplaceAll(jsonTemplate, "%c1", toJSON(car1))
	rawSlice := unmarshalSlice(jsonString)

	// Unmarshal a slice of Vehicles.
	var myVehicleSlice []VehicleIntf
	err = UnmarshalModel(rawSlice, "", &myVehicleSlice, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myVehicleSlice))
	myCar, ok := myVehicleSlice[0].(*Car)
	assert.True(t, ok)
	assert.NotNil(t, myCar)
	car1.AssertEqual(t, *myCar)

	// Unmarshal a slice of Cars directly.
	var myCarSlice []Car
	err = UnmarshalModel(rawSlice, "", &myCarSlice, UnmarshalCar)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myCarSlice))
	car1.AssertEqual(t, myCarSlice[0])
}

func TestUnmarshalModelAbstractMap(t *testing.T) {
	var err error

	jsonTemplate := `{ "car1": %c1 }`
	jsonString := strings.ReplaceAll(jsonTemplate, "%c1", toJSON(car1))
	rawMap := unmarshalMap(jsonString)

	var myVehicleMap map[string]VehicleIntf
	err = UnmarshalModel(rawMap, "", &myVehicleMap, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myVehicleMap))
	myCar, ok := myVehicleMap["car1"].(*Car)
	assert.True(t, ok)
	assert.NotNil(t, myCar)
	car1.AssertEqual(t, *myCar)

	var myCarMap map[string]Car
	err = UnmarshalModel(rawMap, "", &myCarMap, UnmarshalCar)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myCarMap))
	_, foundIt := myCarMap["car1"]
	assert.True(t, foundIt)
	car1.AssertEqual(t, myCarMap["car1"])
}

func TestUnmarshalModelAbstractSliceMap(t *testing.T) {
	var err error

	jsonTemplate := `{ "carSlice1": [ %c1 ] }`
	jsonString := strings.ReplaceAll(jsonTemplate, "%c1", toJSON(car1))
	rawMap := unmarshalMap(jsonString)

	var myVehicleSliceMap map[string][]VehicleIntf
	err = UnmarshalModel(rawMap, "", &myVehicleSliceMap, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myVehicleSliceMap))
	myCar, ok := myVehicleSliceMap["carSlice1"][0].(*Car)
	assert.True(t, ok)
	assert.NotNil(t, myCar)
	car1.AssertEqual(t, *myCar)

	var myCarSliceMap map[string][]Car
	err = UnmarshalModel(rawMap, "", &myCarSliceMap, UnmarshalVehicle)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(myCarSliceMap))
	assert.Equal(t, 1, len(myCarSliceMap["carSlice1"]))
	car1.AssertEqual(t, myCarSliceMap["carSlice1"][0])
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

	// Supply a model slice when a slice of model slices is expected.
	rawSlice = unmarshalSlice(`[ {"foo": "string", "bar": 44} ]`)
	err = UnmarshalModel(rawSlice, "", &modelStruct.ModelSliceSlice, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling [][]core.MyModel"))
	t.Logf("[13] Expected error: %s\n", err.Error())

	// Supply a map of model slices when a map containing a slice of model slices is expected.
	rawMap = unmarshalMap(`{ "prop": [ {"foo": "string", "bar": 44} ] }`)
	err = UnmarshalModel(rawMap, "prop", &modelStruct.ModelSliceSlice, UnmarshalMyModel)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'prop' as [][]core.MyModel"))
	t.Logf("[14] Expected error: %s\n", err.Error())
}

func TestUnmarshalModelAbstractErrors(t *testing.T) {
	var err error
	var myVehicle VehicleIntf
	var mySlice []VehicleIntf
	var myMap map[string]VehicleIntf
	var mySliceMap map[string][]VehicleIntf
	var mySliceSlice [][]VehicleIntf
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

	// Supply a model slice when a slice of model slices is expected.
	rawSlice = unmarshalSlice(`[ { "vehicle_type": "EV", "make": "Ford", "body_style": 44 } ]`)
	err = UnmarshalModel(rawSlice, "", &mySliceSlice, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling [][]core.VehicleIntf"))
	t.Logf("[08] Expected error: %s\n", err.Error())

	// Supply a map of model slices when a map containing a slice of model slices is expected.
	rawMap = unmarshalMap(`{ "prop": [ { "vehicle_type": "EV", "make": "Ford", "body_style": 44 } ] }`)
	err = UnmarshalModel(rawMap, "prop", &mySliceSlice, UnmarshalVehicle)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "error unmarshalling property 'prop' as [][]core.VehicleIntf"))
	t.Logf("[09] Expected error: %s\n", err.Error())
}

// Utility function that unmarshals a JSON string into
func unmarshalSlice(jsonString string) (result []json.RawMessage) {
	err := json.Unmarshal([]byte(jsonString), &result)
	if err != nil {
		panic(err)
	}
	return
}
