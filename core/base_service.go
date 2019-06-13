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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

// common constants for core
const (
	APIKEY                       = "apikey"
	ICP_PREFIX                   = "icp-"
	USER_AGENT                   = "User-Agent"
	AUTHORIZATION                = "Authorization"
	BEARER                       = "Bearer"
	IBM_CREDENTIAL_FILE_ENV      = "IBM_CREDENTIALS_FILE"
	DEFAULT_CREDENTIAL_FILE_NAME = "ibm-credentials.env"
	URL                          = "url"
	USERNAME                     = "username"
	PASSWORD                     = "password"
	IAM_APIKEY                   = "iam_apikey"
	IAM_URL                      = "iam_url"
	SDK_NAME                     = "ibm-go-sdk-core"
	UNKNOWN_ERROR                = "Unknown Error"
	ICP4D                        = "icp4d"
	IAM                          = "iam"
)

// ServiceOptions Service options
type ServiceOptions struct {
	Version            string
	URL                string
	Username           string
	Password           string
	IAMApiKey          string
	IAMAccessToken     string
	IAMURL             string
	IAMClientId        string
	IAMClientSecret    string
	ICP4DAccessToken   string
	ICP4DURL           string
	AuthenticationType string
}

// BaseService Base Service
type BaseService struct {
	Options           *ServiceOptions
	DefaultHeaders    http.Header
	IAMTokenManager   *IAMTokenManager
	ICP4DTokenManager *ICP4DTokenManager
	Client            *http.Client
	UserAgent         string
}

