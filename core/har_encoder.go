//go:build !js

package core

// (C) Copyright IBM Corp. 2025.
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
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	harMaxEntries      = 10000 // Limit entries to prevent unbounded growth
	harBinaryThreshold = 0.05  // Ratio of non-printable chars for binary detection
)

var (
	harOnce     sync.Once
	harEnabled  bool
	harFilePath string
	harMutex    sync.Mutex

	// Regex patterns for redacting secrets in non-JSON content.
	bearerTokenPattern  = regexp.MustCompile(`(?i)(bearer\s+)([a-zA-Z0-9\-._~+/]+=*)`)
	basicAuthPattern    = regexp.MustCompile(`(?i)(basic\s+)([a-zA-Z0-9+/]+=*)`)
	apiKeyPattern       = regexp.MustCompile(`(?i)(apikey[\s:=]+)([a-zA-Z0-9\-._~+/]+)`)
	tokenPattern        = regexp.MustCompile(`(?i)(token[\s:=]+)([a-zA-Z0-9\-._~+/]+)`)
	iamTokenPattern     = regexp.MustCompile(`(?i)(iam[_-]?token[\s:=]+)([a-zA-Z0-9\-._~+/]+)`)
	accessTokenPattern  = regexp.MustCompile(`(?i)(access[_-]?token[\s:=]+)([a-zA-Z0-9\-._~+/]+)`)
	sessionTokenPattern = regexp.MustCompile(`(?i)(session[_-]?token[\s:=]+)([a-zA-Z0-9\-._~+/]+)`)
	passwordPattern     = regexp.MustCompile(`(?i)(password[\s:=]+)([^\s&"'<>]+)`)
	secretPattern       = regexp.MustCompile(`(?i)(secret[\s:=]+)([a-zA-Z0-9\-._~+/]+)`)
	cookiePattern       = regexp.MustCompile(`(?i)(=[^;,\s]{8,})(;|,|$)`)
)

// HAREnabled returns true if HAR recording is enabled via the HAR_ENABLED environment variable.
func HAREnabled() bool {
	harOnce.Do(func() {
		harEnabled = os.Getenv("HAR_ENABLED") == "1"
		if harEnabled {
			customPath := os.Getenv("HAR_FILE_PATH")
			if customPath != "" {
				harFilePath = customPath
			} else {
				harFilePath = filepath.Join(os.TempDir(), "ibm-go-sdk-core.har")
			}
			GetLogger().Info("HAR recording enabled, writing to: %s\n", harFilePath)
		}
	})
	return harEnabled
}

// HAR 1.2 specification structures.
type harNameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type harPostData struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Params   []struct {
		Name  string `json:"name"`
		Value string `json:"value,omitempty"`
	} `json:"params,omitempty"`
}

type harRequest struct {
	Method      string         `json:"method"`
	URL         string         `json:"url"`
	HTTPVersion string         `json:"httpVersion"`
	Headers     []harNameValue `json:"headers"`
	QueryString []harNameValue `json:"queryString"`
	PostData    *harPostData   `json:"postData,omitempty"`
	HeadersSize int64          `json:"headersSize"`
	BodySize    int64          `json:"bodySize"`
}

type harContent struct {
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}

type harResponse struct {
	Status      int            `json:"status"`
	StatusText  string         `json:"statusText"`
	HTTPVersion string         `json:"httpVersion"`
	Headers     []harNameValue `json:"headers"`
	Content     harContent     `json:"content"`
	RedirectURL string         `json:"redirectURL"`
	HeadersSize int64          `json:"headersSize"`
	BodySize    int64          `json:"bodySize"`
}

type harTimings struct {
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
}

type harEntry struct {
	Pageref         string      `json:"pageref"`
	StartedDateTime time.Time   `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Request         harRequest  `json:"request"`
	Response        harResponse `json:"response"`
	Cache           struct{}    `json:"cache"`
	Timings         harTimings  `json:"timings"`
	ServerIPAddress string      `json:"serverIPAddress,omitempty"`
	Connection      string      `json:"connection,omitempty"`
}

type harLog struct {
	Version string `json:"version"`
	Creator struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"creator"`
	Pages   []struct{} `json:"pages"`
	Entries []harEntry `json:"entries"`
}

