// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package zap

import (
	"fmt"
	"time"

	"go.uber.org/zap/zapcore"
)

// In bytes. On Linux, writing 4 kibibytes or less when the file is open in
// append-only mode doesn't require coordination between processes, so this is
// nice default.
const _defaultBufferSize = 4 * 1024

// SamplingConfig FIXME
type SamplingConfig struct {
	Initial    uint16 `json:"initial",yaml:"initial"`
	Thereafter uint16 `json:"therafter",yaml:"thereafter"`
}

// Config FIXME
type Config struct {
	Level             AtomicLevel            `json:"level",yaml:"level"`
	Development       bool                   `json:"development",yaml:"development"`
	DisableCaller     bool                   `json:"disable_caller",yaml:"disable_caller"`
	DisableStacktrace bool                   `json:"disable_stacktrace",yaml:"disable_stacktrace"`
	Sampling          *SamplingConfig        `json:"sampling",yaml:"sampling"`
	Encoding          string                 `json:"encoding",yaml:"encoding"`
	EncoderConfig     zapcore.EncoderConfig  `json:"encoder_config",yaml:"encoder_config"`
	OutputPaths       []string               `json:"output_paths",yaml:"output_paths"`
	DisableBuffering  bool                   `json:"disable_buffering",yaml:"disable_buffering"`
	ErrorOutputPaths  []string               `json:"error_output_paths",yaml:"error_output_paths"`
	InitialFields     map[string]interface{} `json:"initial_fields",yaml:"initial_fields"`
}

// NewProductionConfig FIXME
func NewProductionConfig() Config {
	return Config{
		Level:       DynamicLevel(),
		Development: false,
		Sampling: &SamplingConfig{
			Initial:    100,
			Thereafter: 1000,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.NanosDurationEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// NewDevelopmentConfig FIXME
func NewDevelopmentConfig() Config {
	dyn := DynamicLevel()
	dyn.SetLevel(DebugLevel)

	return Config{
		Level:       dyn,
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "T",
			LevelKey:       "L",
			NameKey:        "N",
			CallerKey:      "C",
			MessageKey:     "M",
			StacktraceKey:  "S",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// Build constructs a logger from the Config and Options.
func (cfg Config) Build(opts ...Option) (*Logger, error) {
	sink, errSink, err := cfg.openSinks()
	if err != nil {
		return nil, err
	}
	enc, err := cfg.buildEncoder()
	if err != nil {
		return nil, err
	}

	fac := zapcore.WriterFacility(enc, sink, cfg.Level)
	if cfg.Sampling != nil {
		fac = zapcore.Sample(fac, time.Second, int(cfg.Sampling.Initial), int(cfg.Sampling.Thereafter))
	}

	return New(fac, cfg.buildOptions()...).WithOptions(opts...), nil
}

func (cfg Config) buildOptions() []Option {
	stackLevel := WarnLevel
	if cfg.Development {
		stackLevel = ErrorLevel
	}
	// FIXME: left off here.
}

func (cfg Config) openSinks() (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	bufSize := _defaultBufferSize
	if cfg.Development || cfg.DisableBuffering {
		bufSize = 0
	}
	sink, err := Open(bufSize, cfg.OutputPaths...)
	if err != nil {
		return nil, nil, err
	}
	errSink, err := Open(bufSize, cfg.ErrorOutputPaths...)
	if err != nil {
		return nil, nil, err
	}
	return sink, errSink, nil
}

func (cfg Config) buildEncoder() (zapcore.Encoder, error) {
	switch cfg.Encoding {
	case "json":
		return zapcore.NewJSONEncoder(cfg.EncoderConfig), nil
	case "console":
		// TODO: Use the console encoder.
		return zapcore.NewJSONEncoder(cfg.EncoderConfig), nil
	}
	return nil, fmt.Errorf("unknown encoding %q", cfg.Encoding)
}
