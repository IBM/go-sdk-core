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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOrderedMaps(t *testing.T) {
	om := NewOrderedMaps()
	assert.NotNil(t, om)
	assert.Equal(t, 0, len(om.maps))
}

func TestOrderedMapsAdd(t *testing.T) {
	om := &OrderedMaps{}
	assert.Equal(t, 0, len(om.maps))

	om.Add("key", "value")
	assert.Equal(t, 1, len(om.maps))
	assert.Equal(t, om.maps[0].Key, "key")
	assert.Equal(t, om.maps[0].Value, "value")
}

func TestOrderedMapsGetMaps(t *testing.T) {
	om := &OrderedMaps{}
	om.Add("key", "value")

	maps := om.GetMaps()
	assert.Equal(t, 1, len(maps))
	assert.Equal(t, maps[0].Key, "key")
	assert.Equal(t, maps[0].Value, "value")
}
