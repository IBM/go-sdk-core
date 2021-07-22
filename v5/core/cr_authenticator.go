package core

// (C) Copyright IBM Corp. 2021.
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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ComputeResourceAuthenticator implements an IAM-based authentication schema where by it
// retrieves a "compute resource token" from the local compute resource (VM)
// and uses that to obtain an IAM access token by invoking the IAM "get token" operation with grant-type=cr-token.
// The resulting IAM access token is then added to outbound requests in an Authorization header
// of the form:
//
// 		Authorization: Bearer <access-token>
//
type ComputeResourceAuthenticator struct {

	// [optional] The name of the file containing the injected CR token value (applies to
	// IKS-managed compute resources).
	// Default value: "/var/run/secrets/tokens/vault-token"
	CRTokenFilename string

	// [optional] The base endpoint URL to be used for invoking operations of the compute resource's
	// local Instance Metadata Service (applies to VSI-managed compute resources).
	// Default value: "http://169.254.169.254"
	InstanceMetadataServiceURL string

	// [optional] The name of the linked trusted IAM profile to be used when obtaining the IAM access token
	// (a CR token might map to multiple IAM profiles).
	// One of IAMProfileName or IAMProfileID must be specified.
	// Default value: ""
	IAMProfileName string

	// [optional] The id of the linked trusted IAM profile to be used when obtaining the IAM access token
	// (a CR token might map to multiple IAM profiles).
	// One of IAMProfileName or IAMProfileID must be specified.
	// Default value: ""
	IAMProfileID string

	// [optional] The IAM token server's base endpoint URL.
	// Default value: "https://iam.cloud.ibm.com"
	URL string

	// [optional] The ClientID and ClientSecret fields are used to form a "basic auth"
	// Authorization header for interactions with the IAM token server.
	// If neither field is specified, then no Authorization header will be sent
	// with token server requests.
	// These fields are both optional, but must be specified together.
	// Default value: ""
	ClientID     string
	ClientSecret string

	// [optional] A flag that indicates whether verification of the server's SSL certificate
	// should be disabled.
	// Default value: false
	DisableSSLVerification bool

	// [optional] The "scope" to use when fetching the access token from the IAM token server.
	// This can be used to obtain an access token with a specific scope.
	// Default value: ""
	Scope string

	// [optional] A set of key/value pairs that will be sent as HTTP headers in requests
	// made to the IAM token server.
	// Default value: nil
	Headers map[string]string

	// [optional] The http.Client object used in interacts with the IAM token server.
	// If not specified by the user, a suitable default Client will be constructed.
	Client *http.Client

	// The cached IAM access token and its expiration time.
	tokenData *iamTokenData

	// Mutex to synchronize access to the tokenData field.
	tokenDataMutex sync.Mutex
}

const (
	defaultCRTokenFilename = "/var/run/secrets/tokens/vault-token"
	defaultImdsEndpoint    = "http://169.254.169.254"
	imdsVersionDate        = "2021-07-15"
	imdsMetadataFlavor     = "ibm"
	iamGrantTypeCRToken    = "urn:ibm:params:oauth:grant-type:cr-token" // #nosec G101
	crtokenLifetime        = 300
)

var craRequestTokenMutex sync.Mutex

// NewComputeResourceAuthenticator constructs a new ComputeResourceAuthenticator instance with the supplied values and
// invokes the ComputeResourceAuthenticator's Validate() method.
func NewComputeResourceAuthenticator(crtokenFilename string, instanceMetadataServiceURL, iamProfileName string, iamProfileID string,
	url string, clientID string, clientSecret string, disableSSLVerification bool, scope string,
	headers map[string]string) (*ComputeResourceAuthenticator, error) {
	authenticator := &ComputeResourceAuthenticator{
		CRTokenFilename:            crtokenFilename,
		InstanceMetadataServiceURL: instanceMetadataServiceURL,
		IAMProfileName:             iamProfileName,
		IAMProfileID:               iamProfileID,
		URL:                        url,
		ClientID:                   clientID,
		ClientSecret:               clientSecret,
		DisableSSLVerification:     disableSSLVerification,
		Scope:                      scope,
		Headers:                    headers,
	}

	// Make sure the config is valid.
	err := authenticator.Validate()
	if err != nil {
		return nil, err
	}

	return authenticator, nil
}

