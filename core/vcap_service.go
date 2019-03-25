package core

/**
 * Copyright 2019 IBM All Rights Reserved.
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
	"os"
)

type Service struct {
	Credentials Credential `json:"credentials,omitempty"`
}

type Credential struct {
	URL      string `json:"url,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	APIKey   string `json:"apikey,omitempty"`
}

func LoadFromVCAPServices(serviceName string) *Credential {
	vcapServices := os.Getenv("VCAP_SERVICES")
	if vcapServices != "" {
		var rawServices map[string][]Service
		if vcapServices != "" {
			if err := json.Unmarshal([]byte(vcapServices), &rawServices); err != nil {
				return nil
			}
			for name, instances := range rawServices {
				if name == serviceName {
					creds := &instances[0].Credentials
					return creds
				}
			}
		}
	}
	return nil
}
