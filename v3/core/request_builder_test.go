package core

// (C) Copyright IBM Corp. 2019.
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
	"io"
	"os"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func setup() *RequestBuilder {
	return NewRequestBuilder("GET")
}

func TestNewRequestBuilder(t *testing.T) {
	request := setup()
	assert.Equal(t, "GET", request.Method, "Got incorrect method types")
}

func TestConstructHTTPURL(t *testing.T) {
	endPoint := "https://gateway.watsonplatform.net/assistant/api"
	pathSegments := []string{"v1/workspaces", "message"}
	pathParameters := []string{"xxxxx"}
	request := setup()
	want := "https://gateway.watsonplatform.net/assistant/api/v1/workspaces/xxxxx/message"
	_, err := request.ConstructHTTPURL(endPoint, pathSegments, pathParameters)
	assert.Nil(t, err)
	assert.Equal(t, want, request.URL.String(), "Invalid construction of url")
}

func TestConstructHTTPURLWithNoPathParam(t *testing.T) {
	endPoint := "https://gateway.watsonplatform.net/assistant/api"
	pathSegments := []string{"v1/workspaces"}
	request := setup()
	want := "https://gateway.watsonplatform.net/assistant/api/v1/workspaces"
	_, err := request.ConstructHTTPURL(endPoint, pathSegments, nil)
	assert.Nil(t, err)
	assert.Equal(t, want, request.URL.String(), "Invalid construction of url")
}

func TestConstructHTTPURLWithEmptyPathSegments(t *testing.T) {
	endPoint := "https://gateway.watsonplatform.net/assistant/api"
	pathSegments := []string{"v1/workspaces", "", "segment", ""}
	pathParameters := []string{"param1", "param2", "param3", "param4"}
	request := setup()
	want := "https://gateway.watsonplatform.net/assistant/api/v1/workspaces/param1/param2/segment/param3/param4"
	_, err := request.ConstructHTTPURL(endPoint, pathSegments, pathParameters)
	assert.Nil(t, err)
	assert.Equal(t, want, request.URL.String(), "Invalid construction of url")
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
	assert.Equal(t, true, strings.HasPrefix(err.Error(), "There was an error parsing the service URL:"))
}

func TestAddQuery(t *testing.T) {
	request := setup()
	request.AddQuery("VERSION", "2018-22-09")
	assert.Equal(t, 1, len(request.Query), "Didnt set the query param")
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

func TestSetBodyContentError(t *testing.T) {
	request := setup()
	_, err := request.SetBodyContent("", nil, nil, 200)
	assert.Nil(t, request.Body)
	assert.Equal(t, err.Error(), "Invalid type for non-JSON body content: int")
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
	endPoint := "https://gateway.watsonplatform.net/assistant/api"
	pathSegments := []string{"v1/workspaces", "message"}
	pathParameters := []string{"xxxxx"}
	wantURL := "https://gateway.watsonplatform.net/assistant/api/xxxxx/v1/workspaces?Version=2018-22-09"

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
