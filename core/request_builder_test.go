//go:build all || fast || basesvc
// +build all fast basesvc

package core

// (C) Copyright IBM Corp. 2019, 2020.
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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	assert "github.com/stretchr/testify/assert"
)

func setup() *RequestBuilder {
	return NewRequestBuilder("GET")
}

func gzipDecompress(srcReader io.Reader) []byte {
	gzipDecompressor, _ := NewGzipDecompressionReader(srcReader)
	decompressedBuf := new(bytes.Buffer)
	_, err := decompressedBuf.ReadFrom(gzipDecompressor)
	if err != nil {
		panic(err)
	}
	return decompressedBuf.Bytes()
}

func TestNewRequestBuilder(t *testing.T) {
	request := setup()
	assert.Equal(t, "GET", request.Method, "Got incorrect method types")
}

func TestResolveRequestURL(t *testing.T) {
	request := setup()
	pathParams := map[string]string{
		"workspace_id": "xxxxx",
		"message_id":   "yyyyy",
	}

	expectedURL := "https://myservice.cloud.ibm.com/assistant/v1/workspaces/xxxxx/message/yyyyy"

	_, err := request.ResolveRequestURL("https://myservice.cloud.ibm.com/assistant", "v1/workspaces/{workspace_id}/message/{message_id}", pathParams)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())

	_, err = request.ResolveRequestURL("https://myservice.cloud.ibm.com/assistant/", "v1/workspaces/{workspace_id}/message/{message_id}", pathParams)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())

	_, err = request.ResolveRequestURL("https://myservice.cloud.ibm.com/assistant/", "/v1/workspaces/{workspace_id}/message/{message_id}", pathParams)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())

	_, err = request.ResolveRequestURL("https://myservice.cloud.ibm.com", "/assistant/v1/workspaces/{workspace_id}/message/{message_id}", pathParams)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())

	expectedURL = "https://myservice.cloud.ibm.com/assistant/v1/workspaces/xxxxx/message/yyyyy/again/yyyyy"

	_, err = request.ResolveRequestURL("https://myservice.cloud.ibm.com", "/assistant/v1/workspaces/{workspace_id}/message/{message_id}/again/{message_id}", pathParams)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())

	expectedURL = "https://myservice.cloud.ibm.com/api/v1"

	_, err = request.ResolveRequestURL("https://myservice.cloud.ibm.com/api/v1", "", nil)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())

	expectedURL = "https://myservice.cloud.ibm.com/api/v1/"

	_, err = request.ResolveRequestURL("https://myservice.cloud.ibm.com/api/v1", "/", nil)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())

	_, err = request.ResolveRequestURL("https://myservice.cloud.ibm.com/api/v1/", "/", nil)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())
}

func TestResolveRequestURLEncodedValues(t *testing.T) {
	request := setup()
	pathParams := map[string]string{
		"workspace_id": "ws/1",
		"message_id":   "message #2",
	}

	expectedURL := "https://host.com/assistant/v1/workspaces/ws%2F1/message/message%20%232"

	_, err := request.ResolveRequestURL("https://host.com/assistant", "v1/workspaces/{workspace_id}/message/{message_id}", pathParams)
	assert.Nil(t, err)
	assert.Equal(t, expectedURL, request.URL.String())
}

