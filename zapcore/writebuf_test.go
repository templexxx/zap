package zapcore_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "github.com/templexxx/zap/zapcore"
)

func TestIoCore_ReOpen(t *testing.T) {
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

	np := "zapcore-drop"
	os.Rename(temp.Name(), np)
	defer os.Remove(np)
	core.ReOpen()

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
	logged, err = ioutil.ReadFile(temp.Name())
	require.NoError(t, err, "Failed to read from temp file.")
	require.Equal(
		t,
		`{"level":"info","msg":"info","k":1,"k":3}`+"\n"+
			`{"level":"warn","msg":"warn","k":1,"k":4}`+"\n",
		string(logged),
		"Unexpected log output.",
	)

	np2 := "zapcore-drop2"
	os.Rename(temp.Name(), np2)
	defer os.Remove(np2)
	core.ReOpen()

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
	logged, err = ioutil.ReadFile(temp.Name())
	require.NoError(t, err, "Failed to read from temp file.")
	require.Equal(
		t,
		`{"level":"info","msg":"info","k":1,"k":3}`+"\n"+
			`{"level":"warn","msg":"warn","k":1,"k":4}`+"\n",
		string(logged),
		"Unexpected log output.",
	)
}

func TestIoCore_Flush(t *testing.T) {
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

	time.Sleep(2 * time.Second) // wait for flushing.

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