// NewBaseService Instantiate a Base Service
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

	if service.Options.AuthenticationType != "" {
		service.Options.AuthenticationType = strings.ToLower(service.Options.AuthenticationType)
	}

	err := service.checkCredentials()
	if err != nil {
		return nil, err
	}

	service.SetUserAgent(service.BuildUserAgent())

	// 1. Credentials are passed in constructor
	if options.AuthenticationType == IAM || hasIAMCredentials(options.IAMApiKey, options.IAMAccessToken) {
		tokenManager, err := NewIAMTokenManager(
			options.IAMApiKey,
			options.IAMURL,
			options.IAMAccessToken,
			options.IAMClientId,
			options.IAMClientSecret)
		if err != nil {
			return nil, err
		}
		service.IAMTokenManager = tokenManager
	} else if usesBasicForIAM(options.Username, options.Password) {
		tokenManager, err := NewIAMTokenManager(
			options.Password,
			options.IAMURL,
			options.IAMAccessToken,
			options.IAMClientId,
			options.IAMClientSecret)
		if err != nil {
			return nil, err
		}
		service.IAMTokenManager = tokenManager
		service.Options.IAMApiKey = options.Password
		service.Options.Username = ""
		service.Options.Password = ""
	} else if isForICP4D(options.AuthenticationType, options.ICP4DAccessToken) {
		if options.ICP4DAccessToken == "" && options.ICP4DURL == "" {
			return nil, fmt.Errorf("The ICP4DURL is mandatory for ICP4D")
		}
		service.ICP4DTokenManager = NewICP4DTokenManager(
			options.ICP4DURL,
			options.Username,
			options.Password,
			options.ICP4DAccessToken)
	} else if isForICP(options.IAMApiKey) {
		service.Options.Username = APIKEY
		service.Options.Password = options.IAMApiKey
	}

	// 2. Credentials from credential file
	if displayName != "" && service.Options.Username == "" &&
		service.IAMTokenManager == nil && service.ICP4DTokenManager == nil {
		serviceName := strings.ToLower(strings.Replace(displayName, " ", "_", -1))
		service.loadFromCredentialFile(serviceName, "=")
	}

	// 3. Try accessing VCAP_SERVICES env variable
	if service.Options.Username == "" && service.IAMTokenManager == nil && service.ICP4DTokenManager == nil {
		credential := LoadFromVCAPServices(serviceName)
		if credential != nil {
			if credential.URL != "" {
				service.SetURL(credential.URL)
			}

			if credential.APIKey != "" {
				err := service.SetIAMAPIKey(credential.APIKey)
				if err != nil {
					return nil, err
				}
			} else if credential.Username != "" && credential.Password != "" {
				service.SetUsernameAndPassword(credential.Username, credential.Password)
			}
		}

		if service.Options.Username == "" && service.IAMTokenManager == nil && service.ICP4DTokenManager == nil {
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

// SetIAMAccessToken Sets the IAM access token
func (service *BaseService) SetIAMAccessToken(iamAccessToken string) {
	if service.IAMTokenManager != nil {
		service.IAMTokenManager.SetIAMAccessToken(iamAccessToken)
	} else {
		tokenManager, _ := NewIAMTokenManager("", "", iamAccessToken, "", "")
		service.IAMTokenManager = tokenManager
	}
	service.Options.IAMAccessToken = iamAccessToken
}

// SetICP4DAccessToken Sets the ICP4D access token
func (service *BaseService) SetICP4DAccessToken(icp4dAccessToken string) {
	if service.ICP4DTokenManager != nil {
		service.ICP4DTokenManager.SetICP4DAccessToken(icp4dAccessToken)
	} else {
		tokenManager := NewICP4DTokenManager("", "", "", icp4dAccessToken)
		service.ICP4DTokenManager = tokenManager
	}
	service.Options.ICP4DAccessToken = icp4dAccessToken
}

// SetIAMAPIKey Sets the IAM API key
func (service *BaseService) SetIAMAPIKey(iamAPIKey string) error {
	if HasBadFirstOrLastChar(iamAPIKey) {
		return fmt.Errorf("The credentials shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your credentials")
	}
	if service.IAMTokenManager != nil {
		service.IAMTokenManager.SetIAMAPIKey(iamAPIKey)
	} else {
		tokenManager, err := NewIAMTokenManager(iamAPIKey, "", "",
			service.Options.IAMClientId, service.Options.IAMClientSecret)
		if err != nil {
			return err
		}
		service.IAMTokenManager = tokenManager
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

	if service.ICP4DTokenManager != nil {
		service.ICP4DTokenManager.DisableSSLVerification()
	}
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
	if service.IAMTokenManager != nil {
		token, err := service.IAMTokenManager.GetToken()
		if err != nil {
			return nil, err
		}
		req.Header.Add(AUTHORIZATION, fmt.Sprintf(`%s %s`, BEARER, token))
	} else if service.ICP4DTokenManager != nil {
		token, err := service.ICP4DTokenManager.GetToken()
		if err != nil {
			return nil, err
		}
		req.Header.Add(AUTHORIZATION, fmt.Sprintf(`%s %s`, BEARER, token))
	} else if service.Options.Username != "" && service.Options.Password != "" {
		req.SetBasicAuth(service.Options.Username, service.Options.Password)
	}

	// Perform the request
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

func isForICP(credential string) bool {
	return strings.HasPrefix(credential, ICP_PREFIX)
}

func isForICP4D(authenticationType, icp4dAccessToken string) bool {
	return authenticationType == ICP4D || icp4dAccessToken != ""
}

func usesBasicForIAM(username, password string) bool {
	return username == APIKEY && !isForICP(password)
}

func hasIAMCredentials(iamAPIKey, iamAccessToken string) bool {
	return (iamAPIKey != "" || iamAccessToken != "") && !isForICP(iamAPIKey)
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

func (service *BaseService) checkCredentials() error {
	credentialsToCheck := map[string]string{
		"URL":         service.Options.URL,
		"username":    service.Options.Username,
		"password":    service.Options.Password,
		"credentials": service.Options.IAMApiKey,
	}

	for k, v := range credentialsToCheck {
		if HasBadFirstOrLastChar(v) {
			return fmt.Errorf("The %s shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding your %s", k, k)
		}
	}
	return nil
}

func (service *BaseService) setCredentialBasedOnType(serviceName, key, value string) {
	if strings.Contains(key, serviceName) {
		if strings.Contains(key, APIKEY) {
			service.SetIAMAPIKey(value)
		} else if strings.Contains(key, URL) {
			service.SetURL(value)
		} else if strings.Contains(key, USERNAME) {
			service.Options.Username = value
		} else if strings.Contains(key, PASSWORD) {
			service.Options.Password = value
		} else if strings.Contains(key, IAM_APIKEY) {
			service.SetIAMAPIKey(value)
		}
	}
}
