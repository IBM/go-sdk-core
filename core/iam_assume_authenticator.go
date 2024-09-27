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
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"sync"
	"time"
)

// IamAssumeAuthenticator obtains an IAM access token using the IAM "get-token" operation's
// "assume" grant type. The authenticator obtains an initial IAM access token from a
// user-supplied apikey, then exchanges this initial IAM access token for another IAM access token
// that has "assumed the identity" of the specified trusted profile.
//
// The resulting IAM access token is added to each outbound request
// in an Authorization header of the form:
//
//	Authorization: Bearer <access-token>
type IamAssumeAuthenticator struct {

	// Specify exactly one of [IAMProfileID, IAMProfileCRN, or IAMProfileName] to
	// identify the trusted profile whose identity should be used.
	// If IAMProfileID or IAMProfileCRN is used, the trusted profile must exist
	// in the same account.
	// If and only if IAMProfileName is used, then AccountID must also be
	// specified to indicate the account that contains the trusted profile.
	IAMProfileID   string
	IAMProfileCRN  string
	IAMProfileName string

	// If and only if IAMProfileName is used to specify the trusted profile,
	// then IAMAccountID must also be specified to indicate the account that
	// contains the trusted profile.
	IAMAccountID string

	// The URL representing the IAM token server's endpoint; If not specified,
	// a suitable default value will be used [optional].
	URL     string
	urlInit sync.Once

	// A flag that indicates whether verification of the server's SSL certificate
	// should be disabled; defaults to false [optional].
	DisableSSLVerification bool

	// A set of key/value pairs that will be sent as HTTP headers in requests
	// made to the token server [optional].
	Headers map[string]string

	// The http.Client object used to invoke token server requests.
	// If not specified by the user, a suitable default Client will be constructed [optional].
	Client     *http.Client
	clientInit sync.Once

	// The User-Agent header value to be included with each token request.
	userAgent     string
	userAgentInit sync.Once

	// The cached token and expiration time.
	tokenData *iamTokenData

	// Mutex to make the tokenData field thread safe.
	tokenDataMutex sync.Mutex

	// An IamAuthenticator instance used to obtain the user's IAM access token from the apikey.
	iamDelegate *IamAuthenticator
}

const (
	iamGrantTypeAssume = "urn:ibm:params:oauth:grant-type:assume"
)

var (
	iamAssumeRequestTokenMutex sync.Mutex
)

// IamAssumeAuthenticatorBuilder is used to construct an IamAssumeAuthenticator instance.
type IamAssumeAuthenticatorBuilder struct {

	// Properties needed to construct an IamAuthenticator instance.
	IamAuthenticator

	// Properties needed to construct an IamAssumeAuthenticator instance.
	IamAssumeAuthenticator
}

// NewIamAssumeAuthenticatorBuilder returns a new builder struct that
// can be used to construct an IamAssumeAuthenticator instance.
func NewIamAssumeAuthenticatorBuilder() *IamAssumeAuthenticatorBuilder {
	return &IamAssumeAuthenticatorBuilder{}
}

// SetIAMProfileID sets the IAMProfileID field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetIAMProfileID(s string) *IamAssumeAuthenticatorBuilder {
	builder.IamAssumeAuthenticator.IAMProfileID = s
	return builder
}

// SetIAMProfileCRN sets the IAMProfileCRN field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetIAMProfileCRN(s string) *IamAssumeAuthenticatorBuilder {
	builder.IamAssumeAuthenticator.IAMProfileCRN = s
	return builder
}

// SetIAMProfileName sets the IAMProfileName field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetIAMProfileName(s string) *IamAssumeAuthenticatorBuilder {
	builder.IamAssumeAuthenticator.IAMProfileName = s
	return builder
}

// SetIAMAccountID sets the IAMAccountID field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetIAMAccountID(s string) *IamAssumeAuthenticatorBuilder {
	builder.IamAssumeAuthenticator.IAMAccountID = s
	return builder
}

// SetApiKey sets the ApiKey field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetApiKey(s string) *IamAssumeAuthenticatorBuilder {
	builder.IamAuthenticator.ApiKey = s
	return builder
}

// SetURL sets the URL field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetURL(s string) *IamAssumeAuthenticatorBuilder {
	builder.IamAuthenticator.URL = s
	builder.IamAssumeAuthenticator.URL = s
	return builder
}