func TestResolveRequestURLErrors(t *testing.T) {
	request := setup()

	pathParams1 := map[string]string{
		"tenant_id":   "tenant-123",
		"resource_id": "",
	}

	_, err := request.ResolveRequestURL("https://host.com", "/v1/{tenant_id}/resources/{resource_id}/", pathParams1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'resource_id' is empty")

	_, err = request.ResolveRequestURL("", "/v1/{tenant_id}/resources/{resource_id}/", pathParams1)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "service URL is empty")

	_, err = request.ResolveRequestURL("://host.com", "/v1/path1", nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error parsing service URL")
}

func TestConstructHTTPURL(t *testing.T) {
	endPoint := "https://api.us-south.assistant.watson.cloud.ibm.com"
	pathSegments := []string{"v1/workspaces", "message"}
	pathParameters := []string{"xxxxx"}
	request := setup()
	want := "https://api.us-south.assistant.watson.cloud.ibm.com/v1/workspaces/xxxxx/message"
	_, err := request.ConstructHTTPURL(endPoint, pathSegments, pathParameters)
	assert.Nil(t, err)
	assert.Equal(t, want, request.URL.String(), "Invalid construction of url")
}

func TestConstructHTTPURLWithNoPathParam(t *testing.T) {
	endPoint := "https://api.us-south.assistant.watson.cloud.ibm.com"
	pathSegments := []string{"v1/workspaces"}
	request := setup()
	want := "https://api.us-south.assistant.watson.cloud.ibm.com/v1/workspaces"
	_, err := request.ConstructHTTPURL(endPoint, pathSegments, nil)
	assert.Nil(t, err)
	assert.Equal(t, want, request.URL.String(), "Invalid construction of url")
}

func TestConstructHTTPURLWithEmptyPathSegments(t *testing.T) {
	endPoint := "https://api.us-south.assistant.watson.cloud.ibm.com"
	pathSegments := []string{"v1/workspaces", "", "segment", ""}
	pathParameters := []string{"param1", "param2", "param3", "param4"}
	request := setup()
	want := "https://api.us-south.assistant.watson.cloud.ibm.com/v1/workspaces/param1/param2/segment/param3/param4"
	_, err := request.ConstructHTTPURL(endPoint, pathSegments, pathParameters)
	assert.Nil(t, err)
	assert.Equal(t, want, request.URL.String(), "Invalid construction of url")
}

func TestConstructHTTPURLWithEmptyPathParam(t *testing.T) {
	endPoint := "https://api.us-south.assistant.watson.cloud.ibm.com"
	pathSegments := []string{"v1/workspaces", "segment"}
	pathParameters := []string{""}
	request := setup()
	_, err := request.ConstructHTTPURL(endPoint, pathSegments, pathParameters)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "'[0]' is empty")
}

func TestConstructHTTPURLMissingURL(t *testing.T) {
	request := setup()
	_, err := request.ConstructHTTPURL("", nil, nil)
	assert.NotNil(t, err)
	assert.Equal(t, ERRORMSG_SERVICE_URL_MISSING, err.Error())
}

func TestConstructHTTPURLInvalidURL(t *testing.T) {
	request := setup()
	_, err := request.ConstructHTTPURL(":<badscheme>", nil, nil)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "error parsing service URL:")
}

func TestAddQuery(t *testing.T) {
	request := setup()
	request.AddQuery("VERSION", "2018-22-09")
	assert.Equal(t, 1, len(request.Query))
}

func TestAddQuerySlice(t *testing.T) {
	request := setup()
	float64Slice := []float64{float64(9.56), float64(4.56), float64(2.4)}
	err := request.AddQuerySlice("float_64_params_array", float64Slice)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(request.Query), "Didnt set the query param")
	builtQuery := request.Query["float_64_params_array"][0]
	expected := "9.56,4.56,2.4"
	assert.NotNil(t, builtQuery)
	assert.Equal(t, expected, builtQuery)
}

func TestAddQuerySliceError(t *testing.T) {
	request := setup()
	var slice interface{}
	err := request.AddQuerySlice("bad_input_param", slice)

	assert.NotNil(t, err)
	assert.Equal(t, 0, len(request.Query), "Query should be empty")
}

func TestAddHeader(t *testing.T) {
	request := setup()
	request.AddHeader("Content-Type", "application/json")
	assert.Equal(t, 1, len(request.Header), "Didnt set the header pair")
}

func readStream(body io.Reader) string {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(body)
	return buf.String()
}

