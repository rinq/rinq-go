package cbor

import (
	"bytes"
	"io"
	"sync"

	"github.com/ugorji/go/codec"
)

// Nil is the nil value encoded in CBOR
var Nil []byte

var encoders sync.Pool
var decoders sync.Pool

// Encode writes v to w in CBOR format.
func Encode(w io.Writer, v interface{}) error {
	e := encoders.Get().(*codec.Encoder)
	defer encoders.Put(e)

	e.Reset(w)
	return e.Encode(v)
}

// MustEncode writes v to w in CBOR format, or panics if unable to do so.
func MustEncode(w io.Writer, v interface{}) {
	e := encoders.Get().(*codec.Encoder)
	defer encoders.Put(e)

	e.Reset(w)
	e.MustEncode(v)
}

// Decode reads CBOR data from r and unpacks into v.
func Decode(r io.Reader, v interface{}) error {
	d := decoders.Get().(*codec.Decoder)
	defer decoders.Put(d)

	d.Reset(r)
	return d.Decode(v)
}

// MustDecode reads CBOR data from r and unpacks into v, or panics if unable to
// do so.
func MustDecode(r io.Reader, v interface{}) {
	d := decoders.Get().(*codec.Decoder)
	defer decoders.Put(d)

	d.Reset(r)
	d.MustDecode(v)
}

// DecodeBytes parses CBOR data in b and unpacks into v.
func DecodeBytes(b []byte, v interface{}) error {
	d := decoders.Get().(*codec.Decoder)
	defer decoders.Put(d)

	d.ResetBytes(b)
	return d.Decode(v)
}

// MustDecodeBytes parses CBOR data in b and unpacks into v, or panics if unable
// to do so.
func MustDecodeBytes(b []byte, v interface{}) {
	d := decoders.Get().(*codec.Decoder)
	defer decoders.Put(d)

	d.ResetBytes(b)
	d.MustDecode(v)
}

func init() {
	var handle codec.CborHandle

	encoders.New = func() interface{} {
		return codec.NewEncoder(nil, &handle)
	}

	decoders.New = func() interface{} {
		return codec.NewDecoder(nil, &handle)
	}

	e := encoders.Get().(*codec.Encoder)
	defer encoders.Put(e)

	var buffer bytes.Buffer
	e.Reset(&buffer)
	e.MustEncode(nil)

	Nil = buffer.Bytes()
}
