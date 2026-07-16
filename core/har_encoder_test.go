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

package core

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testBearerToken  = "Bearer fake-test-token-xxx"                         // pragma: allowlist secret
	testBasicAuth    = "Basic dGVzdDp0ZXN0"                                 // pragma: allowlist secret
	testAPIKey       = "apikey: test-api-key-placeholder"                   //nolint:gosec // pragma: allowlist secret
	testGenericToken = "token: test-token-val"                              // pragma: allowlist secret
	testIAMToken     = "iam_token: test-iam-val"                            // pragma: allowlist secret
	testAccessToken  = "access_token: test-access-val"                      // pragma: allowlist secret
	testSessionToken = "session-token: test-session-val"                    // pragma: allowlist secret
	testPassword     = "password: test-pass-placeholder"                    // pragma: allowlist secret
	testSecret       = "secret: test-secret-placeholder"                    // pragma: allowlist secret
	testCookie       = "Cookie: session=long-cookie-value-here-for-testing" // pragma: allowlist secret
)

func TestHAREnabled(t *testing.T) {
	origEnabled := os.Getenv("HAR_ENABLED")
	origPath := os.Getenv("HAR_FILE_PATH")
	defer os.Setenv("HAR_ENABLED", origEnabled)
	defer os.Setenv("HAR_FILE_PATH", origPath)

	os.Setenv("HAR_ENABLED", "0")
	os.Setenv("HAR_FILE_PATH", "")
	harOnce = sync.Once{}
	if HAREnabled() {
		t.Error("expected HAR disabled when HAR_ENABLED=0")
	}

	os.Setenv("HAR_ENABLED", "1")
	harOnce = sync.Once{}
	if !HAREnabled() {
		t.Error("expected HAR enabled when HAR_ENABLED=1")
	}
}

func TestRedactSecretValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Bearer Token", testBearerToken, "[REDACTED_BEARER_TOKEN]"},
		{"Basic Auth", testBasicAuth, "[REDACTED_BASIC_AUTH]"},
		{"API Key", testAPIKey, "[REDACTED_API_KEY]"},
		{"Generic Token", testGenericToken, "[REDACTED_TOKEN]"},
		{"IAM Token", testIAMToken, "[REDACTED_IAM_TOKEN]"},
		{"Access Token", testAccessToken, "[REDACTED_ACCESS_TOKEN]"},
		{"Session Token", testSessionToken, "[REDACTED_SESSION_TOKEN]"},
		{"Password", testPassword, "[REDACTED_PASSWORD]"},
		{"Secret", testSecret, "[REDACTED_SECRET]"},
		{"Cookie", testCookie, "[REDACTED_COOKIE]"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := redactSecretValue(tc.input)
			if !strings.Contains(result, tc.expected) {
				t.Fatalf("expected redaction marker %q in %q", tc.expected, result)
			}
			var originalSecret string
			for _, sep := range []string{": ", "=", ":"} {
				parts := strings.SplitN(tc.input, sep, 2)
				if len(parts) == 2 {
					originalSecret = strings.TrimSpace(parts[1])
					break
				}
			}
			if originalSecret != "" && len(originalSecret) > 5 && strings.Contains(result, originalSecret) {
				t.Fatalf("original secret %q still visible in %q", originalSecret, result)
			}
		})
	}
}

func TestRedactJSONSecretsViaProcessBodyContent(t *testing.T) {
	body := []byte(`{"apikey":"test-api-key-placeholder","password":"p","nested":{"token":"zzz"}}`) // pragma: allowlist secret

	text, enc := processBodyContent(body, true, "application/json")
	if enc != "" {
		t.Fatalf("expected no encoding for JSON, got %q", enc)
	}
	if strings.Contains(text, "test-key-val") || strings.Contains(text, `"password":"p"`) || strings.Contains(text, `"token":"zzz"`) {
		t.Fatalf("JSON secrets not redacted in %q", text)
	}
}

func TestConvertHeaders(t *testing.T) {
	headers := http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Bearer fake-test-token-for-testing"}, // pragma: allowlist secret
		"User-Agent":    []string{"TestAgent/1.0"},
		"X-API-Key":     []string{"test-api-key-val"}, // pragma: allowlist secret
	}

	result := convertHeaders(headers, true)

	if len(result) != 4 {
		t.Errorf("expected 4 headers, got %d", len(result))
	}

	for _, nv := range result {
		switch nv.Name {
		case "Authorization":
			if !strings.Contains(nv.Value, "[REDACTED_") {
				t.Errorf("Authorization header not redacted: %s", nv.Value)
			}
			if strings.Contains(nv.Value, "fake-test-token") {
				t.Error("Original token still visible in Authorization header")
			}
		case "X-API-Key":
			if nv.Value != "test-api-key-val" {
				t.Errorf("X-API-Key unexpectedly altered: %q", nv.Value)
			}
		case "Content-Type":
			if nv.Value != "application/json" {
				t.Error("Non-sensitive header should not be redacted")
			}
		}
	}
}