// newComputeResourceAuthenticatorFromMap constructs a new ComputeResourceAuthenticator instance from a map containing
// configuration properties.
func newComputeResourceAuthenticatorFromMap(properties map[string]string) (authenticator *ComputeResourceAuthenticator, err error) {
	if properties == nil {
		return nil, fmt.Errorf(ERRORMSG_PROPS_MAP_NIL)
	}

	// Grab the AUTH_DISABLE_SSL string property and convert to a boolean.
	disableSSL, err := strconv.ParseBool(properties[PROPNAME_AUTH_DISABLE_SSL])
	if err != nil {
		disableSSL = false
	}

	authenticator, err = NewComputeResourceAuthenticator(
		properties[PROPNAME_CRTOKEN_FILENAME],
		properties[PROPNAME_INSTANCE_METADATA_SERVICE_URL],
		properties[PROPNAME_IAM_PROFILE_NAME],
		properties[PROPNAME_IAM_PROFILE_ID],
		properties[PROPNAME_AUTH_URL],
		properties[PROPNAME_CLIENT_ID],
		properties[PROPNAME_CLIENT_SECRET],
		disableSSL,
		properties[PROPNAME_SCOPE],
		nil)

	return
}

// AuthenticationType returns the authentication type for this authenticator.
func (*ComputeResourceAuthenticator) AuthenticationType() string {
	return AUTHTYPE_CRAUTH
}