func TestSetBodyContentJSON(t *testing.T) {
	testStructure := &TestStructure{
		Name: "wonder woman",
	}
	body := make(map[string]interface{})
	body["name"] = testStructure.Name
	want := "{\"name\":\"wonder woman\"}\n"

	request := setup()
	_, _ = request.SetBodyContentJSON(body)
	assert.NotNil(t, request.Body)
	assert.Equal(t, want, readStream(request.Body))

	request.Body = nil
	_, _ = request.SetBodyContent("", body, "", "")
	assert.NotNil(t, request.Body)
	assert.Equal(t, want, readStream(request.Body))

	request.Body = nil
	_, _ = request.SetBodyContent("", nil, body, "")
	assert.NotNil(t, request.Body)
	assert.Equal(t, want, readStream(request.Body))

	_, err := request.SetBodyContent("", make(chan int), nil, nil)
	assert.NotNil(t, err)

	_, errAgain := request.SetBodyContent("", nil, make(chan int), nil)
	assert.NotNil(t, errAgain)
}

func TestSetBodyContentString(t *testing.T) {
	var str = "hello GO SDK"
	request := setup()
	_, _ = request.SetBodyContentString(str)
	assert.NotNil(t, request.Body)
	assert.Equal(t, str, readStream(request.Body))
}

func TestSetBodyContentStream(t *testing.T) {
	pwd, _ := os.Getwd()
	var testFile io.ReadCloser
	testFile, err := os.Open(pwd + "/../resources/test_file.txt")
	assert.Nil(t, err)

	request := setup()
	_, _ = request.SetBodyContent("", nil, nil, testFile)
	assert.NotNil(t, request.Body)
	assert.Equal(t, "hello world from text file", readStream(request.Body))

	request.Body = nil
	testFile, _ = os.Open(pwd + "/../resources/test_file.txt")
	_, _ = request.SetBodyContent("", nil, nil, &testFile)
	assert.Equal(t, "hello world from text file", readStream(request.Body))
}

func TestSetBodyContent1(t *testing.T) {
	var str = "hello GO SDK"
	request := setup()
	_, _ = request.SetBodyContent("text/plain", nil, nil, str)
	assert.NotNil(t, request.Body)
	assert.Equal(t, str, readStream(request.Body))
}

func TestSetBodyContent2(t *testing.T) {
	var str = "hello GO SDK"
	request := setup()
	_, _ = request.SetBodyContent("text/plain", nil, nil, &str)
	assert.NotNil(t, request.Body)
	assert.Equal(t, str, readStream(request.Body))
}

// Test the case when SetBodyContent is given a nil pointer for jsonContent.
func TestSetBodyContent3(t *testing.T) {
	var (
		str  = "hello GO SDK"
		json *string
	)
	request := setup()
	_, _ = request.SetBodyContent("text/plain", json, nil, &str)
	assert.NotNil(t, request.Body)
	assert.Equal(t, str, readStream(request.Body))
}

func TestSetBodyContentError(t *testing.T) {
	request := setup()
	_, err := request.SetBodyContent("", nil, nil, 200)
	assert.Nil(t, request.Body)
	assert.Equal(t, err.Error(), "Invalid type for non-JSON body content: int")
}

func TestSetBodyContentNoContent(t *testing.T) {
	request := setup()
	_, err := request.SetBodyContent("", nil, nil, nil)
	assert.Nil(t, request.Body)
	assert.Equal(t, err.Error(), "No body content provided")
}

func TestBuildWithMultipartFormEmptyFileName(t *testing.T) {
	builder := NewRequestBuilder("POST")
	_, err := builder.ConstructHTTPURL("test.com", nil, nil)
	assert.Nil(t, err)

	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09").
		AddFormData("hello1", "", "text/plain", "Hello GO SDK").
		AddFormData("hello2", "", "", "Hello GO SDK again")
	request, _ := builder.Build()
	assert.NotNil(t, request.Body, "Couldnt build successfully")
}