func TestIsSensitiveHeader(t *testing.T) {
	tests := []struct {
		name        string
		isSensitive bool
	}{
		{"Authorization", true},
		{"Cookie", true},
		{"Set-Cookie", true},
		{"X-Auth-Token", true},
		{"X-API-Key", true},
		{"Api-Key", true},
		{"Session-Token", true},
		{"X-Secret", true},
		{"Password", true},
		{"Content-Type", false},
		{"User-Agent", false},
		{"Accept", false},
		{"Host", false},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := isSensitiveHeader(tc.name)
			if got != tc.isSensitive {
				t.Fatalf("isSensitiveHeader(%q)=%v, want %v", tc.name, got, tc.isSensitive)
			}
		})
	}
}

func TestProcessBodyContent(t *testing.T) {
	t.Run("Text Content", func(t *testing.T) {
		textBody := []byte(`{"user":"john","password":"test-pass-placeholder"}`) // pragma: allowlist secret
		text, encoding := processBodyContent(textBody, true, "application/json")
		if encoding != "" {
			t.Fatal("text content should not have encoding")
		}
		if !strings.Contains(text, "[REDACTED_PASSWORD]") || strings.Contains(text, "testPass") {
			t.Fatal("password in text body should be redacted")
		}
	})

	t.Run("Binary Content", func(t *testing.T) {
		binaryBody := []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		}
		text, encoding := processBodyContent(binaryBody, false, "")
		if encoding != "base64" {
			t.Fatalf("binary content should be base64 encoded, got %q", encoding)
		}
		if _, err := base64.StdEncoding.DecodeString(text); err != nil {
			t.Fatalf("invalid base64 encoding: %v", err)
		}
	})

	t.Run("Empty Body", func(t *testing.T) {
		text, encoding := processBodyContent([]byte{}, false, "")
		if text != "" || encoding != "" {
			t.Fatal("empty body should return empty strings")
		}
	})
}

func TestIsBinaryContent(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		isBinary bool
	}{
		{"Plain Text", []byte("This is plain text"), false},
		{"JSON", []byte(`{"key":"value"}`), false},
		{"PNG Header", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, true},
		{"PDF Header", []byte{0x25, 0x50, 0x44, 0x46, 0x2D, 0x31, 0x2E, 0x34}, false},
		{"Empty", []byte{}, false},
		{"Text with Newlines", []byte("Line 1\nLine 2\nLine 3"), false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isBinaryContent(tc.data)
			if got != tc.isBinary {
				t.Fatalf("isBinaryContent=%v, want %v", got, tc.isBinary)
			}
		})
	}
}

func TestBuildHARRequest(t *testing.T) {
	reqURL, _ := url.Parse("https://api.example.com/v1/resource?key=value&token=test")
	req := &http.Request{
		Method: "POST",
		URL:    reqURL,
		Proto:  "HTTP/1.1",
		Header: http.Header{
			"Content-Type":  []string{"application/json"},
			"Authorization": []string{"Bearer fake-test-bearer-token"}, // pragma: allowlist secret
		},
	}
	body := []byte(`{"data":"value"}`)

	harReq := buildHARRequest(req, body, "application/json")

	if harReq.Method != "POST" {
		t.Fatalf("method=%s, want POST", harReq.Method)
	}
	if harReq.HTTPVersion != "HTTP/1.1" {
		t.Fatalf("httpVersion=%s, want HTTP/1.1", harReq.HTTPVersion)
	}
	if len(harReq.QueryString) != 2 {
		t.Fatalf("expected 2 query params, got %d", len(harReq.QueryString))
	}
	authHeaderFound := false
	for _, h := range harReq.Headers {
		if h.Name == "Authorization" {
			authHeaderFound = true
			if !strings.Contains(h.Value, "[REDACTED_") || strings.Contains(h.Value, "fake-test-bearer") {
				t.Fatal("authorization header not redacted in HAR request")
			}
		}
	}
	if !authHeaderFound {
		t.Fatal("authorization header missing from HAR request")
	}
	if harReq.PostData == nil || harReq.PostData.MimeType != "application/json" {
		t.Fatal("postData missing or wrong MIME type")
	}
}

