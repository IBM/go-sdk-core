// +build all fast log

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
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func stringLogger(level LogLevel) (buf *bytes.Buffer, logger Logger) {
	buf = new(bytes.Buffer)
	logger = &SDKLoggerImpl{
		logLevel: level,
		goLogger: log.New(buf, "", 0),
	}
	return
}
func TestLogNone(t *testing.T) {
	buf, l := stringLogger(LevelNone)

	l.Error("error msg")
	assert.Empty(t, buf.String())

	l.Warn("warn msg")
	assert.Empty(t, buf.String())

	l.Info("info msg")
	assert.Empty(t, buf.String())

	l.Debug("debug msg")
	assert.Empty(t, buf.String())
}

func TestLogError(t *testing.T) {
	buf, l := stringLogger(LevelError)

	l.Error("error msg")
	assert.Equal(t, "[Error] error msg\n", buf.String())

	buf.Reset()
	l.Warn("warn msg")
	assert.Empty(t, buf.String())

	buf.Reset()
	l.Info("info msg")
	assert.Empty(t, buf.String())

	buf.Reset()
	l.Debug("debug msg")
	assert.Empty(t, buf.String())
}

func TestLogWarn(t *testing.T) {
	buf, l := stringLogger(LevelWarn)

	l.Error("error msg")
	assert.Equal(t, "[Error] error msg\n", buf.String())

	buf.Reset()
	l.Warn("warn msg")
	assert.Equal(t, "[Warn] warn msg\n", buf.String())

	buf.Reset()
	l.Info("info msg")
	assert.Empty(t, buf.String())

	buf.Reset()
	l.Debug("debug msg")
	assert.Empty(t, buf.String())
}

func TestLogInfo(t *testing.T) {
	buf, l := stringLogger(LevelInfo)

	l.Error("error msg")
	assert.Equal(t, "[Error] error msg\n", buf.String())

	buf.Reset()
	l.Warn("warn msg")
	assert.Equal(t, "[Warn] warn msg\n", buf.String())

	buf.Reset()
	l.Info("info msg")
	assert.Equal(t, "[Info] info msg\n", buf.String())

	buf.Reset()
	l.Debug("debug msg")
	assert.Empty(t, buf.String())
}

func TestLogDebug(t *testing.T) {
	buf, l := stringLogger(LevelDebug)

	l.Error("error msg")
	assert.Equal(t, "[Error] error msg\n", buf.String())

	buf.Reset()
	l.Warn("warn msg")
	assert.Equal(t, "[Warn] warn msg\n", buf.String())

	buf.Reset()
	l.Info("info msg")
	assert.Equal(t, "[Info] info msg\n", buf.String())

	buf.Reset()
	l.Debug("debug msg")
	assert.Equal(t, "[Debug] debug msg\n", buf.String())
}

func TestSetLoggingLevel(t *testing.T) {
	buf, l := stringLogger(LevelError)

	SetLogger(l)

	l.Debug("debug msg")
	assert.Empty(t, buf.String())
	buf.Reset()

	GetLogger().Debug("debug msg")
	assert.Empty(t, buf.String())
	buf.Reset()

	SetLoggingLevel(LevelDebug)

	l.Debug("debug msg")
	assert.Equal(t, "[Debug] debug msg\n", buf.String())
	buf.Reset()

	GetLogger().Debug("debug msg")
	assert.Equal(t, "[Debug] debug msg\n", buf.String())
	buf.Reset()
}