// SetClientIDSecret sets the ClientId and ClientSecret fields in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetClientIDSecret(clientID, clientSecret string) *IamAssumeAuthenticatorBuilder {
	builder.IamAuthenticator.ClientId = clientID
	builder.IamAuthenticator.ClientSecret = clientSecret
	return builder
}

// SetDisableSSLVerification sets the DisableSSLVerification field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetDisableSSLVerification(b bool) *IamAssumeAuthenticatorBuilder {
	builder.IamAuthenticator.DisableSSLVerification = b
	builder.IamAssumeAuthenticator.DisableSSLVerification = b
	return builder
}

// SetScope sets the Scope field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetScope(s string) *IamAssumeAuthenticatorBuilder {
	builder.IamAuthenticator.Scope = s
	return builder
}

// SetHeaders sets the Headers field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetHeaders(headers map[string]string) *IamAssumeAuthenticatorBuilder {
	builder.IamAuthenticator.Headers = headers
	builder.IamAssumeAuthenticator.Headers = headers
	return builder
}

// SetClient sets the Client field in the builder.
func (builder *IamAssumeAuthenticatorBuilder) SetClient(client *http.Client) *IamAssumeAuthenticatorBuilder {
	builder.IamAuthenticator.Client = client
	builder.IamAssumeAuthenticator.Client = client
	return builder
}

// Build() returns a validated instance of the IamAssumeAuthenticator with the config that was set in the builder.
func (builder *IamAssumeAuthenticatorBuilder) Build() (*IamAssumeAuthenticator, error) {
	err := builder.IamAuthenticator.Validate()
	if err != nil {
		return nil, RepurposeSDKProblem(err, "validation-failed")
	}

	err = builder.IamAssumeAuthenticator.Validate()
	if err != nil {
		return nil, RepurposeSDKProblem(err, "validation-failed")
	}

	// If we passed validation, then save our IamAuthenticator instance.
	builder.IamAssumeAuthenticator.iamDelegate = &builder.IamAuthenticator

	return &builder.IamAssumeAuthenticator, nil
}

// Validate will verify the authenticator's configuration.
func (authenticator *IamAssumeAuthenticator) Validate() error {
	var numParams int
	if authenticator.IAMProfileCRN != "" {
		numParams++
	}
	if authenticator.IAMProfileID != "" {
		numParams++
	}
	if authenticator.IAMProfileName != "" {
		numParams++
	}

	// 1. The user should specify exactly one of IAMProfileID, IAMProfileCRN, or IAMProfileName
	if numParams != 1 {
		err := fmt.Errorf(ERRORMSG_EXCLUSIVE_PROPS_ERROR, "IAMProfileCRN, IAMProfileID", "IAMProfileName")
		return SDKErrorf(err, "", "exc-props", getComponentInfo())
	}

	// 2. The user should specify IAMAccountID if and only if IAMProfileName is also specified.
	if (authenticator.IAMProfileName == "") != (authenticator.IAMAccountID == "") {
		err := errors.New(ERRORMSG_ACCOUNTID_PROP_ERROR)
		return SDKErrorf(err, "", "both-props", getComponentInfo())
	}

	return nil
}

// client returns the authenticator's http client after potentially initializing it.
func (authenticator *IamAssumeAuthenticator) client() *http.Client {
	authenticator.clientInit.Do(func() {
		if authenticator.Client == nil {
			authenticator.Client = DefaultHTTPClient()
			authenticator.Client.Timeout = time.Second * 30

			// If the user told us to disable SSL verification, then do it now.
			if authenticator.DisableSSLVerification {
				transport := &http.Transport{
					// #nosec G402
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}
				authenticator.Client.Transport = transport
			}
		}
	})
	return authenticator.Client
}

// getUserAgent returns the User-Agent header value to be included in each token request invoked by the authenticator.
func (authenticator *IamAssumeAuthenticator) getUserAgent() string {
	authenticator.userAgentInit.Do(func() {
		authenticator.userAgent = fmt.Sprintf("%s/%s-%s %s", sdkName, "iam-assume-authenticator", __VERSION__, SystemInfo())
	})
	return authenticator.userAgent
}

