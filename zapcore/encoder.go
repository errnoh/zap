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

package zapcore

import (
	"strings"
	"time"
)

// A LevelEncoder serializes a Level to a primitive type.
type LevelEncoder func(Level, PrimitiveArrayEncoder)

// LowercaseLevelEncoder serializes a Level to a lowercase string. For example,
// InfoLevel is serialized to "info".
func LowercaseLevelEncoder(l Level, enc PrimitiveArrayEncoder) {
	enc.AppendString(l.String())
}

// CapitalLevelEncoder serializes a Level to an all-caps string. For example,
// InfoLevel is serialized to "INFO".
func CapitalLevelEncoder(l Level, enc PrimitiveArrayEncoder) {
	enc.AppendString(strings.ToUpper(l.String()))
}

// UnmarshalText unmarshals text to a LevelEncoder. "capital" is unmarshaled to
// CapitalLevelEncoder, and anything else is unmarshaled to LowercaseLevelEncoder.
func (e *LevelEncoder) UnmarshalText(text []byte) error {
	switch string(text) {
	case "capital":
		*e = CapitalLevelEncoder
	default:
		*e = LowercaseLevelEncoder
	}
	return nil
}

// A TimeEncoder serializes a time.Time to a primitive type.
type TimeEncoder func(time.Time, PrimitiveArrayEncoder)

// EpochTimeEncoder serializes a time.Time to a floating-point number of seconds
// since the Unix epoch.
func EpochTimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	nanos := t.UnixNano()
	sec := float64(nanos) / float64(time.Second)
	enc.AppendFloat64(sec)
}

// EpochMillisTimeEncoder serializes a time.Time to an int64 number of milliseconds
// since the Unix epoch.
func EpochMillisTimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	nanos := t.UnixNano()
	millis := nanos / int64(time.Millisecond)
	enc.AppendInt64(millis)
}

// ISO8601TimeEncoder serializes a time.Time to an ISO8601-formatted string.
func ISO8601TimeEncoder(t time.Time, enc PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05Z0700"))
}

// UnmarshalText unmarshals text to a TimeEncoder. "iso8601" and "ISO8601" are
// unmarshaled to ISO8601TimeEncoder, "millis" is unmarshaled to
// EpochMillisTimeEncoder, and anything else is unmarshaled to EpochTimeEncoder.
func (e *TimeEncoder) UnmarshalText(text []byte) error {
	switch string(text) {
	case "iso8601", "ISO8601":
		*e = ISO8601TimeEncoder
	case "millis":
		*e = EpochMillisTimeEncoder
	default:
		*e = EpochTimeEncoder
	}
	return nil
}

// A DurationEncoder serializes a time.Duration to a primitive type.
type DurationEncoder func(time.Duration, PrimitiveArrayEncoder)

// NanosDurationEncoder serializes a time.Duration to an int64 number of
// nanoseconds elapsed.
func NanosDurationEncoder(d time.Duration, enc PrimitiveArrayEncoder) {
	enc.AppendInt64(int64(d))
}

// StringDurationEncoder serializes a time.Duration using its built-in String
// method.
func StringDurationEncoder(d time.Duration, enc PrimitiveArrayEncoder) {
	enc.AppendString(d.String())
}

// UnmarshalText unmarshals text to a DurationEncoder. "string" is unmarshaled
// to StringDurationEncoder, and anything else is unmarshaled to
// NanosDurationEncoder.
func (e *DurationEncoder) UnmarshalText(text []byte) error {
	switch string(text) {
	case "string":
		*e = StringDurationEncoder
	default:
		*e = NanosDurationEncoder
	}
	return nil
}

// An EncoderConfig allows users to configure the concrete encoders supplied by
// zapcore.
type EncoderConfig struct {
	// Set the keys used for each log entry.
	MessageKey    string `json:"message_key",yaml:"message_key"`
	LevelKey      string `json:"level_key",yaml:"level_key"`
	TimeKey       string `json:"time_key",yaml:"time_key"`
	NameKey       string `json:"name_key",yaml:"name_key"`
	CallerKey     string `json:"caller_key",yaml:"caller_key"`
	StacktraceKey string `json:"stacktrace_key",yaml:"stacktrace_key"`
	// Configure the primitive representations of common complex types. For
	// example, some users may want all time.Times serialized as floating-point
	// seconds since epoch, while others may prefer ISO8601 strings.
	EncodeLevel    LevelEncoder    `json:"level_encoder",yaml:"level_encoder"`
	EncodeTime     TimeEncoder     `json:"time_encoder",yaml:"time_encoder"`
	EncodeDuration DurationEncoder `json:"duration_encoder",yaml:"duration_encoder"`
}