// Authenticate adds IAM authentication information to the request.
//
// The IAM access token will be added to the request's headers in the form:
//
// 		Authorization: Bearer <access-token>
//
func (authenticator *ComputeResourceAuthenticator) Authenticate(request *http.Request) error {
	token, err := authenticator.GetToken()
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// getTokenData returns the tokenData field from the authenticator with synchronization.
func (authenticator *ComputeResourceAuthenticator) getTokenData() *iamTokenData {
	authenticator.tokenDataMutex.Lock()
	defer authenticator.tokenDataMutex.Unlock()

	return authenticator.tokenData
}

// setTokenData sets the 'tokenData' field in the authenticator with synchronization.
func (authenticator *ComputeResourceAuthenticator) setTokenData(tokenData *iamTokenData) {
	authenticator.tokenDataMutex.Lock()
	defer authenticator.tokenDataMutex.Unlock()

	authenticator.tokenData = tokenData
}

// Validate the authenticator's configuration.
//
// Ensures that one of IAMProfileName or IAMProfileID are specified, and the ClientId and ClientSecret pair are
// mutually inclusive.
func (authenticator *ComputeResourceAuthenticator) Validate() error {

	// Check to make sure that one of IAMProfileName or IAMProfileID are specified.
	if authenticator.IAMProfileName == "" && authenticator.IAMProfileID == "" {
		return fmt.Errorf(ERRORMSG_EXCLUSIVE_PROPS_ERROR, "IAMProfileName", "IAMProfileID")
	}

	// Validate ClientId and ClientSecret.  They must both be specified togther or neither should be specified.
	if authenticator.ClientID == "" && authenticator.ClientSecret == "" {
		// Do nothing as this is the valid scenario
	} else {
		// Since it is NOT the case that both properties are empty, make sure BOTH are specified.
		if authenticator.ClientID == "" {
			return fmt.Errorf(ERRORMSG_PROP_MISSING, "ClientID")
		}

		if authenticator.ClientSecret == "" {
			return fmt.Errorf(ERRORMSG_PROP_MISSING, "ClientSecret")
		}
	}

	//
	// Note: I'm not convinced that we should actually do this type of validation here
	// in the Validate() method.   For now, we'll just catch this failure as part of the
	// GetToken() processing later when an actual access token is needed.
	//
	// // Finally, check to make sure that we can in fact retrieve the CR token value.
	// // We do this here as a preliminary validation check (fail fast),
	// // rather than wait until we actually try to retrieve an access token.
	// crToken := authenticator.retrieveCRToken()
	// if crToken == "" {
	// 	return fmt.Errorf(ERRORMSG_UNABLE_RETRIEVE_CRTOKEN)
	// }

	return nil
}

// GetToken returns an access token to be used in an Authorization header.
// Whenever a new token is needed (when a token doesn't yet exist or the existing token has expired),
// a new access token is fetched from the token server.
func (authenticator *ComputeResourceAuthenticator) GetToken() (string, error) {
	if authenticator.getTokenData() == nil || !authenticator.getTokenData().isTokenValid() {
		GetLogger().Debug("Performing synchronous token fetch...")
		// synchronously request the token
		err := authenticator.synchronizedRequestToken()
		if err != nil {
			return "", err
		}
	} else if authenticator.getTokenData().needsRefresh() {
		GetLogger().Debug("Performing background asynchronous token fetch...")
		// If refresh needed, kick off a go routine in the background to get a new token
		//nolint: errcheck
		go authenticator.invokeRequestTokenData()
	} else {
		GetLogger().Debug("Using cached access token...")
	}

	// return an error if the access token is not valid or was not fetched
	if authenticator.getTokenData() == nil || authenticator.getTokenData().AccessToken == "" {
		return "", fmt.Errorf("Error while trying to get access token")
	}

	return authenticator.getTokenData().AccessToken, nil
}

// synchronizedRequestToken will check if the authenticator currently has
// a valid cached access token.
// If yes, then nothing else needs to be done.
// If no, then a blocking request is made to obtain a new IAM access token.
func (authenticator *ComputeResourceAuthenticator) synchronizedRequestToken() error {
	craRequestTokenMutex.Lock()
	defer craRequestTokenMutex.Unlock()
	// if cached token is still valid, then just continue to use it
	if authenticator.getTokenData() != nil && authenticator.getTokenData().isTokenValid() {
		return nil
	}

	return authenticator.invokeRequestTokenData()
}

// invokeRequestTokenData requests a new token from the IAM token server and
// unmarshals the response to produce the authenticator's 'tokenData' field (cache).
// Returns an error if the token was unable to be fetched, otherwise returns nil.
func (authenticator *ComputeResourceAuthenticator) invokeRequestTokenData() error {
	tokenResponse, err := authenticator.RequestToken()
	if err != nil {
		return err
	}

	if tokenData, err := newIamTokenData(tokenResponse); err != nil {
		return err
	} else {
		authenticator.setTokenData(tokenData)
	}

	return nil
}

// RequestToken first retrieves a CR token value from the current compute resource, then uses
// that to obtain a new IAM access token from the IAM token server.
func (authenticator *ComputeResourceAuthenticator) RequestToken() (*IamTokenServerResponse, error) {
	var err error
	var operationPath string = "/identity/token"

	// First, retrieve the CR token value for this compute resource.
	crToken, err := authenticator.retrieveCRToken()
	if crToken == "" {
		if err == nil {
			err = fmt.Errorf(ERRORMSG_UNABLE_RETRIEVE_CRTOKEN + ": reason unknown")
		}
		return nil, NewAuthenticationError(&DetailedResponse{}, err)
	}

	// Use the default IAM URL if one was not specified by the user.
	url := authenticator.URL
	if url == "" {
		url = defaultIamTokenServerEndpoint
	} else {
		// Canonicalize the URL by removing the operation path if it was specified by the user.
		url = strings.TrimSuffix(url, operationPath)
	}

	// Set up the request for the IAM "get token" invocation.
	builder := NewRequestBuilder(POST)
	_, err = builder.ResolveRequestURL(url, operationPath, nil)
	if err != nil {
		return nil, NewAuthenticationError(&DetailedResponse{}, err)
	}

	builder.AddHeader(CONTENT_TYPE, FORM_URL_ENCODED_HEADER)
	builder.AddHeader(Accept, APPLICATION_JSON)
	builder.AddFormData("grant_type", "", "", iamGrantTypeCRToken) // #nosec G101
	builder.AddFormData("cr_token", "", "", crToken)

	// We previously verified that one of IBMProfileID or IAMProfileName are specified,
	// so just process them individually here.
	// If both are specified, that's ok too (they must map to the same profile though).
	if authenticator.IAMProfileID != "" {
		builder.AddFormData("profile_id", "", "", authenticator.IAMProfileID)
	}
	if authenticator.IAMProfileName != "" {
		builder.AddFormData("profile_name", "", "", authenticator.IAMProfileName)
	}

	// If the scope was specified, add that form param to the request.
	if authenticator.Scope != "" {
		builder.AddFormData("scope", "", "", authenticator.Scope)
	}

	// Add user-defined headers to request.
	for headerName, headerValue := range authenticator.Headers {
		builder.AddHeader(headerName, headerValue)
	}

	req, err := builder.Build()
	if err != nil {
		return nil, NewAuthenticationError(&DetailedResponse{}, err)
	}

	// If client id and secret were configured by the user, then set them on the request
	// as a basic auth header.
	if authenticator.ClientID != "" && authenticator.ClientSecret != "" {
		req.SetBasicAuth(authenticator.ClientID, authenticator.ClientSecret)
	}

	// If the authenticator does not have a Client, create one now.
	if authenticator.Client == nil {
		authenticator.Client = &http.Client{
			Timeout: time.Second * 30,
		}

		// If the user told us to disable SSL verification, then do it now.
		if authenticator.DisableSSLVerification {
			transport := &http.Transport{
				/* #nosec G402 */
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			authenticator.Client.Transport = transport
		}
	}

	// If debug is enabled, then dump the request.
	if GetLogger().IsLogLevelEnabled(LevelDebug) {
		buf, dumpErr := httputil.DumpRequestOut(req, req.Body != nil)
		if dumpErr == nil {
			GetLogger().Debug("Request:\n%s\n", string(buf))
		} else {
			GetLogger().Debug(fmt.Sprintf("error while attempting to log outbound request: %s", dumpErr.Error()))
		}
	}

	GetLogger().Debug("Invoking IAM 'get token' operation: %s", builder.URL)
	resp, err := authenticator.Client.Do(req)
	if err != nil {
		return nil, NewAuthenticationError(&DetailedResponse{}, err)
	}
	GetLogger().Debug("Returned from IAM 'get token' operation, received status code %d", resp.StatusCode)

	// If debug is enabled, then dump the response.
	if GetLogger().IsLogLevelEnabled(LevelDebug) {
		buf, dumpErr := httputil.DumpResponse(resp, req.Body != nil)
		if dumpErr == nil {
			GetLogger().Debug("Response:\n%s\n", string(buf))
		} else {
			GetLogger().Debug(fmt.Sprintf("error while attempting to log inbound response: %s", dumpErr.Error()))
		}
	}

	// Check for a bad status code and handle an operation error.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buff := new(bytes.Buffer)
		_, _ = buff.ReadFrom(resp.Body)
		resp.Body.Close()

		// Create a DetailedResponse to be included in the error below.
		detailedResponse := &DetailedResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			RawResult:  buff.Bytes(),
		}

		iamErrorMsg := string(detailedResponse.RawResult)
		if iamErrorMsg == "" {
			iamErrorMsg = "IAM error response not available"
		}
		err = fmt.Errorf(ERRORMSG_IAM_GETTOKEN_ERROR, detailedResponse.StatusCode, builder.URL, iamErrorMsg)
		return nil, NewAuthenticationError(detailedResponse, err)
	}

	// Good response, so unmarshal the response body into an IamTokenServerResponse instance.
	tokenResponse := &IamTokenServerResponse{}
	_ = json.NewDecoder(resp.Body).Decode(tokenResponse)
	defer resp.Body.Close()

	return tokenResponse, nil
}

