//go:build all || fast || basesvc
// +build all fast basesvc

package core

// (C) Copyright IBM Corp. 2020.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func testRoundTripBytes(t *testing.T, src []byte) {
	// Compress the input string and store in a buffer.
	srcReader := bytes.NewReader(src)
	gzipCompressor, err := NewGzipCompressionReader(srcReader)
	assert.Nil(t, err)
	compressedBuf := new(bytes.Buffer)
	_, err = compressedBuf.ReadFrom(gzipCompressor)
	assert.Nil(t, err)
	t.Log("Compressed length: ", compressedBuf.Len())

	// Now uncompress the compressed bytes and store in another buffer.
	bytesReader := bytes.NewReader(compressedBuf.Bytes())
	gzipDecompressor, err := NewGzipDecompressionReader(bytesReader)
	assert.Nil(t, err)
	decompressedBuf := new(bytes.Buffer)
	_, err = decompressedBuf.ReadFrom(gzipDecompressor)
	assert.Nil(t, err)
	t.Log("Uncompressed length: ", decompressedBuf.Len())

	// Verify that the uncompressed bytes produce the original string.
	assert.Equal(t, src, decompressedBuf.Bytes())
}
func TestGzipCompressionString1(t *testing.T) {
	testRoundTripBytes(t, []byte("Hello world!"))
}

func TestGzipCompressionString2(t *testing.T) {
	s := "This is a somewhat longer string, which we'll try to use in our compression/decompression testing.  Hopefully this will workout ok, but who knows???"
	testRoundTripBytes(t, []byte(s))
}

func TestGzipCompressionString3(t *testing.T) {
	s := "This is a string that should be able to be compressed by a LOT......................................................................................................................................................................................................................................................................................................................................................."
	testRoundTripBytes(t, []byte(s))
}

func TestGzipCompressionJSON1(t *testing.T) {
	jsonString := `{
		"rules": [
		  {
			"request_id": "request-0",
			"rule": {
			  "account_id": "44890a2fd24641a5a111738e358686cc",
			  "name": "Go Test Rule #1",
			  "description": "This is the description for Go Test Rule #1.",
			  "rule_type": "user_defined",
			  "target": {
				"service_name": "config-gov-sdk-integration-test-service",
				"resource_kind": "bucket",
				"additional_target_attributes": [
				  {
					"name": "resource_id",
					"operator": "is_not_empty"
				  }
				]
			  },
			  "required_config": {
				"description": "allowed_gb\u003c=20 \u0026\u0026 location=='us-east'",
				"and": [
				  {
					"property": "allowed_gb",
					"operator": "num_less_than_equals",
					"value": "20"
				  },
				  {
					"property": "location",
					"operator": "string_equals",
					"value": "us-east"
				  }
				]
			  },
			  "enforcement_actions": [
				{
				  "action": "disallow"
				}
			  ],
			  "labels": [
				"GoSDKIntegrationTest"
			  ]
			}
		  }
		],
		"Transaction-Id": "bb5bac98-fa55-4125-97a8-578811c39c81",
		"Headers": null
	  }`

	testRoundTripBytes(t, []byte(jsonString))
}

func TestGzipCompressionJSON2(t *testing.T) {
	s := make([]string, 0)

	// Create a large string slice with repeated values, which will result in a small compressed string.
	for i := 0; i < 100000; i++ {
		s = append(s, "This")
		s = append(s, "is")
		s = append(s, "a")
		s = append(s, "test")
		s = append(s, "that ")
		s = append(s, "should")
		s = append(s, "demonstrate")
		s = append(s, "lots")
		s = append(s, "of")
		s = append(s, "compression")
	}

	jsonString := toJSON(s)

	testRoundTripBytes(t, []byte(jsonString))
}
