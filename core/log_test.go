//go:build all || fast || log
// +build all fast log

package core

// (C) Copyright IBM Corp. 2020, 2021.
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

func TestSetLogLevel(t *testing.T) {
	l := NewLogger(LevelNone, nil, nil)
	assert.NotNil(t, l)

	errorLogger := l.errorLog()
	assert.NotNil(t, errorLogger)

	infoLogger := l.infoLog()
	assert.NotNil(t, infoLogger)

	l.SetLogLevel(LevelError)
	assert.Equal(t, LevelError, l.GetLogLevel())
	assert.True(t, l.IsLogLevelEnabled(LevelError))
	assert.False(t, l.IsLogLevelEnabled(LevelWarn))
	assert.False(t, l.IsLogLevelEnabled(LevelInfo))
	assert.False(t, l.IsLogLevelEnabled(LevelDebug))

	l.SetLogLevel(LevelWarn)
	assert.Equal(t, LevelWarn, l.GetLogLevel())
	assert.True(t, l.IsLogLevelEnabled(LevelError))
	assert.True(t, l.IsLogLevelEnabled(LevelWarn))
	assert.False(t, l.IsLogLevelEnabled(LevelInfo))
	assert.False(t, l.IsLogLevelEnabled(LevelDebug))

	l.SetLogLevel(LevelInfo)
	assert.Equal(t, LevelInfo, l.GetLogLevel())
	assert.True(t, l.IsLogLevelEnabled(LevelError))
	assert.True(t, l.IsLogLevelEnabled(LevelWarn))
	assert.True(t, l.IsLogLevelEnabled(LevelInfo))
	assert.False(t, l.IsLogLevelEnabled(LevelDebug))

	l.SetLogLevel(LevelDebug)
	assert.Equal(t, LevelDebug, l.GetLogLevel())
	assert.True(t, l.IsLogLevelEnabled(LevelError))
	assert.True(t, l.IsLogLevelEnabled(LevelWarn))
	assert.True(t, l.IsLogLevelEnabled(LevelInfo))
	assert.True(t, l.IsLogLevelEnabled(LevelDebug))
}

func TestSetLoggingLevel(t *testing.T) {
	l := NewLogger(LevelNone, nil, nil)
	assert.NotNil(t, l)

	SetLogger(l)

	SetLoggingLevel(LevelError)
	assert.Equal(t, LevelError, GetLogger().GetLogLevel())

	SetLoggingLevel(LevelWarn)
	assert.Equal(t, LevelWarn, GetLogger().GetLogLevel())

	SetLoggingLevel(LevelInfo)
	assert.Equal(t, LevelInfo, GetLogger().GetLogLevel())

	SetLoggingLevel(LevelDebug)
	assert.Equal(t, LevelDebug, GetLogger().GetLogLevel())

}

func stringLogger(level LogLevel) (stdout *bytes.Buffer, stderr *bytes.Buffer, logger Logger) {
	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)

	logger = &SDKLoggerImpl{
		logLevel:    level,
		infoLogger:  log.New(stdout, "", 0),
		errorLogger: log.New(stderr, "", 0),
	}
	return
}

func TestLogNone(t *testing.T) {
	stdout, stderr, l := stringLogger(LevelNone)

	l.Error("error msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())

	l.Warn("warn msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())

	l.Info("info msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())

	l.Debug("debug msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestLogError(t *testing.T) {
	stdout, stderr, l := stringLogger(LevelError)

	l.Error("error msg")
	assert.Empty(t, stdout.String())
	assert.Equal(t, "[Error] error msg\n", stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Warn("warn msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Info("info msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Debug("debug msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestLogWarn(t *testing.T) {
	stdout, stderr, l := stringLogger(LevelWarn)

	l.Error("error msg")
	assert.Empty(t, stdout.String())
	assert.Equal(t, "[Error] error msg\n", stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Warn("warn msg")
	assert.Equal(t, "[Warn] warn msg\n", stdout.String())
	assert.Empty(t, stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Info("info msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Debug("debug msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestLogInfo(t *testing.T) {
	stdout, stderr, l := stringLogger(LevelInfo)

	l.Error("error msg")
	assert.Empty(t, stdout.String())
	assert.Equal(t, "[Error] error msg\n", stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Warn("warn msg")
	assert.Equal(t, "[Warn] warn msg\n", stdout.String())
	assert.Empty(t, stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Info("info msg")
	assert.Equal(t, "[Info] info msg\n", stdout.String())
	assert.Empty(t, stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Debug("debug msg")
	assert.Empty(t, stdout.String())
	assert.Empty(t, stderr.String())
}

func TestLogDebug(t *testing.T) {
	stdout, stderr, l := stringLogger(LevelDebug)

	l.Error("error msg")
	assert.Empty(t, stdout.String())
	assert.Equal(t, "[Error] error msg\n", stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Warn("warn msg")
	assert.Equal(t, "[Warn] warn msg\n", stdout.String())
	assert.Empty(t, stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Info("info msg")
	assert.Equal(t, "[Info] info msg\n", stdout.String())
	assert.Empty(t, stderr.String())

	stdout.Reset()
	stderr.Reset()
	l.Debug("debug msg")
	assert.Equal(t, "[Debug] debug msg\n", stdout.String())
	assert.Empty(t, stderr.String())
}
