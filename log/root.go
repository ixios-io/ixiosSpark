// IxiosSpark is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// This file is part of the IxiosSpark library, which builds upon the source code of the geth library.
// The IxiosSpark source code is distributed with the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
// Copyright 2025 The ixiosSpark Authors, Copyright 2015-2024 The go-ethereum Authors (geth)
// You should have received a copy of the GNU Lesser General Public License
// with IxiosSpark. If not, see <http://www.gnu.org/licenses/>.

package log

import (
	"os"
	"sync/atomic"

	"golang.org/x/exp/slog"
)

var root atomic.Value

func init() {
	root.Store(&logger{slog.New(DiscardHandler())})
}

// SetDefault sets the default global logger
func SetDefault(l Logger) {
	root.Store(l)
	if lg, ok := l.(*logger); ok {
		slog.SetDefault(lg.inner)
	}
}

// Root returns the root logger
func Root() Logger {
	return root.Load().(Logger)
}

// The following functions bypass the exported logger methods (logger.Debug,
// etc.) to keep the call depth the same for all paths to logger.Write so
// runtime.Caller(2) always refers to the call site in client code.

// Trace is a convenient alias for Root().Trace
//
// Log a message at the trace level with context key/value pairs
//
// # Usage
//
//	log.Trace("msg")
//	log.Trace("msg", "key1", val1)
//	log.Trace("msg", "key1", val1, "key2", val2)
func Trace(msg string, ctx ...interface{}) {
	Root().Write(LevelTrace, msg, ctx...)
}

// Debug is a convenient alias for Root().Debug
//
// Log a message at the debug level with context key/value pairs
//
// # Usage Examples
//
//	log.Debug("msg")
//	log.Debug("msg", "key1", val1)
//	log.Debug("msg", "key1", val1, "key2", val2)
func Debug(msg string, ctx ...interface{}) {
	Root().Write(slog.LevelDebug, msg, ctx...)
}

// Info is a convenient alias for Root().Info
//
// Log a message at the info level with context key/value pairs
//
// # Usage Examples
//
//	log.Info("msg")
//	log.Info("msg", "key1", val1)
//	log.Info("msg", "key1", val1, "key2", val2)
func Info(msg string, ctx ...interface{}) {
	Root().Write(slog.LevelInfo, msg, ctx...)
}

// Warn is a convenient alias for Root().Warn
//
// Log a message at the warn level with context key/value pairs
//
// # Usage Examples
//
//	log.Warn("msg")
//	log.Warn("msg", "key1", val1)
//	log.Warn("msg", "key1", val1, "key2", val2)
func Warn(msg string, ctx ...interface{}) {
	Root().Write(slog.LevelWarn, msg, ctx...)
}

// Error is a convenient alias for Root().Error
//
// Log a message at the error level with context key/value pairs
//
// # Usage Examples
//
//	log.Error("msg")
//	log.Error("msg", "key1", val1)
//	log.Error("msg", "key1", val1, "key2", val2)
func Error(msg string, ctx ...interface{}) {
	Root().Write(slog.LevelError, msg, ctx...)
}

// Crit is a convenient alias for Root().Crit
//
// Log a message at the crit level with context key/value pairs, and then exit.
//
// # Usage Examples
//
//	log.Crit("msg")
//	log.Crit("msg", "key1", val1)
//	log.Crit("msg", "key1", val1, "key2", val2)
func Crit(msg string, ctx ...interface{}) {
	Root().Write(LevelCrit, msg, ctx...)
	os.Exit(1)
}

// New returns a new logger with the given context.
// New is a convenient alias for Root().New
func New(ctx ...interface{}) Logger {
	return Root().With(ctx...)
}
