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
	"fmt"
	"net/http"
)

// DetailedResponse : Generic response for IBM API
type DetailedResponse struct {
	StatusCode int         // HTTP status code
	Headers    http.Header // HTTP response headers
	Result     interface{} // response from service
}

// GetHeaders returns the headers
func (response *DetailedResponse) GetHeaders() http.Header {
	return response.Headers
}

// GetStatusCode returns the HTTP status code
func (response *DetailedResponse) GetStatusCode() int {
	return response.StatusCode
}

// GetResult returns the result from the service
func (response *DetailedResponse) GetResult() interface{} {
	return response.Result
}

func (response *DetailedResponse) String() string {
	output, err := json.MarshalIndent(response, "", "    ")
	if err == nil {
		return fmt.Sprintf("%+v\n", string(output))
	}
	return fmt.Sprintf("Response")
}