// newIamAssumeAuthenticatorFromMap constructs a new IamAssumeAuthenticator instance from a map.
func newIamAssumeAuthenticatorFromMap(properties map[string]string) (authenticator *IamAssumeAuthenticator, err error) {
	if properties == nil {
		err := errors.New(ERRORMSG_PROPS_MAP_NIL)
		return nil, SDKErrorf(err, "", "missing-props", getComponentInfo())
	}

	disableSSL, err := strconv.ParseBool(properties[PROPNAME_AUTH_DISABLE_SSL])
	if err != nil {
		disableSSL = false
	}

	authenticator, err = NewIamAssumeAuthenticatorBuilder().
		SetIAMProfileID(properties[PROPNAME_IAM_PROFILE_ID]).
		SetIAMProfileCRN(properties[PROPNAME_IAM_PROFILE_CRN]).
		SetIAMProfileName(properties[PROPNAME_IAM_PROFILE_NAME]).
		SetIAMAccountID(properties[PROPNAME_IAM_ACCOUNT_ID]).
		SetApiKey(properties[PROPNAME_APIKEY]).
		SetURL(properties[PROPNAME_AUTH_URL]).
		SetClientIDSecret(properties[PROPNAME_CLIENT_ID], properties[PROPNAME_CLIENT_SECRET]).
		SetDisableSSLVerification(disableSSL).
		SetScope(properties[PROPNAME_SCOPE]).
		Build()

	return
}

// AuthenticationType returns the authentication type for this authenticator.
func (*IamAssumeAuthenticator) AuthenticationType() string {
	return AUTHTYPE_IAM_ASSUME
}

// Authenticate adds IAM authentication information to the request.
//
// The IAM access token will be added to the request's headers in the form:
//
//	Authorization: Bearer <access-token>
func (authenticator *IamAssumeAuthenticator) Authenticate(request *http.Request) error {
	token, err := authenticator.GetToken()
	if err != nil {
		return RepurposeSDKProblem(err, "get-token-fail")
	}

	request.Header.Set("Authorization", "Bearer "+token)
	GetLogger().Debug("Authenticated outbound request (type=%s)\n", authenticator.AuthenticationType())
	return nil
}

// url returns the authenticator's URL property after potentially initializing it.
func (authenticator *IamAssumeAuthenticator) url() string {
	authenticator.urlInit.Do(func() {
		if authenticator.URL == "" {
			// If URL was not specified, then use the default IAM endpoint.
			authenticator.URL = defaultIamTokenServerEndpoint
		} else {
			// Canonicalize the URL by removing the operation path if it was specified by the user.
			authenticator.URL = strings.TrimSuffix(authenticator.URL, iamAuthOperationPathGetToken)
		}
	})
	return authenticator.URL
}

// getTokenData returns the tokenData field from the authenticator.
func (authenticator *IamAssumeAuthenticator) getTokenData() *iamTokenData {
	authenticator.tokenDataMutex.Lock()
	defer authenticator.tokenDataMutex.Unlock()

	return authenticator.tokenData
}

// setTokenData sets the given iamTokenData to the tokenData field of the authenticator.
func (authenticator *IamAssumeAuthenticator) setTokenData(tokenData *iamTokenData) {
	authenticator.tokenDataMutex.Lock()
	defer authenticator.tokenDataMutex.Unlock()

	authenticator.tokenData = tokenData
}