// retrieveCRToken retrieves the CR token for the current compute resource.
func (authenticator *ComputeResourceAuthenticator) retrieveCRToken() (crToken string, err error) {
	var errs []error
	var crTokenErr error

	// 1. IKS use-case
	// First, try to read the CR token value from a local file.
	crToken, crTokenErr = authenticator.readCRTokenFromFile()
	if crTokenErr != nil {
		errs = append(errs, crTokenErr)
	}

	// 2. VPC/VSI use-case
	// Next, try to obtain the CR token value from the Instance Metadata Service.
	if crToken == "" {
		crToken, crTokenErr = authenticator.retrieveCRTokenFromIMDS()
		if crTokenErr != nil {
			errs = append(errs, crTokenErr)
		}
	}

	// If we're going to return "" for the crToken (an error condition), then
	// gather up each of the error objects that resulted from each attempt at
	// retrieving the CR token value as we want to present the entire list
	// to the caller.
	if crToken == "" {
		var errorMsgs []string
		for _, e := range errs {
			errorMsgs = append(errorMsgs, "\t"+e.Error())
		}
		err = fmt.Errorf(ERRORMSG_UNABLE_RETRIEVE_CRTOKEN + "\n" + strings.Join(errorMsgs, "\n"))
	}

	return
}

// readCRTokenFromFile tries to read the CR token value from the local file system.
func (authenticator *ComputeResourceAuthenticator) readCRTokenFromFile() (crToken string, err error) {

	// Use the default filename if one wasn't supplied by the user.
	crTokenFilename := authenticator.CRTokenFilename
	if crTokenFilename == "" {
		crTokenFilename = defaultCRTokenFilename
	}

	GetLogger().Debug("Attempting to read CR token from file: %s\n", crTokenFilename)

	// Read the entire file into a byte slice, then convert to string.
	var bytes []byte
	bytes, err = ioutil.ReadFile(crTokenFilename)
	if err == nil {
		crToken = string(bytes)
		GetLogger().Debug("Successfully read CR token from file: %s\n", crTokenFilename)
	} else {
		GetLogger().Debug("Failed to read CR token value from file %s: %s\n", crTokenFilename, err.Error())
	}

	return
}

