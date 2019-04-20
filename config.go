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
	"os"

	"github.com/templexxx/zap/zapcore"
)

// Config offers a declarative way to construct a logger. It doesn't do
// anything that can't be done with New, Options, and the various
// zapcore.WriteSyncer and zapcore.Core wrappers, but it's a simpler way to
// toggle common options.
//
// Note that Config intentionally supports only the most common options. More
// unusual logging setups (logging to network connections or message queues,
// splitting output between multiple files, etc.) are possible, but require
// direct use of the zapcore package. For sample code, see the package-level
// BasicConfiguration and AdvancedConfiguration examples.
//
// For an example showing runtime log level changes, see the documentation for
// AtomicLevel.
type Config struct {
	// Level is the minimum enabled logging level. Note that this is a dynamic
	// level, so calling Config.Level.SetLevel will atomically change the log
	// level of all loggers descended from this config.
	Level AtomicLevel `json:"level" yaml:"level"`
	// Encoding sets the logger's encoding. Valid values are "json" and
	// "console", as well as any third-party encodings registered via
	// RegisterEncoder.
	Encoding string `json:"encoding" yaml:"encoding"`
	// EncoderConfig sets options for the chosen encoder. See
	// zapcore.EncoderConfig for details.
	EncoderConfig zapcore.EncoderConfig `json:"encoderConfig" yaml:"encoderConfig"`
	// OutputPath is a URL or file path to write logging output to.
	// See Open for details.
	OutputPath string `json:"outputPath" yaml:"outputPath"`

	// BufSize log write buf,
	// See zapcore/writebuf.go for details.
	BufSize int `json:"bufSize" yaml:"bufSize"`

	// Flush will flush log buf every Flush seconds.
	Flush int `json:"flush" yaml:"flush"`
}

// Build constructs a logger from the Config and Options.
func (cfg Config) Build(opts ...Option) (*Logger, error) {
	enc, err := cfg.buildEncoder()
	if err != nil {
		return nil, err
	}

	syncer, err := openSyncer(cfg)
	if err != nil {
		return nil, err
	}

	log := New(
		zapcore.NewCore(enc, syncer, cfg.Level),
	)
	if len(opts) > 0 {
		log = log.WithOptions(opts...)
	}
	return log, nil
}

func (cfg Config) buildEncoder() (zapcore.Encoder, error) {
	return newEncoder(cfg.Encoding, cfg.EncoderConfig)
}

const defaultFlush = 5

func openSyncer(cfg Config) (zapcore.WriteSyncer, error) {
	switch cfg.OutputPath {
	case "stdout":
		return zapcore.Lock(nopReOpenSyner{os.Stdout}), nil
	case "stderr":
		return zapcore.Lock(nopReOpenSyner{os.Stderr}), nil
	default:
		f, err := os.OpenFile(cfg.OutputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			return nil, err
		}
		if cfg.Flush == 0 {
			cfg.Flush = 5
		}
		return zapcore.Buffer(f, cfg.BufSize, cfg.Flush, cfg.OutputPath), nil
	}
}

type nopReOpenSyner struct {
	*os.File
}

func (nopReOpenSyner) ReOpen() error {
	return nil
}

func DefaultConfig() Config {
	return Config{
		Level:         NewAtomicLevelAt(InfoLevel),
		Encoding:      "json",
		EncoderConfig: DefaultEncoderConf(),
		OutputPath:    "stderr",
	}
}

func DefaultEncoderConf() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:     "time",
		LevelKey:    "level",
		MessageKey:  "msg",
		LineEnding:  zapcore.DefaultLineEnding,
		EncodeLevel: zapcore.LowercaseLevelEncoder,
		EncodeTime:  zapcore.ISO8601TimeEncoder,
	}
}
