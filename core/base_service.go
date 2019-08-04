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
	IBM_CREDENTIAL_FILE_ENV      = "IBM_CREDENTIALS_FILE"
	DEFAULT_CREDENTIAL_FILE_NAME = "ibm-credentials.env"
	SDK_NAME                     = "ibm-go-sdk-core"
	UNKNOWN_ERROR                = "Unknown Error"

	// Names of properties that can be defined in a credential file (e.g. "MYSERVICE_USERNAME=user1").
	CREDPROP_URL                 = "URL"
	CREDPROP_USERNAME            = "USERNAME"
	CREDPROP_PASSWORD            = "PASSWORD"
	CREDPROP_IAM_APIKEY          = "IAM_APIKEY"
	CREDPROP_IAM_ACCESS_TOKEN    = "IAM_ACCESS_TOKEN"
	CREDPROP_IAM_URL             = "IAM_URL"
	CREDPROP_IAM_CLIENT_ID       = "IAM_CLIENT_ID"
	CREDPROP_IAM_CLIENT_SECRET   = "IAM_CLIENT_SECRET"
	CREDPROP_ICP4D_URL           = "ICP4D_URL"
	CREDPROP_ICP4D_ACCESS_TOKEN  = "ICP4D_ACCESS_TOKEN"
	CREDPROP_AUTHENTICATION_TYPE = "AUTHENTICATION_TYPE"

	ERRORMSG_CONFIG_PROPERTY_INVALID = "The %s value shouldn't start or end with curly brackets or quotes. Be sure to remove any {} and \" characters surrounding the %s value."
	ERRORMSG_AUTH_NOT_CONFIGURED     = "Authentication information was not properly configured."
)

// ServiceOptions Service options
type ServiceOptions struct {
	Version    string
	URL        string
	AuthConfig AuthenticatorConfig

	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	Username string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	Password string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	IAMApiKey string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	IAMAccessToken string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	IAMURL string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	IAMClientId string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	IAMClientSecret string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	ICP4DAccessToken string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	ICP4DURL string
	// Deprecated: use AuthConfig field to configure the desired authentication scheme.
	AuthenticationType string
}

// BaseService Base Service
type BaseService struct {
	Options        *ServiceOptions
	DefaultHeaders http.Header
	Client         *http.Client
	UserAgent      string
	authenticator  Authenticator
}

type CredentialProps map[string]string