// This struct models the response to the "create_access_token" operation (PUT instance_identity/v1/token).
type instanceIdentityToken struct {
	AccessToken string `json:"access_token"`

	// The following fields are also present in the response, but
	// we don't need these fields because we're going to use the access token
	// (CR token value) immediately and will never cache it.
	// CreatedAt   string `json:"created_at"`
	// ExpiresAt   string `json:"expires_at"`
	// ExpiresIn   int64  `json:"expires_in"`
}

// retrieveCRTokenFromIMDS tries to retrieve the CR token value by invoking
// the "create_access_token" operation on the compute resource's
// local Instance Metadata Service.
func (authenticator *ComputeResourceAuthenticator) retrieveCRTokenFromIMDS() (crToken string, err error) {
	var operationPath string = "instance_identity/v1/token"

	url := authenticator.InstanceMetadataServiceURL
	if url == "" {
		url = defaultImdsEndpoint
	} else {
		// Canonicalize the URL by removing the operation path if it was specified by the user.
		url = strings.TrimSuffix(url, operationPath)
	}

	// Set up the request to invoke the "create_access_token" operation.
	builder := NewRequestBuilder(PUT)
	_, err = builder.ResolveRequestURL(url, operationPath, nil)
	if err != nil {
		return
	}

	// Set the params and request body.
	builder.AddQuery("version", imdsVersionDate)
	builder.AddHeader(CONTENT_TYPE, APPLICATION_JSON)
	builder.AddHeader(Accept, APPLICATION_JSON)
	builder.AddHeader("Metadata-Flavor", imdsMetadataFlavor)

	requestBody := fmt.Sprintf(`{"expires_in": %d}`, crtokenLifetime)
	_, _ = builder.SetBodyContentString(requestBody)

	req, err := builder.Build()
	if err != nil {
		return
	}

	// Create a Client with 30 sec timeout.
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// If debug is enabled, then dump the request.
	if GetLogger().IsLogLevelEnabled(LevelDebug) {
		buf, dumpErr := httputil.DumpRequestOut(req, req.Body != nil)
		if dumpErr == nil {
			GetLogger().Debug("Request:\n%s\n", string(buf))
		} else {
			GetLogger().Debug(fmt.Sprintf("error while attempting to log outbound request: %s", dumpErr.Error()))
		}
	}

	// Invoke the request.
	GetLogger().Debug("Invoking IMDS 'create_access_token' operation: %s", builder.URL)
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	GetLogger().Debug("Returned from IMDS 'create_access_token' operation, received status code %d", resp.StatusCode)

	// If debug is enabled, then dump the response.
	if GetLogger().IsLogLevelEnabled(LevelDebug) {
		buf, dumpErr := httputil.DumpResponse(resp, resp.Body != nil)
		if dumpErr == nil {
			GetLogger().Debug("Response:\n%s\n", string(buf))
		} else {
			GetLogger().Debug(fmt.Sprintf("error while attempting to log inbound response: %s", dumpErr.Error()))
		}
	}

	// Check for a bad status code and handle the operation error.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		buff := new(bytes.Buffer)
		_, _ = buff.ReadFrom(resp.Body)
		resp.Body.Close()

		// Create a DetailedResponse to be included in the error below.
		detailedResponse := &DetailedResponse{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			RawResult:  buff.Bytes(),
		}

		imdsErrorMsg := string(detailedResponse.RawResult)
		if imdsErrorMsg == "" {
			imdsErrorMsg = "IMDS error response not available"
		}
		err = fmt.Errorf(ERRORMSG_IMDS_OPERATION_ERROR, detailedResponse.StatusCode, builder.URL, imdsErrorMsg)
		return
	}

	// IMDS operation invocation must have worked, so unmarshal the response and retrieve the CR token.
	createTokenResponse := &instanceIdentityToken{}
	_ = json.NewDecoder(resp.Body).Decode(createTokenResponse)
	defer resp.Body.Close()

	// The CR token value is returned in the "access_token" field of the response object.
	crToken = createTokenResponse.AccessToken

	return
}