func TestBuildHARResponse(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"Set-Cookie":   []string{"session=test-session-id-for-testing"}, // pragma: allowlist secret
		},
	}
	body := []byte(`{"result":"success","token":"test-api-token-val"}`) // pragma: allowlist secret

	harResp := buildHARResponse(resp, nil, body, "application/json", nil)

	if harResp.Status != 200 {
		t.Fatalf("status=%d, want 200", harResp.Status)
	}
	cookieFound := false
	for _, h := range harResp.Headers {
		if h.Name == "Set-Cookie" {
			cookieFound = true
			if !strings.Contains(h.Value, "[REDACTED_") {
				t.Fatal("Set-Cookie header not redacted")
			}
		}
	}
	if !cookieFound {
		t.Fatal("Set-Cookie header missing")
	}
	if !strings.Contains(harResp.Content.Text, "[REDACTED_") || strings.Contains(harResp.Content.Text, "test-api-token-val") {
		t.Fatal("response body token not redacted")
	}
}

func TestHARArchiveStructure(t *testing.T) {
	archive := createNewHARArchive()
	if archive.Log.Version != "1.2" {
		t.Fatalf("version=%s, want 1.2", archive.Log.Version)
	}
	if archive.Log.Creator.Name != "ibm-go-sdk-core" {
		t.Fatal("wrong creator name")
	}
	if archive.Log.Entries == nil || len(archive.Log.Entries) != 0 {
		t.Fatal("new archive should have 0 entries")
	}
}

func TestHARAppendWithCopies(t *testing.T) {
	tmpFile := "/tmp/test-har-" + time.Now().Format("20060102-150405") + ".har"
	defer os.Remove(tmpFile)

	os.Setenv("HAR_ENABLED", "1")
	os.Setenv("HAR_FILE_PATH", tmpFile)
	harOnce = sync.Once{}

	reqURL, _ := url.Parse("https://api.example.com/test")
	req := &http.Request{
		Method: "GET",
		URL:    reqURL,
		Proto:  "HTTP/1.1",
		Header: http.Header{
			"Authorization": []string{"Bearer test-bearer-token-val"}, // pragma: allowlist secret
		},
	}

	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Proto:      "HTTP/1.1",
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	startTime := time.Now()
	endTime := startTime.Add(100 * time.Millisecond)

	reqBody := []byte(`{"request":"data"}`)
	respBody := []byte(`{"response":"data"}`)

	HARAppendWithCopies(req, resp, startTime, endTime, nil, reqBody, respBody, "application/json", "application/json")

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read HAR file: %v", err)
	}

	var archive harArchive
	if err := json.Unmarshal(data, &archive); err != nil {
		t.Fatalf("failed to parse HAR file: %v", err)
	}

	if len(archive.Log.Entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(archive.Log.Entries))
	}

	entry := archive.Log.Entries[0]
	authFound := false
	for _, h := range entry.Request.Headers {
		if h.Name == "Authorization" {
			authFound = true
			if !strings.Contains(h.Value, "[REDACTED_") || strings.Contains(h.Value, "test-bearer-token") {
				t.Fatal("authorization header not redacted in HAR file")
			}
		}
	}
	if !authFound {
		t.Fatal("authorization header missing from HAR entry")
	}
	if entry.Request.Method != "GET" {
		t.Fatalf("method=%s, want GET", entry.Request.Method)
	}
	if entry.Response.Status != 200 {
		t.Fatalf("status=%d, want 200", entry.Response.Status)
	}
}

func TestMultipleHAREntries(t *testing.T) {
	tmpFile := "/tmp/test-har-multi-" + time.Now().Format("20060102-150405") + ".har"
	defer os.Remove(tmpFile)

	os.Setenv("HAR_ENABLED", "1")
	os.Setenv("HAR_FILE_PATH", tmpFile)
	harOnce = sync.Once{}

	for i := 0; i < 3; i++ {
		reqURL, _ := url.Parse("https://api.example.com/test")
		req := &http.Request{
			Method: "GET",
			URL:    reqURL,
			Proto:  "HTTP/1.1",
		}
		resp := &http.Response{
			StatusCode: 200,
			Status:     "200 OK",
			Proto:      "HTTP/1.1",
		}
		HARAppendWithCopies(req, resp, time.Now(), time.Now(), nil, nil, nil, "", "")
	}

	data, _ := os.ReadFile(tmpFile)
	var archive harArchive
	_ = json.Unmarshal(data, &archive)

	if len(archive.Log.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(archive.Log.Entries))
	}
}
