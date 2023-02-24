//go:build all || fast
// +build all fast

package core

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

//
// The purpose of this testcase is to ensure that dynamic properties with nil values are
// correctly serialized as JSON null values.
// In this testcase we have a struct that simulates a generated model with additional properties
// of type string (actuall a *string).
// In addition to the struct, we have methods SetProperty() and GetProperty() which would normally
// be generated for a dynamic model.
// And to round out the simulation, we have methods MarshalJSON() and unmarshalDynamicModel() which
// also simulate methods that are generated for a dynamic model.
// Note that if the SDK generator is modified to change the way in which any of these methods are generated,
// this testcase can be updated to reflect the new generated code and the testcase can continue to
// serve as a test of serialize null dynamic property values.

type dynamicModel struct {
	Prop1                *string `json:"prop1,omitempty"`
	Prop2                *int64  `json:"prop2,omitempty"`
	additionalProperties map[string]*string
}

func (o *dynamicModel) SetProperty(key string, value *string) {
	if o.additionalProperties == nil {
		o.additionalProperties = make(map[string]*string)
	}
	o.additionalProperties[key] = value
}

func (o *dynamicModel) GetProperty(key string) *string {
	return o.additionalProperties[key]
}

func (o *dynamicModel) MarshalJSON() (buffer []byte, err error) {
	m := make(map[string]interface{})
	if len(o.additionalProperties) > 0 {
		for k, v := range o.additionalProperties {
			m[k] = v
		}
	}
	if o.Prop1 != nil {
		m["prop1"] = o.Prop1
	}
	if o.Prop2 != nil {
		m["prop2"] = o.Prop2
	}
	buffer, err = json.Marshal(m)
	return
}

func unmarshalDynamicModel(m map[string]json.RawMessage, result interface{}) (err error) {
	obj := new(dynamicModel)
	err = UnmarshalPrimitive(m, "prop1", &obj.Prop1)
	if err != nil {
		return
	}
	delete(m, "prop1")
	err = UnmarshalPrimitive(m, "prop2", &obj.Prop2)
	if err != nil {
		return
	}
	delete(m, "prop2")
	for k := range m {
		var v *string
		e := UnmarshalPrimitive(m, k, &v)
		if e != nil {
			err = e
			return
		}
		obj.SetProperty(k, v)
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj))
	return
}

func TestAdditionalPropertiesNull(t *testing.T) {
	// Construct an instance of the model so that it includes a dynamic property with value nil.
	model := &dynamicModel{
		Prop1: StringPtr("foo"),
		Prop2: Int64Ptr(38),
	}
	model.SetProperty("bar", nil)

	// Serialize to JSON and ensure that the nil dynamic property value was explicitly serialized as JSON null.
	b, err := json.Marshal(model)
	jsonString := string(b)
	assert.Nil(t, err)
	t.Logf("Serialized model: %s\n", jsonString)
	assert.Contains(t, jsonString, `"bar":null`)

	// Next, deserialize the json string into a map of RawMessages to simulate how the SDK code will
	// deserialize a response body.
	var rawMap map[string]json.RawMessage
	err = json.NewDecoder(bytes.NewReader(b)).Decode(&rawMap)
	assert.Nil(t, err)
	assert.NotNil(t, rawMap)

	// Use the "generated" unmarshalDynamicModel function to unmarshal the raw map into a model instance.
	var newModel *dynamicModel
	err = unmarshalDynamicModel(rawMap, &newModel)
	assert.Nil(t, err)
	assert.NotNil(t, newModel)
	t.Logf("newModel: %+v\n", *newModel)

	// Make sure the new model is the same as the original model.
	assert.Equal(t, model, newModel)
}