type harArchive struct {
	Log harLog `json:"log"`
}

// HARAppendWithCopies appends a request/response pair to the HAR file.
func HARAppendWithCopies(
	req *http.Request,
	resp *http.Response,
	startTime, endTime time.Time,
	callErr error,
	reqBody, respBody []byte,
	reqContentType, respContentType string,
) {
	if !HAREnabled() || req == nil {
		return
	}

	entry := buildHAREntry(req, resp, startTime, endTime, callErr, reqBody, respBody, reqContentType, respContentType)

	harMutex.Lock()
	defer harMutex.Unlock()

	archive := readOrCreateHARArchive()

	if len(archive.Log.Entries) >= harMaxEntries {
		GetLogger().Warn("HAR file reached maximum entries (%d), rotating...\n", harMaxEntries)
		rotateHARFile()
		archive = createNewHARArchive()
	}

	archive.Log.Entries = append(archive.Log.Entries, entry)
	writeHARArchive(archive)
}

func buildHAREntry(
	req *http.Request,
	resp *http.Response,
	startTime, endTime time.Time,
	callErr error,
	reqBody, respBody []byte,
	reqContentType, respContentType string,
) harEntry {
	entry := harEntry{
		Pageref:         "page_1",
		StartedDateTime: startTime.UTC(),
		Time:            float64(endTime.Sub(startTime).Milliseconds()),
		Request:         buildHARRequest(req, reqBody, reqContentType),
		Response:        buildHARResponse(resp, callErr, respBody, respContentType, req),
		Cache:           struct{}{},
		Timings: harTimings{
			Send:    -1,
			Wait:    float64(endTime.Sub(startTime).Milliseconds()),
			Receive: -1,
		},
	}
	return entry
}

func buildHARRequest(req *http.Request, reqBody []byte, contentType string) harRequest {
	harReq := harRequest{
		Method:      req.Method,
		URL:         req.URL.String(),
		HTTPVersion: getHTTPVersion(req.Proto),
		Headers:     convertHeaders(req.Header, true),
		QueryString: convertQueryString(req.URL),
		HeadersSize: -1,
		BodySize:    int64(len(reqBody)),
	}

	if len(reqBody) > 0 {
		text, encoding := processBodyContent(reqBody, true, contentType)
		if text != "" || encoding != "" {
			harReq.PostData = &harPostData{
				MimeType: contentType,
				Text:     text,
			}
		}
	}

	return harReq
}

func buildHARResponse(resp *http.Response, callErr error, respBody []byte, contentType string, req *http.Request) harResponse {
	harResp := harResponse{
		Status:      getStatusCode(resp, callErr),
		StatusText:  getStatusText(resp, callErr),
		HTTPVersion: getHTTPVersionFromResponse(resp, req),
		Headers:     convertHeaders(getResponseHeaders(resp), false),
		HeadersSize: -1,
		BodySize:    int64(len(respBody)),
		RedirectURL: "",
	}

	text, encoding := processBodyContent(respBody, false, contentType)
	harResp.Content = harContent{
		Size:     int64(len(respBody)),
		MimeType: contentType,
		Text:     text,
		Encoding: encoding,
	}

	if resp != nil && resp.StatusCode >= 300 && resp.StatusCode < 400 {
		harResp.RedirectURL = resp.Header.Get("Location")
	}

	return harResp
}

func convertHeaders(headers http.Header, isRequest bool) []harNameValue {
	if headers == nil {
		return []harNameValue{}
	}
	var result []harNameValue
	for name, values := range headers {
		for _, value := range values {
			if isSensitiveHeader(name) {
				value = redactSecretValue(value)
			}
			result = append(result, harNameValue{
				Name:  name,
				Value: value,
			})
		}
	}
	return result
}