// ObjectEncoder is a strongly-typed, encoding-agnostic interface for adding a
// map- or struct-like object to the logging context. Like maps, ObjectEncoders
// aren't safe for concurrent use (though typical use shouldn't require locks).
type ObjectEncoder interface {
	// Logging-specific marshalers.
	AddArray(key string, marshaler ArrayMarshaler) error
	AddObject(key string, marshaler ObjectMarshaler) error

	// Built-in types.
	AddBool(key string, value bool)
	AddByte(key string, value byte)
	AddComplex128(key string, value complex128)
	AddComplex64(key string, value complex64)
	AddDuration(key string, value time.Duration)
	AddFloat64(key string, value float64)
	AddFloat32(key string, value float32)
	AddInt(key string, value int)
	AddInt64(key string, value int64)
	AddInt32(key string, value int32)
	AddInt16(key string, value int16)
	AddInt8(key string, value int8)
	AddRune(key string, value rune)
	AddString(key, value string)
	AddTime(key string, value time.Time)
	AddUint(key string, value uint)
	AddUint64(key string, value uint64)
	AddUint32(key string, value uint32)
	AddUint16(key string, value uint16)
	AddUint8(key string, value uint8)
	AddUintptr(key string, value uintptr)

	// AddReflected uses reflection to serialize arbitrary objects, so it's slow
	// and allocation-heavy.
	AddReflected(key string, value interface{}) error
	// OpenNamespace opens an isolated namespace where all subsequent fields will
	// be added. Applications can use namespaces to prevent key collisions when
	// injecting loggers into sub-components or third-party libraries.
	OpenNamespace(key string)
}

// ArrayEncoder is a strongly-typed, encoding-agnostic interface for adding
// array-like objects to the logging context. Of note, it supports mixed-type
// arrays even though they aren't typical in Go. Like slices, ArrayEncoders
// aren't safe for concurrent use (though typical use shouldn't require locks).
type ArrayEncoder interface {
	// Built-in types
	PrimitiveArrayEncoder

	// Logging-specific marshalers.
	AppendArray(ArrayMarshaler) error
	AppendObject(ObjectMarshaler) error

	// Time-related types.
	AppendDuration(time.Duration)
	AppendTime(time.Time)

	// AppendReflected uses reflection to serialize arbitrary objects, so it's
	// slow and allocation-heavy.
	AppendReflected(value interface{}) error
}

// PrimitiveArrayEncoder is a subset of the ArrayEncoder type that only covers
// the basic types.
//
// It exists largely so that Duration-, Time-, and LevelEncoders can't make
// recursive calls.
type PrimitiveArrayEncoder interface {
	AppendBool(bool)
	AppendByte(byte)
	AppendComplex128(complex128)
	AppendComplex64(complex64)
	AppendFloat64(float64)
	AppendFloat32(float32)
	AppendInt(int)
	AppendInt64(int64)
	AppendInt32(int32)
	AppendInt16(int16)
	AppendInt8(int8)
	AppendRune(rune)
	AppendString(string)
	AppendUint(uint)
	AppendUint64(uint64)
	AppendUint32(uint32)
	AppendUint16(uint16)
	AppendUint8(uint8)
	AppendUintptr(uintptr)
}

// Encoder is a format-agnostic interface for all log entry marshalers. Since
// log encoders don't need to support the same wide range of use cases as
// general-purpose marshalers, it's possible to make them faster and
// lower-allocation.
//
// Implementations of the ObjectEncoder interface's methods can, of course,
// freely modify the receiver. However, the Clone and EncodeEntry methods will
// be called concurrently and shouldn't modify the receiver.
type Encoder interface {
	ObjectEncoder

	// Clone copies the encoder, ensuring that adding fields to the copy doesn't
	// affect the original.
	Clone() Encoder

	// EncodeEntry encodes an entry and fields, along with any accumulated
	// context, into a byte buffer and returns it.
	EncodeEntry(Entry, []Field) ([]byte, error)
}