// NewBaseService Instantiate a Base Service
func NewBaseService(options *ServiceOptions, serviceName, displayName string) (*BaseService, error) {
	if HasBadFirstOrLastChar(options.URL) {
		return nil, fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "URL", "URL")
	}

	service := BaseService{
		Options: options,

		Client: &http.Client{
			Timeout: time.Second * 30,
		},
	}

	// Set a default value for the User-Agent http header.
	service.SetUserAgent(service.BuildUserAgent())

	// If the AuthConfig property was specified, then use it to configure an Authenticator.
	// Otherwise, we'll try to use the deprecated auth-related properties to configure one.
	if options.AuthConfig != nil {
		err := service.SetAuthenticator(options.AuthConfig)
		if err != nil {
			return nil, err
		}
	} else {
		// The AuthConfig property was not specified, so we'll support the deprecated auth-related properties.
		var err error

		if service.Options.AuthenticationType != "" {
			service.Options.AuthenticationType = strings.ToLower(service.Options.AuthenticationType)
		}

		err = service.checkCredentials()
		if err != nil {
			return nil, err
		}

		var authConfig AuthenticatorConfig

		// 1. Credentials are passed in service options struct.
		authConfig = service.getAuthConfigFromServiceOptions()

		// 2. Credentials from credential file
		if authConfig == nil && displayName != "" {
			serviceName := strings.ToUpper(strings.Replace(displayName, " ", "_", -1))
			credentialProps, err := service.loadFromCredentialFile(serviceName, "=")
			if err != nil {
				return nil, err
			}

			if credentialProps != nil {
				// Try to form an AuthenticatorConfig from the credential file properties.
				authConfig = service.getAuthConfigFromCredentialProps(credentialProps)
			}
		}

		// 3. Try accessing VCAP_SERVICES env variable
		if authConfig == nil {
			credential := LoadFromVCAPServices(serviceName)
			if credential != nil {
				// Create a map from the Credential object.
				props := make(map[string]string)
				props[CREDPROP_URL] = credential.URL
				props[CREDPROP_IAM_APIKEY] = credential.APIKey
				props[CREDPROP_USERNAME] = credential.Username
				props[CREDPROP_PASSWORD] = credential.Password

				// Obtain an AuthenticatorConfig from the map of properties.
				authConfig = service.getAuthConfigFromCredentialProps(props)
			}
		}

		// If we have a non-nil authConfig from one of the sources above, then create an Authenticator from it.
		if authConfig != nil {
			err = service.SetAuthenticator(authConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	return &service, nil
}

// SetAuthenticator instantiates an Authenticator for the specified 'config' object
// and stores the Authenticator instance on the BaseService.
func (service *BaseService) SetAuthenticator(config AuthenticatorConfig) error {
	if config == nil {
		service.authenticator = nil
	} else {
		authObj, err := NewAuthenticator(config)
		if err != nil {
			return err
		}
		service.authenticator = authObj
	}

	return nil
}

// SetUsernameAndPassword Sets the Username and Password
//
// Deprecated: Use SetAuthenticator() instead.
func (service *BaseService) SetUsernameAndPassword(username string, password string) error {
	if HasBadFirstOrLastChar(username) {
		return fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "username", "username")
	}
	if HasBadFirstOrLastChar(password) {
		return fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "password", "password")
	}

	props := make(map[string]string)
	props[CREDPROP_USERNAME] = username
	props[CREDPROP_PASSWORD] = password

	authConfig := service.getAuthConfigFromCredentialProps(props)
	if authConfig != nil {
		err := service.SetAuthenticator(authConfig)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetIAMAccessToken Sets the IAM access token
//
// Deprecated: Use SetAuthenticator() instead.
func (service *BaseService) SetIAMAccessToken(iamAccessToken string) {
	authConfig := &IAMConfig{
		AccessToken: iamAccessToken,
	}
	service.SetAuthenticator(authConfig)
}

// SetICP4DAccessToken Sets the ICP4D access token
//
// Deprecated: Use SetAuthenticator() instead.
func (service *BaseService) SetICP4DAccessToken(icp4dAccessToken string) {
	authConfig := &ICP4DConfig{
		AccessToken: icp4dAccessToken,
	}
	service.SetAuthenticator(authConfig)
}

// SetIAMAPIKey Sets the IAM API key
//
// Deprecated: Use SetAuthenticator() instead.
func (service *BaseService) SetIAMAPIKey(iamAPIKey string) error {
	authConfig := &IAMConfig{
		ApiKey:       iamAPIKey,
		ClientId:     service.Options.IAMClientId,
		ClientSecret: service.Options.IAMClientSecret,
	}

	err := service.SetAuthenticator(authConfig)
	if err == nil {
		service.Options.IAMApiKey = iamAPIKey
	}

	return err
}

// SetURL sets the service URL
func (service *BaseService) SetURL(url string) error {
	if HasBadFirstOrLastChar(url) {
		return fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, "URL", "URL")
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
	if service.authenticator != nil {
		err := service.authenticator.Authenticate(req)
		if err != nil {
			return nil, err
		}
	} else {
		// Otherwise, we have no Authenticator... ERROR.
		return nil, fmt.Errorf(ERRORMSG_AUTH_NOT_CONFIGURED)
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

func isForICP(credential string) bool {
	return strings.HasPrefix(credential, ICP_PREFIX)
}

func isForICP4D(authenticationType, icp4dAccessToken string) bool {
	return authenticationType == AUTHTYPE_ICP4D || icp4dAccessToken != ""
}

func usesBasicForIAM(username, password string) bool {
	return username == APIKEY && !isForICP(password)
}

func hasIAMCredentials(iamAPIKey, iamAccessToken string) bool {
	return (iamAPIKey != "" || iamAccessToken != "") && !isForICP(iamAPIKey)
}

func (service *BaseService) loadFromCredentialFile(serviceName string, separator string) (CredentialProps, error) {
	var props map[string]string

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
			return nil, err
		}
		defer file.Close()

		// Parse the lines from the credential file and create a map of the properties related to this service.
		props = make(map[string]string)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var line = scanner.Text()

			// Parse the line into <property>=<value>
			var lineParts = strings.Split(line, separator)

			// Do we have a property name and a value (separated by '=')?
			if len(lineParts) == 2 {
				// Does the property name contain the service name?
				// If so, then compute the key by filtering out the service name,
				// then store the key/value pair in the map.
				index := strings.Index(lineParts[0], serviceName)
				if (index == 0) && (len(lineParts[0]) > len(serviceName)+1) {
					key := lineParts[0][len(serviceName)+1:]
					value := lineParts[1]
					props[key] = value
				}
			}
		}
	}
	return props, nil
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
			return fmt.Errorf(ERRORMSG_CONFIG_PROPERTY_INVALID, k, k)
		}
	}
	return nil
}

