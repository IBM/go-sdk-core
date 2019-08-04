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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// common constants for core
const (
	USER_AGENT    = "User-Agent"
	SDK_NAME      = "ibm-go-sdk-core"
	UNKNOWN_ERROR = "Unknown Error"
)

// ServiceOptions Service options
type ServiceOptions struct {
	Version       string
	URL           string
	Authenticator Authenticator
}

// BaseService Base Service
type BaseService struct {
	Options        *ServiceOptions
	DefaultHeaders http.Header
	Client         *http.Client
	UserAgent      string
}

type CredentialProps map[string]string

// NewBaseService Instantiate a Base Service
func NewBaseService(options *ServiceOptions, serviceName, displayName string) (*BaseService, error) {
	if HasBadFirstOrLastChar(options.URL) {
		return nil, fmt.Errorf(ERRORMSG_PROP_INVALID, "URL")
	}

	if options.Authenticator == nil {
		return nil, fmt.Errorf(ERRORMSG_NO_AUTHENTICATOR)
	}

	if err := options.Authenticator.Validate(); err != nil {
		return nil, err
	}

	service := BaseService{
		Options: options,

		Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	// Set a default value for the User-Agent http header.
	service.SetUserAgent(service.BuildUserAgent())

	// TODO: try to load service properties from external config (url, disable-ssl).

	return &service, nil
}

// SetURL sets the service URL
func (service *BaseService) SetURL(url string) error {
	if HasBadFirstOrLastChar(url) {
		return fmt.Errorf(ERRORMSG_PROP_INVALID, "URL")
	}

	service.Options.URL = url
	return nil
}

// SetDefaultHeaders sets HTTP headers to be sent in every request.
func (service *BaseService) SetDefaultHeaders(headers http.Header) {
	service.DefaultHeaders = headers
}

// SetHTTPClient updates the client handling the requests
func (service *BaseService) SetHTTPClient(client *http.Client) {
	service.Client = client
}

// DisableSSLVerification skips SSL verification
func (service *BaseService) DisableSSLVerification() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	service.Client.Transport = tr
}

// BuildUserAgent : Builds the user agent string
func (service *BaseService) BuildUserAgent() string {
	return fmt.Sprintf("%s-%s %s", SDK_NAME, __VERSION__, SystemInfo())
}

// SetUserAgent : Sets the user agent value
func (service *BaseService) SetUserAgent(userAgentString string) {
	if userAgentString == "" {
		service.UserAgent = service.BuildUserAgent()
	}
	service.UserAgent = userAgentString
}

// Request performs the HTTP request
func (service *BaseService) Request(req *http.Request, result interface{}) (*DetailedResponse, error) {
	// Add default headers
	if service.DefaultHeaders != nil {
		for k, v := range service.DefaultHeaders {
			req.Header.Add(k, strings.Join(v, ""))
		}
	}

	// Check if user agent is present.
	userAgent := req.Header.Get(USER_AGENT)
	if userAgent == "" {
		req.Header.Add(USER_AGENT, service.UserAgent)
	}

	// Add authentication to the outbound request.
	if service.Options.Authenticator == nil {
		return nil, fmt.Errorf(ERRORMSG_NO_AUTHENTICATOR)
	}
	err := service.Options.Authenticator.Authenticate(req)
	if err != nil {
		return nil, err
	}

	// Perform the request.
	resp, err := service.Client.Do(req)
	if err != nil {
		return nil, err
	}

	response := new(DetailedResponse)
	response.Headers = resp.Header
	response.StatusCode = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp != nil {
			response.Result = resp
			message := getErrorMessage(resp)
			return response, fmt.Errorf(message)
		}
	}

	contentType := resp.Header.Get(CONTENT_TYPE)
	if contentType != "" {
		if IsJSONMimeType(contentType) && result != nil {
			json.NewDecoder(resp.Body).Decode(&result)
			response.Result = result
			defer resp.Body.Close()
		}
	}

	if response.Result == nil && result != nil {
		response.Result = resp.Body
	}

	return response, nil
}

// Errors : a struct for errors array
type Errors struct {
	Errors []Error `json:"errors,omitempty"`
}

// Error : specifies the error
type Error struct {
	Message string `json:"message,omitempty"`
}

func getErrorMessage(response *http.Response) string {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return UNKNOWN_ERROR
	}

	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return string(body)
	}

	if _, ok := data["errors"]; ok {
		var errors Errors
		json.Unmarshal(body, &errors)
		return errors.Errors[0].Message
	}

	if val, ok := data["error"]; ok {
		return val.(string)
	}

	if val, ok := data["message"]; ok {
		return val.(string)
	}

	if val, ok := data["errorMessage"]; ok {
		return val.(string)
	}

	return UNKNOWN_ERROR
}
