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

package zapcore_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/templexxx/zap/zapcore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeInt64Field(key string, val int) Field {
	return Field{Type: Int64Type, Integer: int64(val), Key: key}
}

func TestIOCore(t *testing.T) {
	temp, err := ioutil.TempFile("", "zapcore-test-iocore")
	require.NoError(t, err, "Failed to create temp file.")
	defer os.Remove(temp.Name())

	// Drop timestamps for simpler assertions (timestamp encoding is tested
	// elsewhere).
	cfg := testEncoderConfig()
	cfg.TimeKey = ""

	core := NewCore(
		NewJSONEncoder(cfg),
		Buffer(temp, 32*1024, 1, temp.Name()),
		InfoLevel,
	).With([]Field{makeInt64Field("k", 1)})
	defer assert.NoError(t, core.Sync(), "Expected Syncing a temp file to succeed.")

	if ce := core.Check(Entry{Level: DebugLevel, Message: "debug"}, nil); ce != nil {
		ce.Write(makeInt64Field("k", 2))
	}
	if ce := core.Check(Entry{Level: InfoLevel, Message: "info"}, nil); ce != nil {
		ce.Write(makeInt64Field("k", 3))
	}
	if ce := core.Check(Entry{Level: WarnLevel, Message: "warn"}, nil); ce != nil {
		ce.Write(makeInt64Field("k", 4))
	}
	core.Sync()
	logged, err := ioutil.ReadFile(temp.Name())
	require.NoError(t, err, "Failed to read from temp file.")
	require.Equal(
		t,
		`{"level":"info","msg":"info","k":1,"k":3}`+"\n"+
			`{"level":"warn","msg":"warn","k":1,"k":4}`+"\n",
		string(logged),
		"Unexpected log output.",
	)
}
