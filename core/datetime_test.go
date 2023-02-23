//go:build all || fast
// +build all fast

package core

/**
 * (C) Copyright IBM Corp. 2020.
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
	"encoding/json"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/stretchr/testify/assert"
)

func _testDateTime(t *testing.T, src string, expected string) {
	// parse "src" into a DateTime value.package core
	dt, err := strfmt.ParseDateTime(src)
	assert.Nil(t, err)

	// format the DateTime value and verify.
	actual := dt.String()
	assert.NotNil(t, actual)

	assert.Equal(t, expected, actual)
}

func TestDateTime(t *testing.T) {
	// RFC 3339 with various flavors of tz-offset
	_testDateTime(t, "2016-06-20T04:25:16.218Z", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T04:25:16.218+0000", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T04:25:16.218+00", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T05:25:16.218+01", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T04:25:16.218-0000", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T04:25:16.218-00", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T00:25:16.218-0400", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T00:25:16.218-04", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T07:25:16.218+0300", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T07:25:16.218+03", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T04:25:16Z", "2016-06-20T04:25:16.000Z")
	_testDateTime(t, "2016-06-20T01:25:16-0300", "2016-06-20T04:25:16.000Z")
	_testDateTime(t, "2016-06-20T01:25:16-03:00", "2016-06-20T04:25:16.000Z")
	_testDateTime(t, "2016-06-20T08:55:16+04:30", "2016-06-20T04:25:16.000Z")
	_testDateTime(t, "2016-06-20T16:25:16+12:00", "2016-06-20T04:25:16.000Z")

	// RFC 3339 with nanoseconds for the Catalog-Managements of the world.
	_testDateTime(t, "2020-03-12T10:52:12.866305005-04:00", "2020-03-12T14:52:12.866Z")
	_testDateTime(t, "2020-03-12T14:52:12.866305005Z", "2020-03-12T14:52:12.866Z")
	_testDateTime(t, "2020-03-12T16:52:12.866305005+02:30", "2020-03-12T14:22:12.866Z")
	_testDateTime(t, "2020-03-12T14:52:12.866305Z", "2020-03-12T14:52:12.866Z")

	// UTC datetime with no TZ.
	_testDateTime(t, "2016-06-20T04:25:16.218", "2016-06-20T04:25:16.218Z")
	_testDateTime(t, "2016-06-20T04:25:16", "2016-06-20T04:25:16.000Z")

	// Dialog datetime.
	_testDateTime(t, "2016-06-20 04:25:16", "2016-06-20T04:25:16.000Z")

	// Alchemy datetime.
	// _testDateTime(t, "20160620T042516", "2016-06-20T04:25:16.000Z")

	// IAM Identity Service.
	_testDateTime(t, "2020-11-10T12:28+0000", "2020-11-10T12:28:00.000Z")
	_testDateTime(t, "2020-11-10T07:28-0500", "2020-11-10T12:28:00.000Z")
	_testDateTime(t, "2020-11-10T12:28Z", "2020-11-10T12:28:00.000Z")
}

type DateTimeModel struct {
	WsVictory *strfmt.DateTime `json:"ws_victory"`
}

type DateModel struct {
	WsVictory *strfmt.Date `json:"ws_victory"`
}

func roundTripTestDate(t *testing.T, inputJSON string, expectedOutputJSON string) {

	// Unmarshal inputJSON into a DateTimeModel instance
	var dModel *DateModel = nil
	err := json.Unmarshal([]byte(inputJSON), &dModel)
	assert.Nil(t, err)

	// Now marshal the model instance and verify the resulting JSON string.
	buf, err := json.Marshal(dModel)
	assert.Nil(t, err)
	actualOutputJSON := string(buf)

	t.Logf("Date input: %s, output: %s\n", inputJSON, actualOutputJSON)
	assert.Equal(t, expectedOutputJSON, actualOutputJSON)
}
func roundTripTestDateTime(t *testing.T, inputJSON string, expectedOutputJSON string) {

	// Unmarshal inputJSON into a DateTimeModel instance
	var dtModel *DateTimeModel = nil
	err := json.Unmarshal([]byte(inputJSON), &dtModel)
	assert.Nil(t, err)

	// Now marshal the model instance and verify the resulting JSON string.
	buf, err := json.Marshal(dtModel)
	assert.Nil(t, err)
	actualOutputJSON := string(buf)

	t.Logf("DateTime input: %s, output: %s\n", inputJSON, actualOutputJSON)
	assert.Equal(t, expectedOutputJSON, actualOutputJSON)
}

func TestModelsDateTime(t *testing.T) {
	// RFC 3339 date-time with milliseconds with Z tz-offset.
	roundTripTestDateTime(t, `{"ws_victory":"1903-10-13T21:30:00.000Z"}`, `{"ws_victory":"1903-10-13T21:30:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1903-10-13T21:30:00.00011Z"}`, `{"ws_victory":"1903-10-13T21:30:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1903-10-13T21:30:00.0001134Z"}`, `{"ws_victory":"1903-10-13T21:30:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1903-10-13T21:30:00.000113456Z"}`, `{"ws_victory":"1903-10-13T21:30:00.000Z"}`)

	// RFC 3339 date-time without milliseconds with Z tz-offset.
	roundTripTestDateTime(t, `{"ws_victory":"1912-10-16T19:34:00Z"}`, `{"ws_victory":"1912-10-16T19:34:00.000Z"}`)

	// RFC 3339 date-time with milliseconds with non-Z tz-offset.
	roundTripTestDateTime(t, `{"ws_victory":"1915-10-13T16:15:00.000-03:00"}`, `{"ws_victory":"1915-10-13T19:15:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1915-10-13T22:15:00.000+0300"}`, `{"ws_victory":"1915-10-13T19:15:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1915-10-13T16:15:00.000-03"}`, `{"ws_victory":"1915-10-13T19:15:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1915-10-13T22:15:00.000+03"}`, `{"ws_victory":"1915-10-13T19:15:00.000Z"}`)

	// RFC 3339 date-time without milliseconds with non-Z tz-offset.
	roundTripTestDateTime(t, `{"ws_victory":"1916-10-12T13:43:00-05:00"}`, `{"ws_victory":"1916-10-12T18:43:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1916-10-12T13:43:00-05"}`, `{"ws_victory":"1916-10-12T18:43:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1916-10-12T21:13:00+0230"}`, `{"ws_victory":"1916-10-12T18:43:00.000Z"}`)

	// RFC 3339 with nanoseconds for the Catalog-Managements of the world.
	roundTripTestDateTime(t, `{"ws_victory":"1916-10-12T13:43:00.866305005-05:00"}`, `{"ws_victory":"1916-10-12T18:43:00.866Z"}`)

	// UTC date-time with no tz.
	roundTripTestDateTime(t, `{"ws_victory":"1918-09-11T19:06:00.000"}`, `{"ws_victory":"1918-09-11T19:06:00.000Z"}`)
	roundTripTestDateTime(t, `{"ws_victory":"1918-09-11T19:06:00"}`, `{"ws_victory":"1918-09-11T19:06:00.000Z"}`)

	// Dialog date-time.
	roundTripTestDateTime(t, `{"ws_victory":"2004-10-28 04:39:00"}`, `{"ws_victory":"2004-10-28T04:39:00.000Z"}`)
}

func TestModelsDate(t *testing.T) {
	roundTripTestDate(t, `{"ws_victory":"1903-10-13"}`, `{"ws_victory":"1903-10-13"}`)
	roundTripTestDate(t, `{"ws_victory":"1912-10-16"}`, `{"ws_victory":"1912-10-16"}`)
	roundTripTestDate(t, `{"ws_victory":"1915-10-13"}`, `{"ws_victory":"1915-10-13"}`)
	roundTripTestDate(t, `{"ws_victory":"1916-10-12"}`, `{"ws_victory":"1916-10-12"}`)
	roundTripTestDate(t, `{"ws_victory":"1918-09-11"}`, `{"ws_victory":"1918-09-11"}`)
	roundTripTestDate(t, `{"ws_victory":"2004-10-28"}`, `{"ws_victory":"2004-10-28"}`)
	roundTripTestDate(t, `{"ws_victory":"2007-10-29"}`, `{"ws_victory":"2007-10-29"}`)
	roundTripTestDate(t, `{"ws_victory":"2013-10-31"}`, `{"ws_victory":"2013-10-31"}`)
	roundTripTestDate(t, `{"ws_victory":"2018-10-29"}`, `{"ws_victory":"2018-10-29"}`)
}

func TestDateTimeUtil(t *testing.T) {
	dateVar := strfmt.Date(time.Now())
	fmtDate, err := ParseDate(dateVar.String())
	assert.Nil(t, err)
	assert.Equal(t, dateVar.String(), fmtDate.String())

	fmtDate, err = ParseDate("not a date")
	assert.Equal(t, strfmt.Date{}, fmtDate)
	assert.NotNil(t, err)

	fmtDate, err = ParseDate("")
	assert.Equal(t, strfmt.Date(time.Unix(0, 0).UTC()), fmtDate)
	assert.Nil(t, err)

	dateTimeVar := strfmt.DateTime(time.Now())
	var fmtDTime strfmt.DateTime
	fmtDTime, err = ParseDateTime(dateTimeVar.String())
	assert.Nil(t, err)
	assert.Equal(t, dateTimeVar.String(), fmtDTime.String())

	fmtDTime, err = ParseDateTime("not a datetime")
	assert.Equal(t, strfmt.DateTime{}, fmtDTime)
	assert.NotNil(t, err)

	fmtDTime, err = ParseDateTime("")
	assert.Equal(t, strfmt.NewDateTime(), fmtDTime)
	assert.Nil(t, err)
}
