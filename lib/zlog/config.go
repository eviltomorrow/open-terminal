// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zlog

import (
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultLogMaxSize = 100 // MB
)

// FileLogConfig serializes file log related config in toml/json.
type FileLogConfig struct {
	// Log filename, leave empty to disable file log.
	Filename string `toml:"filename" json:"filename"`
	// Max size for a single file, in MB.
	MaxSize int `toml:"max-size" json:"max-size"`
	// Max log keep days, default is never deleting.
	MaxDays int `toml:"max-days" json:"max-days"`
	// Maximum number of old log files to retain.
	MaxBackups int `toml:"max-backups" json:"max-backups"`
	// Compression function for rotated files.
	// Currently only `gzip` and empty are supported, empty means compression disabled.
	Compression string `toml:"compression" json:"compression"`
}

// Config serializes log related config in toml/json.
type Config struct {
	// Enable std log
	DisableStdlog bool `toml:"disable-stdlog" json:"disable-stdlog"`
	// Log level.
	Level string `toml:"level" json:"level" mapstructure:"level"`
	// Log format. One of json or text.
	Format string `toml:"format" json:"format" mapstructure:"format"`
	// Disable automatic timestamps in output.
	DisableTimestamp bool `json:"-"`
	// File log config.
	File FileLogConfig `toml:"file" json:"file" mapstructure:"file"`
	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `json:"-"`
	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool `json:"-"`
	// DisableStacktrace completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	DisableStacktrace bool `json:"-"`
	// DisableErrorVerbose stops annotating logs with the full verbose error
	// message.
	DisableErrorVerbose bool `json:"-"`
	// SamplingConfig sets a sampling strategy for the logger. Sampling caps the
	// global CPU and I/O load that logging puts on your process while attempting
	// to preserve a representative subset of your logs.
	//
	// Values configured here are per-second. See zapcore.NewSampler for details.
	Sampling *zap.SamplingConfig `json:"-"`
	// ErrorOutputPath is a path to write internal logger errors to.
	// If this field is not set, the internal logger errors will be sent to the same file as in File field.
	// Note: if we want to output the logger errors to stderr, we can just set this field to "stderr"
	ErrorOutputPath string `json:"-"`
	// Timeout for writing log, if TiDB hang on writing log, make it panic.
	// The value is seconds, 0 means no timeout
	Timeout int `json:"-"`
}

func (c *Config) Validate() error {
	for _, allowLevel := range []string{"debug", "info", "warn", "error"} {
		if allowLevel == strings.ToLower(c.Level) {
			break
		}
	}
	return nil
}

// ZapProperties records some information about zap.
type ZapProperties struct {
	Core      zapcore.Core
	Syncer    zapcore.WriteSyncer
	ErrSyncer zapcore.WriteSyncer
	Level     zap.AtomicLevel
}

func (cfg *Config) buildOptions(errSink zapcore.WriteSyncer) []zap.Option {
	opts := []zap.Option{zap.ErrorOutput(errSink)}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	stackLevel := zap.ErrorLevel
	if cfg.Development {
		stackLevel = zap.WarnLevel
	}
	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, int(cfg.Sampling.Initial), int(cfg.Sampling.Thereafter))
		}))
	}
	return opts
}
