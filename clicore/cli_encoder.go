package clilog

import (
	"sync"
	"time"

	"github.com/solo-io/glooshot/pkg/pregoutils-clilog/internal/clibufferpool"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap/buffer"
)

// Unlike zap's built-in console encoder, this encoder just prints strings, in the manner of fmt.Println.
type cliEncoder struct {
	buf        *buffer.Buffer
	printedKey string
}

// these interface methods are irrelevant to the current needs of the CLI encoder
func (c *cliEncoder) AddArray(key string, marshaler zapcore.ArrayMarshaler) error   { return nil }
func (c *cliEncoder) AddObject(key string, marshaler zapcore.ObjectMarshaler) error { return nil }
func (c *cliEncoder) AddBinary(key string, value []byte)                            {}
func (c *cliEncoder) AddByteString(key string, value []byte)                        {}
func (c *cliEncoder) AddBool(key string, value bool)                                {}
func (c *cliEncoder) AddComplex128(key string, value complex128)                    {}
func (c *cliEncoder) AddComplex64(key string, value complex64)                      {}
func (c *cliEncoder) AddDuration(key string, value time.Duration)                   {}
func (c *cliEncoder) AddFloat64(key string, value float64)                          {}
func (c *cliEncoder) AddFloat32(key string, value float32)                          {}
func (c *cliEncoder) AddInt(key string, value int)                                  {}
func (c *cliEncoder) AddInt64(key string, value int64)                              {}
func (c *cliEncoder) AddInt32(key string, value int32)                              {}
func (c *cliEncoder) AddInt16(key string, value int16)                              {}
func (c *cliEncoder) AddInt8(key string, value int8)                                {}
func (c *cliEncoder) AddTime(key string, value time.Time)                           {}
func (c *cliEncoder) AddString(key, val string)                                     {}
func (c *cliEncoder) AddUint(key string, value uint)                                {}
func (c *cliEncoder) AddUint64(key string, value uint64)                            {}
func (c *cliEncoder) AddUint32(key string, value uint32)                            {}
func (c *cliEncoder) AddUint16(key string, value uint16)                            {}
func (c *cliEncoder) AddUint8(key string, value uint8)                              {}
func (c *cliEncoder) AddUintptr(key string, value uintptr)                          {}
func (c *cliEncoder) AddReflected(key string, value interface{}) error              { return nil }
func (c *cliEncoder) OpenNamespace(key string)                                      {}

func NewCliEncoder(printedKey string) zapcore.Encoder {
	return &cliEncoder{
		printedKey: printedKey,
		buf:        clibufferpool.Get(),
	}
}

var _encoderPool = sync.Pool{New: func() interface{} {
	return &cliEncoder{}
}}

func getCliEncoder() *cliEncoder {
	return _encoderPool.Get().(*cliEncoder)
}

func putCliEncoder(c *cliEncoder) {
	c.buf = nil
	c.printedKey = ""
	_encoderPool.Put(c)
}

func (c cliEncoder) Clone() zapcore.Encoder {
	clone := c.clone()
	clone.buf.Write(c.buf.Bytes())
	return clone
}
func (c *cliEncoder) clone() *cliEncoder {
	clone := getCliEncoder()
	clone.buf = clibufferpool.Get()
	clone.printedKey = c.printedKey
	return clone
}

//EncodeEntry implements the distinguishing features of this encoder type.
func (c cliEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := c.clone()
	for _, f := range fields {
		if f.Key == c.printedKey && f.String != "" {
			final.buf.AppendString(f.String + "\n")
		}
	}
	ret := final.buf
	putCliEncoder(final)
	return ret, nil
}
