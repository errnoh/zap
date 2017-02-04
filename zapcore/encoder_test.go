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
	"testing"
	"time"

	. "go.uber.org/zap/zapcore"

	"github.com/stretchr/testify/assert"
)

func TestLevelEncoders(t *testing.T) {
	tests := []struct {
		text     string
		level    Level
		expected interface{}
	}{
		{"", InfoLevel, "info"},
		{"something-random", InfoLevel, "info"},
		{"default", InfoLevel, "info"},
		{"default", DPanicLevel, "dpanic"},
		{"capital", InfoLevel, "INFO"},
		{"capital", DPanicLevel, "DPANIC"},
	}
	for _, tt := range tests {
		var unmarshaled LevelEncoder
		err := unmarshaled.UnmarshalText([]byte(tt.text))
		assert.NoError(t, err, `Unexpected error unmarshaling text %q to LevelEncoder.`, tt.text)

		encoder := NewMapObjectEncoder()
		assert.NoError(t, encoder.AddArray("k", ArrayMarshalerFunc(func(enc ArrayEncoder) error {
			unmarshaled(tt.level, enc)
			return nil
		})), "Unexpected error using LevelEncoder %q.", tt.text)
		assert.Equal(
			t,
			tt.expected,
			encoder.Fields["k"].([]interface{})[0],
			"Unexpected output from LevelEncoder %q.", tt.text,
		)
	}
}

func TestTimeEncoders(t *testing.T) {
	// FIXME
	_epoch := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		text     string
		time     time.Time
		expected interface{} // expected output from serializing Unix time 0
	}{
		{"", _epoch, float64(0)},
		{"something-random", _epoch, float64(0)},
		{"epoch", time.Unix(1, int64(500*time.Millisecond)), 1.5},
		{"millis", time.Unix(1, int64(500*time.Millisecond)), int64(1500)},
		{"iso8601", _epoch, "1970-01-01T00:00:00Z"},
		{"ISO8601", _epoch, "1970-01-01T00:00:00Z"},
	}
	for _, tt := range tests {
		var unmarshaled TimeEncoder
		err := unmarshaled.UnmarshalText([]byte(tt.text))
		assert.NoError(t, err, `Unexpected error unmarshaling text %q to LevelEncoder.`, tt.text)

		encoder := NewMapObjectEncoder()
		assert.NoError(t, encoder.AddArray("k", ArrayMarshalerFunc(func(enc ArrayEncoder) error {
			unmarshaled(tt.time, enc)
			return nil
		})), "Unexpected error using TimeEncoder %q.", tt.text)
		assert.Equal(
			t,
			tt.expected,
			encoder.Fields["k"].([]interface{})[0],
			"Unexpected output from TimeEncoder %q.", tt.text,
		)
	}
}