// GetToken returns an access token to be used in an Authorization header.
// Whenever a new token is needed (when a token doesn't yet exist, needs to be refreshed,
// or the existing token has expired), a new access token is fetched from the token server.
func (authenticator *IamAssumeAuthenticator) GetToken() (string, error) {
	if authenticator.getTokenData() == nil || !authenticator.getTokenData().isTokenValid() {
		GetLogger().Debug("Performing synchronous token fetch...")
		// synchronously request the token
		err := authenticator.synchronizedRequestToken()
		if err != nil {
			return "", RepurposeSDKProblem(err, "request-token-fail")
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
		err := fmt.Errorf("Error while trying to get access token")
		return "", SDKErrorf(err, "", "no-token", getComponentInfo())
	}

	return authenticator.getTokenData().AccessToken, nil
}

// synchronizedRequestToken will synchronously fetch a new access token.
func (authenticator *IamAssumeAuthenticator) synchronizedRequestToken() error {
	iamAssumeRequestTokenMutex.Lock()
	defer iamAssumeRequestTokenMutex.Unlock()
	// if cached token is still valid, then just continue to use it
	if authenticator.getTokenData() != nil && authenticator.getTokenData().isTokenValid() {
		return nil
	}

	return authenticator.invokeRequestTokenData()
}

// invokeRequestTokenData requests a new token from the token server and
// unmarshals the token information to the tokenData cache. Returns
// an error if the token was unable to be fetched, otherwise returns nil
func (authenticator *IamAssumeAuthenticator) invokeRequestTokenData() error {
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

// RequestToken fetches a new access token from the token server and
// returns the response structure.
func (authenticator *IamAssumeAuthenticator) RequestToken() (*IamTokenServerResponse, error) {
	// Step 1: Obtain the user's IAM access token.
	userAccessToken, err := authenticator.iamDelegate.GetToken()
	if err != nil {
		return nil, RepurposeSDKProblem(err, "iam-error")
	}

	// Step 2: Exchange the user's access token for one that reflects the trusted profile
	// by invoking the getToken-assume operation.
	builder := NewRequestBuilder(POST)
	_, err = builder.ResolveRequestURL(authenticator.url(), iamAuthOperationPathGetToken, nil)
	if err != nil {
		return nil, RepurposeSDKProblem(err, "url-resolve-error")
	}

	builder.AddHeader(CONTENT_TYPE, "application/x-www-form-urlencoded")
	builder.AddHeader(Accept, APPLICATION_JSON)
	builder.AddHeader(headerNameUserAgent, authenticator.getUserAgent())

	builder.AddFormData("grant_type", "", "", iamGrantTypeAssume)
	builder.AddFormData("access_token", "", "", userAccessToken)
	if authenticator.IAMProfileCRN != "" {
		builder.AddFormData("profile_crn", "", "", authenticator.IAMProfileCRN)
	} else if authenticator.IAMProfileID != "" {
		builder.AddFormData("profile_id", "", "", authenticator.IAMProfileID)
	} else {
		builder.AddFormData("profile_name", "", "", authenticator.IAMProfileName)
		builder.AddFormData("account", "", "", authenticator.IAMAccountID)
	}

	// Add user-defined headers to request.
	for headerName, headerValue := range authenticator.Headers {
		builder.AddHeader(headerName, headerValue)
	}

	req, err := builder.Build()
	if err != nil {
		return nil, RepurposeSDKProblem(err, "request-build-error")
	}

	// If debug is enabled, then dump the request.
	if GetLogger().IsLogLevelEnabled(LevelDebug) {
		buf, dumpErr := httputil.DumpRequestOut(req, req.Body != nil)
		if dumpErr == nil {
			GetLogger().Debug("Request:\n%s\n", RedactSecrets(string(buf)))
		} else {
			GetLogger().Debug(fmt.Sprintf("error while attempting to log outbound request: %s", dumpErr.Error()))
		}
	}

	GetLogger().Debug("Invoking IAM 'get token (assume)' operation: %s", builder.URL)
	resp, err := authenticator.client().Do(req)
	if err != nil {
		err = SDKErrorf(err, "", "request-error", getComponentInfo())
		return nil, err
	}
	GetLogger().Debug("Returned from IAM 'get token (assume)' operation, received status code %d", resp.StatusCode)

	// If debug is enabled, then dump the response.
	if GetLogger().IsLogLevelEnabled(LevelDebug) {
		buf, dumpErr := httputil.DumpResponse(resp, req.Body != nil)
		if dumpErr == nil {
			GetLogger().Debug("Response:\n%s\n", RedactSecrets(string(buf)))
		} else {
			GetLogger().Debug(fmt.Sprintf("error while attempting to log inbound response: %s", dumpErr.Error()))
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		detailedResponse, err := processErrorResponse(resp)
		authError := authenticationErrorf(err, detailedResponse, "get_token", authenticator.getComponentInfo())

		// The err Summary is typically the message computed for the HTTPError instance in
		// processErrorResponse(). If the response body is non-JSON, the message will be generic
		// text based on the status code but authenticators have always used the stringified
		// RawResult, so update that here for compatibility.
		iamErrorMsg := err.Summary
		if detailedResponse.RawResult != nil {
			// RawResult is only populated if the response body is
			// non-JSON and we couldn't extract a message.
			iamErrorMsg = string(detailedResponse.RawResult)
		}

		authError.Summary = iamErrorMsg

		return nil, authError
	}

	tokenResponse := &IamTokenServerResponse{}
	_ = json.NewDecoder(resp.Body).Decode(tokenResponse)
	defer resp.Body.Close() // #nosec G307
	return tokenResponse, nil
}

func (authenticator *IamAssumeAuthenticator) getComponentInfo() *ProblemComponent {
	return NewProblemComponent("iam_identity_services", "")
}