// Inspect the various fields in the ServiceOptions struct and instantiate a suitable
// AuthenticatorConfig.
func (service *BaseService) getAuthConfigFromServiceOptions() AuthenticatorConfig {
	var props CredentialProps

	// First, set up a map containing all the properties.
	// Note: this will similar to one obtained by loading credentials from a file.
	props = make(map[string]string)
	props[CREDPROP_URL] = service.Options.URL
	props[CREDPROP_USERNAME] = service.Options.Username
	props[CREDPROP_PASSWORD] = service.Options.Password
	props[CREDPROP_IAM_APIKEY] = service.Options.IAMApiKey
	props[CREDPROP_IAM_ACCESS_TOKEN] = service.Options.IAMAccessToken
	props[CREDPROP_IAM_URL] = service.Options.IAMURL
	props[CREDPROP_IAM_CLIENT_ID] = service.Options.IAMClientId
	props[CREDPROP_IAM_CLIENT_SECRET] = service.Options.IAMClientSecret
	props[CREDPROP_ICP4D_URL] = service.Options.ICP4DURL
	props[CREDPROP_ICP4D_ACCESS_TOKEN] = service.Options.ICP4DAccessToken
	props[CREDPROP_AUTHENTICATION_TYPE] = service.Options.AuthenticationType

	// Now just return an AuthenticatorConfig instance from the properties in the map.
	return service.getAuthConfigFromCredentialProps(props)
}

// This function will try to instantiate an AuthenticatorConfig from the properties found in "props".
func (service *BaseService) getAuthConfigFromCredentialProps(props CredentialProps) AuthenticatorConfig {
	var authConfig AuthenticatorConfig

	// If the service's URL property was specified AND the service option URL field is empty,
	// then set the URL on the service.
	if props[CREDPROP_URL] != "" && service.Options != nil && service.Options.URL == "" {
		service.SetURL(props[CREDPROP_URL])
	}

	if props[CREDPROP_AUTHENTICATION_TYPE] == AUTHTYPE_IAM || hasIAMCredentials(props[CREDPROP_IAM_APIKEY], props[CREDPROP_IAM_ACCESS_TOKEN]) {
		authConfig = &IAMConfig{
			URL:          props[CREDPROP_IAM_URL],
			ApiKey:       props[CREDPROP_IAM_APIKEY],
			AccessToken:  props[CREDPROP_IAM_ACCESS_TOKEN],
			ClientId:     props[CREDPROP_IAM_CLIENT_ID],
			ClientSecret: props[CREDPROP_IAM_CLIENT_SECRET],
		}
	} else if usesBasicForIAM(props[CREDPROP_USERNAME], props[CREDPROP_PASSWORD]) {
		authConfig = &IAMConfig{
			URL:          props[CREDPROP_IAM_URL],
			ApiKey:       props[CREDPROP_PASSWORD],
			AccessToken:  props[CREDPROP_IAM_ACCESS_TOKEN],
			ClientId:     props[CREDPROP_IAM_CLIENT_ID],
			ClientSecret: props[CREDPROP_IAM_CLIENT_SECRET],
		}
	} else if isForICP4D(props[CREDPROP_AUTHENTICATION_TYPE], props[CREDPROP_ICP4D_ACCESS_TOKEN]) {
		authConfig = &ICP4DConfig{
			URL:         props[CREDPROP_ICP4D_URL],
			Username:    props[CREDPROP_USERNAME],
			Password:    props[CREDPROP_PASSWORD],
			AccessToken: props[CREDPROP_ICP4D_ACCESS_TOKEN],
		}
	} else if isForICP(props[CREDPROP_IAM_APIKEY]) {
		authConfig = &BasicAuthConfig{
			Username: APIKEY,
			Password: props[CREDPROP_IAM_APIKEY],
		}
	} else if props[CREDPROP_USERNAME] != "" && props[CREDPROP_PASSWORD] != "" {
		authConfig = &BasicAuthConfig{
			Username: props[CREDPROP_USERNAME],
			Password: props[CREDPROP_PASSWORD],
		}
	} else if props[CREDPROP_AUTHENTICATION_TYPE] == AUTHTYPE_NOAUTH {
		authConfig = &NoauthConfig{}
	}

	return authConfig
}