func TestBuildWithMultipartForm(t *testing.T) {
	var str = "hello"
	json1 := make(map[string]interface{})
	json1["name1"] = "test name1"

	json2 := make(map[string]interface{})
	json2["name2"] = "test name2"

	builder := NewRequestBuilder("POST")
	_, err := builder.ConstructHTTPURL("test.com", nil, nil)
	assert.Nil(t, err)

	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09").
		AddFormData("name1", "json1.json", "application/json", json1).
		AddFormData("name2", "json2.json", "application/json", json2).
		AddFormData("hello", "", "text/plain", "Hello GO SDK").
		AddFormData("hello", "", "text/plain", &str)

	pwd, _ := os.Getwd()
	var testFile io.ReadCloser
	testFile, err = os.Open(pwd + "/../resources/test_file.txt")
	assert.Nil(t, err, "Could not open file")
	builder.AddFormData("test_file1", "test_file.txt", "application/octet-stream", testFile)
	builder.AddFormData("test_file2", "test_file.txt", "application/octet-stream", &testFile)

	request, err := builder.Build()
	assert.Nil(t, err, "Couldnt build successfully")
	assert.NotNil(t, request)
	assert.NotNil(t, request.Body)
	defer testFile.Close()
}

func TestGzipSetBodyContentJSON(t *testing.T) {
	testStructure := &TestStructure{
		Name: "wonder woman",
	}
	body := make(map[string]interface{})
	body["name"] = testStructure.Name
	want := "{\"name\":\"wonder woman\"}\n"

	builder := NewRequestBuilder("POST")
	_, _ = builder.ConstructHTTPURL("test.com", nil, nil)
	assert.False(t, builder.EnableGzipCompression)
	builder.EnableGzipCompression = true

	_, _ = builder.SetBodyContentJSON(body)
	assert.NotNil(t, builder.Body)

	request, err := builder.Build()
	assert.Nil(t, err)

	// Make sure the Content-Encoding header was set.
	contentEncoding := request.Header.Get(CONTENT_ENCODING)
	assert.NotEmpty(t, contentEncoding)
	assert.Equal(t, "gzip", contentEncoding)

	// Make sure the request body is the compressed JSON string.
	uncompressedBody := gzipDecompress(request.Body)
	assert.Equal(t, want, string(uncompressedBody))
}

func TestGzipSetBodyContentString(t *testing.T) {
	want := "This is an example of a request body in the form of a string.  This will be gzip-compressed........................................"
	builder := NewRequestBuilder("POST")
	builder.EnableGzipCompression = true
	_, _ = builder.ConstructHTTPURL("test.com", nil, nil)
	_, _ = builder.SetBodyContentString(want)
	assert.NotNil(t, builder.Body)

	request, err := builder.Build()
	assert.Nil(t, err)

	// Make sure the Content-Encoding header was set.
	contentEncoding := request.Header.Get(CONTENT_ENCODING)
	assert.NotEmpty(t, contentEncoding)
	assert.Equal(t, "gzip", contentEncoding)

	// Make sure the request body is the compressed JSON string.
	uncompressedBody := gzipDecompress(request.Body)
	assert.Equal(t, want, string(uncompressedBody))
}

func TestGzipSetBodyContentStream(t *testing.T) {
	var err error
	var bodyStream io.Reader
	bodyStream, err = os.Open("../resources/test_file.txt")
	assert.Nil(t, err)

	builder := NewRequestBuilder("POST")
	builder.EnableGzipCompression = true
	_, _ = builder.ConstructHTTPURL("test.com", nil, nil)
	_, _ = builder.SetBodyContentStream(bodyStream)
	assert.NotNil(t, builder.Body)

	request, err := builder.Build()
	assert.Nil(t, err)

	// Make sure the Content-Encoding header was set.
	contentEncoding := request.Header.Get(CONTENT_ENCODING)
	assert.NotEmpty(t, contentEncoding)
	assert.Equal(t, "gzip", contentEncoding)

	var expectedStream io.ReadCloser
	expectedStream, _ = os.Open("../resources/test_file.txt")
	expectedBuf := new(bytes.Buffer)
	_, err = expectedBuf.ReadFrom(expectedStream)
	assert.Nil(t, err)
	expectedStream.Close()

	// Make sure the request body is the compressed JSON string.
	uncompressedBody := gzipDecompress(request.Body)
	assert.Equal(t, expectedBuf.Bytes(), uncompressedBody)
}

