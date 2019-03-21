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
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// common constants for core
const (
	API_KEY                      = "apikey"
	ICP_PREFIX                   = "icp-"
	USER_AGENT                   = "User-Agent"
	AUTHORIZATION                = "Authorization"
	BEARER                       = "Bearer"
	IBM_CREDENTIAL_FILE_ENV      = "IBM_CREDENTIALS_FILE"
	DEFAULT_CREDENTIAL_FILE_NAME = "ibm-credentials.env"
	URL                          = "url"
	USERNAME                     = "username"
	PASSWORD                     = "password"
	IAM_API_KEY                  = "iam_apikey"
	IAM_URL                      = "iam_url"
	SDK_NAME                     = "ibm-go-sdk-core"
)

// ServiceOptions Service options
type ServiceOptions struct {
	Version        string
	URL            string
	Username       string
	Password       string
	IAMApiKey      string
	IAMAccessToken string
	IAMURL         string
}

// WatsonService Base Service
type BaseService struct {
	Options        *ServiceOptions
	DefaultHeaders http.Header
	TokenManager   *TokenManager
	Client         *http.Client
	UserAgent      string
}

// NewBaseService Instantiate a Watson Service
func NewBaseService(options *ServiceOptions, serviceName, displayName string) (*BaseService, error) {
	if HasBadFirstOrLastChar(options.URL) {
		return nil, fmt.Errorf("The URL shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your URL")
	}

	service := BaseService{
		Options: options,

		Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	service.SetUserAgent(service.BuildUserAgent())

	// 1. Credentials are passed in constructor
	if options.Username != "" && options.Password != "" {
		if options.Username == API_KEY && !strings.HasPrefix(options.Password, ICP_PREFIX) {
			if err := service.SetTokenManager(options.Password, options.IAMAccessToken, options.IAMURL); err != nil {
				return nil, err
			}
		} else {
			if err := service.SetUsernameAndPassword(options.Username, options.Password); err != nil {
				return nil, err
			}
		}
	} else if options.IAMAccessToken != "" || options.IAMApiKey != "" {
		if options.IAMApiKey != "" && strings.HasPrefix(options.IAMApiKey, ICP_PREFIX) {
			if err := service.SetUsernameAndPassword(API_KEY, options.IAMApiKey); err != nil {
				return nil, err
			}
		} else {
			if err := service.SetTokenManager(options.IAMApiKey, options.IAMAccessToken, options.IAMURL); err != nil {
				return nil, err
			}
		}
	}

	// 2. Credentials from credential file
	if displayName != "" && service.Options.Username == "" && service.TokenManager == nil {
		serviceName := strings.ToLower(strings.Replace(displayName, " ", "_", -1))
		service.loadFromCredentialFile(serviceName, "=")
	}

	// 3. Try accessing VCAP_SERVICES env variable
	if service.Options.Username == "" && service.TokenManager == nil {
		credential := LoadFromVCAPServices(serviceName)
		if credential != nil {
			if credential.URL != "" {
				service.SetURL(credential.URL)
			}

			if credential.APIKey != "" {
				service.SetTokenManager(credential.APIKey, "", "")
			} else if credential.Username != "" && credential.Password != "" {
				service.SetUsernameAndPassword(credential.Username, credential.Password)
			}
		}

		if service.Options.Username == "" && service.TokenManager == nil {
			return nil, fmt.Errorf("you must specify an IAM API key or username and password service credentials")
		}
	}

	return &service, nil
}

// SetUsernameAndPassword Sets the Username and Password
func (service *BaseService) SetUsernameAndPassword(username string, password string) error {
	if HasBadFirstOrLastChar(username) {
		return fmt.Errorf("The username shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your username")
	}
	if HasBadFirstOrLastChar(password) {
		return fmt.Errorf("The password shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your password")
	}
	service.Options.Username = username
	service.Options.Password = password
	return nil
}

// SetTokenManager Sets the Token Manager for IAM Authentication
func (service *BaseService) SetTokenManager(iamAPIKey string, iamAccessToken string, iamURL string) error {
	if HasBadFirstOrLastChar(iamAPIKey) {
		return fmt.Errorf("The credentials shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your credentials")
	}
	service.Options.IAMApiKey = iamAPIKey
	service.Options.IAMAccessToken = iamAccessToken
	service.Options.IAMURL = iamURL
	tokenManager := NewTokenManager(iamAPIKey, iamURL, iamAccessToken)
	service.TokenManager = tokenManager
	return nil
}

// SetIAMAccessToken Sets the IAM access token
func (service *BaseService) SetIAMAccessToken(iamAccessToken string) {
	if service.TokenManager != nil {
		service.TokenManager.SetAccessToken(iamAccessToken)
	} else {
		tokenManager := NewTokenManager("", "", iamAccessToken)
		service.TokenManager = tokenManager
	}
	service.Options.IAMAccessToken = iamAccessToken
}

// SetIAMAPIKey Sets the IAM API key
func (service *BaseService) SetIAMAPIKey(iamAPIKey string) error {
	if HasBadFirstOrLastChar(iamAPIKey) {
		return fmt.Errorf("The credentials shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your credentials")
	}
	if service.TokenManager != nil {
		service.TokenManager.SetIAMAPIKey(iamAPIKey)
	} else {
		tokenManager := NewTokenManager(iamAPIKey, "", "")
		service.TokenManager = tokenManager
	}
	service.Options.IAMApiKey = iamAPIKey
	return nil
}

// SetURL sets the service URL
func (service *BaseService) SetURL(url string) error {
	if HasBadFirstOrLastChar(url) {
		return fmt.Errorf("The URL shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your URL")
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

	// Check if user agent is present
	userAgent := req.Header.Get(USER_AGENT)
	if userAgent == "" {
		req.Header.Add(USER_AGENT, service.UserAgent)
	}

	// Add authentication
	if service.TokenManager != nil {
		token, _ := service.TokenManager.GetToken()
		req.Header.Add(AUTHORIZATION, fmt.Sprintf(`%s %s`, BEARER, token))
	} else if service.Options.Username != "" && service.Options.Password != "" {
		req.SetBasicAuth(service.Options.Username, service.Options.Password)
	}

	// Perform the request
	resp, err := service.Client.Do(req)
	if err != nil {
		return nil, err
	}

	// handle the response
	response := new(DetailedResponse)
	response.Headers = resp.Header
	response.StatusCode = resp.StatusCode
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if resp != nil {
			buff := new(bytes.Buffer)
			buff.ReadFrom(resp.Body)
			return response, fmt.Errorf(buff.String())
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

func (service *BaseService) loadFromCredentialFile(serviceName string, separator string) error {
	// File path specified by env variable
	credentialFilePath := os.Getenv(IBM_CREDENTIAL_FILE_ENV)

	// Home directory
	if credentialFilePath == "" {
		var filePath = path.Join(UserHomeDir(), DEFAULT_CREDENTIAL_FILE_NAME)
		if _, err := os.Stat(filePath); err == nil {
			credentialFilePath = filePath
		}
	}

	// Top-level of project directory
	if credentialFilePath == "" {
		dir, _ := os.Getwd()
		var filePath = path.Join(dir, "..", DEFAULT_CREDENTIAL_FILE_NAME)
		if _, err := os.Stat(filePath); err == nil {
			credentialFilePath = filePath
		}
	}

	if credentialFilePath != "" {
		file, err := os.Open(credentialFilePath)
		if err != nil {
			return err
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var line = scanner.Text()
			var keyVal = strings.Split(line, separator)
			if len(keyVal) == 2 {
				service.setCredentialBasedOnType(serviceName, strings.ToLower(keyVal[0]), keyVal[1])
			}
		}
	}
	return nil
}

func (service *BaseService) setCredentialBasedOnType(serviceName, key, value string) {
	if strings.Contains(key, serviceName) {
		if strings.Contains(key, API_KEY) {
			service.SetIAMAPIKey(value)
		} else if strings.Contains(key, URL) {
			service.SetURL(value)
		} else if strings.Contains(key, USERNAME) {
			service.Options.Username = value
		} else if strings.Contains(key, PASSWORD) {
			service.Options.Password = value
		} else if strings.Contains(key, IAM_API_KEY) {
			service.SetIAMAPIKey(value)
		}
	}
}