func convertQueryString(u *url.URL) []harNameValue {
	if u == nil {
		return []harNameValue{}
	}

	var result []harNameValue
	for name, values := range u.Query() {
		for _, value := range values {
			result = append(result, harNameValue{
				Name:  name,
				Value: value,
			})
		}
	}
	return result
}

func processBodyContent(body []byte, isRequest bool, contentType string) (text string, encoding string) {
	if len(body) == 0 {
		return "", ""
	}

	if isBinaryContent(body) {
		return base64.StdEncoding.EncodeToString(body), "base64"
	}

	text = string(body)

	if strings.Contains(strings.ToLower(contentType), "json") ||
		strings.HasPrefix(strings.TrimSpace(text), "{") ||
		strings.HasPrefix(strings.TrimSpace(text), "[") {
		redactedJSON := redactJSONSecrets(text)
		if redactedJSON != "" {
			return redactedJSON, ""
		}
	}

	text = redactSecretValue(text)
	text = RedactSecrets(text)

	return text, ""
}

func redactJSONSecrets(jsonStr string) string {
	var data interface{}

	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return ""
	}

	redacted := redactJSONValue(data)

	result, err := json.MarshalIndent(redacted, "", "  ")
	if err != nil {
		return ""
	}

	return string(result)
}

func redactJSONValue(val interface{}) interface{} {
	switch v := val.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			lowerKey := strings.ToLower(key)
			if isSensitiveJSONKey(lowerKey) {
				result[key] = getRedactionLabel(lowerKey)
			} else {
				result[key] = redactJSONValue(value)
			}
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = redactJSONValue(item)
		}
		return result

	case string:
		if looksLikeToken(v) {
			return "[REDACTED_TOKEN]"
		}
		return v

	default:
		return v
	}
}

func isSensitiveJSONKey(key string) bool {
	sensitiveKeys := []string{
		"token", "apikey", "api_key", "password", "secret",
		"authorization", "auth", "credential", "access_token",
		"refresh_token", "session_token", "bearer", "api-key",
		"iam_token", "session_id", "cookie", "sessionid",
	}
	for _, sensitive := range sensitiveKeys {
		if strings.Contains(key, sensitive) {
			return true
		}
	}
	return false
}

func getRedactionLabel(key string) string {
	if strings.Contains(key, "bearer") {
		return "[REDACTED_BEARER_TOKEN]"
	}
	if strings.Contains(key, "apikey") || strings.Contains(key, "api_key") || strings.Contains(key, "api-key") {
		return "[REDACTED_API_KEY]"
	}
	if strings.Contains(key, "password") {
		return "[REDACTED_PASSWORD]"
	}
	if strings.Contains(key, "secret") {
		return "[REDACTED_SECRET]"
	}
	if strings.Contains(key, "iam_token") || strings.Contains(key, "iam-token") {
		return "[REDACTED_IAM_TOKEN]"
	}
	if strings.Contains(key, "access_token") || strings.Contains(key, "access-token") {
		return "[REDACTED_ACCESS_TOKEN]"
	}
	if strings.Contains(key, "session") {
		return "[REDACTED_SESSION_TOKEN]"
	}
	if strings.Contains(key, "cookie") {
		return "[REDACTED_COOKIE]"
	}
	return "[REDACTED_TOKEN]"
}

func looksLikeToken(s string) bool {
	if len(s) > 32 && regexp.MustCompile(`^[A-Za-z0-9\-._~+/]+=*$`).MatchString(s) {
		return true
	}
	if strings.Count(s, ".") == 2 && len(s) > 50 {
		return true
	}
	return false
}