func TestGzipNoBodyContent(t *testing.T) {
	builder := NewRequestBuilder("GET")
	builder.EnableGzipCompression = true
	_, _ = builder.ConstructHTTPURL("test.com", nil, nil)
	assert.Nil(t, builder.Body)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.Nil(t, request.Body)

	// Make sure the Content-Encoding was NOT set.
	contentEncoding := request.Header.Get(CONTENT_ENCODING)
	assert.Empty(t, contentEncoding)
}

func TestGzipBuildWithMultipartForm(t *testing.T) {
	var err error

	// Create a builder with gzip enabled.
	builder := NewRequestBuilder("POST")
	builder.EnableGzipCompression = true
	_, err = builder.ConstructHTTPURL("test.com", nil, nil)
	assert.Nil(t, err)

	// Create a JSON object and add it as a mime-part.
	s := make([]string, 0)
	for i := 0; i < 10000; i++ {
		s = append(s, "This")
		s = append(s, "is")
		s = append(s, "a")
		s = append(s, "test")
		s = append(s, "of")
		s = append(s, "the")
		s = append(s, "emergency")
		s = append(s, "broadcast")
		s = append(s, "system")
		s = append(s, "!")
	}
	jsonPart := make(map[string][]string)
	jsonPart["string_slice"] = s
	builder.AddFormData("json-part", "part1.json", "application/json", jsonPart)

	// Add a string mime-part.
	stringPart := "This is a string mime-part."
	builder.AddFormData("string-part", "", "text/plain", stringPart)

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request.Body)

	// Make sure the Content-Encoding header was set.
	contentEncoding := request.Header.Get(CONTENT_ENCODING)
	assert.NotEmpty(t, contentEncoding)
	assert.Equal(t, "gzip", contentEncoding)

	// Validation for a request that contains form parts is a challenge because the
	// content-disposition string associated with each mime-part is computed on the fly
	// and not really accessible.
	// So for this test, we'll check to make sure that within the entire gzip-decompressed
	// request body, we can find the content associated with the two mime-parts.
	actualBody := string(gzipDecompress(request.Body))

	expectedJSONPart := toJSON(jsonPart)
	assert.Contains(t, actualBody, expectedJSONPart)
	assert.Contains(t, actualBody, stringPart)
}

func TestURLEncodedForm(t *testing.T) {
	builder := NewRequestBuilder("POST")
	_, err := builder.ConstructHTTPURL("test.com", nil, nil)
	assert.Nil(t, err)

	builder.AddHeader("Content-Type", FORM_URL_ENCODED_HEADER).
		AddQuery("Version", "2018-22-09").
		AddFormData("grant_type", "", "", "lalalala").
		AddFormData("apikey", "", "", "xxxx")

	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)
}

func TestBuildWithMultipartFormWithRepeatedKeys(t *testing.T) {
	var str = "hello"
	json1 := make(map[string]interface{})
	json1["name1"] = "test name1"

	json2 := make(map[string]interface{})
	json2["name2"] = "test name2"

	builder := NewRequestBuilder("POST")
	_, err := builder.ConstructHTTPURL("test.com", nil, nil)
	assert.Nil(t, err)

	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09").
		AddFormData("name", "json1.json", "application/json", json1).
		AddFormData("name", "json2.json", "application/json", json2).
		AddFormData("hello", "", "text/plain", "Hello GO SDK").
		AddFormData("hello", "", "text/plain", &str)

	pwd, _ := os.Getwd()
	var testFile io.ReadCloser
	testFile, err = os.Open(pwd + "/../resources/test_file.txt")
	assert.Nil(t, err, "Could not open file")
	builder.AddFormData("test_file", "test_file1.txt", "application/octet-stream", &testFile)
	builder.AddFormData("test_file", "test_file2.txt", "application/octet-stream", &testFile)

	request, err := builder.Build()
	assert.Nil(t, err, "Couldnt build successfully")
	assert.NotNil(t, request)
	assert.NotNil(t, request.Body)
	err = request.ParseMultipartForm(32 << 20)
	assert.Nil(t, err, "Couldnt parts multipart form successfully")
	vs := request.MultipartForm.File["name"]
	assert.Equal(t, 2, len(vs))
	vs = request.MultipartForm.File["test_file"]
	assert.Equal(t, 2, len(vs))
	defer testFile.Close()
}

