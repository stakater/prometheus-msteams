/*
Copyright 2026 Stakater.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package utility provides utility functions and types for the application,
// such as logging.
package utility

import (
	"fmt"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// LoggerFormat represents the format of the logs.
type LoggerFormat string

const (
	// LogFormatJSON represents the JSON log format.
	LogFormatJSON LoggerFormat = "json"
	// LogFormatFmt represents the logfmt log format.
	LogFormatFmt LoggerFormat = "fmt"
)

// Logger is a wrapper around go-kit log.Logger that provides structured
// logging with different levels.
type Logger struct {
	log log.Logger
}

// NewLogger creates a new Logger with the specified format and debug level.
func NewLogger(format LoggerFormat, debug bool) *Logger {
	l := &Logger{}
	var logger log.Logger
	{
		switch format {
		case LogFormatJSON:
			logger = log.NewJSONLogger(log.NewSyncWriter(os.Stdout))
		case LogFormatFmt:
			logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
		default:
			fmt.Fprintf(os.Stderr, "log-format '%s' is not valid", format)
			os.Exit(1)
		}
		l.log = logger
		l.SetDebug(debug)
		l.log = log.With(l.log, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
	}
	return l
}

// SetDebug sets the debug level for the logger.
// If debug is true, debug messages will be logged.
func (l *Logger) SetDebug(value bool) *Logger {
	if value {
		l.log = level.NewFilter(l.log, level.AllowDebug())
	} else {
		l.log = level.NewFilter(l.log, level.AllowInfo())
	}
	return l
}

// GetLogger returns the underlying go-kit log.Logger.
func (l *Logger) GetLogger() log.Logger {
	return l.log
}

// With adds keyvals to the logger context and returns a new Logger.
func (l *Logger) With(keyvals ...any) *Logger {
	return &Logger{log: log.With(l.log, keyvals...)}
}

// Log logs a message with keyvals.
func (l *Logger) Log(keyvals ...any) {
	err := l.log.Log(keyvals...)
	if err != nil {
		l.stdErr(err)
	}
}

// Debug logs a debug message with keyvals.
func (l *Logger) Debug(keyvals ...any) {
	err := level.Debug(l.log).Log(keyvals...)
	if err != nil {
		l.stdErr(err)
	}
}

// Info logs an info message with keyvals.
func (l *Logger) Info(keyvals ...any) {
	err := level.Info(l.log).Log(keyvals...)
	if err != nil {
		l.stdErr(err)
	}
}

// Warn logs a warning message with keyvals.
func (l *Logger) Warn(keyvals ...any) {
	err := level.Warn(l.log).Log(keyvals...)
	if err != nil {
		l.stdErr(err)
	}
}

// Err logs an error type with additional keyvals.
func (l *Logger) Err(err error, keyvals ...any) {
	keyvals = append([]any{"err", err}, keyvals...)
	err = level.Error(l.log).Log(keyvals...)
	if err != nil {
		l.stdErr(err)
	}
}

// Error logs an error message with keyvals.
func (l *Logger) Error(keyvals ...any) {
	err := level.Error(l.log).Log(keyvals...)
	if err != nil {
		l.stdErr(err)
	}
}

func (l *Logger) stdErr(err error) {
	fmt.Fprintf(os.Stderr, "%q\n", err.Error())
}