func redactSecretValue(value string) string {
	value = bearerTokenPattern.ReplaceAllString(value, "${1}[REDACTED_BEARER_TOKEN]")
	value = basicAuthPattern.ReplaceAllString(value, "${1}[REDACTED_BASIC_AUTH]")
	value = apiKeyPattern.ReplaceAllString(value, "${1}[REDACTED_API_KEY]")
	value = iamTokenPattern.ReplaceAllString(value, "${1}[REDACTED_IAM_TOKEN]")
	value = accessTokenPattern.ReplaceAllString(value, "${1}[REDACTED_ACCESS_TOKEN]")
	value = sessionTokenPattern.ReplaceAllString(value, "${1}[REDACTED_SESSION_TOKEN]")
	value = tokenPattern.ReplaceAllString(value, "${1}[REDACTED_TOKEN]")
	value = passwordPattern.ReplaceAllString(value, "${1}[REDACTED_PASSWORD]")
	value = secretPattern.ReplaceAllString(value, "${1}[REDACTED_SECRET]")
	value = cookiePattern.ReplaceAllString(value, "=[REDACTED_COOKIE]${2}")
	return value
}

func isBinaryContent(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	nonPrintable := 0
	sampleSize := len(data)
	if sampleSize > 8192 {
		sampleSize = 8192
	}

	for i := 0; i < sampleSize; i++ {
		b := data[i]
		if b == 9 || b == 10 || b == 13 {
			continue
		}
		if b < 32 || b > 126 {
			nonPrintable++
		}
	}

	ratio := float64(nonPrintable) / float64(sampleSize)
	return ratio > harBinaryThreshold
}

func isSensitiveHeader(name string) bool {
	lowerName := strings.ToLower(name)
	sensitivePatterns := []string{
		"authorization",
		"cookie",
		"set-cookie",
		"token",
		"apikey",
		"api-key",
		"secret",
		"password",
		"credential",
		"session",
		"x-auth",
		"x-api",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(lowerName, pattern) {
			return true
		}
	}
	return false
}

func getHTTPVersion(proto string) string {
	if proto == "" {
		return "HTTP/1.1"
	}
	return proto
}

func getHTTPVersionFromResponse(resp *http.Response, req *http.Request) string {
	if resp != nil && resp.Proto != "" {
		return resp.Proto
	}
	if req != nil && req.Proto != "" {
		return req.Proto
	}
	return "HTTP/1.1"
}

func getStatusCode(resp *http.Response, err error) int {
	if resp != nil {
		return resp.StatusCode
	}
	if err != nil {
		return 0
	}
	return -1
}

func getStatusText(resp *http.Response, err error) string {
	if resp != nil {
		return resp.Status
	}
	if err != nil {
		return err.Error()
	}
	return ""
}

func getResponseHeaders(resp *http.Response) http.Header {
	if resp == nil {
		return nil
	}
	return resp.Header
}

func readOrCreateHARArchive() harArchive {
	data, err := os.ReadFile(harFilePath)
	if err != nil || len(data) == 0 {
		return createNewHARArchive()
	}

	var archive harArchive
	if err := json.Unmarshal(data, &archive); err != nil {
		GetLogger().Warn("Failed to parse existing HAR file, creating new: %s\n", err.Error())
		return createNewHARArchive()
	}

	return archive
}

func createNewHARArchive() harArchive {
	archive := harArchive{}
	archive.Log.Version = "1.2"
	archive.Log.Creator.Name = "ibm-go-sdk-core"
	archive.Log.Creator.Version = __VERSION__
	archive.Log.Pages = []struct{}{}
	archive.Log.Entries = []harEntry{}
	return archive
}

func writeHARArchive(archive harArchive) {
	data, err := json.MarshalIndent(archive, "", "  ")
	if err != nil {
		GetLogger().Error("Failed to marshal HAR archive: %s\n", err.Error())
		return
	}

	if err := os.WriteFile(harFilePath, data, 0600); err != nil {
		GetLogger().Error("Failed to write HAR file: %s\n", err.Error())
	}
}

func rotateHARFile() {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := strings.TrimSuffix(harFilePath, ".har") + "_" + timestamp + ".har"

	if err := os.Rename(harFilePath, backupPath); err != nil {
		GetLogger().Error("Failed to rotate HAR file: %s\n", err.Error())
	} else {
		GetLogger().Info("Rotated HAR file to: %s\n", backupPath)
	}
}