func TestBuild(t *testing.T) {
	endPoint := "https://api.us-south.assistant.watson.cloud.ibm.com"
	pathSegments := []string{"v1/workspaces", "message"}
	pathParameters := []string{"xxxxx"}
	wantURL := "https://api.us-south.assistant.watson.cloud.ibm.com/xxxxx/v1/workspaces?Version=2018-22-09"

	testStructure := &TestStructure{
		Name: "wonder woman",
	}
	body := make(map[string]interface{})
	body["name"] = testStructure.Name

	builder := NewRequestBuilder("POST")
	_, err := builder.ConstructHTTPURL(endPoint, pathParameters, pathSegments)
	assert.Nil(t, err)

	builder.AddHeader("Content-Type", "Application/json").
		AddQuery("Version", "2018-22-09")

	_, _ = builder.SetBodyContentJSON(body)
	request, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, request)
	assert.Equal(t, wantURL, request.URL.String())
	assert.Equal(t, "Application/json", request.Header["Content-Type"][0])
}

func TestRequestWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"name": "wonder woman"}`)
	}))
	defer server.Close()

	ctx, cancelFunc := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancelFunc()

	builder := NewRequestBuilder("GET")
	builder = builder.WithContext(ctx)
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, _ := NewNoAuthAuthenticator()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.Nil(t, err)
	assert.NotNil(t, detailedResponse)
	assert.Equal(t, http.StatusOK, detailedResponse.StatusCode)
	assert.Equal(t, "application/json", detailedResponse.Headers.Get("Content-Type"))

	result, ok := detailedResponse.Result.(*Foo)
	assert.Equal(t, true, ok)
	assert.NotNil(t, result)
	assert.NotNil(t, foo)
	assert.Equal(t, "wonder woman", *(result.Name))
}

func TestRequestWithContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		time.Sleep(2 * time.Second)
		fmt.Fprint(w, `{"name": "wonder woman"}`)
	}))
	defer server.Close()

	ctx, cancelFunc := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancelFunc()

	builder := NewRequestBuilder("GET")
	builder = builder.WithContext(ctx)
	_, err := builder.ConstructHTTPURL(server.URL, nil, nil)
	assert.Nil(t, err)
	req, _ := builder.Build()

	authenticator, _ := NewNoAuthAuthenticator()

	options := &ServiceOptions{
		URL:           server.URL,
		Authenticator: authenticator,
	}
	service, err := NewBaseService(options)
	assert.Nil(t, err)

	var foo *Foo
	detailedResponse, err := service.Request(req, &foo)
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
	assert.Contains(t, err.Error(), "context deadline exceeded")
	assert.Nil(t, detailedResponse)
}

func TestHostHeader1(t *testing.T) {
	// "baseline" test to ensure that if the Host header is not explicitly set,
	// then the host from the request URL is used.

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL("https://localhost:80/api/v1", "", nil)
	assert.Nil(t, err)

	req, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, req)

	assert.Equal(t, "localhost:80", req.Host)
	t.Logf("Host: %s\n", req.Host)
}

func TestHostHeader2(t *testing.T) {
	// Verify that if the "Host" header is set on the request builder,
	// then the resulting Request object will have its Host field set as well.

	builder := NewRequestBuilder("GET")
	_, err := builder.ResolveRequestURL("https://localhost:80/api/v1", "", nil)
	assert.Nil(t, err)

	builder.AddHeader("Host", "overridehost:81")

	req, err := builder.Build()
	assert.Nil(t, err)
	assert.NotNil(t, req)

	assert.Equal(t, "overridehost:81", req.Host)
	t.Logf("Host: %s\n", req.Host)
}
