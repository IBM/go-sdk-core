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
	"github.com/go-yaml/yaml"
)

type OrderedMaps struct {
	maps []yaml.MapItem
}

func (m *OrderedMaps) Add(key string, value interface{}) {
	m.maps = append(m.maps, yaml.MapItem{
		Key: key,
		Value: value,
	})
}

func (m *OrderedMaps) GetMaps() []yaml.MapItem {
	return m.maps
}

func NewOrderedMaps() *OrderedMaps {
	return &OrderedMaps{}
}
