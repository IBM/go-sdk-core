//go:build all || fast || auth
// +build all fast auth

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
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	// These two access tokens are the result of running curl to invoke
	// the POST /v1/authorize against a CP4D environment.

	// Username/password
	// curl -k -X POST https://<host>/icp4d-api/v1/authorize -H 'Content-Type: application/json' \
	//      -d '{"username": "testuser", "password": "<password>" }'
	jwtUserPwd = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6IlVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhY2Nlc3NfY2F0YWxvZyIsImNhbl9wcm92aXNpb24iLCJzaWduX2luX29ubHkiXSwiZ3JvdXBzIjpbMTAwMDBdLCJzdWIiOiJ0ZXN0dXNlciIsImlzcyI6IktOT1hTU08iLCJhdWQiOiJEU1giLCJ1aWQiOiIxMDAwMzMxMDAzIiwiYXV0aGVudGljYXRvciI6ImRlZmF1bHQiLCJpYXQiOjE2MTA1NDgxNjksImV4cCI6MTYxMDU5MTMzM30.AGbjQwWDQ7KG7Ef5orTH982kLmwExmj0eiDe3nke8frcm0EfshglU1nddIVBhrEI6vkrHZQSUoolLT6Kz1hUrbbRedC6E-XmJwPG9HcfG9BsW6CJ4hN5IbrJDf9maDBvKDLsEjH6YPTiAoMDNKsxLImHFms0GbIREAj_7Q7Xb2jpQYPR1JG32GAclq01deY8n4whE6WeyQqcbHUCGy3Q7sKddqEvT59XjLr1Mwm1uvIGnso_FkWJhvZs_z4aF0rVQes7gJZpOOSPkuA7l08KxvFmX3vF0IqmfudymEqaW9YH2ihAvHQBOJJtIkKaRga2TYyvfcwLFCXOABEi2lBOuQ"

	// Username/apikey
	// curl -k -X POST https://<host>/icp4d-api/v1/authorize -H 'Content-Type: application/json' \
	//      -d '{"username": "testuser", "api_key": "<apikey>" }'
	jwtUserApikey = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InRlc3R1c2VyIiwicm9sZSI6IlVzZXIiLCJwZXJtaXNzaW9ucyI6WyJhY2Nlc3NfY2F0YWxvZyIsImNhbl9wcm92aXNpb24iLCJzaWduX2luX29ubHkiXSwiZ3JvdXBzIjpbMTAwMDBdLCJzdWIiOiJ0ZXN0dXNlciIsImlzcyI6IktOT1hTU08iLCJhdWQiOiJEU1giLCJ1aWQiOiIxMDAwMzMxMDAzIiwiYXV0aGVudGljYXRvciI6ImRlZmF1bHQiLCJpYXQiOjE2MTA1NDgyNDgsImV4cCI6MTYxMDU5MTQxMn0.I8MgxrapKRt0nOn0F41NtLHQ5HGmInZNaJIWcNwyBgLWI5YY_98kpKLecN5d9Ll9g0_lapAFs_b8xpTya0Lvnp2Q81SloRFpDhAMUVHVWq46g2dvZd1JpoFB8NHwrkz2qE_JUHBIonJmQusy8vMm1m1CPy0pE6fTYH1d5EJG2vLo6f2eFiDizLfGxb0ym9lUOkK6dgNZw2T32N8IoSYNan6BQU25Jai6llWRLwZda7R521EPEw2AtPDsd95AxoTd8f4pptxfkL2uXpT35wRguap_09sRlvDTR18Ghs-GbtCh3Do-8OPGEFYKvJkSHNpiXPw8pvHEe5jCGl3l3F5vXQ"
)

func TestParseJWT(t *testing.T) {
	var err error
	var claims *coreJWTClaims

	claims, err = parseJWT(jwtUserPwd)
	assert.Nil(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, int64(1610591333), claims.ExpiresAt)
	assert.Equal(t, int64(1610548169), claims.IssuedAt)

	claims, err = parseJWT(jwtUserApikey)
	assert.Nil(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, int64(1610591412), claims.ExpiresAt)
	assert.Equal(t, int64(1610548248), claims.IssuedAt)
}

func TestParseJWTFail(t *testing.T) {
	_, err := parseJWT("segment1.segment2")
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())

	_, err = parseJWT("====.====.====")
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())

	_, err = parseJWT("segment1.segment2.segment3")
	assert.NotNil(t, err)
	t.Logf("Expected error: %s\n", err.Error())
}

func TestDecodeSegment(t *testing.T) {
	testStringDecoded := "testString\n"
	testStringEncoded := "dGVzdFN0cmluZwo="
	testStringEncodedShort := "dGVzdFN0cmluZwo"
	testStringInvalid := "???!"

	var err error
	var decoded []byte

	decoded, err = decodeSegment(testStringEncoded)
	assert.Nil(t, err)
	assert.Equal(t, testStringDecoded, string(decoded))

	decoded, err = decodeSegment(testStringEncodedShort)
	assert.Nil(t, err)
	assert.Equal(t, testStringDecoded, string(decoded))

	decoded, err = decodeSegment("")
	assert.Nil(t, err)
	assert.Equal(t, []byte{}, decoded)

	_, err = decodeSegment(testStringInvalid)
	assert.NotNil(t, err)
}
